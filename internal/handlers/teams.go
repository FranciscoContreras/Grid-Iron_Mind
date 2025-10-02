package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

type TeamsHandler struct {
	queries *db.TeamQueries
}

func NewTeamsHandler() *TeamsHandler {
	return &TeamsHandler{
		queries: &db.TeamQueries{},
	}
}

// HandleTeams handles GET /teams (list) and GET /teams/:id (single)
func (h *TeamsHandler) HandleTeams(w http.ResponseWriter, r *http.Request) {
	// Parse path to determine if this is a list or single team request
	// Support both v1 and v2 API paths
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/v1/teams")
	path = strings.TrimPrefix(path, "/api/v2/teams")
	path = strings.Trim(path, "/")

	if path == "" {
		// List all teams
		h.listTeams(w, r)
	} else {
		// Check if requesting team players, injuries, or defensive stats
		parts := strings.Split(path, "/")
		if len(parts) >= 2 && parts[1] == "players" {
			h.getTeamPlayers(w, r, parts[0])
		} else if len(parts) >= 2 && parts[1] == "injuries" {
			injuryHandler := NewInjuryHandler()
			injuryHandler.HandleTeamInjuries(w, r)
		} else if len(parts) >= 3 && parts[1] == "defense" && parts[2] == "stats" {
			// Route to defensive stats handler
			defensiveHandler := NewDefensiveHandler()
			defensiveHandler.HandleTeamDefenseStats(w, r)
		} else {
			// Get single team by ID
			h.getTeam(w, r, path)
		}
	}
}

func (h *TeamsHandler) listTeams(w http.ResponseWriter, r *http.Request) {
	cacheKey := cache.TeamsCacheKey()

	// Try cache first
	if cached, err := cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Query database
	teams, err := h.queries.ListTeams(r.Context())
	if err != nil {
		log.Printf("Error listing teams: %v", err)
		response.InternalError(w, "Failed to retrieve teams")
		return
	}

	// Build response
	respData := struct {
		Data interface{} `json:"data"`
		Meta struct {
			Timestamp string `json:"timestamp"`
		} `json:"meta"`
	}{
		Data: teams,
	}
	respData.Meta.Timestamp = getCurrentTimestamp()

	// Marshal and cache
	respJSON, _ := json.Marshal(respData)
	cache.Set(r.Context(), cacheKey, string(respJSON), cache.TTLTeams)

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(respJSON)
}

func (h *TeamsHandler) getTeam(w http.ResponseWriter, r *http.Request, idStr string) {
	// Parse team ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid team ID format")
		return
	}

	// Query database
	team, err := h.queries.GetTeamByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting team: %v", err)
		response.InternalError(w, "Failed to retrieve team")
		return
	}

	if team == nil {
		response.NotFound(w, "Team")
		return
	}

	// Return response
	response.Success(w, team)
}

func (h *TeamsHandler) getTeamPlayers(w http.ResponseWriter, r *http.Request, idStr string) {
	// Parse team ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid team ID format")
		return
	}

	// First verify team exists
	team, err := h.queries.GetTeamByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting team: %v", err)
		response.InternalError(w, "Failed to retrieve team")
		return
	}

	if team == nil {
		response.NotFound(w, "Team")
		return
	}

	// Query team players
	players, err := h.queries.GetTeamPlayers(r.Context(), id)
	if err != nil {
		log.Printf("Error getting team players: %v", err)
		response.InternalError(w, "Failed to retrieve team players")
		return
	}

	// Return response
	response.Success(w, players)
}