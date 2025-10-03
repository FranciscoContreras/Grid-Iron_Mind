package yahoo

import "fmt"

// GameResponse represents a Yahoo Fantasy game
type GameResponse struct {
	Game Game `json:"fantasy_content>game" xml:"fantasy_content>game"`
}

// Game represents NFL fantasy game data
type Game struct {
	GameKey     string `json:"game_key" xml:"game_key"`
	GameID      string `json:"game_id" xml:"game_id"`
	Name        string `json:"name" xml:"name"`
	Code        string `json:"code" xml:"code"`
	Type        string `json:"type" xml:"type"`
	URL         string `json:"url" xml:"url"`
	Season      string `json:"season" xml:"season"`
	IsGameOver  int    `json:"is_game_over" xml:"is_game_over"`
	CurrentWeek int    `json:"current_week" xml:"current_week"`
}

// LeagueResponse represents league data
type LeagueResponse struct {
	League League `json:"fantasy_content>league" xml:"fantasy_content>league"`
}

// League represents a fantasy league
type League struct {
	LeagueKey      string         `json:"league_key" xml:"league_key"`
	LeagueID       string         `json:"league_id" xml:"league_id"`
	Name           string         `json:"name" xml:"name"`
	URL            string         `json:"url" xml:"url"`
	DraftStatus    string         `json:"draft_status" xml:"draft_status"`
	NumTeams       int            `json:"num_teams" xml:"num_teams"`
	CurrentWeek    int            `json:"current_week" xml:"current_week"`
	StartWeek      int            `json:"start_week" xml:"start_week"`
	EndWeek        int            `json:"end_week" xml:"end_week"`
	Season         int            `json:"season" xml:"season"`
	IsFinished     int            `json:"is_finished" xml:"is_finished"`
	ScoringType    string         `json:"scoring_type" xml:"scoring_type"`
	Standings      *Standings     `json:"standings,omitempty" xml:"standings,omitempty"`
	Scoreboard     *Scoreboard    `json:"scoreboard,omitempty" xml:"scoreboard,omitempty"`
}

// Standings represents league standings
type Standings struct {
	Teams []Team `json:"teams>team" xml:"teams>team"`
}

// Scoreboard represents weekly matchups
type Scoreboard struct {
	Week     int       `json:"week" xml:"week"`
	Matchups []Matchup `json:"matchups>matchup" xml:"matchups>matchup"`
}

// Matchup represents a head-to-head matchup
type Matchup struct {
	Week       int       `json:"week" xml:"week"`
	WeekStart  string    `json:"week_start" xml:"week_start"`
	WeekEnd    string    `json:"week_end" xml:"week_end"`
	Status     string    `json:"status" xml:"status"`
	IsPlayoffs int       `json:"is_playoffs" xml:"is_playoffs"`
	Teams      []Team    `json:"teams>team" xml:"teams>team"`
}

// TeamResponse represents team data
type TeamResponse struct {
	Team Team `json:"fantasy_content>team" xml:"fantasy_content>team"`
}

// Team represents a fantasy team
type Team struct {
	TeamKey         string          `json:"team_key" xml:"team_key"`
	TeamID          string          `json:"team_id" xml:"team_id"`
	Name            string          `json:"name" xml:"name"`
	IsOwnedByUser   int             `json:"is_owned_by_current_user" xml:"is_owned_by_current_user"`
	URL             string          `json:"url" xml:"url"`
	TeamLogos       []TeamLogo      `json:"team_logos>team_logo" xml:"team_logos>team_logo"`
	WaiverPriority  int             `json:"waiver_priority" xml:"waiver_priority"`
	NumberOfMoves   int             `json:"number_of_moves" xml:"number_of_moves"`
	NumberOfTrades  int             `json:"number_of_trades" xml:"number_of_trades"`
	Managers        []Manager       `json:"managers>manager" xml:"managers>manager"`
	Standings       *TeamStandings  `json:"team_standings,omitempty" xml:"team_standings,omitempty"`
	Roster          *Roster         `json:"roster,omitempty" xml:"roster,omitempty"`
	TeamPoints      *TeamPoints     `json:"team_points,omitempty" xml:"team_points,omitempty"`
	TeamProjected   *TeamPoints     `json:"team_projected_points,omitempty" xml:"team_projected_points,omitempty"`
}

// TeamLogo represents team logo URLs
type TeamLogo struct {
	Size string `json:"size" xml:"size"`
	URL  string `json:"url" xml:"url"`
}

// Manager represents a team manager
type Manager struct {
	ManagerID       string `json:"manager_id" xml:"manager_id"`
	Nickname        string `json:"nickname" xml:"nickname"`
	GUID            string `json:"guid" xml:"guid"`
	IsCommissioner  int    `json:"is_commissioner" xml:"is_commissioner"`
	IsCurrentLogin  int    `json:"is_current_login" xml:"is_current_login"`
	Email           string `json:"email" xml:"email"`
	ImageURL        string `json:"image_url" xml:"image_url"`
}

