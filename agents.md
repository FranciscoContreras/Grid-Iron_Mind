# Grid Iron Mind - Agent Instructions

# Grid Iron Mind - Agent Instructions

## Project Overview

Grid Iron Mind is a high-performance NFL data lake with AI-enriched insights via REST API. The API serves as the comprehensive data layer for multiple applications with intelligent analysis on top of raw data. Performance is critical. Deploy on Vercel.

Project name: Grid Iron Mind
API base: gridironmind.vercel.app/api/v1

## AI Enrichment Features

The API enriches raw NFL data with intelligent analysis and predictions.

### AI Capabilities

Player performance predictions (next game projections)
Injury risk assessment based on workload and history
Game outcome predictions with confidence scores
Player similarity and comparison analysis
Stat trend detection and anomaly identification
Fantasy football scoring predictions
Play calling tendency analysis
Natural language query support

### AI Implementation Approach

Use Claude API for complex analysis and natural language
Use statistical models for predictions (implement in Go)
Cache AI results with longer TTL (15-60 minutes)
Pre-compute common predictions during off-hours
Store AI results in separate tables with timestamps
Track AI prediction accuracy over time

### AI Data Schema

Add these tables to database:

Predictions table:

- id (UUID primary key)
- prediction_type (text: game_outcome, player_performance, injury_risk)
- entity_id (UUID: player_id or game_id)
- prediction_data (JSONB: flexible structure for different prediction types)
- confidence_score (decimal 0-1)
- generated_at (timestamp)
- valid_until (timestamp)
- actual_outcome (JSONB: populated after event occurs)
- accuracy_score (decimal: calculated after outcome known)

AI_Analysis table:

- id (UUID primary key)
- analysis_type (text: player_comparison, trend_analysis, play_tendency)
- subject_ids (JSONB: array of relevant entity IDs)
- analysis_result (JSONB: structured analysis output)
- metadata (JSONB: model version, parameters used)
- created_at (timestamp)
- expires_at (timestamp)

### AI API Endpoints

POST /api/v1/ai/analyze
Body: { "type": "player_comparison", "player_ids": ["id1", "id2"] }
Response: detailed comparison with stats, trends, insights

GET /api/v1/ai/predict/game/:game_id
Response: game outcome prediction with confidence score

GET /api/v1/ai/predict/player/:player_id/next-game
Response: projected stats for next game

POST /api/v1/ai/query
Body: { "question": "Who are the top 5 running backs this season?" }
Response: natural language answer with data

GET /api/v1/ai/insights/player/:player_id
Response: AI-generated insights about player trends, strengths, risks

GET /api/v1/ai/fantasy/rankings
Query params: position, week, scoring_format
Response: AI-powered fantasy rankings with projections

### AI Processing Pipeline

Raw data ingestion from ESPN API
Statistical preprocessing and feature engineering
AI model inference (Claude API or local models)
Result validation and confidence scoring
Cache results in Redis and database
Serve via API endpoints
Post-event: compare predictions to actual outcomes
Model performance tracking and improvement

### Claude API Integration for AI

Use Claude Sonnet 4 for complex analysis
Implement retry logic with exponential backoff
Keep prompts concise and structured
Request JSON responses for parsing
Track token usage and costs
Cache Claude responses aggressively
Use system prompts to ensure consistent output format

Example Claude integration in Go:

