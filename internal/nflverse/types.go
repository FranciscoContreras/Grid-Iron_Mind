package nflverse

// PlayerStats represents comprehensive player statistics from nflverse
type PlayerStats struct {
	PlayerID        string  `json:"player_id"`
	PlayerName      string  `json:"player_name"`
	PlayerDisplayName string `json:"player_display_name"`
	Position        string  `json:"position"`
	PositionGroup   string  `json:"position_group"`
	HeadshotURL     string  `json:"headshot_url"`
	Season          int     `json:"season"`
	Week            int     `json:"week"`
	SeasonType      string  `json:"season_type"`
	TeamAbbr        string  `json:"recent_team"`
	OpponentTeam    string  `json:"opponent_team"`

	// Completions and attempts
	Completions     int     `json:"completions"`
	Attempts        int     `json:"attempts"`
	PassingYards    float64 `json:"passing_yards"`
	PassingTDs      int     `json:"passing_tds"`
	Interceptions   int     `json:"interceptions"`
	Sacks           float64 `json:"sacks"`
	SackYards       float64 `json:"sack_yards"`
	SackFumbles     int     `json:"sack_fumbles"`
	SackFumblesLost int     `json:"sack_fumbles_lost"`
	PassingAirYards float64 `json:"passing_air_yards"`
	PassingYardsAfterCatch float64 `json:"passing_yards_after_catch"`
	PassingFirstDowns int   `json:"passing_first_downs"`
	PassingEPA      float64 `json:"passing_epa"`
	Passing2ptConversions int `json:"passing_2pt_conversions"`

	// Rushing stats
	Carries         int     `json:"carries"`
	RushingYards    float64 `json:"rushing_yards"`
	RushingTDs      int     `json:"rushing_tds"`
	RushingFumbles  int     `json:"rushing_fumbles"`
	RushingFumblesLost int  `json:"rushing_fumbles_lost"`
	RushingFirstDowns int   `json:"rushing_first_downs"`
	RushingEPA      float64 `json:"rushing_epa"`
	Rushing2ptConversions int `json:"rushing_2pt_conversions"`

	// Receiving stats
	Receptions      int     `json:"receptions"`
	Targets         int     `json:"targets"`
	ReceivingYards  float64 `json:"receiving_yards"`
	ReceivingTDs    int     `json:"receiving_tds"`
	ReceivingFumbles int    `json:"receiving_fumbles"`
	ReceivingFumblesLost int `json:"receiving_fumbles_lost"`
	ReceivingAirYards float64 `json:"receiving_air_yards"`
	ReceivingYardsAfterCatch float64 `json:"receiving_yards_after_catch"`
	ReceivingFirstDowns int `json:"receiving_first_downs"`
	ReceivingEPA    float64 `json:"receiving_epa"`
	Receiving2ptConversions int `json:"receiving_2pt_conversions"`

	// Fantasy points
	FantasyPoints        float64 `json:"fantasy_points"`
	FantasyPointsPPR     float64 `json:"fantasy_points_ppr"`
	FantasyPointsHalfPPR float64 `json:"fantasy_points_half_ppr"`
}

// Schedule represents game schedule data
type Schedule struct {
	Season      int    `json:"season"`
	GameType    string `json:"game_type"`
	Week        int    `json:"week"`
	GameID      string `json:"game_id"`
	OldGameID   string `json:"old_game_id"`
	GameDate    string `json:"gameday"`
	Weekday     string `json:"weekday"`
	GameTime    string `json:"gametime"`
	AwayTeam    string `json:"away_team"`
	AwayScore   int    `json:"away_score"`
	HomeTeam    string `json:"home_team"`
	HomeScore   int    `json:"home_score"`
	Location    string `json:"location"`
	Result      int    `json:"result"`
	Total       int    `json:"total"`
	Overtime    int    `json:"overtime"`
	Roof        string `json:"roof"`
	Surface     string `json:"surface"`
	Temp        int    `json:"temp"`
	Wind        int    `json:"wind"`
	AwayQBName  string `json:"away_qb_name"`
	HomeQBName  string `json:"home_qb_name"`
	AwayCoach   string `json:"away_coach"`
	HomeCoach   string `json:"home_coach"`
	StadiumID   string `json:"stadium_id"`
	Stadium     string `json:"stadium"`
}

// Roster represents player roster information
type Roster struct {
	Season            int     `json:"season"`
	TeamAbbr          string  `json:"team"`
	Position          string  `json:"position"`
	DepthChartPosition string `json:"depth_chart_position"`
	JerseyNumber      int     `json:"jersey_number"`
	Status            string  `json:"status"`
	PlayerName        string  `json:"full_name"`
	FirstName         string  `json:"first_name"`
	LastName          string  `json:"last_name"`
	BirthDate         string  `json:"birth_date"`
	HeightFeet        int     `json:"height"`
	Weight            int     `json:"weight"`
	College           string  `json:"college"`
	PlayerID          string  `json:"gsis_id"`
	ESPNPlayerID      string  `json:"espn_id"`
	YahooPlayerID     string  `json:"yahoo_id"`
	RotowirePlayerID  string  `json:"rotowire_id"`
	PFFPlayerID       string  `json:"pff_id"`
	FantasyDataID     string  `json:"fantasy_data_id"`
	SleeperPlayerID   string  `json:"sleeper_id"`
	Years             int     `json:"years_exp"`
	HeadshotURL       string  `json:"headshot_url"`
	DraftClub         string  `json:"draft_club"`
	DraftNumber       int     `json:"draft_number"`
}

