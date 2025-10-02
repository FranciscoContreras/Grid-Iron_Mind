package ingestion

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/francisco/gridironmind/internal/espn"
	"github.com/google/uuid"
)

// SyncTeamStats syncs team statistics for completed games in a season/week
//
// This function:
//   - Fetches game details from ESPN API for each game
//   - Extracts box score team statistics
//   - Inserts/updates game_team_stats table
//
// Example usage:
//
//	if err := service.SyncTeamStats(ctx, 2025, 4); err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) SyncTeamStats(ctx context.Context, season, week int) error {
	log.Printf("Starting team stats sync for season %d week %d...", season, week)

	// Get all completed games for the season/week
	query := `
		SELECT id, nfl_game_id
		FROM games
		WHERE season = $1
		AND week = $2
		AND status = 'completed'
		AND nfl_game_id IS NOT NULL
		ORDER BY game_date
	`

	rows, err := s.dbPool.Query(ctx, query, season, week)
	if err != nil {
		return fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	games := []struct {
		ID        uuid.UUID
		NFLGameID string
	}{}

	for rows.Next() {
		var game struct {
			ID        uuid.UUID
			NFLGameID string
		}
		if err := rows.Scan(&game.ID, &game.NFLGameID); err != nil {
			log.Printf("Failed to scan game row: %v", err)
			continue
		}
		games = append(games, game)
	}

	if len(games) == 0 {
		log.Printf("No completed games found for season %d week %d", season, week)
		return nil
	}

	log.Printf("Found %d completed games to sync team stats", len(games))

	synced := 0
	for _, game := range games {
		if err := s.syncGameTeamStats(ctx, game.ID, game.NFLGameID); err != nil {
			log.Printf("Failed to sync team stats for game %s: %v", game.ID, err)
			continue
		}
		synced++
	}

	log.Printf("Successfully synced team stats for %d/%d games", synced, len(games))
	return nil
}

// syncGameTeamStats syncs team statistics for a single game
func (s *Service) syncGameTeamStats(ctx context.Context, gameID uuid.UUID, nflGameID string) error {
	// Fetch game details from ESPN
	gameDetail, err := s.espnClient.FetchGameDetails(ctx, nflGameID)
	if err != nil {
		return fmt.Errorf("failed to fetch game details: %w", err)
	}

	// Check if boxscore data exists
	if len(gameDetail.BoxScore.Teams) == 0 {
		return fmt.Errorf("no boxscore data available for game %s", nflGameID)
	}

	// Process each team's stats
	for _, teamBox := range gameDetail.BoxScore.Teams {
		if err := s.insertOrUpdateTeamStats(ctx, gameID, teamBox); err != nil {
			log.Printf("Failed to insert team stats: %v", err)
			continue
		}
	}

	return nil
}

// insertOrUpdateTeamStats inserts or updates team statistics for a game
func (s *Service) insertOrUpdateTeamStats(ctx context.Context, gameID uuid.UUID, teamBox struct {
	Team       espn.TeamInfo
	Statistics []struct {
		Name             string
		DisplayValue     string
		Value            interface{}
		Label            string
		Abbreviation     string
	}
}) error {
	// Find team by ESPN ID
	teamIDStr := teamBox.Team.ID
	if teamIDStr == "" {
		return fmt.Errorf("team ID is empty in boxscore")
	}

	nflTeamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		return fmt.Errorf("invalid team ID: %w", err)
	}

	// Look up our team_id by nfl_id
	var teamID uuid.UUID
	err = s.dbPool.QueryRow(ctx, "SELECT id FROM teams WHERE nfl_id = $1", nflTeamID).Scan(&teamID)
	if err != nil {
		return fmt.Errorf("team not found for nfl_id %d: %w", nflTeamID, err)
	}

	// Parse statistics from ESPN boxscore
	stats := parseTeamStats(teamBox.Statistics)

	// Insert or update team stats
	query := `
		INSERT INTO game_team_stats (
			game_id,
			team_id,
			first_downs,
			total_yards,
			passing_yards,
			rushing_yards,
			third_down_attempts,
			third_down_conversions,
			third_down_pct,
			turnovers,
			fumbles_lost,
			interceptions_thrown,
			penalties,
			penalty_yards,
			possession_time,
			completions,
			pass_attempts,
			sacks_allowed,
			rushing_attempts,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW()
		)
		ON CONFLICT (game_id, team_id)
		DO UPDATE SET
			first_downs = EXCLUDED.first_downs,
			total_yards = EXCLUDED.total_yards,
			passing_yards = EXCLUDED.passing_yards,
			rushing_yards = EXCLUDED.rushing_yards,
			third_down_attempts = EXCLUDED.third_down_attempts,
			third_down_conversions = EXCLUDED.third_down_conversions,
			third_down_pct = EXCLUDED.third_down_pct,
			turnovers = EXCLUDED.turnovers,
			fumbles_lost = EXCLUDED.fumbles_lost,
			interceptions_thrown = EXCLUDED.interceptions_thrown,
			penalties = EXCLUDED.penalties,
			penalty_yards = EXCLUDED.penalty_yards,
			possession_time = EXCLUDED.possession_time,
			completions = EXCLUDED.completions,
			pass_attempts = EXCLUDED.pass_attempts,
			sacks_allowed = EXCLUDED.sacks_allowed,
			rushing_attempts = EXCLUDED.rushing_attempts
	`

	_, err = s.dbPool.Exec(ctx, query,
		gameID,
		teamID,
		stats.FirstDowns,
		stats.TotalYards,
		stats.PassingYards,
		stats.RushingYards,
		stats.ThirdDownAttempts,
		stats.ThirdDownConversions,
		stats.ThirdDownPct,
		stats.Turnovers,
		stats.FumblesLost,
		stats.InterceptionsThrown,
		stats.Penalties,
		stats.PenaltyYards,
		stats.PossessionTime,
		stats.Completions,
		stats.PassAttempts,
		stats.SacksAllowed,
		stats.RushingAttempts,
	)

	if err != nil {
		return fmt.Errorf("failed to insert team stats: %w", err)
	}

	log.Printf("Synced team stats for game %s, team %s", gameID, teamID)
	return nil
}

