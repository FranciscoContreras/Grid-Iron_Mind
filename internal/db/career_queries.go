package db

import (
	"context"
	"fmt"

	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

// CareerQueries handles database operations for player career data
type CareerQueries struct{}

// GetPlayerCareerStats retrieves all career statistics for a player
func (q *CareerQueries) GetPlayerCareerStats(ctx context.Context, playerID uuid.UUID) ([]models.PlayerCareerStats, error) {
	query := `
		SELECT
			pcs.id, pcs.player_id, pcs.season, pcs.team_id,
			pcs.games_played, pcs.games_started,
			pcs.passing_yards, pcs.passing_tds, pcs.passing_ints,
			pcs.passing_completions, pcs.passing_attempts, pcs.passing_rating,
			pcs.rushing_yards, pcs.rushing_tds, pcs.rushing_attempts, pcs.rushing_long,
			pcs.receptions, pcs.receiving_yards, pcs.receiving_tds,
			pcs.receiving_targets, pcs.receiving_long,
			pcs.tackles, pcs.sacks, pcs.interceptions,
			pcs.forced_fumbles, pcs.fumble_recoveries, pcs.passes_defended,
			pcs.field_goals_made, pcs.field_goals_attempted,
			pcs.extra_points_made, pcs.extra_points_attempted,
			pcs.created_at, pcs.updated_at,
			t.id as team_id, t.name as team_name, t.abbreviation as team_abbreviation,
			t.city as team_city
		FROM player_career_stats pcs
		LEFT JOIN teams t ON pcs.team_id = t.id
		WHERE pcs.player_id = $1
		ORDER BY pcs.season DESC
	`

	rows, err := pool.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query career stats: %w", err)
	}
	defer rows.Close()

	var stats []models.PlayerCareerStats
	for rows.Next() {
		var s models.PlayerCareerStats
		var teamID *uuid.UUID
		var teamName, teamAbbr, teamCity *string

		err := rows.Scan(
			&s.ID, &s.PlayerID, &s.Season, &s.TeamID,
			&s.GamesPlayed, &s.GamesStarted,
			&s.PassingYards, &s.PassingTDs, &s.PassingInts,
			&s.PassingCompletions, &s.PassingAttempts, &s.PassingRating,
			&s.RushingYards, &s.RushingTDs, &s.RushingAttempts, &s.RushingLong,
			&s.Receptions, &s.ReceivingYards, &s.ReceivingTDs,
			&s.ReceivingTargets, &s.ReceivingLong,
			&s.Tackles, &s.Sacks, &s.Interceptions,
			&s.ForcedFumbles, &s.FumbleRecoveries, &s.PassesDefended,
			&s.FieldGoalsMade, &s.FieldGoalsAttempted,
			&s.ExtraPointsMade, &s.ExtraPointsAttempted,
			&s.CreatedAt, &s.UpdatedAt,
			&teamID, &teamName, &teamAbbr, &teamCity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan career stats: %w", err)
		}

		// Only populate team if we have team data from JOIN
		if teamID != nil && teamName != nil {
			s.Team = &models.Team{
				ID:           *teamID,
				Name:         *teamName,
				Abbreviation: *teamAbbr,
				City:         *teamCity,
			}
		}

		stats = append(stats, s)
	}

	return stats, nil
}

// GetPlayerCareerStatsBySeason retrieves career stats for a specific season
func (q *CareerQueries) GetPlayerCareerStatsBySeason(ctx context.Context, playerID uuid.UUID, season int) (*models.PlayerCareerStats, error) {
	query := `
		SELECT
			pcs.id, pcs.player_id, pcs.season, pcs.team_id,
			pcs.games_played, pcs.games_started,
			pcs.passing_yards, pcs.passing_tds, pcs.passing_ints,
			pcs.passing_completions, pcs.passing_attempts, pcs.passing_rating,
			pcs.rushing_yards, pcs.rushing_tds, pcs.rushing_attempts, pcs.rushing_long,
			pcs.receptions, pcs.receiving_yards, pcs.receiving_tds,
			pcs.receiving_targets, pcs.receiving_long,
			pcs.tackles, pcs.sacks, pcs.interceptions,
			pcs.forced_fumbles, pcs.fumble_recoveries, pcs.passes_defended,
			pcs.field_goals_made, pcs.field_goals_attempted,
			pcs.extra_points_made, pcs.extra_points_attempted,
			pcs.created_at, pcs.updated_at
		FROM player_career_stats pcs
		WHERE pcs.player_id = $1 AND pcs.season = $2
	`

	var s models.PlayerCareerStats
	err := pool.QueryRow(ctx, query, playerID, season).Scan(
		&s.ID, &s.PlayerID, &s.Season, &s.TeamID,
		&s.GamesPlayed, &s.GamesStarted,
		&s.PassingYards, &s.PassingTDs, &s.PassingInts,
		&s.PassingCompletions, &s.PassingAttempts, &s.PassingRating,
		&s.RushingYards, &s.RushingTDs, &s.RushingAttempts, &s.RushingLong,
		&s.Receptions, &s.ReceivingYards, &s.ReceivingTDs,
		&s.ReceivingTargets, &s.ReceivingLong,
		&s.Tackles, &s.Sacks, &s.Interceptions,
		&s.ForcedFumbles, &s.FumbleRecoveries, &s.PassesDefended,
		&s.FieldGoalsMade, &s.FieldGoalsAttempted,
		&s.ExtraPointsMade, &s.ExtraPointsAttempted,
		&s.CreatedAt, &s.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get career stats for season: %w", err)
	}

	return &s, nil
}

