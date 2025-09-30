package models

import (
	"time"

	"github.com/google/uuid"
)

// Game represents an NFL game
type Game struct {
	ID         uuid.UUID  `json:"id"`
	NFLGameID  string     `json:"nfl_game_id" validate:"required"`
	HomeTeamID uuid.UUID  `json:"home_team_id" validate:"required"`
	AwayTeamID uuid.UUID  `json:"away_team_id" validate:"required"`
	GameDate   time.Time  `json:"game_date" validate:"required"`
	Season     int        `json:"season" validate:"required"`
	Week       int        `json:"week" validate:"required"`
	HomeScore  *int       `json:"home_score,omitempty"`
	AwayScore  *int       `json:"away_score,omitempty"`
	Status     string     `json:"status" validate:"required,oneof=scheduled live final"`
	CreatedAt  time.Time  `json:"created_at"`

	// Optional nested team data
	HomeTeam *Team `json:"home_team,omitempty"`
	AwayTeam *Team `json:"away_team,omitempty"`
}