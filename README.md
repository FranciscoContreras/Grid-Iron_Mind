# Grid Iron Mind API

High-performance NFL data lake with AI-enriched insights via REST API.

## API Base URL

```
https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1
```

## Available Endpoints

### Players

#### List Players
```
GET /api/v1/players
```

**Query Parameters:**
- `limit` (optional, default: 50, max: 100) - Number of results per page
- `offset` (optional, default: 0) - Pagination offset
- `position` (optional) - Filter by position (QB, RB, WR, TE, etc.)
- `team` (optional) - Filter by team UUID
- `status` (optional) - Filter by status (active, injured, inactive)

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players?position=QB&limit=10"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "nfl_id": 3139477,
      "name": "Patrick Mahomes",
      "position": "QB",
      "team_id": "uuid",
      "jersey_number": 15,
      "height_inches": 75,
      "weight_pounds": 230,
      "status": "active",
      "created_at": "2025-09-30T12:00:00Z",
      "updated_at": "2025-09-30T12:00:00Z"
    }
  ],
  "meta": {
    "total": 100,
    "limit": 10,
    "offset": 0,
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

#### Get Single Player
```
GET /api/v1/players/:id
```

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players/uuid"
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "name": "Patrick Mahomes",
    "position": "QB",
    ...
  },
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

### Teams

#### List All Teams
```
GET /api/v1/teams
```

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/teams"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "nfl_id": 12,
      "name": "Chiefs",
      "abbreviation": "KC",
      "city": "Kansas City",
      "conference": "AFC",
      "division": "West",
      "stadium": "GEHA Field at Arrowhead Stadium",
      "created_at": "2025-09-30T12:00:00Z",
      "updated_at": "2025-09-30T12:00:00Z"
    }
  ],
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

#### Get Single Team
```
GET /api/v1/teams/:id
```

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/teams/uuid"
```

#### Get Team Roster
```
GET /api/v1/teams/:id/players
```

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/teams/uuid/players"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Patrick Mahomes",
      "position": "QB",
      "jersey_number": 15,
      ...
    }
  ],
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

### Games

#### List Games
```
GET /api/v1/games
```

**Query Parameters:**
- `limit` (optional, default: 50, max: 100) - Number of results per page
- `offset` (optional, default: 0) - Pagination offset
- `season` (optional) - Filter by season year (e.g., 2024)
- `week` (optional) - Filter by week number
- `team` (optional) - Filter by team UUID
- `status` (optional) - Filter by status (scheduled, in_progress, final)

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games?season=2024&week=1"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "home_team_id": "uuid",
      "away_team_id": "uuid",
      "game_date": "2024-09-08T20:15:00Z",
      "season": 2024,
      "week": 1,
      "status": "final",
      "home_score": 27,
      "away_score": 20
    }
  ],
  "meta": {
    "total": 16,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

#### Get Single Game
```
GET /api/v1/games/:id
```

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games/uuid"
```

### Statistics

#### Get Game Statistics
```
GET /api/v1/games/:gameID/stats
```

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games/uuid/stats"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "game_id": "uuid",
      "player_id": "uuid",
      "passing_yards": 350,
      "passing_tds": 3,
      "rushing_yards": 45,
      "rushing_tds": 1,
      "receptions": 0,
      "receiving_yards": 0,
      "receiving_tds": 0
    }
  ],
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

#### Get Player Statistics
```
GET /api/v1/players/:playerID/stats
```

**Query Parameters:**
- `season` (optional) - Filter by season year
- `week` (optional) - Filter by week number
- `limit` (optional, default: 50) - Number of results

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players/uuid/stats?season=2024"
```

#### Get Stats Leaders
```
GET /api/v1/stats/leaders
```

**Query Parameters:**
- `category` (required) - Stat category (passing_yards, rushing_yards, receiving_yards, passing_tds, rushing_tds, receiving_tds)
- `season` (optional) - Filter by season year
- `limit` (optional, default: 10) - Number of results

**Example:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/stats/leaders?category=passing_yards&season=2024&limit=10"
```

