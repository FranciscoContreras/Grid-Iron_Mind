package nflverse

// CSV models for NFLverse data parsing
// These structs map directly to CSV column names

// PlayerStatCSV represents a row from player_stats.csv
type PlayerStatCSV struct {
	PlayerID        string  `csv:"player_id"`
	PlayerName      string  `csv:"player_display_name"`
	Position        string  `csv:"position"`
	PositionGroup   string  `csv:"position_group"`
	Season          int     `csv:"season"`
	Week            int     `csv:"week"`
	SeasonType      string  `csv:"season_type"`
	TeamAbbr        string  `csv:"recent_team"`
	OpponentTeam    string  `csv:"opponent_team"`

	// Passing
	Completions     int     `csv:"completions"`
	Attempts        int     `csv:"attempts"`
	PassingYards    float64 `csv:"passing_yards"`
	PassingTDs      int     `csv:"passing_tds"`
	Interceptions   int     `csv:"interceptions"`
	Sacks           float64 `csv:"sacks"`
	SackYards       float64 `csv:"sack_yards"`
	SackFumbles     int     `csv:"sack_fumbles"`
	SackFumblesLost int     `csv:"sack_fumbles_lost"`

	// Rushing
	Carries         int     `csv:"carries"`
	RushingYards    float64 `csv:"rushing_yards"`
	RushingTDs      int     `csv:"rushing_tds"`
	RushingFumbles  int     `csv:"rushing_fumbles"`
	RushingFumblesLost int  `csv:"rushing_fumbles_lost"`

	// Receiving
	Receptions      int     `csv:"receptions"`
	Targets         int     `csv:"targets"`
	ReceivingYards  float64 `csv:"receiving_yards"`
	ReceivingTDs    int     `csv:"receiving_tds"`
	ReceivingFumbles int    `csv:"receiving_fumbles"`
	ReceivingFumblesLost int `csv:"receiving_fumbles_lost"`

	// Kicking (if player is kicker)
	FieldGoalsMade      int `csv:"fg_made"`
	FieldGoalsAttempted int `csv:"fg_att"`
	ExtraPointsMade     int `csv:"pat_made"`
	ExtraPointsAttempted int `csv:"pat_att"`

	// Fantasy
	FantasyPoints       float64 `csv:"fantasy_points"`
	FantasyPointsPPR    float64 `csv:"fantasy_points_ppr"`
}

// ScheduleCSV represents a row from schedule.csv
type ScheduleCSV struct {
	Season      int    `csv:"season"`
	GameType    string `csv:"game_type"`
	Week        int    `csv:"week"`
	GameID      string `csv:"game_id"`
	OldGameID   string `csv:"old_game_id"`
	GameDay     string `csv:"gameday"`
	Weekday     string `csv:"weekday"`
	GameTime    string `csv:"gametime"`
	AwayTeam    string `csv:"away_team"`
	AwayScore   int    `csv:"away_score"`
	HomeTeam    string `csv:"home_team"`
	HomeScore   int    `csv:"home_score"`
	Location    string `csv:"location"`
	Result      int    `csv:"result"`
	Total       int    `csv:"total"`
	Overtime    int    `csv:"overtime"`
	Roof        string `csv:"roof"`
	Surface     string `csv:"surface"`
	Temp        int    `csv:"temp"`
	Wind        int    `csv:"wind"`
	AwayQBName  string `csv:"away_qb_name"`
	HomeQBName  string `csv:"home_qb_name"`
	AwayCoach   string `csv:"away_coach"`
	HomeCoach   string `csv:"home_coach"`
	RefereeName string `csv:"referee"`
	StadiumID   string `csv:"stadium_id"`
	Stadium     string `csv:"stadium"`
}

// RosterCSV represents a row from roster.csv
type RosterCSV struct {
	Season            int    `csv:"season"`
	TeamAbbr          string `csv:"team"`
	Position          string `csv:"position"`
	DepthChartPosition string `csv:"depth_chart_position"`
	JerseyNumber      int    `csv:"jersey_number"`
	Status            string `csv:"status"`
	FullName          string `csv:"full_name"`
	FirstName         string `csv:"first_name"`
	LastName          string `csv:"last_name"`
	BirthDate         string `csv:"birth_date"`
	Height            int    `csv:"height"` // in inches
	Weight            int    `csv:"weight"`
	College           string `csv:"college"`
	PlayerID          string `csv:"gsis_id"` // Our primary player ID
	ESPNPlayerID      string `csv:"espn_id"`
	YahooPlayerID     string `csv:"yahoo_id"`
	RotowirePlayerID  string `csv:"rotowire_id"`
	PFFPlayerID       string `csv:"pff_id"`
	FantasyDataID     string `csv:"fantasy_data_id"`
	SleeperPlayerID   string `csv:"sleeper_id"`
	YearsExp          int    `csv:"years_exp"`
	HeadshotURL       string `csv:"headshot_url"`
	DraftClub         string `csv:"draft_club"`
	DraftNumber       int    `csv:"draft_number"`
	EntryYear         int    `csv:"entry_year"`
	RookieYear        int    `csv:"rookie_year"`
	DraftRound        int    `csv:"draft_round"`
	DraftPick         int    `csv:"draft_pick"` // Overall pick number
}

