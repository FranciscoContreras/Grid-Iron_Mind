# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Grid Iron Mind is a high-performance NFL data API providing real-time and historical statistics, player information, team data, game schedules, injury reports, and AI-powered insights. The system serves comprehensive NFL data with sub-200ms response times for database queries and aggressive caching for optimal performance.

**Production API:** `https://nfl.wearemachina.com/api/v1`

## Tech Stack

- **Backend:** Go 1.21+
- **Database:** PostgreSQL with pgx/v5 driver
- **Cache:** Redis (go-redis/v8)
- **Deployment:** Heroku (primary), supports Vercel serverless
- **AI:** Claude 3.5 Sonnet via Anthropic API
- **Data Sources:** ESPN NFL APIs, NFLverse data

## Architecture

### Project Structure

```
cmd/
  server/           # Main API server entrypoint
  sync2025/         # 2025 season data sync CLI tool
internal/
  cache/            # Redis caching layer with TTL management
  config/           # Configuration loading
  db/               # Database connection pool and queries
  espn/             # ESPN API client
  nflverse/         # NFLverse data client
  ingestion/        # Data ingestion service
  handlers/         # HTTP request handlers (players, teams, games, stats, AI, admin)
  middleware/       # CORS, auth, rate limiting, error handling
  models/           # Data models (Player, Team, Game, Stats, Injury)
  ai/               # Claude API integration
  weather/          # Weather API client
pkg/
  response/         # JSON response formatting utilities
  validation/       # Input validation
dashboard/          # Web dashboard (HTML/JS/CSS)
migrations/         # SQL migration files
schema.sql          # Complete database schema
```

### Key Architectural Principles

1. **Monolithic HTTP Server:** Single Go binary runs HTTP server with all routes (cmd/server/main.go)
2. **Layered Architecture:** Handlers → Queries/Services → Database (thin handlers, fat queries)
3. **Middleware Stack:** CORS → LogRequest → RecoverPanic → RateLimit/Auth → Handler
4. **Dual Deployment:** Heroku (primary) and Vercel serverless (supported)
5. **Connection Pooling:** Shared pgxpool.Pool for all database operations
6. **Aggressive Caching:** Redis with endpoint-specific TTLs (5min-1hr)

## Common Commands

### Development

```bash
# Install dependencies
go mod download

# Run local server (with auto-reload on port 8080)
go run cmd/server/main.go

# Build server binary
go build -o bin/server cmd/server/main.go

# Build sync tool
go build -o bin/sync2025 cmd/sync2025/main.go

# Access local dashboard
open http://localhost:8080
```

### Database

```bash
# Connect to database
psql $DATABASE_URL

# Apply schema (includes all migrations)
psql $DATABASE_URL -f schema.sql

# Check connection
psql $DATABASE_URL -c "SELECT COUNT(*) FROM players;"
```

### Data Sync (2025 Season)

```bash
# Build sync tool
make build

# Full initial sync (30-60 min)
make sync-full

# Daily update sync (2-5 min)
make sync-update

# Live game day sync (continuous)
make sync-live

# Sync player stats only
make sync-stats

# Sync injuries only (fast)
make sync-injuries

# Install automated cron schedule
make install-cron

# View sync logs
make logs

# Check database status
make db-status
```

The sync tool (`cmd/sync2025/main.go`) supports five modes:
- **full**: Complete initial load of all 2025 data
- **update**: Refresh current week's games and stats
- **live**: Continuous updates during game days
- **stats**: Player statistics only
- **injuries**: Injury reports only

### Deployment

```bash
# Deploy to Heroku
git push heroku main

# View Heroku logs
heroku logs --tail

# Check Heroku status
heroku ps
```

### Testing

```bash
# No test files currently exist in the codebase
# When adding tests, use:
go test ./...
go test -v -run TestName ./internal/handlers/
```

## API Design

### Server Configuration (cmd/server/main.go)

