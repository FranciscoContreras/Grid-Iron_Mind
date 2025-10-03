# Live Game Data Sync System - Comprehensive Guide

## Overview

The Grid Iron Mind API now includes an **intelligent internal scheduler** that automatically keeps game data up-to-date during live NFL games. The system runs within the API server process and dynamically adjusts sync frequency based on game schedule and current time.

## Architecture

### Multi-Tier Sync Strategy

The scheduler operates in **4 different modes**, automatically switching based on NFL season and game schedule:

#### 1. **Live Mode** (Every 1 minute)
- **When**: During peak game hours
  - **Thursday**: 5pm-11pm PT (8pm-11pm ET) - Thursday Night Football
  - **Sunday**: 10am-11pm PT (1pm-11pm ET) - All Sunday games
  - **Monday**: 5pm-11pm PT (8pm-11pm ET) - Monday Night Football
  - **Saturday** (Week 15+): 10am-11pm PT - Late season Saturday games
- **Purpose**: Capture live scores, status changes, and in-game updates in real-time

#### 2. **Active Mode** (Every 5 minutes)
- **When**: Game days outside of peak game hours
- **Purpose**: Catch pre-game updates, post-game stats, and injury reports

#### 3. **Standard Mode** (Every 15 minutes)
- **When**: Non-game days during the NFL season
- **Purpose**: Keep rosters, standings, and stats updated between games

#### 4. **Idle Mode** (Every 1 hour)
- **When**: NFL offseason (March-August)
- **Purpose**: Minimal maintenance syncing for offseason activity

### What Gets Synced

Each sync iteration performs the following operations:

1. **Game Scores & Status** (Always)
   - Fetches current scoreboard from ESPN
   - Updates game scores, status (scheduled/in_progress/completed)
   - Updates current quarter, game clock, attendance
   - Captures venue and broadcast information

2. **Team Statistics** (During season)
   - Box score stats for completed games
   - Offensive/defensive metrics
   - Time of possession, turnovers, penalties
   - Third/fourth down efficiency, red zone stats

3. **Injury Reports** (Once daily at 3am PT)
   - All team injury reports from ESPN
   - Player status, injury type, return timeline
   - Updates player statuses (active/injured/inactive)

4. **Cache Invalidation** (After sync)
   - Clears relevant cache entries for updated data
   - Ensures fresh data is returned to API clients
   - Patterns: games, teams, stats, standings, defense

## How It Works

### 1. Server Startup

When the API server starts:

```go
// Scheduler is initialized with default config
schedulerConfig := scheduler.DefaultConfig(cfg.WeatherAPIKey)
autoScheduler := scheduler.NewScheduler(schedulerConfig)
autoScheduler.Start()
```

The scheduler immediately:
1. Runs an initial sync
2. Detects current season and mode
3. Schedules next sync based on mode

### 2. Game Day Detection

The `GameDetector` component:
- Queries database for games scheduled today
- Checks current time against NFL game hours
- Determines appropriate sync mode
- Provides game summary for logging

Example detection logic:
```go
seasonInfo := utils.GetCurrentSeason()  // Get current NFL season/week
hasGamesToday := detector.HasGamesToday(ctx)  // Check DB for today's games
isGameTime := detector.IsGameTime()  // Check if we're in game hours

if isGameTime && hasGamesToday {
    mode = SyncModeLive  // Every 1 minute
} else if hasGamesToday {
    mode = SyncModeActive  // Every 5 minutes
} else {
    mode = SyncModeStandard  // Every 15 minutes
}
```

### 3. Sync Execution

Each sync iteration:

```
========================================
[SCHEDULER] Starting sync iteration at 14:23:05
========================================
[SCHEDULER] Season: 2025, Week: 5, Active: true
[SCHEDULER] Today's games: 16 total (3 live, 7 scheduled, 6 completed)

[SCHEDULER] [1/4] Syncing games...
[SCHEDULER] Fetching games for season 2025, week 5
[SCHEDULER] ✓ Games synced successfully

[SCHEDULER] [2/4] Syncing team stats...
[SCHEDULER] Syncing team stats for season 2025, week 5
[SCHEDULER] ✓ Team stats synced successfully

[SCHEDULER] [3/4] Skipping injury sync (not scheduled)

[SCHEDULER] [4/4] Clearing cache...
[SCHEDULER] ✓ Cache cleared successfully

========================================
[SCHEDULER] Sync completed successfully in 2.3s
========================================
[SCHEDULER] Next sync in 1m0s (mode: live)
```

### 4. Graceful Shutdown

When server shuts down:
```go
defer autoScheduler.Stop()
```

The scheduler:
- Receives shutdown signal via context
- Completes current sync if running
- Exits sync loop cleanly
- Closes connections

## Configuration

### Environment Variables

Control scheduler behavior via environment variables:

```bash
# Disable auto-sync entirely (defaults to enabled)
ENABLE_AUTO_SYNC=false

# Weather API key for weather enrichment during sync (optional)
WEATHER_API_KEY=your_key_here

# Database and Redis configs (required for sync to work)
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
```

### Default Intervals

Configured in `internal/scheduler/config.go`:

```go
LiveInterval:     1 * time.Minute   // During games
ActiveInterval:   5 * time.Minute   // Game days, off-hours
StandardInterval: 15 * time.Minute  // Regular season, no games
IdleInterval:     1 * time.Hour     // Offseason
```

### Customization

To customize sync behavior, modify `scheduler.DefaultConfig()`:

```go
schedulerConfig := scheduler.DefaultConfig(cfg.WeatherAPIKey)

// Override specific settings
schedulerConfig.LiveInterval = 30 * time.Second  // Faster live updates
schedulerConfig.SyncStats = false  // Disable stats sync
schedulerConfig.ClearCache = false  // Don't clear cache
```

## Admin API Endpoints

Control and monitor the scheduler via API:

### 1. Get Scheduler Status

```bash
GET /api/v1/admin/scheduler/status
Authorization: Bearer YOUR_API_KEY
```

**Response:**
```json
{
  "data": {
    "enabled": true,
    "running": true,
    "current_mode": "live",
    "next_sync": "2025-10-02T14:24:05Z",
    "last_sync": "2025-10-02T14:23:05Z",
    "last_error": "",
    "sync_count": 47,
    "error_count": 0,
    "interval": "1m0s",
    "season_info": "Season 2025, Week 5 (Regular Season)",
    "games_summary": "16 total (3 live, 7 scheduled, 6 completed)"
  },
  "meta": {
    "timestamp": "2025-10-02T14:23:15Z"
  }
}
```

### 2. Trigger Manual Sync

Force an immediate sync (useful for testing or urgent updates):

```bash
POST /api/v1/admin/scheduler/trigger
Authorization: Bearer YOUR_API_KEY
```

**Response:**
```json
{
  "data": {
    "message": "Sync triggered successfully",
    "status": "running"
  }
}
```

### 3. Update Configuration

Update scheduler settings at runtime:

```bash
POST /api/v1/admin/scheduler/configure
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json

{
  "enabled": true,
  "mode": "live",
  "sync_games": true,
  "sync_stats": true,
  "sync_injuries": false,
  "clear_cache": true
}
```

## Monitoring & Debugging

### Log Output

The scheduler produces detailed logs for monitoring:

```
[SCHEDULER] Starting auto-sync scheduler...
[SCHEDULER] Configuration: enabled=true, mode=
[SCHEDULER] Mode changed:  → live
[SCHEDULER] Next sync in 1m0s (mode: live)

========================================
[SCHEDULER] Starting sync iteration at 14:23:05
========================================
[SCHEDULER] Season: 2025, Week: 5, Active: true
[SCHEDULER] Today's games: 16 total (3 live, 7 scheduled, 6 completed)
[SCHEDULER] [1/4] Syncing games...
[SCHEDULER] Fetching games for season 2025, week 5
[SCHEDULER] ✓ Games synced successfully
...
========================================
[SCHEDULER] Sync completed successfully in 2.3s
========================================
```

### Error Handling

Errors are logged but don't stop the scheduler:

```
[SCHEDULER] ERROR syncing games: ESPN API timeout
[SCHEDULER] WARNING syncing team stats: no completed games found
[SCHEDULER] Sync completed with ERRORS in 1.2s
```

The scheduler will:
- Log errors clearly with context
- Continue to next sync iteration
- Track error count in status
- Retry on next scheduled sync

### Metrics

Monitor scheduler health via metrics endpoint:

```bash
GET /api/v1/metrics/health
```

Check for:
- Last sync time
- Error rates
- Database connection health
- API response times

## Implementation Details

### File Structure

```
internal/scheduler/
├── config.go          # Configuration and sync modes
├── game_detector.go   # Game day/time detection logic
└── scheduler.go       # Main scheduler service

internal/handlers/
└── scheduler.go       # Admin API handlers for scheduler

cmd/server/main.go     # Server startup with scheduler integration
```

### Key Components

#### `Scheduler` Service
- Manages sync loop in goroutine
- Determines mode dynamically
- Executes sync operations
- Tracks status and metrics

#### `GameDetector`
- Queries database for game schedule
- Checks if current time is during games
- Detects live games in progress
- Provides game summaries

#### `Config`
- Defines sync modes and intervals
- Controls what gets synced
- Configurable at startup

### Data Flow

```
Server Starts
    ↓
Scheduler.Start()
    ↓
Run Initial Sync → [Games] → [Stats] → [Injuries] → [Cache]
    ↓
Determine Mode (GameDetector)
    ↓
Schedule Next Sync
    ↓
Wait for Interval
    ↓
Run Sync → [Games] → [Stats] → [Injuries] → [Cache]
    ↓
(Loop until shutdown)
```