```go
type ClaudeRequest struct {
    Model      string    `json:"model"`
    MaxTokens  int       `json:"max_tokens"`
    Messages   []Message `json:"messages"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func analyzePlayerComparison(player1, player2 PlayerStats) (string, error) {
    prompt := fmt.Sprintf(`Compare these NFL players. Return JSON only.
Player 1: %v
Player 2: %v

Respond with:
{
  "stronger_player": "player1 or player2",
  "key_differences": ["string array"],
  "recommendation": "string"
}`, player1, player2)

    // Call Claude API
    // Parse JSON response
    // Return analysis
}
```

### AI Performance Requirements

AI endpoints response time under 2 seconds
Cache hit rate above 80% for common queries
Prediction accuracy tracked and displayed
Fallback to cached or statistical methods if Claude API slow
Rate limit AI endpoints more strictly (10 per minute)
Monitor token usage and costs daily

## Tech Stack

Backend: Go 1.21+
Database: PostgreSQL
Cache: Redis
Deployment: Vercel Serverless Functions
API Format: REST with JSON

## Architecture Goals

Fast response times (under 50ms for cached queries)
Clean API design with versioning
Comprehensive NFL player data storage
Easy integration for external applications
Proper error handling and validation
Rate limiting and authentication ready

## Project Structure

/api
/v1
players.go
teams.go
stats.go
games.go
/internal
/db
postgres.go
queries.go
/models
player.go
team.go
stats.go
/cache
redis.go
/handlers
players.go
teams.go
/middleware
auth.go
ratelimit.go
cors.go
/pkg
/response
json.go
/validation
validate.go
vercel.json
go.mod
README.md

## Database Schema Design

Players table:

- id (UUID primary key)
- nfl_id (unique integer)
- name (text)
- position (text)
- team_id (UUID foreign key)
- jersey_number (integer)
- height_inches (integer)
- weight_pounds (integer)
- birth_date (date)
- college (text)
- draft_year (integer)
- draft_round (integer)
- draft_pick (integer)
- status (text: active, injured, inactive)
- created_at (timestamp)
- updated_at (timestamp)

Teams table:

- id (UUID primary key)
- nfl_id (unique integer)
- name (text)
- abbreviation (text, unique)
- city (text)
- conference (text)
- division (text)
- stadium (text)
- created_at (timestamp)
- updated_at (timestamp)

Game_Stats table:

- id (UUID primary key)
- player_id (UUID foreign key)
- game_id (UUID foreign key)
- season (integer)
- week (integer)
- passing_yards (integer)
- rushing_yards (integer)
- receiving_yards (integer)
- touchdowns (integer)
- interceptions (integer)
- completions (integer)
- attempts (integer)
- targets (integer)
- receptions (integer)
- created_at (timestamp)

Games table:

- id (UUID primary key)
- nfl_game_id (unique text)
- home_team_id (UUID foreign key)
- away_team_id (UUID foreign key)
- game_date (timestamp)
- season (integer)
- week (integer)
- home_score (integer)
- away_score (integer)
- status (text: scheduled, live, final)
- created_at (timestamp)

Add indexes on:

- players.nfl_id
- players.team_id
- teams.abbreviation
- game_stats.player_id
- game_stats.game_id
- game_stats.season
- games.game_date

## API Endpoints

GET /api/v1/players
Query params: team, position, status, limit, offset
Response: paginated list of players

GET /api/v1/players/:id
Response: single player with current season stats

GET /api/v1/players/:id/stats
Query params: season, week_start, week_end
Response: player stats for specified range

GET /api/v1/teams
Response: all teams

GET /api/v1/teams/:id
Response: single team with roster

GET /api/v1/teams/:id/players
Response: all players on team

GET /api/v1/games
Query params: season, week, team
Response: list of games

GET /api/v1/games/:id
Response: single game with stats

GET /api/v1/stats/leaders
Query params: stat_type, season, position, limit
Response: leaderboard for specified stat

## Coding Standards

Use standard Go project layout
Follow Go naming conventions (camelCase for unexported, PascalCase for exported)
Return errors, do not panic
Use context.Context for all database operations
Use prepared statements for all queries
Validate all input before database calls
Use struct tags for JSON serialization
Keep handlers thin, move logic to service layer
Write tests for all business logic
Use environment variables for configuration

## Error Response Format

{
"error": {
"code": "PLAYER_NOT_FOUND",
"message": "Player with ID 123 does not exist",
"status": 404
}
}

## Success Response Format

Single resource:
{
"data": { player object },
"meta": {
"timestamp": "2025-09-30T12:00:00Z"
}
}

Collection:
{
"data": [ array of objects ],
"meta": {
"total": 1500,
"limit": 50,
"offset": 0,
"timestamp": "2025-09-30T12:00:00Z"
}
}

## Performance Requirements

Cache frequently accessed data in Redis with 5 minute TTL
Use connection pooling for PostgreSQL
Implement query result pagination (max 100 per page)
Add database query timeout of 5 seconds
Log slow queries over 100ms
Use SELECT only needed columns
Avoid N+1 queries with proper JOINs or batch loading

## Vercel Configuration

Create vercel.json:
{
"functions": {
"api/**/*.go": {
"runtime": "vercel-go@3.1.0"
}
}
}

Each endpoint needs its own file in /api directory
Use net/http standard library
Set proper CORS headers
Return correct HTTP status codes

## Environment Variables

DATABASE_URL (PostgreSQL connection string)
REDIS_URL (Redis connection string)
API_KEY (for authentication)
ENVIRONMENT (production, staging, development)

## Data Sources - ESPN API

Primary data source: ESPN NFL APIs
No authentication required for public endpoints
Rate limiting: unknown, implement exponential backoff

### Core Endpoints for Data Lake

**Players:**

- Active roster: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/athletes?limit=1000&active=true`
- Player overview: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/overview`
- Player stats: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/gamelog`
- Player splits: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/splits`

**Teams:**

- All teams: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams`
- Team roster: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams/{TEAM_ID}?enable=roster,projection,stats`
- Team schedule: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams/{TEAM_ID}/schedule?season={YEAR}`
- Team injuries: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/teams/{TEAM_ID}/injuries?limit=100`

**Games:**

- Scoreboard (all games): `site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?dates={YEAR}&seasontype={SEASONTYPE}`
- Game summary: `site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event={EVENT_ID}`
- Play by play: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/events/{EVENT_ID}/competitions/{EVENT_ID}/plays?limit=300`