- Single HTTP server with net/http ServeMux
- Graceful shutdown on SIGINT/SIGTERM
- Middleware applied via wrapper functions (applyMiddleware, applyAIMiddleware)
- Serves dashboard static files from ./dashboard at root
- Health check at /health and /api/v1/health
- Port from $PORT env var (default 8080)

### Endpoint Patterns

**Public Endpoints** (standard rate limit 100/min):
- `/api/v1/players` - List players with filters
- `/api/v1/players/:id` - Single player details
- `/api/v1/players/:id/career` - Player career stats
- `/api/v1/players/:id/history` - Player team history
- `/api/v1/players/:id/injuries` - Player injuries
- `/api/v1/teams` - List all teams
- `/api/v1/teams/:id` - Single team details
- `/api/v1/teams/:id/players` - Team roster
- `/api/v1/games` - List games with filters
- `/api/v1/games/:id` - Single game details
- `/api/v1/stats/leaders` - Stat leaders
- `/api/v1/stats/game/:gameID` - Game statistics
- `/api/v1/weather/*` - Weather endpoints

**AI Endpoints** (strict rate limit 10/min, requires API key):
- `/api/v1/ai/predict/game/:id` - Game outcome prediction
- `/api/v1/ai/predict/player/:id` - Player performance prediction
- `/api/v1/ai/insights/player/:id` - Player analysis
- `/api/v1/ai/query` - Natural language query

**Admin Endpoints** (for data ingestion):
- `/api/v1/admin/sync/teams` - Sync teams
- `/api/v1/admin/sync/rosters` - Sync rosters
- `/api/v1/admin/sync/games` - Sync games
- `/api/v1/admin/sync/full` - Full sync
- `/api/v1/admin/sync/historical/*` - Historical data sync
- `/api/v1/admin/sync/nflverse/*` - NFLverse enrichment
- `/api/v1/admin/sync/weather` - Weather enrichment
- `/api/v1/admin/sync/team-stats` - Team statistics
- `/api/v1/admin/sync/injuries` - Injury reports
- `/api/v1/admin/keys/generate` - Generate API keys

### Response Format (pkg/response)

All responses use consistent JSON format via pkg/response helpers:

**Success (single resource):**
```go
response.Success(w, data)
// Returns: {"data": {...}, "meta": {"timestamp": "..."}}
```

**Success (collection with pagination):**
```go
response.SuccessWithPagination(w, data, total, limit, offset)
// Returns: {"data": [...], "meta": {"total": 100, "limit": 50, "offset": 0, "timestamp": "..."}}
```

**Error:**
```go
response.Error(w, 404, "NOT_FOUND", "Player not found")
// Returns: {"error": {"code": "NOT_FOUND", "message": "...", "status": 404}}
```

**Helper functions:**
- `response.NotFound(w, "Player")` - 404 error
- `response.BadRequest(w, "Invalid ID")` - 400 error
- `response.InternalError(w, "Database error")` - 500 error
- `response.Unauthorized(w, "Invalid API key")` - 401 error
- `response.TooManyRequests(w, retryAfter)` - 429 error

## Database Architecture

### Connection Management (internal/db/postgres.go)

- Global pgxpool.Pool managed by db package
- `db.Connect(ctx, config)` - Initialize pool
- `db.GetPool()` - Access pool from handlers/queries
- `db.Close()` - Close pool on shutdown
- `db.HealthCheck(ctx)` - Verify connectivity

Pool configuration:
- MaxConns/MinConns from env (default 25/5)
- MaxConnLifetime: 1 hour
- MaxConnIdleTime: 30 minutes
- HealthCheckPeriod: 1 minute

### Query Patterns

Queries are organized in dedicated files:
- `internal/db/queries.go` - Player queries (PlayerQueries struct)
- `internal/db/game_queries.go` - Game queries
- `internal/db/career_queries.go` - Career stats queries
- `internal/db/injury_queries.go` - Injury queries

**Query struct pattern:**
```go
type PlayerQueries struct{}

func (q *PlayerQueries) GetByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
    pool := GetPool()
    // Use pool.QueryRow or pool.Query
}
```

