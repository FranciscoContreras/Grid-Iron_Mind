package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

type DefensiveHandler struct {
	queries *db.DefensiveQueries
}

func NewDefensiveHandler() *DefensiveHandler {
	return &DefensiveHandler{
		queries: &db.DefensiveQueries{},
	}
}

// HandleTeamDefenseStats handles GET /api/v1/teams/:teamId/defense/stats
func (h *DefensiveHandler) HandleTeamDefenseStats(w http.ResponseWriter, r *http.Request) {
	// Extract team ID from path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		response.BadRequest(w, "Invalid path")
		return
	}

	teamID, err := uuid.Parse(pathParts[3]) // /api/v1/teams/:teamId/defense/stats
	if err != nil {
		response.BadRequest(w, "Invalid team ID")
		return
	}

	// Get query parameters
	query := r.URL.Query()
	seasonStr := query.Get("season")
	if seasonStr == "" {
		response.BadRequest(w, "season parameter is required")
		return
	}

	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		response.BadRequest(w, "Invalid season parameter")
		return
	}

	var week *int
	if weekStr := query.Get("week"); weekStr != "" {
		weekVal, err := strconv.Atoi(weekStr)
		if err != nil {
			response.BadRequest(w, "Invalid week parameter")
			return
		}
		week = &weekVal
	}

	// Get defensive stats
	stats, err := h.queries.GetTeamDefensiveStats(r.Context(), teamID, season, week)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			response.NotFound(w, "Defensive stats")
			return
		}
		response.InternalError(w, "Failed to fetch defensive stats")
		return
	}

	response.Success(w, stats)
}

// HandleDefensiveRankings handles GET /api/v1/defense/rankings
func (h *DefensiveHandler) HandleDefensiveRankings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	seasonStr := query.Get("season")
	if seasonStr == "" {
		response.BadRequest(w, "season parameter is required")
		return
	}

	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		response.BadRequest(w, "Invalid season parameter")
		return
	}

	category := query.Get("category")
	if category == "" {
		category = "overall"
	}

	// Validate category
	validCategories := map[string]bool{
		"overall":        true,
		"pass":           true,
		"rush":           true,
		"points_allowed": true,
	}

	if !validCategories[category] {
		response.BadRequest(w, "Invalid category. Must be one of: overall, pass, rush, points_allowed")
		return
	}

	// Get rankings
	rankings, err := h.queries.GetDefensiveRankings(r.Context(), season, category)
	if err != nil {
		response.InternalError(w, "Failed to fetch defensive rankings")
		return
	}

	if len(rankings) == 0 {
		response.Success(w, []interface{}{}) // Return empty array
		return
	}

	response.Success(w, rankings)
}

// HandlePlayerVsDefense handles GET /api/v1/players/:playerId/vs-defense/:teamId
func (h *DefensiveHandler) HandlePlayerVsDefense(w http.ResponseWriter, r *http.Request) {
	// Extract player ID and defense team ID from path
	// Path: /api/v1/players/:playerId/vs-defense/:teamId
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 6 {
		response.BadRequest(w, "Invalid path")
		return
	}

	playerID, err := uuid.Parse(pathParts[3])
	if err != nil {
		response.BadRequest(w, "Invalid player ID")
		return
	}

	defenseTeamID, err := uuid.Parse(pathParts[5])
	if err != nil {
		response.BadRequest(w, "Invalid defense team ID")
		return
	}

	// Get query parameters
	query := r.URL.Query()
	var season *int
	if seasonStr := query.Get("season"); seasonStr != "" {
		seasonVal, err := strconv.Atoi(seasonStr)
		if err != nil {
			response.BadRequest(w, "Invalid season parameter")
			return
		}
		season = &seasonVal
	}

	limit := 5 // Default
	if limitStr := query.Get("limit"); limitStr != "" {
		limitVal, err := strconv.Atoi(limitStr)
		if err != nil {
			response.BadRequest(w, "Invalid limit parameter")
			return
		}
		if limitVal > 0 && limitVal <= 50 {
			limit = limitVal
		}
	}

	// Get player vs defense stats
	result, err := h.queries.GetPlayerVsDefense(r.Context(), playerID, defenseTeamID, season, limit)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") || strings.Contains(err.Error(), "not found") {
			response.NotFound(w, "Player or team")
			return
		}
		response.InternalError(w, "Failed to fetch player vs defense stats")
		return
	}

	response.Success(w, result)
}
