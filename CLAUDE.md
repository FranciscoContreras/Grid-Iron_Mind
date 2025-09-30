# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Grid Iron Mind is a high-performance NFL data lake with AI-enriched insights via REST API. The system serves as a comprehensive data layer for multiple applications, providing both raw NFL data and intelligent AI analysis. Performance is critical - target response times are under 50ms for cached queries and under 200ms for database queries. The project deploys to Vercel as serverless functions.

**API Base:** gridironmind.vercel.app/api/v1

## Tech Stack

- **Backend:** Go 1.21+
- **Database:** PostgreSQL with pgx/v5 driver
- **Cache:** Redis (go-redis/v8)
- **Deployment:** Vercel Serverless Functions (vercel-go@3.1.0)
- **AI:** Claude API (Sonnet 4) for complex analysis
- **Data Source:** ESPN NFL APIs (public endpoints)

## Architecture

### Project Structure

```
/api/v1/              # Vercel serverless function endpoints (one file per route)
/internal/
  /cache/             # Redis caching layer
  /database/          # Database connection setup
  /db/                # Database queries and operations
  /handlers/          # HTTP request handlers
  /middleware/        # Auth, rate limiting, CORS
  /models/            # Data models (Player, Team, Stats, etc.)
  /services/          # Business logic layer
/pkg/
  /response/          # JSON response formatting
  /validation/        # Input validation utilities
/migrations/          # SQL migration files
```

### Key Architectural Principles

1. **Serverless Design:** Each API endpoint is a separate file in `/api/v1/` directory for Vercel deployment
2. **Layered Architecture:** Handlers → Services → Database, keep handlers thin
3. **Caching Strategy:** Redis caching with 5-minute TTL for frequent queries, 15-60 minutes for AI results
4. **Stateless API:** No session state, authentication via API keys
5. **Performance First:** Connection pooling, query result pagination (max 100/page), indexed queries

### Database Schema

The database has 6 main tables:

- **teams:** NFL team data (32 teams)
- **players:** Player profiles with team relationships
- **games:** Game schedule and scores
- **game_stats:** Per-game player statistics
- **predictions:** AI predictions (game outcomes, player performance, injury risk)
- **ai_analysis:** AI analysis results (player comparisons, trends, play tendencies)

All tables use UUID primary keys. See `migrations/001_initial_schema.sql` for complete schema.

## Common Commands

### Development

```bash
# Install dependencies
go mod download

# Run database migrations
psql $DATABASE_URL -f migrations/001_initial_schema.sql

# Build for local testing
go build -o bin/api ./api/v1/...

# Run tests
go test ./...

# Run specific test
go test -v -run TestPlayerHandler ./internal/handlers/
```

### Deployment

```bash
# Deploy to Vercel (automatic via git push)
vercel deploy

# Deploy to production
vercel --prod
```

### Database

```bash
# Connect to database
psql $DATABASE_URL

# Run migration
psql $DATABASE_URL -f migrations/001_initial_schema.sql

# Reset database (careful!)
psql $DATABASE_URL -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
psql $DATABASE_URL -f migrations/001_initial_schema.sql
```

## API Design

### Endpoint Pattern

All endpoints follow REST conventions with `/api/v1/` prefix:

- **Players:** `/api/v1/players`, `/api/v1/players/:id`, `/api/v1/players/:id/stats`
- **Teams:** `/api/v1/teams`, `/api/v1/teams/:id`, `/api/v1/teams/:id/players`
- **Games:** `/api/v1/games`, `/api/v1/games/:id`
- **Stats:** `/api/v1/stats/leaders`
- **AI:** `/api/v1/ai/analyze`, `/api/v1/ai/predict/game/:game_id`, `/api/v1/ai/predict/player/:player_id/next-game`, `/api/v1/ai/query`, `/api/v1/ai/insights/player/:player_id`, `/api/v1/ai/fantasy/rankings`

### Response Format