// TeamStandings represents team standings info
type TeamStandings struct {
	Rank            int     `json:"rank" xml:"rank"`
	PlayoffSeed     int     `json:"playoff_seed" xml:"playoff_seed"`
	OutcomeTotals   Outcome `json:"outcome_totals" xml:"outcome_totals"`
	PointsFor       float64 `json:"points_for,string" xml:"points_for"`
	PointsAgainst   float64 `json:"points_against,string" xml:"points_against"`
}

// Outcome represents win/loss/tie record
type Outcome struct {
	Wins       int     `json:"wins,string" xml:"wins"`
	Losses     int     `json:"losses,string" xml:"losses"`
	Ties       int     `json:"ties,string" xml:"ties"`
	Percentage float64 `json:"percentage,string" xml:"percentage"`
}

// TeamPoints represents team scoring
type TeamPoints struct {
	CoverageType string  `json:"coverage_type" xml:"coverage_type"`
	Week         int     `json:"week,string" xml:"week"`
	Total        float64 `json:"total,string" xml:"total"`
}

// Roster represents team roster
type Roster struct {
	Week    int      `json:"week,string" xml:"week"`
	Players []Player `json:"players>player" xml:"players>player"`
}

// PlayersResponse represents player collection
type PlayersResponse struct {
	Players []Player `json:"fantasy_content>players>player" xml:"fantasy_content>players>player"`
}

// Player represents NFL player fantasy data
type Player struct {
	PlayerKey             string              `json:"player_key" xml:"player_key"`
	PlayerID              string              `json:"player_id" xml:"player_id"`
	Name                  PlayerName          `json:"name" xml:"name"`
	EditorialPlayerKey    string              `json:"editorial_player_key" xml:"editorial_player_key"`
	EditorialTeamKey      string              `json:"editorial_team_key" xml:"editorial_team_key"`
	EditorialTeamFullName string              `json:"editorial_team_full_name" xml:"editorial_team_full_name"`
	EditorialTeamAbbr     string              `json:"editorial_team_abbr" xml:"editorial_team_abbr"`
	ByeWeeks              ByeWeeks            `json:"bye_weeks" xml:"bye_weeks"`
	UniformNumber         string              `json:"uniform_number" xml:"uniform_number"`
	DisplayPosition       string              `json:"display_position" xml:"display_position"`
	Headshot              Headshot            `json:"headshot" xml:"headshot"`
	ImageURL              string              `json:"image_url" xml:"image_url"`
	IsUndroppable         int                 `json:"is_undroppable" xml:"is_undroppable"`
	PositionType          string              `json:"position_type" xml:"position_type"`
	EligiblePositions     []string            `json:"eligible_positions>position" xml:"eligible_positions>position"`
	HasPlayerNotes        int                 `json:"has_player_notes" xml:"has_player_notes"`
	PlayerNotes           *PlayerNotes        `json:"player_notes_last_timestamp,omitempty" xml:"player_notes_last_timestamp,omitempty"`
	SelectedPosition      *SelectedPosition   `json:"selected_position,omitempty" xml:"selected_position,omitempty"`
	PlayerStats           *PlayerStats        `json:"player_stats,omitempty" xml:"player_stats,omitempty"`
	PlayerPoints          *PlayerPoints       `json:"player_points,omitempty" xml:"player_points,omitempty"`
	PlayerProjected       *PlayerStats        `json:"player_projected_stats,omitempty" xml:"player_projected_stats,omitempty"`
	OwnershipStatus       string              `json:"ownership>ownership_type" xml:"ownership>ownership_type"`
	PercentOwned          float64             `json:"percent_owned>value,string" xml:"percent_owned>value"`
}

// PlayerName represents player name components
type PlayerName struct {
	Full   string `json:"full" xml:"full"`
	First  string `json:"first" xml:"first"`
	Last   string `json:"last" xml:"last"`
	ASCII  string `json:"ascii_first" xml:"ascii_first"`
}

// ByeWeeks represents player bye weeks
type ByeWeeks struct {
	Week int `json:"week,string" xml:"week"`
}

// Headshot represents player headshot image
type Headshot struct {
	URL  string `json:"url" xml:"url"`
	Size string `json:"size" xml:"size"`
}

// PlayerNotes represents player news/notes
type PlayerNotes struct {
	Timestamp int64  `json:",string" xml:",chardata"`
	Note      string `json:"note" xml:"note"`
}

// SelectedPosition represents roster position
type SelectedPosition struct {
	CoverageType string `json:"coverage_type" xml:"coverage_type"`
	Week         int    `json:"week,string" xml:"week"`
	Position     string `json:"position" xml:"position"`
}

// PlayerStats represents player statistics
type PlayerStats struct {
	CoverageType string `json:"coverage_type" xml:"coverage_type"`
	Week         int    `json:"week,string" xml:"week"`
	Stats        []Stat `json:"stats>stat" xml:"stats>stat"`
}

