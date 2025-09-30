package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

type CareerHandler struct {
	queries *db.CareerQueries
}

func NewCareerHandler() *CareerHandler {
	return &CareerHandler{
		queries: &db.CareerQueries{},
	}
}

// HandlePlayerCareerStats handles GET /players/:id/career
func (h *CareerHandler) HandlePlayerCareerStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Extract player ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/players/")
	path = strings.TrimSuffix(path, "/career")
	playerID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PLAYER_ID", "Player ID must be a valid UUID")
		return
	}

	// Check cache
	cacheKey := cache.CacheKey("player_career", playerID.String())
	if cached, err := cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Get career stats
	stats, err := h.queries.GetPlayerCareerStats(r.Context(), playerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve career stats")
		return
	}

	// Get team history
	history, err := h.queries.GetPlayerTeamHistory(r.Context(), playerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve team history")
		return
	}

	respData := map[string]interface{}{
		"player_id":     playerID,
		"career_stats":  stats,
		"team_history":  history,
		"total_seasons": len(stats),
	}

	// Cache response for 1 hour
	respJSON := response.ToJSON(respData)
	cache.Set(r.Context(), cacheKey, respJSON, 1*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(respJSON))
}

// HandlePlayerTeamHistory handles GET /players/:id/history
func (h *CareerHandler) HandlePlayerTeamHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Extract player ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/players/")
	path = strings.TrimSuffix(path, "/history")
	playerID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PLAYER_ID", "Player ID must be a valid UUID")
		return
	}

	// Check cache
	cacheKey := cache.CacheKey("player_history", playerID.String())
	if cached, err := cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Get team history
	history, err := h.queries.GetPlayerTeamHistory(r.Context(), playerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve team history")
		return
	}

	respData := map[string]interface{}{
		"player_id":     playerID,
		"team_history":  history,
		"total_teams":   len(history),
	}

	// Cache response for 1 hour
	respJSON := response.ToJSON(respData)
	cache.Set(r.Context(), cacheKey, respJSON, 1*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(respJSON))
}