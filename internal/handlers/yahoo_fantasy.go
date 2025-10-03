package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/utils"
	"github.com/francisco/gridironmind/internal/yahoo"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

// YahooFantasyHandler handles Yahoo Fantasy data requests
type YahooFantasyHandler struct {
	ingestionService *yahoo.IngestionService
}

// NewYahooFantasyHandler creates a new Yahoo Fantasy handler
func NewYahooFantasyHandler(ingestionService *yahoo.IngestionService) *YahooFantasyHandler {
	return &YahooFantasyHandler{
		ingestionService: ingestionService,
	}
}

// HandlePlayerRankings returns fantasy player rankings
func (h *YahooFantasyHandler) HandlePlayerRankings(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	position := query.Get("position")
	limit := parseIntParam(query.Get("limit"), 50)

	// Get season and week (default to current)
	seasonInfo := utils.GetCurrentSeason()
	season := parseIntParam(query.Get("season"), seasonInfo.Year)
	week := parseIntParam(query.Get("week"), seasonInfo.CurrentWeek)

	// Validate parameters
	if limit > 100 {
		limit = 100
	}

	// Get ranked players
	players, err := h.ingestionService.GetTopRankedPlayers(r.Context(), season, week, position, limit)
	if err != nil {
		response.InternalError(w, "Failed to fetch rankings")
		return
	}

	response.Success(w, map[string]interface{}{
		"season":   season,
		"week":     week,
		"position": position,
		"count":    len(players),
		"players":  players,
	})
}

// HandlePlayerProjection returns a player's weekly fantasy projection
func (h *YahooFantasyHandler) HandlePlayerProjection(w http.ResponseWriter, r *http.Request) {
	// Extract player ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/fantasy/projections/")
	playerIDStr := strings.Split(path, "/")[0]

	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid player ID")
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	seasonInfo := utils.GetCurrentSeason()
	season := parseIntParam(query.Get("season"), seasonInfo.Year)
	week := parseIntParam(query.Get("week"), seasonInfo.CurrentWeek)

	// Get projection
	projection, err := h.ingestionService.GetPlayerProjection(r.Context(), playerID, season, week)
	if err != nil {
		// Check if player exists
		pool := db.GetPool()
		var exists bool
		checkQuery := "SELECT EXISTS(SELECT 1 FROM players WHERE id = $1)"
		_ = pool.QueryRow(r.Context(), checkQuery, playerID).Scan(&exists)

		if !exists {
			response.NotFound(w, "Player")
			return
		}

		// Player exists but no projection data
		response.Success(w, map[string]interface{}{
			"player_id":  playerID,
			"season":     season,
			"week":       week,
			"projection": nil,
			"message":    "No projection data available for this week",
		})
		return
	}

	response.Success(w, map[string]interface{}{
		"player_id":  playerID,
		"season":     season,
		"week":       week,
		"projection": projection,
	})
}