**Handler pattern:**
```go
type PlayersHandler struct {
    queries *db.PlayerQueries
}

func (h *PlayersHandler) HandlePlayers(w http.ResponseWriter, r *http.Request) {
    // Parse path to route to correct method
    // Call queries methods
    // Use response helpers
}
```

### Schema Overview (schema.sql)

Core tables:
- **teams** - 32 NFL teams with stadium metadata (lat/lon, surface, capacity)
- **players** - Player profiles with team_id FK, draft info, physical stats
- **games** - Game schedule with nfl_game_id, scores, season/week
- **game_stats** - Per-game player statistics (composite key: player_id + game_id)
- **player_season_stats** - Aggregated season stats (added in migration 002)
- **player_team_history** - Player team changes over time
- **player_injuries** - Current injury reports with status/timeline
- **game_team_stats** - Team statistics per game
- **game_weather** - Weather conditions per game
- **predictions** - AI predictions with confidence scores
- **ai_analysis** - AI analysis results with expiration

All tables use UUID primary keys. Indexes on foreign keys, dates, season/week, and frequently queried columns.

## Caching Strategy (internal/cache)

Redis-based caching with TTL management:

**Cache Keys (internal/cache/keys.go):**
```go
cache.PlayerKey(id)           // "player:{id}"
cache.PlayersListKey(filters) // "players:list:{filters_hash}"
cache.TeamKey(id)             // "team:{id}"
// etc.
```

**TTLs by endpoint:**
- Teams: 1 hour (infrequent changes)
- Players: 15 minutes
- Games: 5 minutes (live updates)
- Stats: 5 minutes
- Stats Leaders: 10 minutes
- AI Predictions (Game): 15 minutes
- AI Predictions (Player): 30 minutes
- AI Analysis: 1 hour

**Usage pattern:**
```go
// Try cache first
cached, err := cache.Get(ctx, cacheKey)
if err == nil && cached != "" {
    w.Header().Set("X-Cache", "HIT")
    w.Write([]byte(cached))
    return
}

// Query database
data := queryDatabase()

// Cache result
cache.Set(ctx, cacheKey, response.ToJSON(data), ttl)
w.Header().Set("X-Cache", "MISS")
```

Cache is optional - if Redis unavailable, API continues without caching.

## Data Ingestion

### ESPN API Client (internal/espn/client.go)

Primary data source for live NFL data:
- FetchTeams() - All 32 teams
- FetchTeamRoster(teamID) - Current roster
- FetchScoreboard() - Live scores and status
- FetchPlayerDetails(athleteID) - Player profile

**Key ESPN endpoints used:**
- `site.api.espn.com/apis/site/v2/sports/football/nfl/teams`
- `site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard`
- `sports.core.api.espn.com/v2/sports/football/leagues/nfl/athletes`

### NFLverse Client (internal/nflverse/client.go)

Enhanced statistics and historical data:
- CSV parsing from nflverse-data GitHub releases
- Player season stats, advanced metrics, Next Gen Stats
- Schedule enrichment, weather data

### Ingestion Service (internal/ingestion/service.go)

Orchestrates data sync operations:
- SyncTeams() - Load team data
- SyncRosters() - Load player rosters
- SyncGames() - Load game schedule and scores
- SyncTeamStats() - Load team statistics
- SyncInjuries() - Load injury reports
- SyncPlayerStats() - Load player statistics from NFLverse

**Access via admin endpoints or sync CLI tool.**

### 2025 Season Sync Tool (cmd/sync2025/main.go)

Standalone CLI for automated data sync:
- Loads .env for DATABASE_URL
- Supports 5 modes: full, update, live, stats, injuries
- Full sync: Teams → Rosters → All 18 weeks of games → Stats → Injuries
- Update sync: Refresh current week only
- Live sync: Continuous loop during game days
- Designed for cron scheduling (see scripts/crontab-2025.txt)

## AI Integration

