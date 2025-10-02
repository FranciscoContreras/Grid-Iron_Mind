# Grid Iron Mind API - Endpoint Analysis & Testing Plan

## Complete Endpoint Inventory

### Public Query Endpoints (GET)

| # | Endpoint | Handler | Sub-Routes | Status | Test Priority |
|---|----------|---------|------------|--------|---------------|
| 1 | `/api/v1/players` | PlayersHandler | List all players | ✅ | HIGH |
| 2 | `/api/v1/players/:id` | PlayersHandler | Single player details | ✅ | HIGH |
| 3 | `/api/v1/players/:id/career` | CareerHandler | Player career stats | ✅ | HIGH |
| 4 | `/api/v1/players/:id/history` | CareerHandler | Player team history | ✅ | MEDIUM |
| 5 | `/api/v1/players/:id/injuries` | InjuryHandler | Player injuries | ✅ | MEDIUM |
| 6 | `/api/v1/players/:id/advanced-stats` | AdvancedStatsHandler | Next Gen Stats | 🆕 | HIGH |
| 7 | `/api/v1/players/:id/vs-defense/:team` | DefensiveHandler | Player vs defense | ✅ | LOW |
| 8 | `/api/v1/teams` | TeamsHandler | List all teams | ✅ | HIGH |
| 9 | `/api/v1/teams/:id` | TeamsHandler | Single team details | ✅ | HIGH |
| 10 | `/api/v1/teams/:id/players` | TeamsHandler | Team roster | ✅ | MEDIUM |
| 11 | `/api/v1/games` | GamesHandler | List games | ✅ | HIGH |
| 12 | `/api/v1/games/:id` | GamesHandler | Single game details | ✅ | HIGH |
| 13 | `/api/v1/games/:id/stats` | GamesHandler | Game team stats | ✅ | MEDIUM |
| 14 | `/api/v1/games/:id/scoring-plays` | GamesHandler | Scoring timeline | 🆕 | MEDIUM |
| 15 | `/api/v1/stats/leaders` | StatsHandler | Stat leaders | ✅ | MEDIUM |
| 16 | `/api/v1/stats/game/:gameId` | StatsHandler | Game player stats | ✅ | MEDIUM |
| 17 | `/api/v1/standings` | StandingsHandler | Team standings | 🆕 | HIGH |
| 18 | `/api/v1/defense/rankings` | DefensiveHandler | Defensive rankings | ✅ | LOW |
| 19 | `/api/v1/weather/current` | WeatherHandler | Current weather | ✅ | LOW |
| 20 | `/api/v1/weather/historical` | WeatherHandler | Historical weather | ✅ | LOW |
| 21 | `/api/v1/weather/forecast` | WeatherHandler | Weather forecast | ✅ | LOW |
| 22 | `/health` | healthCheck | Health status | ✅ | HIGH |
| 23 | `/api/v1/health` | healthCheck | Health status | ✅ | HIGH |
| 24 | `/api/v1/metrics/database` | MetricsHandler | DB metrics | ✅ | MEDIUM |
| 25 | `/api/v1/metrics/health` | MetricsHandler | Health metrics | ✅ | MEDIUM |

### Admin Sync Endpoints (POST)

