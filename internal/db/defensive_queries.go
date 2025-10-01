package db

import (
	"context"
	"fmt"

	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

type DefensiveQueries struct{}

// GetTeamDefensiveStats retrieves defensive stats for a team
func (q *DefensiveQueries) GetTeamDefensiveStats(ctx context.Context, teamID uuid.UUID, season int, week *int) (*models.TeamDefensiveStats, error) {
	pool := GetPool()

	query := `
		SELECT
			d.id, d.team_id, d.season, d.week,
			d.points_allowed, d.points_allowed_per_game,
			d.yards_allowed, d.yards_allowed_per_game,
			d.pass_yards_allowed, d.pass_yards_allowed_per_game,
			d.rush_yards_allowed, d.rush_yards_allowed_per_game,
			d.sacks, d.sacks_yards,
			d.interceptions, d.interception_yards, d.interception_touchdowns,
			d.forced_fumbles, d.fumble_recoveries, d.fumble_recovery_touchdowns,
			d.third_down_attempts, d.third_down_conversions_allowed, d.third_down_percentage,
			d.red_zone_attempts, d.red_zone_touchdowns_allowed, d.red_zone_percentage,
			d.pass_attempts_allowed, d.pass_completions_allowed, d.pass_touchdowns_allowed,
			d.rush_attempts_allowed, d.rush_touchdowns_allowed,
			d.penalties, d.penalty_yards,
			d.defensive_rank, d.pass_defense_rank, d.rush_defense_rank, d.points_allowed_rank,
			d.games_played, d.created_at, d.updated_at,
			t.abbreviation, t.name
		FROM team_defensive_stats d
		JOIN teams t ON d.team_id = t.id
		WHERE d.team_id = $1 AND d.season = $2
	`

	args := []interface{}{teamID, season}

	if week != nil {
		query += " AND d.week = $3"
		args = append(args, *week)
	} else {
		query += " AND d.week IS NULL"
	}

	var stats models.TeamDefensiveStats
	err := pool.QueryRow(ctx, query, args...).Scan(
		&stats.ID, &stats.TeamID, &stats.Season, &stats.Week,
		&stats.PointsAllowed, &stats.PointsAllowedPerGame,
		&stats.YardsAllowed, &stats.YardsAllowedPerGame,
		&stats.PassYardsAllowed, &stats.PassYardsAllowedPerGame,
		&stats.RushYardsAllowed, &stats.RushYardsAllowedPerGame,
		&stats.Sacks, &stats.SacksYards,
		&stats.Interceptions, &stats.InterceptionYards, &stats.InterceptionTouchdowns,
		&stats.ForcedFumbles, &stats.FumbleRecoveries, &stats.FumbleRecoveryTouchdowns,
		&stats.ThirdDownAttempts, &stats.ThirdDownConversionsAllowed, &stats.ThirdDownPercentage,
		&stats.RedZoneAttempts, &stats.RedZoneTouchdownsAllowed, &stats.RedZonePercentage,
		&stats.PassAttemptsAllowed, &stats.PassCompletionsAllowed, &stats.PassTouchdownsAllowed,
		&stats.RushAttemptsAllowed, &stats.RushTouchdownsAllowed,
		&stats.Penalties, &stats.PenaltyYards,
		&stats.DefensiveRank, &stats.PassDefenseRank, &stats.RushDefenseRank, &stats.PointsAllowedRank,
		&stats.GamesPlayed, &stats.CreatedAt, &stats.UpdatedAt,
		&stats.TeamAbbr, &stats.TeamName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get team defensive stats: %w", err)
	}

	return &stats, nil
}

// GetDefensiveRankings retrieves league-wide defensive rankings
func (q *DefensiveQueries) GetDefensiveRankings(ctx context.Context, season int, category string) ([]models.DefensiveRanking, error) {
	pool := GetPool()

	var query string
	var orderBy string
	var valueColumn string

	switch category {
	case "pass":
		orderBy = "d.pass_yards_allowed_per_game ASC"
		valueColumn = "d.pass_yards_allowed_per_game"
	case "rush":
		orderBy = "d.rush_yards_allowed_per_game ASC"
		valueColumn = "d.rush_yards_allowed_per_game"
	case "points_allowed":
		orderBy = "d.points_allowed_per_game ASC"
		valueColumn = "d.points_allowed_per_game"
	default: // overall
		orderBy = "d.yards_allowed_per_game ASC"
		valueColumn = "d.yards_allowed_per_game"
		category = "overall"
	}

	query = fmt.Sprintf(`
		SELECT
			ROW_NUMBER() OVER (ORDER BY %s) as rank,
			d.team_id,
			t.abbreviation,
			t.name,
			%s as value,
			d.season
		FROM team_defensive_stats d
		JOIN teams t ON d.team_id = t.id
		WHERE d.season = $1 AND d.week IS NULL
		ORDER BY %s
	`, orderBy, valueColumn, orderBy)

	rows, err := pool.Query(ctx, query, season)
	if err != nil {
		return nil, fmt.Errorf("failed to get defensive rankings: %w", err)
	}
	defer rows.Close()

	var rankings []models.DefensiveRanking
	for rows.Next() {
		var r models.DefensiveRanking
		r.Category = category

		err := rows.Scan(
			&r.Rank,
			&r.TeamID,
			&r.TeamAbbr,
			&r.TeamName,
			&r.Value,
			&r.Season,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan defensive ranking: %w", err)
		}

		rankings = append(rankings, r)
	}

	return rankings, nil
}

// GetPlayerVsDefense retrieves a player's historical performance against a specific defense
func (q *DefensiveQueries) GetPlayerVsDefense(ctx context.Context, playerID, defenseTeamID uuid.UUID, season *int, limit int) (*models.PlayerVsDefenseResponse, error) {
	pool := GetPool()

	// Build query with optional season filter
	query := `
		SELECT
			pvd.id, pvd.player_id, pvd.defense_team_id, pvd.game_id,
			pvd.season, pvd.week,
			pvd.passing_yards, pvd.passing_tds, pvd.interceptions_thrown,
			pvd.rushing_yards, pvd.rushing_tds,
			pvd.receptions, pvd.receiving_yards, pvd.receiving_tds,
			pvd.fantasy_points_standard, pvd.fantasy_points_ppr, pvd.fantasy_points_half_ppr,
			pvd.created_at, pvd.updated_at
		FROM player_vs_defense_history pvd
		WHERE pvd.player_id = $1 AND pvd.defense_team_id = $2
	`

	args := []interface{}{playerID, defenseTeamID}
	argNum := 3

	if season != nil {
		query += fmt.Sprintf(" AND pvd.season = $%d", argNum)
		args = append(args, *season)
		argNum++
	}

	query += " ORDER BY pvd.season DESC, pvd.week DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, limit)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query player vs defense: %w", err)
	}
	defer rows.Close()

	var games []models.PlayerVsDefenseHistory
	for rows.Next() {
		var game models.PlayerVsDefenseHistory
		err := rows.Scan(
			&game.ID, &game.PlayerID, &game.DefenseTeamID, &game.GameID,
			&game.Season, &game.Week,
			&game.PassingYards, &game.PassingTds, &game.InterceptionsThrown,
			&game.RushingYards, &game.RushingTds,
			&game.Receptions, &game.ReceivingYards, &game.ReceivingTds,
			&game.FantasyPointsStandard, &game.FantasyPointsPPR, &game.FantasyPointsHalfPPR,
			&game.CreatedAt, &game.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player vs defense game: %w", err)
		}

		games = append(games, game)
	}

	// Get player and team names
	var playerName, defenseAbbr string
	err = pool.QueryRow(ctx, "SELECT name FROM players WHERE id = $1", playerID).Scan(&playerName)
	if err != nil {
		return nil, fmt.Errorf("failed to get player name: %w", err)
	}

	err = pool.QueryRow(ctx, "SELECT abbreviation FROM teams WHERE id = $1", defenseTeamID).Scan(&defenseAbbr)
	if err != nil {
		return nil, fmt.Errorf("failed to get team abbreviation: %w", err)
	}

	// Calculate averages
	var averages *models.PlayerVsDefenseAverages
	if len(games) > 0 {
		averages = calculatePlayerVsDefenseAverages(games)
	}

	response := &models.PlayerVsDefenseResponse{
		PlayerID:        playerID,
		PlayerName:      playerName,
		DefenseTeamID:   defenseTeamID,
		DefenseTeamAbbr: defenseAbbr,
		Games:           games,
		Averages:        averages,
	}

	return response, nil
}