**Stats:**

- Season leaders: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/seasons/{YEAR}/types/{SEASONTYPE}/leaders`
- Team statistics: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/seasons/{YEAR}/types/{SEASONTYPE}/teams/{TEAM_ID}/statistics`

### Ingestion Strategy

Create separate ingestion service outside API routes
Run scheduled jobs to fetch and store data
Process data in batches (players, teams, games, stats)
Store raw API responses and transformed data
Update players/teams daily
Update game stats every 5 minutes during live games
Handle ESPN API changes with versioned parsers

## Authentication

Implement API key authentication
Add middleware to check X-API-Key header
Return 401 for missing or invalid keys
Rate limit by API key

## Rate Limiting

100 requests per minute per API key
Return 429 with Retry-After header when exceeded
Use Redis to track request counts

## Monitoring

Log all requests with duration
Track error rates by endpoint
Monitor database connection pool
Alert on response times over 200ms

## Build Phases - Step by Step Implementation

Follow these phases in order. Complete each phase fully before moving to the next. Test each phase before proceeding.

### Phase 1: Foundation Setup

**Objective:** Create project structure, database, and basic infrastructure.

**Tasks:**

1. Initialize Go project with modules

   - Run `go mod init gridironmind`
   - Create all directories from project structure
   - Set up .gitignore for Go projects
2. Create database schema SQL file

   - Write CREATE TABLE statements for all tables
   - Add all indexes from schema design
   - Include sample data insert statements for testing
   - Save as `schema.sql` in project root
3. Set up PostgreSQL connection

   - Create `/internal/db/postgres.go`
   - Implement connection pooling with pgx
   - Add health check function
   - Read connection string from environment variable
4. Create base models

   - Define structs in `/internal/models/` for Player, Team, Game, GameStats
   - Add JSON tags for API serialization
   - Add validation tags
5. Create response utilities

   - Build `/pkg/response/json.go` with success and error response functions
   - Implement standard response formats from agents.md
6. Set up environment configuration

   - Create `.env.example` file with all required variables
   - Create config loader in `/internal/config/`
   - Document all environment variables

**Completion Criteria:**

- Project compiles without errors
- Database schema file is complete and tested
- Database connection works
- All directories exist with placeholder files

**Testing:**

- Run `go build` successfully
- Execute schema.sql on test database
- Connect to database and verify tables exist

---

### Phase 2: Basic API Endpoints

**Objective:** Implement core CRUD operations for players and teams.

**Tasks:**

1. Create players handler

   - GET /api/v1/players (list with pagination)
   - GET /api/v1/players/:id (single player)
   - Implement query parameters: limit, offset, position, team, status
2. Create teams handler

   - GET /api/v1/teams (list all teams)
   - GET /api/v1/teams/:id (single team)
   - GET /api/v1/teams/:id/players (team roster)
3. Build database query layer

   - Create `/internal/db/queries.go`
   - Write prepared statements for all queries
   - Implement pagination logic
   - Add query timeouts (5 seconds)
4. Set up API routing

   - Create `/api/v1/players.go` for Vercel function
   - Create `/api/v1/teams.go` for Vercel function
   - Implement proper HTTP method handling
   - Add CORS headers
5. Implement error handling

   - Create error types (NotFound, ValidationError, DatabaseError)
   - Build middleware for consistent error responses
   - Add logging for all errors
6. Create Vercel configuration

   - Write `vercel.json` with function configs
   - Set up proper routes
   - Configure environment variables

**Completion Criteria:**

- All player endpoints return proper JSON
- All team endpoints return proper JSON
- Pagination works correctly
- Error responses follow standard format
- CORS headers are set

**Testing:**

- Test each endpoint with curl or Postman
- Verify pagination with different limit/offset values
- Test error cases (invalid IDs, missing parameters)
- Check response times (should be under 200ms)

---

### Phase 2.5: Simple Testing Dashboard

**Objective:** Build a simple web dashboard to test API endpoints and view player data.

**Tasks:**

1. Create dashboard directory structure

   - Create `/dashboard/` folder
   - Create `index.html` as main dashboard page
   - Create `styles.css` for basic styling
   - Create `app.js` for API calls and interactivity
2. Build dashboard HTML structure

   - Header with project name and status indicators
   - Navigation tabs (Players, Teams, Stats, API Testing)
   - Data display area with tables
   - Search and filter controls
   - Response viewer for raw API data
3. Implement API testing interface

   - Endpoint selector dropdown (all available endpoints)
   - Parameter input fields (dynamic based on endpoint)
   - Execute button to make API calls
   - Response display area (formatted JSON)
   - Response time and status code display
4. Build player data viewer

   - Table showing all players with pagination
   - Search by name, position, team
   - Filter by position, status, team
   - Click player to see full details
   - Display player stats if available
5. Create team roster viewer

   - Team selector dropdown
   - Display team info and logo
   - Show complete roster in table format
   - Sort by position, jersey number, name
