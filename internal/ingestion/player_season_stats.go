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
	"time"

	"github.com/google/uuid"
)

// SyncPlayerSeasonStats syncs player season statistics from NFLverse data
//
// This function:
//   - Fetches player stats CSV from NFLverse GitHub releases
//   - Aggregates weekly stats into season totals
//   - Inserts/updates player_season_stats table
//
// Example usage:
//
//	if err := service.SyncPlayerSeasonStats(ctx, 2024); err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) SyncPlayerSeasonStats(ctx context.Context, season int) error {
	log.Printf("Starting player season stats sync for season %d...", season)

	// Fetch CSV data from NFLverse
	url := fmt.Sprintf("https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_%d.csv", season)

	log.Printf("Fetching player stats from: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch NFLverse data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NFLverse API returned status %d", resp.StatusCode)
	}

	// Parse CSV
	reader := csv.NewReader(resp.Body)

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

	// Aggregate stats by player and season
	type SeasonStats struct {
		PlayerID    string
		PlayerName  string
		Position    string
		TeamAbbr    string
		GamesPlayed int
		GamesStarted int

		// Passing
		PassAttempts     int
		PassCompletions  int
		PassYards        float64
		PassTDs          int
		PassINTs         int
		Sacks            float64
		SackYards        float64

		// Rushing
		RushAttempts     int
		RushYards        float64
		RushTDs          int
		Fumbles          int
		FumblesLost      int

		// Receiving
		Receptions       int
		RecYards         float64
		RecTDs           int
		Targets          int
	}

	playerStats := make(map[string]*SeasonStats)

	// Process each row
	rowCount := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading CSV row: %v", err)
			continue
		}

		rowCount++

		// Only process regular season stats
		seasonType := getCSVValue(row, colIndex, "season_type")
		if seasonType != "REG" {
			continue
		}

		// Extract player info
		playerID := getCSVValue(row, colIndex, "player_id")
		if playerID == "" {
			continue
		}

		// Initialize or get existing stats
		if playerStats[playerID] == nil {
			playerStats[playerID] = &SeasonStats{
				PlayerID:   playerID,
				PlayerName: getCSVValue(row, colIndex, "player_display_name"),
				Position:   getCSVValue(row, colIndex, "position"),
				TeamAbbr:   getCSVValue(row, colIndex, "recent_team"),
			}
		}

		stats := playerStats[playerID]

		// Count games (each row is a week)
		stats.GamesPlayed++

		// Aggregate passing stats
		stats.PassAttempts += getCSVInt(row, colIndex, "attempts")
		stats.PassCompletions += getCSVInt(row, colIndex, "completions")
		stats.PassYards += getCSVFloat(row, colIndex, "passing_yards")
		stats.PassTDs += getCSVInt(row, colIndex, "passing_tds")
		stats.PassINTs += getCSVInt(row, colIndex, "interceptions")
		stats.Sacks += getCSVFloat(row, colIndex, "sacks")
		stats.SackYards += getCSVFloat(row, colIndex, "sack_yards")

		// Aggregate rushing stats
		stats.RushAttempts += getCSVInt(row, colIndex, "carries")
		stats.RushYards += getCSVFloat(row, colIndex, "rushing_yards")
		stats.RushTDs += getCSVInt(row, colIndex, "rushing_tds")
		stats.Fumbles += getCSVInt(row, colIndex, "rushing_fumbles") + getCSVInt(row, colIndex, "sack_fumbles")
		stats.FumblesLost += getCSVInt(row, colIndex, "rushing_fumbles_lost") + getCSVInt(row, colIndex, "sack_fumbles_lost")

		// Aggregate receiving stats
		stats.Receptions += getCSVInt(row, colIndex, "receptions")
		stats.RecYards += getCSVFloat(row, colIndex, "receiving_yards")
		stats.RecTDs += getCSVInt(row, colIndex, "receiving_tds")
		stats.Targets += getCSVInt(row, colIndex, "targets")
	}

	log.Printf("Parsed %d rows, aggregated stats for %d players", rowCount, len(playerStats))

	// Insert aggregated stats into database
	inserted := 0
	updated := 0

	for _, stats := range playerStats {
		// Look up player by GSIS ID (NFLverse player_id)
		var playerUUID uuid.UUID
		var teamUUID uuid.UUID

		// Try to find player by GSIS ID (stored in players table)
		err := s.dbPool.QueryRow(ctx, `
			SELECT id FROM players WHERE name ILIKE $1 LIMIT 1
		`, stats.PlayerName).Scan(&playerUUID)

		if err != nil {
			// Player not found, skip
			log.Printf("Player not found: %s (GSIS: %s)", stats.PlayerName, stats.PlayerID)
			continue
		}

		// Look up team by abbreviation
		err = s.dbPool.QueryRow(ctx, `
			SELECT id FROM teams WHERE abbreviation = $1 LIMIT 1
		`, stats.TeamAbbr).Scan(&teamUUID)

		if err != nil {
			// Team not found, set to NULL
			log.Printf("Team not found for abbreviation: %s", stats.TeamAbbr)
		}

		// Calculate averages
		var rushAvg, recAvg float64
		if stats.RushAttempts > 0 {
			rushAvg = stats.RushYards / float64(stats.RushAttempts)
		}
		if stats.Receptions > 0 {
			recAvg = stats.RecYards / float64(stats.Receptions)
		}

		// Calculate passer rating (simplified NFL passer rating formula)
		var passerRating float64
		if stats.PassAttempts > 0 {
			a := ((float64(stats.PassCompletions) / float64(stats.PassAttempts)) - 0.3) * 5
			b := ((stats.PassYards / float64(stats.PassAttempts)) - 3) * 0.25
			c := (float64(stats.PassTDs) / float64(stats.PassAttempts)) * 20
			d := 2.375 - ((float64(stats.PassINTs) / float64(stats.PassAttempts)) * 25)

			// Clamp values
			a = clamp(a, 0, 2.375)
			b = clamp(b, 0, 2.375)
			c = clamp(c, 0, 2.375)
			d = clamp(d, 0, 2.375)

			passerRating = ((a + b + c + d) / 6) * 100
		}

		// Upsert to database
		query := `
			INSERT INTO player_season_stats (
				player_id, season, team_id, position,
				games_played, games_started,
				passing_attempts, passing_completions, passing_yards, passing_tds, passing_ints, passing_rating,
				sacks, sack_yards,
				rushing_attempts, rushing_yards, rushing_tds, rushing_avg,
				fumbles, fumbles_lost,
				receptions, receiving_yards, receiving_tds, receiving_avg, targets,
				created_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, NOW()
			)
			ON CONFLICT (player_id, season)
			DO UPDATE SET
				team_id = EXCLUDED.team_id,
				position = EXCLUDED.position,
				games_played = EXCLUDED.games_played,
				games_started = EXCLUDED.games_started,
				passing_attempts = EXCLUDED.passing_attempts,
				passing_completions = EXCLUDED.passing_completions,
				passing_yards = EXCLUDED.passing_yards,
				passing_tds = EXCLUDED.passing_tds,
				passing_ints = EXCLUDED.passing_ints,
				passing_rating = EXCLUDED.passing_rating,
				sacks = EXCLUDED.sacks,
				sack_yards = EXCLUDED.sack_yards,
				rushing_attempts = EXCLUDED.rushing_attempts,
				rushing_yards = EXCLUDED.rushing_yards,
				rushing_tds = EXCLUDED.rushing_tds,
				rushing_avg = EXCLUDED.rushing_avg,
				fumbles = EXCLUDED.fumbles,
				fumbles_lost = EXCLUDED.fumbles_lost,
				receptions = EXCLUDED.receptions,
				receiving_yards = EXCLUDED.receiving_yards,
				receiving_tds = EXCLUDED.receiving_tds,
				receiving_avg = EXCLUDED.receiving_avg,
				targets = EXCLUDED.targets
		`

		tag, err := s.dbPool.Exec(ctx, query,
			playerUUID, season, teamUUID, stats.Position,
			stats.GamesPlayed, stats.GamesStarted,
			stats.PassAttempts, stats.PassCompletions, int(stats.PassYards), stats.PassTDs, stats.PassINTs, passerRating,
			int(stats.Sacks), int(stats.SackYards),
			stats.RushAttempts, int(stats.RushYards), stats.RushTDs, rushAvg,
			stats.Fumbles, stats.FumblesLost,
			stats.Receptions, int(stats.RecYards), stats.RecTDs, recAvg, stats.Targets,
		)

		if err != nil {
			log.Printf("Failed to upsert stats for player %s: %v", stats.PlayerName, err)
			continue
		}

		if tag.RowsAffected() > 0 {
			if strings.Contains(tag.String(), "INSERT") {
				inserted++
			} else {
				updated++
			}
		}
	}

	log.Printf("Player season stats sync completed: %d inserted, %d updated", inserted, updated)
	return nil
}