// TeamStatsData holds parsed team statistics
type TeamStatsData struct {
	FirstDowns           int
	TotalYards           int
	PassingYards         int
	RushingYards         int
	ThirdDownAttempts    int
	ThirdDownConversions int
	ThirdDownPct         float64
	Turnovers            int
	FumblesLost          int
	InterceptionsThrown  int
	Penalties            int
	PenaltyYards         int
	PossessionTime       string
	Completions          int
	PassAttempts         int
	SacksAllowed         int
	RushingAttempts      int
}

// parseTeamStats extracts team statistics from ESPN boxscore format
func parseTeamStats(statistics []struct {
	Name         string
	DisplayValue string
	Value        interface{}
	Label        string
}) TeamStatsData {
	stats := TeamStatsData{}

	for _, stat := range statistics {
		value := getIntValue(stat.Value)

		switch stat.Name {
		case "firstDowns":
			stats.FirstDowns = value
		case "totalYards":
			stats.TotalYards = value
		case "netPassingYards":
			stats.PassingYards = value
		case "rushingYards":
			stats.RushingYards = value
		case "thirdDownEff":
			// Format: "5-12" (conversions-attempts)
			parts := parseEfficiency(stat.DisplayValue)
			if len(parts) == 2 {
				stats.ThirdDownConversions = parts[0]
				stats.ThirdDownAttempts = parts[1]
				if parts[1] > 0 {
					stats.ThirdDownPct = float64(parts[0]) / float64(parts[1]) * 100
				}
			}
		case "turnovers":
			stats.Turnovers = value
		case "fumblesLost":
			stats.FumblesLost = value
		case "passesIntercepted":
			stats.InterceptionsThrown = value
		case "penalties":
			// Format: "5-35" (penalties-yards)
			parts := parseEfficiency(stat.DisplayValue)
			if len(parts) == 2 {
				stats.Penalties = parts[0]
				stats.PenaltyYards = parts[1]
			}
		case "possessionTime":
			stats.PossessionTime = stat.DisplayValue
		case "completionAttempts":
			// Format: "20-30" (completions-attempts)
			parts := parseEfficiency(stat.DisplayValue)
			if len(parts) == 2 {
				stats.Completions = parts[0]
				stats.PassAttempts = parts[1]
			}
		case "totalPenaltiesYards":
			parts := parseEfficiency(stat.DisplayValue)
			if len(parts) == 2 {
				stats.Penalties = parts[0]
				stats.PenaltyYards = parts[1]
			}
		case "sacksYardsLost":
			// Format: "3-21" (sacks-yards)
			parts := parseEfficiency(stat.DisplayValue)
			if len(parts) >= 1 {
				stats.SacksAllowed = parts[0]
			}
		case "rushingAttempts":
			stats.RushingAttempts = value
		}
	}

	return stats
}

// getIntValue safely extracts an integer from interface{} value
func getIntValue(val interface{}) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if num, err := strconv.Atoi(v); err == nil {
			return num
		}
	}
	return 0
}

// parseEfficiency parses efficiency strings like "5-12" into [5, 12]
func parseEfficiency(s string) []int {
	parts := []string{}
	current := ""
	for _, ch := range s {
		if ch == '-' {
			parts = append(parts, current)
			current = ""
		} else if ch >= '0' && ch <= '9' {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	result := []int{}
	for _, p := range parts {
		if num, err := strconv.Atoi(p); err == nil {
			result = append(result, num)
		}
	}
	return result
}

// getTeamID extracts team ID from team info
func getTeamID(teamInfo interface{}) (string, bool) {
	if m, ok := teamInfo.(map[string]interface{}); ok {
		if id, ok := m["id"].(string); ok {
			return id, true
		}
	}
	return "", false
}

// SyncTeamStatsForSeason syncs team statistics for all completed games in a season
func (s *Service) SyncTeamStatsForSeason(ctx context.Context, season int) error {
	log.Printf("Starting team stats sync for entire season %d...", season)

	// Sync each week (1-18 for regular season)
	for week := 1; week <= 18; week++ {
		log.Printf("Syncing week %d/%d...", week, 18)
		if err := s.SyncTeamStats(ctx, season, week); err != nil {
			log.Printf("Failed to sync week %d: %v", week, err)
			// Continue with next week even if this one fails
			continue
		}
	}

	log.Printf("Completed team stats sync for season %d", season)
	return nil
}