6. Add stats leaderboard viewer

   - Stat type selector (passing yards, rushing yards, etc.)
   - Position filter
   - Top 10/25/50 selector
   - Sortable table with player names and stats
7. Implement data refresh

   - Add refresh button for each section
   - Show loading indicators during API calls
   - Display errors clearly if API calls fail
   - Cache responses in browser for 5 minutes
8. Style the dashboard

   - Clean, modern design
   - Responsive layout (works on mobile)
   - Use simple CSS (no frameworks needed)
   - Dark mode option
   - NFL team colors where appropriate
9. Add dashboard deployment

   - Deploy dashboard to Vercel as static site
   - Configure CORS on API to allow dashboard domain
   - Or include dashboard in `/public` folder served by API

**Dashboard Features:**

API Tester:

- Test any endpoint with custom parameters
- View raw JSON responses
- Copy response to clipboard
- Save common test queries

Player Browser:

- Paginated player list
- Quick search and filters
- Player detail modal with stats
- Link to ESPN profile

Team Viewer:

- All 32 NFL teams
- Full roster with player cards
- Team stats and schedule
- Injury reports

Stats Dashboard:

- Current season leaderboards
- Compare players side by side
- Trend charts (if time permits)
- Export data as CSV

**Tech Stack for Dashboard:**

- Pure HTML/CSS/JavaScript (no framework needed)
- Fetch API for HTTP requests
- LocalStorage for caching
- CSS Grid for layout
- Responsive design

**File Structure:**

```
/dashboard
  index.html          (main page)
  styles.css          (all styling)
  app.js              (API calls, DOM manipulation)
  /assets
    logo.png
    team-logos/       (NFL team logos if needed)
```

**Example Dashboard Layout:**

```
+------------------------------------------+
| Grid Iron Mind Dashboard      [Refresh] |
+------------------------------------------+
| [Players] [Teams] [Stats] [API Testing] |
+------------------------------------------+
|                                          |
| Search: [________] Position: [All ▼]    |
|                                          |
| +------------------------------------+   |
| | Name        | Team  | Pos | Status |  |
| |-------------|-------|-----|--------|  |
| | P. Mahomes  | KC    | QB  | Active |  |
| | J. Allen    | BUF   | QB  | Active |  |
| | ...                                |   |
| +------------------------------------+   |
|                                          |
| << Prev | Page 1 of 50 | Next >>        |
+------------------------------------------+
| Status: Connected | Response: 45ms      |
+------------------------------------------+
```

**Completion Criteria:**

- Dashboard loads and displays correctly
- All tabs/sections functional
- Can test all API endpoints
- Player data displays in table
- Search and filters work
- Pagination functions properly
- Errors display clearly
- Works on desktop and mobile

**Testing:**

- Open dashboard in browser
- Test each navigation tab
- Make API calls to all endpoints
- Search for players by name
- Filter by position and team
- Test pagination (next/prev)
- Verify error handling with invalid requests
- Test on mobile browser

---

### Phase 3: Data Ingestion Service

**Objective:** Build service to fetch and store NFL data from ESPN API.

**Tasks:**

1. Create ESPN API client

   - Build `/internal/espn/client.go`
   - Implement rate limiting and retry logic
   - Add response parsing for each endpoint type
   - Handle API errors gracefully
2. Build data ingestion functions

   - Create `/cmd/ingest/main.go` for running ingestion jobs
   - Implement FetchAndStorePlayers()
   - Implement FetchAndStoreTeams()
   - Implement FetchAndStoreGames()
   - Add progress logging
3. Implement data transformation

   - Create mappers from ESPN format to internal models
   - Handle missing or null fields
   - Validate data before insertion
4. Set up database upsert logic

   - Create upsert queries (INSERT ON CONFLICT UPDATE)
   - Handle duplicate detection by ESPN IDs
   - Update existing records without losing data
5. Create ingestion scheduler

   - Build simple CLI tool for manual runs
   - Add flags for different data types (--players, --teams, --games)
   - Add date range support for historical data
6. Add data validation

   - Verify required fields exist
   - Check data types and ranges
   - Log validation errors without stopping ingestion

**Completion Criteria:**

- Can fetch players from ESPN API
- Can fetch teams from ESPN API
- Data is transformed correctly
- Data is stored in database without errors
- Duplicate records are handled properly

**Testing:**

- Run ingestion for 5 players manually
- Verify data appears correctly in database
- Run ingestion twice and verify no duplicates
- Check that updates work for existing records
- Test error handling with invalid ESPN responses

---

### Phase 4: Game Stats and Advanced Endpoints

**Objective:** Add game statistics, schedules, and complex query endpoints.

**Tasks:**

1. Implement game endpoints

   - GET /api/v1/games (list with filters)
   - GET /api/v1/games/:id (game details)
   - Query params: season, week, team
   - Include team info and scores in response