// PlayerPoints represents fantasy points
type PlayerPoints struct {
	CoverageType string  `json:"coverage_type" xml:"coverage_type"`
	Week         int     `json:"week,string" xml:"week"`
	Total        float64 `json:"total,string" xml:"total"`
}

// Stat represents a single statistic
type Stat struct {
	StatID int     `json:"stat_id,string" xml:"stat_id"`
	Value  float64 `json:"value,string" xml:"value"`
}

// Transaction represents league transactions
type Transaction struct {
	TransactionKey  string          `json:"transaction_key" xml:"transaction_key"`
	TransactionID   string          `json:"transaction_id" xml:"transaction_id"`
	Type            string          `json:"type" xml:"type"`
	Status          string          `json:"status" xml:"status"`
	Timestamp       int64           `json:"timestamp,string" xml:"timestamp"`
	Players         []Player        `json:"players>player" xml:"players>player"`
	FAABBid         int             `json:"faab_bid,string" xml:"faab_bid"`
}

// DraftResult represents draft pick information
type DraftResult struct {
	Pick       int    `json:"pick,string" xml:"pick"`
	Round      int    `json:"round,string" xml:"round"`
	TeamKey    string `json:"team_key" xml:"team_key"`
	PlayerKey  string `json:"player_key" xml:"player_key"`
}

// StatCategory represents a fantasy scoring category
type StatCategory struct {
	StatID           int     `json:"stat_id" xml:"stat_id"`
	Enabled          int     `json:"enabled" xml:"enabled"`
	Name             string  `json:"name" xml:"name"`
	DisplayName      string  `json:"display_name" xml:"display_name"`
	SortOrder        int     `json:"sort_order" xml:"sort_order"`
	PositionType     string  `json:"position_type" xml:"position_type"`
	StatPositionType []string `json:"stat_position_types>stat_position_type" xml:"stat_position_types>stat_position_type"`
	IsOnlyDisplayStat int    `json:"is_only_display_stat" xml:"is_only_display_stat"`
}

// StatModifier represents scoring rules
type StatModifier struct {
	StatID int     `json:"stat_id,string" xml:"stat_id"`
	Value  float64 `json:"value,string" xml:"value"`
}

// NFL Stat ID Constants (common Yahoo Fantasy Football stat IDs)
const (
	StatPassingYards        = 4
	StatPassingTouchdowns   = 5
	StatInterceptions       = 6
	StatRushingYards        = 9
	StatRushingTouchdowns   = 10
	StatReceptions          = 11
	StatReceivingYards      = 12
	StatReceivingTouchdowns = 13
	StatReturnTouchdowns    = 14
	StatTwoPointConversions = 15
	StatFumblesLost         = 18
	StatFieldGoals0to19     = 19
	StatFieldGoals20to29    = 20
	StatFieldGoals30to39    = 21
	StatFieldGoals40to49    = 22
	StatFieldGoals50Plus    = 23
	StatPATMade             = 24
	StatPointsAllowed0      = 25
	StatPointsAllowed1to6   = 26
	StatPointsAllowed7to13  = 27
	StatPointsAllowed14to20 = 28
	StatPointsAllowed21to27 = 29
	StatPointsAllowed28to34 = 30
	StatPointsAllowed35Plus = 31
	StatSacks               = 32
	StatInterceptionsDefense = 33
	StatFumblesRecovered    = 34
	StatTouchdownsDefense   = 35
	StatSafeties            = 36
	StatBlockedKicks        = 37
	StatYardsAllowed0to99   = 38
	StatYardsAllowed100to199 = 39
	StatYardsAllowed200to299 = 40
	StatYardsAllowed300to399 = 41
	StatYardsAllowed400to499 = 42
	StatYardsAllowed500Plus = 43
)

// Helper function to get stat name by ID
func GetStatName(statID int) string {
	statNames := map[int]string{
		StatPassingYards:        "Passing Yards",
		StatPassingTouchdowns:   "Passing Touchdowns",
		StatInterceptions:       "Interceptions",
		StatRushingYards:        "Rushing Yards",
		StatRushingTouchdowns:   "Rushing Touchdowns",
		StatReceptions:          "Receptions",
		StatReceivingYards:      "Receiving Yards",
		StatReceivingTouchdowns: "Receiving Touchdowns",
		StatReturnTouchdowns:    "Return Touchdowns",
		StatTwoPointConversions: "Two-Point Conversions",
		StatFumblesLost:         "Fumbles Lost",
		StatSacks:               "Sacks",
		StatFumblesRecovered:    "Fumbles Recovered",
		StatTouchdownsDefense:   "Defensive/Special Teams Touchdowns",
		StatSafeties:            "Safeties",
		StatBlockedKicks:        "Blocked Kicks",
	}

	if name, ok := statNames[statID]; ok {
		return name
	}
	return fmt.Sprintf("Stat %d", statID)
}
