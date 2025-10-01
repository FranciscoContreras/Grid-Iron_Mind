# 2025 Season Data Sync - Quick Start

## Setup

### 1. Build the sync tool

```bash
go build -o bin/sync2025 cmd/sync2025/main.go
```

### 2. Run initial full sync

This will load all 2025 season data into your database:

```bash
./bin/sync2025 full
```

**What gets loaded**:
- ✓ All 32 NFL teams
- ✓ Complete rosters (~1,700 players)
- ✓ All 2025 games (272 regular season games across 18 weeks)
- ✓ Team statistics for completed games
- ✓ Player season statistics
- ✓ Current injury reports

**Time**: 30-60 minutes

### 3. Set up automated updates

```bash
# Install the cron schedule
crontab scripts/crontab-2025.txt

# Or run manually
./scripts/sync-2025-schedule.sh
```

## Usage

### Available Commands

```bash
# Full initial sync (run once)
./bin/sync2025 full

# Regular update (run daily)
./bin/sync2025 update

# Live updates during games (run on game days)
./bin/sync2025 live

# Player stats only
./bin/sync2025 stats

# Injury reports only (fast)
./bin/sync2025 injuries
```

### Recommended Schedule

- **Game Days (Sun/Mon)**: Run `live` mode during game hours
- **Monday Morning**: Run `update` to catch roster moves
- **Tuesday-Saturday**: Run `update` once daily
- **Daily**: Run `injuries` for latest injury news

## Monitoring

View sync logs:

```bash
tail -f logs/sync-2025.log
```

Check database:

```sql
-- Recent games
SELECT * FROM games
WHERE season = 2025
ORDER BY game_date DESC
LIMIT 10;

-- Player stats count
SELECT COUNT(*) FROM player_season_stats WHERE season = 2025;

-- Team stats coverage
SELECT COUNT(*) FROM game_team_stats gts
JOIN games g ON gts.game_id = g.id
WHERE g.season = 2025;
```

## Troubleshooting

**Build fails**: Make sure Go 1.21+ is installed
```bash
go version
```

**Database connection error**: Check your `.env` file has valid `DATABASE_URL`

**ESPN rate limiting**: Wait 5 minutes between full syncs

**No games found**: Verify the current NFL week is correct

## Documentation

Full documentation: [docs/2025-DATA-SYNC.md](docs/2025-DATA-SYNC.md)

## Architecture

```
cmd/sync2025/main.go          # Main sync orchestrator
  ├── internal/ingestion/     # Data ingestion service
  ├── internal/espn/          # ESPN API client
  ├── internal/nflverse/      # NFLverse data client
  └── internal/db/            # Database operations

scripts/
  ├── sync-2025-schedule.sh   # Automated sync wrapper
  └── crontab-2025.txt        # Cron schedule config
```

## What Gets Synced

### From ESPN API
- Teams and metadata (logos, colors, stadium info)
- Current rosters with player details
- Game schedule and live scores
- Game status updates (scheduled, in_progress, completed)
- Team game statistics (box scores)
- Injury reports with details

### From NFLverse
- Advanced player statistics
- Historical performance data
- Enhanced schedule information
- Next Gen Stats (when available)

## Database Tables Updated

- `teams` - NFL team information
- `players` - Player profiles and metadata
- `games` - Game schedule and scores
- `game_team_stats` - Team statistics per game
- `player_season_stats` - Player statistics per season
- `player_injuries` - Current injury reports

## Environment Variables

Required in `.env`:

```bash
DATABASE_URL=postgresql://user:pass@host:5432/dbname
WEATHER_API_KEY=your_key_here  # Optional
REDIS_URL=redis://localhost:6379  # Optional
```

## Next Steps

After initial sync:

1. ✓ Verify data in database
2. ✓ Set up automated cron job
3. ✓ Configure monitoring/alerts
4. ✓ Test API endpoints with 2025 data
5. ✓ Set up backup schedule

## Performance

- **Initial full sync**: 30-60 minutes
- **Daily update**: 2-5 minutes
- **Live game sync**: Continuous (5 min intervals)
- **Injury sync**: 1-2 minutes
- **Database size**: ~500MB for full season
