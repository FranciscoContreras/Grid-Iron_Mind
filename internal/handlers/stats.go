package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/francisco/gridironmind/internal/autofetch"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/models"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/francisco/gridironmind/pkg/validation"
	"github.com/google/uuid"
)

type StatsHandler struct {
	queries          *db.StatsQueries
	autoFetchEnabled bool
	orchestrator     *autofetch.Orchestrator
}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		queries:          &db.StatsQueries{},
		autoFetchEnabled: true,
		orchestrator:     autofetch.NewOrchestrator(""),
	}
}

// HandleGameStats handles GET /stats/game/:gameID - returns player stats for a game
func (h *StatsHandler) HandleGameStats(w http.ResponseWriter, r *http.Request) {
	// Extract game ID from path: /api/v1/stats/game/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/stats/game/")
	path = strings.TrimSuffix(path, "/")

	gameID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_GAME_ID", "Game ID must be a valid UUID")
		return
	}

	log.Printf("%s %s from %s - Getting player stats for game %s", r.Method, r.URL.Path, r.RemoteAddr, gameID)

	stats, err := h.queries.GetGameStats(r.Context(), gameID)
	if err != nil {
		log.Printf("Error getting game stats for %s: %v", gameID, err)
		response.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve game stats")
		return
	}

	// AUTO-FETCH: If no stats found for game, try to fetch them
	if len(stats) == 0 && h.autoFetchEnabled {
		log.Printf("[AUTO-FETCH] No stats found for game %s, attempting auto-fetch", gameID)

		if err := h.orchestrator.FetchStatsIfMissing(r.Context(), gameID); err != nil {
			log.Printf("[AUTO-FETCH] Failed to fetch stats for game %s: %v", gameID, err)
			// Continue with empty result
		} else {
			// Retry query after fetch
			stats, err = h.queries.GetGameStats(r.Context(), gameID)
			if err == nil && len(stats) > 0 {
				log.Printf("[AUTO-FETCH] Successfully fetched and returned %d stats", len(stats))
				w.Header().Set("X-Auto-Fetched", "true")
			}
		}
	}

	response.Success(w, stats)
}

// HandlePlayerStats handles GET /players/:playerID/stats
func (h *StatsHandler) HandlePlayerStats(w http.ResponseWriter, r *http.Request) {
	// Extract player ID from path: /api/v1/players/{id}/stats
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/players/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		response.Error(w, http.StatusBadRequest, "INVALID_PATH", "Invalid player stats path")
		return
	}

	playerID, err := uuid.Parse(parts[0])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PLAYER_ID", "Player ID must be a valid UUID")
		return
	}

	log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := validation.ValidateLimit(validation.ParseIntParam(limitStr, 50))
	offset := validation.ValidateOffset(validation.ParseIntParam(offsetStr, 0))

	var filters models.StatsFilters
	filters.Limit = limit
	filters.Offset = offset

	if seasonStr := r.URL.Query().Get("season"); seasonStr != "" {
		season, err := strconv.Atoi(seasonStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_SEASON", "Season must be a valid year")
			return
		}
		filters.Season = season
	}

	if weekStr := r.URL.Query().Get("week"); weekStr != "" {
		week, err := strconv.Atoi(weekStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_WEEK", "Week must be a valid number")
			return
		}
		filters.Week = week
	}

	stats, err := h.queries.GetPlayerStats(r.Context(), playerID, filters)
	if err != nil {
		log.Printf("Error getting player stats for %s: %v", playerID, err)
		response.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve player stats")
		return
	}

	response.Success(w, stats)
}

// HandleStatsLeaders handles GET /stats/leaders
func (h *StatsHandler) HandleStatsLeaders(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Parse query parameters
	category := r.URL.Query().Get("category")
	if category == "" {
		category = "passing_yards" // default
	}

	// Validate category
	validCategories := map[string]bool{
		"passing_yards":       true,
		"passing_touchdowns":  true,
		"rushing_yards":       true,
		"rushing_touchdowns":  true,
		"receiving_yards":     true,
		"receiving_touchdowns": true,
	}

	if !validCategories[category] {
		response.Error(w, http.StatusBadRequest, "INVALID_CATEGORY", "Invalid stats category")
		return
	}

	seasonStr := r.URL.Query().Get("season")
	if seasonStr == "" {
		seasonStr = "2025" // default to current season
	}

	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_SEASON", "Season must be a valid year")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			response.Error(w, http.StatusBadRequest, "INVALID_LIMIT", "Limit must be between 1 and 100")
			return
		}
	}

	leaders, err := h.queries.GetStatsLeaders(r.Context(), category, season, limit)
	if err != nil {
		log.Printf("Error getting stats leaders: %v", err)
		response.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve stats leaders")
		return
	}

	response.Success(w, leaders)
}