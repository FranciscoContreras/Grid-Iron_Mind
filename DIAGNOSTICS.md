# Player Diagnostics Guide

## Overview

This guide helps diagnose and fix missing player data in the Grid Iron Mind database. Use these tools to identify which top fantasy players are missing and sync them from ESPN.

## Problem

The database may be missing high-profile players like:
- Saquon Barkley (PHI, RB)
- Other top fantasy performers

This typically happens when:
1. ESPN roster sync hasn't run recently
2. Player data changed since last sync
3. Name mismatches between ESPN and our database

## Diagnostic Tools

### 1. SQL Diagnostic Script (Recommended - No Go Required)

**Location:** `scripts/diagnose-missing-players.sql`

**What it checks:**
- Total player count (active vs inactive)
- Players by position
- Top 30 fantasy players for 2024-2025 season
- Team roster completeness
- Name mismatch detection

**Run locally:**
```bash
# If DATABASE_URL is set
make diagnose-players
```

**Run on Heroku:**
```bash
# Connect to Heroku database and run diagnostic
heroku pg:psql --app grid-iron-mind -f scripts/diagnose-missing-players.sql

# Or manually copy/paste SQL
heroku pg:psql --app grid-iron-mind < scripts/diagnose-missing-players.sql
```

**Sample Output:**
```
=== PLAYER DATABASE DIAGNOSTIC ===

=== Total Player Count ===
 total_players | active_players | inactive_players
---------------+----------------+------------------
          1842 |           1654 |              188

=== Top Fantasy Players Check ===
  expected_name   |         status          | position | team
------------------+-------------------------+----------+------
 Saquon Barkley   | ✗ MISSING              |          |
 Lamar Jackson    | ✓ FOUND: L. Jackson    | QB       | BAL
 Josh Allen       | ✓ FOUND: J. Allen      | QB       | BUF
 ...
```

### 2. Go Diagnostic Tool (Advanced)

**Location:** `cmd/diagnose-players/main.go`

**Features:**
- Queries ESPN API directly to find missing players
- Shows full ESPN player data for missing players
- Exits with error code if players missing

**Run (requires Go installed):**
```bash
make diagnose-players-go
```

## Fix Options

### Option 1: Rust Pipeline Import (Recommended - Most Comprehensive)

The **best solution** is to use the Rust data pipeline which imports **ALL active players** from NFLverse (more comprehensive than ESPN):

```bash
cd nfl-data-pipeline

# Import 2025 season (recommended)
make sync-year YEAR=2025

# Or full historical import (2010-2025)
make sync-full
```

**Why Rust pipeline?**
- ✅ NFLverse has comprehensive roster data (1800+ players)
- ✅ Includes all active players like Saquon Barkley
- ✅ Fast parallel processing
- ✅ Historical data back to 2010

**Expected Time:**
- Single season: 2-3 minutes
- Full import: 30-60 minutes

### Option 2: ESPN API Sync (Alternative, May Have Gaps)

Sync team rosters from ESPN API:

**Via API:**
```bash
# Trigger roster sync via API
curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
  -H "X-API-Key: YOUR_API_KEY"
```

**Expected Time:** 2-5 minutes

⚠️ **Note:** ESPN API may not have all players. Use Rust pipeline for complete coverage.

### Option 2: Check Specific Player via API

Use the API to search for a player by name:

```bash
# Search for Saquon Barkley
curl "https://nfl.wearemachina.com/api/v1/players?search=barkley"

# If not found, trigger roster sync
curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
  -H "X-API-Key: YOUR_API_KEY"
```

### Option 3: Manual ESPN Data Fetch

Query ESPN API directly to see if player exists:

```bash
# Fetch Philadelphia Eagles roster (team ID: 21)
curl "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams/21?enable=roster"
```

## Common Issues & Solutions

### Issue: "Player not found in database"

**Diagnosis:**
```bash
# Run diagnostic
make diagnose-players
# or
heroku pg:psql --app grid-iron-mind -f scripts/diagnose-missing-players.sql
```

