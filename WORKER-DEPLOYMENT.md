# Worker Deployment Guide

## Overview

The Rust data pipeline runs continuously on Heroku as a **worker dyno**, automatically syncing NFL player data from NFLverse.

## Architecture

```
Heroku Worker Dyno
    └── run-worker-with-init.sh
        ├── First Run: Import 2025 & 2024 seasons (one-time)
        └── Continuous: Update mode based on game schedule
            ├── Game Days (Sun/Mon/Thu): Every 5 minutes
            └── Off Days: Every 30 minutes
```

## Worker Dyno Status

### Check Worker Status
```bash
heroku ps --app grid-iron-mind
```

### View Worker Logs
```bash
# Live logs
heroku logs --tail --dyno worker --app grid-iron-mind

# Recent logs
heroku logs --dyno worker --app grid-iron-mind --num 200
```

### Scale Worker
```bash
# Start worker
heroku ps:scale worker=1 --app grid-iron-mind

# Stop worker
heroku ps:scale worker=0 --app grid-iron-mind

# Restart worker
heroku ps:restart worker --app grid-iron-mind
```

## How It Works

### First Run (Initial Import)
When the worker starts for the first time:

1. **Checks initialization flag** (`/tmp/nfl-pipeline-initialized`)
2. **Imports 2025 season:**
   - All active players (~1,800)
   - Player stats
   - Next Gen Stats
3. **Imports 2024 season** (for context)
4. **Creates flag file** to prevent re-import
5. **Starts continuous update loop**

**Expected time:** 5-10 minutes

### Continuous Updates
After initialization, the worker runs in a loop:

**Game Days:**
- **Sunday 1PM-11PM:** Update every 5 minutes
- **Monday 8PM-11PM:** Update every 5 minutes (MNF)
- **Thursday 8PM-11PM:** Update every 5 minutes (TNF)

**Off Hours:**
- All other times: Update every 30 minutes

## Worker Script

The worker uses: `nfl-data-pipeline/run-worker-with-init.sh`

**Key Features:**
- ✅ Automatic initial import on first run
- ✅ Game day detection (frequent updates)
- ✅ Error handling (continues on failure)
- ✅ Persistent initialization flag

## Deployment

### 1. Commit Worker Changes
```bash
git add Procfile nfl-data-pipeline/run-worker-with-init.sh
git commit -m "Add worker with initial import"
git push heroku main
```

### 2. Restart Worker to Trigger Init
```bash
# Restart to run new script
heroku ps:restart worker --app grid-iron-mind

# Watch it import
heroku logs --tail --dyno worker --app grid-iron-mind
```

### 3. Verify Data Import
```bash
# Check player count (should be 1800+)
heroku pg:psql --app grid-iron-mind -c "SELECT COUNT(*) FROM players WHERE status = 'active';"

# Run diagnostic
make diagnose-heroku
```

## Monitoring

### Expected Log Output

**Initial Import:**
```
🏈 NFL Data Pipeline Worker Starting
📅 Date: Fri Oct  3 18:30:00 UTC 2025
🚀 FIRST RUN: Importing 2025 season rosters...
  [1/4] Importing rosters for 2025...
  ✅ Rosters: 1854 players
  ⏭️  Schedule: Skipping (use Go importer with ESPN API)
  [3/4] Importing player stats for 2025...
  ✅ Player Stats: 34521 records
  [4/4] Importing NGS passing for 2025...
  ✅ NGS Passing: 492 records
✅ 2025 season import complete!
📚 Importing 2024 season for context...
✅ Initial data population complete!

📅 Day: 5, Hour: 18
🔄 Starting continuous update loop...
📊 Regular day - Running update mode
⏱️  Sleeping 30 minutes (off hours)...
```

**Continuous Updates:**
```
📊 Regular day - Running update mode
  ✅ Updated 47 player records
  ✅ Updated 12 game records
⏱️  Sleeping 30 minutes (off hours)...
```

