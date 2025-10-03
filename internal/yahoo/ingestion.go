package yahoo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IngestionService handles importing Yahoo Fantasy data into the database
type IngestionService struct {
	client *Client
	dbPool *pgxpool.Pool
}

// NewIngestionService creates a new Yahoo data ingestion service
func NewIngestionService(client *Client) *IngestionService {
	return &IngestionService{
		client: client,
		dbPool: db.GetPool(),
	}
}

// SyncPlayerRankings syncs player rankings for a specific week
func (s *IngestionService) SyncPlayerRankings(ctx context.Context, season, week int, position string) error {
	log.Printf("[YAHOO] Syncing player rankings for %s, season %d, week %d", position, season, week)

	// Fetch rankings from Yahoo API
	rankings, err := s.client.FetchPlayerRankings(ctx, position, week)
	if err != nil {
		return fmt.Errorf("failed to fetch rankings: %w", err)
	}

	if rankings == nil || len(rankings.Players) == 0 {
		log.Printf("[YAHOO] No rankings found for %s week %d", position, week)
		return nil
	}

	// Upsert each player's ranking
	count := 0
	for rank, player := range rankings.Players {
		if err := s.upsertPlayerRanking(ctx, player, season, week, rank+1); err != nil {
			log.Printf("[YAHOO] Failed to upsert ranking for player %s: %v", player.Name.Full, err)
			continue
		}
		count++
	}

	log.Printf("[YAHOO] Synced %d %s rankings for week %d", count, position, week)
	return nil
}

// upsertPlayerRanking inserts or updates a player's weekly ranking
func (s *IngestionService) upsertPlayerRanking(ctx context.Context, player Player, season, week, rank int) error {
	// Find player in our database by Yahoo player key or name
	var playerID uuid.UUID

	// Try to find by Yahoo player key first (stored in a custom field or by matching)
	// For now, we'll match by name and team
	query := `
		SELECT p.id
		FROM players p
		JOIN teams t ON p.team_id = t.id
		WHERE LOWER(p.name) = LOWER($1)
		AND t.abbreviation = $2
		LIMIT 1
	`

	err := s.dbPool.QueryRow(ctx, query, player.Name.Full, player.EditorialTeamAbbr).Scan(&playerID)
	if err != nil {
		// Player not found - this could be a new player or name mismatch
		log.Printf("[YAHOO] Player not found in database: %s (%s)", player.Name.Full, player.EditorialTeamAbbr)
		return nil // Skip rather than fail
	}

	// Upsert ranking
	upsertQuery := `
		INSERT INTO yahoo_player_rankings (
			player_id, yahoo_player_key, season, week, position,
			overall_rank, position_rank, percent_owned, percent_started,
			scoring_format, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (player_id, season, week, scoring_format)
		DO UPDATE SET
			yahoo_player_key = EXCLUDED.yahoo_player_key,
			overall_rank = EXCLUDED.overall_rank,
			position_rank = EXCLUDED.position_rank,
			percent_owned = EXCLUDED.percent_owned,
			percent_started = EXCLUDED.percent_started,
			updated_at = NOW()
	`

	_, err = s.dbPool.Exec(ctx, upsertQuery,
		playerID,
		player.PlayerKey,
		season,
		week,
		player.DisplayPosition,
		rank,
		nil, // position_rank - calculate separately
		player.PercentOwned,
		nil, // percent_started - not in basic response
		"standard",
	)

	return err
}

// SyncPlayerProjections syncs weekly projections for players
func (s *IngestionService) SyncPlayerProjections(ctx context.Context, season, week int, playerKeys []string) error {
	if len(playerKeys) == 0 {
		return fmt.Errorf("no player keys provided")
	}

	log.Printf("[YAHOO] Syncing projections for %d players, week %d", len(playerKeys), week)

	// Fetch projections from Yahoo (batch request)
	projections, err := s.client.FetchPlayerProjections(ctx, playerKeys, week)
	if err != nil {
		return fmt.Errorf("failed to fetch projections: %w", err)
	}

	if projections == nil || len(projections.Players) == 0 {
		log.Printf("[YAHOO] No projections found for week %d", week)
		return nil
	}

	// Upsert each player's projections
	count := 0
	for _, player := range projections.Players {
		if err := s.upsertPlayerProjection(ctx, player, season, week); err != nil {
			log.Printf("[YAHOO] Failed to upsert projection for player %s: %v", player.Name.Full, err)
			continue
		}
		count++
	}

	log.Printf("[YAHOO] Synced %d player projections for week %d", count, week)
	return nil
}