### Multi-Provider AI Service with Automatic Fallback

The system supports multiple AI providers with automatic failover for resilience:

**AI Service (internal/ai/service.go):**
- Manages multiple AI providers (Claude, Grok)
- Automatic fallback when primary provider fails
- Configurable primary provider (first available)
- Returns provider used in each response

**Providers:**
1. **Claude** (internal/ai/claude.go) - Claude 3.5 Sonnet
2. **Grok** (internal/ai/grok.go) - Grok Beta from xAI

**Fallback Logic:**
```go
// Service tries primary provider, falls back to secondary automatically
prediction, provider, err := aiService.PredictGameOutcome(ctx, home, away, stats...)
log.Printf("Generated using %s", provider) // "claude" or "grok"
```

### Claude API Client (internal/ai/claude.go)

- Uses Claude 3.5 Sonnet model
- Structured JSON-only responses
- Token usage tracking
- Response caching (15-60 min TTL)

### Grok API Client (internal/ai/grok.go)

- Uses Grok Beta model from xAI
- OpenAI-compatible chat completions API
- Structured JSON responses
- Zero temperature for consistent results

### AI Handlers (internal/handlers/ai.go)

- HandlePredictGame - Game outcome prediction
- HandlePredictPlayer - Player performance prediction
- HandleAnalyzePlayer - Deep player analysis
- HandleAIQuery - Natural language queries

All AI endpoints:
1. Require API key authentication
2. Apply strict rate limiting (10/min)
3. Support multiple AI providers with automatic fallback
4. Cache results aggressively
5. Return structured predictions with confidence scores
6. Include `ai_provider` field in response indicating which AI was used

## Authentication & Rate Limiting

### API Key Authentication (internal/middleware/auth.go)

- Applied only to AI endpoints via `applyAIMiddleware`
- Checks `X-API-Key` header or `Authorization: Bearer` token
- Validates against API_KEY and UNLIMITED_API_KEY env vars
- If no API keys configured, auth is bypassed (dev mode)
- Returns 401 for invalid/missing keys

### Rate Limiting (internal/middleware/ratelimit.go)

**Two tiers:**
1. **StandardRateLimit** - 100 requests/minute (public endpoints)
2. **StrictRateLimit** - 10 requests/minute (AI endpoints)

**Implementation:**
- Redis-based counters with 1-minute expiration
- Returns 429 with Retry-After header when exceeded
- Unlimited API keys bypass all limits
- Headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
- If Redis unavailable, allows all requests

## Middleware Stack (internal/middleware)

### Standard Endpoints
```go
func applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return middleware.CORS(
        middleware.LogRequest(
            middleware.RecoverPanic(
                middleware.StandardRateLimit(handler),
            ),
        ),
    )
}
```

### AI Endpoints
```go
func applyAIMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return middleware.CORS(
        middleware.LogRequest(
            middleware.RecoverPanic(
                middleware.APIKeyAuth(
                    middleware.StrictRateLimit(handler),
                ),
            ),
        ),
    )
}
```

### Available Middleware

- **CORS** (cors.go) - Allow all origins, standard headers, GET/POST/OPTIONS
- **LogRequest** (in main.go via LogRequest) - Log method, path, duration
- **RecoverPanic** (errors.go) - Catch panics, return 500 error
- **StandardRateLimit** (ratelimit.go) - 100/min limit
- **StrictRateLimit** (ratelimit.go) - 10/min limit
- **APIKeyAuth** (auth.go) - API key validation

## Environment Variables

**Required:**
- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - HTTP server port (default 8080)

**Optional:**
- `REDIS_URL` - Redis connection string (caching disabled if not set)
- `API_KEY` - Standard API key (auth disabled if not set)
- `UNLIMITED_API_KEY` - Unlimited rate limit API key
- `CLAUDE_API_KEY` - Claude API key for primary AI (optional with Grok)
- `GROK_API_KEY` - Grok API key from xAI for AI fallback/primary (optional with Claude)
- `WEATHER_API_KEY` - Weather API key for weather endpoints
- `DB_MAX_CONNS` - Max database connections (default 25)
- `DB_MIN_CONNS` - Min database connections (default 5)
- `ENVIRONMENT` - production, staging, or development

