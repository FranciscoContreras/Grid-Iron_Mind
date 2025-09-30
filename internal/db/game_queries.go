package db

import (
	"context"
	"fmt"

	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

type GameQueries struct{}

// ListGames retrieves games with optional filters
func (q *GameQueries) ListGames(ctx context.Context, filters models.GameFilters) ([]models.Game, int, error) {
	pool := GetPool()
	if pool == nil {
		return nil, 0, fmt.Errorf("database connection not initialized")
	}

	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if filters.Season > 0 {
		whereClause += fmt.Sprintf(" AND season = $%d", argPos)
		args = append(args, filters.Season)
		argPos++
	}

	if filters.Week > 0 {
		whereClause += fmt.Sprintf(" AND week = $%d", argPos)
		args = append(args, filters.Week)
		argPos++
	}

	if filters.TeamID != uuid.Nil {
		whereClause += fmt.Sprintf(" AND (home_team_id = $%d OR away_team_id = $%d)", argPos, argPos)
		args = append(args, filters.TeamID)
		argPos++
	}

	if filters.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, filters.Status)
		argPos++
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM games %s", whereClause)
	if err := pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count games: %w", err)
	}

	// Fetch games
	query := fmt.Sprintf(`
		SELECT id, nfl_game_id, season, week, game_date,
		       home_team_id, away_team_id, home_score, away_score, status,
		       created_at
		FROM games
		%s
		ORDER BY game_date DESC, id
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var games []models.Game
	for rows.Next() {
		var g models.Game
		if err := rows.Scan(
			&g.ID, &g.EspnGameID, &g.SeasonYear, &g.Week,
			&g.GameDate, &g.HomeTeamID, &g.AwayTeamID, &g.HomeScore,
			&g.AwayScore, &g.Status, &g.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, g)
	}

	return games, total, nil
}

// GetGameByID retrieves a single game by ID
func (q *GameQueries) GetGameByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, nfl_game_id, season, week, game_date,
		       home_team_id, away_team_id, home_score, away_score, status,
		       created_at
		FROM games
		WHERE id = $1
	`

	var g models.Game
	err := pool.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.EspnGameID, &g.SeasonYear, &g.Week,
		&g.GameDate, &g.HomeTeamID, &g.AwayTeamID, &g.HomeScore,
		&g.AwayScore, &g.Status, &g.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %w", err)
	}

	return &g, nil
}

type StatsQueries struct{}

