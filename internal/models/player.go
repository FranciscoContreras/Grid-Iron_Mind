package models

import (
	"time"

	"github.com/google/uuid"
)

// Player represents an NFL player
type Player struct {
	ID           uuid.UUID  `json:"id"`
	NFLID        *string    `json:"nfl_id,omitempty"`
	Name         string     `json:"name" validate:"required"`
	Position     string     `json:"position" validate:"required"`
	TeamID       *uuid.UUID `json:"team_id,omitempty"`
	JerseyNumber *int       `json:"jersey_number,omitempty"`
	HeightInches *int       `json:"height_inches,omitempty"`
	WeightPounds *int       `json:"weight_pounds,omitempty"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	BirthCity    *string    `json:"birth_city,omitempty"`
	BirthState   *string    `json:"birth_state,omitempty"`
	BirthCountry *string    `json:"birth_country,omitempty"`
	College      *string    `json:"college,omitempty"`
	DraftYear    *int       `json:"draft_year,omitempty"`
	DraftRound   *int       `json:"draft_round,omitempty"`
	DraftPick    *int       `json:"draft_pick,omitempty"`
	RookieYear   *int       `json:"rookie_year,omitempty"`
	YearsPro     *int       `json:"years_pro,omitempty"`
	HeadshotURL  *string    `json:"headshot_url,omitempty"`
	Status       string     `json:"status" validate:"required,oneof=active injured inactive"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Optional nested team data
	Team *Team `json:"team,omitempty"`
}

// PlayerWithStats extends Player with current season stats
type PlayerWithStats struct {
	Player
	Stats *PlayerSeasonStats `json:"stats,omitempty"`
}

// PlayerSeasonStats represents aggregated stats for a player
type PlayerSeasonStats struct {
	Season        int `json:"season"`
	GamesPlayed   int `json:"games_played"`
	PassingYards  int `json:"passing_yards"`
	RushingYards  int `json:"rushing_yards"`
	ReceivingYards int `json:"receiving_yards"`
	Touchdowns    int `json:"touchdowns"`
	Interceptions int `json:"interceptions"`
}

// PlayerCareerStats represents complete career statistics
type PlayerCareerStats struct {
	ID             uuid.UUID  `json:"id"`
	PlayerID       uuid.UUID  `json:"player_id"`
	Season         int        `json:"season"`
	TeamID         *uuid.UUID `json:"team_id,omitempty"`
	GamesPlayed    int        `json:"games_played"`
	GamesStarted   int        `json:"games_started"`

	// Passing
	PassingYards      int     `json:"passing_yards"`
	PassingTDs        int     `json:"passing_tds"`
	PassingInts       int     `json:"passing_ints"`
	PassingCompletions int    `json:"passing_completions"`
	PassingAttempts   int     `json:"passing_attempts"`
	PassingRating     *float64 `json:"passing_rating,omitempty"`

	// Rushing
	RushingYards    int  `json:"rushing_yards"`
	RushingTDs      int  `json:"rushing_tds"`
	RushingAttempts int  `json:"rushing_attempts"`
	RushingLong     *int `json:"rushing_long,omitempty"`

	// Receiving
	Receptions      int  `json:"receptions"`
	ReceivingYards  int  `json:"receiving_yards"`
	ReceivingTDs    int  `json:"receiving_tds"`
	ReceivingTargets int `json:"receiving_targets"`
	ReceivingLong   *int `json:"receiving_long,omitempty"`

	// Defensive
	Tackles         int      `json:"tackles"`
	Sacks           *float64 `json:"sacks,omitempty"`
	Interceptions   int      `json:"interceptions"`
	ForcedFumbles   int      `json:"forced_fumbles"`
	FumbleRecoveries int     `json:"fumble_recoveries"`
	PassesDefended  int      `json:"passes_defended"`

	// Kicking
	FieldGoalsMade     int `json:"field_goals_made"`
	FieldGoalsAttempted int `json:"field_goals_attempted"`
	ExtraPointsMade     int `json:"extra_points_made"`
	ExtraPointsAttempted int `json:"extra_points_attempted"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Optional nested data
	Team *Team `json:"team,omitempty"`
}

// PlayerTeamHistory represents which teams a player has played for
type PlayerTeamHistory struct {
	ID           uuid.UUID  `json:"id"`
	PlayerID     uuid.UUID  `json:"player_id"`
	TeamID       uuid.UUID  `json:"team_id"`
	SeasonStart  int        `json:"season_start"`
	SeasonEnd    *int       `json:"season_end,omitempty"`
	Position     string     `json:"position"`
	JerseyNumber *int       `json:"jersey_number,omitempty"`
	IsCurrent    bool       `json:"is_current"`
	CreatedAt    time.Time  `json:"created_at"`

	// Optional nested data
	Team *Team `json:"team,omitempty"`
}