**Development mode triggers:**
- No API_KEY set → Authentication bypassed
- No REDIS_URL set → Caching disabled, rate limiting allows all
- No AI keys set → AI endpoints return 503
- One AI key set → That provider is primary (no fallback)
- Both AI keys set → First configured is primary, other is fallback

## Code Patterns

### Handler Structure

```go
// internal/handlers/players.go
type PlayersHandler struct {
    queries *db.PlayerQueries
}

func NewPlayersHandler() *PlayersHandler {
    return &PlayersHandler{
        queries: &db.PlayerQueries{},
    }
}

func (h *PlayersHandler) HandlePlayers(w http.ResponseWriter, r *http.Request) {
    // 1. Method check
    if r.Method != http.MethodGet {
        response.Error(w, 405, "METHOD_NOT_ALLOWED", "Only GET allowed")
        return
    }

    // 2. Parse path to route sub-endpoints
    path := strings.TrimPrefix(r.URL.Path, "/api/v1/players")
    if path == "" {
        h.listPlayers(w, r)
    } else if strings.HasSuffix(path, "/career") {
        // Delegate to specialized handler
    } else {
        h.getPlayer(w, r, path)
    }
}

func (h *PlayersHandler) listPlayers(w http.ResponseWriter, r *http.Request) {
    // 3. Parse and validate query params
    limit := validation.ValidateLimit(validation.ParseIntParam(query.Get("limit"), 50))

    // 4. Check cache
    cacheKey := cache.PlayersListKey(filters)
    if cached, err := cache.Get(r.Context(), cacheKey); err == nil {
        w.Header().Set("X-Cache", "HIT")
        w.Write([]byte(cached))
        return
    }

    // 5. Query database
    players, total, err := h.queries.List(r.Context(), filters)
    if err != nil {
        response.InternalError(w, "Failed to fetch players")
        return
    }

    // 6. Cache result
    cache.Set(r.Context(), cacheKey, response.ToJSON(players), 15*time.Minute)

    // 7. Return response
    w.Header().Set("X-Cache", "MISS")
    response.SuccessWithPagination(w, players, total, limit, offset)
}
```

### Database Query Pattern

```go
// internal/db/queries.go
type PlayerQueries struct{}

func (q *PlayerQueries) GetByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
    pool := GetPool()

    query := `
        SELECT id, nfl_id, name, position, team_id, jersey_number,
               height_inches, weight_pounds, birth_date, college,
               draft_year, draft_round, draft_pick, status,
               created_at, updated_at
        FROM players
        WHERE id = $1
    `

    var player models.Player
    err := pool.QueryRow(ctx, query, id).Scan(
        &player.ID, &player.NFLID, &player.Name, &player.Position,
        &player.TeamID, &player.JerseyNumber, &player.HeightInches,
        &player.WeightPounds, &player.BirthDate, &player.College,
        &player.DraftYear, &player.DraftRound, &player.DraftPick,
        &player.Status, &player.CreatedAt, &player.UpdatedAt,
    )

    if err == pgx.ErrNoRows {
        return nil, ErrPlayerNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }

    return &player, nil
}
```

### Model Definition Pattern

```go
// internal/models/player.go
type Player struct {
    ID            uuid.UUID  `json:"id"`
    NFLID         int        `json:"nfl_id"`
    Name          string     `json:"name"`
    Position      string     `json:"position"`
    TeamID        *uuid.UUID `json:"team_id,omitempty"`
    JerseyNumber  *int       `json:"jersey_number,omitempty"`
    HeightInches  *int       `json:"height_inches,omitempty"`
    WeightPounds  *int       `json:"weight_pounds,omitempty"`
    BirthDate     *time.Time `json:"birth_date,omitempty"`
    College       *string    `json:"college,omitempty"`
    DraftYear     *int       `json:"draft_year,omitempty"`
    DraftRound    *int       `json:"draft_round,omitempty"`
    DraftPick     *int       `json:"draft_pick,omitempty"`
    Status        string     `json:"status"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}
