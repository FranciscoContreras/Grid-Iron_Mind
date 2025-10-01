package models

import (
	"time"

	"github.com/google/uuid"
)

// TeamDefensiveStats represents defensive statistics for a team in a season/week
type TeamDefensiveStats struct {
	ID     uuid.UUID  `json:"id"`
	TeamID uuid.UUID  `json:"team_id"`
	Season int        `json:"season"`
	Week   *int       `json:"week,omitempty"` // NULL for season-long stats

	// Points & Yards Allowed
	PointsAllowed          int      `json:"points_allowed"`
	PointsAllowedPerGame   *float64 `json:"points_allowed_per_game,omitempty"`
	YardsAllowed           int      `json:"yards_allowed"`
	YardsAllowedPerGame    *float64 `json:"yards_allowed_per_game,omitempty"`
	PassYardsAllowed       int      `json:"pass_yards_allowed"`
	PassYardsAllowedPerGame *float64 `json:"pass_yards_allowed_per_game,omitempty"`
	RushYardsAllowed       int      `json:"rush_yards_allowed"`
	RushYardsAllowedPerGame *float64 `json:"rush_yards_allowed_per_game,omitempty"`

	// Defensive Plays
	Sacks                    int `json:"sacks"`
	SacksYards               int `json:"sacks_yards"`
	Interceptions            int `json:"interceptions"`
	InterceptionYards        int `json:"interception_yards"`
	InterceptionTouchdowns   int `json:"interception_touchdowns"`
	ForcedFumbles            int `json:"forced_fumbles"`
	FumbleRecoveries         int `json:"fumble_recoveries"`
	FumbleRecoveryTouchdowns int `json:"fumble_recovery_touchdowns"`

	// Third Down & Red Zone Defense
	ThirdDownAttempts           int      `json:"third_down_attempts"`
	ThirdDownConversionsAllowed int      `json:"third_down_conversions_allowed"`
	ThirdDownPercentage         *float64 `json:"third_down_percentage,omitempty"`
	RedZoneAttempts             int      `json:"red_zone_attempts"`
	RedZoneTouchdownsAllowed    int      `json:"red_zone_touchdowns_allowed"`
	RedZonePercentage           *float64 `json:"red_zone_percentage,omitempty"`

	// Additional Metrics
	PassAttemptsAllowed    int `json:"pass_attempts_allowed"`
	PassCompletionsAllowed int `json:"pass_completions_allowed"`
	PassTouchdownsAllowed  int `json:"pass_touchdowns_allowed"`
	RushAttemptsAllowed    int `json:"rush_attempts_allowed"`
	RushTouchdownsAllowed  int `json:"rush_touchdowns_allowed"`
	Penalties              int `json:"penalties"`
	PenaltyYards           int `json:"penalty_yards"`

	// Rankings
	DefensiveRank     *int `json:"defensive_rank,omitempty"`
	PassDefenseRank   *int `json:"pass_defense_rank,omitempty"`
	RushDefenseRank   *int `json:"rush_defense_rank,omitempty"`
	PointsAllowedRank *int `json:"points_allowed_rank,omitempty"`

	GamesPlayed int       `json:"games_played"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Joined fields (not in DB)
	TeamAbbr *string `json:"team_abbr,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
}

