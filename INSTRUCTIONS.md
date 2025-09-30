# Grid Iron Mind API - Instructions

A comprehensive NFL data API providing real-time and historical statistics, player information, team data, game schedules, injury reports, and AI-powered insights.

**Base URL:** `https://nfl.wearemachina.com/api/v1`

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Public Endpoints](#public-endpoints)
4. [AI Endpoints (Requires API Key)](#ai-endpoints-requires-api-key)
5. [Admin Endpoints](#admin-endpoints)
6. [Response Format](#response-format)
7. [Error Handling](#error-handling)
8. [Rate Limits](#rate-limits)
9. [Examples](#examples)

---

## Getting Started

### Health Check

Verify the API is operational:

```bash
curl https://nfl.wearemachina.com/api/v1/health
```

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "service": "Grid Iron Mind API",
    "version": "1.0.0"
  }
}
```

### API Documentation

View the interactive API documentation:
```
https://nfl.wearemachina.com/api-docs.html
```

---

## Authentication

### Public Endpoints

Most endpoints are **publicly accessible** and do not require authentication:
- Teams, players, games, stats, injuries, weather

### AI Endpoints

AI-powered endpoints require an **API key** via the `X-API-Key` header.

#### Generate an API Key

**Standard API Key** (rate limited):
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/admin/keys/generate \
  -H "Content-Type: application/json" \
  -d '{
    "unlimited": false,
    "label": "My Application"
  }'
```

**Unlimited API Key** (no rate limits):
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/admin/keys/generate \
  -H "Content-Type: application/json" \
  -d '{
    "unlimited": true,
    "label": "Production Application"
  }'
```

**Response:**
```json
{
  "data": {
    "api_key": "gim_a1b2c3d4e5f6789...",
    "type": "unlimited",
    "label": "Production Application",
    "unlimited": true,
    "message": "API key generated successfully. Store this key securely - it cannot be retrieved again."
  }
}
```

⚠️ **Important:** Save the API key immediately. It cannot be retrieved after generation.

#### Using Your API Key

Include the API key in the `X-API-Key` header:

```bash
curl -X POST https://nfl.wearemachina.com/api/v1/ai/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: gim_a1b2c3d4e5f6789..." \
  -d '{
    "query": "Who are the top 5 quarterbacks this season?"
  }'
```

---

## Public Endpoints

### Teams

#### List All Teams
```bash
GET /api/v1/teams
```

**Example:**
```bash
curl https://nfl.wearemachina.com/api/v1/teams
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Buffalo Bills",
      "abbreviation": "BUF",
      "city": "Buffalo",
      "conference": "AFC",
      "division": "East",
      "stadium": "Highmark Stadium",
      "logo_url": "https://...",
      "nfl_id": 2
    }
  ],
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z"
  }
}
```

#### Get Single Team
```bash
GET /api/v1/teams/{id}
```

**Example:**
```bash
curl https://nfl.wearemachina.com/api/v1/teams/550e8400-e29b-41d4-a716-446655440000
```

#### Get Team Roster
```bash
GET /api/v1/teams/{id}/players
```

**Example:**
```bash
curl https://nfl.wearemachina.com/api/v1/teams/550e8400-e29b-41d4-a716-446655440000/players
```

#### Get Team Injuries
```bash
GET /api/v1/teams/{id}/injuries
```

**Example:**
```bash
curl https://nfl.wearemachina.com/api/v1/teams/550e8400-e29b-41d4-a716-446655440000/injuries
```

**Response:**
```json
{
  "data": {
    "team_id": "uuid",
    "team_name": "Buffalo Bills",
    "injuries": [
      {
        "id": "uuid",
        "player_id": "uuid",
        "status": "Questionable",
        "injury_type": "Ankle",
        "body_location": "Right Ankle",
        "return_date": "2025-10-01",
        "player": {
          "name": "Josh Allen",
          "position": "QB"
        }
      }
    ]
  }
}
```

---

### Players

#### List Players (with filters)
```bash
GET /api/v1/players?limit={limit}&offset={offset}&position={position}&team={team_id}&status={status}
```

**Query Parameters:**
- `limit` (default: 50, max: 100) - Number of results
- `offset` (default: 0) - Pagination offset
- `position` - Filter by position (QB, RB, WR, TE, etc.)
- `team` - Filter by team UUID
- `status` - Filter by status (active, injured, suspended)

**Example:**
```bash
# Get all quarterbacks
curl "https://nfl.wearemachina.com/api/v1/players?position=QB&limit=20"

