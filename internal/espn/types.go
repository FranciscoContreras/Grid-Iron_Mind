package espn

import (
	"strings"
	"time"
)

// FlexibleTime wraps time.Time to handle multiple ESPN date formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON handles multiple date formats from ESPN API
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")

	// Try common ESPN formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04Z",           // ESPN's simplified format
		"2006-01-02T15:04:05Z",        // ESPN's format with seconds
		"2006-01-02T15:04:05.000Z",    // ESPN's format with milliseconds
		time.RFC3339Nano,
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, str)
		if err == nil {
			ft.Time = t
			return nil
		}
		lastErr = err
	}

	return lastErr
}

// SeasonType represents ESPN season type information
type SeasonType struct {
	ID           string `json:"id"`
	Type         int    `json:"type"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

// ScoreboardResponse represents the ESPN scoreboard API response
type ScoreboardResponse struct {
	Leagues []struct {
		Season struct {
			Year int        `json:"year"`
			Type SeasonType `json:"type"`
			Slug string     `json:"slug"`
		} `json:"season"`
	} `json:"leagues"`
	Events []Event `json:"events"`
}

// Event represents a game/event
type Event struct {
	ID          string    `json:"id"`
	UID         string    `json:"uid"`
	Date        FlexibleTime `json:"date"`
	Name        string    `json:"name"`
	ShortName   string    `json:"shortName"`
	Season      Season    `json:"season"`
	Week        Week      `json:"week"`
	Competitions []Competition `json:"competitions"`
	Status      Status    `json:"status"`
}

// Season represents season information (from Event)
type Season struct {
	Year int    `json:"year"`
	Type int    `json:"type"`
	Slug string `json:"slug"`
}

// Week represents week information
type Week struct {
	Number int `json:"number"`
}

// Competition represents game competition details
type Competition struct {
	ID          string       `json:"id"`
	UID         string       `json:"uid"`
	Date        FlexibleTime `json:"date"`
	Attendance  int          `json:"attendance"`
	Venue       Venue        `json:"venue"`
	Competitors []Competitor `json:"competitors"`
	Status      Status       `json:"status"`
}

// Venue represents game venue
type Venue struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Address  struct {
		City  string `json:"city"`
		State string `json:"state"`
	} `json:"address"`
}

// Competitor represents a team in a game
type Competitor struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	Type       string `json:"type"`
	Order      int    `json:"order"`
	HomeAway   string `json:"homeAway"`
	Team       TeamInfo `json:"team"`
	Score      string   `json:"score"`
	Statistics []Statistic `json:"statistics"`
}

// TeamInfo represents basic team information
type TeamInfo struct {
	ID           string `json:"id"`
	UID          string `json:"uid"`
	Location     string `json:"location"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	DisplayName  string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
	Color        string `json:"color"`
	AlternateColor string `json:"alternateColor"`
	Logo         string `json:"logo"`
}

// Statistic represents a game statistic
type Statistic struct {
	Name             string `json:"name"`
	Abbreviation     string `json:"abbreviation"`
	DisplayValue     string `json:"displayValue"`
	RankDisplayValue string `json:"rankDisplayValue,omitempty"`
}

// Status represents game status
type Status struct {
	Clock        float64 `json:"clock"`
	DisplayClock string  `json:"displayClock"`
	Period       int     `json:"period"`
	Type         struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		State       string `json:"state"`
		Completed   bool   `json:"completed"`
		Description string `json:"description"`
		Detail      string `json:"detail"`
		ShortDetail string `json:"shortDetail"`
	} `json:"type"`
}

// TeamsResponse represents the teams list API response
type TeamsResponse struct {
	Sports []struct {
		Leagues []struct {
			Teams []TeamEntry `json:"teams"`
		} `json:"leagues"`
	} `json:"sports"`
}

// TeamEntry wraps team data
type TeamEntry struct {
	Team TeamDetail `json:"team"`
}