// Helper functions for CSV parsing

func getCSVValue(row []string, colIndex map[string]int, colName string) string {
	if idx, ok := colIndex[colName]; ok && idx < len(row) {
		return strings.TrimSpace(row[idx])
	}
	return ""
}

func getCSVInt(row []string, colIndex map[string]int, colName string) int {
	val := getCSVValue(row, colIndex, colName)
	if val == "" || val == "NA" {
		return 0
	}
	i, _ := strconv.Atoi(val)
	return i
}

func getCSVFloat(row []string, colIndex map[string]int, colName string) float64 {
	val := getCSVValue(row, colIndex, colName)
	if val == "" || val == "NA" {
		return 0
	}
	f, _ := strconv.ParseFloat(val, 64)
	return f
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// SyncPlayerSeasonStatsRange syncs player stats for multiple seasons
func (s *Service) SyncPlayerSeasonStatsRange(ctx context.Context, startSeason, endSeason int) error {
	log.Printf("Starting player season stats sync for seasons %d-%d...", startSeason, endSeason)

	for season := startSeason; season <= endSeason; season++ {
		log.Printf("Syncing season %d...", season)
		if err := s.SyncPlayerSeasonStats(ctx, season); err != nil {
			log.Printf("Failed to sync season %d: %v", season, err)
			// Continue with next season
			continue
		}

		// Rate limiting - be nice to GitHub
		time.Sleep(2 * time.Second)
	}

	log.Printf("Completed player season stats sync for seasons %d-%d", startSeason, endSeason)
	return nil
}
