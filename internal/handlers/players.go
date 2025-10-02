package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/francisco/gridironmind/internal/autofetch"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/francisco/gridironmind/pkg/validation"
	"github.com/google/uuid"
)

type PlayersHandler struct {
	queries          *db.PlayerQueries
	autoFetchEnabled bool
	orchestrator     *autofetch.Orchestrator
}

func NewPlayersHandler() *PlayersHandler {
	return &PlayersHandler{
		queries:          &db.PlayerQueries{},
		autoFetchEnabled: true,
		orchestrator:     autofetch.NewOrchestrator(""),
	}
}

// HandlePlayers handles GET /players (list), GET /players/:id (single), GET /players/:id/career, and GET /players/:id/history
func (h *PlayersHandler) HandlePlayers(w http.ResponseWriter, r *http.Request) {
	// Parse path to determine the request type
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/players")
	path = strings.Trim(path, "/")

	if path == "" {
		// List players
		h.listPlayers(w, r)
	} else if strings.HasSuffix(path, "/career") {
		// Get player career stats
		careerHandler := NewCareerHandler()
		careerHandler.HandlePlayerCareerStats(w, r)
	} else if strings.HasSuffix(path, "/history") {
		// Get player team history
		careerHandler := NewCareerHandler()
		careerHandler.HandlePlayerTeamHistory(w, r)
	} else if strings.HasSuffix(path, "/injuries") {
		// Get player injuries
		injuryHandler := NewInjuryHandler()
		injuryHandler.HandlePlayerInjuries(w, r)
	} else if strings.HasSuffix(path, "/advanced-stats") {
		// Get player advanced stats (Next Gen Stats)
		advancedStatsHandler := NewAdvancedStatsHandler()
		advancedStatsHandler.HandleAdvancedStats(w, r)
	} else if strings.Contains(path, "/vs-defense/") {
		// Get player vs defense stats
		defensiveHandler := NewDefensiveHandler()
		defensiveHandler.HandlePlayerVsDefense(w, r)
	} else {
		// Get single player by ID
		h.getPlayer(w, r, path)
	}
}

func (h *PlayersHandler) listPlayers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	limit := validation.ValidateLimit(validation.ParseIntParam(query.Get("limit"), 50))
	offset := validation.ValidateOffset(validation.ParseIntParam(query.Get("offset"), 0))
	position := strings.TrimSpace(query.Get("position"))
	status := strings.TrimSpace(query.Get("status"))
	teamIDStr := strings.TrimSpace(query.Get("team"))

	// Build filters
	filters := db.PlayerFilters{
		Limit:  limit,
		Offset: offset,
	}

	// Validate and set position filter
	if position != "" {
		position = strings.ToUpper(position)
		if err := validation.ValidatePosition(position); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		filters.Position = position
	}

	// Validate and set status filter
	if status != "" {
		status = strings.ToLower(status)
		if err := validation.ValidateStatus(status); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
		filters.Status = status
	}

	// Parse team ID filter
	if teamIDStr != "" {
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			response.BadRequest(w, "Invalid team ID format")
			return
		}
		filters.TeamID = teamID
	}

	// Query database
	players, total, err := h.queries.ListPlayers(r.Context(), filters)
	if err != nil {
		log.Printf("Error listing players: %v", err)
		response.InternalError(w, "Failed to retrieve players")
		return
	}

	// AUTO-FETCH: If no players found and filtering by team, try to fetch rosters
	if total == 0 && h.autoFetchEnabled && filters.TeamID != uuid.Nil {
		log.Printf("[AUTO-FETCH] No players found for team %s, attempting auto-fetch", filters.TeamID)

		// Ensure teams exist first, then fetch rosters
		teamQueries := &db.TeamQueries{}
		team, err := teamQueries.GetTeamByID(r.Context(), filters.TeamID)
		if err == nil && team != nil {
			// Team exists, fetch rosters
			if err := h.orchestrator.FetchGamesIfMissing(r.Context(), 2025, 1); err != nil {
				log.Printf("[AUTO-FETCH] Failed initial setup: %v", err)
			}

			// Retry query after fetch
			players, total, err = h.queries.ListPlayers(r.Context(), filters)
			if err == nil && total > 0 {
				log.Printf("[AUTO-FETCH] Successfully fetched and returned %d players", total)
				w.Header().Set("X-Auto-Fetched", "true")
			}
		}
	}

	// Return response with pagination
	response.SuccessWithPagination(w, players, total, limit, offset)
}

func (h *PlayersHandler) getPlayer(w http.ResponseWriter, r *http.Request, idStr string) {
	// Parse player ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid player ID format")
		return
	}

	// Query database
	player, err := h.queries.GetPlayerByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting player: %v", err)
		response.InternalError(w, "Failed to retrieve player")
		return
	}

	if player == nil {
		response.NotFound(w, "Player")
		return
	}

	// Return response
	response.Success(w, player)
}