# Grid Iron Mind - Comprehensive Data Analysis

**Generated:** 2025-09-30
**Purpose:** Analyze API data structures to optimize database schema and storage strategy

---

## Executive Summary

This analysis examines data from:
- ✅ **ESPN API**: 7 endpoints successfully fetched
- ⚠️ **NFLverse API**: API structure changed, need alternative access
- ✅ **WeatherAPI.com**: 1 sample successfully fetched

**Key Findings:**
1. ESPN provides rich nested data that we're currently underutilizing
2. Current schema is missing many valuable fields
3. Need better relationship modeling for stats and historical data
4. Weather integration working but limited data captured

---

## 1. ESPN API Analysis

### 1.1 Scoreboard Data (`espn_scoreboard_current.json`)

**Current Structure We Capture:**
```
games table:
- id, nfl_game_id, season, week, game_date
- home_team_id, away_team_id
- home_score, away_score
- status
```

**Missing Fields We Should Capture:**

#### Competition Level
- `competition.attendance` → games.attendance
- `competition.venue.id` → games.venue_id
- `competition.venue.fullName` → games.venue_name
- `competition.venue.address.city` → games.venue_city
- `competition.venue.address.state` → games.venue_state
- `competition.broadcasters` → new table: game_broadcasts

#### Status Details
- `status.type.id` → more detailed status tracking
- `status.type.description` → status_description
- `status.period` → current_period
- `status.displayClock` → game_clock

#### Team Details (per competition)
- `competitor.records` → team season records
- `competitor.statistics` → game-level team stats
- `competitor.leaders` → top performers per game

### 1.2 Teams Data (`espn_teams.json`)

**Rich Data Available:**
```json
{
  "id": "12",
  "uid": "s:20~l:28~t:12",
  "slug": "kansas-city-chiefs",
  "location": "Kansas City",
  "name": "Chiefs",
  "abbreviation": "KC",
  "displayName": "Kansas City Chiefs",
  "shortDisplayName": "Chiefs",
  "color": "e31837",
  "alternateColor": "ffb612",
  "isActive": true,
  "logos": [...],
  "links": [...],
  "record": {...},
  "standingSummary": "..."
}
```

**Recommendations:**
1. Add `teams.uid` for ESPN's universal ID
2. Add `teams.slug` for URL-friendly names
3. Add `teams.alternate_color`
4. Add `teams.logo_url` (array or JSON)
5. Add `teams.is_active` for historical teams

### 1.3 Team Detail with Roster (`espn_team_detail_chiefs.json`)

**Roster Data Structure:**
```json
{
  "athlete": {
    "id": "3139477",
    "fullName": "Patrick Mahomes",
    "displayName": "Patrick Mahomes",
    "shortName": "P. Mahomes",
    "jersey": "15",
    "position": {
      "name": "Quarterback",
      "abbreviation": "QB"
    },
    "headshot": "...",
    "age": 28,
    "experience": {...},
    "status": {...}
  }
}
```

**Current schema captures:**
- name, position, jersey_number, height, weight

**Missing:**
- Short names/display names
- Headshot URLs
- Position details (not just abbreviation)
- Age/experience tracking
- Status details

### 1.4 Player Overview (`espn_player_mahomes.json`)

**Incredibly Rich Data:**
- Career stats summary
- Latest news
- Recent games
- Team history
- Awards and achievements
- Bio information (birthplace, college)
- Social media links

**New Tables Needed:**
```sql
-- Player career summary by season
CREATE TABLE player_season_stats (
    id UUID PRIMARY KEY,
    player_id UUID REFERENCES players(id),
    season INT NOT NULL,
    team_id UUID REFERENCES teams(id),
    games_played INT,
    games_started INT,
    -- Position-specific stats
    passing_yards INT,
    passing_tds INT,
    interceptions INT,
    -- ... many more fields
    UNIQUE(player_id, season)
);

-- Player news/articles
CREATE TABLE player_news (
    id UUID PRIMARY KEY,
    player_id UUID REFERENCES players(id),
    headline TEXT,
    description TEXT,
    published TIMESTAMP,
    link TEXT,
    source TEXT
);

-- Player team history
CREATE TABLE player_team_history (
    id UUID PRIMARY KEY,
    player_id UUID REFERENCES players(id),
    team_id UUID REFERENCES teams(id),
    start_season INT,
    end_season INT,
    UNIQUE(player_id, team_id, start_season)
);
```

### 1.5 Game Detail (`espn_game_detail.json`)

**Box Score Data:**
- Team statistics (first downs, total yards, turnovers, possession time)
- Scoring summary by quarter
- Drive summaries
- Play-by-play data (optional, very large)

**Recommended Tables:**