```

## Dashboard (dashboard/)

Static web dashboard served at root path:
- Single-page app (index.html, app.js, styles.css)
- Features: Browse players, view teams, API endpoint tester
- API base URL configurable in app.js
- Dark mode support
- No build step - pure HTML/JS/CSS

Access locally: `http://localhost:8080`

## Deployment Targets

### Heroku (Primary)
- Procfile: `web: bin/server` (expects compiled binary)
- Buildpack: heroku/go
- Addons: heroku-postgresql, heroku-redis
- Environment variables configured via Heroku config
- See README.md "Deployment" section for full setup

### Vercel (Supported)
- Not currently active, but codebase supports serverless functions
- Would require creating /api files and configuring vercel.json
- Current architecture is optimized for Heroku monolithic deployment

## Performance Guidelines

- **Database:** Use connection pool, limit queries to 5s timeout, select only needed columns
- **Caching:** Check cache first, set appropriate TTLs, include X-Cache headers
- **Pagination:** Default 50/page, max 100/page, always include total count
- **Indexes:** Query optimizer relies on indexes in schema.sql
- **AI:** Cache aggressively (15-60min), implement timeouts, fallback to stats if slow
- **Logging:** Log slow queries >100ms, track error rates, monitor pool stats

## Auto-Fetch System

### Overview

The API includes an intelligent auto-fetch system that automatically retrieves missing data when requests return empty results. This creates a self-healing data layer that fetches scheduled games, teams, players, and stats on-demand.

**Key Feature:** When you query `/api/v1/games?season=2025&week=5` and the database returns empty, the system automatically:
1. Detects the empty result
2. Fetches the game schedule from ESPN API
3. Stores it in the database
4. Returns the fetched data
5. Sets `X-Auto-Fetched: true` header

### Architecture

**Components:**
- `internal/utils/season.go` - Season/week detection utilities
- `internal/autofetch/orchestrator.go` - Auto-fetch orchestration logic
- Integrated into handlers (games, players, stats)

**How it works:**
```
User Request → Handler → DB Query → Empty?
                              ↓
                            YES → Auto-Fetch Orchestrator
                              ↓
                       ESPN API Fetch → Store in DB
                              ↓
                       Retry DB Query → Return Data
```

### Season Detection (internal/utils/season.go)

Automatically determines current NFL season and week:

```go
seasonInfo := utils.GetCurrentSeason()
// Returns: Year, CurrentWeek, IsOffseason, IsPreseason, IsRegular, IsPostseason

season, week := utils.GetSeasonWeek(date)
// Get season/week for specific date

shouldFetch := utils.ShouldFetchGames(season, week)
// Determines if auto-fetch should run (current season + previous 1 year)
```

**Season Logic:**
- NFL season year starts in September (e.g., 2025 season = Sep 2025 - Feb 2026)
- Regular season: 18 weeks starting first Thursday of September
- Week calculation based on days since season start
- Offseason detection for March-August

### Auto-Fetch Orchestrator (internal/autofetch/orchestrator.go)

Handles automatic data fetching with deduplication and cascade logic:

**Methods:**
- `FetchGamesIfMissing(ctx, season, week)` - Fetch specific week's games
- `FetchAllSeasonGames(ctx, season)` - Fetch entire season schedule
- `FetchPlayerIfMissing(ctx, nflID)` - Fetch player by NFL ID
- `FetchTeamIfMissing(ctx, nflID)` - Fetch team by NFL ID
- `FetchStatsIfMissing(ctx, gameID)` - Fetch game statistics
- `AutoFetchCurrentWeek(ctx)` - Fetch current week's data

