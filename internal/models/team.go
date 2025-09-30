package models

import (
	"time"

	"github.com/google/uuid"
)

// Team represents an NFL team
type Team struct {
	ID              uuid.UUID  `json:"id"`
	NFLID           *int       `json:"nfl_id,omitempty"`
	Name            string     `json:"name" validate:"required"`
	Abbreviation    string     `json:"abbreviation" validate:"required"`
	City            string     `json:"city" validate:"required"`
	Conference      string     `json:"conference" validate:"required,oneof=AFC NFC"`
	Division        string     `json:"division" validate:"required,oneof=North South East West"`
	Stadium         *string    `json:"stadium,omitempty"`
	StadiumLat      *float64   `json:"stadium_lat,omitempty"`
	StadiumLon      *float64   `json:"stadium_lon,omitempty"`
	StadiumType     *string    `json:"stadium_type,omitempty"` // outdoor, indoor, retractable
	StadiumSurface  *string    `json:"stadium_surface,omitempty"` // grass, turf
	StadiumCapacity *int       `json:"stadium_capacity,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}