2. Add player stats endpoints

   - GET /api/v1/players/:id/stats (career stats)
   - Query params: season, week_start, week_end
   - GET /api/v1/players/:id/gamelog (game by game stats)
3. Create stats leaderboard endpoint

   - GET /api/v1/stats/leaders
   - Query params: stat_type, season, position, limit
   - Sort by specified stat
   - Include player and team info
4. Build team schedule endpoint

   - GET /api/v1/teams/:id/schedule
   - Query params: season
   - Include game results and upcoming games
5. Extend ingestion service

   - Add FetchAndStoreGameStats()
   - Fetch play by play data
   - Update game scores in real-time during live games
6. Optimize database queries

   - Add compound indexes for common queries
   - Use JOINs instead of multiple queries
   - Add query result logging for slow queries

**Completion Criteria:**

- All game endpoints return correct data
- Player stats endpoints work with filters
- Leaderboard returns sorted results
- Team schedules show complete season
- Game stats are stored correctly

**Testing:**

- Query games for specific week/season
- Get player stats for multiple seasons
- Test leaderboard with different stat types
- Verify schedule shows past and future games
- Check that stats match ESPN website

---

### Phase 5: Caching and Performance

**Objective:** Add Redis caching to improve response times.

**Tasks:**

1. Set up Redis connection

   - Create `/internal/cache/redis.go`
   - Implement connection with retry logic
   - Add health check function
   - Handle Redis unavailability gracefully
2. Implement cache layer

   - Create cache key generation functions
   - Add cache wrapper for database queries
   - Set appropriate TTLs (5 min for hot data, 60 min for cold)
   - Implement cache invalidation logic
3. Add caching to endpoints

   - Cache player lists (5 min TTL)
   - Cache team rosters (15 min TTL)
   - Cache game results (permanent until game ends)
   - Cache leaderboards (10 min TTL)
4. Build cache middleware

   - Create middleware to check cache before handler
   - Add cache headers to responses (Cache-Control, ETag)
   - Implement cache warming for popular queries
5. Add performance monitoring

   - Log cache hit/miss rates
   - Track response times by endpoint
   - Monitor database connection pool usage
6. Optimize data structures

   - Use efficient JSON encoding
   - Minimize response payload sizes
   - Add field filtering (only return requested fields)

**Completion Criteria:**

- Redis connection works
- Cache hit rate above 50% for common queries
- Response times under 50ms for cached data
- Response times under 200ms for uncached data
- Cache invalidation works correctly

**Testing:**

- Make same request twice, verify cache hit on second
- Check response times improve with caching
- Verify cache expires after TTL
- Test with Redis unavailable (should still work)
- Monitor cache hit rates over time

---

### Phase 6: AI Integration

**Objective:** Add AI-powered predictions and analysis endpoints.

**Tasks:**

1. Set up AI database tables

   - Create predictions and ai_analysis tables
   - Add indexes for query performance
   - Create migration file
2. Build Claude API client

   - Create `/internal/ai/claude.go`
   - Implement request/response handling
   - Add retry logic for rate limits
   - Track token usage
3. Implement prediction functions

   - Create game outcome predictor
   - Build player performance predictor
   - Add injury risk assessment
   - Store predictions in database
4. Create analysis functions

   - Build player comparison analyzer
   - Implement trend detection
   - Create play tendency analyzer
5. Build AI API endpoints

   - POST /api/v1/ai/analyze (player comparisons, trends)
   - GET /api/v1/ai/predict/game/:game_id
   - GET /api/v1/ai/predict/player/:player_id/next-game
   - POST /api/v1/ai/query (natural language)
   - GET /api/v1/ai/insights/player/:player_id
6. Implement AI caching strategy

   - Cache AI responses for 15-60 minutes
   - Store predictions in database
   - Track prediction accuracy
   - Update accuracy scores after games
7. Build statistical models

   - Create simple regression models in Go for fast predictions
   - Use historical data for training
   - Fall back to statistical models if Claude API slow

**Completion Criteria:**

- Claude API integration works
- All AI endpoints return valid responses
- Predictions are stored in database
- Cache works for AI responses
- Statistical fallback functions properly

**Testing:**

- Request player comparison, verify response quality
- Get game prediction, check confidence score
- Test natural language query endpoint
- Verify predictions are cached
- Check response times under 2 seconds

---

### Phase 7: Authentication and Rate Limiting

**Objective:** Secure API with authentication and prevent abuse.

**Tasks:**

1. Implement API key system

   - Create api_keys table in database
   - Generate secure random keys
   - Store key hashes (not plain keys)
   - Add key metadata (created_at, last_used, rate_limit_tier)
2. Build authentication middleware

   - Check X-API-Key header
   - Validate key against database
   - Return 401 for missing/invalid keys
   - Update last_used timestamp
3. Create rate limiting

   - Use Redis to track request counts
   - Implement sliding window rate limiting
   - Different limits per endpoint type (100/min standard, 10/min AI)
   - Return 429 with Retry-After header
