# 2025 NFL Season Data Sync Guide

This guide explains how to load and maintain up-to-date 2025 NFL season data in the Grid Iron Mind database.

## Overview

The `sync2025` tool provides multiple modes for syncing 2025 season data:

- **Full Sync**: Initial complete load of all 2025 data
- **Update Sync**: Regular updates for current week
- **Live Sync**: Real-time updates during game day (every 5 minutes)
- **Stats Sync**: Player statistics only
- **Injuries Sync**: Injury reports only

## Quick Start

### 1. Build the Sync Tool

```bash
cd /path/to/gridironmind
go build -o bin/sync2025 cmd/sync2025/main.go
```

### 2. Initial Full Load

Run this once to load all 2025 data:

```bash
./bin/sync2025 full
```

This will:
- ✓ Sync all 32 NFL teams
- ✓ Load current rosters for all teams
- ✓ Import all 2025 games (weeks 1-18)
- ✓ Load team stats for completed games
- ✓ Import player season statistics
- ✓ Load current injury reports

**Time estimate**: 30-60 minutes for complete initial load

### 3. Set Up Automated Updates

Install the cron schedule:

```bash
# Create logs directory
mkdir -p logs

# Install crontab
crontab scripts/crontab-2025.txt
```

Or use the wrapper script manually:

```bash
./scripts/sync-2025-schedule.sh
```

## Sync Modes

### Full Sync

Complete initial load of all 2025 season data.

```bash
./bin/sync2025 full
```

**When to use**:
- First time setup
- After major database changes
- To rebuild complete dataset

**What it syncs**:
1. All teams and metadata
2. Complete rosters (all players)
3. All games for 2025 season (weeks 1-18)
4. Team stats for all completed games
5. Player season statistics from NFLverse
6. Current injury reports

### Update Sync

Regular update of current week's data.

```bash
./bin/sync2025 update
```

**When to use**:
- Daily updates during the season
- After games complete
- When rosters change

**What it syncs**:
1. Current rosters (trades, signings)
2. Current week's games
3. Team stats for current week
4. Updated player season stats
5. Current injury reports

**Time estimate**: 2-5 minutes

### Live Sync

Continuous updates during game day (runs for duration specified or until stopped).

```bash
./bin/sync2025 live
```

**When to use**:
- During Sunday/Monday game days
- For real-time score updates
- Live game tracking

**What it syncs** (every 5 minutes):
- Current game scores
- Game status updates
- Team stats for newly completed games

**Note**: This mode runs continuously. Stop with Ctrl+C or use timeout:

```bash
# Run for 3 hours
timeout 3h ./bin/sync2025 live
```

### Stats Sync

Player statistics only (no games or rosters).

```bash
./bin/sync2025 stats
```

**When to use**:
- After weekly stats are finalized
- To update player performance data
- Debugging stats-related issues

**What it syncs**:
- Player season stats from NFLverse
- Individual career stats from ESPN

**Time estimate**: 10-15 minutes

### Injuries Sync

Injury reports only (fast operation).

```bash
./bin/sync2025 injuries
```

**When to use**:
- Daily before games
- After practice reports
- When injury news breaks

**What it syncs**:
- Current injury status for all players
- Return dates
- Injury details

**Time estimate**: 1-2 minutes

## Automated Scheduling

### Recommended Schedule

The provided cron schedule (`scripts/crontab-2025.txt`) implements this strategy:

| When | What | Why |
|------|------|-----|
| **Sunday 1pm-11pm ET** | Hourly updates | Game day - live scores |
| **Monday 8pm-11pm ET** | Hourly updates | Monday Night Football |
| **Monday 9am** | Full roster refresh | Process weekend transactions |
| **Tuesday-Saturday 6am** | Daily update | Off-day maintenance |
| **Daily 3pm** | Injury reports | Before practice reports |

### Custom Schedule

You can customize the schedule by editing `scripts/crontab-2025.txt`:

