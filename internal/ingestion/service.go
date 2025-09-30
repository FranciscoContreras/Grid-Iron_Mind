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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles data ingestion from ESPN API
type Service struct {
	espnClient *espn.Client
	dbPool     *pgxpool.Pool
}

// NewService creates a new ingestion service
func NewService() *Service {
	return &Service{
		espnClient: espn.NewClient(),
		dbPool:     db.GetPool(),
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
			id, espnGameID, event.Season.Year, event.Season.Type, event.Week.Number,
			event.Date, homeTeamID, awayTeamID, homeScore, awayScore, status, now, now,
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