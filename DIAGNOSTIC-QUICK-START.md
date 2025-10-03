# Player Diagnostic - Quick Start

## Problem: Players Are Missing

Example: Saquon Barkley doesn't show up in `/api/v1/players?search=barkley`

## Solution: 3-Step Fix

### Step 1: Diagnose (30 seconds)

```bash
make diagnose-heroku
```

**Expected Output:**
```
=== Top Fantasy Players Check ===
 Saquon Barkley   | ✗ MISSING
 Lamar Jackson    | ✓ FOUND: L. Jackson
 ...

=== Missing Players Summary ===
 total_checked | found_count | missing_count
---------------+-------------+---------------
            30 |          28 |             2
```

### Step 2: Fix (2-5 minutes)

```bash
# Get your API key
heroku config:get API_KEY --app grid-iron-mind

# Sync rosters
curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
  -H "X-API-Key: YOUR_API_KEY_HERE"
```

**Wait for:** `{"message": "Rosters sync started in background", "status": "processing"}`

### Step 3: Verify (30 seconds)

```bash
# Re-run diagnostic
make diagnose-heroku

# Should show: missing_count | 0
```

## Alternative: Use Sync Tool

```bash
# Sync everything (recommended for first-time setup)
make sync-full

# Daily update (faster, for ongoing maintenance)
make sync-update
```

## Common Issues

### Issue: "Heroku CLI not found"

**Fix:**
```bash
# macOS
brew tap heroku/brew && brew install heroku

# Then login
heroku login
```

### Issue: "psql: command not found"

**Fix:**
```bash
# macOS
brew install postgresql

# Ubuntu/Debian
sudo apt-get install postgresql-client
```

### Issue: "Permission denied: API key required"

**Fix:**
```bash
# Get your API key from Heroku
heroku config:get API_KEY --app grid-iron-mind

# Use it in the request
curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
  -H "X-API-Key: YOUR_KEY"
```

## Files Reference

| File | Purpose |
|------|---------|
| `scripts/diagnose-missing-players.sql` | SQL diagnostic script |
| `scripts/heroku-diagnose.sh` | Heroku diagnostic runner |
| `DIAGNOSTICS.md` | Complete diagnostic guide |
| `PLAYER-DIAGNOSTICS-SUMMARY.md` | Implementation details |

## Commands Reference

| Command | Purpose | Time |
|---------|---------|------|
| `make diagnose-heroku` | Check Heroku database | 5s |
| `make diagnose-players` | Check local database | 2s |
| `make sync-update` | Update current data | 2-5m |
| `make sync-full` | Full season sync | 30-60m |

## Expected Player Counts

- **Total:** 1,800 - 2,000 players
- **Active:** 1,600 - 1,700 players
- **Per Team:** 50-60 players
- **Per Position:**
  - QB: ~100
  - RB: ~150
  - WR: ~200
  - TE: ~100

## Next Steps

After fixing missing players:

1. **Set up automated sync:**
   ```bash
   make install-cron
   ```

2. **Monitor health:**
   ```bash
   # Add to daily cron
   0 4 * * * cd /path/to/gridironmind && make diagnose-heroku
   ```

3. **Use the API:**
   ```bash
   curl "https://nfl.wearemachina.com/api/v1/players?search=barkley"
   # Should return Saquon Barkley
   ```

## Support

- 📖 Full guide: [DIAGNOSTICS.md](DIAGNOSTICS.md)
- 🏗️ Implementation: [PLAYER-DIAGNOSTICS-SUMMARY.md](PLAYER-DIAGNOSTICS-SUMMARY.md)
- 📚 Project docs: [README.md](README.md)
- 🧠 Architecture: [CLAUDE.md](CLAUDE.md)
