package ingestion

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/espn"
	"github.com/francisco/gridironmind/internal/models"
	"github.com/francisco/gridironmind/internal/nflverse"
	"github.com/francisco/gridironmind/internal/weather"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles data ingestion from ESPN, nflverse, and weather APIs
type Service struct {
	espnClient     *espn.Client
	nflverseClient *nflverse.Client
	weatherClient  *weather.Client
	dbPool         *pgxpool.Pool
}

// NewService creates a new ingestion service
func NewService(weatherAPIKey string) *Service {
	var weatherClient *weather.Client
	if weatherAPIKey != "" {
		weatherClient = weather.NewClient(weatherAPIKey)
	}

	return &Service{
		espnClient:     espn.NewClient(),
		nflverseClient: nflverse.NewClient(),
		weatherClient:  weatherClient,
		dbPool:         db.GetPool(),
	}
}

// SyncTeams fetches all teams from ESPN and updates the database
func (s *Service) SyncTeams(ctx context.Context) error {
	log.Println("Starting teams sync...")

	teamsResp, err := s.espnClient.FetchAllTeams(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch teams from ESPN: %w", err)
	}

	if len(teamsResp.Sports) == 0 || len(teamsResp.Sports[0].Leagues) == 0 {
		return fmt.Errorf("no teams found in ESPN response")
	}

	teams := teamsResp.Sports[0].Leagues[0].Teams
	log.Printf("Fetched %d teams from ESPN", len(teams))

	for _, teamEntry := range teams {
		team := teamEntry.Team
		if !team.IsActive {
			continue
		}

		nflID, err := strconv.Atoi(team.ID)
		if err != nil {
			log.Printf("Skipping team %s: invalid ID", team.DisplayName)
			continue
		}

		// Check if team exists
		var existingID uuid.UUID
		err = s.dbPool.QueryRow(ctx,
			"SELECT id FROM teams WHERE nfl_id = $1",
			nflID,
		).Scan(&existingID)

		now := time.Now()
		if err != nil {
			// Team doesn't exist, insert it
			id := uuid.New()
			_, err = s.dbPool.Exec(ctx,
				`INSERT INTO teams (id, nfl_id, name, abbreviation, city, conference, division, stadium, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				id, nflID, team.Name, team.Abbreviation, team.Location,
				"", "", "", now, now, // conference, division, stadium will be updated in roster sync
			)
			if err != nil {
				log.Printf("Failed to insert team %s: %v", team.DisplayName, err)
				continue
			}
			log.Printf("Inserted new team: %s (%s)", team.DisplayName, team.Abbreviation)
		} else {
			// Team exists, update it
			_, err = s.dbPool.Exec(ctx,
				`UPDATE teams
				SET name = $1, abbreviation = $2, city = $3, updated_at = $4
				WHERE id = $5`,
				team.Name, team.Abbreviation, team.Location, now, existingID,
			)
			if err != nil {
				log.Printf("Failed to update team %s: %v", team.DisplayName, err)
				continue
			}
			log.Printf("Updated team: %s (%s)", team.DisplayName, team.Abbreviation)
		}
	}

	log.Println("Teams sync completed")
	return nil
}

// SyncTeamRoster fetches and updates roster for a specific team
func (s *Service) SyncTeamRoster(ctx context.Context, espnTeamID string) error {
	log.Printf("Syncing roster for team %s...", espnTeamID)

	teamResp, err := s.espnClient.FetchTeamRoster(ctx, espnTeamID)
	if err != nil {
		return fmt.Errorf("failed to fetch team roster: %w", err)
	}

	// Get our internal team ID
	nflID, _ := strconv.Atoi(espnTeamID)
	var teamID uuid.UUID
	err = s.dbPool.QueryRow(ctx, "SELECT id FROM teams WHERE nfl_id = $1", nflID).Scan(&teamID)
	if err != nil {
		return fmt.Errorf("team not found in database: %w", err)
	}

	// Process each player in the roster
	for _, athlete := range teamResp.Roster.Athletes {
		if err := s.upsertPlayer(ctx, athlete, teamID); err != nil {
			log.Printf("Failed to upsert player %s: %v", athlete.DisplayName, err)
		}
	}

	log.Printf("Roster sync completed for team %s", espnTeamID)
	return nil
}

// SyncAllRosters syncs rosters for all teams
func (s *Service) SyncAllRosters(ctx context.Context) error {
	log.Println("Starting all rosters sync...")

	// Get all teams
	rows, err := s.dbPool.Query(ctx, "SELECT nfl_id FROM teams")
	if err != nil {
		return fmt.Errorf("failed to fetch teams: %w", err)
	}
	defer rows.Close()

	var teamIDs []int
	for rows.Next() {
		var nflID int
		if err := rows.Scan(&nflID); err != nil {
			log.Printf("Failed to scan team ID: %v", err)
			continue
		}
		teamIDs = append(teamIDs, nflID)
	}

	// Sync each team's roster with a delay to avoid rate limiting
	for _, nflID := range teamIDs {
		if err := s.SyncTeamRoster(ctx, strconv.Itoa(nflID)); err != nil {
			log.Printf("Failed to sync roster for team %d: %v", nflID, err)
		}
		// Add delay between requests to avoid rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Println("All rosters sync completed")
	return nil
}

// SyncGames fetches current games/scoreboard and updates database
func (s *Service) SyncGames(ctx context.Context) error {
	log.Println("Starting games sync...")

	scoreboard, err := s.espnClient.FetchScoreboard(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch scoreboard: %w", err)
	}

	log.Printf("Fetched %d games from ESPN", len(scoreboard.Events))

	for _, event := range scoreboard.Events {
		if err := s.upsertGame(ctx, event); err != nil {
			log.Printf("Failed to upsert game %s: %v", event.Name, err)
		}
	}

	log.Println("Games sync completed")
	return nil
}

// upsertPlayer inserts or updates a player in the database
func (s *Service) upsertPlayer(ctx context.Context, athlete espn.Athlete, teamID uuid.UUID) error {
	nflID, err := strconv.Atoi(athlete.ID)
	if err != nil {
		return fmt.Errorf("invalid athlete ID: %w", err)
	}

	jerseyNum := 0
	if athlete.Jersey != "" {
		jerseyNum, _ = strconv.Atoi(athlete.Jersey)
	}

	heightInches := int(athlete.Height)
	weightPounds := int(athlete.Weight)

	status := "active"
	if !athlete.Active {
		status = "inactive"
	}
	if athlete.Status != nil && athlete.Status.Type == "injured" {
		status = "injured"
	}

	position := ""
	if athlete.Position.Abbreviation != "" {
		position = athlete.Position.Abbreviation
	}

	// Check if player exists
	var existingID uuid.UUID
	err = s.dbPool.QueryRow(ctx,
		"SELECT id FROM players WHERE nfl_id = $1",
		nflID,
	).Scan(&existingID)

	now := time.Now()
	if err != nil {
		// Player doesn't exist, insert
		id := uuid.New()
		_, err = s.dbPool.Exec(ctx,
			`INSERT INTO players (id, nfl_id, name, position, team_id, jersey_number, height_inches, weight_pounds, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			id, nflID, athlete.DisplayName, position, teamID, jerseyNum, heightInches, weightPounds, status, now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert player: %w", err)
		}
		log.Printf("Inserted player: %s (%s)", athlete.DisplayName, position)
	} else {
		// Player exists, update
		_, err = s.dbPool.Exec(ctx,
			`UPDATE players
			SET name = $1, position = $2, team_id = $3, jersey_number = $4,
			    height_inches = $5, weight_pounds = $6, status = $7, updated_at = $8
			WHERE id = $9`,
			athlete.DisplayName, position, teamID, jerseyNum, heightInches, weightPounds, status, now, existingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update player: %w", err)
		}
		log.Printf("Updated player: %s (%s)", athlete.DisplayName, position)
	}

	return nil
}

// upsertGame inserts or updates a game in the database
func (s *Service) upsertGame(ctx context.Context, event espn.Event) error {
	if len(event.Competitions) == 0 {
		return fmt.Errorf("no competitions found for event")
	}

	competition := event.Competitions[0]
	if len(competition.Competitors) < 2 {
		return fmt.Errorf("insufficient competitors for game")
	}

	// Find home and away teams
	var homeTeam, awayTeam espn.Competitor
	for _, comp := range competition.Competitors {
		if strings.ToLower(comp.HomeAway) == "home" {
			homeTeam = comp
		} else {
			awayTeam = comp
		}
	}

	// Get our internal team IDs
	homeNflID, _ := strconv.Atoi(homeTeam.ID)
	awayNflID, _ := strconv.Atoi(awayTeam.ID)

	var homeTeamID, awayTeamID uuid.UUID
	err := s.dbPool.QueryRow(ctx, "SELECT id FROM teams WHERE nfl_id = $1", homeNflID).Scan(&homeTeamID)
	if err != nil {
		return fmt.Errorf("home team not found: %w", err)
	}

	err = s.dbPool.QueryRow(ctx, "SELECT id FROM teams WHERE nfl_id = $1", awayNflID).Scan(&awayTeamID)
	if err != nil {
		return fmt.Errorf("away team not found: %w", err)
	}

	homeScore, _ := strconv.Atoi(homeTeam.Score)
	awayScore, _ := strconv.Atoi(awayTeam.Score)

	status := "scheduled"
	statusDetail := event.Status.Type.Description
	if event.Status.Type.Completed {
		status = "completed"
	} else if event.Status.Type.State == "in" {
		status = "in_progress"
	}

	// Extract venue information
	venueID := competition.Venue.ID
	venueName := competition.Venue.FullName
	var venueCity, venueState string
	if competition.Venue.Address.City != "" {
		venueCity = competition.Venue.Address.City
	}
	if competition.Venue.Address.State != "" {
		venueState = competition.Venue.Address.State
	}
	attendance := competition.Attendance

	// Extract status details
	var currentPeriod int
	var gameClock string
	if event.Status.Period > 0 {
		currentPeriod = event.Status.Period
	}
	if event.Status.DisplayClock != "" {
		gameClock = event.Status.DisplayClock
	}

	// Check if game exists
	var existingID uuid.UUID
	err = s.dbPool.QueryRow(ctx,
		"SELECT id FROM games WHERE nfl_game_id = $1",
		event.ID,
	).Scan(&existingID)

	if err != nil {
		// Game doesn't exist, insert with comprehensive data
		id := uuid.New()
		_, err = s.dbPool.Exec(ctx,
			`INSERT INTO games (
				id, nfl_game_id, home_team_id, away_team_id, game_date, season, week,
				home_score, away_score, status, status_detail, current_period, game_clock,
				venue_id, venue_name, venue_city, venue_state, attendance
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
			id, event.ID, homeTeamID, awayTeamID, event.Date.Time, event.Season.Year, event.Week.Number,
			homeScore, awayScore, status, statusDetail, currentPeriod, gameClock,
			venueID, venueName, venueCity, venueState, attendance,
		)
		if err != nil {
			return fmt.Errorf("failed to insert game: %w", err)
		}
		log.Printf("Inserted game: %s at %s", event.Name, venueName)
	} else {
		// Game exists, update scores, status, and details
		_, err = s.dbPool.Exec(ctx,
			`UPDATE games
			SET home_score = $1, away_score = $2, status = $3, status_detail = $4,
			    current_period = $5, game_clock = $6, attendance = $7
			WHERE id = $8`,
			homeScore, awayScore, status, statusDetail,
			currentPeriod, gameClock, attendance, existingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update game: %w", err)
		}
		log.Printf("Updated game: %s (%d-%d) - %s", event.Name, homeScore, awayScore, statusDetail)
	}

	return nil
}

// FullSync performs a complete data sync (teams -> rosters -> games)
func (s *Service) FullSync(ctx context.Context) error {
	log.Println("Starting full data sync...")

	// Step 1: Sync teams
	if err := s.SyncTeams(ctx); err != nil {
		return fmt.Errorf("teams sync failed: %w", err)
	}

	// Step 2: Sync all rosters
	if err := s.SyncAllRosters(ctx); err != nil {
		return fmt.Errorf("rosters sync failed: %w", err)
	}

	// Step 3: Sync games
	if err := s.SyncGames(ctx); err != nil {
		return fmt.Errorf("games sync failed: %w", err)
	}

	log.Println("Full data sync completed successfully")
	return nil
}
// SyncPlayerCareerStats fetches and stores complete career statistics for a player
func (s *Service) SyncPlayerCareerStats(ctx context.Context, playerID uuid.UUID, espnAthleteID string) error {
	log.Printf("Syncing career stats for player %s (ESPN ID: %s)...", playerID, espnAthleteID)

	// Fetch player stats from ESPN
	statsResp, err := s.espnClient.FetchPlayerStats(ctx, espnAthleteID)
	if err != nil {
		return fmt.Errorf("failed to fetch player stats: %w", err)
	}

	careerQueries := &db.CareerQueries{}

	// Process season types (regular season stats)
	for _, seasonType := range statsResp.SeasonTypes {
		if seasonType.Type != 2 { // Only regular season
			continue
		}

		// Create career stats entry
		stats := &models.PlayerCareerStats{
			PlayerID:    playerID,
			Season:      seasonType.Year,
			GamesPlayed: 0,
		}

		// Parse stats from categories
		for _, category := range seasonType.Categories {
			for _, stat := range category.Stats {
				value := 0
				switch v := stat.Value.(type) {
				case float64:
					value = int(v)
				case string:
					value, _ = strconv.Atoi(v)
				}

				// Map ESPN stat names to our fields
				switch stat.Abbreviation {
				case "GP":
					stats.GamesPlayed = value
				case "GS":
					stats.GamesStarted = value
				case "YDS":
					if category.Name == "passing" {
						stats.PassingYards = value
					} else if category.Name == "rushing" {
						stats.RushingYards = value
					} else if category.Name == "receiving" {
						stats.ReceivingYards = value
					}
				case "TD":
					if category.Name == "passing" {
						stats.PassingTDs = value
					} else if category.Name == "rushing" {
						stats.RushingTDs = value
					} else if category.Name == "receiving" {
						stats.ReceivingTDs = value
					}
				case "INT":
					stats.PassingInts = value
				case "REC":
					stats.Receptions = value
				case "TAR":
					stats.ReceivingTargets = value
				case "TAC":
					stats.Tackles = value
				case "SACK":
					if f, ok := stat.Value.(float64); ok {
						sacks := f
						stats.Sacks = &sacks
					}
				}
			}
		}

		// Only insert if player had activity that season
		if stats.GamesPlayed > 0 {
			if err := careerQueries.UpsertPlayerCareerStats(ctx, stats); err != nil {
				log.Printf("Failed to upsert career stats for season %d: %v", stats.Season, err)
			} else {
				log.Printf("Synced stats for season %d: %d games", stats.Season, stats.GamesPlayed)
			}
		}
	}

	log.Printf("Career stats sync completed for player %s", playerID)
	return nil
}

// SyncHistoricalGames fetches all games for a specific season
func (s *Service) SyncHistoricalGames(ctx context.Context, season int) error {
	log.Printf("Starting historical games sync for season %d...", season)

	totalGames := 0

	// Sync each week of the season (18 weeks in regular season)
	for week := 1; week <= 18; week++ {
		log.Printf("Syncing week %d of season %d...", week, season)

		scoreboard, err := s.espnClient.FetchSeasonGames(ctx, season, week)
		if err != nil {
			log.Printf("Failed to fetch games for season %d week %d: %v", season, week, err)
			continue
		}

		for _, event := range scoreboard.Events {
			if err := s.upsertGame(ctx, event); err != nil {
				log.Printf("Failed to upsert game %s: %v", event.Name, err)
			} else {
				totalGames++
			}
		}

		// Rate limiting
		time.Sleep(2 * time.Second)
	}

	log.Printf("Historical games sync completed for season %d: %d games", season, totalGames)
	return nil
}

// SyncMultipleSeasons syncs historical games for multiple seasons
func (s *Service) SyncMultipleSeasons(ctx context.Context, startYear, endYear int) error {
	log.Printf("Starting multi-season sync from %d to %d...", startYear, endYear)

	for year := startYear; year <= endYear; year++ {
		if err := s.SyncHistoricalGames(ctx, year); err != nil {
			log.Printf("Failed to sync season %d: %v", year, err)
		}
		// Longer delay between seasons
		time.Sleep(5 * time.Second)
	}

	log.Println("Multi-season sync completed")
	return nil
}

// SyncNFLversePlayerStats syncs enriched player statistics from nflverse
func (s *Service) SyncNFLversePlayerStats(ctx context.Context, season int) error {
	log.Printf("Starting nflverse player stats sync for season %d...", season)

	stats, err := s.nflverseClient.FetchPlayerStats(ctx, season)
	if err != nil {
		return fmt.Errorf("failed to fetch nflverse player stats: %w", err)
	}

	log.Printf("Fetched %d player stat records from nflverse", len(stats))

	careerQueries := &db.CareerQueries{}
	updatedPlayers := 0

	for _, stat := range stats {
		// Find player by nflverse player_id (need to map to our UUID)
		playerID, err := s.findPlayerByExternalID(ctx, stat.PlayerID)
		if err != nil {
			continue // Player not in our system yet
		}

		// Upsert career stats with nflverse data
		careerStats := &models.PlayerCareerStats{
			PlayerID:     playerID,
			Season:       stat.Season,
			GamesPlayed:  0, // nflverse doesn't track this directly
			PassingYards: int(stat.PassingYards),
			PassingTDs:   stat.PassingTDs,
			PassingInts:  stat.Interceptions,
			RushingYards: int(stat.RushingYards),
			RushingTDs:   stat.RushingTDs,
			Receptions:   stat.Receptions,
			ReceivingYards: int(stat.ReceivingYards),
			ReceivingTDs: stat.ReceivingTDs,
		}

		if err := careerQueries.UpsertPlayerCareerStats(ctx, careerStats); err != nil {
			log.Printf("Failed to upsert career stats for player %s: %v", stat.PlayerName, err)
			continue
		}
		updatedPlayers++
	}

	log.Printf("Successfully updated %d player career stats from nflverse", updatedPlayers)
	return nil
}

// SyncNFLverseSchedule syncs enriched schedule data including weather and venue details
func (s *Service) SyncNFLverseSchedule(ctx context.Context, season int) error {
	log.Printf("Starting nflverse schedule sync for season %d...", season)

	schedules, err := s.nflverseClient.FetchSchedule(ctx, season)
	if err != nil {
		return fmt.Errorf("failed to fetch nflverse schedule: %w", err)
	}

	log.Printf("Fetched %d games from nflverse schedule", len(schedules))

	updatedGames := 0

	for _, sched := range schedules {
		// Update existing games with enhanced data
		query := `
			UPDATE games
			SET
				weather_temp = $1,
				weather_wind_speed = $2,
				venue_name = $3
			WHERE game_id = $4
		`

		_, err := s.dbPool.Exec(ctx, query,
			sched.Temp,
			sched.Wind,
			sched.Stadium,
			sched.GameID,
		)

		if err != nil {
			log.Printf("Failed to update game %s: %v", sched.GameID, err)
			continue
		}
		updatedGames++
	}

	log.Printf("Successfully updated %d games with nflverse schedule data", updatedGames)
	return nil
}

// SyncNFLverseNextGenStats syncs Next Gen Stats for advanced metrics
func (s *Service) SyncNFLverseNextGenStats(ctx context.Context, season int, statType string) error {
	log.Printf("Starting nflverse Next Gen Stats sync for season %d, type: %s...", season, statType)

	ngsStats, err := s.nflverseClient.FetchNextGenStats(ctx, season, statType)
	if err != nil {
		return fmt.Errorf("failed to fetch Next Gen Stats: %w", err)
	}

	log.Printf("Fetched %d Next Gen Stats records", len(ngsStats))

	// Store NGS data in a dedicated table or extend player_career_stats
	// For now, we'll log the availability
	for _, ngs := range ngsStats {
		log.Printf("Player: %s, Team: %s, Season: %d, Week: %d",
			ngs.PlayerName, ngs.TeamAbbr, ngs.Season, ngs.Week)
	}

	log.Println("Next Gen Stats sync completed (logging only - storage TBD)")
	return nil
}

// findPlayerByExternalID finds a player in our database by their external ID (ESPN, nflverse, etc.)
func (s *Service) findPlayerByExternalID(ctx context.Context, externalID string) (uuid.UUID, error) {
	var playerID uuid.UUID
	query := `
		SELECT id FROM players
		WHERE espn_athlete_id = $1
		LIMIT 1
	`
	err := s.dbPool.QueryRow(ctx, query, externalID).Scan(&playerID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("player not found: %w", err)
	}
	return playerID, nil
}

// EnrichGamesWithWeather enriches games with weather data from WeatherAPI.com
func (s *Service) EnrichGamesWithWeather(ctx context.Context, season int) error {
	if s.weatherClient == nil {
		return fmt.Errorf("weather client not initialized (API key missing)")
	}

	log.Printf("Starting weather enrichment for season %d...", season)

	// Get all games for the season that have venue coordinates
	query := `
		SELECT g.id, g.game_date, t.stadium_latitude, t.stadium_longitude, t.city, t.state
		FROM games g
		JOIN teams t ON g.home_team_id = t.id
		WHERE g.season = $1
		AND t.stadium_latitude IS NOT NULL
		AND t.stadium_longitude IS NOT NULL
		AND g.weather_temp IS NULL
		ORDER BY g.game_date
	`

	rows, err := s.dbPool.Query(ctx, query, season)
	if err != nil {
		return fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var gameID uuid.UUID
		var gameDate time.Time
		var lat, lon float64
		var city, state string

		if err := rows.Scan(&gameID, &gameDate, &lat, &lon, &city, &state); err != nil {
			log.Printf("Failed to scan game row: %v", err)
			continue
		}

		// Format date for WeatherAPI (YYYY-MM-DD)
		dateStr := gameDate.Format("2006-01-02")

		// Get historical weather data
		weatherData, err := s.weatherClient.GetHistoricalWeatherByCoordinates(ctx, lat, lon, dateStr)
		if err != nil {
			log.Printf("Failed to fetch weather for game %s on %s: %v", gameID, dateStr, err)
			time.Sleep(1 * time.Second) // Rate limiting
			continue
		}

		// Determine if day game (kickoff before 5pm local time)
		isDayGame := gameDate.Hour() < 17

		// Update game with comprehensive weather data
		updateQuery := `
			UPDATE games
			SET weather_temp = $1,
			    weather_condition = $2,
			    weather_wind_speed = $3,
			    weather_humidity = $4,
			    weather_wind_dir = $5,
			    weather_pressure = $6,
			    weather_visibility = $7,
			    weather_feels_like = $8,
			    weather_precipitation = $9,
			    weather_cloud_cover = $10,
			    is_day_game = $11
			WHERE id = $12
		`

		_, err = s.dbPool.Exec(ctx, updateQuery,
			int(weatherData.Day.AvgTempF),
			weatherData.Day.Condition.Text,
			int(weatherData.Day.MaxWindMPH),
			int(weatherData.Day.AvgHumidity),
			"", // Wind direction not available in historical day data
			int(weatherData.Day.AvgPressureMb),
			int(weatherData.Day.AvgVisMiles),
			int(weatherData.Day.AvgTempF), // Use avg temp as feels-like for historical
			weatherData.Day.TotalPrecipIn,
			int(weatherData.Day.AvgCloud),
			isDayGame,
			gameID,
		)

		if err != nil {
			log.Printf("Failed to update game %s with weather: %v", gameID, err)
			continue
		}

		count++
		log.Printf("Enriched game %s (%s, %s %s) with weather: %s, %.0f°F, %dmph wind, %d%% humidity",
			gameID, dateStr, city, state,
			weatherData.Day.Condition.Text, weatherData.Day.AvgTempF,
			int(weatherData.Day.MaxWindMPH), int(weatherData.Day.AvgHumidity))

		// Rate limiting - WeatherAPI free tier allows 1M calls/month
		// Sleep for 500ms between requests to be respectful
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Weather enrichment completed: %d games updated", count)
	return nil
}

// SyncGameTeamStats syncs team statistics for completed games from ESPN box scores
func (s *Service) SyncGameTeamStats(ctx context.Context, season int, week int) error {
	log.Printf("Starting team stats sync for season %d, week %d...", season, week)

	// Fetch completed games for the given season/week
	query := `
		SELECT id, nfl_game_id, home_team_id, away_team_id
		FROM games
		WHERE season = $1 AND week = $2 AND status = 'completed'
		ORDER BY game_date
	`

	rows, err := s.dbPool.Query(ctx, query, season, week)
	if err != nil {
		return fmt.Errorf("failed to fetch games: %w", err)
	}
	defer rows.Close()

	type gameInfo struct {
		ID         uuid.UUID
		NFLGameID  string
		HomeTeamID uuid.UUID
		AwayTeamID uuid.UUID
	}

	var games []gameInfo
	for rows.Next() {
		var g gameInfo
		if err := rows.Scan(&g.ID, &g.NFLGameID, &g.HomeTeamID, &g.AwayTeamID); err != nil {
			log.Printf("Error scanning game row: %v", err)
			continue
		}
		games = append(games, g)
	}

	if len(games) == 0 {
		log.Printf("No completed games found for season %d, week %d", season, week)
		return nil
	}

	log.Printf("Found %d completed games to sync stats for", len(games))

	count := 0
	var failedFetch, failedBoxScore, failedTeamLookup, failedInsert int

	for _, game := range games {
		log.Printf("Processing game %s (ID: %s)", game.NFLGameID, game.ID)

		// Fetch game details with box score
		gameDetail, err := s.espnClient.FetchGameDetails(ctx, game.NFLGameID)
		if err != nil {
			log.Printf("ERROR: Failed to fetch game details for game %s (NFL ID: %s): %v", game.ID, game.NFLGameID, err)
			log.Printf("ERROR: ESPN API error type: %T", err)
			failedFetch++
			// Return error on first fetch failure to see what's wrong
			return fmt.Errorf("ESPN API fetch failed for game %s: %w", game.NFLGameID, err)
		}

		// Process box score for both teams
		if gameDetail.BoxScore.Teams == nil || len(gameDetail.BoxScore.Teams) < 2 {
			log.Printf("No box score data for game %s", game.NFLGameID)
			failedBoxScore++
			continue
		}

		log.Printf("Found %d teams in box score for game %s", len(gameDetail.BoxScore.Teams), game.NFLGameID)

		for _, teamStats := range gameDetail.BoxScore.Teams {
			// Determine which team this is
			espnTeamID := teamStats.Team.ID
			var teamID uuid.UUID

			// Match ESPN team ID to our team ID (convert string to int)
			var dbTeamID uuid.UUID
			teamQuery := `SELECT id FROM teams WHERE nfl_id::text = $1`
			if err := s.dbPool.QueryRow(ctx, teamQuery, espnTeamID).Scan(&dbTeamID); err != nil {
				log.Printf("ERROR: Could not find team with ESPN ID %s: %v", espnTeamID, err)
				log.Printf("ERROR: This is blocking all team stats sync! Check teams table.")
				return fmt.Errorf("team lookup failed for ESPN ID %s: %w", espnTeamID, err)
			}
			teamID = dbTeamID

			// Parse statistics into a map for easy access
			stats := make(map[string]float64)
			statsDisplay := make(map[string]string)
			for _, stat := range teamStats.Statistics {
				statsDisplay[stat.Name] = stat.DisplayValue

				// Special handling for X-Y format stats
				if stat.Name == "thirdDownEff" || stat.Name == "fourthDownEff" ||
					stat.Name == "redZoneAttempts" || stat.Name == "completionAttempts" ||
					stat.Name == "sacksYardsLost" || stat.Name == "totalPenaltiesYards" {
					// Parse X-Y format for efficiency stats
					parts := strings.Split(stat.DisplayValue, "-")
					if len(parts) == 2 {
						made, _ := strconv.ParseFloat(parts[0], 64)
						att, _ := strconv.ParseFloat(parts[1], 64)
						stats[stat.Name+"_made"] = made
						stats[stat.Name+"_att"] = att
					}
				} else if stat.Value != nil {
					// Use the Value field if available (it's already a number)
					switch v := stat.Value.(type) {
					case float64:
						stats[stat.Name] = v
					case string:
						// Skip string values like "-"
						continue
					}
				}
			}

			// Parse possession time
			possessionTime := statsDisplay["possessionTime"]
			var possessionSeconds int
			if possessionTime != "" {
				parts := strings.Split(possessionTime, ":")
				if len(parts) == 2 {
					minutes, _ := strconv.Atoi(parts[0])
					seconds, _ := strconv.Atoi(parts[1])
					possessionSeconds = (minutes * 60) + seconds
				}
			}

			// Calculate percentages
			var thirdDownPct, fourthDownPct float64
			if stats["thirdDownEff_att"] > 0 {
				thirdDownPct = (stats["thirdDownEff_made"] / stats["thirdDownEff_att"]) * 100
			}
			if stats["fourthDownEff_att"] > 0 {
				fourthDownPct = (stats["fourthDownEff_made"] / stats["fourthDownEff_att"]) * 100
			}

			// Debug: log parsed stats
			log.Printf("Parsed %d stats for ESPN team %s. Sample: totalYards=%.0f, firstDowns=%.0f",
				len(stats), espnTeamID, stats["totalYards"], stats["firstDowns"])

			// Upsert team stats
			log.Printf("Inserting stats for team %s (ESPN ID: %s) in game %s",
				teamID, espnTeamID, game.ID)

			insertQuery := `
				INSERT INTO game_team_stats (
					game_id, team_id,
					first_downs, total_yards, passing_yards, rushing_yards,
					offensive_plays, yards_per_play,
					third_down_attempts, third_down_conversions, third_down_pct,
					fourth_down_attempts, fourth_down_conversions, fourth_down_pct,
					red_zone_attempts, red_zone_scores,
					turnovers, fumbles_lost, interceptions_thrown,
					penalties, penalty_yards,
					possession_time, possession_seconds,
					completions, pass_attempts,
					sacks_allowed, sack_yards,
					rushing_attempts, rushing_avg
				) VALUES (
					$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
					$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29
				)
				ON CONFLICT (game_id, team_id)
				DO UPDATE SET
					first_downs = EXCLUDED.first_downs,
					total_yards = EXCLUDED.total_yards,
					passing_yards = EXCLUDED.passing_yards,
					rushing_yards = EXCLUDED.rushing_yards,
					offensive_plays = EXCLUDED.offensive_plays,
					yards_per_play = EXCLUDED.yards_per_play,
					third_down_attempts = EXCLUDED.third_down_attempts,
					third_down_conversions = EXCLUDED.third_down_conversions,
					third_down_pct = EXCLUDED.third_down_pct,
					fourth_down_attempts = EXCLUDED.fourth_down_attempts,
					fourth_down_conversions = EXCLUDED.fourth_down_conversions,
					fourth_down_pct = EXCLUDED.fourth_down_pct,
					red_zone_attempts = EXCLUDED.red_zone_attempts,
					red_zone_scores = EXCLUDED.red_zone_scores,
					turnovers = EXCLUDED.turnovers,
					fumbles_lost = EXCLUDED.fumbles_lost,
					interceptions_thrown = EXCLUDED.interceptions_thrown,
					penalties = EXCLUDED.penalties,
					penalty_yards = EXCLUDED.penalty_yards,
					possession_time = EXCLUDED.possession_time,
					possession_seconds = EXCLUDED.possession_seconds,
					completions = EXCLUDED.completions,
					pass_attempts = EXCLUDED.pass_attempts,
					sacks_allowed = EXCLUDED.sacks_allowed,
					sack_yards = EXCLUDED.sack_yards,
					rushing_attempts = EXCLUDED.rushing_attempts,
					rushing_avg = EXCLUDED.rushing_avg
			`

			result, err := s.dbPool.Exec(ctx, insertQuery,
				game.ID, teamID,
				int(stats["firstDowns"]), int(stats["totalYards"]),
				int(stats["netPassingYards"]), int(stats["rushingYards"]),
				int(stats["totalOffensivePlays"]), stats["yardsPerPlay"],
				int(stats["thirdDownEff_att"]), int(stats["thirdDownEff_made"]), thirdDownPct,
				int(stats["fourthDownEff_att"]), int(stats["fourthDownEff_made"]), fourthDownPct,
				int(stats["redZoneAttempts_att"]), int(stats["redZoneAttempts_made"]),
				int(stats["turnovers"]), int(stats["fumblesLost"]), int(stats["interceptions"]),
				int(stats["totalPenaltiesYards_att"]), int(stats["totalPenaltiesYards_made"]),
				possessionTime, possessionSeconds,
				int(stats["completionAttempts_made"]), int(stats["completionAttempts_att"]),
				int(stats["sacksYardsLost_att"]), int(stats["sacksYardsLost_made"]),
				int(stats["rushingAttempts"]), stats["yardsPerRushAttempt"],
			)

			if err != nil {
				log.Printf("Failed to insert stats for team %s in game %s: %v", teamID, game.ID, err)
				failedInsert++
				continue
			}

			rowsAffected := result.RowsAffected()
			log.Printf("✓ Synced stats for team %s in game %s (NFL ID: %s) - %d rows affected",
				teamID, game.ID, game.NFLGameID, rowsAffected)
			count++
		}

		// Rate limiting - be respectful to ESPN API
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("========================================")
	log.Printf("Team stats sync completed: %d team stats records created/updated out of %d games", count, len(games))
	log.Printf("Failures: %d fetch errors, %d missing box scores, %d team lookup errors, %d insert errors",
		failedFetch, failedBoxScore, failedTeamLookup, failedInsert)
	log.Printf("========================================")

	if count == 0 {
		return fmt.Errorf("no team stats were synced - processed %d games, failures: fetch=%d boxscore=%d teamlookup=%d insert=%d",
			len(games), failedFetch, failedBoxScore, failedTeamLookup, failedInsert)
	}

	return nil
}