**Response:**
```json
{
  "data": [
    {
      "player_id": "uuid",
      "player_name": "Patrick Mahomes",
      "position": "QB",
      "team_id": "uuid",
      "total_value": 4839,
      "games_played": 17
    }
  ],
  "meta": {
    "category": "passing_yards",
    "season": 2024,
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

### AI Endpoints

**Note:** AI endpoints require API key authentication via `X-API-Key` header or `Authorization: Bearer` token.

#### Predict Game Outcome
```
POST /api/v1/ai/predict/game/:gameID
```

**Example:**
```bash
curl -X POST "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/ai/predict/game/uuid" \
  -H "X-API-Key: your-api-key"
```

**Response:**
```json
{
  "data": {
    "game_id": "uuid",
    "home_team": "Chiefs",
    "away_team": "Bills",
    "prediction": {
      "winner": "Chiefs",
      "predicted_score": {"home": 31, "away": 27},
      "confidence": 75,
      "key_factors": ["Home field advantage", "QB matchup"],
      "analysis": "Detailed analysis..."
    },
    "generated_at": "2025-09-30T12:00:00Z"
  }
}
```

#### Predict Player Performance
```
POST /api/v1/ai/predict/player/:playerID
```

**Example:**
```bash
curl -X POST "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/ai/predict/player/uuid" \
  -H "X-API-Key: your-api-key"
```

#### Analyze Player
```
POST /api/v1/ai/insights/player/:playerID
```

**Example:**
```bash
curl -X POST "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/ai/insights/player/uuid" \
  -H "X-API-Key: your-api-key"
```

#### AI Query
```
POST /api/v1/ai/query
```

**Example:**
```bash
curl -X POST "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/ai/query" \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{"query": "Who are the top 5 quarterbacks this season?"}'
```

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Player not found",
    "status": 404
  }
}
```

