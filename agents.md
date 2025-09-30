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
