# Grid Iron Mind API

High-performance NFL data lake with AI-enriched insights via REST API.

## API Base URL

```
https://gridironmind.vercel.app/api/v1
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
curl "https://gridironmind.vercel.app/api/v1/players?position=QB&limit=10"
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
curl "https://gridironmind.vercel.app/api/v1/players/uuid"
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
curl "https://gridironmind.vercel.app/api/v1/teams"
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
curl "https://gridironmind.vercel.app/api/v1/teams/uuid"
```

#### Get Team Roster
```
GET /api/v1/teams/:id/players
```

**Example:**
```bash
curl "https://gridironmind.vercel.app/api/v1/teams/uuid/players"
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
- `404` - Not Found (resource doesn't exist)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

## Environment Variables

Required environment variables for deployment:

```env
DATABASE_URL=postgres://user:password@host:5432/gridironmind
REDIS_URL=redis://host:6379
API_KEY=your-api-key
ENVIRONMENT=production
CLAUDE_API_KEY=your-claude-api-key
DB_MAX_CONNS=25
DB_MIN_CONNS=5
```

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
- **Cache:** Redis (to be implemented in Phase 5)
- **Deployment:** Vercel Serverless Functions
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

**Phase 3: Data Ingestion Service** 🔄 Next
**Phase 4: Game Stats and Advanced Endpoints** 📋 Planned
**Phase 5: Caching and Performance** 📋 Planned
**Phase 6: AI Integration** 📋 Planned
**Phase 7: Authentication and Rate Limiting** 📋 Planned
**Phase 8: Production Deployment** 📋 Planned

## License

MIT