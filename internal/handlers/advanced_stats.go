package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

type AdvancedStatsHandler struct{}

func NewAdvancedStatsHandler() *AdvancedStatsHandler {
	return &AdvancedStatsHandler{}
}

// HandleAdvancedStats routes to player-specific advanced stats
//
// Routes:
//   - GET /api/v1/players/:id/advanced-stats - Get player's advanced stats
func (h *AdvancedStatsHandler) HandleAdvancedStats(w http.ResponseWriter, r *http.Request) {
	// Extract player ID from path
	path := r.URL.Path; path = strings.TrimPrefix(path, "/api/v1/players/"); path = strings.TrimPrefix(path, "/api/v2/players/")
	path = strings.TrimSuffix(path, "/advanced-stats")

	playerID, err := uuid.Parse(path)
	if err != nil {
		response.BadRequest(w, "Invalid player ID")
		return
	}

	h.getPlayerAdvancedStats(w, r, playerID)
}

// getPlayerAdvancedStats retrieves advanced stats for a player
//
// Query parameters:
//   - season (optional): Filter by season
//   - week (optional): Filter by week (NULL = season totals)
//   - stat_type (optional): Filter by stat type (passing, rushing, receiving)
//
// Example: GET /api/v1/players/{id}/advanced-stats?season=2024&stat_type=passing
func (h *AdvancedStatsHandler) getPlayerAdvancedStats(w http.ResponseWriter, r *http.Request, playerID uuid.UUID) {
	query := r.URL.Query()

	// Build query
	sql := `
		SELECT
			a.id,
			a.player_id,
			p.name as player_name,
			a.season,
			a.week,
			a.stat_type,
			-- Passing
			a.avg_time_to_throw,
			a.avg_completed_air_yards,
			a.avg_intended_air_yards,
			a.avg_air_yards_differential,
			a.max_completed_air_distance,
			a.avg_air_yards_to_sticks,
			a.attempts,
			a.pass_yards,
			a.pass_touchdowns,
			a.interceptions,
			a.completions,
			a.completion_percentage,
			a.expected_completion_percentage,
			a.completion_percentage_above_expectation,
			a.passer_rating,
			-- Rushing
			a.efficiency,
			a.percent_attempts_gte_eight_defenders,
			a.avg_time_to_los,
			a.rush_attempts,
			a.rush_yards,
			a.expected_rush_yards,
			a.rush_yards_over_expected,
			a.avg_rush_yards,
			a.rush_touchdowns,
			-- Receiving
			a.avg_cushion,
			a.avg_separation,
			a.avg_intended_air_yards_receiving,
			a.percent_share_of_intended_air_yards,
			a.receptions,
			a.targets,
			a.catch_percentage,
			a.yards,
			a.rec_touchdowns,
			a.avg_yac,
			a.avg_expected_yac,
			a.avg_yac_above_expectation,
			a.created_at,
			a.updated_at
		FROM advanced_stats a
		JOIN players p ON a.player_id = p.id
		WHERE a.player_id = $1
	`

	args := []interface{}{playerID}
	argNum := 2

	// Filter by season
	if seasonStr := query.Get("season"); seasonStr != "" {
		season, err := strconv.Atoi(seasonStr)
		if err == nil {
			sql += " AND a.season = $" + strconv.Itoa(argNum)
			args = append(args, season)
			argNum++
		}
	}

	// Filter by week
	if weekStr := query.Get("week"); weekStr != "" {
		if weekStr == "season" {
			sql += " AND a.week IS NULL"
		} else {
			week, err := strconv.Atoi(weekStr)
			if err == nil {
				sql += " AND a.week = $" + strconv.Itoa(argNum)
				args = append(args, week)
				argNum++
			}
		}
	}

	// Filter by stat type
	if statType := query.Get("stat_type"); statType != "" {
		sql += " AND a.stat_type = $" + strconv.Itoa(argNum)
		args = append(args, statType)
		argNum++
	}

	sql += " ORDER BY a.season DESC, a.week DESC NULLS FIRST, a.stat_type"

	rows, err := db.GetPool().Query(r.Context(), sql, args...)
	if err != nil {
		log.Printf("Error querying advanced stats: %v", err)
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to fetch advanced stats")
		return
	}
	defer rows.Close()

	type AdvancedStat struct {
		ID         uuid.UUID `json:"id"`
		PlayerID   uuid.UUID `json:"player_id"`
		PlayerName string    `json:"player_name"`
		Season     int       `json:"season"`
		Week       *int      `json:"week,omitempty"`
		StatType   string    `json:"stat_type"`

		// Passing
		AvgTimeToThrow                        *float64 `json:"avg_time_to_throw,omitempty"`
		AvgCompletedAirYards                  *float64 `json:"avg_completed_air_yards,omitempty"`
		AvgIntendedAirYards                   *float64 `json:"avg_intended_air_yards,omitempty"`
		AvgAirYardsDifferential               *float64 `json:"avg_air_yards_differential,omitempty"`
		MaxCompletedAirDistance               *int     `json:"max_completed_air_distance,omitempty"`
		AvgAirYardsToSticks                   *float64 `json:"avg_air_yards_to_sticks,omitempty"`
		Attempts                              *int     `json:"attempts,omitempty"`
		PassYards                             *int     `json:"pass_yards,omitempty"`
		PassTouchdowns                        *int     `json:"pass_touchdowns,omitempty"`
		Interceptions                         *int     `json:"interceptions,omitempty"`
		Completions                           *int     `json:"completions,omitempty"`
		CompletionPercentage                  *float64 `json:"completion_percentage,omitempty"`
		ExpectedCompletionPercentage          *float64 `json:"expected_completion_percentage,omitempty"`
		CompletionPercentageAboveExpectation  *float64 `json:"completion_percentage_above_expectation,omitempty"`
		PasserRating                          *float64 `json:"passer_rating,omitempty"`

		// Rushing
		Efficiency                         *float64 `json:"efficiency,omitempty"`
		PercentAttemptsGteEightDefenders   *float64 `json:"percent_attempts_gte_eight_defenders,omitempty"`
		AvgTimeToLOS                       *float64 `json:"avg_time_to_los,omitempty"`
		RushAttempts                       *int     `json:"rush_attempts,omitempty"`
		RushYards                          *int     `json:"rush_yards,omitempty"`
		ExpectedRushYards                  *int     `json:"expected_rush_yards,omitempty"`
		RushYardsOverExpected              *int     `json:"rush_yards_over_expected,omitempty"`
		AvgRushYards                       *float64 `json:"avg_rush_yards,omitempty"`
		RushTouchdowns                     *int     `json:"rush_touchdowns,omitempty"`

		// Receiving
		AvgCushion                         *float64 `json:"avg_cushion,omitempty"`
		AvgSeparation                      *float64 `json:"avg_separation,omitempty"`
		AvgIntendedAirYardsReceiving       *float64 `json:"avg_intended_air_yards_receiving,omitempty"`
		PercentShareOfIntendedAirYards     *float64 `json:"percent_share_of_intended_air_yards,omitempty"`
		Receptions                         *int     `json:"receptions,omitempty"`
		Targets                            *int     `json:"targets,omitempty"`
		CatchPercentage                    *float64 `json:"catch_percentage,omitempty"`
		Yards                              *int     `json:"yards,omitempty"`
		RecTouchdowns                      *int     `json:"rec_touchdowns,omitempty"`
		AvgYAC                             *float64 `json:"avg_yac,omitempty"`
		AvgExpectedYAC                     *float64 `json:"avg_expected_yac,omitempty"`
		AvgYACAboveExpectation             *float64 `json:"avg_yac_above_expectation,omitempty"`

		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	var stats []AdvancedStat

	for rows.Next() {
		var s AdvancedStat
		err := rows.Scan(
			&s.ID, &s.PlayerID, &s.PlayerName, &s.Season, &s.Week, &s.StatType,
			// Passing
			&s.AvgTimeToThrow, &s.AvgCompletedAirYards, &s.AvgIntendedAirYards,
			&s.AvgAirYardsDifferential, &s.MaxCompletedAirDistance, &s.AvgAirYardsToSticks,
			&s.Attempts, &s.PassYards, &s.PassTouchdowns, &s.Interceptions,
			&s.Completions, &s.CompletionPercentage,
			&s.ExpectedCompletionPercentage, &s.CompletionPercentageAboveExpectation,
			&s.PasserRating,
			// Rushing
			&s.Efficiency, &s.PercentAttemptsGteEightDefenders, &s.AvgTimeToLOS,
			&s.RushAttempts, &s.RushYards, &s.ExpectedRushYards,
			&s.RushYardsOverExpected, &s.AvgRushYards, &s.RushTouchdowns,
			// Receiving
			&s.AvgCushion, &s.AvgSeparation, &s.AvgIntendedAirYardsReceiving,
			&s.PercentShareOfIntendedAirYards,
			&s.Receptions, &s.Targets, &s.CatchPercentage,
			&s.Yards, &s.RecTouchdowns,
			&s.AvgYAC, &s.AvgExpectedYAC, &s.AvgYACAboveExpectation,
			&s.CreatedAt, &s.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error scanning advanced stat: %v", err)
			continue
		}

		stats = append(stats, s)
	}

	if len(stats) == 0 {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "No advanced stats found for player")
		return
	}

	response.Success(w, stats)
}