4. Add rate limit tiers

   - Free tier: 100 requests/hour
   - Basic tier: 1000 requests/hour
   - Pro tier: 10000 requests/hour
   - Store tier in api_keys table
5. Build admin endpoints

   - POST /api/v1/admin/keys (create new API key)
   - GET /api/v1/admin/keys (list keys)
   - DELETE /api/v1/admin/keys/:id (revoke key)
   - Require admin authentication
6. Add request logging

   - Log all requests with API key, endpoint, timestamp
   - Store in database for analytics
   - Track usage by key

**Completion Criteria:**

- API requires valid key for all endpoints
- Rate limiting works correctly
- Different tiers have different limits
- Admin endpoints work
- Request logging captures all traffic

**Testing:**

- Test endpoints without API key (should fail)
- Test with valid API key (should work)
- Exceed rate limit and verify 429 response
- Test different rate limit tiers
- Verify admin endpoints create/revoke keys

---

### Phase 8: Production Deployment

**Objective:** Deploy to Vercel and prepare for production traffic.

**Tasks:**

1. Set up production database

   - Create PostgreSQL database on hosting service (Vercel Postgres, Supabase, or Railway)
   - Run schema migrations
   - Set up connection pooling
   - Configure backups
2. Set up production Redis

   - Create Redis instance (Vercel KV or Upstash)
   - Configure persistence
   - Set memory limits
3. Configure Vercel project

   - Create new Vercel project
   - Set all environment variables
   - Configure custom domain (if applicable)
   - Set up proper regions
4. Run initial data ingestion

   - Fetch all current season data
   - Fetch historical data (past 3 seasons)
   - Verify data completeness
   - Set up scheduled ingestion (daily)
5. Performance testing

   - Load test API endpoints
   - Verify response times under load
   - Check database connection pool behavior
   - Monitor memory usage
6. Set up monitoring

   - Configure error tracking (Sentry or similar)
   - Set up uptime monitoring
   - Add performance dashboards
   - Configure alerts for errors and slow responses
7. Create API documentation

   - Document all endpoints in README
   - Add example requests/responses
   - Document authentication
   - Provide code examples
8. Build status page

   - Create /api/v1/health endpoint
   - Show database status
   - Show Redis status
   - Show API version

**Completion Criteria:**

- API deployed to Vercel
- Database populated with data
- All endpoints working in production
- Monitoring configured
- Documentation complete
- Performance meets targets

**Testing:**

- Test all endpoints in production
- Run load tests with realistic traffic
- Verify caching works in production
- Check error handling in production
- Monitor for several hours/days

---

## Phase Execution Guidelines

**For each phase:**

1. Read the phase objectives and tasks completely
2. Ask clarifying questions if anything is unclear
3. Build all components for the phase
4. Test thoroughly before marking complete
5. Document any deviations or issues
6. Get approval before moving to next phase

**Do not skip phases or combine them unless explicitly told to do so.**

**If you encounter blockers:**

- Document the issue clearly
- Propose solutions or alternatives
- Wait for guidance before proceeding

**After completing each phase:**

- Summarize what was built
- Show test results
- List any known issues or limitations
- Confirm readiness for next phase

## Important Notes

Focus on read performance over write performance
Data consistency is important (use transactions)
API should be stateless
Version all endpoints (start with v1)
Document all endpoints in README
Keep response times under 50ms for cached data
Keep response times under 200ms for database queries
Test with production-like data volumes

## Project Overview

Grid Iron Mind is a high-performance NFL data lake with AI-enriched insights via REST API. The API serves as the comprehensive data layer for multiple applications with intelligent analysis on top of raw data. Performance is critical. Deploy on Vercel.

Project name: Grid Iron Mind
API base: gridironmind.vercel.app/api/v1

## AI Enrichment Features

The API enriches raw NFL data with intelligent analysis and predictions.

### AI Capabilities

Player performance predictions (next game projections)
Injury risk assessment based on workload and history
Game outcome predictions with confidence scores
Player similarity and comparison analysis
Stat trend detection and anomaly identification
Fantasy football scoring predictions
Play calling tendency analysis
Natural language query support

### AI Implementation Approach

Use Claude API for complex analysis and natural language
Use statistical models for predictions (implement in Go)
Cache AI results with longer TTL (15-60 minutes)
Pre-compute common predictions during off-hours
Store AI results in separate tables with timestamps
Track AI prediction accuracy over time

### AI Data Schema

Add these tables to database:

Predictions table:

- id (UUID primary key)
- prediction_type (text: game_outcome, player_performance, injury_risk)
- entity_id (UUID: player_id or game_id)
- prediction_data (JSONB: flexible structure for different prediction types)
- confidence_score (decimal 0-1)
- generated_at (timestamp)
- valid_until (timestamp)
- actual_outcome (JSONB: populated after event occurs)
- accuracy_score (decimal: calculated after outcome known)