# Get Bills roster
curl "https://nfl.wearemachina.com/api/v1/players?team=550e8400-e29b-41d4-a716-446655440000"
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Josh Allen",
      "position": "QB",
      "jersey_number": "17",
      "height": "6-5",
      "weight": 237,
      "age": 28,
      "experience": 6,
      "college": "Wyoming",
      "status": "active",
      "team_id": "uuid"
    }
  ],
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z",
    "total": 1843,
    "limit": 20,
    "offset": 0
  }
}
```

#### Get Player Details
```bash
GET /api/v1/players/{id}
```

#### Get Player Career Stats
```bash
GET /api/v1/players/{id}/career
```

**Example:**
```bash
curl https://nfl.wearemachina.com/api/v1/players/550e8400-e29b-41d4-a716-446655440000/career
```

**Response:**
```json
{
  "data": [
    {
      "season": 2024,
      "games_played": 17,
      "passing_yards": 4306,
      "passing_touchdowns": 29,
      "interceptions": 18,
      "rushing_yards": 762,
      "rushing_touchdowns": 15
    }
  ]
}
```

#### Get Player Team History
```bash
GET /api/v1/players/{id}/history
```

#### Get Player Injuries
```bash
GET /api/v1/players/{id}/injuries
```

**Example:**
```bash
curl https://nfl.wearemachina.com/api/v1/players/550e8400-e29b-41d4-a716-446655440000/injuries
```

---

### Games

#### List Games
```bash
GET /api/v1/games?season={year}&week={week}&team={team_id}
```

**Example:**
```bash
# Get all games for week 1 of 2024 season
curl "https://nfl.wearemachina.com/api/v1/games?season=2024&week=1"
```

#### Get Game Details
```bash
GET /api/v1/games/{id}
```

#### Get Game Stats
```bash
GET /api/v1/stats/game/{id}
```

---

### Stats

#### Get League Leaders
```bash
GET /api/v1/stats/leaders?category={category}&limit={limit}&season={year}
```

**Categories:**
- `passing` - Passing yards
- `rushing` - Rushing yards
- `receiving` - Receiving yards
- `touchdowns` - Total touchdowns
- `interceptions` - Interceptions thrown
- `sacks` - Sacks

**Example:**
```bash
# Top 10 passers
curl "https://nfl.wearemachina.com/api/v1/stats/leaders?category=passing&limit=10"

# Top rushers for 2023 season
curl "https://nfl.wearemachina.com/api/v1/stats/leaders?category=rushing&season=2023"
```

---

### Weather

#### Get Current Weather
```bash
GET /api/v1/weather/current?location={city}
```

**Example:**
```bash
curl "https://nfl.wearemachina.com/api/v1/weather/current?location=Buffalo,NY"
```

#### Get Weather Forecast
```bash
GET /api/v1/weather/forecast?location={city}&days={days}
```

#### Get Historical Weather
```bash
GET /api/v1/weather/historical?location={city}&date={YYYY-MM-DD}
```

---

## AI Endpoints (Requires API Key)

All AI endpoints require the `X-API-Key` header.

### Game Prediction
```bash
POST /api/v1/ai/predict/game/{game_id}
```

**Example:**
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/ai/predict/game/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: gim_your_api_key"
```

### Player Performance Prediction
```bash
POST /api/v1/ai/predict/player/{player_id}
```

**Request Body:**
```json
{
  "game_id": "uuid",
  "opponent_team_id": "uuid"
}
```

### Player Insights
```bash
POST /api/v1/ai/insights/player/{player_id}
```

**Example:**
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/ai/insights/player/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: gim_your_api_key" \
  -d '{
    "season": 2024,
    "analysis_type": "season_summary"
  }'
```

### Natural Language Query
```bash
POST /api/v1/ai/query
```

**Request Body:**
```json
{
  "query": "Who are the top 5 quarterbacks this season by passing yards?"
}
```

**Example:**
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/ai/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: gim_your_api_key" \
  -d '{
    "query": "Compare Josh Allen and Patrick Mahomes this season"
  }'
```

---

## Admin Endpoints

### Data Synchronization

These endpoints trigger background jobs to sync data from NFL sources.

#### Sync Teams
```bash
POST /api/v1/admin/sync/teams
```

#### Sync All Rosters
```bash
POST /api/v1/admin/sync/rosters
```

#### Sync Current Games
```bash
POST /api/v1/admin/sync/games
```

#### Sync Injury Reports
```bash
POST /api/v1/admin/sync/injuries
```

