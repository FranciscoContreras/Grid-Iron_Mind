package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

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
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Parse path to determine if this is a list or single team request
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/teams")
	path = strings.Trim(path, "/")

	if path == "" {
		// List all teams
		h.listTeams(w, r)
	} else {
		// Check if requesting team players
		parts := strings.Split(path, "/")
		if len(parts) == 2 && parts[1] == "players" {
			h.getTeamPlayers(w, r, parts[0])
		} else {
			// Get single team by ID
			h.getTeam(w, r, path)
		}
	}
}

func (h *TeamsHandler) listTeams(w http.ResponseWriter, r *http.Request) {
	// Query database
	teams, err := h.queries.ListTeams(r.Context())
	if err != nil {
		log.Printf("Error listing teams: %v", err)
		response.InternalError(w, "Failed to retrieve teams")
		return
	}

	// Return response
	response.Success(w, teams)
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