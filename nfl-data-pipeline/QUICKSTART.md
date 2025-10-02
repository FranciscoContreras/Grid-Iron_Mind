# Quick Start Guide - NFL Data Pipeline (Rust)

## Prerequisites

You need Rust installed on your system. If you don't have it:

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
```

Verify installation:
```bash
rustc --version
cargo --version
```

## Setup

### 1. Navigate to Pipeline Directory

```bash
cd nfl-data-pipeline
```

### 2. Configure Database Connection

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` with your Heroku database URL:
```bash
# Get your Heroku database URL
heroku config:get DATABASE_URL

# Edit .env
nano .env
```

Update the `DATABASE_URL` line with your actual Heroku PostgreSQL connection string.

### 3. Build the Pipeline

```bash
make build
```

This compiles an optimized release binary at `target/release/nfl-data-pipeline`.

## Usage

### Test with Single Season (Recommended First)

Start with importing just 2024 data to verify everything works:

```bash
make sync-year YEAR=2024
```

**Expected output:**
```
🏈 NFL Data Pipeline Starting
Mode: year
Year range: 2024-2024
📅 Importing data for year 2024...
  ✅ Rosters: 3216 players
  ✅ Schedule: 272 games
  ✅ Player Stats: 45000+ records
  ✅ NGS Passing: 1500+ records
✅ Year 2024 import complete
✅ Pipeline completed successfully!
```

**Duration:** 3-5 minutes

### Full Historical Import (2010-2025)

Once the test succeeds, run the full import:

```bash
make sync-full
```

**Expected output:**
```
⚠️  WARNING: Full import takes 30-60 minutes
📦 Importing 16 seasons of data (2010-2025)...
[Progress bar: 1/16] Importing 2010
  ✅ Rosters: 2800 players
  ✅ Schedule: 256 games
  ✅ Player Stats: 38000+ records
[Progress bar: 2/16] Importing 2011
...
```

**Duration:** 30-60 minutes (depending on network speed)

**What gets imported:**
- 16 seasons (2010-2025)
- ~50,000 unique players (deduplicated)
- ~4,300 games
- ~720,000 player stat records
- ~160,000 Next Gen Stats records (2016+)

### Other Commands

**Update current week only** (fast daily sync):
```bash
make sync-update
```

**Live game day sync** (continuous updates):
```bash
make sync-live
```

**Validate existing data:**
```bash
psql $DATABASE_URL -c "SELECT COUNT(*) FROM players;"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM games;"
psql $DATABASE_URL -c "SELECT season, COUNT(*) FROM games GROUP BY season ORDER BY season;"
```

## Automation

### Set Up Automated Cron Jobs

For continuous data updates, install the cron schedule:

```bash
make install-cron
```

This displays the cron configuration. To activate it:

```bash
crontab -e
```

Then paste the schedule from `scripts/crontab-2025.txt`.

**Schedule summary:**
- **Daily 3 AM**: Update previous day's games
- **Sunday 1-11 PM**: Live game updates
- **Monday 8-11 PM**: Monday Night Football
- **Thursday 8-11 PM**: Thursday Night Football
- **Tuesday 5 AM**: Weekly player stats sync
- **Wednesday 10 AM**: Injury report sync

## Monitoring

### View Logs

```bash
make logs
```

### Check Database Status

```bash
make db-status
```

**Output:**
```
🗄️  Database Status:
 players | 50000
 games   | 4300
 stats   | 720000

season | data_type     | status    | records_imported
-------+---------------+-----------+------------------
2025   | rosters       | completed | 3200
2025   | schedule      | completed | 272
2025   | player_stats  | completed | 45000
...
```

## Troubleshooting

### Connection Errors

If you see `connection refused` or `authentication failed`:

1. Verify DATABASE_URL is correct:
   ```bash
   echo $DATABASE_URL
   heroku config:get DATABASE_URL
   ```

2. Test connection manually:
   ```bash
   psql $DATABASE_URL -c "SELECT 1;"
   ```

3. Ensure Heroku database is running:
   ```bash
   heroku pg:info
   ```

### Team Abbreviation Errors

If you see "team not found" errors like `OAK not found`:

- The pipeline automatically maps historical teams (OAK→LV, STL→LA, SD→LAC)
- If errors persist, check that teams table is populated:
  ```bash
  psql $DATABASE_URL -c "SELECT abbreviation FROM teams ORDER BY abbreviation;"
  ```

### Memory Issues

If the pipeline crashes with "out of memory":

1. Reduce batch size in `.env`:
   ```
   BATCH_SIZE=250
   ```

2. Run year-by-year instead of full import:
   ```bash
   for year in {2010..2025}; do
     make sync-year YEAR=$year
   done
   ```

### Network Timeouts

If downloads fail:

1. Increase retry count in `.env`:
   ```
   MAX_RETRIES=5
   ```

2. Check nflfastr availability:
   ```bash
   curl -I https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_2024.csv
   ```

## Data Sources

**Primary:** NFLverse (nflfastr)
- Player stats: `https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_{year}.csv`
- Rosters: `.../rosters/roster_{year}.csv`
- Schedule: `.../schedules/sched_{year}.csv`
- NGS: `.../nextgen_stats/ngs_{year}_passing.csv`

**Secondary:** ESPN API (for current season gaps)

## Next Steps

After successful import:

1. **Verify API Performance:**
   ```bash
   curl https://nfl.wearemachina.com/api/v1/players?limit=100
   curl https://nfl.wearemachina.com/api/v1/games?season=2024&week=1
   ```

2. **Check Query Speed:**
   - Should see sub-200ms response times
   - X-Cache headers should show hits after first query

3. **Set Up Automation:**
   - Install cron jobs for continuous updates
   - Monitor sync logs daily

4. **Backfill Missing Data:**
   - Use admin endpoints to fill any gaps
   - Run validation queries to identify missing records

## Support

For issues or questions:
- Check logs: `make logs`
- Validate database: `make db-status`
- Rebuild: `make clean && make build`
- GitHub: https://github.com/francisco/gridironmind
