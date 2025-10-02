package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

type StandingsHandler struct{}

func NewStandingsHandler() *StandingsHandler {
	return &StandingsHandler{}
}

// HandleStandings returns standings for a season/week
//
// Query parameters:
//   - season (required): NFL season year
//   - week (optional): Specific week (default: latest week with standings)
//   - division (optional): Filter by division (e.g., "AFC East")
//   - conference (optional): Filter by conference ("AFC" or "NFC")
//
// Example: GET /api/v1/standings?season=2025&week=4
func (h *StandingsHandler) HandleStandings(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	seasonStr := r.URL.Query().Get("season")
	if seasonStr == "" {
		response.BadRequest(w, "Season parameter is required")
		return
	}

	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		response.BadRequest(w, "Invalid season")
		return
	}

	weekStr := r.URL.Query().Get("week")
	division := r.URL.Query().Get("division")
	conference := r.URL.Query().Get("conference")

	var week *int
	if weekStr != "" {
		weekNum, err := strconv.Atoi(weekStr)
		if err != nil {
			response.BadRequest(w, "Invalid week")
			return
		}
		week = &weekNum
	}

	// If no week specified, get latest week with standings
	if week == nil {
		var latestWeek int
		err := db.GetPool().QueryRow(r.Context(), `
			SELECT COALESCE(MAX(week), 0)
			FROM team_standings
			WHERE season = $1
		`, season).Scan(&latestWeek)

		if err != nil || latestWeek == 0 {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "No standings found for season")
			return
		}
		week = &latestWeek
	}

	log.Printf("Fetching standings for season %d week %d", season, *week)

	// Build query
	query := `
		SELECT
			ts.id,
			ts.team_id,
			t.name as team_name,
			t.abbreviation as team_abbr,
			t.conference,
			t.division,
			ts.wins,
			ts.losses,
			ts.ties,
			ts.win_pct,
			ts.points_for,
			ts.points_against,
			ts.point_differential,
			ts.home_wins,
			ts.home_losses,
			ts.away_wins,
			ts.away_losses,
			ts.division_wins,
			ts.division_losses,
			ts.conference_wins,
			ts.conference_losses,
			ts.current_streak,
			ts.division_rank,
			ts.conference_rank,
			ts.playoff_seed
		FROM team_standings ts
		JOIN teams t ON ts.team_id = t.id
		WHERE ts.season = $1 AND ts.week = $2
	`

	args := []interface{}{season, *week}

	if division != "" {
		query += " AND t.division = $3"
		args = append(args, division)
	} else if conference != "" {
		query += " AND t.conference = $3"
		args = append(args, conference)
	}

	query += " ORDER BY t.conference, t.division, ts.division_rank"

	rows, err := db.GetPool().Query(r.Context(), query, args...)
	if err != nil {
		log.Printf("Error querying standings: %v", err)
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to fetch standings")
		return
	}
	defer rows.Close()

	type Standing struct {
		ID                uuid.UUID  `json:"id"`
		TeamID            uuid.UUID  `json:"team_id"`
		TeamName          string     `json:"team_name"`
		TeamAbbr          string     `json:"team_abbr"`
		Conference        string     `json:"conference"`
		Division          string     `json:"division"`
		Wins              int        `json:"wins"`
		Losses            int        `json:"losses"`
		Ties              int        `json:"ties"`
		WinPct            float64    `json:"win_pct"`
		PointsFor         int        `json:"points_for"`
		PointsAgainst     int        `json:"points_against"`
		PointDifferential int        `json:"point_differential"`
		HomeWins          int        `json:"home_wins"`
		HomeLosses        int        `json:"home_losses"`
		AwayWins          int        `json:"away_wins"`
		AwayLosses        int        `json:"away_losses"`
		DivisionWins      int        `json:"division_wins"`
		DivisionLosses    int        `json:"division_losses"`
		ConferenceWins    int        `json:"conference_wins"`
		ConferenceLosses  int        `json:"conference_losses"`
		CurrentStreak     *string    `json:"current_streak,omitempty"`
		DivisionRank      *int       `json:"division_rank,omitempty"`
		ConferenceRank    *int       `json:"conference_rank,omitempty"`
		PlayoffSeed       *int       `json:"playoff_seed,omitempty"`
	}

	var standings []Standing

	for rows.Next() {
		var s Standing
		err := rows.Scan(
			&s.ID, &s.TeamID, &s.TeamName, &s.TeamAbbr, &s.Conference, &s.Division,
			&s.Wins, &s.Losses, &s.Ties, &s.WinPct,
			&s.PointsFor, &s.PointsAgainst, &s.PointDifferential,
			&s.HomeWins, &s.HomeLosses,
			&s.AwayWins, &s.AwayLosses,
			&s.DivisionWins, &s.DivisionLosses,
			&s.ConferenceWins, &s.ConferenceLosses,
			&s.CurrentStreak,
			&s.DivisionRank, &s.ConferenceRank, &s.PlayoffSeed,
		)
		if err != nil {
			log.Printf("Error scanning standing: %v", err)
			continue
		}
		standings = append(standings, s)
	}

	if len(standings) == 0 {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "No standings found")
		return
	}

	response.Success(w, standings)
}