// PlayByPlay represents detailed play-by-play data
type PlayByPlay struct {
	PlayID              string  `json:"play_id"`
	GameID              string  `json:"game_id"`
	HomeTeam            string  `json:"home_team"`
	AwayTeam            string  `json:"away_team"`
	Season              int     `json:"season"`
	Week                int     `json:"week"`
	PossessionTeam      string  `json:"posteam"`
	DefensiveTeam       string  `json:"defteam"`
	Quarter             int     `json:"qtr"`
	Down                int     `json:"down"`
	YardsToGo           int     `json:"ydstogo"`
	YardLine            int     `json:"yardline_100"`
	GameSeconds         int     `json:"game_seconds_remaining"`
	PlayType            string  `json:"play_type"`
	PlayTypeNFL         string  `json:"play_type_nfl"`
	Description         string  `json:"desc"`
	Yards               int     `json:"yards_gained"`
	PasserID            string  `json:"passer_player_id"`
	PasserName          string  `json:"passer_player_name"`
	ReceiverID          string  `json:"receiver_player_id"`
	ReceiverName        string  `json:"receiver_player_name"`
	RusherID            string  `json:"rusher_player_id"`
	RusherName          string  `json:"rusher_player_name"`
	PassLength          string  `json:"pass_length"`
	PassLocation        string  `json:"pass_location"`
	AirYards            float64 `json:"air_yards"`
	YardsAfterCatch     float64 `json:"yards_after_catch"`
	RunLocation         string  `json:"run_location"`
	RunGap              string  `json:"run_gap"`
	EPA                 float64 `json:"epa"`
	WPA                 float64 `json:"wpa"`
	SuccessPlay         int     `json:"success"`
	FirstDownPlay       int     `json:"first_down"`
	Touchdown           int     `json:"touchdown"`
	PassTD              int     `json:"pass_touchdown"`
	RushTD              int     `json:"rush_touchdown"`
	Interception        int     `json:"interception"`
	Fumble              int     `json:"fumble"`
	CompletedPass       int     `json:"complete_pass"`
	Sack                int     `json:"sack"`
	Penalty             int     `json:"penalty"`
	Score               int     `json:"score"`
	PossessionTeamScore int     `json:"posteam_score_post"`
	DefensiveTeamScore  int     `json:"defteam_score_post"`
}

// NextGenStats represents Next Gen Stats data
type NextGenStats struct {
	Season        int     `json:"season"`
	Week          int     `json:"week"`
	PlayerID      string  `json:"player_gsis_id"`
	PlayerName    string  `json:"player_display_name"`
	PlayerFirstName string `json:"player_first_name"`
	PlayerLastName  string `json:"player_last_name"`
	Position      string  `json:"player_position"`
	TeamAbbr      string  `json:"team_abbr"`

	// Passing NGS
	AvgTimeToThrow      float64 `json:"avg_time_to_throw"`
	AvgCompletedAirYards float64 `json:"avg_completed_air_yards"`
	AvgIntendedAirYards float64 `json:"avg_intended_air_yards"`
	AvgAirYardsDifferential float64 `json:"avg_air_yards_differential"`
	MaxCompletedAirDistance float64 `json:"max_completed_air_distance"`
	AvgAirYardsToSticks float64 `json:"avg_air_yards_to_sticks"`

	// Rushing NGS
	EfficiencyRate      float64 `json:"efficiency"`
	PercentAttemptsGTE8Defenders float64 `json:"percent_attempts_gte_8_defenders"`
	AvgTimeToLOS        float64 `json:"avg_time_to_los"`
	RushYardsOverExpected float64 `json:"rush_yards_over_expected"`
	AvgRushYards        float64 `json:"avg_rush_yards"`
	RushPct             float64 `json:"rush_pct"`

	// Receiving NGS
	AvgCushion          float64 `json:"avg_cushion"`
	AvgSeparation       float64 `json:"avg_separation"`
	AvgIntendedAirYardsReceiving float64 `json:"avg_intended_air_yards"`
	PercentShareOfIntendedAirYards float64 `json:"percent_share_of_intended_air_yards"`
	CatchPct            float64 `json:"catch_percentage"`
	AvgYAC              float64 `json:"avg_yac"`
	AvgExpectedYAC      float64 `json:"avg_expected_yac"`
	AvgYACAboveExpectation float64 `json:"avg_yac_above_expectation"`
}

// DepthChart represents team depth chart information
type DepthChart struct {
	Season           int    `json:"season"`
	ClubCode         string `json:"club_code"`
	Week             int    `json:"week"`
	GameType         string `json:"game_type"`
	DepthTeam        string `json:"depth_team"`
	LastName         string `json:"last_name"`
	FirstName        string `json:"first_name"`
	FormationRank    int    `json:"formation_rank"`
	Position         string `json:"position"`
	JerseyNumber     int    `json:"jersey_number"`
	PlayerID         string `json:"gsis_id"`
	ESPNPlayerID     string `json:"espn_id"`
	PlayerFirstName  string `json:"player_first_name"`
	PlayerLastName   string `json:"player_last_name"`
	PlayerFullName   string `json:"full_name"`
}

// Injury represents injury report data
type Injury struct {
	Season           int    `json:"season"`
	GameType         string `json:"game_type"`
	Week             int    `json:"week"`
	TeamAbbr         string `json:"team"`
	Position         string `json:"position"`
	PlayerName       string `json:"full_name"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	ReportPrimaryInjury string `json:"report_primary_injury"`
	ReportSecondaryInjury string `json:"report_secondary_injury"`
	ReportStatus     string `json:"report_status"`
	PracticeStatus   string `json:"practice_status"`
	DateModified     string `json:"date_modified"`
	PlayerID         string `json:"gsis_id"`
}