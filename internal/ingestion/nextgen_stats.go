package ingestion

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// SyncNextGenStats syncs Next Gen Stats from NFLverse for a given season
//
// NFLverse provides Next Gen Stats in three categories:
//   - Passing NGS
//   - Rushing NGS
//   - Receiving NGS
//
// This function:
//   - Downloads CSV files from NFLverse GitHub releases
//   - Parses Next Gen Stats for passing, rushing, and receiving
//   - Matches players by name and team
//   - Stores in advanced_stats table with stat_type field
//   - Supports both weekly and season-aggregated stats
//
// Example usage:
//
//	if err := service.SyncNextGenStats(ctx, 2024, "passing"); err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) SyncNextGenStats(ctx context.Context, season int, statType string) error {
	log.Printf("Syncing Next Gen Stats (%s) for season %d...", statType, season)

	// Validate stat type
	validTypes := map[string]bool{
		"passing":   true,
		"rushing":   true,
		"receiving": true,
	}

	if !validTypes[statType] {
		return fmt.Errorf("invalid stat type: %s (must be passing, rushing, or receiving)", statType)
	}

	// NFLverse Next Gen Stats URLs
	url := fmt.Sprintf("https://github.com/nflverse/nflverse-data/releases/download/nextgen_stats/ngs_%s_%d.csv", statType, season)

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

		// Extract common fields
		seasonTypeStr := getCSVValue(row, colIndex, "season_type")
		if seasonTypeStr != "REG" {
			skippedCount++
			continue // Only regular season
		}

		playerName := getCSVValue(row, colIndex, "player_display_name")
		if playerName == "" {
			playerName = getCSVValue(row, colIndex, "player_name")
		}
		if playerName == "" {
			log.Printf("Skipping row with no player name")
			skippedCount++
			continue
		}

		teamAbbr := getCSVValue(row, colIndex, "team_abbr")
		if teamAbbr == "" {
			log.Printf("Skipping %s with no team", playerName)
			skippedCount++
			continue
		}

		// Get week (NULL = season totals)
		weekStr := getCSVValue(row, colIndex, "week")
		var week *int
		if weekStr != "" && weekStr != "0" {
			w, err := strconv.Atoi(weekStr)
			if err == nil {
				week = &w
			}
		}

		// Find player
		playerID := s.findPlayerByNameAndTeam(ctx, playerName, teamAbbr, season)
		if playerID == nil {
			log.Printf("Could not find player: %s (%s)", playerName, teamAbbr)
			skippedCount++
			continue
		}

		// Insert based on stat type
		var insertErr error
		switch statType {
		case "passing":
			insertErr = s.insertPassingNGS(ctx, *playerID, season, week, row, colIndex)
		case "rushing":
			insertErr = s.insertRushingNGS(ctx, *playerID, season, week, row, colIndex)
		case "receiving":
			insertErr = s.insertReceivingNGS(ctx, *playerID, season, week, row, colIndex)
		}

		if insertErr != nil {
			log.Printf("Failed to insert NGS for %s: %v", playerName, insertErr)
			skippedCount++
			continue
		}

		insertedCount++
	}

	log.Printf("Next Gen Stats (%s) sync complete: %d rows processed, %d inserted, %d skipped",
		statType, totalRows, insertedCount, skippedCount)

	return nil
}

