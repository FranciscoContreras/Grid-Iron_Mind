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
Files in ./data_exploration:
- espn_game_detail.json (981K)
- espn_player_mahomes.json (385K)
- espn_player_stats_mahomes.json (204B)
- espn_scoreboard_current.json (665K)
- espn_standings.json (430K)
- espn_team_detail_chiefs.json (652K)
- espn_teams.json (253K)
- nflverse_nextgen_passing_2024.json (42B)
- nflverse_nextgen_receiving_2024.json (42B)
- nflverse_nextgen_rushing_2024.json (42B)
- nflverse_player_stats_2024.json (42B)
- nflverse_rosters_2024.json (42B)
- nflverse_schedule_2024.json (42B)
- weather_current_coords.json (1.3K)
- weather_current_kc.json (42B)
- weather_forecast_kc.json (42B)
- weather_historical_kc.json (42B)