// TeamDetail represents detailed team information
type TeamDetail struct {
	ID           string `json:"id"`
	UID          string `json:"uid"`
	Slug         string `json:"slug"`
	Location     string `json:"location"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	DisplayName  string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
	Color        string `json:"color"`
	AlternateColor string `json:"alternateColor"`
	IsActive     bool   `json:"isActive"`
	Logos        []Logo `json:"logos"`
}

// Logo represents team logo
type Logo struct {
	Href   string `json:"href"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// TeamResponse represents a single team with roster
type TeamResponse struct {
	Team   TeamDetail `json:"team"`
	Roster struct {
		Athletes []Athlete `json:"athletes"`
	} `json:"roster"`
}

// PlayersResponse represents the active players API response
type PlayersResponse struct {
	Items      []PlayerRef `json:"items"`
	Count      int         `json:"count"`
	PageIndex  int         `json:"pageIndex"`
	PageSize   int         `json:"pageSize"`
	PageCount  int         `json:"pageCount"`
}

// PlayerRef represents a player reference from the athletes list
type PlayerRef struct {
	Ref string `json:"$ref"`
}

// PlayerOverviewResponse represents detailed player information
type PlayerOverviewResponse struct {
	Athlete Athlete `json:"athlete"`
	Statistics []struct {
		Season Season `json:"season"`
		Stats  map[string]interface{} `json:"stats"`
	} `json:"statistics"`
}

// Athlete represents a player
type Athlete struct {
	ID           string    `json:"id"`
	UID          string    `json:"uid"`
	GUID         string    `json:"guid"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	FullName     string    `json:"fullName"`
	DisplayName  string    `json:"displayName"`
	ShortName    string    `json:"shortName"`
	Weight       float64   `json:"weight"`
	Height       float64   `json:"height"`
	Age          int       `json:"age"`
	DateOfBirth  string    `json:"dateOfBirth"`
	BirthPlace   Place     `json:"birthPlace"`
	Position     Position  `json:"position"`
	Jersey       string    `json:"jersey"`
	Active       bool      `json:"active"`
	Team         *TeamInfo `json:"team,omitempty"`
	Headshot     *Headshot `json:"headshot,omitempty"`
	Status       *PlayerStatus `json:"status,omitempty"`
}

// Place represents a location
type Place struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

// Position represents player position
type Position struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName"`
	Abbreviation string `json:"abbreviation"`
}

// Headshot represents player photo
type Headshot struct {
	Href string `json:"href"`
	Alt  string `json:"alt"`
}

// PlayerStatus represents player status (active, injured, etc.)
type PlayerStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PlayerStatsResponse represents career statistics response
type PlayerStatsResponse struct {
	Splits struct {
		Categories []struct {
			Name  string `json:"name"`
			Stats []struct {
				Name         string      `json:"name"`
				DisplayName  string      `json:"displayName"`
				Abbreviation string      `json:"abbreviation"`
				Value        interface{} `json:"value"`
			} `json:"stats"`
		} `json:"categories"`
	} `json:"splits"`
	SeasonTypes []struct {
		Year        int    `json:"year"`
		Type        int    `json:"type"`
		Name        string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Categories  []struct {
			Name  string `json:"name"`
			Stats []struct {
				Name         string      `json:"name"`
				DisplayName  string      `json:"displayName"`
				Abbreviation string      `json:"abbreviation"`
				Value        interface{} `json:"value"`
			} `json:"stats"`
		} `json:"categories"`
	} `json:"seasonTypes"`
}

// GameDetailResponse represents detailed game information
type GameDetailResponse struct {
	Header struct {
		ID          string      `json:"id"`
		UID         string      `json:"uid"`
		Season      Season      `json:"season"`
		Week        int         `json:"week"` // Just a number in game details API
		GameNote    string      `json:"gameNote"`
		Competition Competition `json:"competition"`
	} `json:"header"`
	BoxScore struct {
		Teams []struct {
			Team       TeamInfo `json:"team"`
			Statistics []struct {
				Name             string      `json:"name"`
				DisplayValue     string      `json:"displayValue"`
				Value            interface{} `json:"value"` // can be float64 or string "-"
				Label            string      `json:"label"`
				Abbreviation     string      `json:"abbreviation"`
			} `json:"statistics"`
		} `json:"teams"`
		Players []struct {
			Team       TeamInfo `json:"team"`
			Statistics []struct {
				Name    string `json:"name"`
				Keys    []string `json:"keys"`
				Text    string `json:"text"`
				Labels  []string `json:"labels"`
				Athletes []struct {
					Athlete Athlete  `json:"athlete"`
					Stats   []string `json:"stats"`
				} `json:"athletes"`
			} `json:"statistics"`
		} `json:"players"`
	} `json:"boxscore"`
	Weather *struct {
		DisplayValue string `json:"displayValue"`
		Temperature  int    `json:"temperature"`
		HighTemperature int `json:"highTemperature"`
		Condition    string `json:"conditionId"`
		Link         struct {
			Text string `json:"text"`
		} `json:"link"`
	} `json:"weather,omitempty"`
}