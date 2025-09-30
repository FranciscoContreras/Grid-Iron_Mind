package db

import (
	"context"
	"fmt"
	"time"

	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PlayerQueries handles all player-related database operations
type PlayerQueries struct{}

// ListPlayers retrieves players with filters and pagination
func (q *PlayerQueries) ListPlayers(ctx context.Context, filters PlayerFilters) ([]models.Player, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Build query with filters
	query := `
		SELECT p.id, p.nfl_id, p.name, p.position, p.team_id, p.jersey_number,
		       p.height_inches, p.weight_pounds, p.birth_date, p.college,
		       p.draft_year, p.draft_round, p.draft_pick, p.status,
		       p.created_at, p.updated_at
		FROM players p
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM players p WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filters.Position != "" {
		query += fmt.Sprintf(" AND p.position = $%d", argCount)
		countQuery += fmt.Sprintf(" AND p.position = $%d", argCount)
		args = append(args, filters.Position)
		argCount++
	}

	if filters.TeamID != uuid.Nil {
		query += fmt.Sprintf(" AND p.team_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND p.team_id = $%d", argCount)
		args = append(args, filters.TeamID)
		argCount++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND p.status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND p.status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	// Get total count
	var total int
	err := pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count players: %w", err)
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY p.name LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filters.Limit, filters.Offset)

	// Execute query
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var p models.Player
		err := rows.Scan(
			&p.ID, &p.NFLID, &p.Name, &p.Position, &p.TeamID, &p.JerseyNumber,
			&p.HeightInches, &p.WeightPounds, &p.BirthDate, &p.College,
			&p.DraftYear, &p.DraftRound, &p.DraftPick, &p.Status,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, p)
	}

	return players, total, nil
}

// GetPlayerByID retrieves a single player by ID
func (q *PlayerQueries) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT p.id, p.nfl_id, p.name, p.position, p.team_id, p.jersey_number,
		       p.height_inches, p.weight_pounds, p.birth_date, p.college,
		       p.draft_year, p.draft_round, p.draft_pick, p.status,
		       p.created_at, p.updated_at
		FROM players p
		WHERE p.id = $1
	`

	var p models.Player
	err := pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.NFLID, &p.Name, &p.Position, &p.TeamID, &p.JerseyNumber,
		&p.HeightInches, &p.WeightPounds, &p.BirthDate, &p.College,
		&p.DraftYear, &p.DraftRound, &p.DraftPick, &p.Status,
		&p.CreatedAt, &p.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	return &p, nil
}

// TeamQueries handles all team-related database operations
type TeamQueries struct{}

// ListTeams retrieves all teams
func (q *TeamQueries) ListTeams(ctx context.Context) ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, nfl_id, name, abbreviation, city, conference, division,
		       stadium, created_at, updated_at
		FROM teams
		ORDER BY name
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	teams := []models.Team{}
	for rows.Next() {
		var t models.Team
		err := rows.Scan(
			&t.ID, &t.NFLID, &t.Name, &t.Abbreviation, &t.City,
			&t.Conference, &t.Division, &t.Stadium, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, t)
	}

	return teams, nil
}

// GetTeamByID retrieves a single team by ID
func (q *TeamQueries) GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, nfl_id, name, abbreviation, city, conference, division,
		       stadium, created_at, updated_at
		FROM teams
		WHERE id = $1
	`

	var t models.Team
	err := pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.NFLID, &t.Name, &t.Abbreviation, &t.City,
		&t.Conference, &t.Division, &t.Stadium, &t.CreatedAt, &t.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &t, nil
}

// GetTeamPlayers retrieves all players for a team
func (q *TeamQueries) GetTeamPlayers(ctx context.Context, teamID uuid.UUID) ([]models.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT p.id, p.nfl_id, p.name, p.position, p.team_id, p.jersey_number,
		       p.height_inches, p.weight_pounds, p.birth_date, p.college,
		       p.draft_year, p.draft_round, p.draft_pick, p.status,
		       p.created_at, p.updated_at
		FROM players p
		WHERE p.team_id = $1
		ORDER BY p.jersey_number, p.name
	`

	rows, err := pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query team players: %w", err)
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var p models.Player
		err := rows.Scan(
			&p.ID, &p.NFLID, &p.Name, &p.Position, &p.TeamID, &p.JerseyNumber,
			&p.HeightInches, &p.WeightPounds, &p.BirthDate, &p.College,
			&p.DraftYear, &p.DraftRound, &p.DraftPick, &p.Status,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, p)
	}

	return players, nil
}

// PlayerFilters holds filter parameters for player queries
type PlayerFilters struct {
	Position string
	TeamID   uuid.UUID
	Status   string
	Limit    int
	Offset   int
}