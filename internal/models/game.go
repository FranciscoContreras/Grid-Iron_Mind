package models

import (
	"time"

	"github.com/google/uuid"
)

// Game represents an NFL game
type Game struct {
	ID          uuid.UUID `json:"id"`
	EspnGameID  int       `json:"espn_game_id"`
	SeasonYear  int       `json:"season_year"`
	SeasonType  int       `json:"season_type"`
	Week        int       `json:"week"`
	GameDate    time.Time `json:"game_date"`
	HomeTeamID  uuid.UUID `json:"home_team_id"`
	AwayTeamID  uuid.UUID `json:"away_team_id"`
	HomeScore   int       `json:"home_score"`
	AwayScore   int       `json:"away_score"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Optional nested team data
	HomeTeam *Team `json:"home_team,omitempty"`
	AwayTeam *Team `json:"away_team,omitempty"`
}

// GameFilters for filtering games
type GameFilters struct {
	Season int
	Week   int
	TeamID uuid.UUID
	Status string
	Limit  int
	Offset int
}