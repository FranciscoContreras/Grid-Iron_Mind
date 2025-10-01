# Auto-Fetch System - Deployment Checklist

## Pre-Deployment

### 1. Review Code Changes
- ✅ `internal/utils/season.go` - Season detection utility
- ✅ `internal/autofetch/orchestrator.go` - Auto-fetch orchestrator
- ✅ `internal/handlers/games.go` - Modified games handler
- ✅ `CLAUDE.md` - Updated documentation
- ✅ `AUTO-FETCH-IMPLEMENTATION.md` - Implementation guide

### 2. Verify Database Schema
```bash
# Connect to database
psql $DATABASE_URL

# Verify tables exist
\dt

# Check games table structure
\d games

# Verify we have teams (should be 32)
SELECT COUNT(*) FROM teams;
```

### 3. Local Testing (Optional)
```bash
# Build server
go build -o bin/server cmd/server/main.go

# Run locally
./bin/server

# Test auto-fetch
curl "http://localhost:8080/api/v1/games?season=2025&week=5"
```

## Deployment to Heroku

### 1. Commit Changes
```bash
git status
git add internal/utils/season.go
git add internal/autofetch/orchestrator.go
git add internal/handlers/games.go
git add CLAUDE.md
git add AUTO-FETCH-IMPLEMENTATION.md
git add DEPLOYMENT-CHECKLIST.md
git commit -m "Add intelligent auto-fetch system for self-healing data layer

- Auto-fetches scheduled games when API returns empty
- Season/week detection utility
- Orchestrator with deduplication and cascade fetching
- Integrated into games handler
- Solves Priority #1: Week 5 scheduled games now return data"
```

### 2. Deploy to Heroku
```bash
# Push to Heroku
git push heroku main

# Wait for build to complete
# Should see: "Verifying deploy... done."
```

### 3. Monitor Deployment
```bash
# Check build logs
heroku logs --tail

# Verify server started
heroku ps

# Expected output:
# web.1: up 2025/10/01 12:00:00 -0700 (~ 1m ago)
```

## Post-Deployment Testing

### 1. Health Check
```bash
# Verify API is up
curl https://nfl.wearemachina.com/api/v1/health

# Expected:
# {
#   "data": {
#     "status": "healthy",
#     "service": "Grid Iron Mind API",
#     "version": "1.0.0"
#   }
# }
```

### 2. Test Auto-Fetch for Week 5
```bash
# This is THE critical test - Priority #1
curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=5"

# Expected:
# - Should return 16 games (scheduled matchups)
# - Status: "scheduled"
# - home_score and away_score: null
# - game_date: future dates
# - X-Auto-Fetched: true (in headers)
```

### 3. Test Auto-Fetch for Multiple Weeks
```bash
# Week 6
curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=6"

# Week 7
curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=7"

# Week 18 (last week of regular season)
curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=18"
```

### 4. Verify Headers
```bash
# Check response headers
curl -I "https://nfl.wearemachina.com/api/v1/games?season=2025&week=5"

# Look for:
# X-Auto-Fetched: true (first request)
# X-Cache: MISS
# Content-Type: application/json
```

### 5. Test Concurrent Requests (Deduplication)
```bash
# Run 5 concurrent requests
for i in {1..5}; do
  curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=10" &
done
wait

# Check logs - should only see ONE fetch operation
heroku logs --tail | grep "\[AUTO-FETCH\]"
```

### 6. Test Already-Fetched Data (Fast Path)
```bash
# Request same data again - should be fast (no auto-fetch)
time curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=5"

# Expected:
# - Response time: < 200ms
# - NO X-Auto-Fetched header (data already in DB)
# - Same 16 games returned
```

### 7. Verify Fantasy Football App Works
```bash
# This is what the Fantasy Grid app needs
curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=5" | jq '.data[] | {home_team_id, away_team_id, game_date, status}'

# Should show all Week 5 matchups with:
# - Valid team IDs
# - Future game dates
# - status: "scheduled"
```

## Monitoring