```cron
# Every 30 minutes on Sunday
*/30 13-23 * * 0 cd $PROJECT_PATH && ./bin/sync2025 update

# Twice daily Monday-Saturday
0 6,18 * * 1-6 cd $PROJECT_PATH && ./bin/sync2025 update

# Injuries three times daily
0 9,15,21 * * * cd $PROJECT_PATH && ./bin/sync2025 injuries
```

## Data Sources

The sync tool pulls from multiple sources:

### ESPN API
- Teams and rosters
- Game schedule and scores
- Player metadata
- Team game statistics
- Injury reports

### NFLverse
- Advanced player statistics
- Enhanced schedule data
- Next Gen Stats (when available)

## Monitoring

### Check Sync Status

View recent sync logs:

```bash
tail -f logs/sync-2025.log
```

### Database Queries

Check data freshness:

```sql
-- Most recent game update
SELECT game_date, status, updated_at
FROM games
WHERE season = 2025
ORDER BY updated_at DESC
LIMIT 10;

-- Player stats coverage
SELECT season, COUNT(*) as player_count
FROM player_season_stats
WHERE season = 2025
GROUP BY season;

-- Team stats coverage
SELECT COUNT(*) as games_with_stats
FROM game_team_stats gts
JOIN games g ON gts.game_id = g.id
WHERE g.season = 2025;
```

### Troubleshooting

**Problem**: Sync fails with "rate limited"
**Solution**: ESPN is rate limiting. Wait 5 minutes and retry. Consider adding delays.

**Problem**: No games found for current week
**Solution**: Check if games are scheduled. Use `./bin/sync2025 full` to resync entire season.

**Problem**: Player stats not updating
**Solution**: Run `./bin/sync2025 stats` manually. NFLverse may have delays.

**Problem**: Database connection errors
**Solution**: Verify `DATABASE_URL` in `.env` file. Check database is running.

## Performance

### Optimization Tips

1. **Use indexes**: The schema includes optimized indexes for common queries
2. **Batch operations**: The sync tool batches database operations
3. **Rate limiting**: Built-in delays prevent API throttling
4. **Incremental updates**: Update mode only syncs current week

### Resource Usage

- **CPU**: Minimal (I/O bound)
- **Memory**: ~100MB during sync
- **Network**: ~10-50MB per full week sync
- **Database**: ~500MB for complete 2025 season

## API Rate Limits

### ESPN API
- **Limit**: Not officially documented
- **Safe rate**: 1 request per second
- **Built-in delays**: 500ms - 2 seconds between requests

### NFLverse
- **Limit**: None (public data)
- **Files**: CSV files hosted on GitHub
- **Refresh**: Daily updates

## Environment Variables

Required in `.env`:

```bash
# Database connection
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# Optional: Weather data enrichment
WEATHER_API_KEY=your_api_key_here

# Redis cache (if used)
REDIS_URL=redis://localhost:6379
```

## Manual Operations

### Sync Specific Week

```bash
# Edit cmd/sync2025/main.go to accept week parameter
# Or run SQL queries directly
```

### Resync Single Team

```bash
# Use the existing ingestion service methods
go run cmd/sync2025/main.go
```

### Export Data

```bash
# Export 2025 games to CSV
psql $DATABASE_URL -c "COPY (SELECT * FROM games WHERE season = 2025) TO STDOUT CSV HEADER" > games-2025.csv
```

## Best Practices

1. **Start with full sync**: Always run `full` mode first
2. **Monitor logs**: Keep an eye on `logs/sync-2025.log`
3. **Check data quality**: Run validation queries after syncs
4. **Rate limit awareness**: Don't run multiple syncs simultaneously
5. **Backup database**: Before major syncs, backup your database
6. **Test in staging**: Test sync scripts in non-production first

## Support

For issues or questions:
- Check logs: `logs/sync-2025.log`
- Review ESPN API status
- Verify database connection
- Check NFLverse data availability

## Future Enhancements

Potential improvements:
- [ ] Webhook notifications on sync completion
- [ ] Sync status dashboard
- [ ] Automatic retry with exponential backoff
- [ ] Parallel team syncing
- [ ] Real-time game event streaming
- [ ] Historical seasons (2020-2024)