| # | Endpoint | Handler | Purpose | Status | Test Priority |
|---|----------|---------|---------|--------|---------------|
| 26 | `/api/v1/admin/sync/teams` | AdminHandler | Sync teams from ESPN | ✅ | HIGH |
| 27 | `/api/v1/admin/sync/rosters` | AdminHandler | Sync team rosters | ✅ | HIGH |
| 28 | `/api/v1/admin/sync/games` | AdminHandler | Sync game schedule | ✅ | HIGH |
| 29 | `/api/v1/admin/sync/full` | AdminHandler | Full initial sync | ✅ | MEDIUM |
| 30 | `/api/v1/admin/sync/historical/season` | AdminHandler | Sync historical season | ✅ | MEDIUM |
| 31 | `/api/v1/admin/sync/historical/seasons` | AdminHandler | Sync multiple seasons | ✅ | MEDIUM |
| 32 | `/api/v1/admin/sync/nflverse/stats` | AdminHandler | NFLverse stats | ✅ | MEDIUM |
| 33 | `/api/v1/admin/sync/nflverse/schedule` | AdminHandler | NFLverse schedule | ✅ | LOW |
| 34 | `/api/v1/admin/sync/nflverse/nextgen` | AdminHandler | Legacy NGS endpoint | ⚠️ | LOW |
| 35 | `/api/v1/admin/sync/nextgen-stats` | AdminHandler | Next Gen Stats sync | 🆕 | HIGH |
| 36 | `/api/v1/admin/sync/weather` | AdminHandler | Weather enrichment | ✅ | MEDIUM |
| 37 | `/api/v1/admin/sync/team-stats` | AdminHandler | Team stats sync | ✅ | MEDIUM |
| 38 | `/api/v1/admin/sync/injuries` | AdminHandler | Injury reports | ✅ | MEDIUM |
| 39 | `/api/v1/admin/sync/scoring-plays` | AdminHandler | Scoring plays sync | 🆕 | MEDIUM |
| 40 | `/api/v1/admin/sync/player-season-stats` | AdminHandler | Player career stats | 🆕 | HIGH |
| 41 | `/api/v1/admin/calc/standings` | AdminHandler | Calculate standings | 🆕 | HIGH |
| 42 | `/api/v1/admin/keys/generate` | AdminHandler | Generate API key | ✅ | LOW |

### Style/Documentation Endpoints

| # | Endpoint | Handler | Purpose | Status |
|---|----------|---------|---------|--------|
| 43 | `/api/v1/style/check` | StyleAgentHandler | Style validation | ✅ |
| 44 | `/api/v1/style/rules` | StyleAgentHandler | Style rules | ✅ |
| 45 | `/api/v1/style/example` | StyleAgentHandler | Style examples | ✅ |
| 46 | `/style-guide.html` | StyleAgentHandler | Style guide page | ✅ |
| 47 | `/api-docs.html` | Static file | API documentation | ✅ |
| 48 | `/ui-system.html` | Static file | UI system docs | ✅ |
| 49 | `/` | Static files | Dashboard | ✅ |

**Total Endpoints: 49**
- **Public Query (GET):** 25 endpoints
- **Admin Sync (POST):** 17 endpoints
- **Style/Docs:** 7 endpoints

**Legend:**
- ✅ = Existing, should work
- 🆕 = Newly implemented, needs testing
- ⚠️ = Legacy/deprecated

## Testing Plan

### Phase 1: Database Verification

**Check what data exists:**

```sql
-- Teams
SELECT COUNT(*) FROM teams;
SELECT * FROM teams LIMIT 5;

-- Players
SELECT COUNT(*) FROM players;
SELECT COUNT(*) FROM players WHERE team_id IS NOT NULL;

-- Games
SELECT COUNT(*) FROM games;
SELECT season, week, COUNT(*) FROM games GROUP BY season, week ORDER BY season DESC, week DESC;

-- Game stats
SELECT COUNT(*) FROM game_stats;
SELECT COUNT(*) FROM game_team_stats;
SELECT COUNT(*) FROM game_scoring_plays;

-- Player stats
SELECT COUNT(*) FROM player_season_stats;
SELECT COUNT(*) FROM advanced_stats;

-- Standings
SELECT COUNT(*) FROM team_standings;
SELECT season, week, COUNT(*) FROM team_standings GROUP BY season, week ORDER BY season DESC, week DESC;

-- Injuries
SELECT COUNT(*) FROM player_injuries;
```

### Phase 2: Sync Critical Data (If Missing)

**Priority Order:**

1. **Teams** (foundation)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/teams \
     -H "X-API-Key: $ADMIN_KEY"
   ```

2. **Rosters** (players)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \
     -H "X-API-Key: $ADMIN_KEY"
   ```

3. **Games** (2025 season)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/games \
     -H "X-API-Key: $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"season": 2025}'
   ```

4. **Team Stats** (for recent games)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/team-stats \
     -H "X-API-Key: $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"season": 2025, "week": 5}'
   ```

5. **Standings** (calculate)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/calc/standings \
     -H "X-API-Key: $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"season": 2025}'
   ```

6. **Player Season Stats** (2024 for data)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/player-season-stats \
     -H "X-API-Key: $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"season": 2024}'
   ```

