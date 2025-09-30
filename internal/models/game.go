package models

import (
	"time"

	"github.com/google/uuid"
)

// Game represents an NFL game
type Game struct {
	ID               uuid.UUID `json:"id"`
	EspnGameID       string    `json:"espn_game_id"`
	SeasonYear       int       `json:"season_year"`
	SeasonType       int       `json:"season_type"`
	Week             int       `json:"week"`
	GameDate         time.Time `json:"game_date"`
	HomeTeamID       uuid.UUID `json:"home_team_id"`
	AwayTeamID       uuid.UUID `json:"away_team_id"`
	HomeScore        int       `json:"home_score"`
	AwayScore        int       `json:"away_score"`
	Status           string    `json:"status"`

	// Venue information
	VenueID          *string   `json:"venue_id,omitempty"`
	VenueName        *string   `json:"venue_name,omitempty"`
	VenueCity        *string   `json:"venue_city,omitempty"`
	VenueState       *string   `json:"venue_state,omitempty"`
	VenueType        *string   `json:"venue_type,omitempty"`
	Attendance       *int      `json:"attendance,omitempty"`

	// Weather conditions
	WeatherTemp      *int      `json:"weather_temp,omitempty"` // Fahrenheit
	WeatherCondition *string   `json:"weather_condition,omitempty"` // clear, rain, snow, etc.
	WeatherWindSpeed *int      `json:"weather_wind_speed,omitempty"` // mph
	WeatherHumidity  *int      `json:"weather_humidity,omitempty"` // percentage

	// Additional metadata
	GameTimeET       *string   `json:"game_time_et,omitempty"`
	PlayoffRound     *string   `json:"playoff_round,omitempty"` // wild-card, divisional, conference, super-bowl

	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Optional nested team data
	HomeTeam *Team `json:"home_team,omitempty"`
	AwayTeam *Team `json:"away_team,omitempty"`

	// Team name fields for quick access (populated from JOINs)
	HomeTeamName string `json:"home_team_name,omitempty"`
	HomeTeamAbbr string `json:"home_team_abbr,omitempty"`
	AwayTeamName string `json:"away_team_name,omitempty"`
	AwayTeamAbbr string `json:"away_team_abbr,omitempty"`
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