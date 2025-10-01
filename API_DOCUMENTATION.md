# Grid Iron Mind API Documentation

**Version:** 1.0.0
**Base URL:** `https://nfl.wearemachina.com/api/v1`
**Local Development:** `http://localhost:8080/api/v1`

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Rate Limiting](#rate-limiting)
4. [Response Format](#response-format)
5. [Endpoints](#endpoints)
   - [Players](#players-endpoints)
   - [Teams](#teams-endpoints)
   - [Games](#games-endpoints)
   - [Statistics](#statistics-endpoints)
   - [Defensive Stats](#defensive-statistics-endpoints)
   - [Injuries](#injury-endpoints)
   - [Career Stats](#career-endpoints)
   - [Weather](#weather-endpoints)
   - [AI Features](#ai-endpoints)
   - [AI Data Garden](#ai-data-garden-endpoints)
   - [Admin](#admin-endpoints)
6. [Data Models](#data-models)
7. [Error Codes](#error-codes)
8. [Auto-Fetch System](#auto-fetch-system)

---

## Overview

Grid Iron Mind is a high-performance NFL data API providing:
- ⚡ Sub-200ms response times for database queries
- 📊 Real-time and historical NFL statistics
- 🏈 Complete player, team, and game data
- 🤖 AI-powered predictions and insights
- 🌡️ Weather data and game conditions
- 🏥 Injury reports and tracking
- 📈 Fantasy football analytics
- 🎯 Defensive matchup analysis

### Key Features

- **Auto-Fetch System**: Automatically retrieves missing data on-demand
- **Multi-Provider AI**: Claude 3.5 Sonnet and Grok with automatic fallback
- **Aggressive Caching**: Redis-backed caching with endpoint-specific TTLs
- **Comprehensive Data**: 2020-2025 seasons with ongoing updates

---

## Authentication

### Public Endpoints
Most endpoints are publicly accessible without authentication:
- Players, Teams, Games
- Statistics, Weather
- Defensive stats, Injuries

### Protected Endpoints (API Key Required)
AI-powered endpoints require authentication via API key:

**Header Format:**
```http
X-API-Key: your_api_key_here
```

**Alternative (Bearer Token):**
```http
Authorization: Bearer your_api_key_here
```

**Endpoints Requiring Authentication:**
- `/api/v1/ai/*` - All AI prediction and analysis endpoints
- `/api/v1/garden/*` - AI Data Garden endpoints (except `/health`)

**Getting an API Key:**
```http
POST /api/v1/admin/keys/generate
Content-Type: application/json

{
  "description": "My application",
  "tier": "standard"
}
```

---

## Rate Limiting

### Standard Rate Limit
**Public Endpoints:** 100 requests/minute

### Strict Rate Limit
**AI Endpoints:** 10 requests/minute

### Rate Limit Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1696176000
```

**429 Response:**
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "status": 429
  }
}
```

---

## Response Format

### Success Response (Single Resource)
```json
{
  "data": {
    "id": "uuid",
    "name": "Patrick Mahomes",
    ...
  },
  "meta": {
    "timestamp": "2025-10-01T12:00:00Z"
  }
}
```

### Success Response (Collection with Pagination)
```json
{
  "data": [
    { "id": "uuid", ... },
    { "id": "uuid", ... }
  ],
  "meta": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-10-01T12:00:00Z"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Player not found",
    "status": 404
  }
}
```

### Special Headers
- `X-Cache: HIT|MISS` - Indicates cache status
- `X-Auto-Fetched: true` - Data was just auto-fetched
- `X-AI-Provider: claude|grok` - Which AI provider was used

---

## Endpoints

## Players Endpoints

### List Players
```http
GET /api/v1/players
```

**Query Parameters:**
- `limit` (int, optional) - Results per page (default: 50, max: 100)
- `offset` (int, optional) - Pagination offset (default: 0)
- `position` (string, optional) - Filter by position (QB, RB, WR, TE, etc.)
- `status` (string, optional) - Filter by status (active, injured, inactive)
- `team` (uuid, optional) - Filter by team ID

**Example:**
```http
GET /api/v1/players?position=QB&limit=20&status=active
```

**Response:**
```json
{
  "data": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "nfl_id": 4046838,
      "name": "Patrick Mahomes",
      "position": "QB",
      "team_id": "c1d6c4d6-3a8d-4e48-9c3a-8f7e6d5c4b3a",
      "jersey_number": 15,
      "height_inches": 75,
      "weight_pounds": 225,
      "birth_date": "1995-09-17T00:00:00Z",
      "college": "Texas Tech",
      "draft_year": 2017,
      "draft_round": 1,
      "draft_pick": 10,
      "status": "active",
      "headshot_url": "https://...",
      "created_at": "2024-09-30T12:00:00Z",
      "updated_at": "2024-10-01T12:00:00Z"
    }
  ],
  "meta": {
    "total": 1875,
    "limit": 20,
    "offset": 0,
    "timestamp": "2025-10-01T12:00:00Z"
  }
}
```

---

### Get Single Player
```http
GET /api/v1/players/:id
```

**Path Parameters:**
- `id` (uuid, required) - Player UUID

**Example:**
```http
GET /api/v1/players/a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**Response:** Single player object (same structure as list item)

---

### Get Player Career Stats
```http
GET /api/v1/players/:id/career
```

**Response:**
```json
{
  "data": {
    "player_id": "uuid",
    "total_seasons": 7,
    "career_stats": [
      {
        "season": 2024,
        "team_name": "Kansas City Chiefs",
        "games_played": 17,
        "passing_yards": 4839,
        "passing_tds": 41,
        "rushing_yards": 284,
        "rushing_tds": 4,
        "receptions": 0,
        "receiving_yards": 0,
        "receiving_tds": 0
      }
    ],
    "team_history": [
      {
        "team_name": "Kansas City Chiefs",
        "position": "QB",
        "season_start": 2017,
        "season_end": null,
        "is_current": true
      }
    ]
  }
}
```

---

### Get Player Team History
```http
GET /api/v1/players/:id/history
```

**Response:**
```json
{
  "data": [
    {
      "player_id": "uuid",
      "team_id": "uuid",
      "team_name": "Kansas City Chiefs",
      "team_abbreviation": "KC",
      "position": "QB",
      "season_start": 2017,
      "season_end": null,
      "is_current": true
    }
  ]
}
```

---

### Get Player Injuries
```http
GET /api/v1/players/:id/injuries
```

**Response:**
```json
{
  "data": {
    "player_id": "uuid",
    "injuries": [
      {
        "id": "uuid",
        "player_id": "uuid",
        "injury_type": "Ankle",
        "status": "Questionable",
        "description": "High ankle sprain",
        "injury_date": "2024-10-01T00:00:00Z",
        "expected_return_date": "2024-10-15T00:00:00Z",
        "last_updated": "2024-10-01T12:00:00Z"
      }
    ],
    "count": 1
  }
}
```

---

### Get Player vs Defense Stats
```http
GET /api/v1/players/:playerId/vs-defense/:teamId
```

**Path Parameters:**
- `playerId` (uuid, required) - Player UUID
- `teamId` (uuid, required) - Defense team UUID

**Query Parameters:**
- `season` (int, optional) - Filter to specific season
- `limit` (int, optional) - Number of games (default: 5, max: 50)

**Example:**
```http
GET /api/v1/players/abc.../vs-defense/def...?season=2024&limit=10
```

**Response:**
```json
{
  "data": {
    "player_id": "uuid",
    "player_name": "Christian McCaffrey",
    "defense_team_id": "uuid",
    "defense_team_abbr": "SEA",
    "games": [
      {
        "id": "uuid",
        "season": 2024,
        "week": 5,
        "passing_yards": 0,
        "passing_tds": 0,
        "interceptions_thrown": 0,
        "rushing_yards": 87,
        "rushing_tds": 1,
        "receptions": 6,
        "receiving_yards": 54,
        "receiving_tds": 1,
        "fantasy_points_standard": 24.1,
        "fantasy_points_ppr": 30.1,
        "fantasy_points_half_ppr": 27.1
      }
    ],
    "averages": {
      "games_played": 5,
      "fantasy_points_per_game_standard": 22.4,
      "fantasy_points_per_game_ppr": 27.8,
      "fantasy_points_per_game_half_ppr": 25.1,
      "yards_per_game": 156.2,
      "touchdowns_per_game": 1.8
    }
  }
}
```

---

## Teams Endpoints

### List All Teams
```http
GET /api/v1/teams
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "nfl_id": 12,
      "name": "Kansas City Chiefs",
      "abbreviation": "KC",
      "city": "Kansas City",
      "conference": "AFC",
      "division": "West",
      "logo_url": "https://...",
      "stadium_name": "GEHA Field at Arrowhead Stadium",
      "stadium_city": "Kansas City",
      "stadium_state": "MO",
      "stadium_capacity": 76416,
      "stadium_surface": "Grass",
      "stadium_type": "outdoor",
      "stadium_lat": 39.0489,
      "stadium_lon": -94.4839,
      "created_at": "2024-09-30T12:00:00Z",
      "updated_at": "2024-10-01T12:00:00Z"
    }
  ],
  "meta": {
    "timestamp": "2025-10-01T12:00:00Z"
  }
}
```

---

### Get Single Team
```http
GET /api/v1/teams/:id
```

**Response:** Single team object

---

### Get Team Players (Roster)
```http
GET /api/v1/teams/:id/players
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
  ]
}
```

---

### Get Team Injuries
```http
GET /api/v1/teams/:id/injuries
```

**Response:**
```json
{
  "data": {
    "team_id": "uuid",
    "injuries_by_status": {
      "Out": [...],
      "Doubtful": [...],
      "Questionable": [...],
      "Probable": [...]
    },
    "total_count": 8
  }
}
```

---

### Get Team Defensive Stats
```http
GET /api/v1/teams/:teamId/defense/stats
```

**Query Parameters:**
- `season` (int, required) - NFL season year
- `week` (int, optional) - Specific week (omit for season-long)

**Example:**
```http
GET /api/v1/teams/abc.../defense/stats?season=2024&week=5
```

**Response:**
```json
{
  "data": {
    "id": "uuid",
    "team_id": "uuid",
    "team_abbr": "SF",
    "team_name": "San Francisco 49ers",
    "season": 2024,
    "week": 5,
    "points_allowed": 78,
    "points_allowed_per_game": 15.6,
    "yards_allowed": 1567,
    "yards_allowed_per_game": 313.4,
    "pass_yards_allowed": 1023,
    "rush_yards_allowed": 544,
    "sacks": 18,
    "interceptions": 6,
    "forced_fumbles": 4,
    "third_down_percentage": 36.92,
    "red_zone_percentage": 53.33,
    "defensive_rank": 3,
    "pass_defense_rank": 5,
    "rush_defense_rank": 2,
    "games_played": 5
  }
}
```

---

## Games Endpoints

### List Games
```http
GET /api/v1/games
```

**Query Parameters:**
- `season` (int, optional) - Filter by season (e.g., 2024)
- `week` (int, optional) - Filter by week (1-18)
- `team` (uuid, optional) - Filter by team ID
- `status` (string, optional) - Filter by status (scheduled, in_progress, final)
- `limit` (int, optional) - Results per page (default: 50)
- `offset` (int, optional) - Pagination offset

**Example:**
```http
GET /api/v1/games?season=2024&week=5
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "nfl_game_id": "401671654",
      "season": 2024,
      "week": 5,
      "game_type": "regular",
      "home_team_id": "uuid",
      "away_team_id": "uuid",
      "home_team_name": "Kansas City Chiefs",
      "away_team_name": "Las Vegas Raiders",
      "home_team_abbr": "KC",
      "away_team_abbr": "LV",
      "home_score": 27,
      "away_score": 20,
      "status": "final",
      "game_date": "2024-10-01T20:15:00Z",
      "venue_name": "GEHA Field at Arrowhead Stadium",
      "venue_city": "Kansas City",
      "venue_state": "MO",
      "venue_type": "outdoor",
      "weather_temp": 72,
      "weather_condition": "Clear",
      "weather_wind_speed": 8,
      "weather_humidity": 45,
      "created_at": "2024-09-30T12:00:00Z",
      "updated_at": "2024-10-01T23:30:00Z"
    }
  ],
  "meta": {
    "total": 16,
    "limit": 50,
    "offset": 0
  }
}
```

---

### Get Single Game
```http
GET /api/v1/games/:id
```

**Response:** Single game object

---

### Get Game Stats
```http
GET /api/v1/stats/game/:gameId
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "player_id": "uuid",
      "game_id": "uuid",
      "player": {
        "name": "Patrick Mahomes",
        "position": "QB"
      },
      "passing_attempts": 35,
      "passing_completions": 25,
      "passing_yards": 291,
      "passing_touchdowns": 2,
      "passing_interceptions": 0,
      "rushing_attempts": 3,
      "rushing_yards": 14,
      "rushing_touchdowns": 0,
      "receiving_receptions": 0,
      "receiving_yards": 0,
      "receiving_touchdowns": 0,
      "receiving_targets": 0
    }
  ]
}
```

---

## Statistics Endpoints

### Get Stat Leaders
```http
GET /api/v1/stats/leaders
```

**Query Parameters:**
- `season` (int, required) - Season year
- `category` (string, required) - Stat category
  - `passing_yards`, `passing_tds`, `rushing_yards`, `rushing_tds`
  - `receiving_yards`, `receiving_tds`, `receptions`
- `limit` (int, optional) - Number of leaders (default: 10, max: 50)

**Example:**
```http
GET /api/v1/stats/leaders?season=2024&category=passing_yards&limit=10
```

**Response:**
```json
{
  "data": [
    {
      "rank": 1,
      "player_id": "uuid",
      "player_name": "Patrick Mahomes",
      "team_name": "Kansas City Chiefs",
      "position": "QB",
      "value": 4839,
      "games_played": 17
    }
  ]
}
```

---

## Defensive Statistics Endpoints

### Get Defensive Rankings
```http
GET /api/v1/defense/rankings
```

**Query Parameters:**
- `season` (int, required) - NFL season year
- `category` (string, optional) - Ranking category (default: `overall`)
  - `overall` - Total yards allowed per game
  - `pass` - Pass yards allowed per game
  - `rush` - Rush yards allowed per game
  - `points_allowed` - Points allowed per game

**Example:**
```http
GET /api/v1/defense/rankings?season=2024&category=pass
```

**Response:**
```json
{
  "data": [
    {
      "rank": 1,
      "team_id": "uuid",
      "team_abbr": "BAL",
      "team_name": "Baltimore Ravens",
      "category": "pass",
      "value": 189.3,
      "season": 2024
    }
  ]
}
```

---

## Injury Endpoints

### Get All Injuries (League-Wide)
```http
GET /api/v1/injuries
```

**Query Parameters:**
- `status` (string, optional) - Filter by status (Out, Doubtful, Questionable, Probable)
- `position` (string, optional) - Filter by position

---

## Career Endpoints

Covered under Players endpoints:
- `GET /api/v1/players/:id/career` - Full career stats
- `GET /api/v1/players/:id/history` - Team history

---

## Weather Endpoints

### Get Current Weather
```http
GET /api/v1/weather/current
```

**Query Parameters (one required):**
- `location` (string) - City, State (e.g., "Kansas City, MO")
- `lat` + `lon` (float) - Latitude and longitude

**Example:**
```http
GET /api/v1/weather/current?location=Kansas+City,MO
GET /api/v1/weather/current?lat=39.0489&lon=-94.4839
```

**Response:**
```json
{
  "data": {
    "temperature": 72.5,
    "condition": "Clear",
    "wind_speed": 8.2,
    "wind_direction": "S",
    "humidity": 45,
    "pressure": 1013.25,
    "visibility": 10.0,
    "location": "Kansas City, MO",
    "timestamp": "2025-10-01T14:30:00Z"
  }
}
```

---

### Get Historical Weather
```http
GET /api/v1/weather/historical
```

**Query Parameters:**
- `location` OR `lat`+`lon` (required)
- `date` (string, required) - Format: YYYY-MM-DD

**Example:**
```http
GET /api/v1/weather/historical?location=Kansas+City,MO&date=2024-09-15
```

---

### Get Weather Forecast
```http
GET /api/v1/weather/forecast
```

**Query Parameters:**
- `location` OR `lat`+`lon` (required)
- `days` (int, optional) - Forecast days (default: 7, max: 14)

---

## AI Endpoints

**⚠️ Requires API Key Authentication**

### Predict Game Outcome
```http
POST /api/v1/ai/predict/game/:gameId
```

**Headers:**
```http
X-API-Key: your_api_key
Content-Type: application/json
```

**Response:**
```json
{
  "data": {
    "game_id": "uuid",
    "home_team": "Kansas City Chiefs",
    "away_team": "Las Vegas Raiders",
    "prediction": {
      "winner": "Kansas City Chiefs",
      "confidence": 78.5,
      "predicted_score": {
        "home": 27,
        "away": 20
      },
      "key_factors": [
        "Home field advantage",
        "QB performance differential",
        "Defensive matchup favorability"
      ],
      "analysis": "The Chiefs are favored due to..."
    },
    "ai_provider": "claude",
    "generated_at": "2025-10-01T12:00:00Z"
  }
}
```

---

### Predict Player Performance
```http
POST /api/v1/ai/predict/player/:playerId
```

**Response:**
```json
{
  "data": {
    "player_id": "uuid",
    "player_name": "Patrick Mahomes",
    "position": "QB",
    "prediction": {
      "passing_yards": {
        "predicted": 285,
        "range": [240, 330],
        "confidence": 72
      },
      "passing_tds": {
        "predicted": 2,
        "range": [1, 3],
        "confidence": 68
      },
      "fantasy_points": {
        "standard": 22.4,
        "ppr": 22.4,
        "confidence": 70
      },
      "analysis": "Based on matchup history and defensive rankings..."
    },
    "ai_provider": "grok",
    "generated_at": "2025-10-01T12:00:00Z"
  }
}
```

---

### Analyze Player
```http
POST /api/v1/ai/insights/player/:playerId
```

**Response:**
```json
{
  "data": {
    "player_id": "uuid",
    "player_name": "Patrick Mahomes",
    "analysis": {
      "strengths": [
        "Exceptional arm strength",
        "Elite decision making under pressure",
        "Mobility extends plays"
      ],
      "weaknesses": [
        "Occasional overconfidence leads to turnovers"
      ],
      "recent_performance": "Outstanding 4-game stretch with 12 TDs...",
      "injury_concerns": "No current concerns",
      "fantasy_outlook": "Top 3 QB ROS with elite upside",
      "matchup_analysis": "Favorable schedule ahead..."
    },
    "ai_provider": "claude"
  }
}
```

---

### Natural Language Query
```http
POST /api/v1/ai/query
```

**Request Body:**
```json
{
  "query": "Who are the top 5 running backs this season?"
}
```

**Response:**
```json
{
  "data": {
    "query": "Who are the top 5 running backs this season?",
    "answer": "The top 5 running backs in 2024 by total yards are:\n1. Christian McCaffrey (SF) - 1,459 yards\n2. Derrick Henry (BAL) - 1,367 yards...",
    "confidence": 95,
    "sources": ["player_season_stats", "games"],
    "ai_provider": "claude"
  }
}
```

---

## AI Data Garden Endpoints

**🌱 Self-Healing Database System**

### Health Check
```http
GET /api/v1/garden/health
```

**Public - No API Key Required**

**Response:**
```json
{
  "data": {
    "health_report": {
      "overall_health": "healthy",
      "checked_at": "2025-10-01T12:00:00Z",
      "issues": [],
      "recommendations": [
        "Consider syncing week 6 data",
        "Some player stats missing for recent games"
      ],
      "data_freshness": {
        "players": "2024-10-01",
        "games": "2024-10-01",
        "stats": "2024-09-30"
      }
    }
  }
}
```

---

### Health Check with Auto-Heal
```http
POST /api/v1/garden/health
```

**Requires API Key**

Runs health check and automatically fixes detected issues.

---

### Natural Language Data Query
```http
POST /api/v1/garden/query
```

**Requires API Key**

**Request:**
```json
{
  "query": "Show me all games where the weather was below 32 degrees"
}
```

**Response:**
```json
{
  "data": {
    "query": "Show me all games where the weather was below 32 degrees",
    "sql": "SELECT * FROM games WHERE weather_temp < 32",
    "results": [...],
    "explanation": "Found 47 games played in freezing conditions"
  }
}
```

---

### AI Data Enrichment
```http
POST /api/v1/garden/enrich/player/:playerId
```

**Requires API Key**

Enriches player data with AI-generated insights and analysis.

---

### Sync Schedule
```http
GET /api/v1/garden/schedule
```

**Response:**
```json
{
  "data": {
    "next_sync": "2025-10-02T03:00:00Z",
    "last_sync": "2025-10-01T03:00:00Z",
    "sync_status": "healthy",
    "upcoming_tasks": [
      {
        "task": "sync_week_6_games",
        "scheduled_at": "2025-10-02T03:00:00Z"
      }
    ]
  }
}
```

---

### Garden Status
```http
GET /api/v1/garden/status
```

**Response:**
```json
{
  "data": {
    "is_healthy": true,
    "ai_available": true,
    "database_connected": true,
    "cache_available": true,
    "auto_heal_enabled": true,
    "last_health_check": "2025-10-01T12:00:00Z"
  }
}
```

---

## Admin Endpoints

### Sync Teams
```http
POST /api/v1/admin/sync/teams
```

Fetches and syncs all 32 NFL teams from ESPN API.

---

### Sync Rosters
```http
POST /api/v1/admin/sync/rosters
```

Fetches current rosters for all teams.

---

### Sync Games
```http
POST /api/v1/admin/sync/games
```

**Query Parameters:**
- `season` (int, required)
- `week` (int, required)

---

### Full Sync
```http
POST /api/v1/admin/sync/full
```

Performs complete sync: Teams → Rosters → Games → Stats

---

### Sync Historical Season
```http
POST /api/v1/admin/sync/historical/season
```

**Query Parameters:**
- `year` (int, required)

---

### Sync NFLverse Stats
```http
POST /api/v1/admin/sync/nflverse/stats
```

**Query Parameters:**
- `season` (int, required)

---

### Sync Injuries
```http
POST /api/v1/admin/sync/injuries
```

Fetches latest injury reports.

---

### Generate API Key
```http
POST /api/v1/admin/keys/generate
```

**Request:**
```json
{
  "description": "My App",
  "tier": "standard"
}
```

**Response:**
```json
{
  "data": {
    "api_key": "gim_1234567890abcdef...",
    "description": "My App",
    "tier": "standard",
    "created_at": "2025-10-01T12:00:00Z"
  }
}
```

---

## Data Models

### Player
```typescript
{
  id: UUID
  nfl_id: number
  name: string
  position: string
  team_id: UUID | null
  jersey_number: number | null
  height_inches: number | null
  weight_pounds: number | null
  birth_date: Date | null
  college: string | null
  draft_year: number | null
  draft_round: number | null
  draft_pick: number | null
  status: string // "active" | "injured" | "inactive"
  headshot_url: string | null
  created_at: Date
  updated_at: Date
}
```

### Team
```typescript
{
  id: UUID
  nfl_id: number
  name: string
  abbreviation: string
  city: string
  conference: string // "AFC" | "NFC"
  division: string // "North" | "South" | "East" | "West"
  logo_url: string
  stadium_name: string
  stadium_city: string
  stadium_state: string
  stadium_capacity: number
  stadium_surface: string
  stadium_type: string // "outdoor" | "indoor" | "retractable"
  stadium_lat: number
  stadium_lon: number
  created_at: Date
  updated_at: Date
}
```

### Game
```typescript
{
  id: UUID
  nfl_game_id: string
  season: number
  week: number
  game_type: string // "regular" | "playoff" | "preseason"
  home_team_id: UUID
  away_team_id: UUID
  home_score: number | null
  away_score: number | null
  status: string // "scheduled" | "in_progress" | "final"
  game_date: Date
  venue_name: string
  venue_city: string
  venue_state: string
  venue_type: string
  weather_temp: number | null
  weather_condition: string | null
  weather_wind_speed: number | null
  weather_humidity: number | null
  created_at: Date
  updated_at: Date
}
```

### Game Stats
```typescript
{
  id: UUID
  player_id: UUID
  game_id: UUID
  passing_attempts: number
  passing_completions: number
  passing_yards: number
  passing_touchdowns: number
  passing_interceptions: number
  rushing_attempts: number
  rushing_yards: number
  rushing_touchdowns: number
  receiving_receptions: number
  receiving_yards: number
  receiving_touchdowns: number
  receiving_targets: number
}
```

### Injury
```typescript
{
  id: UUID
  player_id: UUID
  injury_type: string
  status: string // "Out" | "Doubtful" | "Questionable" | "Probable"
  description: string
  injury_date: Date
  expected_return_date: Date | null
  last_updated: Date
}
```

---

## Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `BAD_REQUEST` | 400 | Invalid request parameters |
| `UNAUTHORIZED` | 401 | Missing or invalid API key |
| `NOT_FOUND` | 404 | Resource not found |
| `METHOD_NOT_ALLOWED` | 405 | HTTP method not allowed |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |
| `AI_UNAVAILABLE` | 503 | AI service not configured |

---

## Auto-Fetch System

The API includes intelligent auto-fetching that automatically retrieves missing data when queried.

### How It Works

1. You query an endpoint (e.g., `/api/v1/games?season=2025&week=1`)
2. If the database returns empty results:
   - System automatically fetches data from ESPN API
   - Stores it in the database
   - Returns the freshly fetched data
3. Response includes `X-Auto-Fetched: true` header

### Supported Auto-Fetch

- **Games**: Automatically fetches game schedules
- **Players**: Fetches rosters when individual player missing
- **Stats**: Fetches game statistics from NFLverse
- **Teams**: Ensures all 32 teams exist

### Example

```http
GET /api/v1/games?season=2025&week=1

Response Headers:
X-Auto-Fetched: true
X-Cache: MISS
```

The first request may take 2-10 seconds, but subsequent requests are served instantly from the database.

---

## Best Practices

### Pagination
Always use `limit` and `offset` for large datasets:
```http
GET /api/v1/players?limit=50&offset=0
```

### Caching
Check `X-Cache` header to monitor cache performance:
```http
X-Cache: HIT  # Served from cache (fast)
X-Cache: MISS # Queried database (slower)
```

### Rate Limits
Monitor rate limit headers to avoid 429 errors:
```http
X-RateLimit-Remaining: 95
```

### AI Endpoints
- Cache AI results client-side (responses rarely change)
- Use specific queries for better accuracy
- Monitor `ai_provider` field for provider transparency

---

## Support & Resources

- **API Base URL**: https://nfl.wearemachina.com/api/v1
- **Dashboard**: https://nfl.wearemachina.com
- **Health Check**: https://nfl.wearemachina.com/health
- **Documentation**: https://nfl.wearemachina.com/api-docs.html

---

**Last Updated:** October 1, 2025
**Version:** 1.0.0