// NextGenStatsPassingCSV represents Next Gen Stats for passing
type NextGenStatsPassingCSV struct {
	Season                      int     `csv:"season"`
	Week                        int     `csv:"week"`
	PlayerID                    string  `csv:"player_gsis_id"`
	PlayerName                  string  `csv:"player_display_name"`
	PlayerFirstName             string  `csv:"player_first_name"`
	PlayerLastName              string  `csv:"player_last_name"`
	Position                    string  `csv:"player_position"`
	TeamAbbr                    string  `csv:"team_abbr"`
	AvgTimeToThrow              float64 `csv:"avg_time_to_throw"`
	AvgCompletedAirYards        float64 `csv:"avg_completed_air_yards"`
	AvgIntendedAirYards         float64 `csv:"avg_intended_air_yards"`
	AvgAirYardsDifferential     float64 `csv:"avg_air_yards_differential"`
	MaxCompletedAirDistance     int     `csv:"max_completed_air_distance"`
	AvgAirYardsToSticks         float64 `csv:"avg_air_yards_to_sticks"`
	Attempts                    int     `csv:"attempts"`
	PassYards                   int     `csv:"pass_yards"`
	PassTouchdowns              int     `csv:"pass_touchdowns"`
	Interceptions               int     `csv:"interceptions"`
	Completions                 int     `csv:"completions"`
	CompletionPercentage        float64 `csv:"completion_percentage"`
	ExpectedCompletionPercentage float64 `csv:"expected_completion_percentage"`
	CompletionPercentageAboveExpectation float64 `csv:"completion_percentage_above_expectation"`
	PasserRating                float64 `csv:"passer_rating"`
}

// NextGenStatsRushingCSV represents Next Gen Stats for rushing
type NextGenStatsRushingCSV struct {
	Season                       int     `csv:"season"`
	Week                         int     `csv:"week"`
	PlayerID                     string  `csv:"player_gsis_id"`
	PlayerName                   string  `csv:"player_display_name"`
	PlayerFirstName              string  `csv:"player_first_name"`
	PlayerLastName               string  `csv:"player_last_name"`
	Position                     string  `csv:"player_position"`
	TeamAbbr                     string  `csv:"team_abbr"`
	Efficiency                   float64 `csv:"efficiency"`
	PercentAttemptsGteEightDefenders float64 `csv:"percent_attempts_gte_eight_defenders"`
	AvgTimeToLOS                 float64 `csv:"avg_time_to_los"`
	RushAttempts                 int     `csv:"rush_attempts"`
	RushYards                    int     `csv:"rush_yards"`
	ExpectedRushYards            int     `csv:"expected_rush_yards"`
	RushYardsOverExpected        int     `csv:"rush_yards_over_expected"`
	AvgRushYards                 float64 `csv:"avg_rush_yards"`
	RushTouchdowns               int     `csv:"rush_touchdowns"`
}

// NextGenStatsReceivingCSV represents Next Gen Stats for receiving
type NextGenStatsReceivingCSV struct {
	Season                         int     `csv:"season"`
	Week                           int     `csv:"week"`
	PlayerID                       string  `csv:"player_gsis_id"`
	PlayerName                     string  `csv:"player_display_name"`
	PlayerFirstName                string  `csv:"player_first_name"`
	PlayerLastName                 string  `csv:"player_last_name"`
	Position                       string  `csv:"player_position"`
	TeamAbbr                       string  `csv:"team_abbr"`
	AvgCushion                     float64 `csv:"avg_cushion"`
	AvgSeparation                  float64 `csv:"avg_separation"`
	AvgIntendedAirYards            float64 `csv:"avg_intended_air_yards"`
	PercentShareOfIntendedAirYards float64 `csv:"percent_share_of_intended_air_yards"`
	Receptions                     int     `csv:"receptions"`
	Targets                        int     `csv:"targets"`
	CatchPercentage                float64 `csv:"catch_percentage"`
	Yards                          int     `csv:"yards"`
	RecTouchdowns                  int     `csv:"rec_touchdowns"`
	AvgYAC                         float64 `csv:"avg_yac"`
	AvgExpectedYAC                 float64 `csv:"avg_expected_yac"`
	AvgYACAboveExpectation         float64 `csv:"avg_yac_above_expectation"`
}

// InjuryCSV represents a row from injuries.csv
type InjuryCSV struct {
	Season                  int    `csv:"season"`
	GameType                string `csv:"game_type"`
	Week                    int    `csv:"week"`
	TeamAbbr                string `csv:"team"`
	Position                string `csv:"position"`
	FullName                string `csv:"full_name"`
	FirstName               string `csv:"first_name"`
	LastName                string `csv:"last_name"`
	ReportPrimaryInjury     string `csv:"report_primary_injury"`
	ReportSecondaryInjury   string `csv:"report_secondary_injury"`
	ReportStatus            string `csv:"report_status"`
	PracticeStatus          string `csv:"practice_status"`
	DateModified            string `csv:"date_modified"`
	PlayerID                string `csv:"gsis_id"`
}