```sql
-- Team stats per game
CREATE TABLE game_team_stats (
    id UUID PRIMARY KEY,
    game_id UUID REFERENCES games(id),
    team_id UUID REFERENCES teams(id),
    first_downs INT,
    total_yards INT,
    passing_yards INT,
    rushing_yards INT,
    turnovers INT,
    penalties INT,
    penalty_yards INT,
    possession_time VARCHAR(10), -- "MM:SS"
    third_down_efficiency VARCHAR(10), -- "5-12"
    fourth_down_efficiency VARCHAR(10),
    UNIQUE(game_id, team_id)
);

-- Scoring plays
CREATE TABLE game_scoring_plays (
    id UUID PRIMARY KEY,
    game_id UUID REFERENCES games(id),
    team_id UUID REFERENCES teams(id),
    quarter INT,
    time_remaining VARCHAR(10),
    play_type VARCHAR(50), -- TD, FG, Safety, etc.
    description TEXT,
    home_score INT, -- score after this play
    away_score INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 1.6 Standings (`espn_standings.json`)

**Contains:**
- Division standings
- Conference standings
- Win/loss records
- Home/away records
- Division records
- Playoff clinching status

**Recommended:**
```sql
-- Could be computed view OR cached table
CREATE TABLE team_standings (
    id UUID PRIMARY KEY,
    team_id UUID REFERENCES teams(id),
    season INT,
    week INT, -- or NULL for season total
    wins INT,
    losses INT,
    ties INT,
    win_pct DECIMAL(5,3),
    points_for INT,
    points_against INT,
    point_differential INT,
    home_record VARCHAR(10),
    away_record VARCHAR(10),
    division_record VARCHAR(10),
    conference_record VARCHAR(10),
    streak VARCHAR(10), -- "W3", "L2", etc.
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, season, week)
);
```

---

## 2. Weather API Analysis (`weather_current_coords.json`)

**Current Weather Response:**
```json
{
  "location": {
    "name": "Kansas City",
    "region": "Missouri",
    "country": "United States of America",
    "lat": 39.1,
    "lon": -94.58,
    "tz_id": "America/Chicago",
    "localtime": "2025-09-30 8:21"
  },
  "current": {
    "temp_c": 22.2,
    "temp_f": 72.0,
    "is_day": 1,
    "condition": {
      "text": "Partly cloudy",
      "icon": "//cdn.weatherapi.com/weather/64x64/day/116.png",
      "code": 1003
    },
    "wind_mph": 5.6,
    "wind_degree": 170,
    "wind_dir": "S",
    "pressure_mb": 1017.0,
    "precip_in": 0.0,
    "humidity": 59,
    "cloud": 25,
    "feelslike_f": 72.0,
    "vis_miles": 9.0,
    "uv": 7.0,
    "gust_mph": 6.9
  }
}
```

**Currently Storing:**
- weather_temp (F)
- weather_condition (text)
- weather_wind_speed (mph)
- weather_humidity (%)

**Missing Valuable Fields:**
- `wind_direction` - "S", "NW", etc.
- `pressure_mb` - barometric pressure
- `visibility_miles` - visibility
- `is_day` - day/night game
- `feels_like_f` - perceived temperature
- `cloud_coverage` - percentage
- `precipitation` - rain/snow amount
- `uv_index` - sun intensity

**Recommendation:**
Expand weather fields or create separate `game_weather_detail` table:

```sql
ALTER TABLE games ADD COLUMN weather_wind_dir VARCHAR(10);
ALTER TABLE games ADD COLUMN weather_pressure INT;
ALTER TABLE games ADD COLUMN weather_visibility INT;
ALTER TABLE games ADD COLUMN weather_feels_like INT;
ALTER TABLE games ADD COLUMN weather_precipitation DECIMAL(4,2);
ALTER TABLE games ADD COLUMN weather_cloud_cover INT;
ALTER TABLE games ADD COLUMN weather_uv_index DECIMAL(3,1);
ALTER TABLE games ADD COLUMN is_day_game BOOLEAN;
```

---

## 3. Current Schema Gaps

### 3.1 Missing Tables

1. **player_season_stats** - Career stats by season
2. **game_team_stats** - Team performance per game
3. **game_scoring_plays** - Scoring timeline
4. **team_standings** - Weekly/season standings
5. **player_news** - News articles
6. **game_broadcasts** - TV/streaming info
7. **advanced_stats** - Next Gen Stats (when we get NFLverse access)

### 3.2 Missing Columns

**teams table:**
- uid, slug, alternate_color, logo_urls (JSONB)
- stadium_name, stadium_capacity, stadium_type, stadium_surface
- stadium_latitude, stadium_longitude (already in migrations!)

**players table:**
- short_name, display_name
- headshot_url, birth_city, birth_state, birth_country
- college, draft_year, draft_round, draft_pick
- rookie_year, years_pro
- status (active, injured, retired)

**games table:**
- venue_id, venue_name, venue_city, venue_state, venue_type
- attendance
- current_period, game_clock, status_description
- weather fields (expanded as above)
- game_time_et, playoff_round

---

## 4. Recommended Schema Changes

### 4.1 Immediate Additions (High Value, Low Effort)

```sql
-- Add missing venue/weather fields that migrations already support
-- (These columns exist but aren't being populated)

-- Populate during game sync:
UPDATE games SET
    venue_id = competition.venue.id,
    venue_name = competition.venue.fullName,
    venue_city = competition.venue.address.city,
    venue_state = competition.venue.address.state,
    attendance = competition.attendance;

-- Expand weather during enrichment:
ALTER TABLE games ADD COLUMN weather_wind_dir VARCHAR(10);
ALTER TABLE games ADD COLUMN weather_feels_like INT;
ALTER TABLE games ADD COLUMN weather_visibility INT;
```

### 4.2 Medium Priority (New Tables)

```sql
-- Team stats per game
CREATE TABLE game_team_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id),
    first_downs INT DEFAULT 0,
    total_yards INT DEFAULT 0,
    passing_yards INT DEFAULT 0,
    rushing_yards INT DEFAULT 0,
    turnovers INT DEFAULT 0,
    penalties INT DEFAULT 0,
    penalty_yards INT DEFAULT 0,
    possession_time VARCHAR(10),
    third_down_att INT DEFAULT 0,
    third_down_conv INT DEFAULT 0,
    fourth_down_att INT DEFAULT 0,
    fourth_down_conv INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, team_id)
);