// TeamDefensiveStatsVsPosition represents how a team's defense performs against specific positions
type TeamDefensiveStatsVsPosition struct {
	ID       uuid.UUID `json:"id"`
	TeamID   uuid.UUID `json:"team_id"`
	Season   int       `json:"season"`
	Week     *int      `json:"week,omitempty"`
	Position string    `json:"position"` // QB, RB, WR, TE

	// Fantasy Points Allowed
	FantasyPointsAllowedStandard *float64 `json:"fantasy_points_allowed_standard,omitempty"`
	FantasyPointsAllowedPPR      *float64 `json:"fantasy_points_allowed_ppr,omitempty"`
	FantasyPointsAllowedHalfPPR  *float64 `json:"fantasy_points_allowed_half_ppr,omitempty"`
	FantasyPointsPerGameStandard *float64 `json:"fantasy_points_per_game_standard,omitempty"`
	FantasyPointsPerGamePPR      *float64 `json:"fantasy_points_per_game_ppr,omitempty"`
	FantasyPointsPerGameHalfPPR  *float64 `json:"fantasy_points_per_game_half_ppr,omitempty"`

	// Position-Specific Stats
	PassingYardsAllowed   int `json:"passing_yards_allowed,omitempty"`
	PassingTdsAllowed     int `json:"passing_tds_allowed,omitempty"`
	InterceptionsForced   int `json:"interceptions_forced,omitempty"`
	SacksRecorded         int `json:"sacks_recorded,omitempty"`
	RushingYardsAllowed   int `json:"rushing_yards_allowed,omitempty"`
	RushingTdsAllowed     int `json:"rushing_tds_allowed,omitempty"`
	ReceptionsAllowed     int `json:"receptions_allowed,omitempty"`
	ReceivingYardsAllowed int `json:"receiving_yards_allowed,omitempty"`
	ReceivingTdsAllowed   int `json:"receiving_tds_allowed,omitempty"`

	RankVsPosition *int `json:"rank_vs_position,omitempty"`
	GamesPlayed    int  `json:"games_played"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Joined fields
	TeamAbbr *string `json:"team_abbr,omitempty"`
	TeamName *string `json:"team_name,omitempty"`
}

// PlayerVsDefenseHistory tracks a player's historical performance against specific defenses
type PlayerVsDefenseHistory struct {
	ID             uuid.UUID  `json:"id"`
	PlayerID       uuid.UUID  `json:"player_id"`
	DefenseTeamID  uuid.UUID  `json:"defense_team_id"`
	GameID         *uuid.UUID `json:"game_id,omitempty"`
	Season         int        `json:"season"`
	Week           int        `json:"week"`

	// Game Stats
	PassingYards       int `json:"passing_yards"`
	PassingTds         int `json:"passing_tds"`
	InterceptionsThrown int `json:"interceptions_thrown"`
	RushingYards       int `json:"rushing_yards"`
	RushingTds         int `json:"rushing_tds"`
	Receptions         int `json:"receptions"`
	ReceivingYards     int `json:"receiving_yards"`
	ReceivingTds       int `json:"receiving_tds"`

	// Fantasy Points
	FantasyPointsStandard *float64 `json:"fantasy_points_standard,omitempty"`
	FantasyPointsPPR      *float64 `json:"fantasy_points_ppr,omitempty"`
	FantasyPointsHalfPPR  *float64 `json:"fantasy_points_half_ppr,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PlayerVsDefenseResponse is the API response for player vs defense queries
type PlayerVsDefenseResponse struct {
	PlayerID      uuid.UUID                   `json:"player_id"`
	PlayerName    string                      `json:"player_name"`
	DefenseTeamID uuid.UUID                   `json:"defense_team_id"`
	DefenseTeamAbbr string                    `json:"defense_team_abbr"`
	Games         []PlayerVsDefenseHistory    `json:"games"`
	Averages      *PlayerVsDefenseAverages    `json:"averages,omitempty"`
}

// PlayerVsDefenseAverages contains average performance metrics
type PlayerVsDefenseAverages struct {
	GamesPlayed                  int      `json:"games_played"`
	FantasyPointsPerGameStandard *float64 `json:"fantasy_points_per_game_standard,omitempty"`
	FantasyPointsPerGamePPR      *float64 `json:"fantasy_points_per_game_ppr,omitempty"`
	FantasyPointsPerGameHalfPPR  *float64 `json:"fantasy_points_per_game_half_ppr,omitempty"`
	YardsPerGame                 *float64 `json:"yards_per_game,omitempty"`
	TouchdownsPerGame            *float64 `json:"touchdowns_per_game,omitempty"`
}

// DefensiveRanking represents a team's ranking in a specific defensive category
type DefensiveRanking struct {
	Rank     int       `json:"rank"`
	TeamID   uuid.UUID `json:"team_id"`
	TeamAbbr string    `json:"team_abbr"`
	TeamName string    `json:"team_name"`
	Category string    `json:"category"` // overall, pass, rush, points_allowed
	Value    float64   `json:"value"`    // yards/points per game
	Season   int       `json:"season"`
}
