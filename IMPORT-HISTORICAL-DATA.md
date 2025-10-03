# Import Historical NFL Data (Last 10 Years)

## Quick Start

Import the last 10 years of NFL data (2015-2025) using the Rust pipeline:

```bash
cd nfl-data-pipeline

# Import 10 years of data (2015-2025)
./target/release/nfl-data-pipeline --mode full --start-year 2015 --end-year 2025
```

**Expected Time:** 30-45 minutes
**Data Imported:**
- ~15,000+ unique players
- ~2,700+ games
- ~400,000+ player stat records
- Next Gen Stats (2016+)

## Prerequisites

### 1. Build the Rust Pipeline

```bash
cd nfl-data-pipeline

# Install Rust (if not already installed)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# Build release binary
cargo build --release
```

### 2. Configure Database Connection

Create `.env` file in `nfl-data-pipeline/`:

```env
DATABASE_URL=your_postgres_connection_string
```

**Get Heroku database URL:**
```bash
heroku config:get DATABASE_URL --app grid-iron-mind
```

## Import Options

### Option 1: Last 10 Years (2015-2025) - Recommended

```bash
cd nfl-data-pipeline
make sync-full
```

This runs:
```bash
./target/release/nfl-data-pipeline --mode full --start-year 2010 --end-year 2025
```

### Option 2: Custom Year Range

```bash
cd nfl-data-pipeline

# Last 5 years only
./target/release/nfl-data-pipeline --mode full --start-year 2020 --end-year 2025

# Specific years
./target/release/nfl-data-pipeline --mode full --start-year 2022 --end-year 2024
```

### Option 3: Single Season

```bash
cd nfl-data-pipeline

# Import just 2024
make sync-year YEAR=2024

# Or manually
./target/release/nfl-data-pipeline --mode year --year 2024
```

### Option 4: Dry Run (Test First)

```bash
cd nfl-data-pipeline

# Test import without writing to database
./target/release/nfl-data-pipeline --mode full --start-year 2024 --end-year 2025 --dry-run
```

## What Gets Imported

For each season (2015-2025), the pipeline imports:

### 1. **Rosters** (Players)
- All active NFL players for that season
- Player names, positions, teams
- Physical stats (height, weight)
- College, draft info
- Jersey numbers

**Source:** `https://github.com/nflverse/nflverse-data/releases/download/rosters/roster_{year}.csv`

### 2. **Player Statistics**
- Weekly game stats (passing, rushing, receiving)
- Season totals
- Advanced metrics

**Source:** `https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_{year}.csv`

### 3. **Next Gen Stats** (2016+)
- Passing stats (air yards, completion probability)
- Rushing stats (rush attempts, yards before contact)
- Receiving stats (separation, catch rate)

**Source:** `https://github.com/nflverse/nflverse-data/releases/download/nextgen_stats/ngs_{year}_passing.csv`

## Progress Monitoring

The pipeline shows progress as it runs:

```
🏈 NFL Data Pipeline Starting
Mode: full
Year range: 2015-2025

[00:02:15] ████████████████████████████████ 11/11 2024

📅 Importing data for year 2024...
  ✅ Rosters: 1842 players
  ⏭️  Schedule: Skipping (use Go importer with ESPN API)
  ✅ Player Stats: 32847 records
  ✅ NGS Passing: 485 records
✅ Year 2024 import complete

✅ Pipeline completed successfully!
```

## Verify Import

After import completes, check database:

```bash
cd nfl-data-pipeline
make db-status
```

**Expected output:**
```
🗄️  Database Status:
 players
---------
    15483

 games
-------
    2856

 stats
---------
  427589

 season | data_type    | status    | records_imported
--------+--------------+-----------+------------------
   2025 | rosters      | completed |             1854
   2024 | rosters      | completed |             1842
   2024 | player_stats | completed |            32847
   ...
```

## Performance Tips

### Faster Import

1. **Use Release Build** (already done by `make sync-full`):
   ```bash
   cargo build --release
   ```

2. **Increase Batch Size** (edit `nfl-data-pipeline/src/config.rs`):
   ```rust
   batch_size: 1000  // Default: 500
   ```

3. **Parallel Processing**:
   Already enabled via `rayon` crate

### Troubleshooting Slow Imports

- **Slow network?** Pipeline downloads large CSV files. Use wired connection if possible.
- **Database timeout?** Check DATABASE_URL points to correct server.
- **Out of memory?** Import one season at a time instead of full range.

