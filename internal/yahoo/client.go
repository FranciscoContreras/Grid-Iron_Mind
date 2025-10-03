package yahoo

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	baseURL       = "https://fantasysports.yahooapis.com/fantasy/v2"
	authURL       = "https://api.login.yahoo.com/oauth2/request_auth"
	tokenURL      = "https://api.login.yahoo.com/oauth2/get_token"
	nflGameCode   = "nfl"
	userAgent     = "GridIronMind/1.0"
	currentSeason = "2025"
)

// Client handles Yahoo Fantasy Sports API requests with OAuth2
type Client struct {
	httpClient *http.Client
	config     *oauth2.Config
	token      *oauth2.Token
}

// Config holds Yahoo API configuration
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewClient creates a new Yahoo Fantasy Sports API client
func NewClient(cfg Config) *Client {
	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: []string{},
	}

	return &Client{
		config: oauth2Config,
	}
}

// SetToken sets the OAuth2 token for authenticated requests
func (c *Client) SetToken(token *oauth2.Token) {
	c.token = token
	c.httpClient = c.config.Client(context.Background(), token)
}

// GetAuthURL returns the OAuth2 authorization URL
func (c *Client) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state)
}

// ExchangeCode exchanges an authorization code for an access token
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	c.SetToken(token)
	return token, nil
}

// doRequest performs an authenticated HTTP request
func (c *Client) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("client not authenticated - call SetToken first")
	}

	reqURL := fmt.Sprintf("%s%s", baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	// Retry with exponential backoff
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Handle rate limiting
		if resp.StatusCode == 429 {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("rate limited after %d attempts", maxRetries)
			}
			time.Sleep(time.Duration(attempt+2) * 2 * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Yahoo API returned status %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}

	return nil, fmt.Errorf("request failed after retries")
}

// FetchNFLGame retrieves NFL game information for the current season
func (c *Client) FetchNFLGame(ctx context.Context) (*GameResponse, error) {
	endpoint := fmt.Sprintf("/game/%s", nflGameCode)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response GameResponse
	if err := json.Unmarshal(body, &response); err != nil {
		// Try XML parsing as fallback (Yahoo sometimes returns XML)
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse game response: %w", err)
		}
	}

	return &response, nil
}

// FetchPlayerStats retrieves player statistics
func (c *Client) FetchPlayerStats(ctx context.Context, playerKeys []string) (*PlayersResponse, error) {
	if len(playerKeys) == 0 {
		return nil, fmt.Errorf("no player keys provided")
	}

	// Yahoo API supports batch requests with comma-separated keys
	keys := strings.Join(playerKeys, ",")
	endpoint := fmt.Sprintf("/players;player_keys=%s/stats", url.QueryEscape(keys))

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response PlayersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse players response: %w", err)
		}
	}

	return &response, nil
}

// FetchLeagueStandings retrieves league standings
func (c *Client) FetchLeagueStandings(ctx context.Context, leagueKey string) (*LeagueResponse, error) {
	endpoint := fmt.Sprintf("/league/%s/standings", leagueKey)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response LeagueResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse league response: %w", err)
		}
	}

	return &response, nil
}

// FetchPlayerRankings retrieves fantasy player rankings
func (c *Client) FetchPlayerRankings(ctx context.Context, position string, week int) (*PlayersResponse, error) {
	endpoint := fmt.Sprintf("/game/%s/players;position=%s", nflGameCode, position)
	if week > 0 {
		endpoint += fmt.Sprintf(";week=%d", week)
	}
	endpoint += "/stats"

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response PlayersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse rankings response: %w", err)
		}
	}

	return &response, nil
}

// FetchWeeklyMatchups retrieves matchup data for a specific week
func (c *Client) FetchWeeklyMatchups(ctx context.Context, leagueKey string, week int) (*LeagueResponse, error) {
	endpoint := fmt.Sprintf("/league/%s/scoreboard;week=%d", leagueKey, week)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response LeagueResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse matchups response: %w", err)
		}
	}

	return &response, nil
}

// FetchPlayerProjections retrieves player projections for a week
func (c *Client) FetchPlayerProjections(ctx context.Context, playerKeys []string, week int) (*PlayersResponse, error) {
	if len(playerKeys) == 0 {
		return nil, fmt.Errorf("no player keys provided")
	}

	keys := strings.Join(playerKeys, ",")
	endpoint := fmt.Sprintf("/players;player_keys=%s;type=week;week=%d/stats", url.QueryEscape(keys), week)

	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response PlayersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse projections response: %w", err)
		}
	}

	return &response, nil
}

// SearchPlayers searches for players by name
func (c *Client) SearchPlayers(ctx context.Context, searchQuery string) (*PlayersResponse, error) {
	endpoint := fmt.Sprintf("/game/%s/players;search=%s", nflGameCode, url.QueryEscape(searchQuery))
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response PlayersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		if err := xml.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse search response: %w", err)
		}
	}

	return &response, nil
}
