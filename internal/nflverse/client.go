package nflverse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL   = "https://nflreadr.nflverse.com"
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
func (c *Client) FetchPlayerStats(ctx context.Context, season int) ([]PlayerStats, error) {
	url := fmt.Sprintf("%s/player_stats?season=%d", baseURL, season)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var stats []PlayerStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse player stats: %w", err)
	}

	return stats, nil
}

// FetchWeeklyStats fetches weekly player statistics
func (c *Client) FetchWeeklyStats(ctx context.Context, season int, week int) ([]PlayerStats, error) {
	url := fmt.Sprintf("%s/player_stats?season=%d&week=%d", baseURL, season, week)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var stats []PlayerStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse weekly stats: %w", err)
	}

	return stats, nil
}

// FetchSchedule fetches game schedule for a season
func (c *Client) FetchSchedule(ctx context.Context, season int) ([]Schedule, error) {
	url := fmt.Sprintf("%s/schedules?season=%d", baseURL, season)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var schedule []Schedule
	if err := json.Unmarshal(body, &schedule); err != nil {
		return nil, fmt.Errorf("failed to parse schedule: %w", err)
	}

	return schedule, nil
}

// FetchRosters fetches team rosters for a season
func (c *Client) FetchRosters(ctx context.Context, season int) ([]Roster, error) {
	url := fmt.Sprintf("%s/rosters?season=%d", baseURL, season)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var rosters []Roster
	if err := json.Unmarshal(body, &rosters); err != nil {
		return nil, fmt.Errorf("failed to parse rosters: %w", err)
	}

	return rosters, nil
}

// FetchPlayByPlay fetches play-by-play data for a season
func (c *Client) FetchPlayByPlay(ctx context.Context, season int, week int) ([]PlayByPlay, error) {
	url := fmt.Sprintf("%s/pbp?season=%d&week=%d", baseURL, season, week)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var pbp []PlayByPlay
	if err := json.Unmarshal(body, &pbp); err != nil {
		return nil, fmt.Errorf("failed to parse play-by-play: %w", err)
	}

	return pbp, nil
}

// FetchNextGenStats fetches Next Gen Stats data
func (c *Client) FetchNextGenStats(ctx context.Context, season int, statType string) ([]NextGenStats, error) {
	url := fmt.Sprintf("%s/nextgen_stats?season=%d&stat_type=%s", baseURL, season, statType)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var ngs []NextGenStats
	if err := json.Unmarshal(body, &ngs); err != nil {
		return nil, fmt.Errorf("failed to parse Next Gen Stats: %w", err)
	}

	return ngs, nil
}

// FetchDepthCharts fetches team depth charts
func (c *Client) FetchDepthCharts(ctx context.Context, season int) ([]DepthChart, error) {
	url := fmt.Sprintf("%s/depth_charts?season=%d", baseURL, season)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var charts []DepthChart
	if err := json.Unmarshal(body, &charts); err != nil {
		return nil, fmt.Errorf("failed to parse depth charts: %w", err)
	}

	return charts, nil
}

// FetchInjuries fetches injury reports
func (c *Client) FetchInjuries(ctx context.Context, season int, week int) ([]Injury, error) {
	url := fmt.Sprintf("%s/injuries?season=%d&week=%d", baseURL, season, week)
	body, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	var injuries []Injury
	if err := json.Unmarshal(body, &injuries); err != nil {
		return nil, fmt.Errorf("failed to parse injuries: %w", err)
	}

	return injuries, nil
}