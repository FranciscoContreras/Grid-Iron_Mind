package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// DataExplorer explores all available data sources and saves raw responses
type DataExplorer struct {
	httpClient *http.Client
	outputDir  string
}

func NewDataExplorer() *DataExplorer {
	return &DataExplorer{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		outputDir: "./data_exploration",
	}
}

func (e *DataExplorer) saveResponse(filename string, data interface{}) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Pretty print JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	filepath := fmt.Sprintf("%s/%s", e.outputDir, filename)
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("✓ Saved: %s (%d bytes)", filepath, len(jsonData))
	return nil
}

func (e *DataExplorer) fetchJSON(ctx context.Context, url string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GridIronMind-DataExplorer/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// ESPN API Endpoints
func (e *DataExplorer) exploreESPN(ctx context.Context) {
	log.Println("\n=== ESPN API EXPLORATION ===")

	endpoints := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "espn_scoreboard_current",
			url:         "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard",
			description: "Current week scoreboard with live scores",
		},
		{
			name:        "espn_teams",
			url:         "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams",
			description: "All NFL teams with basic info",
		},
		{
			name:        "espn_team_detail_chiefs",
			url:         "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams/12?enable=roster,projection,stats",
			description: "Detailed team info with roster (Chiefs example)",
		},
		{
			name:        "espn_player_mahomes",
			url:         "https://site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/3139477/overview",
			description: "Player overview (Patrick Mahomes example)",
		},
		{
			name:        "espn_player_stats_mahomes",
			url:         "https://site.api.espn.com/apis/site/v2/sports/football/nfl/athletes/3139477/statistics",
			description: "Player career statistics (Mahomes example)",
		},
		{
			name:        "espn_game_detail",
			url:         "https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=401547638",
			description: "Detailed game info with stats (example game)",
		},
		{
			name:        "espn_standings",
			url:         "https://site.api.espn.com/apis/v2/sports/football/nfl/standings",
			description: "NFL standings by division",
		},
		{
			name:        "espn_news",
			url:         "https://site.api.espn.com/apis/site/v2/sports/football/nfl/news",
			description: "Latest NFL news",
		},
	}

	for _, ep := range endpoints {
		log.Printf("\nFetching: %s", ep.description)
		data, err := e.fetchJSON(ctx, ep.url)
		if err != nil {
			log.Printf("✗ Error: %v", err)
			continue
		}

		if err := e.saveResponse(ep.name+".json", data); err != nil {
			log.Printf("✗ Save error: %v", err)
		}

		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

// NFLverse API Endpoints
func (e *DataExplorer) exploreNFLverse(ctx context.Context) {
	log.Println("\n=== NFLVERSE API EXPLORATION ===")

	endpoints := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "nflverse_player_stats_2024",
			url:         "https://nflreadr.nflverse.com/player_stats?season=2024",
			description: "Player stats for 2024 season",
		},
		{
			name:        "nflverse_schedule_2024",
			url:         "https://nflreadr.nflverse.com/schedule?season=2024",
			description: "Schedule for 2024 season",
		},
		{
			name:        "nflverse_nextgen_passing_2024",
			url:         "https://nflreadr.nflverse.com/nextgen_stats?season=2024&stat_type=passing",
			description: "Next Gen Stats - Passing for 2024",
		},
		{
			name:        "nflverse_nextgen_rushing_2024",
			url:         "https://nflreadr.nflverse.com/nextgen_stats?season=2024&stat_type=rushing",
			description: "Next Gen Stats - Rushing for 2024",
		},
		{
			name:        "nflverse_nextgen_receiving_2024",
			url:         "https://nflreadr.nflverse.com/nextgen_stats?season=2024&stat_type=receiving",
			description: "Next Gen Stats - Receiving for 2024",
		},
		{
			name:        "nflverse_rosters_2024",
			url:         "https://nflreadr.nflverse.com/rosters?season=2024",
			description: "Team rosters for 2024",
		},
		{
			name:        "nflverse_draft_picks_2024",
			url:         "https://nflreadr.nflverse.com/draft_picks?season=2024",
			description: "Draft picks for 2024",
		},
	}

	for _, ep := range endpoints {
		log.Printf("\nFetching: %s", ep.description)

		// NFLverse returns arrays directly
		req, err := http.NewRequestWithContext(ctx, "GET", ep.url, nil)
		if err != nil {
			log.Printf("✗ Error: %v", err)
			continue
		}

		req.Header.Set("User-Agent", "GridIronMind-DataExplorer/1.0")
		resp, err := e.httpClient.Do(req)
		if err != nil {
			log.Printf("✗ Error: %v", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Printf("✗ Error reading response: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("✗ HTTP %d: %s", resp.StatusCode, string(body))
			continue
		}

		// Try to parse as array
		var result interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("✗ Parse error: %v", err)
			continue
		}

		if err := e.saveResponse(ep.name+".json", result); err != nil {
			log.Printf("✗ Save error: %v", err)
		}

		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

// WeatherAPI Endpoints
func (e *DataExplorer) exploreWeatherAPI(ctx context.Context, apiKey string) {
	log.Println("\n=== WEATHER API EXPLORATION ===")

	if apiKey == "" {
		log.Println("⚠ WEATHER_API_KEY not set, skipping weather exploration")
		return
	}

	endpoints := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "weather_current_kc",
			url:         fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=Kansas City,MO&aqi=no", apiKey),
			description: "Current weather in Kansas City",
		},
		{
			name:        "weather_forecast_kc",
			url:         fmt.Sprintf("https://api.weatherapi.com/v1/forecast.json?key=%s&q=Kansas City,MO&days=3", apiKey),
			description: "3-day forecast for Kansas City",
		},
		{
			name:        "weather_historical_kc",
			url:         fmt.Sprintf("https://api.weatherapi.com/v1/history.json?key=%s&q=Kansas City,MO&dt=2024-09-15", apiKey),
			description: "Historical weather for Kansas City (example date)",
		},
		{
			name:        "weather_current_coords",
			url:         fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=39.0997,-94.5786&aqi=no", apiKey),
			description: "Current weather by coordinates (Arrowhead Stadium)",
		},
	}

	for _, ep := range endpoints {
		log.Printf("\nFetching: %s", ep.description)
		data, err := e.fetchJSON(ctx, ep.url)
		if err != nil {
			log.Printf("✗ Error: %v", err)
			continue
		}

		if err := e.saveResponse(ep.name+".json", data); err != nil {
			log.Printf("✗ Save error: %v", err)
		}

		time.Sleep(500 * time.Millisecond) // Rate limiting
	}
}

