package db

import (
	"context"
	"fmt"
	"strings"
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
	baseSelect := `
		SELECT p.id, p.nfl_id, p.name, p.position, p.team_id, p.jersey_number,
		       p.height_inches, p.weight_pounds, p.birth_date, p.college,
		       p.draft_year, p.draft_round, p.draft_pick, p.status,
		       p.created_at, p.updated_at
	`

	// Add relevance scoring if search is provided
	var query string
	if filters.Search != "" {
		query = baseSelect + `,
		CASE
			WHEN LOWER(p.name) = $1 THEN 100
			WHEN LOWER(p.name) LIKE $2 THEN 80
			WHEN LOWER(p.name) LIKE $3 THEN 60
			ELSE 0
		END as relevance_score
		FROM players p
		WHERE LOWER(p.name) LIKE $3
		`
	} else {
		query = baseSelect + `
		FROM players p
		WHERE 1=1
		`
	}

	countQuery := `SELECT COUNT(*) FROM players p WHERE 1=1`
	args := []interface{}{}
	argCount := 1

	// Add search filter
	if filters.Search != "" {
		searchLower := strings.ToLower(strings.TrimSpace(filters.Search))
		// Add search args for relevance scoring
		args = append(args, searchLower)                  // $1 - exact match
		args = append(args, searchLower+"%")              // $2 - starts with
		args = append(args, "%"+searchLower+"%")          // $3 - contains
		argCount = 4 // Next arg starts at $4

		// Add search to count query
		countQuery += fmt.Sprintf(" AND LOWER(p.name) LIKE $%d", 1)
	}

	// Apply other filters
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

	// Get total count (use separate args for count query)
	var total int
	var countArgs []interface{}
	if filters.Search != "" {
		searchLower := strings.ToLower(strings.TrimSpace(filters.Search))
		countArgs = append(countArgs, "%"+searchLower+"%")
	}
	countArgIdx := len(countArgs) + 1
	if filters.Position != "" {
		countArgs = append(countArgs, filters.Position)
	}
	if filters.TeamID != uuid.Nil {
		countArgs = append(countArgs, filters.TeamID)
	}
	if filters.Status != "" {
		countArgs = append(countArgs, filters.Status)
	}

	err := pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count players: %w", err)
	}

	// Add ordering and pagination
	if filters.Search != "" {
		// Order by relevance score first, then name
		query += fmt.Sprintf(" ORDER BY relevance_score DESC, p.name ASC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	} else {
		// No search - just order by name
		query += fmt.Sprintf(" ORDER BY p.name ASC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	}
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
		var relevanceScore *int // Ignore relevance score in scan

		if filters.Search != "" {
			err := rows.Scan(
				&p.ID, &p.NFLID, &p.Name, &p.Position, &p.TeamID, &p.JerseyNumber,
				&p.HeightInches, &p.WeightPounds, &p.BirthDate, &p.College,
				&p.DraftYear, &p.DraftRound, &p.DraftPick, &p.Status,
				&p.CreatedAt, &p.UpdatedAt, &relevanceScore,
			)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to scan player: %w", err)
			}
		} else {
			err := rows.Scan(
				&p.ID, &p.NFLID, &p.Name, &p.Position, &p.TeamID, &p.JerseyNumber,
				&p.HeightInches, &p.WeightPounds, &p.BirthDate, &p.College,
				&p.DraftYear, &p.DraftRound, &p.DraftPick, &p.Status,
				&p.CreatedAt, &p.UpdatedAt,
			)
			if err != nil {
				return nil, 0, fmt.Errorf("failed to scan player: %w", err)
			}
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

// GetByNFLID retrieves a player by their NFL ID
func (q *PlayerQueries) GetByNFLID(ctx context.Context, nflID int) (*models.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT p.id, p.nfl_id, p.name, p.position, p.team_id, p.jersey_number,
		       p.height_inches, p.weight_pounds, p.birth_date, p.college,
		       p.draft_year, p.draft_round, p.draft_pick, p.status,
		       p.created_at, p.updated_at
		FROM players p
		WHERE p.nfl_id = $1
	`

	var p models.Player
	err := pool.QueryRow(ctx, query, nflID).Scan(
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

// List retrieves all teams
func (q *TeamQueries) List(ctx context.Context) ([]models.Team, error) {
	return q.ListTeams(ctx)
}

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

// GetByNFLID retrieves a team by their NFL ID
func (q *TeamQueries) GetByNFLID(ctx context.Context, nflID int) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT id, nfl_id, name, abbreviation, city, conference, division,
		       stadium, created_at, updated_at
		FROM teams
		WHERE nfl_id = $1
	`

	var t models.Team
	err := pool.QueryRow(ctx, query, nflID).Scan(
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
	Search   string
	Position string
	TeamID   uuid.UUID
	Status   string
	Limit    int
	Offset   int
}