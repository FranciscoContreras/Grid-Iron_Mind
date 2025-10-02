package ingestion

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/francisco/gridironmind/internal/espn"
	"github.com/google/uuid"
)

// SyncScoringPlays syncs scoring plays for completed games in a season/week
//
// This function:
//   - Fetches game details from ESPN API for each game
//   - Extracts scoring plays timeline
//   - Parses player names from play descriptions
//   - Inserts/updates game_scoring_plays table
//
// Example usage:
//
//	if err := service.SyncScoringPlays(ctx, 2025, 4); err != nil {
//	    log.Fatal(err)
//	}
func (s *Service) SyncScoringPlays(ctx context.Context, season, week int) error {
	log.Printf("Starting scoring plays sync for season %d week %d...", season, week)

	// Get all completed games for the season/week
	query := `
		SELECT id, nfl_game_id, home_team_id, away_team_id
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

	type gameInfo struct {
		ID         uuid.UUID
		NFLGameID  string
		HomeTeamID uuid.UUID
		AwayTeamID uuid.UUID
	}

	var games []gameInfo
	for rows.Next() {
		var game gameInfo
		if err := rows.Scan(&game.ID, &game.NFLGameID, &game.HomeTeamID, &game.AwayTeamID); err != nil {
			log.Printf("Failed to scan game row: %v", err)
			continue
		}
		games = append(games, game)
	}

	if len(games) == 0 {
		log.Printf("No completed games found for season %d week %d", season, week)
		return nil
	}

	log.Printf("Found %d completed games to sync scoring plays", len(games))

	synced := 0
	totalPlays := 0
	for _, game := range games {
		count, err := s.syncGameScoringPlays(ctx, game.ID, game.NFLGameID, game.HomeTeamID, game.AwayTeamID)
		if err != nil {
			log.Printf("Failed to sync scoring plays for game %s: %v", game.ID, err)
			continue
		}
		synced++
		totalPlays += count
	}

	log.Printf("Successfully synced scoring plays: %d/%d games, %d total plays", synced, len(games), totalPlays)
	return nil
}

// syncGameScoringPlays syncs scoring plays for a single game
func (s *Service) syncGameScoringPlays(ctx context.Context, gameID uuid.UUID, nflGameID string, homeTeamID, awayTeamID uuid.UUID) (int, error) {
	// Fetch game details from ESPN
	gameDetail, err := s.espnClient.FetchGameDetails(ctx, nflGameID)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch game details: %w", err)
	}

	// Check if scoring plays data exists
	if len(gameDetail.ScoringPlays) == 0 {
		log.Printf("No scoring plays data for game %s", nflGameID)
		return 0, nil
	}

	log.Printf("Found %d scoring plays for game %s", len(gameDetail.ScoringPlays), nflGameID)

	// Delete existing plays for this game (clean slate for re-sync)
	_, err = s.dbPool.Exec(ctx, "DELETE FROM game_scoring_plays WHERE game_id = $1", gameID)
	if err != nil {
		return 0, fmt.Errorf("failed to delete existing plays: %w", err)
	}

	// Process each scoring play
	count := 0
	for i, play := range gameDetail.ScoringPlays {
		if err := s.insertScoringPlay(ctx, gameID, homeTeamID, awayTeamID, play, i+1); err != nil {
			log.Printf("Failed to insert scoring play: %v", err)
			continue
		}
		count++
	}

	log.Printf("Synced %d scoring plays for game %s", count, gameID)
	return count, nil
}

// insertScoringPlay inserts a scoring play into the database
func (s *Service) insertScoringPlay(ctx context.Context, gameID, homeTeamID, awayTeamID uuid.UUID, play espn.ScoringPlay, sequence int) error {
	// Map ESPN team ID to our team UUID
	espnTeamIDStr := play.Team.ID
	espnTeamID, err := strconv.Atoi(espnTeamIDStr)
	if err != nil {
		return fmt.Errorf("invalid ESPN team ID: %w", err)
	}

	var teamID uuid.UUID
	err = s.dbPool.QueryRow(ctx, "SELECT id FROM teams WHERE nfl_id = $1", espnTeamID).Scan(&teamID)
	if err != nil {
		return fmt.Errorf("team not found for nfl_id %d: %w", espnTeamID, err)
	}

	// Parse play details
	playType := play.Type.Abbreviation        // TD, FG, Safety, etc.
	scoringType := play.ScoringType.Name      // touchdown, field-goal, etc.
	description := strings.TrimSpace(play.Text)
	quarter := play.Period.Number
	timeRemaining := play.Clock.DisplayValue

	// Calculate points based on play type
	points := calculatePoints(playType, scoringType)

	// Parse player names from description (basic extraction)
	scoringPlayerName, assistPlayerName := parsePlayerNames(description, scoringType)

	// Look up player IDs (best effort)
	var scoringPlayerID, assistPlayerID *uuid.UUID
	if scoringPlayerName != "" {
		if pid := s.findPlayerByName(ctx, scoringPlayerName, teamID); pid != nil {
			scoringPlayerID = pid
		}
	}
	if assistPlayerName != "" {
		if pid := s.findPlayerByName(ctx, assistPlayerName, teamID); pid != nil {
			assistPlayerID = pid
		}
	}

	// Insert scoring play
	query := `
		INSERT INTO game_scoring_plays (
			game_id,
			team_id,
			quarter,
			time_remaining,
			sequence_number,
			play_type,
			scoring_type,
			points,
			description,
			scoring_player_id,
			assist_player_id,
			home_score,
			away_score,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()
		)
	`

	_, err = s.dbPool.Exec(ctx, query,
		gameID,
		teamID,
		quarter,
		timeRemaining,
		sequence,
		playType,
		scoringType,
		points,
		description,
		scoringPlayerID,
		assistPlayerID,
		play.HomeScore,
		play.AwayScore,
	)

	if err != nil {
		return fmt.Errorf("failed to insert scoring play: %w", err)
	}

	return nil
}

// calculatePoints determines point value based on play type
func calculatePoints(playType, scoringType string) int {
	switch playType {
	case "TD":
		return 6 // Touchdown (XP added separately)
	case "FG":
		return 3 // Field goal
	case "XP", "PAT":
		return 1 // Extra point
	case "2PT":
		return 2 // Two-point conversion
	case "SF", "SFTY":
		return 2 // Safety
	default:
		// Try to infer from scoring type
		if strings.Contains(scoringType, "field-goal") {
			return 3
		} else if strings.Contains(scoringType, "touchdown") {
			return 6
		} else if strings.Contains(scoringType, "safety") {
			return 2
		}
		return 0
	}
}

// parsePlayerNames extracts player names from play description
//
// Examples:
//   - "Patrick Mahomes 15 Yd pass from Josh Allen (Harrison Butker Kick)" → ("Patrick Mahomes", "Josh Allen")
//   - "Travis Kelce 8 Yd pass from Patrick Mahomes (Harrison Butker Kick)" → ("Travis Kelce", "Patrick Mahomes")
//   - "Derrick Henry 25 Yd Run (Ryan Succop Kick)" → ("Derrick Henry", "")
//   - "Harrison Butker 45 Yd Field Goal" → ("Harrison Butker", "")
func parsePlayerNames(description, scoringType string) (scoringPlayer, assistPlayer string) {
	// Clean description
	desc := strings.TrimSpace(description)

	// Pattern for passing touchdowns: "Receiver X Yd pass from QB (...)"
	passingTDRegex := regexp.MustCompile(`^([A-Za-z\.\-'\s]+?)\s+\d+\s+Yd\s+pass\s+from\s+([A-Za-z\.\-'\s]+?)(?:\s+\(|$)`)
	if matches := passingTDRegex.FindStringSubmatch(desc); len(matches) == 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
	}

	// Pattern for rushing touchdowns: "Runner X Yd Run (...)"
	rushingTDRegex := regexp.MustCompile(`^([A-Za-z\.\-'\s]+?)\s+\d+\s+Yd\s+Run(?:\s+\(|$)`)
	if matches := rushingTDRegex.FindStringSubmatch(desc); len(matches) == 2 {
		return strings.TrimSpace(matches[1]), ""
	}

	// Pattern for field goals: "Kicker X Yd Field Goal"
	fgRegex := regexp.MustCompile(`^([A-Za-z\.\-'\s]+?)\s+\d+\s+Yd\s+Field\s+Goal`)
	if matches := fgRegex.FindStringSubmatch(desc); len(matches) == 2 {
		return strings.TrimSpace(matches[1]), ""
	}

	// Pattern for extra points/kicks in parentheses: "(...) (Kicker Kick)"
	kickRegex := regexp.MustCompile(`\(([A-Za-z\.\-'\s]+?)\s+Kick\)`)
	if matches := kickRegex.FindStringSubmatch(desc); len(matches) == 2 {
		// This is just the kicker, not the primary scorer
		// Already handled in the main patterns above
	}

	// Pattern for defensive touchdowns: "Player Fumble Return/Interception (...)"
	defTDRegex := regexp.MustCompile(`^([A-Za-z\.\-'\s]+?)\s+\d+\s+Yd\s+(Fumble|Interception|Kickoff|Punt)\s+(Return|Recovered)`)
	if matches := defTDRegex.FindStringSubmatch(desc); len(matches) >= 2 {
		return strings.TrimSpace(matches[1]), ""
	}

	// Fallback: extract first name (player who scored)
	nameRegex := regexp.MustCompile(`^([A-Za-z\.\-'\s]+?)\s+\d+`)
	if matches := nameRegex.FindStringSubmatch(desc); len(matches) == 2 {
		return strings.TrimSpace(matches[1]), ""
	}

	return "", ""
}

// findPlayerByName finds a player ID by name (case-insensitive, fuzzy match)
func (s *Service) findPlayerByName(ctx context.Context, name string, teamID uuid.UUID) *uuid.UUID {
	// First try exact match
	var id uuid.UUID
	query := `SELECT id FROM players WHERE LOWER(name) = LOWER($1) AND team_id = $2 LIMIT 1`
	err := s.dbPool.QueryRow(ctx, query, name, teamID).Scan(&id)
	if err == nil {
		return &id
	}

	// Try fuzzy match (last name only)
	parts := strings.Fields(name)
	if len(parts) > 0 {
		lastName := parts[len(parts)-1]
		query := `SELECT id FROM players WHERE LOWER(name) LIKE LOWER($1) AND team_id = $2 LIMIT 1`
		err := s.dbPool.QueryRow(ctx, query, "%"+lastName+"%", teamID).Scan(&id)
		if err == nil {
			return &id
		}
	}

	return nil
}

// SyncScoringPlaysForSeason syncs scoring plays for all completed games in a season
func (s *Service) SyncScoringPlaysForSeason(ctx context.Context, season int) error {
	log.Printf("Starting scoring plays sync for entire season %d...", season)

	// Sync each week (1-18 for regular season)
	for week := 1; week <= 18; week++ {
		log.Printf("Syncing scoring plays for week %d/%d...", week, 18)
		if err := s.SyncScoringPlays(ctx, season, week); err != nil {
			log.Printf("Failed to sync scoring plays for week %d: %v", week, err)
			// Continue with next week even if this one fails
			continue
		}
	}

	log.Printf("Completed scoring plays sync for season %d", season)
	return nil
}