// UpsertPlayerCareerStats inserts or updates career stats for a player/season
func (q *CareerQueries) UpsertPlayerCareerStats(ctx context.Context, stats *models.PlayerCareerStats) error {
	query := `
		INSERT INTO player_career_stats (
			player_id, season, team_id, games_played, games_started,
			passing_yards, passing_tds, passing_ints, passing_completions, passing_attempts, passing_rating,
			rushing_yards, rushing_tds, rushing_attempts, rushing_long,
			receptions, receiving_yards, receiving_tds, receiving_targets, receiving_long,
			tackles, sacks, interceptions, forced_fumbles, fumble_recoveries, passes_defended,
			field_goals_made, field_goals_attempted, extra_points_made, extra_points_attempted
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23, $24, $25, $26,
			$27, $28, $29, $30
		)
		ON CONFLICT (player_id, season)
		DO UPDATE SET
			team_id = EXCLUDED.team_id,
			games_played = EXCLUDED.games_played,
			games_started = EXCLUDED.games_started,
			passing_yards = EXCLUDED.passing_yards,
			passing_tds = EXCLUDED.passing_tds,
			passing_ints = EXCLUDED.passing_ints,
			passing_completions = EXCLUDED.passing_completions,
			passing_attempts = EXCLUDED.passing_attempts,
			passing_rating = EXCLUDED.passing_rating,
			rushing_yards = EXCLUDED.rushing_yards,
			rushing_tds = EXCLUDED.rushing_tds,
			rushing_attempts = EXCLUDED.rushing_attempts,
			rushing_long = EXCLUDED.rushing_long,
			receptions = EXCLUDED.receptions,
			receiving_yards = EXCLUDED.receiving_yards,
			receiving_tds = EXCLUDED.receiving_tds,
			receiving_targets = EXCLUDED.receiving_targets,
			receiving_long = EXCLUDED.receiving_long,
			tackles = EXCLUDED.tackles,
			sacks = EXCLUDED.sacks,
			interceptions = EXCLUDED.interceptions,
			forced_fumbles = EXCLUDED.forced_fumbles,
			fumble_recoveries = EXCLUDED.fumble_recoveries,
			passes_defended = EXCLUDED.passes_defended,
			field_goals_made = EXCLUDED.field_goals_made,
			field_goals_attempted = EXCLUDED.field_goals_attempted,
			extra_points_made = EXCLUDED.extra_points_made,
			extra_points_attempted = EXCLUDED.extra_points_attempted,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := pool.Exec(ctx, query,
		stats.PlayerID, stats.Season, stats.TeamID, stats.GamesPlayed, stats.GamesStarted,
		stats.PassingYards, stats.PassingTDs, stats.PassingInts, stats.PassingCompletions, stats.PassingAttempts, stats.PassingRating,
		stats.RushingYards, stats.RushingTDs, stats.RushingAttempts, stats.RushingLong,
		stats.Receptions, stats.ReceivingYards, stats.ReceivingTDs, stats.ReceivingTargets, stats.ReceivingLong,
		stats.Tackles, stats.Sacks, stats.Interceptions, stats.ForcedFumbles, stats.FumbleRecoveries, stats.PassesDefended,
		stats.FieldGoalsMade, stats.FieldGoalsAttempted, stats.ExtraPointsMade, stats.ExtraPointsAttempted,
	)

	return err
}

// GetPlayerTeamHistory retrieves the team history for a player
func (q *CareerQueries) GetPlayerTeamHistory(ctx context.Context, playerID uuid.UUID) ([]models.PlayerTeamHistory, error) {
	query := `
		SELECT
			pth.id, pth.player_id, pth.team_id, pth.season_start, pth.season_end,
			pth.position, pth.jersey_number, pth.is_current, pth.created_at,
			t.id, t.name, t.abbreviation, t.city, t.conference, t.division
		FROM player_team_history pth
		JOIN teams t ON pth.team_id = t.id
		WHERE pth.player_id = $1
		ORDER BY pth.season_start DESC
	`

	rows, err := pool.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query team history: %w", err)
	}
	defer rows.Close()

	var history []models.PlayerTeamHistory
	for rows.Next() {
		var h models.PlayerTeamHistory
		var team models.Team

		err := rows.Scan(
			&h.ID, &h.PlayerID, &h.TeamID, &h.SeasonStart, &h.SeasonEnd,
			&h.Position, &h.JerseyNumber, &h.IsCurrent, &h.CreatedAt,
			&team.ID, &team.Name, &team.Abbreviation, &team.City, &team.Conference, &team.Division,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team history: %w", err)
		}

		h.Team = &team
		history = append(history, h)
	}

	return history, nil
}

// UpsertPlayerTeamHistory inserts or updates player team history
func (q *CareerQueries) UpsertPlayerTeamHistory(ctx context.Context, history *models.PlayerTeamHistory) error {
	query := `
		INSERT INTO player_team_history (
			player_id, team_id, season_start, season_end, position, jersey_number, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (player_id, team_id, season_start)
		DO UPDATE SET
			season_end = EXCLUDED.season_end,
			position = EXCLUDED.position,
			jersey_number = EXCLUDED.jersey_number,
			is_current = EXCLUDED.is_current
	`

	_, err := pool.Exec(ctx, query,
		history.PlayerID, history.TeamID, history.SeasonStart, history.SeasonEnd,
		history.Position, history.JerseyNumber, history.IsCurrent,
	)

	return err
}