AI_Analysis table:

- id (UUID primary key)
- analysis_type (text: player_comparison, trend_analysis, play_tendency)
- subject_ids (JSONB: array of relevant entity IDs)
- analysis_result (JSONB: structured analysis output)
- metadata (JSONB: model version, parameters used)
- created_at (timestamp)
- expires_at (timestamp)

### AI API Endpoints

POST /api/v1/ai/analyze
Body: { "type": "player_comparison", "player_ids": ["id1", "id2"] }
Response: detailed comparison with stats, trends, insights

GET /api/v1/ai/predict/game/:game_id
Response: game outcome prediction with confidence score

GET /api/v1/ai/predict/player/:player_id/next-game
Response: projected stats for next game

POST /api/v1/ai/query
Body: { "question": "Who are the top 5 running backs this season?" }
Response: natural language answer with data

GET /api/v1/ai/insights/player/:player_id
Response: AI-generated insights about player trends, strengths, risks

GET /api/v1/ai/fantasy/rankings
Query params: position, week, scoring_format
Response: AI-powered fantasy rankings with projections

### AI Processing Pipeline

Raw data ingestion from ESPN API
Statistical preprocessing and feature engineering
AI model inference (Claude API or local models)
Result validation and confidence scoring
Cache results in Redis and database
Serve via API endpoints
Post-event: compare predictions to actual outcomes
Model performance tracking and improvement

### Claude API Integration for AI

Use Claude Sonnet 4 for complex analysis
Implement retry logic with exponential backoff
Keep prompts concise and structured
Request JSON responses for parsing
Track token usage and costs
Cache Claude responses aggressively
Use system prompts to ensure consistent output format

Example Claude integration in Go:

```go
type ClaudeRequest struct {
    Model      string    `json:"model"`
    MaxTokens  int       `json:"max_tokens"`
    Messages   []Message `json:"messages"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

