package db

import (
	"context"
	"fmt"

	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

// InjuryQueries handles database operations for player injuries
type InjuryQueries struct{}

// GetPlayerInjuries retrieves all injuries for a player
func (q *InjuryQueries) GetPlayerInjuries(ctx context.Context, playerID uuid.UUID) ([]models.PlayerInjury, error) {
	query := `
		SELECT
			pi.id, pi.player_id, pi.team_id, pi.game_id,
			pi.status, pi.status_abbreviation,
			pi.injury_type, pi.body_location, pi.detail, pi.side,
			pi.injury_date, pi.return_date,
			pi.espn_injury_id, pi.last_updated, pi.created_at,
			p.id, p.name, p.position,
			t.id, t.name, t.abbreviation, t.city
		FROM player_injuries pi
		JOIN players p ON pi.player_id = p.id
		LEFT JOIN teams t ON pi.team_id = t.id
		WHERE pi.player_id = $1
		ORDER BY pi.created_at DESC
	`

	rows, err := pool.Query(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query player injuries: %w", err)
	}
	defer rows.Close()

	var injuries []models.PlayerInjury
	for rows.Next() {
		var inj models.PlayerInjury
		var player models.Player
		var teamID *uuid.UUID
		var teamName, teamAbbr, teamCity *string

		err := rows.Scan(
			&inj.ID, &inj.PlayerID, &inj.TeamID, &inj.GameID,
			&inj.Status, &inj.StatusAbbreviation,
			&inj.InjuryType, &inj.BodyLocation, &inj.Detail, &inj.Side,
			&inj.InjuryDate, &inj.ReturnDate,
			&inj.ESPNInjuryID, &inj.LastUpdated, &inj.CreatedAt,
			&player.ID, &player.Name, &player.Position,
			&teamID, &teamName, &teamAbbr, &teamCity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan injury: %w", err)
		}

		inj.Player = &player

		if teamID != nil && teamName != nil {
			inj.Team = &models.Team{
				ID:           *teamID,
				Name:         *teamName,
				Abbreviation: *teamAbbr,
				City:         *teamCity,
			}
		}

		injuries = append(injuries, inj)
	}

	return injuries, nil
}

// GetTeamInjuries retrieves all current injuries for a team
func (q *InjuryQueries) GetTeamInjuries(ctx context.Context, teamID uuid.UUID) ([]models.PlayerInjury, error) {
	query := `
		SELECT
			pi.id, pi.player_id, pi.team_id, pi.game_id,
			pi.status, pi.status_abbreviation,
			pi.injury_type, pi.body_location, pi.detail, pi.side,
			pi.injury_date, pi.return_date,
			pi.espn_injury_id, pi.last_updated, pi.created_at,
			p.id, p.name, p.position, p.jersey_number
		FROM player_injuries pi
		JOIN players p ON pi.player_id = p.id
		WHERE pi.team_id = $1
		ORDER BY
			CASE pi.status
				WHEN 'Out' THEN 1
				WHEN 'Doubtful' THEN 2
				WHEN 'Questionable' THEN 3
				ELSE 4
			END,
			p.name
	`

	rows, err := pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query team injuries: %w", err)
	}
	defer rows.Close()

	var injuries []models.PlayerInjury
	for rows.Next() {
		var inj models.PlayerInjury
		var player models.Player

		err := rows.Scan(
			&inj.ID, &inj.PlayerID, &inj.TeamID, &inj.GameID,
			&inj.Status, &inj.StatusAbbreviation,
			&inj.InjuryType, &inj.BodyLocation, &inj.Detail, &inj.Side,
			&inj.InjuryDate, &inj.ReturnDate,
			&inj.ESPNInjuryID, &inj.LastUpdated, &inj.CreatedAt,
			&player.ID, &player.Name, &player.Position, &player.JerseyNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan injury: %w", err)
		}

		inj.Player = &player
		injuries = append(injuries, inj)
	}

	return injuries, nil
}

// UpsertPlayerInjury inserts or updates a player injury
func (q *InjuryQueries) UpsertPlayerInjury(ctx context.Context, injury *models.PlayerInjury) error {
	query := `
		INSERT INTO player_injuries (
			player_id, team_id, game_id,
			status, status_abbreviation,
			injury_type, body_location, detail, side,
			injury_date, return_date,
			espn_injury_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (player_id, injury_type, status)
		DO UPDATE SET
			team_id = EXCLUDED.team_id,
			game_id = EXCLUDED.game_id,
			status = EXCLUDED.status,
			status_abbreviation = EXCLUDED.status_abbreviation,
			body_location = EXCLUDED.body_location,
			detail = EXCLUDED.detail,
			side = EXCLUDED.side,
			injury_date = EXCLUDED.injury_date,
			return_date = EXCLUDED.return_date,
			espn_injury_id = EXCLUDED.espn_injury_id,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := pool.Exec(ctx, query,
		injury.PlayerID, injury.TeamID, injury.GameID,
		injury.Status, injury.StatusAbbreviation,
		injury.InjuryType, injury.BodyLocation, injury.Detail, injury.Side,
		injury.InjuryDate, injury.ReturnDate,
		injury.ESPNInjuryID,
	)

	return err
}

// DeleteOldInjuries removes injuries older than specified days that are resolved
func (q *InjuryQueries) DeleteOldInjuries(ctx context.Context, daysOld int) error {
	query := `
		DELETE FROM player_injuries
		WHERE created_at < NOW() - INTERVAL '%d days'
		AND status IN ('Healthy', 'Active', 'Cleared')
	`

	_, err := pool.Exec(ctx, fmt.Sprintf(query, daysOld))
	return err
}
