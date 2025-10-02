package nflverse

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
)

const (
	// Base URL for NFLverse data releases
	nflverseBaseURL = "https://github.com/nflverse/nflverse-data/releases/download"
)

// CSVParser handles downloading and parsing NFLverse CSV files
type CSVParser struct {
	httpClient *http.Client
}

// NewCSVParser creates a new CSV parser
func NewCSVParser() *CSVParser {
	return &CSVParser{
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Increased timeout for large files
		},
	}
}

// downloadCSV downloads a CSV file from NFLverse
func (p *CSVParser) downloadCSV(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "GridIronMind/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// ParsePlayerStats downloads and parses player stats for a given season
func (p *CSVParser) ParsePlayerStats(ctx context.Context, season int) ([]*PlayerStatCSV, error) {
	url := fmt.Sprintf("%s/player_stats/player_stats_%d.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download player stats: %w", err)
	}

	var stats []*PlayerStatCSV
	if err := gocsv.UnmarshalBytes(csvData, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return stats, nil
}

// ParseSchedule downloads and parses schedule for a given season
func (p *CSVParser) ParseSchedule(ctx context.Context, season int) ([]*ScheduleCSV, error) {
	url := fmt.Sprintf("%s/schedules/sched_%d.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download schedule: %w", err)
	}

	var schedule []*ScheduleCSV
	if err := gocsv.UnmarshalBytes(csvData, &schedule); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return schedule, nil
}

// ParseRosters downloads and parses rosters for a given season
func (p *CSVParser) ParseRosters(ctx context.Context, season int) ([]*RosterCSV, error) {
	url := fmt.Sprintf("%s/rosters/roster_%d.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download rosters: %w", err)
	}

	var rosters []*RosterCSV
	if err := gocsv.UnmarshalBytes(csvData, &rosters); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return rosters, nil
}

// ParseNextGenStatsPassing downloads and parses NGS passing stats
func (p *CSVParser) ParseNextGenStatsPassing(ctx context.Context, season int) ([]*NextGenStatsPassingCSV, error) {
	url := fmt.Sprintf("%s/nextgen_stats/ngs_%d_passing.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download NGS passing stats: %w", err)
	}

	var stats []*NextGenStatsPassingCSV
	if err := gocsv.UnmarshalBytes(csvData, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return stats, nil
}

// ParseNextGenStatsRushing downloads and parses NGS rushing stats
func (p *CSVParser) ParseNextGenStatsRushing(ctx context.Context, season int) ([]*NextGenStatsRushingCSV, error) {
	url := fmt.Sprintf("%s/nextgen_stats/ngs_%d_rushing.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download NGS rushing stats: %w", err)
	}

	var stats []*NextGenStatsRushingCSV
	if err := gocsv.UnmarshalBytes(csvData, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return stats, nil
}

// ParseNextGenStatsReceiving downloads and parses NGS receiving stats
func (p *CSVParser) ParseNextGenStatsReceiving(ctx context.Context, season int) ([]*NextGenStatsReceivingCSV, error) {
	url := fmt.Sprintf("%s/nextgen_stats/ngs_%d_receiving.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download NGS receiving stats: %w", err)
	}

	var stats []*NextGenStatsReceivingCSV
	if err := gocsv.UnmarshalBytes(csvData, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return stats, nil
}

// ParseInjuries downloads and parses injury reports for a given season
func (p *CSVParser) ParseInjuries(ctx context.Context, season int) ([]*InjuryCSV, error) {
	url := fmt.Sprintf("%s/injuries/injuries_%d.csv", nflverseBaseURL, season)

	csvData, err := p.downloadCSV(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to download injuries: %w", err)
	}

	var injuries []*InjuryCSV
	if err := gocsv.UnmarshalBytes(csvData, &injuries); err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	return injuries, nil
}

// FilterRegularSeason filters schedule/stats to only include regular season games
func FilterRegularSeason(schedules []*ScheduleCSV) []*ScheduleCSV {
	var filtered []*ScheduleCSV
	for _, s := range schedules {
		if strings.ToLower(s.GameType) == "reg" || s.GameType == "REG" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// FilterPlayoffs filters schedule/stats to only include playoff games
func FilterPlayoffs(schedules []*ScheduleCSV) []*ScheduleCSV {
	var filtered []*ScheduleCSV
	for _, s := range schedules {
		gameType := strings.ToLower(s.GameType)
		if gameType == "post" || gameType == "playoffs" {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
