package ingestion

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// SyncPlayByPlay syncs play-by-play data from NFLverse for a given season and week
//
// NFLverse provides comprehensive play-by-play data with:
//   - Every play in every game
//   - Player involvement (passer, rusher, receiver)
//   - Advanced metrics (EPA, WPA, success rate)
//   - Play outcomes (first downs, TDs, turnovers)
//
// This function:
//   - Downloads CSV from NFLverse GitHub releases
//   - Parses play-by-play data
//   - Matches players and teams by ID
//   - Stores in play_by_play table
//   - Updates materialized views
//
// Example usage:
//
//	if err := service.SyncPlayByPlay(ctx, 2024, 1); err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) SyncPlayByPlay(ctx context.Context, season, week int) error {
	log.Printf("Syncing play-by-play data for season %d week %d...", season, week)

	// NFLverse play-by-play URL
	url := fmt.Sprintf("https://github.com/nflverse/nflverse-data/releases/download/pbp/play_by_play_%d.csv", season)

	log.Printf("Fetching from: %s", url)

	// Download CSV
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NFLverse returned status %d for %s", resp.StatusCode, url)
	}

	// Parse CSV
	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = -1 // Allow variable fields

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create column index map
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	log.Printf("Found %d columns in CSV", len(header))

	// Cache team and game lookups
	teamCache := make(map[string]uuid.UUID)
	gameCache := make(map[string]uuid.UUID)

	// Track stats
	totalRows := 0
	insertedCount := 0
	skippedCount := 0

	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV row: %v", err)
			continue
		}

		totalRows++

		// Filter by week
		playWeek := getCSVInt(row, colIndex, "week")
		if playWeek != week {
			continue
		}

		// Get game ID
		nflGameID := getCSVValue(row, colIndex, "game_id")
		if nflGameID == "" {
			skippedCount++
			continue
		}

		// Find game in database
		gameID, found := gameCache[nflGameID]
		if !found {
			var tempID uuid.UUID
			err := s.dbPool.QueryRow(ctx, `
				SELECT id FROM games WHERE nfl_game_id = $1
			`, nflGameID).Scan(&tempID)

			if err != nil {
				log.Printf("Game not found for nfl_game_id %s", nflGameID)
				skippedCount++
				continue
			}
			gameID = tempID
			gameCache[nflGameID] = gameID
		}

		// Get team IDs
		homeTeam := getCSVValue(row, colIndex, "home_team")
		awayTeam := getCSVValue(row, colIndex, "away_team")
		possTeam := getCSVValue(row, colIndex, "posteam")
		defTeam := getCSVValue(row, colIndex, "defteam")

		homeTeamID := s.findTeamID(ctx, homeTeam, teamCache)
		awayTeamID := s.findTeamID(ctx, awayTeam, teamCache)
		possTeamID := s.findTeamID(ctx, possTeam, teamCache)
		defTeamID := s.findTeamID(ctx, defTeam, teamCache)

		// Get player IDs
		passerGSIS := getCSVValue(row, colIndex, "passer_player_id")
		receiverGSIS := getCSVValue(row, colIndex, "receiver_player_id")
		rusherGSIS := getCSVValue(row, colIndex, "rusher_player_id")

		var passerID, receiverID, rusherID *uuid.UUID

		if passerGSIS != "" {
			passerID = s.findPlayerByGSISOrName(ctx, passerGSIS, getCSVValue(row, colIndex, "passer_player_name"))
		}
		if receiverGSIS != "" {
			receiverID = s.findPlayerByGSISOrName(ctx, receiverGSIS, getCSVValue(row, colIndex, "receiver_player_name"))
		}
		if rusherGSIS != "" {
			rusherID = s.findPlayerByGSISOrName(ctx, rusherGSIS, getCSVValue(row, colIndex, "rusher_player_name"))
		}

		// Insert play
		if err := s.insertPlay(ctx, gameID, nflGameID, season, week, row, colIndex,
			homeTeamID, awayTeamID, possTeamID, defTeamID,
			passerID, receiverID, rusherID); err != nil {
			log.Printf("Failed to insert play: %v", err)
			skippedCount++
			continue
		}

		insertedCount++
	}

	log.Printf("Play-by-play sync complete: %d rows processed, %d inserted, %d skipped",
		totalRows, insertedCount, skippedCount)

	// Refresh materialized views
	log.Printf("Refreshing materialized views...")
	if err := s.refreshPlayByPlayViews(ctx); err != nil {
		log.Printf("Failed to refresh views: %v", err)
	}

	return nil
}

