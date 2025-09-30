package espn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL     = "https://site.api.espn.com"
	coreBaseURL = "https://sports.core.api.espn.com"
	webBaseURL  = "https://site.web.api.espn.com"
	userAgent   = "GridIronMind/1.0"
)

// Client handles ESPN API requests
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new ESPN API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with retries and rate limiting
func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	// Exponential backoff for retries
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
			return nil, fmt.Errorf("ESPN API returned status %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}

	return nil, fmt.Errorf("request failed after retries")
}

// FetchScoreboard fetches current NFL scoreboard
func (c *Client) FetchScoreboard(ctx context.Context) (*ScoreboardResponse, error) {
	url := fmt.Sprintf("%s/apis/site/v2/sports/football/nfl/scoreboard", baseURL)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response ScoreboardResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse scoreboard response: %w", err)
	}

	return &response, nil
}

// FetchTeamRoster fetches roster for a specific team
func (c *Client) FetchTeamRoster(ctx context.Context, teamID string) (*TeamResponse, error) {
	url := fmt.Sprintf("%s/apis/site/v2/sports/football/nfl/teams/%s?enable=roster,projection,stats", baseURL, teamID)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response TeamResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse team response: %w", err)
	}

	return &response, nil
}

// FetchAllTeams fetches all NFL teams
func (c *Client) FetchAllTeams(ctx context.Context) (*TeamsResponse, error) {
	url := fmt.Sprintf("%s/apis/site/v2/sports/football/nfl/teams", baseURL)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response TeamsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse teams response: %w", err)
	}

	return &response, nil
}

// FetchActivePlayers fetches all active NFL players
func (c *Client) FetchActivePlayers(ctx context.Context, limit int) (*PlayersResponse, error) {
	if limit == 0 {
		limit = 1000
	}
	url := fmt.Sprintf("%s/v2/sports/football/leagues/nfl/athletes?limit=%d&active=true", coreBaseURL, limit)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response PlayersResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse players response: %w", err)
	}

	return &response, nil
}

// FetchPlayerOverview fetches detailed player information
func (c *Client) FetchPlayerOverview(ctx context.Context, athleteID string) (*PlayerOverviewResponse, error) {
	url := fmt.Sprintf("%s/apis/common/v3/sports/football/nfl/athletes/%s/overview", webBaseURL, athleteID)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var response PlayerOverviewResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse player overview: %w", err)
	}

	return &response, nil
}