7. **Next Gen Stats** (2024 passing)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/nextgen-stats \
     -H "X-API-Key: $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"season": 2024, "stat_type": "passing"}'
   ```

8. **Scoring Plays** (if games completed)
   ```bash
   curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/scoring-plays \
     -H "X-API-Key: $ADMIN_KEY" \
     -H "Content-Type: application/json" \
     -d '{"season": 2025, "week": 5}'
   ```

### Phase 3: Test All Public Endpoints

**High Priority Tests:**

```bash
# 1. Health check
curl https://nfl.wearemachina.com/health

# 2. Teams list
curl https://nfl.wearemachina.com/api/v1/teams

# 3. Single team
curl https://nfl.wearemachina.com/api/v1/teams/{team-id}

# 4. Team roster
curl https://nfl.wearemachina.com/api/v1/teams/{team-id}/players

# 5. Players list
curl https://nfl.wearemachina.com/api/v1/players?limit=10

# 6. Single player
curl https://nfl.wearemachina.com/api/v1/players/{player-id}

# 7. Player career stats
curl https://nfl.wearemachina.com/api/v1/players/{player-id}/career

# 8. Player advanced stats (NEW)
curl https://nfl.wearemachina.com/api/v1/players/{player-id}/advanced-stats?season=2024

# 9. Games list
curl "https://nfl.wearemachina.com/api/v1/games?season=2025&week=5"

# 10. Single game
curl https://nfl.wearemachina.com/api/v1/games/{game-id}

# 11. Game stats
curl https://nfl.wearemachina.com/api/v1/games/{game-id}/stats

# 12. Scoring plays (NEW)
curl https://nfl.wearemachina.com/api/v1/games/{game-id}/scoring-plays

# 13. Standings (NEW)
curl "https://nfl.wearemachina.com/api/v1/standings?season=2025"
```

**Medium Priority Tests:**

```bash
# 14. Stat leaders
curl https://nfl.wearemachina.com/api/v1/stats/leaders?category=passing&limit=10

# 15. Game player stats
curl https://nfl.wearemachina.com/api/v1/stats/game/{game-id}

# 16. Player injuries
curl https://nfl.wearemachina.com/api/v1/players/{player-id}/injuries

# 17. Database metrics
curl https://nfl.wearemachina.com/api/v1/metrics/database

# 18. Health metrics
curl https://nfl.wearemachina.com/api/v1/metrics/health
```

### Phase 4: Automated Testing Script

Create `test_all_endpoints.sh`:

```bash
#!/bin/bash

# Configuration
BASE_URL="https://nfl.wearemachina.com"
ADMIN_KEY="your-admin-key-here"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TOTAL=0
PASSED=0
FAILED=0

test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_status="${3:-200}"

    TOTAL=$((TOTAL + 1))
    echo -n "Testing $name... "

    response=$(curl -s -w "\n%{http_code}" "$url")
    status=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | head -n -1)

    if [ "$status" == "$expected_status" ]; then
        # Check if response has data
        if echo "$body" | jq -e '.data' > /dev/null 2>&1; then
            echo -e "${GREEN}PASS${NC} (status: $status, has data)"
            PASSED=$((PASSED + 1))
        else
            echo -e "${YELLOW}WARN${NC} (status: $status, no data field)"
            PASSED=$((PASSED + 1))
        fi
    else
        echo -e "${RED}FAIL${NC} (expected: $expected_status, got: $status)"
        echo "Response: $body" | head -c 200
        echo ""
        FAILED=$((FAILED + 1))
    fi
}

echo "=== Grid Iron Mind API Test Suite ==="
echo ""

# Health checks
echo "--- Health Checks ---"
test_endpoint "Health (root)" "$BASE_URL/health"
test_endpoint "Health (API)" "$BASE_URL/api/v1/health"

# Teams
echo ""
echo "--- Teams ---"
test_endpoint "Teams list" "$BASE_URL/api/v1/teams"
# Will get team ID from response for next tests

# Players
echo ""
echo "--- Players ---"
test_endpoint "Players list" "$BASE_URL/api/v1/players?limit=10"

