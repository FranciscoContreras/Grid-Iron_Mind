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

// HandleGames handles both GET /games (list) and GET /games/:id (single) and GET /games/:id/stats
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
	} else if strings.HasSuffix(path, "/stats") {
		// Get team stats for game
		h.HandleGameStats(w, r)
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

// HandleGameStats returns team statistics for a specific game
func (h *GamesHandler) HandleGameStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET method is allowed")
		return
	}

	// Extract game ID from path: /api/v1/games/:id/stats
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/games/")
	path = strings.TrimSuffix(path, "/stats")

	id, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "Game ID must be a valid UUID")
		return
	}

	log.Printf("Fetching team stats for game %s", id)

	// Query team stats for this game
	query := `
		SELECT
			gts.id,
			gts.game_id,
			gts.team_id,
			t.name as team_name,
			t.abbreviation as team_abbr,
			gts.first_downs,
			gts.total_yards,
			gts.passing_yards,
			gts.rushing_yards,
			gts.offensive_plays,
			gts.yards_per_play,
			gts.third_down_attempts,
			gts.third_down_conversions,
			gts.third_down_pct,
			gts.fourth_down_attempts,
			gts.fourth_down_conversions,
			gts.fourth_down_pct,
			gts.red_zone_attempts,
			gts.red_zone_scores,
			gts.turnovers,
			gts.fumbles_lost,
			gts.interceptions_thrown,
			gts.penalties,
			gts.penalty_yards,
			gts.possession_time,
			gts.possession_seconds,
			gts.completions,
			gts.pass_attempts,
			gts.sacks_allowed,
			gts.sack_yards,
			gts.rushing_attempts,
			gts.rushing_avg
		FROM game_team_stats gts
		JOIN teams t ON gts.team_id = t.id
		WHERE gts.game_id = $1
		ORDER BY t.abbreviation
	`

	rows, err := db.GetPool().Query(r.Context(), query, id)
	if err != nil {
		log.Printf("Error querying team stats for game %s: %v", id, err)
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to fetch team stats")
		return
	}
	defer rows.Close()

	type TeamStats struct {
		ID                    uuid.UUID `json:"id"`
		GameID                uuid.UUID `json:"game_id"`
		TeamID                uuid.UUID `json:"team_id"`
		TeamName              string    `json:"team_name"`
		TeamAbbr              string    `json:"team_abbr"`
		FirstDowns            int       `json:"first_downs"`
		TotalYards            int       `json:"total_yards"`
		PassingYards          int       `json:"passing_yards"`
		RushingYards          int       `json:"rushing_yards"`
		OffensivePlays        int       `json:"offensive_plays"`
		YardsPerPlay          float64   `json:"yards_per_play"`
		ThirdDownAttempts     int       `json:"third_down_attempts"`
		ThirdDownConversions  int       `json:"third_down_conversions"`
		ThirdDownPct          float64   `json:"third_down_pct"`
		FourthDownAttempts    int       `json:"fourth_down_attempts"`
		FourthDownConversions int       `json:"fourth_down_conversions"`
		FourthDownPct         float64   `json:"fourth_down_pct"`
		RedZoneAttempts       int       `json:"red_zone_attempts"`
		RedZoneScores         int       `json:"red_zone_scores"`
		Turnovers             int       `json:"turnovers"`
		FumblesLost           int       `json:"fumbles_lost"`
		InterceptionsThrown   int       `json:"interceptions_thrown"`
		Penalties             int       `json:"penalties"`
		PenaltyYards          int       `json:"penalty_yards"`
		PossessionTime        string    `json:"possession_time"`
		PossessionSeconds     int       `json:"possession_seconds"`
		Completions           int       `json:"completions"`
		PassAttempts          int       `json:"pass_attempts"`
		SacksAllowed          int       `json:"sacks_allowed"`
		SackYards             int       `json:"sack_yards"`
		RushingAttempts       int       `json:"rushing_attempts"`
		RushingAvg            float64   `json:"rushing_avg"`
	}

	var stats []TeamStats
	for rows.Next() {
		var s TeamStats
		err := rows.Scan(
			&s.ID, &s.GameID, &s.TeamID, &s.TeamName, &s.TeamAbbr,
			&s.FirstDowns, &s.TotalYards, &s.PassingYards, &s.RushingYards,
			&s.OffensivePlays, &s.YardsPerPlay,
			&s.ThirdDownAttempts, &s.ThirdDownConversions, &s.ThirdDownPct,
			&s.FourthDownAttempts, &s.FourthDownConversions, &s.FourthDownPct,
			&s.RedZoneAttempts, &s.RedZoneScores,
			&s.Turnovers, &s.FumblesLost, &s.InterceptionsThrown,
			&s.Penalties, &s.PenaltyYards,
			&s.PossessionTime, &s.PossessionSeconds,
			&s.Completions, &s.PassAttempts,
			&s.SacksAllowed, &s.SackYards,
			&s.RushingAttempts, &s.RushingAvg,
		)
		if err != nil {
			log.Printf("Error scanning team stats: %v", err)
			continue
		}
		stats = append(stats, s)
	}

	if len(stats) == 0 {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "No team stats found for this game")
		return
	}

	response.Success(w, stats)
}