# Team Stats API - Ready for Testing

## Overview

The team statistics sync and API endpoints are now fully implemented and ready for testing.

## What's Complete

### 1. ✅ Team Stats Sync Implementation
- **File:** `internal/ingestion/team_stats.go`
- **Functions:**
  - `SyncTeamStats(season, week)` - Sync specific week
  - `SyncTeamStatsForSeason(season)` - Sync entire season (18 weeks)
  - `syncGameTeamStats()` - Per-game sync logic
  - `insertOrUpdateTeamStats()` - Database upsert
  - `parseTeamStats()` - ESPN boxscore parser

### 2. ✅ Admin Endpoint
- **Endpoint:** `POST /api/v1/admin/sync/team-stats`
- **File:** `internal/handlers/admin.go` (line 431)
- **Status:** Wired up and calls `SyncTeamStats()`

### 3. ✅ Query Endpoint
- **Endpoint:** `GET /api/v1/games/:id/stats`
- **File:** `internal/handlers/games.go` (HandleGameStats)
- **Returns:** Team statistics for both teams in a game
- **Status:** Fully implemented with team name joins

### 4. ✅ Database Schema
- **Table:** `game_team_stats`
- **Migration:** `003_enhance_comprehensive_schema.sql`
- **Constraints:** UNIQUE(game_id, team_id) for upsert safety
- **Indexes:** game_id, team_id, total_yards

## How to Test

### Step 1: Sync Team Stats for a Specific Week

```bash
# Sync Week 4 of 2025 season
curl -X POST "http://localhost:8080/api/v1/admin/sync/team-stats" \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "season": 2025,
    "week": 4
  }'
```

**Expected Response:**
```json
{
  "data": {
    "message": "Team stats sync completed for season 2025, week 4",
    "season": 2025,
    "week": 4,
    "status": "success"
  }
}
```

### Step 2: Query Team Stats for a Game

First, get a game ID:
```bash
curl "http://localhost:8080/api/v1/games?season=2025&week=4&limit=1"
```

Then query team stats:
```bash
# Replace {game-id} with actual UUID
curl "http://localhost:8080/api/v1/games/{game-id}/stats"
```

**Expected Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "game_id": "uuid",
      "team_id": "uuid",
      "team_name": "Kansas City Chiefs",
      "team_abbr": "KC",
      "first_downs": 24,
      "total_yards": 398,
      "passing_yards": 275,
      "rushing_yards": 123,
      "third_down_attempts": 12,
      "third_down_conversions": 7,
      "third_down_pct": 58.33,
      "turnovers": 1,
      "fumbles_lost": 0,
      "interceptions_thrown": 1,
      "penalties": 5,
      "penalty_yards": 45,
      "possession_time": "32:15",
      "completions": 25,
      "pass_attempts": 35,
      "sacks_allowed": 2,
      "rushing_attempts": 28
    },
    {
      "team_name": "Los Angeles Chargers",
      "team_abbr": "LAC",
      ...
    }
  ]
}
```

### Step 3: Sync Entire Season (Background Task)

For backfilling historical data:

```bash
# This will sync all 18 weeks of 2024 season (takes ~15-20 min)
curl -X POST "http://localhost:8080/api/v1/admin/sync/team-stats-season" \
  -H "X-API-Key: your-admin-key" \
  -H "Content-Type: application/json" \
  -d '{
    "season": 2024
  }'