// insertPassingNGS inserts passing Next Gen Stats
func (s *Service) insertPassingNGS(ctx context.Context, playerID uuid.UUID, season int, week *int, row []string, colIndex map[string]int) error {
	query := `
		INSERT INTO advanced_stats (
			player_id, season, week, stat_type,
			avg_time_to_throw, avg_completed_air_yards, avg_intended_air_yards,
			avg_air_yards_differential, max_completed_air_distance, avg_air_yards_to_sticks,
			attempts, pass_yards, pass_touchdowns, interceptions,
			completions, completion_percentage,
			expected_completion_percentage, completion_percentage_above_expectation,
			passer_rating,
			updated_at
		) VALUES (
			$1, $2, $3, 'passing',
			$4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW()
		)
		ON CONFLICT (player_id, season, week, stat_type)
		DO UPDATE SET
			avg_time_to_throw = EXCLUDED.avg_time_to_throw,
			avg_completed_air_yards = EXCLUDED.avg_completed_air_yards,
			avg_intended_air_yards = EXCLUDED.avg_intended_air_yards,
			avg_air_yards_differential = EXCLUDED.avg_air_yards_differential,
			max_completed_air_distance = EXCLUDED.max_completed_air_distance,
			avg_air_yards_to_sticks = EXCLUDED.avg_air_yards_to_sticks,
			attempts = EXCLUDED.attempts,
			pass_yards = EXCLUDED.pass_yards,
			pass_touchdowns = EXCLUDED.pass_touchdowns,
			interceptions = EXCLUDED.interceptions,
			completions = EXCLUDED.completions,
			completion_percentage = EXCLUDED.completion_percentage,
			expected_completion_percentage = EXCLUDED.expected_completion_percentage,
			completion_percentage_above_expectation = EXCLUDED.completion_percentage_above_expectation,
			passer_rating = EXCLUDED.passer_rating,
			updated_at = NOW()
	`

	_, err := s.dbPool.Exec(ctx, query,
		playerID, season, week,
		getCSVFloat(row, colIndex, "avg_time_to_throw"),
		getCSVFloat(row, colIndex, "avg_completed_air_yards"),
		getCSVFloat(row, colIndex, "avg_intended_air_yards"),
		getCSVFloat(row, colIndex, "avg_air_yards_differential"),
		getCSVInt(row, colIndex, "max_completed_air_distance"),
		getCSVFloat(row, colIndex, "avg_air_yards_to_sticks"),
		getCSVInt(row, colIndex, "attempts"),
		getCSVInt(row, colIndex, "pass_yards"),
		getCSVInt(row, colIndex, "pass_touchdowns"),
		getCSVInt(row, colIndex, "interceptions"),
		getCSVInt(row, colIndex, "completions"),
		getCSVFloat(row, colIndex, "completion_percentage"),
		getCSVFloat(row, colIndex, "xcomp_pct"), // expected completion %
		getCSVFloat(row, colIndex, "cpoe"),      // completion % over expected
		getCSVFloat(row, colIndex, "passer_rating"),
	)

	return err
}

// insertRushingNGS inserts rushing Next Gen Stats
func (s *Service) insertRushingNGS(ctx context.Context, playerID uuid.UUID, season int, week *int, row []string, colIndex map[string]int) error {
	query := `
		INSERT INTO advanced_stats (
			player_id, season, week, stat_type,
			efficiency, percent_attempts_gte_eight_defenders, avg_time_to_los,
			rush_attempts, rush_yards, expected_rush_yards, rush_yards_over_expected,
			avg_rush_yards, rush_touchdowns,
			updated_at
		) VALUES (
			$1, $2, $3, 'rushing',
			$4, $5, $6, $7, $8, $9, $10, $11, $12, NOW()
		)
		ON CONFLICT (player_id, season, week, stat_type)
		DO UPDATE SET
			efficiency = EXCLUDED.efficiency,
			percent_attempts_gte_eight_defenders = EXCLUDED.percent_attempts_gte_eight_defenders,
			avg_time_to_los = EXCLUDED.avg_time_to_los,
			rush_attempts = EXCLUDED.rush_attempts,
			rush_yards = EXCLUDED.rush_yards,
			expected_rush_yards = EXCLUDED.expected_rush_yards,
			rush_yards_over_expected = EXCLUDED.rush_yards_over_expected,
			avg_rush_yards = EXCLUDED.avg_rush_yards,
			rush_touchdowns = EXCLUDED.rush_touchdowns,
			updated_at = NOW()
	`

	_, err := s.dbPool.Exec(ctx, query,
		playerID, season, week,
		getCSVFloat(row, colIndex, "efficiency"),
		getCSVFloat(row, colIndex, "percent_attempts_gte_eight_defenders"),
		getCSVFloat(row, colIndex, "avg_time_to_los"),
		getCSVInt(row, colIndex, "rush_attempts"),
		getCSVInt(row, colIndex, "rush_yards"),
		getCSVInt(row, colIndex, "expected_rush_yards"),
		getCSVInt(row, colIndex, "rush_yards_over_expected"),
		getCSVFloat(row, colIndex, "avg_rush_yards"),
		getCSVInt(row, colIndex, "rush_touchdowns"),
	)

	return err
}