// insertPlay inserts a single play into the database
func (s *Service) insertPlay(ctx context.Context, gameID uuid.UUID, nflGameID string,
	season, week int, row []string, colIndex map[string]int,
	homeTeamID, awayTeamID, possTeamID, defTeamID *uuid.UUID,
	passerID, receiverID, rusherID *uuid.UUID) error {

	playID := getCSVValue(row, colIndex, "play_id")

	query := `
		INSERT INTO play_by_play (
			play_id, game_id, nfl_game_id,
			home_team_id, away_team_id, possession_team_id, defensive_team_id,
			season, week, quarter, down, yards_to_go, yard_line, game_seconds_remaining,
			play_type, play_type_nfl, description, yards_gained,
			passer_player_id, receiver_player_id, rusher_player_id,
			pass_length, pass_location, air_yards, yards_after_catch,
			run_location, run_gap,
			epa, wpa, success_play,
			first_down, touchdown, pass_touchdown, rush_touchdown,
			interception, fumble, completed_pass, sack, penalty,
			possession_team_score, defensive_team_score,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
			$15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27,
			$28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39,
			$40, $41, NOW()
		)
		ON CONFLICT (play_id, game_id)
		DO UPDATE SET
			quarter = EXCLUDED.quarter,
			down = EXCLUDED.down,
			yards_to_go = EXCLUDED.yards_to_go,
			yard_line = EXCLUDED.yard_line,
			play_type = EXCLUDED.play_type,
			description = EXCLUDED.description,
			yards_gained = EXCLUDED.yards_gained,
			epa = EXCLUDED.epa,
			wpa = EXCLUDED.wpa,
			success_play = EXCLUDED.success_play,
			first_down = EXCLUDED.first_down,
			touchdown = EXCLUDED.touchdown,
			pass_touchdown = EXCLUDED.pass_touchdown,
			rush_touchdown = EXCLUDED.rush_touchdown,
			interception = EXCLUDED.interception,
			fumble = EXCLUDED.fumble,
			completed_pass = EXCLUDED.completed_pass,
			sack = EXCLUDED.sack,
			penalty = EXCLUDED.penalty,
			possession_team_score = EXCLUDED.possession_team_score,
			defensive_team_score = EXCLUDED.defensive_team_score,
			updated_at = NOW()
	`

	_, err := s.dbPool.Exec(ctx, query,
		playID, gameID, nflGameID,
		homeTeamID, awayTeamID, possTeamID, defTeamID,
		season, week,
		getCSVInt(row, colIndex, "qtr"),
		getCSVInt(row, colIndex, "down"),
		getCSVInt(row, colIndex, "ydstogo"),
		getCSVInt(row, colIndex, "yardline_100"),
		getCSVInt(row, colIndex, "game_seconds_remaining"),
		getCSVValue(row, colIndex, "play_type"),
		getCSVValue(row, colIndex, "play_type_nfl"),
		getCSVValue(row, colIndex, "desc"),
		getCSVInt(row, colIndex, "yards_gained"),
		passerID, receiverID, rusherID,
		getCSVValue(row, colIndex, "pass_length"),
		getCSVValue(row, colIndex, "pass_location"),
		getCSVFloat(row, colIndex, "air_yards"),
		getCSVFloat(row, colIndex, "yards_after_catch"),
		getCSVValue(row, colIndex, "run_location"),
		getCSVValue(row, colIndex, "run_gap"),
		getCSVFloat(row, colIndex, "epa"),
		getCSVFloat(row, colIndex, "wpa"),
		getCSVInt(row, colIndex, "success"),
		getCSVInt(row, colIndex, "first_down"),
		getCSVInt(row, colIndex, "touchdown"),
		getCSVInt(row, colIndex, "pass_touchdown"),
		getCSVInt(row, colIndex, "rush_touchdown"),
		getCSVInt(row, colIndex, "interception"),
		getCSVInt(row, colIndex, "fumble"),
		getCSVInt(row, colIndex, "complete_pass"),
		getCSVInt(row, colIndex, "sack"),
		getCSVInt(row, colIndex, "penalty"),
		getCSVInt(row, colIndex, "posteam_score_post"),
		getCSVInt(row, colIndex, "defteam_score_post"),
	)

	return err
}

