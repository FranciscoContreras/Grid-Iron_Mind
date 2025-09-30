package models

import (
	"time"

	"github.com/google/uuid"
)

// GameStats represents player statistics for a single game
type GameStats struct {
	ID             uuid.UUID `json:"id"`
	PlayerID       uuid.UUID `json:"player_id" validate:"required"`
	GameID         uuid.UUID `json:"game_id" validate:"required"`
	Season         int       `json:"season" validate:"required"`
	Week           int       `json:"week" validate:"required"`
	PassingYards   int       `json:"passing_yards"`
	RushingYards   int       `json:"rushing_yards"`
	ReceivingYards int       `json:"receiving_yards"`
	Touchdowns     int       `json:"touchdowns"`
	Interceptions  int       `json:"interceptions"`
	Completions    int       `json:"completions"`
	Attempts       int       `json:"attempts"`
	Targets        int       `json:"targets"`
	Receptions     int       `json:"receptions"`
	CreatedAt      time.Time `json:"created_at"`

	// Optional nested data
	Player *Player `json:"player,omitempty"`
	Game   *Game   `json:"game,omitempty"`
}

// StatLeader represents a player in a leaderboard
type StatLeader struct {
	PlayerID     uuid.UUID `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	Position     string    `json:"position"`
	TeamID       uuid.UUID `json:"team_id"`
	TeamAbbr     string    `json:"team_abbreviation"`
	StatValue    int       `json:"stat_value"`
	GamesPlayed  int       `json:"games_played"`
}