**Common Error Codes:**
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing or invalid API key)
- `404` - Not Found (resource doesn't exist)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error
- `503` - Service Unavailable (AI service not configured)

## Authentication and Rate Limiting

### API Keys

API keys are required for AI endpoints. Include your API key in requests using either:

**X-API-Key Header:**
```bash
curl -H "X-API-Key: your-api-key" https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/ai/query
```

**Authorization Bearer Token:**
```bash
curl -H "Authorization: Bearer your-api-key" https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/ai/query
```

### Rate Limits

All endpoints have rate limiting enabled:

**Standard Endpoints** (Players, Teams, Games, Stats):
- **Limit:** 100 requests per minute
- **Headers:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

**AI Endpoints:**
- **Limit:** 10 requests per minute (with API key required)
- **Headers:** `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

**Unlimited Access:**
- Special API keys can be generated with unlimited access
- Unlimited keys bypass all rate limits
- Shows `X-RateLimit-Limit: unlimited` in response headers

### Generating API Keys

Admin endpoint to generate API keys:

```bash
# Generate standard API key
curl -X POST "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/keys/generate" \
  -H "Content-Type: application/json" \
  -d '{"unlimited": false, "label": "production-key"}'

# Generate unlimited API key
curl -X POST "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/keys/generate" \
  -H "Content-Type: application/json" \
  -d '{"unlimited": true, "label": "internal-unlimited"}'
```

**Response:**
```json
{
  "data": {
    "api_key": "gim_683a2f8c97767b9b4cf5403c6cfecca7174938be2f884625f09c0f6630f890b3",
    "type": "unlimited",
    "label": "internal-unlimited",
    "unlimited": true,
    "message": "API key generated successfully. Store this key securely - it cannot be retrieved again."
  }
}
```

## Caching

Redis-based caching is implemented for optimal performance:

**Cache TTLs:**
- **Teams:** 1 hour
- **Players:** 15 minutes
- **Games:** 5 minutes
- **Stats:** 5 minutes
- **Stats Leaders:** 10 minutes
- **AI Predictions (Game):** 15 minutes
- **AI Predictions (Player):** 30 minutes
- **AI Analysis:** 1 hour

**Cache Headers:**
- `X-Cache: HIT` - Response served from cache
- `X-Cache: MISS` - Response fetched from database

## Environment Variables

Required environment variables for deployment:

```env
# Database (required)
DATABASE_URL=postgres://user:password@host:5432/gridironmind
DB_MAX_CONNS=25
DB_MIN_CONNS=5

# Redis (optional - caching disabled if not set)
REDIS_URL=redis://host:6379

# API Keys (optional - dev mode if not set)
API_KEY=your-standard-api-key
UNLIMITED_API_KEY=your-unlimited-api-key

# AI Integration (optional - AI endpoints disabled if not set)
CLAUDE_API_KEY=your-claude-api-key

# Environment
ENVIRONMENT=production
```

**Note:**
- When `API_KEY` is not set, authentication is bypassed (development mode)
- When `UNLIMITED_API_KEY` is set, that specific key gets unlimited rate limit access
- When `REDIS_URL` is not set, caching is disabled and rate limiting allows all requests
- When `CLAUDE_API_KEY` is not set, AI endpoints return 503 Service Unavailable

## Dashboard

A simple web dashboard is included for testing the API and viewing data:

```
http://localhost:8000/dashboard/
```

**Features:**
- Browse and search players with filters
- View team rosters
- Interactive API endpoint tester
- Dark mode support
- Response time tracking

See [dashboard/README.md](dashboard/README.md) for detailed instructions.

## Local Development

1. Install Go 1.21+
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up PostgreSQL database:
   ```bash
   psql $DATABASE_URL -f schema.sql
   ```

4. Create `.env` file with required variables

5. Run the server:
   ```bash
   go run cmd/server/main.go
   # Server runs on http://localhost:8080
   ```

6. Open the dashboard at `http://localhost:8080`

## Deployment

### Heroku (Recommended)

```bash
# Create Heroku app
heroku create gridironmind

# Add addons
heroku addons:create heroku-postgresql:essential-0
heroku addons:create heroku-redis:mini

# Set environment variables
heroku config:set ENVIRONMENT=production
heroku config:set API_KEY=your-key
heroku config:set CLAUDE_API_KEY=your-key

# Deploy
git push heroku main

# Setup database
heroku pg:psql < schema.sql
```

See [HEROKU_DEPLOY.md](HEROKU_DEPLOY.md) for complete deployment guide.

## Database Schema

Run the schema to create all tables:
```bash
psql $DATABASE_URL -f schema.sql
```

This creates:
- `teams` - NFL team data
- `players` - Player profiles
- `games` - Game schedule and scores
- `game_stats` - Per-game player statistics
- `predictions` - AI predictions
- `ai_analysis` - AI analysis results

## Architecture

- **Backend:** Go 1.21+
- **Database:** PostgreSQL with pgx/v5
- **Cache:** Redis with go-redis/v8
- **AI:** Claude 3.5 Sonnet via Anthropic API
- **Deployment:** Heroku
- **API Format:** REST with JSON

## Project Status

**Phase 1: Foundation Setup** ✅ Complete
- Project structure created
- Database schema implemented
- Base models and utilities built

**Phase 2: Basic API Endpoints** ✅ Complete
- Players endpoints (list, get)
- Teams endpoints (list, get, roster)
- Database query layer
- CORS middleware
- Error handling

**Phase 2.5: Testing Dashboard** ✅ Complete
- Interactive web dashboard
- Player browser with search/filters
- Team roster viewer
- API endpoint tester
- Dark mode and caching

**Phase 3: Data Ingestion Service** ✅ Complete
- ESPN API integration
- Teams sync endpoint
- Rosters sync endpoint
- Games sync endpoint
- Full sync orchestration
- Admin endpoints for data management

**Phase 4: Game Stats and Advanced Endpoints** ✅ Complete
- Games endpoints (list, get) with filters
- Game statistics endpoints
- Player statistics endpoints
- Stats leaders endpoint
- Database query layer for stats

**Phase 5: Caching and Performance** ✅ Complete
- Redis integration
- Cache wrapper with TTLs
- Endpoint caching (teams, players, games, stats)
- Cache invalidation on data sync
- X-Cache headers (HIT/MISS)

**Phase 6: AI Integration** ✅ Complete
- Claude API client
- Game outcome predictions
- Player performance predictions
- Player analysis and insights
- General AI query endpoint
- AI response caching

**Phase 7: Authentication and Rate Limiting** ✅ Complete
- API key authentication middleware
- Redis-based rate limiting
- Standard rate limits (100/min)
- Strict rate limits for AI (10/min)
- Rate limit headers
- Unlimited API key support

**Phase 8: Production Deployment** ✅ Complete
- Deployed to Heroku
- PostgreSQL database configured
- Redis caching enabled
- Environment variables configured
- Database migrations automated
- API fully operational at https://grid-iron-mind-71cc9734eaf4.herokuapp.com/

## License

MIT