// insertReceivingNGS inserts receiving Next Gen Stats
func (s *Service) insertReceivingNGS(ctx context.Context, playerID uuid.UUID, season int, week *int, row []string, colIndex map[string]int) error {
	query := `
		INSERT INTO advanced_stats (
			player_id, season, week, stat_type,
			avg_cushion, avg_separation, avg_intended_air_yards_receiving,
			percent_share_of_intended_air_yards,
			receptions, targets, catch_percentage, yards, rec_touchdowns,
			avg_yac, avg_expected_yac, avg_yac_above_expectation,
			updated_at
		) VALUES (
			$1, $2, $3, 'receiving',
			$4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW()
		)
		ON CONFLICT (player_id, season, week, stat_type)
		DO UPDATE SET
			avg_cushion = EXCLUDED.avg_cushion,
			avg_separation = EXCLUDED.avg_separation,
			avg_intended_air_yards_receiving = EXCLUDED.avg_intended_air_yards_receiving,
			percent_share_of_intended_air_yards = EXCLUDED.percent_share_of_intended_air_yards,
			receptions = EXCLUDED.receptions,
			targets = EXCLUDED.targets,
			catch_percentage = EXCLUDED.catch_percentage,
			yards = EXCLUDED.yards,
			rec_touchdowns = EXCLUDED.rec_touchdowns,
			avg_yac = EXCLUDED.avg_yac,
			avg_expected_yac = EXCLUDED.avg_expected_yac,
			avg_yac_above_expectation = EXCLUDED.avg_yac_above_expectation,
			updated_at = NOW()
	`

	_, err := s.dbPool.Exec(ctx, query,
		playerID, season, week,
		getCSVFloat(row, colIndex, "avg_cushion"),
		getCSVFloat(row, colIndex, "avg_separation"),
		getCSVFloat(row, colIndex, "avg_intended_air_yards"),
		getCSVFloat(row, colIndex, "percent_share_of_intended_air_yards"),
		getCSVInt(row, colIndex, "receptions"),
		getCSVInt(row, colIndex, "targets"),
		getCSVFloat(row, colIndex, "catch_percentage"),
		getCSVInt(row, colIndex, "yards"),
		getCSVInt(row, colIndex, "rec_touchdowns"),
		getCSVFloat(row, colIndex, "avg_yac"),
		getCSVFloat(row, colIndex, "avg_expected_yac"),
		getCSVFloat(row, colIndex, "avg_yac_above_expectation"),
	)

	return err
}

// findPlayerByNameAndTeam finds a player by name and team for a specific season
func (s *Service) findPlayerByNameAndTeam(ctx context.Context, playerName, teamAbbr string, season int) *uuid.UUID {
	// Try exact match first
	var playerID uuid.UUID
	err := s.dbPool.QueryRow(ctx, `
		SELECT p.id
		FROM players p
		JOIN teams t ON p.team_id = t.id
		WHERE LOWER(p.name) = LOWER($1)
		AND t.abbreviation = $2
		LIMIT 1
	`, playerName, teamAbbr).Scan(&playerID)

	if err == nil {
		return &playerID
	}

	// Try fuzzy match by last name
	parts := strings.Split(playerName, " ")
	if len(parts) < 2 {
		return nil
	}
	lastName := parts[len(parts)-1]

	err = s.dbPool.QueryRow(ctx, `
		SELECT p.id
		FROM players p
		JOIN teams t ON p.team_id = t.id
		WHERE LOWER(p.name) LIKE LOWER($1)
		AND t.abbreviation = $2
		LIMIT 1
	`, "%"+lastName+"%", teamAbbr).Scan(&playerID)

	if err == nil {
		return &playerID
	}

	// Try without team filter (player may have been traded)
	err = s.dbPool.QueryRow(ctx, `
		SELECT id
		FROM players
		WHERE LOWER(name) = LOWER($1)
		LIMIT 1
	`, playerName).Scan(&playerID)

	if err == nil {
		return &playerID
	}

	return nil
}

// SyncAllNextGenStats syncs all three Next Gen Stats types for a season
func (s *Service) SyncAllNextGenStats(ctx context.Context, season int) error {
	log.Printf("Syncing all Next Gen Stats for season %d...", season)

	statTypes := []string{"passing", "rushing", "receiving"}

	for _, statType := range statTypes {
		log.Printf("Syncing %s NGS...", statType)
		if err := s.SyncNextGenStats(ctx, season, statType); err != nil {
			log.Printf("Failed to sync %s NGS: %v", statType, err)
			// Continue with other types even if one fails
			continue
		}
	}

	log.Printf("All Next Gen Stats sync complete for season %d", season)
	return nil
}

// SyncNextGenStatsRange syncs Next Gen Stats for multiple seasons
func (s *Service) SyncNextGenStatsRange(ctx context.Context, startSeason, endSeason int, statType string) error {
	log.Printf("Syncing Next Gen Stats (%s) for seasons %d-%d...", statType, startSeason, endSeason)

	for season := startSeason; season <= endSeason; season++ {
		log.Printf("Processing season %d...", season)
		if err := s.SyncNextGenStats(ctx, season, statType); err != nil {
			log.Printf("Failed to sync season %d: %v", season, err)
			// Continue with next season
			continue
		}
	}

	log.Printf("Next Gen Stats (%s) range sync complete", statType)
	return nil
}