// GetGameStats retrieves all stats for a specific game
func (q *StatsQueries) GetGameStats(ctx context.Context, gameID uuid.UUID) ([]models.GameStat, error) {
	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT id, player_id, game_id, team_id, season, week,
		       passing_yards, passing_touchdowns, passing_interceptions, passing_completions, passing_attempts,
		       rushing_yards, rushing_touchdowns, rushing_attempts,
		       receiving_yards, receiving_touchdowns, receiving_receptions, receiving_targets,
		       created_at, updated_at
		FROM game_stats
		WHERE game_id = $1
		ORDER BY passing_yards DESC, rushing_yards DESC, receiving_yards DESC
	`

	rows, err := pool.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to query game stats: %w", err)
	}
	defer rows.Close()

	var stats []models.GameStat
	for rows.Next() {
		var s models.GameStat
		if err := rows.Scan(
			&s.ID, &s.PlayerID, &s.GameID, &s.TeamID, &s.SeasonYear, &s.Week,
			&s.PassingYards, &s.PassingTouchdowns, &s.PassingInterceptions,
			&s.PassingCompletions, &s.PassingAttempts,
			&s.RushingYards, &s.RushingTouchdowns, &s.RushingAttempts,
			&s.ReceivingYards, &s.ReceivingTouchdowns, &s.ReceivingReceptions,
			&s.ReceivingTargets, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stat: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetPlayerStats retrieves all stats for a specific player
func (q *StatsQueries) GetPlayerStats(ctx context.Context, playerID uuid.UUID, filters models.StatsFilters) ([]models.GameStat, error) {
	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	whereClause := "WHERE player_id = $1"
	args := []interface{}{playerID}
	argPos := 2

	if filters.Season > 0 {
		whereClause += fmt.Sprintf(" AND season = $%d", argPos)
		args = append(args, filters.Season)
		argPos++
	}

	if filters.Week > 0 {
		whereClause += fmt.Sprintf(" AND week = $%d", argPos)
		args = append(args, filters.Week)
		argPos++
	}

	query := fmt.Sprintf(`
		SELECT id, player_id, game_id, team_id, season, week,
		       passing_yards, passing_touchdowns, passing_interceptions, passing_completions, passing_attempts,
		       rushing_yards, rushing_touchdowns, rushing_attempts,
		       receiving_yards, receiving_touchdowns, receiving_receptions, receiving_targets,
		       created_at, updated_at
		FROM game_stats
		%s
		ORDER BY season DESC, week DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query player stats: %w", err)
	}
	defer rows.Close()

	var stats []models.GameStat
	for rows.Next() {
		var s models.GameStat
		if err := rows.Scan(
			&s.ID, &s.PlayerID, &s.GameID, &s.TeamID, &s.SeasonYear, &s.Week,
			&s.PassingYards, &s.PassingTouchdowns, &s.PassingInterceptions,
			&s.PassingCompletions, &s.PassingAttempts,
			&s.RushingYards, &s.RushingTouchdowns, &s.RushingAttempts,
			&s.ReceivingYards, &s.ReceivingTouchdowns, &s.ReceivingReceptions,
			&s.ReceivingTargets, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan stat: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetStatsLeaders retrieves top performers by stat category
func (q *StatsQueries) GetStatsLeaders(ctx context.Context, category string, season int, limit int) ([]models.PlayerStatLeader, error) {
	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Determine which stat to aggregate
	var statColumn string
	switch category {
	case "passing_yards":
		statColumn = "SUM(passing_yards) as total_stat, SUM(passing_touchdowns) as touchdowns, SUM(passing_interceptions) as interceptions"
	case "passing_touchdowns":
		statColumn = "SUM(passing_touchdowns) as total_stat, SUM(passing_yards) as yards, SUM(passing_interceptions) as interceptions"
	case "rushing_yards":
		statColumn = "SUM(rushing_yards) as total_stat, SUM(rushing_touchdowns) as touchdowns, SUM(rushing_attempts) as attempts"
	case "rushing_touchdowns":
		statColumn = "SUM(rushing_touchdowns) as total_stat, SUM(rushing_yards) as yards, SUM(rushing_attempts) as attempts"
	case "receiving_yards":
		statColumn = "SUM(receiving_yards) as total_stat, SUM(receiving_touchdowns) as touchdowns, SUM(receiving_receptions) as receptions"
	case "receiving_touchdowns":
		statColumn = "SUM(receiving_touchdowns) as total_stat, SUM(receiving_yards) as yards, SUM(receiving_receptions) as receptions"
	default:
		return nil, fmt.Errorf("invalid category: %s", category)
	}

	query := fmt.Sprintf(`
		SELECT
			gs.player_id,
			p.name as player_name,
			p.position,
			p.jersey_number,
			gs.team_id,
			t.abbreviation as team_abbr,
			%s,
			COUNT(*) as games_played
		FROM game_stats gs
		JOIN players p ON gs.player_id = p.id
		JOIN teams t ON gs.team_id = t.id
		WHERE gs.season = $1
		GROUP BY gs.player_id, p.name, p.position, p.jersey_number, gs.team_id, t.abbreviation
		ORDER BY total_stat DESC
		LIMIT $2
	`, statColumn)

	rows, err := pool.Query(ctx, query, season, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats leaders: %w", err)
	}
	defer rows.Close()

	var leaders []models.PlayerStatLeader
	for rows.Next() {
		var l models.PlayerStatLeader
		var secondaryStat, tertiaryStat int

		if err := rows.Scan(
			&l.PlayerID, &l.PlayerName, &l.Position, &l.JerseyNumber,
			&l.TeamID, &l.TeamAbbr, &l.TotalStat, &secondaryStat, &tertiaryStat,
			&l.GamesPlayed,
		); err != nil {
			return nil, fmt.Errorf("failed to scan leader: %w", err)
		}

		l.Category = category
		leaders = append(leaders, l)
	}

	return leaders, nil
}