**Features:**
- **Deduplication:** Prevents multiple concurrent fetches of same resource
- **Cascade Fetching:** Automatically fetches dependencies (e.g., games require teams)
- **Graceful Failures:** Logs errors but doesn't fail requests
- **Background Enrichment:** Can enrich data asynchronously

### Handler Integration

**Games Handler Example:**
```go
// internal/handlers/games.go
func (h *GamesHandler) listGames(w http.ResponseWriter, r *http.Request) {
    // ... parse filters ...

    games, total, err := h.queries.ListGames(r.Context(), filters)

    // AUTO-FETCH: If no games found and season/week specified
    if total == 0 && h.autoFetchEnabled && filters.Season > 0 && filters.Week > 0 {
        log.Printf("[AUTO-FETCH] Attempting fetch for season %d week %d",
                   filters.Season, filters.Week)

        if err := h.orchestrator.FetchGamesIfMissing(r.Context(),
                                                      filters.Season,
                                                      filters.Week); err == nil {
            // Retry query after fetch
            games, total, err = h.queries.ListGames(r.Context(), filters)
            w.Header().Set("X-Auto-Fetched", "true")
        }
    }

    response.SuccessWithPagination(w, games, total, limit, offset)
}
```

### Configuration

Auto-fetch is **enabled by default** in handlers. To disable:

```go
handler := NewGamesHandler()
handler.autoFetchEnabled = false  // Disable auto-fetch
```

Environment variables (optional):
- `WEATHER_API_KEY` - For weather enrichment during fetch
- No additional config required - works out of the box

### Fetch Rules

**Games:**
- Auto-fetch if season is current year or previous year (during early season)
- Validates week range (1-18)
- Fetches scheduled games even if not played yet

**Teams:**
- Auto-fetches all 32 teams if fewer than 32 exist
- Runs automatically before fetching games/players

**Players:**
- Fetches all rosters when individual player missing
- Ensures teams exist first

**Stats:**
- Fetches from NFLverse when game stats missing
- Season-specific fetching

### Response Headers

When auto-fetch succeeds:
- `X-Auto-Fetched: true` - Indicates data was just fetched
- `X-Cache: MISS` - Not from cache (newly fetched)

### Logging

Auto-fetch operations are logged with `[AUTO-FETCH]` prefix:
```
[AUTO-FETCH] No games found for season 2025 week 5, attempting auto-fetch
[AUTO-FETCH] Fetching games for season 2025 week 5
[AUTO-FETCH] Successfully fetched and returned 16 games
```

### Use Cases

1. **Future Week Queries:** Query week 5 before it's played → Returns scheduled matchups
2. **Historical Data Gaps:** Query missing historical week → Fetches if within range
3. **New Season Start:** First query of new season → Automatically loads schedule
4. **Player Lookups:** Query unknown player → Fetches latest rosters
5. **Stats Backfill:** Query stats for completed game → Fetches from NFLverse

### Performance Considerations

- Auto-fetch adds latency on first miss (2-10 seconds depending on data size)
- Subsequent requests served from database (fast)
- Deduplication prevents thundering herd on concurrent requests
- Failed fetches logged but don't break API (graceful degradation)
- Async enrichment possible for non-blocking updates

### Best Practices

1. **Preload Data:** Use sync CLI tool for bulk loading
2. **Monitor Logs:** Track auto-fetch frequency and failures
3. **Cache Invalidation:** Clear cache after auto-fetch if using Redis
4. **Rate Limiting:** ESPN API calls respect rate limits
5. **Error Handling:** Auto-fetch failures are non-fatal

## Data Flow Summary

1. **Ingestion:** ESPN/NFLverse → Ingestion Service → PostgreSQL
2. **API Request:** Client → Middleware → Handler → Cache/DB → Response
3. **Auto-Fetch:** DB Empty → Orchestrator → ESPN API → DB → Response
4. **Caching:** Handler checks Redis → Miss = DB query + cache write
5. **AI:** Handler → Claude API → Cache → Response (with retry/fallback)
6. **Sync:** Cron → sync2025 CLI → Ingestion Service → Database