// Helper function to calculate averages
func calculatePlayerVsDefenseAverages(games []models.PlayerVsDefenseHistory) *models.PlayerVsDefenseAverages {
	if len(games) == 0 {
		return nil
	}

	var totalStandard, totalPPR, totalHalfPPR float64
	var totalYards, totalTouchdowns int
	gamesPlayed := len(games)

	for _, game := range games {
		if game.FantasyPointsStandard != nil {
			totalStandard += *game.FantasyPointsStandard
		}
		if game.FantasyPointsPPR != nil {
			totalPPR += *game.FantasyPointsPPR
		}
		if game.FantasyPointsHalfPPR != nil {
			totalHalfPPR += *game.FantasyPointsHalfPPR
		}

		totalYards += game.PassingYards + game.RushingYards + game.ReceivingYards
		totalTouchdowns += game.PassingTds + game.RushingTds + game.ReceivingTds
	}

	avgStandard := totalStandard / float64(gamesPlayed)
	avgPPR := totalPPR / float64(gamesPlayed)
	avgHalfPPR := totalHalfPPR / float64(gamesPlayed)
	avgYards := float64(totalYards) / float64(gamesPlayed)
	avgTouchdowns := float64(totalTouchdowns) / float64(gamesPlayed)

	return &models.PlayerVsDefenseAverages{
		GamesPlayed:                  gamesPlayed,
		FantasyPointsPerGameStandard: &avgStandard,
		FantasyPointsPerGamePPR:      &avgPPR,
		FantasyPointsPerGameHalfPPR:  &avgHalfPPR,
		YardsPerGame:                 &avgYards,
		TouchdownsPerGame:            &avgTouchdowns,
	}
}
