#!/bin/bash

# Grid Iron Mind - Data Source Explorer
# Fetches sample data from all available APIs to understand structure

WEATHER_API_KEY="65371fc2086c49efbf1123516253009"
OUTPUT_DIR="./data_exploration"

echo "Grid Iron Mind - Data Source Explorer"
echo "====================================="
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to fetch and save JSON
fetch_data() {
    local name=$1
    local url=$2
    local description=$3

    echo "Fetching: $description"
    curl -s -H "User-Agent: GridIronMind-Explorer/1.0" "$url" | python3 -m json.tool > "$OUTPUT_DIR/${name}.json" 2>&1

    if [ $? -eq 0 ]; then
        local size=$(ls -lh "$OUTPUT_DIR/${name}.json" | awk '{print $5}')
        echo "✓ Saved: ${name}.json ($size)"
    else
        echo "✗ Failed to fetch $name"
    fi

    sleep 0.5  # Rate limiting
}

echo ""
echo "=== ESPN API EXPLORATION ==="
echo ""

fetch_data "espn_scoreboard_current" \
    "https://site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard" \
    "Current week scoreboard with live scores"

fetch_data "espn_teams" \
    "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams" \
    "All NFL teams with basic info"

fetch_data "espn_team_detail_chiefs" \
    "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams/12?enable=roster,projection,stats" \
    "Detailed team info with roster (Chiefs)"

fetch_data "espn_player_mahomes" \
    "https://site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/3139477/overview" \
    "Player overview (Patrick Mahomes)"

fetch_data "espn_player_stats_mahomes" \
    "https://site.api.espn.com/apis/site/v2/sports/football/nfl/athletes/3139477/statistics" \
    "Player career statistics (Mahomes)"

fetch_data "espn_game_detail" \
    "https://site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event=401547638" \
    "Detailed game info with box score"

fetch_data "espn_standings" \
    "https://site.api.espn.com/apis/v2/sports/football/nfl/standings" \
    "NFL standings by division"

echo ""
echo "=== NFLVERSE API EXPLORATION ==="
echo ""

fetch_data "nflverse_player_stats_2024" \
    "https://nflreadr.nflverse.com/player_stats?season=2024&limit=100" \
    "Player stats for 2024 season (sample 100)"

fetch_data "nflverse_schedule_2024" \
    "https://nflreadr.nflverse.com/schedule?season=2024" \
    "Schedule for 2024 season"

fetch_data "nflverse_nextgen_passing_2024" \
    "https://nflreadr.nflverse.com/nextgen_stats?season=2024&stat_type=passing&limit=50" \
    "Next Gen Stats - Passing 2024 (sample 50)"

fetch_data "nflverse_nextgen_rushing_2024" \
    "https://nflreadr.nflverse.com/nextgen_stats?season=2024&stat_type=rushing&limit=50" \
    "Next Gen Stats - Rushing 2024 (sample 50)"

fetch_data "nflverse_nextgen_receiving_2024" \
    "https://nflreadr.nflverse.com/nextgen_stats?season=2024&stat_type=receiving&limit=50" \
    "Next Gen Stats - Receiving 2024 (sample 50)"

fetch_data "nflverse_rosters_2024" \
    "https://nflreadr.nflverse.com/rosters?season=2024&limit=100" \
    "Team rosters for 2024 (sample 100)"

echo ""
echo "=== WEATHER API EXPLORATION ==="
echo ""

fetch_data "weather_current_kc" \
    "https://api.weatherapi.com/v1/current.json?key=${WEATHER_API_KEY}&q=Kansas City,MO&aqi=no" \
    "Current weather in Kansas City"

fetch_data "weather_forecast_kc" \
    "https://api.weatherapi.com/v1/forecast.json?key=${WEATHER_API_KEY}&q=Kansas City,MO&days=3" \
    "3-day forecast for Kansas City"

fetch_data "weather_historical_kc" \
    "https://api.weatherapi.com/v1/history.json?key=${WEATHER_API_KEY}&q=Kansas City,MO&dt=2024-09-15" \
    "Historical weather for Kansas City"

fetch_data "weather_current_coords" \
    "https://api.weatherapi.com/v1/current.json?key=${WEATHER_API_KEY}&q=39.0997,-94.5786&aqi=no" \
    "Current weather by coordinates (Arrowhead)"

echo ""
echo "=== GENERATING ANALYSIS REPORT ==="
echo ""

# Generate comprehensive README
cat > "$OUTPUT_DIR/README.md" << 'EOF'
# Data Source Exploration Report

## Purpose
This report analyzes all available data sources to understand:
1. Data structure and format
2. Available fields and their types
3. Relationships between entities
4. Data quality and completeness
5. Optimal storage strategy

## Data Sources Analyzed

### 1. ESPN API
Files: `espn_*.json`

