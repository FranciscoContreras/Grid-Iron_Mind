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

	espnGameID, _ := strconv.Atoi(event.ID)
	homeScore, _ := strconv.Atoi(homeTeam.Score)
	awayScore, _ := strconv.Atoi(awayTeam.Score)

	status := "scheduled"
	if event.Status.Type.Completed {
		status = "completed"
	} else if event.Status.Type.State == "in" {
		status = "in_progress"
	}

	// Check if game exists
	var existingID uuid.UUID
	err = s.dbPool.QueryRow(ctx,
		"SELECT id FROM games WHERE espn_game_id = $1",
		espnGameID,
	).Scan(&existingID)

	now := time.Now()
	if err != nil {
		// Game doesn't exist, insert
		id := uuid.New()
		_, err = s.dbPool.Exec(ctx,
			`INSERT INTO games (id, espn_game_id, season_year, season_type, week, game_date, home_team_id, away_team_id, home_score, away_score, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			id, espnGameID, event.Season.Year, event.Season.Type.Type, event.Week.Number,
			event.Date.Time, homeTeamID, awayTeamID, homeScore, awayScore, status, now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert game: %w", err)
		}
		log.Printf("Inserted game: %s", event.Name)
	} else {
		// Game exists, update scores and status
		_, err = s.dbPool.Exec(ctx,
			`UPDATE games
			SET home_score = $1, away_score = $2, status = $3, updated_at = $4
			WHERE id = $5`,
			homeScore, awayScore, status, now, existingID,
		)
		if err != nil {
			return fmt.Errorf("failed to update game: %w", err)
		}
		log.Printf("Updated game: %s (%d-%d)", event.Name, homeScore, awayScore)
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

		// Update game with weather data
		updateQuery := `
			UPDATE games
			SET weather_temp = $1,
			    weather_condition = $2,
			    weather_wind_speed = $3,
			    weather_humidity = $4
			WHERE id = $5
		`

		_, err = s.dbPool.Exec(ctx, updateQuery,
			int(weatherData.Day.AvgTempF),
			weatherData.Day.Condition.Text,
			int(weatherData.Day.MaxWindMPH),
			int(weatherData.Day.AvgHumidity),
			gameID,
		)

		if err != nil {
			log.Printf("Failed to update game %s with weather: %v", gameID, err)
			continue
		}

		count++
		log.Printf("Enriched game %s (%s, %s %s) with weather: %s, %.0f°F",
			gameID, dateStr, city, state,
			weatherData.Day.Condition.Text, weatherData.Day.AvgTempF)

		// Rate limiting - WeatherAPI free tier allows 1M calls/month
		// Sleep for 500ms between requests to be respectful
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Weather enrichment completed: %d games updated", count)
	return nil
}