**Single resource:**
```json
{
  "data": { /* resource object */ },
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

**Collection:**
```json
{
  "data": [ /* array of objects */ ],
  "meta": {
    "total": 1500,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

**Error:**
```json
{
  "error": {
    "code": "PLAYER_NOT_FOUND",
    "message": "Player with ID 123 does not exist",
    "status": 404
  }
}
```

## AI Implementation

### AI Features

The system provides AI-enriched insights:
- Player performance predictions (next game projections)
- Injury risk assessment
- Game outcome predictions with confidence scores
- Player similarity and comparison analysis
- Stat trend detection and anomaly identification
- Fantasy football scoring predictions
- Play calling tendency analysis
- Natural language query support

### AI Processing Pipeline

1. Raw data ingestion from ESPN API
2. Statistical preprocessing and feature engineering
3. AI model inference (Claude API or local statistical models)
4. Result validation and confidence scoring
5. Cache results in Redis and database
6. Serve via API endpoints
7. Post-event: compare predictions to actual outcomes for accuracy tracking

### Claude API Integration

- Use Claude Sonnet 4 for complex analysis
- Implement retry logic with exponential backoff
- Request JSON-only responses for parsing
- Cache Claude responses aggressively (15-60 minutes)
- Track token usage and costs
- Fallback to statistical methods if Claude API is slow
- Rate limit AI endpoints more strictly (10 per minute)

Example pattern:
```go
type ClaudeRequest struct {
    Model      string    `json:"model"`
    MaxTokens  int       `json:"max_tokens"`
    Messages   []Message `json:"messages"`
}

func analyzeWithClaude(prompt string) (map[string]interface{}, error) {
    // Build request with structured prompt requesting JSON output
    // Call Claude API with retry logic
    // Parse JSON response
    // Cache result with 15-60 minute TTL
    // Return analysis
}
```

## Data Sources

### ESPN API

Primary data source is ESPN NFL public APIs (no auth required). Implement exponential backoff for rate limiting.

**Key endpoints:**
- Active roster: `sports.core.api.espn.com/v2/sports/football/leagues/nfl/athletes?limit=1000&active=true`
- Player overview: `site.web.api.espn.com/apis/common/v3/sports/football/nfl/athletes/{ATHLETE_ID}/overview`
- Team roster: `site.api.espn.com/apis/site/v2/sports/football/nfl/teams/{TEAM_ID}?enable=roster,projection,stats`
- Scoreboard: `site.api.espn.com/apis/site/v2/sports/football/nfl/scoreboard`

### Data Ingestion Strategy

- Create separate ingestion service outside API routes
- Run scheduled jobs to fetch and store data
- Update players/teams daily
- Update game stats every 5 minutes during live games
- Store raw API responses and transformed data
- Handle API changes with versioned parsers

## Code Standards

### Go Conventions

- Follow standard Go project layout
- Use camelCase for unexported, PascalCase for exported
- Return errors, do not panic
- Use `context.Context` for all database operations
- Use prepared statements for all queries
- Validate all input before database calls
- Use struct tags for JSON serialization (e.g., `json:"player_id"`)
- Keep handlers thin, move logic to service layer

### Performance Requirements

- Use connection pooling for PostgreSQL
- Implement query result pagination (max 100 per page)
- Add database query timeout of 5 seconds
- Log slow queries over 100ms
- SELECT only needed columns
- Avoid N+1 queries with proper JOINs or batch loading
- AI endpoint response time under 2 seconds
- Cache hit rate above 80% for common queries

### Error Handling

- Always return errors up the stack
- Use custom error types for domain errors
- Log errors with context (request ID, user ID, etc.)
- Return appropriate HTTP status codes
- Never expose internal errors to API consumers

## Environment Variables

Required environment variables (see `.env.example`):

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `API_KEY`: API key for authentication
- `ENVIRONMENT`: production, staging, or development
- `CLAUDE_API_KEY`: Claude API key for AI features

## Vercel Configuration

The `vercel.json` configures Go runtime for all files in `/api/**/*.go`. Each endpoint needs its own file in the `/api` directory structure. Use `net/http` standard library, set proper CORS headers, and return correct HTTP status codes.

## Authentication & Security

- Implement API key authentication via `X-API-Key` header
- Middleware checks API keys and returns 401 for invalid/missing keys
- Rate limiting: 100 requests per minute per API key (429 with Retry-After header)
- Use Redis to track request counts
- Rate limit AI endpoints more strictly (10 per minute)

## Monitoring

- Log all requests with duration
- Track error rates by endpoint
- Monitor database connection pool
- Alert on response times over 200ms
- Track AI prediction accuracy over time
- Monitor token usage and costs daily