**Solution:**
```bash
# Sync rosters
curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
  -H "X-API-Key: YOUR_API_KEY"
```

### Issue: "Player exists with different name"

**Example:** Database has "S. Barkley" but search for "Saquon Barkley" fails

**Diagnosis:**
```sql
-- Check for name variations
SELECT name, position, status
FROM players
WHERE name ILIKE '%barkley%';
```

**Solution:**
- Update player search to use fuzzy matching
- Re-sync rosters to get full names
- Add player name aliases table

### Issue: "Team has too few players"

**Diagnosis:**
```sql
-- Check team roster sizes
SELECT t.name, t.abbreviation, COUNT(p.id) as player_count
FROM teams t
LEFT JOIN players p ON p.team_id = t.id
GROUP BY t.id, t.name, t.abbreviation
ORDER BY player_count ASC;
```

**Solution:**
```bash
# Sync specific team roster (e.g., Philadelphia Eagles)
curl -X POST "https://nfl.wearemachina.com/api/v1/admin/sync/rosters" \
  -H "X-API-Key: YOUR_API_KEY"
```

## Monitoring & Automation

### Schedule Regular Roster Syncs

Add to cron (via `scripts/crontab-2025.txt`):

```cron
# Sync rosters daily at 3 AM
0 3 * * * cd /path/to/gridironmind && make sync-update >> logs/sync.log 2>&1

# Sync rosters before game day (Thursday, Sunday, Monday)
0 6 * * 0,1,4 cd /path/to/gridironmind && make sync-update >> logs/sync.log 2>&1
```

### API Health Check

```bash
# Check database metrics
curl "https://nfl.wearemachina.com/api/v1/metrics/database"

# Response should show:
# {
#   "data": {
#     "total_players": 1800+,
#     "active_players": 1600+,
#     ...
#   }
# }
```

## Quick Reference

| Task | Command |
|------|---------|
| Run diagnostic (SQL) | `make diagnose-players` |
| Run diagnostic on Heroku | `heroku pg:psql --app grid-iron-mind -f scripts/diagnose-missing-players.sql` |
| Sync all rosters | `curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters -H "X-API-Key: KEY"` |
| Search for player | `curl "https://nfl.wearemachina.com/api/v1/players?search=NAME"` |
| Check team roster | `SELECT COUNT(*) FROM players WHERE team_id = (SELECT id FROM teams WHERE abbreviation = 'PHI')` |
| View recent updates | `SELECT COUNT(*) FROM players WHERE updated_at > NOW() - INTERVAL '1 day'` |

## Expected Player Counts

- **Total Players:** ~1,800 - 2,000
- **Active Players:** ~1,600 - 1,700
- **Per Team:** ~50-60 players (including practice squad)
- **Key Positions:**
  - QB: ~100
  - RB: ~150
  - WR: ~200
  - TE: ~100
  - Offensive Line: ~300
  - Defensive: ~800

## Troubleshooting

### Diagnostic shows many missing players

1. Check last sync time:
   ```sql
   SELECT MAX(updated_at) FROM players;
   ```

2. Trigger full roster sync:
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
     -H "X-API-Key: YOUR_API_KEY"
   ```

3. Wait 2-5 minutes, then re-run diagnostic

### ESPN API returns 429 (Rate Limited)

- Wait 60 seconds between requests
- Use sync tool instead of direct API calls
- Schedule syncs during off-peak hours (2-6 AM)

### Player exists in ESPN but not in database

1. Check ESPN team ID mapping:
   ```sql
   SELECT * FROM teams WHERE abbreviation = 'PHI';
   ```

2. Manually trigger sync for that team's roster

3. Check logs for import errors:
   ```bash
   heroku logs --tail --app grid-iron-mind | grep -i "player"
   ```

## Support

If issues persist:
1. Check `logs/diagnose-players.log` for details
2. Review `logs/sync-2025.log` for sync errors
3. Verify ESPN API is responding: `curl "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams"`