// HandleTopProjections returns top projected fantasy players for a week
func (h *YahooFantasyHandler) HandleTopProjections(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	position := query.Get("position")
	limit := parseIntParam(query.Get("limit"), 25)

	seasonInfo := utils.GetCurrentSeason()
	season := parseIntParam(query.Get("season"), seasonInfo.Year)
	week := parseIntParam(query.Get("week"), seasonInfo.CurrentWeek)

	if limit > 100 {
		limit = 100
	}

	// Query top projected players
	pool := db.GetPool()
	dbQuery := `
		SELECT
			p.id, p.name, p.position, t.abbreviation as team,
			proj.projected_points,
			proj.projected_passing_yards, proj.projected_passing_tds,
			proj.projected_rushing_yards, proj.projected_rushing_tds,
			proj.projected_receptions, proj.projected_receiving_yards, proj.projected_receiving_tds
		FROM yahoo_player_projections proj
		JOIN players p ON proj.player_id = p.id
		LEFT JOIN teams t ON p.team_id = t.id
		WHERE proj.season = $1 AND proj.week = $2
	`

	args := []interface{}{season, week}
	argCount := 3

	if position != "" {
		dbQuery += ` AND p.position = $` + strconv.Itoa(argCount)
		args = append(args, position)
		argCount++
	}

	dbQuery += ` ORDER BY proj.projected_points DESC LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := pool.Query(r.Context(), dbQuery, args...)
	if err != nil {
		response.InternalError(w, "Failed to fetch projections")
		return
	}
	defer rows.Close()

	type ProjectedPlayer struct {
		ID              uuid.UUID `json:"id"`
		Name            string    `json:"name"`
		Position        string    `json:"position"`
		Team            string    `json:"team"`
		ProjectedPoints float64   `json:"projected_points"`
		PassingYards    *int      `json:"projected_passing_yards,omitempty"`
		PassingTDs      *int      `json:"projected_passing_tds,omitempty"`
		RushingYards    *int      `json:"projected_rushing_yards,omitempty"`
		RushingTDs      *int      `json:"projected_rushing_tds,omitempty"`
		Receptions      *int      `json:"projected_receptions,omitempty"`
		ReceivingYards  *int      `json:"projected_receiving_yards,omitempty"`
		ReceivingTDs    *int      `json:"projected_receiving_tds,omitempty"`
	}

	var players []ProjectedPlayer
	for rows.Next() {
		var p ProjectedPlayer
		err := rows.Scan(
			&p.ID, &p.Name, &p.Position, &p.Team,
			&p.ProjectedPoints,
			&p.PassingYards, &p.PassingTDs,
			&p.RushingYards, &p.RushingTDs,
			&p.Receptions, &p.ReceivingYards, &p.ReceivingTDs,
		)
		if err != nil {
			continue
		}
		players = append(players, p)
	}

	response.Success(w, map[string]interface{}{
		"season":   season,
		"week":     week,
		"position": position,
		"count":    len(players),
		"players":  players,
	})
}

// HandleOwnershipData returns player ownership percentages
func (h *YahooFantasyHandler) HandleOwnershipData(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	position := query.Get("position")
	minOwnership := parseFloatParam(query.Get("min_owned"), 0.0)
	limit := parseIntParam(query.Get("limit"), 50)

	seasonInfo := utils.GetCurrentSeason()
	season := parseIntParam(query.Get("season"), seasonInfo.Year)
	week := parseIntParam(query.Get("week"), seasonInfo.CurrentWeek)

	if limit > 100 {
		limit = 100
	}

	pool := db.GetPool()
	dbQuery := `
		SELECT
			p.id, p.name, p.position, t.abbreviation as team,
			r.percent_owned, r.overall_rank, r.position_rank
		FROM yahoo_player_rankings r
		JOIN players p ON r.player_id = p.id
		LEFT JOIN teams t ON p.team_id = t.id
		WHERE r.season = $1 AND r.week = $2
		AND r.percent_owned >= $3
	`

	args := []interface{}{season, week, minOwnership}
	argCount := 4

	if position != "" {
		dbQuery += ` AND p.position = $` + strconv.Itoa(argCount)
		args = append(args, position)
		argCount++
	}

	dbQuery += ` ORDER BY r.percent_owned DESC LIMIT $` + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := pool.Query(r.Context(), dbQuery, args...)
	if err != nil {
		response.InternalError(w, "Failed to fetch ownership data")
		return
	}
	defer rows.Close()

	type OwnershipData struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Position     string    `json:"position"`
		Team         string    `json:"team"`
		PercentOwned *float64  `json:"percent_owned"`
		OverallRank  *int      `json:"overall_rank,omitempty"`
		PositionRank *int      `json:"position_rank,omitempty"`
	}

	var players []OwnershipData
	for rows.Next() {
		var p OwnershipData
		err := rows.Scan(
			&p.ID, &p.Name, &p.Position, &p.Team,
			&p.PercentOwned, &p.OverallRank, &p.PositionRank,
		)
		if err != nil {
			continue
		}
		players = append(players, p)
	}

	response.Success(w, map[string]interface{}{
		"season":   season,
		"week":     week,
		"position": position,
		"count":    len(players),
		"players":  players,
	})
}

// parseFloatParam parses a float parameter with a default value
func parseFloatParam(param string, defaultValue float64) float64 {
	if param == "" {
		return defaultValue
	}
	val, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return defaultValue
	}
	return val
}

// parseIntParam parses an integer parameter with a default value
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(param)
	if err != nil {
		return defaultValue
	}
	return val
}