## Performance Considerations

### Resource Usage

**During Live Mode** (1 min intervals):
- ~60 syncs per hour
- ~2-5 seconds per sync
- ~1-2 ESPN API calls per sync
- Minimal CPU/memory impact

**Database Impact:**
- Upsert operations (idempotent)
- Indexed queries for game lookups
- Connection pool shared with API requests

**Network:**
- ESPN API rate limits respected
- Exponential backoff on failures
- 30-second timeout per request

### Optimization

The scheduler is optimized for production:

1. **Concurrent Operations**: Sync runs in separate goroutine
2. **Non-Blocking**: Doesn't block API request handling
3. **Graceful Shutdown**: Completes current sync before exit
4. **Error Recovery**: Continues on errors, logs clearly
5. **Cache Management**: Only clears relevant cache patterns

## Troubleshooting

### Scheduler Not Running

**Check logs for:**
```
[SCHEDULER] Already running
```

**Solution:** Scheduler is already active. Check status endpoint.

### No Data Updates

**Check:**
1. Scheduler enabled: `ENABLE_AUTO_SYNC` not set to `false`
2. Database connection: `DATABASE_URL` valid
3. ESPN API accessible: Network connectivity
4. Current mode: May be in `idle` during offseason

**Debug:**
```bash
# Check scheduler status
curl -H "Authorization: Bearer YOUR_KEY" \
  https://your-api.com/api/v1/admin/scheduler/status

# Trigger manual sync
curl -X POST -H "Authorization: Bearer YOUR_KEY" \
  https://your-api.com/api/v1/admin/scheduler/trigger
```

### Sync Errors

**Common errors:**

1. **"ESPN API timeout"**
   - ESPN servers slow/down
   - Network connectivity issues
   - Will retry on next sync

2. **"Database connection failed"**
   - Check `DATABASE_URL`
   - Verify connection pool settings
   - Check database health

3. **"No games found"**
   - Not an error - no games scheduled
   - Expected during offseason

### High Error Rate

If `error_count` is high:

1. Check ESPN API status
2. Verify database connectivity
3. Review error logs for patterns
4. Consider increasing sync intervals temporarily

## Best Practices

### Production Deployment

1. **Monitor Logs**: Watch for error patterns
2. **Track Metrics**: Use `/api/v1/metrics/health`
3. **Set Alerts**: Alert on high error rates
4. **Test Syncs**: Use trigger endpoint to test
5. **Cache Strategy**: Ensure Redis is available

### Development

1. **Disable for Local**: Set `ENABLE_AUTO_SYNC=false` if not needed
2. **Manual Triggers**: Use trigger endpoint for testing
3. **Check Status**: Monitor mode changes and sync results
4. **Test Modes**: Force different modes by time/data

### Scaling

The scheduler is designed for single-instance deployment:

- **DO**: Run on primary API server
- **DON'T**: Run on multiple instances (creates duplicate syncs)
- **Alternative**: Use external cron + sync CLI tool for multi-instance setups

## Comparison: Internal Scheduler vs External Cron

### Internal Scheduler (Recommended)

**Pros:**
- ✅ Intelligent mode switching
- ✅ Game-aware frequency adjustment
- ✅ No external dependencies
- ✅ Integrated with API lifecycle
- ✅ Real-time status monitoring
- ✅ No cron setup required

**Cons:**
- ❌ Runs per server instance
- ❌ Requires server restart to fully disable

### External Cron + CLI Tool

**Pros:**
- ✅ Centralized control
- ✅ Works with multiple API instances
- ✅ Easy to disable/modify schedule
- ✅ Separate process isolation

**Cons:**
- ❌ Fixed intervals (no intelligence)
- ❌ Requires cron setup
- ❌ External dependency
- ❌ Less responsive to live games

## Future Enhancements

Potential improvements:

1. **Player Stats Sync**: Add per-game player stats during live games
2. **Play-by-Play**: Sync real-time play-by-play data
3. **Smart Intervals**: Further refine intervals based on game status
4. **Webhook Support**: Trigger syncs on ESPN webhook events
5. **Multi-Instance**: Distributed lock for multi-server sync coordination
6. **Configurable API**: Full config updates via API endpoint
7. **Historical Backfill**: Automatic backfill of missing historical data

## Summary

The **Live Game Data Sync System** ensures your API always has the latest NFL data without manual intervention:

- 🎯 **Intelligent**: Auto-detects games and adjusts frequency
- ⚡ **Fast**: 1-minute updates during live games
- 🛡️ **Reliable**: Error handling and retry logic
- 📊 **Monitored**: Real-time status via API
- 🔧 **Configurable**: Control via environment and API
- 🚀 **Production-Ready**: Optimized for performance and reliability

The system starts automatically with your API server and "just works" - keeping your NFL data fresh 24/7 during the season!
