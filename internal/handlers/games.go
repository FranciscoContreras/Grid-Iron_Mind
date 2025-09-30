package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/models"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/francisco/gridironmind/pkg/validation"
	"github.com/google/uuid"
)

type GamesHandler struct {
	queries *db.GameQueries
}

func NewGamesHandler() *GamesHandler {
	return &GamesHandler{
		queries: &db.GameQueries{},
	}
}

// HandleGames handles both GET /games (list) and GET /games/:id (single)
func (h *GamesHandler) HandleGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/games")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		// List games
		h.listGames(w, r)
	} else {
		// Get single game by ID
		h.getGame(w, r, path)
	}
}

func (h *GamesHandler) listGames(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := validation.ValidateLimit(validation.ParseIntParam(limitStr, 50))
	offset := validation.ValidateOffset(validation.ParseIntParam(offsetStr, 0))

	// Parse filters
	var filters models.GameFilters
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
		if err != nil || week < 1 || week > 18 {
			response.Error(w, http.StatusBadRequest, "INVALID_WEEK", "Week must be between 1 and 18")
			return
		}
		filters.Week = week
	}

	if teamIDStr := r.URL.Query().Get("team"); teamIDStr != "" {
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_TEAM_ID", "Team ID must be a valid UUID")
			return
		}
		filters.TeamID = teamID
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = status
	}

	// Query database
	games, total, err := h.queries.ListGames(r.Context(), filters)
	if err != nil {
		log.Printf("Error listing games: %v", err)
		response.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve games")
		return
	}

	response.SuccessWithPagination(w, games, total, filters.Limit, filters.Offset)
}

func (h *GamesHandler) getGame(w http.ResponseWriter, r *http.Request, idStr string) {
	log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Game ID must be a valid UUID")
		return
	}

	game, err := h.queries.GetGameByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting game %s: %v", id, err)
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Game not found")
		return
	}

	response.Success(w, game)
}