package models

import (
	"time"

	"github.com/google/uuid"
)

// GameStat represents player statistics for a single game
type GameStat struct {
	ID                   uuid.UUID `json:"id"`
	PlayerID             uuid.UUID `json:"player_id"`
	GameID               uuid.UUID `json:"game_id"`
	TeamID               uuid.UUID `json:"team_id"`
	SeasonYear           int       `json:"season_year"`
	Week                 int       `json:"week"`
	PassingYards         int       `json:"passing_yards"`
	PassingTouchdowns    int       `json:"passing_touchdowns"`
	PassingInterceptions int       `json:"passing_interceptions"`
	PassingCompletions   int       `json:"passing_completions"`
	PassingAttempts      int       `json:"passing_attempts"`
	RushingYards         int       `json:"rushing_yards"`
	RushingTouchdowns    int       `json:"rushing_touchdowns"`
	RushingAttempts      int       `json:"rushing_attempts"`
	ReceivingYards       int       `json:"receiving_yards"`
	ReceivingTouchdowns  int       `json:"receiving_touchdowns"`
	ReceivingReceptions  int       `json:"receiving_receptions"`
	ReceivingTargets     int       `json:"receiving_targets"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// Optional nested data
	Player *Player `json:"player,omitempty"`
	Game   *Game   `json:"game,omitempty"`
}

// StatsFilters for filtering stats
type StatsFilters struct {
	Season int
	Week   int
	Limit  int
	Offset int
}

// PlayerStatLeader represents a player in a leaderboard
type PlayerStatLeader struct {
	PlayerID     uuid.UUID `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	Position     string    `json:"position"`
	JerseyNumber int       `json:"jersey_number"`
	TeamID       uuid.UUID `json:"team_id"`
	TeamAbbr     string    `json:"team_abbr"`
	Category     string    `json:"category"`
	TotalStat    int       `json:"total_stat"`
	GamesPlayed  int       `json:"games_played"`
}