**Example:**
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/injuries
```

**Response:**
```json
{
  "data": {
    "message": "Injury reports sync started (running in background)",
    "status": "success"
  }
}
```

#### Full Sync (Teams → Rosters → Games)
```bash
POST /api/v1/admin/sync/full
```

#### Sync Historical Season
```bash
POST /api/v1/admin/sync/historical/season
```

**Request Body:**
```json
{
  "year": 2023
}
```

#### Sync Multiple Seasons
```bash
POST /api/v1/admin/sync/historical/seasons
```

**Request Body:**
```json
{
  "start_year": 2020,
  "end_year": 2024
}
```

#### Sync NFLverse Player Stats
```bash
POST /api/v1/admin/sync/nflverse/stats
```

**Request Body:**
```json
{
  "season": 2024
}
```

#### Sync NFLverse Schedule
```bash
POST /api/v1/admin/sync/nflverse/schedule
```

**Request Body:**
```json
{
  "season": 2024
}
```

#### Sync NFLverse Next Gen Stats
```bash
POST /api/v1/admin/sync/nflverse/nextgen
```

**Request Body:**
```json
{
  "season": 2024,
  "stat_type": "passing"
}
```

**Stat Types:** `passing`, `rushing`, `receiving`

#### Enrich Games with Weather Data
```bash
POST /api/v1/admin/sync/weather
```

**Request Body:**
```json
{
  "season": 2024
}
```

#### Sync Team Stats
```bash
POST /api/v1/admin/sync/team-stats
```

**Request Body:**
```json
{
  "season": 2024,
  "week": 1
}
```

---

## Response Format

All endpoints return JSON in a consistent format:

### Success Response
```json
{
  "data": {
    // Response data
  },
  "meta": {
    "timestamp": "2025-09-30T12:00:00Z",
    "total": 100,      // Total records (paginated endpoints)
    "limit": 50,       // Results per page
    "offset": 0        // Current offset
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

---

## Error Handling

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Missing or invalid API key |
| 404 | Not Found - Resource doesn't exist |
| 405 | Method Not Allowed - Wrong HTTP method |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |
| 503 | Service Unavailable - Database connection failed |

### Common Error Codes

- `INVALID_REQUEST` - Malformed request body
- `METHOD_NOT_ALLOWED` - Wrong HTTP method
- `NOT_FOUND` - Resource not found
- `UNHEALTHY` - Service unavailable
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INVALID_API_KEY` - API key invalid or missing
- `SYNC_FAILED` - Data synchronization failed

---

## Rate Limits

### Standard Endpoints (Public)
- **Rate Limit:** 100 requests per minute per IP
- **Headers:**
  - `X-RateLimit-Limit` - Max requests per window
  - `X-RateLimit-Remaining` - Remaining requests
  - `X-RateLimit-Reset` - Time when limit resets

### AI Endpoints (Standard API Key)
- **Rate Limit:** 10 requests per minute
- Stricter limits due to computational cost

### AI Endpoints (Unlimited API Key)
- **Rate Limit:** No limits
- For production applications with high volume

---

## Examples

### Complete Workflow: Get Player Stats

```bash
#!/bin/bash

API_BASE="https://nfl.wearemachina.com/api/v1"

# 1. Find Josh Allen by name (search Bills roster)
BILLS_ID=$(curl -s "$API_BASE/teams" | jq -r '.data[] | select(.abbreviation=="BUF") | .id')

# 2. Get Bills roster
JOSH_ALLEN_ID=$(curl -s "$API_BASE/teams/$BILLS_ID/players" | jq -r '.data[] | select(.name=="Josh Allen") | .id')

# 3. Get Josh Allen's career stats
curl -s "$API_BASE/players/$JOSH_ALLEN_ID/career" | jq '.'

# 4. Get Josh Allen's injury status
curl -s "$API_BASE/players/$JOSH_ALLEN_ID/injuries" | jq '.'
```

### Using AI Endpoints

```bash
#!/bin/bash

API_KEY="gim_your_api_key_here"
API_BASE="https://nfl.wearemachina.com/api/v1"

# Natural language query
curl -X POST "$API_BASE/ai/query" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "query": "Who are the top 5 running backs by rushing yards this season?"
  }' | jq '.'

# Get player insights
curl -X POST "$API_BASE/ai/insights/player/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "season": 2024,
    "analysis_type": "season_summary"
  }' | jq '.'
```

### Sync Data (Admin)

```bash
#!/bin/bash

API_BASE="https://nfl.wearemachina.com/api/v1"

# Sync current week's games
curl -X POST "$API_BASE/admin/sync/games"

# Sync injury reports
curl -X POST "$API_BASE/admin/sync/injuries"

# Sync historical season
curl -X POST "$API_BASE/admin/sync/historical/season" \
  -H "Content-Type: application/json" \
  -d '{"year": 2023}'

# Sync NFLverse stats
curl -X POST "$API_BASE/admin/sync/nflverse/stats" \
  -H "Content-Type: application/json" \
  -d '{"season": 2024}'
```

---

## Support

For issues, feature requests, or questions:
- **GitHub:** [gridironmind repository]
- **API Status:** Check `/api/v1/health` endpoint
- **Documentation:** https://nfl.wearemachina.com/api-docs.html

---

## Changelog

### v1.0.0 (2025-09-30)
- ✅ Complete team metadata (32 teams)
- ✅ Player career statistics (2010-2024, 2,257 records)
- ✅ Injury tracking system
- ✅ Weather integration
- ✅ AI-powered insights
- ✅ NFLverse data enrichment
- ✅ Historical game data
- ✅ API key management