```

**Note:** This endpoint needs to be added if you want season-level sync via API. Currently, `SyncTeamStatsForSeason()` exists but isn't exposed via HTTP endpoint.

## ESPN Statistics Mapped

The sync parses these ESPN boxscore fields:

| ESPN Field | Database Column | Example |
|------------|-----------------|---------|
| `firstDowns` | `first_downs` | 24 |
| `totalYards` | `total_yards` | 398 |
| `netPassingYards` | `passing_yards` | 275 |
| `rushingYards` | `rushing_yards` | 123 |
| `thirdDownEff` | `third_down_*` | "7-12" → 7 conv, 12 att |
| `turnovers` | `turnovers` | 1 |
| `fumblesLost` | `fumbles_lost` | 0 |
| `passesIntercepted` | `interceptions_thrown` | 1 |
| `penalties` | `penalties`, `penalty_yards` | "5-45" → 5 penalties, 45 yards |
| `possessionTime` | `possession_time` | "32:15" |
| `completionAttempts` | `completions`, `pass_attempts` | "25-35" |
| `sacksYardsLost` | `sacks_allowed` | "2-14" → 2 sacks |
| `rushingAttempts` | `rushing_attempts` | 28 |

## Validation Checklist

- [ ] Sync endpoint returns 200 for valid week
- [ ] Sync endpoint returns 400 for invalid week
- [ ] Sync logs show ESPN API calls
- [ ] Database contains 2 rows per game (home + away)
- [ ] Query endpoint returns both teams' stats
- [ ] Stats include team_name and team_abbr from join
- [ ] All numeric fields populated correctly
- [ ] Third down percentage calculated correctly
- [ ] Re-running sync updates existing records (no duplicates)
- [ ] Cache invalidated after sync

## Performance Expectations

- **Single game sync:** ~2-3 seconds (1 ESPN API call)
- **Full week sync (16 games):** ~45-60 seconds
- **Full season sync (18 weeks):** ~15-20 minutes

## Troubleshooting

### Issue: "No completed games found"
- Verify games exist: `SELECT COUNT(*) FROM games WHERE season=2025 AND week=4 AND status='completed'`
- If status is 'scheduled', games haven't been played yet

### Issue: "Team not found for nfl_id"
- Verify teams synced: `SELECT COUNT(*) FROM teams`
- Should have 32 teams
- Run: `POST /api/v1/admin/sync/teams`

### Issue: "No boxscore data available"
- ESPN API may not have stats yet (game too recent or postponed)
- Check ESPN API directly: `site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event={nfl_game_id}`

### Issue: Stats returning NULL values
- The sync only populates fields ESPN provides
- Fields like `offensive_plays`, `yards_per_play`, `fourth_down_*` may be NULL if ESPN doesn't include them
- This is expected - not all stats available for all games

## Next Steps

1. **Test with Real Data**
   - Sync Week 4 of 2025 season
   - Verify stats match ESPN.com box scores
   - Check all fields populated

2. **Backfill Historical Data**
   - Run `SyncTeamStatsForSeason(2024)`
   - Run `SyncTeamStatsForSeason(2025)`
   - Monitor for parsing errors

3. **Create Additional Query Endpoints**
   - `GET /api/v1/teams/:id/stats?season=2025&week=4` - Team stats for specific game
   - `GET /api/v1/teams/:id/stats/season?season=2025` - Aggregated season stats
   - `GET /api/v1/stats/team-leaders?stat=total_yards&season=2025` - Leaderboards

4. **Add to Automated Sync**
   - Add team stats sync to daily cron job
   - Run after games complete (Monday mornings)
   - Include in `sync2025` CLI tool

## Files Modified

1. `internal/ingestion/team_stats.go` - Created (370 lines)
2. `internal/handlers/admin.go` - Updated (line 431: changed to call SyncTeamStats)
3. `internal/handlers/games.go` - Already had HandleGameStats endpoint
4. `migrations/003_enhance_comprehensive_schema.sql` - Already had schema

## API Endpoints Summary

**Admin Endpoints (Require API Key):**
- `POST /api/v1/admin/sync/team-stats` ✅ - Sync specific week

**Public Endpoints:**
- `GET /api/v1/games/:id/stats` ✅ - Get team stats for game

**Missing (Optional Additions):**
- `POST /api/v1/admin/sync/team-stats-season` - Sync entire season
- `GET /api/v1/teams/:id/stats` - Team-specific stats queries
- `GET /api/v1/stats/team-leaders` - Team statistical leaders

## Status

**🟢 READY FOR TESTING**

The team stats sync is fully implemented, wired up, and ready to test with real ESPN data. All core functionality is in place.