// upsertPlayerProjection inserts or updates a player's weekly projection
func (s *IngestionService) upsertPlayerProjection(ctx context.Context, player Player, season, week int) error {
	// Find player in database
	var playerID uuid.UUID
	query := `
		SELECT p.id
		FROM players p
		JOIN teams t ON p.team_id = t.id
		WHERE LOWER(p.name) = LOWER($1)
		AND t.abbreviation = $2
		LIMIT 1
	`

	err := s.dbPool.QueryRow(ctx, query, player.Name.Full, player.EditorialTeamAbbr).Scan(&playerID)
	if err != nil {
		return nil // Skip if player not found
	}

	// Extract projected stats from player stats
	var projectedPoints float64
	var passingYards, passingTDs, interceptions int
	var rushingYards, rushingTDs int
	var receptions, receivingYards, receivingTDs int

	if player.PlayerProjected != nil && len(player.PlayerProjected.Stats) > 0 {
		for _, stat := range player.PlayerProjected.Stats {
			switch stat.StatID {
			case StatPassingYards:
				passingYards = int(stat.Value)
			case StatPassingTouchdowns:
				passingTDs = int(stat.Value)
			case StatInterceptions:
				interceptions = int(stat.Value)
			case StatRushingYards:
				rushingYards = int(stat.Value)
			case StatRushingTouchdowns:
				rushingTDs = int(stat.Value)
			case StatReceptions:
				receptions = int(stat.Value)
			case StatReceivingYards:
				receivingYards = int(stat.Value)
			case StatReceivingTouchdowns:
				receivingTDs = int(stat.Value)
			}
		}
	}

	if player.PlayerPoints != nil {
		projectedPoints = player.PlayerPoints.Total
	}

	// Upsert projection
	upsertQuery := `
		INSERT INTO yahoo_player_projections (
			player_id, yahoo_player_key, season, week,
			projected_points,
			projected_passing_yards, projected_passing_tds, projected_interceptions,
			projected_rushing_yards, projected_rushing_tds,
			projected_receptions, projected_receiving_yards, projected_receiving_tds,
			scoring_format, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		ON CONFLICT (player_id, season, week, scoring_format)
		DO UPDATE SET
			yahoo_player_key = EXCLUDED.yahoo_player_key,
			projected_points = EXCLUDED.projected_points,
			projected_passing_yards = EXCLUDED.projected_passing_yards,
			projected_passing_tds = EXCLUDED.projected_passing_tds,
			projected_interceptions = EXCLUDED.projected_interceptions,
			projected_rushing_yards = EXCLUDED.projected_rushing_yards,
			projected_rushing_tds = EXCLUDED.projected_rushing_tds,
			projected_receptions = EXCLUDED.projected_receptions,
			projected_receiving_yards = EXCLUDED.projected_receiving_yards,
			projected_receiving_tds = EXCLUDED.projected_receiving_tds,
			updated_at = NOW()
	`

	_, err = s.dbPool.Exec(ctx, upsertQuery,
		playerID, player.PlayerKey, season, week,
		projectedPoints,
		passingYards, passingTDs, interceptions,
		rushingYards, rushingTDs,
		receptions, receivingYards, receivingTDs,
		"standard",
	)

	return err
}

// SyncAllPositionRankings syncs rankings for all major positions
func (s *IngestionService) SyncAllPositionRankings(ctx context.Context, season, week int) error {
	positions := []string{"QB", "RB", "WR", "TE", "K", "DEF"}

	for _, position := range positions {
		if err := s.SyncPlayerRankings(ctx, season, week, position); err != nil {
			log.Printf("[YAHOO] Error syncing %s rankings: %v", position, err)
			// Continue with other positions
			continue
		}
		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// GetTopRankedPlayers retrieves top fantasy players for a week
func (s *IngestionService) GetTopRankedPlayers(ctx context.Context, season, week int, position string, limit int) ([]RankedPlayer, error) {
	query := `
		SELECT
			p.id, p.name, p.position, t.abbreviation as team,
			r.overall_rank, r.position_rank, r.percent_owned,
			r.yahoo_player_key
		FROM yahoo_player_rankings r
		JOIN players p ON r.player_id = p.id
		LEFT JOIN teams t ON p.team_id = t.id
		WHERE r.season = $1 AND r.week = $2
	`

	args := []interface{}{season, week}
	argCount := 3

	if position != "" {
		query += fmt.Sprintf(" AND r.position = $%d", argCount)
		args = append(args, position)
		argCount++
	}

	query += " ORDER BY r.overall_rank ASC LIMIT $" + fmt.Sprintf("%d", argCount)
	args = append(args, limit)

	rows, err := s.dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []RankedPlayer
	for rows.Next() {
		var p RankedPlayer
		err := rows.Scan(
			&p.PlayerID, &p.Name, &p.Position, &p.Team,
			&p.OverallRank, &p.PositionRank, &p.PercentOwned,
			&p.YahooPlayerKey,
		)
		if err != nil {
			log.Printf("[YAHOO] Error scanning ranked player: %v", err)
			continue
		}
		players = append(players, p)
	}

	return players, nil
}

// RankedPlayer represents a fantasy ranked player
type RankedPlayer struct {
	PlayerID       uuid.UUID
	Name           string
	Position       string
	Team           string
	OverallRank    *int
	PositionRank   *int
	PercentOwned   *float64
	YahooPlayerKey string
}

// GetPlayerProjection retrieves a player's projection for a week
func (s *IngestionService) GetPlayerProjection(ctx context.Context, playerID uuid.UUID, season, week int) (*PlayerProjection, error) {
	query := `
		SELECT
			projected_points,
			projected_passing_yards, projected_passing_tds, projected_interceptions,
			projected_rushing_yards, projected_rushing_tds,
			projected_receptions, projected_receiving_yards, projected_receiving_tds
		FROM yahoo_player_projections
		WHERE player_id = $1 AND season = $2 AND week = $3
		AND scoring_format = 'standard'
	`

	var proj PlayerProjection
	err := s.dbPool.QueryRow(ctx, query, playerID, season, week).Scan(
		&proj.ProjectedPoints,
		&proj.PassingYards, &proj.PassingTDs, &proj.Interceptions,
		&proj.RushingYards, &proj.RushingTDs,
		&proj.Receptions, &proj.ReceivingYards, &proj.ReceivingTDs,
	)

	if err != nil {
		return nil, err
	}

	return &proj, nil
}

// PlayerProjection represents a player's fantasy projection
type PlayerProjection struct {
	ProjectedPoints float64
	PassingYards    int
	PassingTDs      int
	Interceptions   int
	RushingYards    int
	RushingTDs      int
	Receptions      int
	ReceivingYards  int
	ReceivingTDs    int
}