func analyzePlayerComparison(player1, player2 PlayerStats) (string, error) {
    prompt := fmt.Sprintf(`Compare these NFL players. Return JSON only.
Player 1: %v
Player 2: %v

Respond with:
{
  "stronger_player": "player1 or player2",
  "key_differences": ["string array"],
  "recommendation": "string"
}`, player1, player2)

    // Call Claude API
    // Parse JSON response
    // Return analysis
}
```

### AI Performance Requirements

AI endpoints response time under 2 seconds
Cache hit rate above 80% for common queries
Prediction accuracy tracked and displayed
Fallback to cached or statistical methods if Claude API slow
Rate limit AI endpoints more strictly (10 per minute)
Monitor token usage and costs daily

## Tech Stack

Backend: Go 1.21+
Database: PostgreSQL
Cache: Redis
Deployment: Vercel Serverless Functions
API Format: REST with JSON

## Architecture Goals

Fast response times (under 50ms for cached queries)
Clean API design with versioning
Comprehensive NFL player data storage
Easy integration for external applications
Proper error handling and validation
Rate limiting and authentication ready

## Project Structure

/api
/v1
players.go
teams.go
stats.go
games.go
/internal
/db
postgres.go
queries.go
/models
player.go
team.go
stats.go
/cache
redis.go
/handlers
players.go
teams.go
/middleware
auth.go
ratelimit.go
cors.go
/pkg
/response
json.go
/validation
validate.go
vercel.json
go.mod
README.md

## Database Schema Design

Players table:

- id (UUID primary key)
- nfl_id (unique integer)
- name (text)
- position (text)
- team_id (UUID foreign key)
- jersey_number (integer)
- height_inches (integer)
- weight_pounds (integer)
- birth_date (date)
- college (text)
- draft_year (integer)
- draft_round (integer)
- draft_pick (integer)
- status (text: active, injured, inactive)
- created_at (timestamp)
- updated_at (timestamp)

Teams table:

- id (UUID primary key)
- nfl_id (unique integer)
- name (text)
- abbreviation (text, unique)
- city (text)
- conference (text)
- division (text)
- stadium (text)
- created_at (timestamp)
- updated_at (timestamp)

Game_Stats table:

- id (UUID primary key)
- player_id (UUID foreign key)
- game_id (UUID foreign key)
- season (integer)
- week (integer)
- passing_yards (integer)
- rushing_yards (integer)
- receiving_yards (integer)
- touchdowns (integer)
- interceptions (integer)
- completions (integer)
- attempts (integer)
- targets (integer)
- receptions (integer)
- created_at (timestamp)

Games table:

- id (UUID primary key)
- nfl_game_id (unique text)
- home_team_id (UUID foreign key)
- away_team_id (UUID foreign key)
- game_date (timestamp)
- season (integer)
- week (integer)
- home_score (integer)
- away_score (integer)
- status (text: scheduled, live, final)
- created_at (timestamp)

Add indexes on:

- players.nfl_id
- players.team_id
- teams.abbreviation
- game_stats.player_id
- game_stats.game_id
- game_stats.season
- games.game_date

## API Endpoints

GET /api/v1/players
Query params: team, position, status, limit, offset
Response: paginated list of players

GET /api/v1/players/:id
Response: single player with current season stats

GET /api/v1/players/:id/stats
Query params: season, week_start, week_end
Response: player stats for specified range

GET /api/v1/teams
Response: all teams

GET /api/v1/teams/:id
Response: single team with roster

GET /api/v1/teams/:id/players
Response: all players on team

GET /api/v1/games
Query params: season, week, team
Response: list of games

GET /api/v1/games/:id
Response: single game with stats

GET /api/v1/stats/leaders
Query params: stat_type, season, position, limit
Response: leaderboard for specified stat

## Coding Standards

Use standard Go project layout
Follow Go naming conventions (camelCase for unexported, PascalCase for exported)
Return errors, do not panic
Use context.Context for all database operations
Use prepared statements for all queries
Validate all input before database calls
Use struct tags for JSON serialization
Keep handlers thin, move logic to service layer
Write tests for all business logic
Use environment variables for configuration

## Error Response Format

{
"error": {
"code": "PLAYER_NOT_FOUND",
"message": "Player with ID 123 does not exist",
"status": 404
}
}

## Success Response Format

Single resource:
{
"data": { player object },
"meta": {
"timestamp": "2025-09-30T12:00:00Z"
}
}

Collection:
{
"data": [ array of objects ],
"meta": {
"total": 1500,
"limit": 50,
"offset": 0,
"timestamp": "2025-09-30T12:00:00Z"
}
}

## Performance Requirements

Cache frequently accessed data in Redis with 5 minute TTL
Use connection pooling for PostgreSQL
Implement query result pagination (max 100 per page)
Add database query timeout of 5 seconds
Log slow queries over 100ms
Use SELECT only needed columns
Avoid N+1 queries with proper JOINs or batch loading

## Vercel Configuration

Create vercel.json:
{
"functions": {
"api/**/*.go": {
"runtime": "vercel-go@3.1.0"
}
}
}

Each endpoint needs its own file in /api directory
Use net/http standard library
Set proper CORS headers
Return correct HTTP status codes

## Environment Variables

DATABASE_URL (PostgreSQL connection string)
REDIS_URL (Redis connection string)
API_KEY (for authentication)
ENVIRONMENT (production, staging, development)

## Data Sources - ESPN API

Primary data source: ESPN NFL APIs
No authentication required for public endpoints
Rate limiting: unknown, implement exponential backoff

### Core Endpoints for Data Lake

**Players:**

- Active roster: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/athletes?limit=1000&active=true`
- Player overview: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/overview`
- Player stats: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/gamelog`
- Player splits: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/splits`

**Teams:**

- All teams: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams`
- Team roster: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams/{TEAM_ID}?enable=roster,projection,stats`
- Team schedule: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams/{TEAM_ID}/schedule?season={YEAR}`
- Team injuries: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/teams/{TEAM_ID}/injuries?limit=100`

**Games:**

- Scoreboard (all games): `site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard?dates={YEAR}&seasontype={SEASONTYPE}`
- Game summary: `site.api.espn.com/apis/site/v2/sports/football/nfl/summary?event={EVENT_ID}`
- Play by play: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/events/{EVENT_ID}/competitions/{EVENT_ID}/plays?limit=300`

**Stats:**

- Season leaders: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/seasons/{YEAR}/types/{SEASONTYPE}/leaders`
- Team statistics: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/seasons/{YEAR}/types/{SEASONTYPE}/teams/{TEAM_ID}/statistics`

### Ingestion Strategy

Create separate ingestion service outside API routes
Run scheduled jobs to fetch and store data
Process data in batches (players, teams, games, stats)
Store raw API responses and transformed data
Update players/teams daily
Update game stats every 5 minutes during live games
Handle ESPN API changes with versioned parsers

## Authentication

Implement API key authentication
Add middleware to check X-API-Key header
Return 401 for missing or invalid keys
Rate limit by API key

## Rate Limiting

100 requests per minute per API key
Return 429 with Retry-After header when exceeded
Use Redis to track request counts

## Monitoring

Log all requests with duration
Track error rates by endpoint
Monitor database connection pool
Alert on response times over 200ms

## Next Steps for Initial Build

1. Set up Go project with modules
2. Create database schema and migrations
3. Set up PostgreSQL connection with pgx
4. Implement player endpoints first
5. Add Redis caching layer
6. Create Vercel configuration
7. Deploy and test
8. Add remaining endpoints
9. Implement authentication
10. Add rate limiting

## Important Notes

Focus on read performance over write performance
Data consistency is important (use transactions)
API should be stateless
Version all endpoints (start with v1)
Document all endpoints in README
Keep response times under 50ms for cached data
Keep response times under 200ms for database queries
Test with production-like data volumes