# Games
echo ""
echo "--- Games ---"
test_endpoint "Games (2025)" "$BASE_URL/api/v1/games?season=2025"

# Standings
echo ""
echo "--- Standings ---"
test_endpoint "Standings 2025" "$BASE_URL/api/v1/standings?season=2025"

# Stats
echo ""
echo "--- Stats ---"
test_endpoint "Stat leaders" "$BASE_URL/api/v1/stats/leaders?category=passing"

# Metrics
echo ""
echo "--- Metrics ---"
test_endpoint "Database metrics" "$BASE_URL/api/v1/metrics/database"
test_endpoint "Health metrics" "$BASE_URL/api/v1/metrics/health"

# Summary
echo ""
echo "=== Test Summary ==="
echo "Total: $TOTAL"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Success Rate: $(( PASSED * 100 / TOTAL ))%"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed.${NC}"
    exit 1
fi
```

## Known Issues & Fixes Needed

### Issue 1: Missing Player GSIS IDs
**Affected Endpoints:**
- `/api/v1/players/:id/advanced-stats`
- Any NFLverse integration endpoints

**Fix:** Add `gsis_id` column to players table

```sql
ALTER TABLE players ADD COLUMN gsis_id VARCHAR(50);
CREATE INDEX idx_players_gsis ON players(gsis_id);
```

### Issue 2: Standings May Be Empty
**Affected Endpoints:**
- `/api/v1/standings`

**Fix:** Run standings calculation after syncing games

### Issue 3: Advanced Stats May Be Empty
**Affected Endpoints:**
- `/api/v1/players/:id/advanced-stats`

**Fix:** Sync Next Gen Stats for at least one season

### Issue 4: Scoring Plays May Be Empty
**Affected Endpoints:**
- `/api/v1/games/:id/scoring-plays`

**Fix:** Sync scoring plays for completed games

## Deployment Checklist

### Pre-Deployment

- [ ] Verify all migrations run successfully
- [ ] Check database connection string
- [ ] Verify environment variables are set:
  - `DATABASE_URL`
  - `REDIS_URL` (optional)
  - `ADMIN_API_KEY`
  - `WEATHER_API_KEY` (optional)
- [ ] Run local build test
- [ ] Review error logs

### Deployment Steps

1. **Commit all changes:**
   ```bash
   git add .
   git commit -m "Add Next Gen Stats, Play-by-Play, and comprehensive testing"
   ```

2. **Push to Heroku:**
   ```bash
   git push heroku main
   ```

3. **Run migrations:**
   ```bash
   heroku run 'psql $DATABASE_URL -f migrations/008_add_play_by_play.sql'
   ```

4. **Check logs:**
   ```bash
   heroku logs --tail
   ```

5. **Verify health:**
   ```bash
   curl https://nfl.wearemachina.com/health
   ```

### Post-Deployment

1. **Sync critical data** (in order):
   - Teams
   - Rosters
   - Games (2025 season)
   - Standings
   - Player stats (2024)

2. **Test all new endpoints:**
   - Standings
   - Advanced stats
   - Scoring plays

3. **Monitor logs for errors**

4. **Update API documentation**

## Expected Response Formats

### Success Response
```json
{
  "data": [...],
  "meta": {
    "timestamp": "2025-10-02T10:30:00Z",
    "total": 100,
    "limit": 50,
    "offset": 0
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "status": 404
  }
}
```

### Empty Data Response
```json
{
  "data": []
}
```

## Success Criteria

**All endpoints should:**
- [ ] Return 200 status code
- [ ] Have proper JSON structure
- [ ] Include `data` field
- [ ] Include `meta` field (for collections)
- [ ] Handle errors gracefully
- [ ] Log requests
- [ ] Respect rate limits

**New endpoints specifically:**
- [ ] `/api/v1/standings` returns standings for current season
- [ ] `/api/v1/players/:id/advanced-stats` returns NGS data
- [ ] `/api/v1/games/:id/scoring-plays` returns scoring timeline
- [ ] `/api/v1/admin/sync/nextgen-stats` syncs successfully
- [ ] `/api/v1/admin/calc/standings` calculates correctly

---

*Analysis completed: October 2, 2025*
*Total endpoints: 49*
*New endpoints: 7*