// Generate Analysis Report
func (e *DataExplorer) generateReport() {
	log.Println("\n=== GENERATING ANALYSIS REPORT ===")

	report := `# Data Source Exploration Report

Generated: %s

## Purpose
This report analyzes all available data sources to understand:
1. Data structure and format
2. Available fields and their types
3. Relationships between entities
4. Data quality and completeness
5. Optimal storage strategy

## Data Sources Analyzed

### 1. ESPN API
- **Scoreboard**: Live game data with scores, status, teams
- **Teams**: Team information, rosters, statistics
- **Players**: Player profiles, career stats, biographical data
- **Games**: Detailed game summaries with play-by-play
- **Standings**: Division and conference standings
- **News**: Latest NFL news and updates

### 2. NFLverse API
- **Player Stats**: Comprehensive season statistics
- **Schedule**: Game schedules with dates and matchups
- **Next Gen Stats**: Advanced analytics (passing, rushing, receiving)
- **Rosters**: Team rosters with player details
- **Draft Picks**: Draft history and picks

### 3. WeatherAPI.com
- **Current Weather**: Real-time conditions
- **Forecasts**: Multi-day weather predictions
- **Historical**: Past weather data for game enrichment
- **Location Support**: City names and coordinates

## Key Findings

### Data Structure Observations
Check the JSON files in ./data_exploration/ to analyze:

1. **Nested Objects**: How deeply nested is the data?
2. **Arrays**: What entities are returned as arrays?
3. **Data Types**: String vs Int, nullable fields
4. **Identifiers**: What unique IDs are available?
5. **Relationships**: How are entities linked?

### Recommended Database Schema Changes

Based on the actual data structure, consider:

1. **Normalization**: What should be separate tables?
2. **Indexing**: Which fields are commonly queried?
3. **JSON Columns**: Should some nested data be stored as JSONB?
4. **Caching Strategy**: What data changes frequently vs rarely?
5. **Aggregations**: What pre-computed stats would be useful?

### Next Steps

1. Review all JSON files in ./data_exploration/
2. Document field mappings (API → Database)
3. Identify missing fields in current schema
4. Plan schema migrations
5. Update ingestion logic to capture all relevant data

## Files Generated
`

	// List all files in output directory
	files, err := os.ReadDir(e.outputDir)
	if err != nil {
		log.Printf("Error reading output directory: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			info, _ := file.Info()
			report += fmt.Sprintf("- %s (%.2f KB)\n", file.Name(), float64(info.Size())/1024)
		}
	}

	report += fmt.Sprintf("\n\nTotal files: %d\n", len(files))
	report += "\n## Analysis Instructions\n\n"
	report += "1. Open each JSON file to understand the data structure\n"
	report += "2. Compare against current database schema (schema.sql)\n"
	report += "3. Identify gaps and opportunities\n"
	report += "4. Document findings in data_analysis.md\n"
	report += "5. Propose schema improvements\n"

	reportPath := fmt.Sprintf("%s/README.md", e.outputDir)
	if err := os.WriteFile(reportPath, []byte(fmt.Sprintf(report, time.Now().Format(time.RFC3339))), 0644); err != nil {
		log.Printf("✗ Failed to write report: %v", err)
		return
	}

	log.Printf("✓ Analysis report saved: %s", reportPath)
}

func main() {
	log.Println("Grid Iron Mind - Data Source Explorer")
	log.Println("=====================================")

	ctx := context.Background()
	explorer := NewDataExplorer()

	// Explore all data sources
	explorer.exploreESPN(ctx)
	explorer.exploreNFLverse(ctx)
	explorer.exploreWeatherAPI(ctx, os.Getenv("WEATHER_API_KEY"))

	// Generate analysis report
	explorer.generateReport()

	log.Println("\n✓ Exploration complete! Check ./data_exploration/ for results")
	log.Println("📊 Review the files to understand data structure and plan schema improvements")
}