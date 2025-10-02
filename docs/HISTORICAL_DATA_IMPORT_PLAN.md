# Historical Data Import Plan (2010-2024)

## Executive Summary
Import 15 years of comprehensive NFL historical data to improve API performance and enable advanced analytics. This will populate our database with ~200k+ player stats, ~4k games, and historical records.

## Current State
- **Players**: 2,515 active players
- **Teams**: 32 teams
- **Games**: 272 games (2025 season only)
- **Game Stats**: 0 records
- **Coverage**: 2025 season only
- **Gap**: Missing 2010-2024 (15 years of data)

## Data Sources

### 1. NFLverse (Primary Source)
**URL**: `https://github.com/nflverse/nflverse-data/releases/download`

**Available Datasets** (CSV format):
- `player_stats/player_stats_[YEAR].csv` - Weekly player statistics
- `schedules/sched_[YEAR].csv` - Game schedules with scores
- `rosters/roster_[YEAR].csv` - Team rosters by season
- `nextgen_stats/ngs_[YEAR]_[type].csv` - Next Gen Stats (passing, rushing, receiving)
- `pbp/play_by_play_[YEAR].csv` - Play-by-play data (large files)
- `injuries/injuries_[YEAR].csv` - Injury reports

**Data Quality**: High - Community-maintained, validated against official NFL data

