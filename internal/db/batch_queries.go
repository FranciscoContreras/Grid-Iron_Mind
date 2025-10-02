package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

// BatchQueries handles batch loading operations to prevent N+1 queries
type BatchQueries struct{}

// GetTeamsByIDs retrieves multiple teams by their IDs in a single query
func (q *BatchQueries) GetTeamsByIDs(ctx context.Context, teamIDs []uuid.UUID) (map[uuid.UUID]*models.Team, error) {
	if len(teamIDs) == 0 {
		return map[uuid.UUID]*models.Team{}, nil
	}

	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(teamIDs))
	args := make([]interface{}, len(teamIDs))
	for i, id := range teamIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, nfl_id, name, abbreviation, city, conference, division, stadium
		FROM teams
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	teams := make(map[uuid.UUID]*models.Team)
	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID, &team.NFLID, &team.Name, &team.Abbreviation,
			&team.City, &team.Conference, &team.Division, &team.Stadium,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams[team.ID] = &team
	}

	return teams, nil
}

// GetPlayersByIDs retrieves multiple players by their IDs in a single query
func (q *BatchQueries) GetPlayersByIDs(ctx context.Context, playerIDs []uuid.UUID) (map[uuid.UUID]*models.Player, error) {
	if len(playerIDs) == 0 {
		return map[uuid.UUID]*models.Player{}, nil
	}

	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	// Build placeholders
	placeholders := make([]string, len(playerIDs))
	args := make([]interface{}, len(playerIDs))
	for i, id := range playerIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, nfl_id, name, position, team_id, jersey_number, status
		FROM players
		WHERE id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	players := make(map[uuid.UUID]*models.Player)
	for rows.Next() {
		var player models.Player
		err := rows.Scan(
			&player.ID, &player.NFLID, &player.Name, &player.Position,
			&player.TeamID, &player.JerseyNumber, &player.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players[player.ID] = &player
	}

	return players, nil
}

// GetGameStatsWithDetails retrieves game stats with player and team details in a single query
func (q *BatchQueries) GetGameStatsWithDetails(ctx context.Context, gameID uuid.UUID) ([]map[string]interface{}, error) {
	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT
			gs.id,
			gs.player_id,
			p.name as player_name,
			p.position,
			p.jersey_number,
			t.abbreviation as team_abbr,
			t.name as team_name,
			gs.passing_yards,
			gs.rushing_yards,
			gs.receiving_yards,
			gs.touchdowns,
			gs.interceptions,
			gs.completions,
			gs.attempts,
			gs.targets,
			gs.receptions
		FROM game_stats gs
		JOIN players p ON gs.player_id = p.id
		LEFT JOIN teams t ON p.team_id = t.id
		WHERE gs.game_id = $1
		ORDER BY p.position, p.name
	`

	rows, err := pool.Query(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to query game stats: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var (
			id             uuid.UUID
			playerID       uuid.UUID
			playerName     string
			position       string
			jerseyNumber   *int
			teamAbbr       *string
			teamName       *string
			passingYards   int
			rushingYards   int
			receivingYards int
			touchdowns     int
			interceptions  int
			completions    int
			attempts       int
			targets        int
			receptions     int
		)

		err := rows.Scan(
			&id, &playerID, &playerName, &position, &jerseyNumber,
			&teamAbbr, &teamName,
			&passingYards, &rushingYards, &receivingYards,
			&touchdowns, &interceptions, &completions, &attempts,
			&targets, &receptions,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game stat: %w", err)
		}

		stat := map[string]interface{}{
			"id":              id,
			"player_id":       playerID,
			"player_name":     playerName,
			"position":        position,
			"jersey_number":   jerseyNumber,
			"team_abbr":       teamAbbr,
			"team_name":       teamName,
			"passing_yards":   passingYards,
			"rushing_yards":   rushingYards,
			"receiving_yards": receivingYards,
			"touchdowns":      touchdowns,
			"interceptions":   interceptions,
			"completions":     completions,
			"attempts":        attempts,
			"targets":         targets,
			"receptions":      receptions,
		}

		results = append(results, stat)
	}

	return results, nil
}

// GetGamesWithTeamDetails retrieves games with home/away team details in a single query
func (q *BatchQueries) GetGamesWithTeamDetails(ctx context.Context, season int, week int) ([]map[string]interface{}, error) {
	pool := GetPool()
	if pool == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}

	query := `
		SELECT
			g.id,
			g.nfl_game_id,
			g.game_date,
			g.season,
			g.week,
			g.home_score,
			g.away_score,
			g.status,
			ht.id as home_team_id,
			ht.name as home_team_name,
			ht.abbreviation as home_team_abbr,
			at.id as away_team_id,
			at.name as away_team_name,
			at.abbreviation as away_team_abbr
		FROM games g
		JOIN teams ht ON g.home_team_id = ht.id
		JOIN teams at ON g.away_team_id = at.id
		WHERE g.season = $1 AND g.week = $2
		ORDER BY g.game_date
	`

	rows, err := pool.Query(ctx, query, season, week)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var game struct {
			ID           uuid.UUID
			NFLGameID    string
			GameDate     interface{}
			Season       int
			Week         int
			HomeScore    *int
			AwayScore    *int
			Status       string
			HomeTeamID   uuid.UUID
			HomeTeamName string
			HomeTeamAbbr string
			AwayTeamID   uuid.UUID
			AwayTeamName string
			AwayTeamAbbr string
		}

		err := rows.Scan(
			&game.ID, &game.NFLGameID, &game.GameDate,
			&game.Season, &game.Week, &game.HomeScore, &game.AwayScore, &game.Status,
			&game.HomeTeamID, &game.HomeTeamName, &game.HomeTeamAbbr,
			&game.AwayTeamID, &game.AwayTeamName, &game.AwayTeamAbbr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}

		result := map[string]interface{}{
			"id":          game.ID,
			"nfl_game_id": game.NFLGameID,
			"game_date":   game.GameDate,
			"season":      game.Season,
			"week":        game.Week,
			"home_score":  game.HomeScore,
			"away_score":  game.AwayScore,
			"status":      game.Status,
			"home_team": map[string]interface{}{
				"id":           game.HomeTeamID,
				"name":         game.HomeTeamName,
				"abbreviation": game.HomeTeamAbbr,
			},
			"away_team": map[string]interface{}{
				"id":           game.AwayTeamID,
				"name":         game.AwayTeamName,
				"abbreviation": game.AwayTeamAbbr,
			},
		}

		results = append(results, result)
	}

	return results, nil
}

// BatchLoadTeamsForPlayers adds team details to a slice of players
func (q *BatchQueries) BatchLoadTeamsForPlayers(ctx context.Context, players []models.Player) ([]models.Player, error) {
	if len(players) == 0 {
		return players, nil
	}

	// Collect unique team IDs
	teamIDsMap := make(map[uuid.UUID]bool)
	for _, player := range players {
		if player.TeamID != nil && *player.TeamID != uuid.Nil {
			teamIDsMap[*player.TeamID] = true
		}
	}

	// Convert map to slice
	teamIDs := make([]uuid.UUID, 0, len(teamIDsMap))
	for id := range teamIDsMap {
		teamIDs = append(teamIDs, id)
	}

	// Batch load teams (for future use when Player model includes Team details)
	_, err := q.GetTeamsByIDs(ctx, teamIDs)
	if err != nil {
		return nil, err
	}

	// NOTE: This loads teams but doesn't attach them to players yet
	// To use this, extend the Player model to include a Team field
	// Example:
	//   type Player struct {
	//     ...existing fields...
	//     Team *Team `json:"team,omitempty"`
	//   }

	return players, nil
}
