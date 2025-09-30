package models

import (
	"time"

	"github.com/google/uuid"
)

// PlayerInjury represents an injury report for a player
type PlayerInjury struct {
	ID                  uuid.UUID  `json:"id"`
	PlayerID            uuid.UUID  `json:"player_id"`
	TeamID              *uuid.UUID `json:"team_id,omitempty"`
	GameID              *uuid.UUID `json:"game_id,omitempty"`
	Status              string     `json:"status"`
	StatusAbbreviation  string     `json:"status_abbreviation,omitempty"`
	InjuryType          string     `json:"injury_type,omitempty"`
	BodyLocation        string     `json:"body_location,omitempty"`
	Detail              string     `json:"detail,omitempty"`
	Side                string     `json:"side,omitempty"`
	InjuryDate          *time.Time `json:"injury_date,omitempty"`
	ReturnDate          *time.Time `json:"return_date,omitempty"`
	ESPNInjuryID        string     `json:"espn_injury_id,omitempty"`
	LastUpdated         time.Time  `json:"last_updated"`
	CreatedAt           time.Time  `json:"created_at"`

	// Related entities
	Player *Player `json:"player,omitempty"`
	Team   *Team   `json:"team,omitempty"`
}
