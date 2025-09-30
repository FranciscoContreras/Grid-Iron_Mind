package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

// InjuryHandler handles injury-related requests
type InjuryHandler struct{}

// NewInjuryHandler creates a new injury handler
func NewInjuryHandler() *InjuryHandler {
	return &InjuryHandler{}
}

// HandlePlayerInjuries handles GET /api/v1/players/{id}/injuries
func (h *InjuryHandler) HandlePlayerInjuries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract player ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/players/")
	path = strings.TrimSuffix(path, "/injuries")
	playerID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid player ID")
		return
	}

	ctx := r.Context()
	injuryQueries := db.InjuryQueries{}

	injuries, err := injuryQueries.GetPlayerInjuries(ctx, playerID)
	if err != nil {
		log.Printf("Failed to get player injuries: %v", err)
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve player injuries")
		return
	}

	response.Success(w, map[string]interface{}{
		"player_id": playerID,
		"injuries":  injuries,
		"count":     len(injuries),
	})
}

// HandleTeamInjuries handles GET /api/v1/teams/{id}/injuries
func (h *InjuryHandler) HandleTeamInjuries(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract team ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/teams/")
	path = strings.TrimSuffix(path, "/injuries")
	teamID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid team ID")
		return
	}

	ctx := r.Context()
	injuryQueries := db.InjuryQueries{}

	injuries, err := injuryQueries.GetTeamInjuries(ctx, teamID)
	if err != nil {
		log.Printf("Failed to get team injuries: %v", err)
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve team injuries")
		return
	}

	// Group by status for better display
	grouped := make(map[string][]interface{})
	for _, injury := range injuries {
		status := injury.Status
		if grouped[status] == nil {
			grouped[status] = []interface{}{}
		}
		grouped[status] = append(grouped[status], injury)
	}

	response.Success(w, map[string]interface{}{
		"team_id":  teamID,
		"injuries": injuries,
		"grouped":  grouped,
		"count":    len(injuries),
	})
}