CREATE INDEX idx_game_team_stats_game ON game_team_stats(game_id);
CREATE INDEX idx_game_team_stats_team ON game_team_stats(team_id);
```

### 4.3 Long Term (Historical Stats)

```sql
-- Player career stats by season
CREATE TABLE player_season_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    season INT NOT NULL,
    team_id UUID REFERENCES teams(id),
    position VARCHAR(10),
    games_played INT DEFAULT 0,
    games_started INT DEFAULT 0,

    -- Passing (QB)
    passing_attempts INT DEFAULT 0,
    passing_completions INT DEFAULT 0,
    passing_yards INT DEFAULT 0,
    passing_tds INT DEFAULT 0,
    interceptions INT DEFAULT 0,
    sacks INT DEFAULT 0,
    qb_rating DECIMAL(5,2),

    -- Rushing (RB, QB)
    rushing_attempts INT DEFAULT 0,
    rushing_yards INT DEFAULT 0,
    rushing_tds INT DEFAULT 0,
    rushing_avg DECIMAL(4,2),
    rushing_long INT DEFAULT 0,

    -- Receiving (WR, TE, RB)
    receptions INT DEFAULT 0,
    receiving_yards INT DEFAULT 0,
    receiving_tds INT DEFAULT 0,
    receiving_avg DECIMAL(4,2),
    receiving_long INT DEFAULT 0,
    targets INT DEFAULT 0,

    -- Defense
    tackles INT DEFAULT 0,
    tackles_solo INT DEFAULT 0,
    sacks_defense DECIMAL(4,1) DEFAULT 0,
    interceptions_defense INT DEFAULT 0,
    passes_defended INT DEFAULT 0,
    forced_fumbles INT DEFAULT 0,
    fumble_recoveries INT DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season)
);

CREATE INDEX idx_player_season_stats_player ON player_season_stats(player_id);
CREATE INDEX idx_player_season_stats_season ON player_season_stats(season);
CREATE INDEX idx_player_season_stats_team ON player_season_stats(team_id);
```

---

## 5. Data Ingestion Updates Needed

### 5.1 Game Sync Enhancement

**File:** `internal/ingestion/service.go`
**Function:** `upsertGame()`

**Add:**
```go
// Extract venue information
venueID := competition.Venue.ID
venueName := competition.Venue.FullName
venueCity := competition.Venue.Address.City
venueState := competition.Venue.Address.State
attendance := competition.Attendance