**Coverage**: 1999-present (we'll use 2010-2024)

### 2. ESPN API (Supplementary)
- Current team information
- Player headshots and metadata
- Real-time updates

## Import Strategy

### Phase 1: Foundation (Immediate)
**Goal**: Set up CSV parsing infrastructure
**Duration**: Day 1

1. **Add CSV parsing library**
   ```bash
   go get github.com/gocarina/gocsv
   ```

2. **Create NFLverse CSV parsers**
   - Parse player stats CSV
   - Parse schedule CSV
   - Parse roster CSV
   - Parse Next Gen Stats CSV

3. **Build data transformation layer**
   - Map NFLverse fields to our schema
   - Handle missing/null values
   - Validate data integrity

### Phase 2: Incremental Historical Import (Days 2-3)
**Goal**: Import data year-by-year with validation
**Duration**: 2 days

**Import Order** (oldest to newest):
1. **2010** → 2011 → 2012 → 2013 → 2014
2. **2015** → 2016 → 2017 → 2018 → 2019
3. **2020** → 2021 → 2022 → 2023 → 2024

**Per-Year Process**:
```
For each year:
  1. Download roster CSV → Insert/update players
  2. Download schedule CSV → Insert games
  3. Download player stats CSV → Insert game_stats
  4. Download Next Gen Stats → Insert advanced_stats
  5. Validate counts and relationships
  6. Log progress and errors
```

**Why Year-by-Year**:
- Easier to debug issues
- Can resume if interrupted
- Track progress granularly
- Validate data quality incrementally

### Phase 3: Advanced Data Import (Days 4-5)
**Goal**: Enrich with advanced analytics
**Duration**: 2 days

1. **Next Gen Stats Import**
   - Passing stats (2016-2024)
   - Rushing stats (2018-2024)
   - Receiving stats (2017-2024)

2. **Play-by-Play Data** (Optional - very large)
   - Only import summary data
   - Skip individual play details for now
   - Consider separate archive storage

3. **Historical Injury Reports** (if available)

### Phase 4: Validation & Optimization (Day 6)
**Goal**: Ensure data quality and performance
**Duration**: 1 day

1. **Data Validation**
   - Check foreign key relationships
   - Verify stat totals match known records
   - Identify and fix data gaps

2. **Database Optimization**
   - ANALYZE tables for query planner
   - Verify all indexes are used
   - Check query performance

3. **Cache Warming**
   - Pre-cache popular queries
   - Warm Redis with common requests

## Implementation Plan

### Tool: `cmd/import_historical/main.go`

```go
// Command-line tool for historical data import
Usage:
  import-historical --mode=full                    # Import all years 2010-2024
  import-historical --mode=year --year=2015       # Import specific year
  import-historical --mode=range --start=2010 --end=2015
  import-historical --mode=validate               # Validate existing data
  import-historical --mode=stats                  # Show import progress

Features:
  - Progress tracking with resume capability
  - Detailed logging to file and console
  - Error recovery and retry logic
  - Dry-run mode for testing
  - Concurrent downloads (rate-limited)
  - Database transaction batching
```

### Database Schema Additions

```sql
-- Track import progress
CREATE TABLE IF NOT EXISTS import_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season INT NOT NULL,
    data_type VARCHAR(50) NOT NULL, -- 'rosters', 'games', 'stats', 'ngs'
    status VARCHAR(20) NOT NULL,     -- 'pending', 'in_progress', 'completed', 'failed'
    records_imported INT DEFAULT 0,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(season, data_type)
);

-- Track data quality metrics
CREATE TABLE IF NOT EXISTS data_quality_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season INT NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value JSONB,
    measured_at TIMESTAMP DEFAULT NOW()
);
```

## Estimated Data Volumes

### Per Season (approximate):
- **Games**: 267 games (16 teams × 17 weeks = 256 regular + playoffs)
- **Player Stats**: ~15,000 records (267 games × ~56 players per game)
- **Rosters**: ~1,700 players (32 teams × ~53 players)
- **Next Gen Stats**: ~500 records per stat type

### Total for 15 Years (2010-2024):
- **Games**: ~4,005 games
- **Player Stats**: ~225,000 records
- **Unique Players**: ~8,000-10,000 players
- **Next Gen Stats**: ~40,000 records
- **Database Size**: ~2-3 GB (estimated)

## Performance Considerations

### Import Speed Optimization:
1. **Batch Inserts**: Insert 1000 records per transaction
2. **Parallel Processing**: Download CSVs concurrently (max 5 at a time)
3. **Upsert Strategy**: Use `ON CONFLICT DO UPDATE` for players/teams
4. **Index Management**: Drop indexes during bulk insert, rebuild after
5. **Connection Pooling**: Reuse database connections

### Expected Import Times:
- **Roster per year**: ~30 seconds
- **Games per year**: ~10 seconds
- **Stats per year**: ~2-3 minutes
- **Total per year**: ~4-5 minutes
- **Full 15 years**: ~60-75 minutes

## Risk Mitigation

### Risks & Solutions:

1. **CSV Download Failures**
   - **Risk**: GitHub releases unavailable or rate-limited
   - **Solution**: Retry logic, local caching, resume capability

2. **Data Schema Mismatches**
   - **Risk**: NFLverse CSV columns change over time
   - **Solution**: Flexible column mapping, version detection

3. **Database Connection Issues**
   - **Risk**: Heroku connection limits, timeouts
   - **Solution**: Connection pooling, batch commits, transaction management

4. **Data Corruption**
   - **Risk**: Partial imports, duplicate records
   - **Solution**: Transaction rollback, unique constraints, validation queries

5. **Performance Degradation**
   - **Risk**: Large imports slow down API
   - **Solution**: Run during off-peak hours, use read replicas

## Execution Schedule

### Week 1:
- **Monday**: Build CSV parsers and import tool (Phase 1)
- **Tuesday**: Import 2010-2014 (5 years) + validate
- **Wednesday**: Import 2015-2019 (5 years) + validate
- **Thursday**: Import 2020-2024 (5 years) + validate
- **Friday**: Advanced data import (NGS, injuries)
- **Weekend**: Validation, optimization, cache warming

### Success Metrics:
- ✅ All 15 years imported with <1% error rate
- ✅ API response times remain <200ms
- ✅ Zero data integrity violations
- ✅ 95%+ coverage of known games/players

## Post-Import Actions

1. **Update Documentation**
   - Document data coverage (2010-2024)
   - Update API docs with historical endpoints
   - Add data freshness indicators

2. **Enable Historical Endpoints**
   - `/api/v2/players/{id}/career` - Full career stats
   - `/api/v2/stats/leaders?season=YYYY` - Historical leaders
   - `/api/v2/games?season=YYYY` - Historical games

3. **Analytics Enhancements**
   - AI predictions with 15 years of training data
   - Player comparison across eras
   - Historical trend analysis

4. **Monitoring**
   - Track query performance on historical data
   - Monitor database size growth
   - Set up alerts for data freshness

## Maintenance Strategy

### Ongoing Updates:
- **Weekly**: Update current season stats
- **Daily**: Update injury reports
- **Real-time**: Game scores during season
- **Annual**: Import new season data

### Data Refresh Policy:
- **Current season**: Update daily
- **Previous season**: Update weekly for corrections
- **Historical (>1 year old)**: Immutable, unless corrections needed

## Rollback Plan

If import fails or causes issues:

1. **Immediate Actions**:
   ```sql
   -- Rollback to pre-import state
   DELETE FROM game_stats WHERE season < 2025;
   DELETE FROM games WHERE season < 2025;
   DELETE FROM player_season_stats WHERE season < 2025;
   ```

2. **Restore from Backup**:
   ```bash
   heroku pg:backups:restore --app grid-iron-mind
   ```

3. **Investigate & Fix**:
   - Review error logs
   - Fix import logic
   - Test on staging first

## Appendix A: NFLverse CSV Schemas

### player_stats.csv
```
player_id, player_name, season, week, team, position,
completions, attempts, passing_yards, passing_tds, interceptions,
carries, rushing_yards, rushing_tds,
receptions, targets, receiving_yards, receiving_tds,
fantasy_points, fantasy_points_ppr
```

### schedule.csv
```
game_id, season, week, gameday, gametime,
away_team, away_score, home_team, home_score,
location, roof, surface, temp, wind
```

### roster.csv
```
season, team, position, jersey_number, status,
full_name, first_name, last_name, birth_date,
height, weight, college, gsis_id, espn_id
```

## Appendix B: Useful Queries

### Check Import Progress
```sql
SELECT season, data_type, status, records_imported,
       completed_at - started_at as duration
FROM import_progress
ORDER BY season DESC, data_type;
```

### Validate Data Coverage
```sql
SELECT season,
       COUNT(DISTINCT game_id) as games,
       COUNT(*) as stats_records
FROM game_stats
GROUP BY season
ORDER BY season DESC;
```

### Find Missing Data
```sql
WITH expected AS (
  SELECT generate_series(2010, 2024) as season
)
SELECT e.season,
       COUNT(DISTINCT g.id) as games_count,
       CASE WHEN COUNT(DISTINCT g.id) = 0 THEN 'MISSING'
            WHEN COUNT(DISTINCT g.id) < 250 THEN 'INCOMPLETE'
            ELSE 'COMPLETE' END as status
FROM expected e
LEFT JOIN games g ON e.season = g.season
GROUP BY e.season
ORDER BY e.season;
```