**Key Endpoints:**
- **Scoreboard** (`espn_scoreboard_current.json`): Live game data with scores, status, teams, competitions
- **Teams** (`espn_teams.json`): All 32 teams with basic info
- **Team Detail** (`espn_team_detail_chiefs.json`): Roster, stats, projections for specific team
- **Player Overview** (`espn_player_mahomes.json`): Player bio, stats summary, recent games
- **Player Stats** (`espn_player_stats_mahomes.json`): Career statistics by season
- **Game Detail** (`espn_game_detail.json`): Box score, play-by-play, team stats
- **Standings** (`espn_standings.json`): Division/conference standings

**Key Observations:**
- Rich nested structure with competitions, competitors, teams
- Multiple identifier formats (id, uid, guid)
- Status objects with detailed game state
- Venue information included in games
- Statistics as nested arrays

### 2. NFLverse API
Files: `nflverse_*.json`

**Key Endpoints:**
- **Player Stats** (`nflverse_player_stats_2024.json`): Comprehensive weekly/season stats
- **Schedule** (`nflverse_schedule_2024.json`): Game schedule with stadium, weather placeholders
- **Next Gen Stats**: Advanced metrics (passing, rushing, receiving)
- **Rosters** (`nflverse_rosters_2024.json`): Player roster data with positions, depth charts

**Key Observations:**
- Flat structure (arrays of objects)
- Consistent column names across datasets
- Historical data available
- Advanced metrics (Next Gen Stats) provide deeper insights
- Stadium coordinates included in schedule

### 3. WeatherAPI.com
Files: `weather_*.json`

**Key Endpoints:**
- **Current** (`weather_current_*.json`): Real-time weather conditions
- **Forecast** (`weather_forecast_*.json`): Multi-day predictions
- **Historical** (`weather_historical_*.json`): Past weather data

**Key Observations:**
- Detailed weather metrics (temp, wind, humidity, pressure)
- Location data (lat/lon, timezone)
- Multiple temperature units (F/C)
- Condition descriptions and codes
- UV index, visibility, precipitation

## Analysis Tasks

### 1. Schema Comparison
Compare each JSON structure against current `schema.sql`:
- [ ] Which fields are we missing?
- [ ] Which fields are we storing that aren't needed?
- [ ] What data types need adjustment?
- [ ] What indexes would improve query performance?

### 2. Data Relationships
Map out entity relationships:
- [ ] Team → Players (one-to-many)
- [ ] Game → Teams (many-to-many via home/away)
- [ ] Game → Stats (one-to-many)
- [ ] Player → Stats (one-to-many)
- [ ] Game → Weather (one-to-one)

### 3. Field Mapping
Create mapping table: API Field → Database Column
Example:
```
ESPN event.id → games.nfl_game_id
ESPN team.id → teams.nfl_id
NFLverse player_id → players.nfl_id
```

### 4. Storage Strategy

**Current Schema Issues to Address:**
1. Missing fields from ESPN (venue details, attendance, playoff info)
2. Missing advanced stats from NFLverse (Next Gen Stats)
3. Weather data structure (currently simple, could be richer)
4. Player biographical data (birth place, college, etc.)
5. Historical data (career stats by season)

**Recommended Additions:**
1. **player_season_stats** table for career history
2. **advanced_stats** table for Next Gen Stats
3. **game_weather** expanded fields
4. **team_stadiums** separate table
5. **game_box_scores** for detailed team stats

### 5. Caching Strategy
Determine what should be cached and for how long:
- Teams: Rarely change (24h cache)
- Players: Change with roster moves (4h cache)
- Games: Live during games (no cache), historical (24h cache)
- Stats: Update after games (1h cache during games)
- Weather: Current (15min), historical (permanent)

## Next Steps

1. **Review JSON Files**: Open each file and examine structure
2. **Document Missing Fields**: Create list of fields we should capture
3. **Schema Migration Plan**: Write SQL migrations for new tables/columns
4. **Update Ingestion Logic**: Modify sync functions to capture new data
5. **API Documentation**: Update API docs with new endpoints/fields

## Files Generated
EOF

# List all generated files
echo "Files in $OUTPUT_DIR:" >> "$OUTPUT_DIR/README.md"
ls -lh "$OUTPUT_DIR"/*.json 2>/dev/null | awk '{printf "- %s (%s)\n", $9, $5}' | sed "s|$OUTPUT_DIR/||" >> "$OUTPUT_DIR/README.md"

echo "✓ Analysis report saved: $OUTPUT_DIR/README.md"

echo ""
echo "=== SUMMARY ==="
echo ""
echo "✓ Exploration complete!"
echo "📊 Check ./$OUTPUT_DIR/ for all JSON files"
echo "📝 Read ./$OUTPUT_DIR/README.md for analysis guide"
echo ""
echo "Total files generated:"
ls -1 "$OUTPUT_DIR"/*.json 2>/dev/null | wc -l

echo ""
echo "Next steps:"
echo "1. Review each JSON file to understand data structure"
echo "2. Compare against current schema.sql"
echo "3. Identify missing fields and opportunities"
echo "4. Plan schema improvements and migrations"