// Update INSERT to include venue fields
_, err = s.dbPool.Exec(ctx,
    `INSERT INTO games (
        id, nfl_game_id, home_team_id, away_team_id,
        game_date, season, week, home_score, away_score, status,
        venue_id, venue_name, venue_city, venue_state, attendance
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
    id, event.ID, homeTeamID, awayTeamID,
    event.Date.Time, event.Season.Year, event.Week.Number,
    homeScore, awayScore, status,
    venueID, venueName, venueCity, venueState, attendance,
)
```

### 5.2 Weather Enrichment Enhancement

**File:** `internal/ingestion/service.go`
**Function:** `EnrichGamesWithWeather()`

**Add all weather fields:**
```go
updateQuery := `
    UPDATE games
    SET weather_temp = $1,
        weather_condition = $2,
        weather_wind_speed = $3,
        weather_humidity = $4,
        weather_wind_dir = $5,
        weather_feels_like = $6,
        weather_visibility = $7,
        weather_precipitation = $8
    WHERE id = $9
`
s.dbPool.Exec(ctx, updateQuery,
    int(weatherData.Day.AvgTempF),
    weatherData.Day.Condition.Text,
    int(weatherData.Day.MaxWindMPH),
    int(weatherData.Day.AvgHumidity),
    // Add new fields:
    weatherData.Current.WindDir,
    int(weatherData.Current.FeelsLikeF),
    int(weatherData.Current.VisMiles),
    weatherData.Day.TotalPrecipIn,
    gameID,
)
```

### 5.3 New: Team Stats Sync

**Create new function:**
```go
func (s *Service) SyncGameTeamStats(ctx context.Context, gameID string) error {
    // Fetch game detail from ESPN
    gameDetail, err := s.espnClient.FetchGameDetails(ctx, gameID)

    // Extract team stats from box score
    for _, team := range gameDetail.BoxScore.Teams {
        // Insert/update game_team_stats
    }
}
```

---

## 6. API Response Size Analysis

| Endpoint | Size | Key Data | Frequency |
|----------|------|----------|-----------|
| Scoreboard | 665KB | Live scores, 16 games | Every 5min during games |
| Teams | 253KB | All 32 teams | Once per day |
| Team Detail | 652KB | Roster + stats | Once per day per team |
| Player Overview | 385KB | Career summary | On demand |
| Game Detail | 981KB | Box score, plays | After game completion |
| Standings | 430KB | All divisions | Once per day |

**Total for full sync:** ~50MB for current week data
**Recommendation:** Cache heavily, sync strategically

---

## 7. Query Optimization Recommendations

### 7.1 Missing Indexes

```sql
-- Improve game queries by venue
CREATE INDEX idx_games_venue_city ON games(venue_city) WHERE venue_city IS NOT NULL;

-- Weather analysis queries
CREATE INDEX idx_games_weather ON games(weather_temp, weather_condition)
    WHERE weather_temp IS NOT NULL;

-- Playoff games
CREATE INDEX idx_games_playoff ON games(playoff_round)
    WHERE playoff_round IS NOT NULL;

-- Player position queries
CREATE INDEX idx_players_position ON players(position);

-- Historical season queries
CREATE INDEX idx_games_season ON games(season, week);
```

### 7.2 Materialized Views (Future)

```sql
-- Pre-compute team season stats
CREATE MATERIALIZED VIEW team_season_stats AS
SELECT
    t.id as team_id,
    g.season,
    COUNT(*) as games_played,
    SUM(CASE WHEN g.home_team_id = t.id AND g.home_score > g.away_score THEN 1
             WHEN g.away_team_id = t.id AND g.away_score > g.home_score THEN 1
             ELSE 0 END) as wins,
    -- ... more aggregations
FROM teams t
JOIN games g ON g.home_team_id = t.id OR g.away_team_id = t.id
WHERE g.status = 'completed'
GROUP BY t.id, g.season;

REFRESH MATERIALIZED VIEW team_season_stats;
```

---

## 8. Action Items

### Immediate (This Week)
- [ ] Add venue fields to game sync
- [ ] Expand weather enrichment with all fields
- [ ] Update API documentation with new fields

### Short Term (Next Sprint)
- [ ] Create game_team_stats table and sync logic
- [ ] Add missing indexes
- [ ] Implement team standings calculation

### Medium Term (Next Month)
- [ ] Create player_season_stats table
- [ ] Build historical data backfill script
- [ ] Add NFLverse integration for advanced stats

### Long Term (Future)
- [ ] Play-by-play data storage (optional, huge volume)
- [ ] Real-time game state tracking
- [ ] Predictive analytics tables

---

## 9. Storage Estimates

### Current Data Volume
- Teams: 32 rows (~50KB)
- Players: 1,696 active players (~2MB)
- Games: ~285 games/season (~5MB/season)
- Stats: ~50K stat records/season (~50MB)

### With Proposed Changes
- game_team_stats: ~570 rows/season (~1MB)
- player_season_stats: ~1,696 rows/season (~20MB)
- Historical (10 years): ~500MB total

**Conclusion:** Storage is not a concern. Optimize for query performance and data richness.

---

## 10. Next Steps

1. **Review this analysis** with team
2. **Prioritize features** based on business value
3. **Create migration scripts** for approved schema changes
4. **Update ingestion logic** incrementally
5. **Test with sample data** before full deployment
6. **Monitor performance** after each change

---

*This analysis is based on actual API responses captured on 2025-09-30.*
*All JSON samples are available in the `data_exploration/` directory.*