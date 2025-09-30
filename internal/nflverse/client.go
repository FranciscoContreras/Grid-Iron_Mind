package nflverse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// NFLverse hosts data as CSV files on GitHub releases
	baseURL   = "https://github.com/nflverse/nflverse-data/releases/download"
	userAgent = "GridIronMind/1.0"
)

// Client handles nflverse API requests
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new nflverse API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nflverse API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// FetchPlayerStats fetches player statistics for a given season
// NOTE: NFLverse data is available as CSV files, not JSON API
// This function returns an error indicating the feature is not yet implemented
func (c *Client) FetchPlayerStats(ctx context.Context, season int) ([]PlayerStats, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchWeeklyStats fetches weekly player statistics
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchWeeklyStats(ctx context.Context, season int, week int) ([]PlayerStats, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchSchedule fetches game schedule for a season
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchSchedule(ctx context.Context, season int) ([]Schedule, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchRosters fetches team rosters for a season
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchRosters(ctx context.Context, season int) ([]Roster, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchPlayByPlay fetches play-by-play data for a season
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchPlayByPlay(ctx context.Context, season int, week int) ([]PlayByPlay, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchNextGenStats fetches Next Gen Stats data
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchNextGenStats(ctx context.Context, season int, statType string) ([]NextGenStats, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchDepthCharts fetches team depth charts
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchDepthCharts(ctx context.Context, season int) ([]DepthChart, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}

// FetchInjuries fetches injury reports
// NOTE: Not yet implemented - requires CSV parsing
func (c *Client) FetchInjuries(ctx context.Context, season int, week int) ([]Injury, error) {
	return nil, fmt.Errorf("NFLverse integration requires CSV parsing - not yet implemented")
}