// findTeamID finds a team ID by abbreviation with caching
func (s *Service) findTeamID(ctx context.Context, abbr string, cache map[string]uuid.UUID) *uuid.UUID {
	if abbr == "" {
		return nil
	}

	if id, found := cache[abbr]; found {
		return &id
	}

	var teamID uuid.UUID
	err := s.dbPool.QueryRow(ctx, `
		SELECT id FROM teams WHERE abbreviation = $1
	`, abbr).Scan(&teamID)

	if err == nil {
		cache[abbr] = teamID
		return &teamID
	}

	return nil
}

// findPlayerByGSISOrName finds a player by GSIS ID or name
func (s *Service) findPlayerByGSISOrName(ctx context.Context, gsisID, name string) *uuid.UUID {
	// TODO: Add gsis_id column to players table for exact matching
	// For now, use name matching

	if name == "" {
		return nil
	}

	var playerID uuid.UUID

	// Try exact match
	err := s.dbPool.QueryRow(ctx, `
		SELECT id FROM players
		WHERE LOWER(name) = LOWER($1)
		LIMIT 1
	`, name).Scan(&playerID)

	if err == nil {
		return &playerID
	}

	// Try fuzzy match by last name
	parts := strings.Split(name, " ")
	if len(parts) < 2 {
		return nil
	}
	lastName := parts[len(parts)-1]

	err = s.dbPool.QueryRow(ctx, `
		SELECT id FROM players
		WHERE LOWER(name) LIKE LOWER($1)
		LIMIT 1
	`, "%"+lastName+"%").Scan(&playerID)

	if err == nil {
		return &playerID
	}

	return nil
}

// refreshPlayByPlayViews refreshes materialized views
func (s *Service) refreshPlayByPlayViews(ctx context.Context) error {
	log.Printf("Refreshing game_play_summary view...")
	if _, err := s.dbPool.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY game_play_summary"); err != nil {
		return fmt.Errorf("failed to refresh game_play_summary: %w", err)
	}

	log.Printf("Refreshing player_play_stats view...")
	if _, err := s.dbPool.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY player_play_stats"); err != nil {
		return fmt.Errorf("failed to refresh player_play_stats: %w", err)
	}

	log.Printf("Materialized views refreshed successfully")
	return nil
}

// SyncPlayByPlaySeason syncs play-by-play data for all weeks in a season
func (s *Service) SyncPlayByPlaySeason(ctx context.Context, season int) error {
	log.Printf("Syncing play-by-play for entire season %d...", season)

	// Determine max week with games
	var maxWeek int
	err := s.dbPool.QueryRow(ctx, `
		SELECT COALESCE(MAX(week), 0)
		FROM games
		WHERE season = $1
	`, season).Scan(&maxWeek)

	if err != nil {
		return fmt.Errorf("failed to get max week: %w", err)
	}

	if maxWeek == 0 {
		return fmt.Errorf("no games found for season %d", season)
	}

	log.Printf("Found games through week %d", maxWeek)

	// Sync each week
	for week := 1; week <= maxWeek; week++ {
		log.Printf("Syncing week %d/%d...", week, maxWeek)
		if err := s.SyncPlayByPlay(ctx, season, week); err != nil {
			log.Printf("Failed to sync week %d: %v", week, err)
			// Continue with next week
			continue
		}
	}

	log.Printf("Play-by-play sync complete for season %d", season)
	return nil
}

// SyncPlayByPlayGame syncs play-by-play data for a specific game
func (s *Service) SyncPlayByPlayGame(ctx context.Context, gameID uuid.UUID) error {
	log.Printf("Syncing play-by-play for game %s...", gameID)

	// Get game details
	var season, week int
	var nflGameID string

	err := s.dbPool.QueryRow(ctx, `
		SELECT season, week, nfl_game_id
		FROM games
		WHERE id = $1
	`, gameID).Scan(&season, &week, &nflGameID)

	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Sync the week (which includes this game)
	if err := s.SyncPlayByPlay(ctx, season, week); err != nil {
		return fmt.Errorf("failed to sync play-by-play: %w", err)
	}

	log.Printf("Play-by-play synced for game %s", gameID)
	return nil
}