## Error Handling

The pipeline handles errors gracefully:

- **Network errors:** Automatic retries (3 attempts)
- **Parse errors:** Skips malformed rows, continues import
- **Database errors:** Rolls back transaction, exits cleanly

**View detailed logs:**
```bash
RUST_LOG=debug ./target/release/nfl-data-pipeline --mode full --start-year 2024 --end-year 2025
```

## Common Issues

### "Failed to connect to database"

**Solution:** Check DATABASE_URL in `.env` file

```bash
# Test connection
psql $DATABASE_URL -c "SELECT 1;"
```

### "Download failed: 404"

**Solution:** Some years may not have all data sources. This is normal. Pipeline continues with available data.

### "Player ID not found"

**Solution:** This happens when stats reference a player not in rosters. Pipeline skips these gracefully.

### "Duplicate key violation"

**Solution:** Safe to ignore. Pipeline uses UPSERT, so re-importing is idempotent.

## Incremental Updates

After initial import, use update mode for ongoing maintenance:

```bash
cd nfl-data-pipeline

# Update current season only
make sync-update
```

This checks `import_progress` table and only re-imports seasons with new data.

## Schedule Automated Updates

### Cron Setup

```bash
# Edit crontab
crontab -e

# Add daily update at 3 AM
0 3 * * * cd /path/to/nfl-data-pipeline && make sync-update >> sync.log 2>&1

# Add game day updates (Sunday, Monday, Thursday)
0 13 * * 0 cd /path/to/nfl-data-pipeline && make sync-update >> sync.log 2>&1
0 20 * * 1,4 cd /path/to/nfl-data-pipeline && make sync-update >> sync.log 2>&1
```

### Always-On Worker

For continuous updates (like Heroku worker dyno):

```bash
cd nfl-data-pipeline
./run-worker.sh
```

This script:
- Runs more frequently during game days (5 min intervals)
- Less frequent during off-hours (30 min intervals)
- Never stops (good for worker dynos)

## Data Sources

All data comes from **NFLverse** (nflverse-data GitHub):

- **Rosters:** Comprehensive player data
- **Stats:** Play-by-play derived statistics
- **Next Gen Stats:** AWS tracking data
- **Schedule:** Game schedule and results

**Why NFLverse?**
- ✅ Free and open source
- ✅ Updated weekly during season
- ✅ Historical data back to 1999
- ✅ More complete than ESPN API

## After Import

Once historical data is imported:

1. **Verify data** with diagnostic:
   ```bash
   make diagnose-heroku
   ```

2. **Test API**:
   ```bash
   curl "https://nfl.wearemachina.com/api/v1/players?search=mahomes"
   ```

3. **Check stats**:
   ```bash
   curl "https://nfl.wearemachina.com/api/v1/stats/leaders?stat=passing_yards&season=2024"
   ```

## Expected Player Count

After importing 10 years (2015-2025):

- **Unique Players:** ~15,000-20,000
  - Many players appear in multiple seasons
  - Database uses UPSERT to avoid duplicates
  - Only latest team/status is kept

- **Active Players (2025):** ~1,800
- **Inactive Players:** ~13,000-18,000

**Per Season Breakdown:**
```
2025: ~1,850 players
2024: ~1,840 players
2023: ~1,820 players
2022: ~1,800 players
2021: ~1,780 players
...
```

## Full vs Incremental Import

### Full Import (First Time)
```bash
make sync-full
```
- Imports ALL years from scratch
- Takes 30-60 minutes
- Use for initial setup

### Incremental Update (Ongoing)
```bash
make sync-update
```
- Only updates current season
- Takes 1-2 minutes
- Use for daily maintenance

## Summary

**To import last 10 years of NFL data:**

```bash
cd nfl-data-pipeline
make sync-full
```

This will:
1. Download roster, stats, and Next Gen data from NFLverse
2. Process and validate ~400,000+ records
3. Import into PostgreSQL with UPSERT (no duplicates)
4. Track progress in `import_progress` table
5. Take ~30-60 minutes

**Verify success:**
```bash
make db-status
```

You should see:
- 15,000+ total players
- 2,700+ games
- 400,000+ stat records

**Then run diagnostics:**
```bash
cd ..
make diagnose-heroku
```

All top fantasy players should show ✓ FOUND.
