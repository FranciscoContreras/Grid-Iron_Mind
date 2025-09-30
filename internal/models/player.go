package models

import (
	"time"

	"github.com/google/uuid"
)

// Player represents an NFL player
type Player struct {
	ID           uuid.UUID  `json:"id"`
	NFLID        *int       `json:"nfl_id,omitempty"`
	Name         string     `json:"name" validate:"required"`
	Position     string     `json:"position" validate:"required"`
	TeamID       *uuid.UUID `json:"team_id,omitempty"`
	JerseyNumber *int       `json:"jersey_number,omitempty"`
	HeightInches *int       `json:"height_inches,omitempty"`
	WeightPounds *int       `json:"weight_pounds,omitempty"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	College      *string    `json:"college,omitempty"`
	DraftYear    *int       `json:"draft_year,omitempty"`
	DraftRound   *int       `json:"draft_round,omitempty"`
	DraftPick    *int       `json:"draft_pick,omitempty"`
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