### 1. Check Logs for Auto-Fetch Activity
```bash
# Watch live logs
heroku logs --tail

# Filter for auto-fetch
heroku logs --tail | grep "\[AUTO-FETCH\]"

# Expected log sequence:
# [AUTO-FETCH] No games found for season 2025 week 5, attempting auto-fetch
# [AUTO-FETCH] Fetching games for season 2025 week 5
# [AUTO-FETCH] Only 32 teams found, fetching all teams (if needed)
# [AUTO-FETCH] Successfully fetched and returned 16 games
```

### 2. Database Verification
```bash
# Connect to Heroku database
heroku pg:psql

# Check games were inserted
SELECT season, week, COUNT(*) as game_count
FROM games
WHERE season = 2025
GROUP BY season, week
ORDER BY week;

# Expected: Each week should have ~16 games
# Week 5 | 16
# Week 6 | 16
# etc.
```

### 3. Performance Metrics
```bash
# Check response times
heroku logs --tail | grep "listGames"

# First request (with auto-fetch): 2-10 seconds
# Subsequent requests: < 200ms
```

## Rollback Plan (If Needed)

If something goes wrong:

```bash
# View recent releases
heroku releases

# Rollback to previous version
heroku rollback

# Or rollback to specific version
heroku rollback v123
```

## Success Criteria

✅ **Week 5 Games:** `/api/v1/games?season=2025&week=5` returns 16 scheduled games
✅ **Current Week Detection:** System knows current NFL week
✅ **Fast Subsequent Requests:** Second request is < 200ms
✅ **Deduplication Works:** Concurrent requests don't duplicate fetches
✅ **Graceful Failures:** Failed fetches don't break API
✅ **Clear Logging:** Auto-fetch operations visible in logs
✅ **Fantasy App Works:** Client can detect current week and show matchups

## Next Steps After Deployment

### Immediate (Today)
1. ✅ Deploy auto-fetch system
2. ✅ Test Week 5 scheduled games
3. ✅ Verify Fantasy Football app works
4. ✅ Monitor logs for errors

### Short-term (This Week)
1. Add defensive stats endpoints
2. Add bye week information
3. Cache invalidation after auto-fetch
4. Add metrics tracking

### Medium-term (Next Week)
1. Async background fetching
2. Admin UI for manual triggers
3. Fetch queue system
4. Performance optimization

## Support & Debugging

### Common Issues

**Issue: "fetch already in progress"**
- **Cause:** Concurrent requests triggered multiple fetches
- **Solution:** This is expected - deduplication is working. Wait and retry.

**Issue: "auto-fetch not allowed for season X week Y"**
- **Cause:** Trying to fetch historical data (> 1 year old)
- **Solution:** Use sync CLI tool for bulk historical loads

**Issue: Empty results even after auto-fetch**
- **Cause:** ESPN API returned no data or fetch failed
- **Solution:** Check logs for ESPN API errors. Verify ESPN API is accessible.

**Issue: Slow response times**
- **Cause:** Auto-fetch happening on every request
- **Solution:** Check database - games might not be persisting. Verify DB connection.

### Debug Commands

```bash
# Check server status
heroku ps

# View full logs
heroku logs --tail --num 1000

# Check database
heroku pg:info
heroku pg:psql

# Restart server (if needed)
heroku restart

# View environment variables
heroku config
```

## Contact & Escalation

If issues persist:
1. Check GitHub issues: https://github.com/francisco/gridironmind/issues
2. Review CLAUDE.md documentation
3. Check AUTO-FETCH-IMPLEMENTATION.md for details
4. Monitor Heroku status: https://status.heroku.com

## Final Verification

Before marking as complete, verify:
- [ ] Week 5 games return scheduled matchups
- [ ] Fantasy Football app can detect current week
- [ ] No errors in Heroku logs
- [ ] Response times acceptable (< 10s first, < 200ms subsequent)
- [ ] Database has games for weeks 5-18
- [ ] Auto-fetch only runs once per resource

---

**Deployment Date:** _____________
**Deployed By:** _____________
**Verification Status:** [ ] PASSED  [ ] FAILED
**Notes:** _______________________________