**Game Day:**
```
🔴 SUNDAY GAME DAY - Running update mode (frequent)
  ✅ Updated 183 player stats
  ✅ Updated 16 games
⏱️  Sleeping 5 minutes (game day)...
```

## Troubleshooting

### Worker Not Starting

**Check status:**
```bash
heroku ps --app grid-iron-mind
```

**If stopped, scale up:**
```bash
heroku ps:scale worker=1 --app grid-iron-mind
```

### Worker Crashes

**Check recent crashes:**
```bash
heroku logs --dyno worker --app grid-iron-mind --num 500 | grep -i error
```

**Common issues:**
- Database connection timeout → Check `DATABASE_URL`
- Pipeline binary missing → Redeploy
- Memory limit → Upgrade dyno type

### Initial Import Stuck

**Restart worker:**
```bash
# Remove init flag (forces re-import)
heroku run "rm /tmp/nfl-pipeline-initialized" --app grid-iron-mind

# Restart worker
heroku ps:restart worker --app grid-iron-mind
```

### No Players Imported

**Verify import ran:**
```bash
heroku logs --dyno worker --app grid-iron-mind | grep "2025 season import"
```

**Manual import:**
```bash
heroku run "./nfl-data-pipeline/target/release/nfl-data-pipeline --mode year --year 2025" --app grid-iron-mind
```

## Performance & Costs

### Dyno Usage

**Worker Dyno (Basic):**
- **Cost:** ~$7/month (basic dyno)
- **Memory:** 512MB
- **Hours:** 24/7 (always on)

**Resource Usage:**
- Idle: ~50MB RAM
- Import: ~200MB RAM
- Update: ~100MB RAM

### Optimization

To reduce costs:

1. **Scale down during offseason:**
   ```bash
   heroku ps:scale worker=0 --app grid-iron-mind
   ```

2. **Use scheduled tasks instead** (free with Heroku Scheduler addon):
   ```bash
   heroku addons:create scheduler:standard --app grid-iron-mind
   ```

3. **Manual imports only** (no worker dyno):
   ```bash
   # Run once daily via cron or CI/CD
   heroku run "./nfl-data-pipeline/target/release/nfl-data-pipeline --mode update" --app grid-iron-mind
   ```

## Manual Operations

### Force Full Re-Import

```bash
# Remove init flag
heroku run "rm /tmp/nfl-pipeline-initialized" --app grid-iron-mind

# Restart worker (will re-import)
heroku ps:restart worker --app grid-iron-mind
```

### Import Specific Year

```bash
heroku run "./nfl-data-pipeline/target/release/nfl-data-pipeline --mode year --year 2023" --app grid-iron-mind
```

### Import Historical Range (2015-2025)

```bash
# WARNING: Takes 30-60 minutes, may timeout
heroku run:detached "./nfl-data-pipeline/target/release/nfl-data-pipeline --mode full --start-year 2015 --end-year 2025" --app grid-iron-mind
```

## Verification

After worker has been running:

### 1. Check Player Count
```bash
heroku pg:psql --app grid-iron-mind -c "
SELECT
    COUNT(*) as total,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active
FROM players;
"
```

**Expected:** 1800+ active players

### 2. Run Diagnostic
```bash
make diagnose-heroku
```

**Expected:** 0 missing players

### 3. Test API
```bash
curl "https://nfl.wearemachina.com/api/v1/players?search=barkley"
```

**Expected:** Saquon Barkley returned

## Summary

**Worker is now:**
- ✅ Running continuously on Heroku
- ✅ Importing initial data on first run (2024 & 2025)
- ✅ Auto-updating based on game schedule
- ✅ Handling errors gracefully
- ✅ Syncing to production database

**To deploy:**
```bash
git add -A
git commit -m "Deploy worker with auto-init"
git push heroku main
heroku ps:restart worker
```

**To monitor:**
```bash
heroku logs --tail --dyno worker
```
