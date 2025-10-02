# Grid Iron Mind API Documentation

## Overview

Comprehensive NFL data API providing real-time and historical statistics, player information, team data, game schedules, injury reports, defensive analytics, standings, and performance metrics. Supports 15 years of historical data (2010-2025) with auto-fetch capabilities for missing data.

**Base URLs:**
- **V1:** `https://nfl.wearemachina.com/api/v1`
- **V2:** `https://nfl.wearemachina.com/api/v2` ⭐ **Recommended**

**Version:** 2.0.0

---

## Table of Contents

1. [Authentication](#authentication)
2. [Rate Limiting](#rate-limiting)
3. [Response Format](#response-format)
4. [Error Handling](#error-handling)
5. [Caching](#caching)
6. [Auto-Fetch System](#auto-fetch-system)
7. [Endpoints](#endpoints)
   - [Players](#players)
   - [Teams](#teams)
   - [Games](#games)
   - [Statistics](#statistics)
   - [Standings](#standings)
   - [Defense](#defense)
   - [Weather](#weather)
   - [Metrics](#metrics)
   - [Admin](#admin-endpoints)
8. [Examples](#examples)

---

## Authentication

### API Key Authentication

Admin endpoints require API key authentication via header:

```http
X-API-Key: your-api-key-here
```

Or using Authorization header:

```http
Authorization: Bearer your-api-key-here
```

### Endpoint Access Levels

| Level | Endpoints | Authentication |
|-------|-----------|----------------|
| Public | Players, Teams, Games, Stats, Standings, Defense, Weather, Metrics | None |
| Admin | /admin/sync/*, /admin/calc/*, /admin/keys/* | **Required** |

---

## Rate Limiting

### Rate Limit Tiers

| Tier | Limit | Endpoints |
|------|-------|-----------|
| Standard | 100 req/min | Public API endpoints |
| Admin | 30 req/min | Admin sync endpoints |
| Weather | 60 req/min | Weather API endpoints |
| Unlimited | ∞ | With UNLIMITED_API_KEY |

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1696262400
Retry-After: 30 (when rate limited)
```

---

## Response Format

### Success Response (Single Resource)

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Patrick Mahomes",
    "position": "QB",
    "team_id": "770e8400-e29b-41d4-a716-446655440001",
    "jersey_number": 15,
    "status": "active"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

### Success Response (Collection with Pagination)

```json
{
  "data": [
    { "id": "...", "name": "..." },
    { "id": "...", "name": "..." }
  ],
  "meta": {
    "total": 100,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

### Special Response Headers

- `X-Cache: HIT|MISS` - Indicates if response served from cache
- `X-Auto-Fetched: true` - Indicates data was automatically fetched from source API

---

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Player not found",
    "status": 404
  }
}
```

### Error Codes

| Status | Code | Description |
|--------|------|-------------|
| 400 | BAD_REQUEST | Invalid parameters or malformed request |
| 401 | UNAUTHORIZED | Invalid or missing API key |
| 404 | NOT_FOUND | Resource not found |
| 405 | METHOD_NOT_ALLOWED | HTTP method not supported |
| 429 | RATE_LIMIT_EXCEEDED | Too many requests |
| 500 | INTERNAL_ERROR | Server error |
| 503 | SERVICE_UNAVAILABLE | Service temporarily unavailable |

---

## Caching

### Cache Strategy

The API uses Redis caching with endpoint-specific TTLs:

| Resource | TTL | Reason |
|----------|-----|--------|
| Teams | 1 hour | Infrequent changes |
| Players | 15 minutes | Moderate update frequency |
| Games | 5 minutes | Live updates during games |
| Stats | 5 minutes | Real-time statistics |
| Stat Leaders | 10 minutes | Aggregate data |
| Standings | 10 minutes | Calculated data |
| Defense Rankings | 10 minutes | Aggregate data |
| Weather | 15 minutes | External API data |

### Cache Headers

Responses include `X-Cache` header:
- `HIT` - Served from cache (fast)
- `MISS` - Fetched from database (slower)

---

## Auto-Fetch System

### Overview

The API automatically fetches missing data when requests return empty results. This self-healing system ensures data availability without manual intervention.

### How It Works

1. Client requests data (e.g., `/api/v2/games?season=2025&week=5`)
2. Database returns empty
3. System detects empty result and eligibility for auto-fetch
4. Fetches data from ESPN/NFLverse APIs
5. Stores in database
6. Returns fetched data with `X-Auto-Fetched: true` header

### Auto-Fetch Eligibility

- **Games:** Current season and previous season (during early season)
- **Players:** When querying by team_id and team exists
- **Stats:** For completed games in current/recent seasons
- **Teams:** Always auto-fetches if fewer than 32 teams exist

### Example

```bash
# First request for week 5 (no data in DB)
curl "https://nfl.wearemachina.com/api/v2/games?season=2025&week=5"
# Response includes: X-Auto-Fetched: true
# Takes 2-5 seconds

# Subsequent requests (data in DB)
curl "https://nfl.wearemachina.com/api/v2/games?season=2025&week=5"
# Response includes: X-Cache: HIT
# Takes ~50ms
```

---

## Endpoints

### Players

#### List Players

```http
GET /api/v2/players
```

Retrieve paginated list of players with optional filters.

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `position` | string | Filter by position (QB, RB, WR, TE, K, etc.) | All |
| `status` | string | Filter by status (active, inactive, injured) | All |
| `team` | UUID | Filter by team ID | All |
| `limit` | integer | Results per page (max 100) | 50 |
| `offset` | integer | Pagination offset | 0 |

**Response:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "nfl_id": 3139477,
      "name": "Patrick Mahomes",
      "position": "QB",
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "jersey_number": 15,
      "height_inches": 75,
      "weight_pounds": 230,
      "birth_date": "1995-09-17T00:00:00Z",
      "college": "Texas Tech",
      "draft_year": 2017,
      "draft_round": 1,
      "draft_pick": 10,
      "status": "active",
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-10-01T14:30:00Z"
    }
  ],
  "meta": {
    "total": 2000,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Auto-Fetch:** Enabled when filtering by `team` parameter and no players found.

---

#### Get Player

```http
GET /api/v2/players/{id}
```

Retrieve detailed information for a single player.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Player ID |

**Response:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "nfl_id": 3139477,
    "name": "Patrick Mahomes",
    "position": "QB",
    "team_id": "770e8400-e29b-41d4-a716-446655440001",
    "jersey_number": 15,
    "height_inches": 75,
    "weight_pounds": 230,
    "birth_date": "1995-09-17T00:00:00Z",
    "college": "Texas Tech",
    "draft_year": 2017,
    "draft_round": 1,
    "draft_pick": 10,
    "status": "active",
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-10-01T14:30:00Z"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Player Career Stats

```http
GET /api/v2/players/{id}/career
```

Retrieve aggregated career statistics for a player.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Player ID |

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Filter by specific season | All seasons |

**Response:**

```json
{
  "data": {
    "player_id": "550e8400-e29b-41d4-a716-446655440000",
    "seasons": [
      {
        "season": 2024,
        "games_played": 17,
        "passing_yards": 4839,
        "passing_touchdowns": 39,
        "interceptions": 11,
        "completion_percentage": 67.2,
        "rushing_yards": 389,
        "rushing_touchdowns": 4
      }
    ],
    "career_totals": {
      "games_played": 98,
      "passing_yards": 28424,
      "passing_touchdowns": 219,
      "interceptions": 63
    }
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Player Team History

```http
GET /api/v2/players/{id}/history
```

Retrieve player's team history showing all teams they've played for.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Player ID |

**Response:**

```json
{
  "data": [
    {
      "player_id": "550e8400-e29b-41d4-a716-446655440000",
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "team_name": "Kansas City Chiefs",
      "start_season": 2017,
      "end_season": 2025,
      "is_current": true
    }
  ],
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Player Injuries

```http
GET /api/v2/players/{id}/injuries
```

Retrieve current and recent injury information for a player.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Player ID |

**Response:**

```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "player_id": "550e8400-e29b-41d4-a716-446655440000",
      "injury_type": "Ankle",
      "status": "Questionable",
      "description": "High ankle sprain",
      "return_timeline": "Week 6",
      "week": 5,
      "season": 2025,
      "reported_at": "2025-09-28T18:00:00Z"
    }
  ],
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Player Advanced Stats

```http
GET /api/v2/players/{id}/advanced-stats
```

Retrieve Next Gen Stats and advanced metrics for a player.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Player ID |

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Filter by season | Current season |

**Response:**

```json
{
  "data": {
    "player_id": "550e8400-e29b-41d4-a716-446655440000",
    "season": 2025,
    "avg_time_to_throw": 2.8,
    "avg_completed_air_yards": 8.2,
    "avg_intended_air_yards": 9.1,
    "avg_air_yards_differential": -0.9,
    "completion_percentage_above_expectation": 3.2,
    "max_completed_air_distance": 58,
    "avg_air_yards_to_sticks": 1.2
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Player vs Defense

```http
GET /api/v2/players/{id}/vs-defense/{teamId}
```

Retrieve player's historical performance against a specific defense.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Player ID |
| `teamId` | UUID | Defensive team ID |

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Filter by season | All seasons |
| `limit` | integer | Max games to return (max 50) | 5 |

**Response:**

```json
{
  "data": {
    "player_id": "550e8400-e29b-41d4-a716-446655440000",
    "defense_team_id": "880e8400-e29b-41d4-a716-446655440002",
    "games": [
      {
        "game_id": "990e8400-e29b-41d4-a716-446655440003",
        "season": 2024,
        "week": 12,
        "date": "2024-11-24T20:20:00Z",
        "passing_yards": 356,
        "passing_touchdowns": 3,
        "interceptions": 0,
        "passer_rating": 132.8
      }
    ],
    "totals": {
      "games_played": 8,
      "avg_passing_yards": 289.3,
      "avg_passing_touchdowns": 2.4,
      "avg_interceptions": 0.6,
      "avg_passer_rating": 108.4
    }
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

### Teams

#### List Teams

```http
GET /api/v2/teams
```

Retrieve list of all NFL teams.

**Response:**

```json
{
  "data": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440001",
      "nfl_id": 12,
      "name": "Kansas City Chiefs",
      "abbreviation": "KC",
      "location": "Kansas City",
      "conference": "AFC",
      "division": "West",
      "color": "#E31837",
      "logo_url": "https://a.espncdn.com/i/teamlogos/nfl/500/kc.png",
      "stadium_name": "GEHA Field at Arrowhead Stadium",
      "stadium_location": "Kansas City, MO",
      "stadium_capacity": 76416,
      "stadium_surface": "Grass",
      "stadium_latitude": 39.0489,
      "stadium_longitude": -94.4839,
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-10-01T14:30:00Z"
    }
  ],
  "meta": {
    "total": 32,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Cache TTL:** 1 hour

**Auto-Fetch:** Enabled if fewer than 32 teams exist in database.

---

#### Get Team

```http
GET /api/v2/teams/{id}
```

Retrieve detailed information for a single team.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Team ID |

**Response:**

```json
{
  "data": {
    "id": "770e8400-e29b-41d4-a716-446655440001",
    "nfl_id": 12,
    "name": "Kansas City Chiefs",
    "abbreviation": "KC",
    "location": "Kansas City",
    "conference": "AFC",
    "division": "West",
    "color": "#E31837",
    "logo_url": "https://a.espncdn.com/i/teamlogos/nfl/500/kc.png",
    "stadium_name": "GEHA Field at Arrowhead Stadium",
    "stadium_location": "Kansas City, MO",
    "stadium_capacity": 76416,
    "stadium_surface": "Grass",
    "stadium_latitude": 39.0489,
    "stadium_longitude": -94.4839,
    "created_at": "2025-01-15T10:00:00Z",
    "updated_at": "2025-10-01T14:30:00Z"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Team Roster

```http
GET /api/v2/teams/{id}/players
```

Retrieve current roster for a team.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Team ID |

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `position` | string | Filter by position | All |
| `status` | string | Filter by status | All |

**Response:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Patrick Mahomes",
      "position": "QB",
      "jersey_number": 15,
      "status": "active"
    }
  ],
  "meta": {
    "total": 53,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Team Injuries

```http
GET /api/v2/teams/{id}/injuries
```

Retrieve current injury report for a team.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Team ID |

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Filter by season | Current season |
| `week` | integer | Filter by week | Current week |

**Response:**

```json
{
  "data": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440000",
      "player_id": "550e8400-e29b-41d4-a716-446655440000",
      "player_name": "Travis Kelce",
      "position": "TE",
      "injury_type": "Knee",
      "status": "Probable",
      "description": "Knee contusion",
      "return_timeline": "Week 6",
      "week": 5,
      "season": 2025
    }
  ],
  "meta": {
    "total": 3,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Team Defense Stats

```http
GET /api/v2/teams/{id}/defense/stats
```

Retrieve defensive statistics for a team.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Team ID |

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `season` | integer | Season year (required) |
| `week` | integer | Week number (optional) |

**Response:**

```json
{
  "data": {
    "team_id": "770e8400-e29b-41d4-a716-446655440001",
    "season": 2025,
    "week": 5,
    "games_played": 5,
    "points_allowed": 89,
    "yards_allowed": 1542,
    "passing_yards_allowed": 987,
    "rushing_yards_allowed": 555,
    "sacks": 18,
    "interceptions": 7,
    "fumbles_recovered": 4,
    "safeties": 0,
    "touchdowns": 2,
    "third_down_conversions_allowed": 28,
    "third_down_attempts_allowed": 72,
    "fourth_down_conversions_allowed": 3,
    "fourth_down_attempts_allowed": 8,
    "penalties": 22,
    "penalty_yards": 187
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

### Games

#### List Games

```http
GET /api/v2/games
```

Retrieve paginated list of games with optional filters.

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Filter by season year | Current season |
| `week` | integer | Filter by week number | All weeks |
| `team` | UUID | Filter by team ID (home or away) | All teams |
| `status` | string | Filter by status (scheduled, in_progress, final) | All |
| `limit` | integer | Results per page (max 100) | 50 |
| `offset` | integer | Pagination offset | 0 |

**Response:**

```json
{
  "data": [
    {
      "id": "990e8400-e29b-41d4-a716-446655440003",
      "nfl_game_id": "401547418",
      "season": 2025,
      "week": 5,
      "game_type": "regular",
      "game_date": "2025-10-06T20:20:00Z",
      "home_team_id": "770e8400-e29b-41d4-a716-446655440001",
      "away_team_id": "880e8400-e29b-41d4-a716-446655440002",
      "home_score": 27,
      "away_score": 20,
      "status": "final",
      "broadcast": "NBC",
      "venue": "GEHA Field at Arrowhead Stadium",
      "attendance": 76218,
      "weather_conditions": "Clear, 68°F",
      "created_at": "2025-08-15T10:00:00Z",
      "updated_at": "2025-10-06T23:45:00Z"
    }
  ],
  "meta": {
    "total": 272,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Auto-Fetch:** Enabled for current season and previous season when season/week specified.

---

#### Get Game

```http
GET /api/v2/games/{id}
```

Retrieve detailed information for a single game.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Game ID |

**Response:**

```json
{
  "data": {
    "id": "990e8400-e29b-41d4-a716-446655440003",
    "nfl_game_id": "401547418",
    "season": 2025,
    "week": 5,
    "game_type": "regular",
    "game_date": "2025-10-06T20:20:00Z",
    "home_team_id": "770e8400-e29b-41d4-a716-446655440001",
    "away_team_id": "880e8400-e29b-41d4-a716-446655440002",
    "home_score": 27,
    "away_score": 20,
    "status": "final",
    "broadcast": "NBC",
    "venue": "GEHA Field at Arrowhead Stadium",
    "attendance": 76218,
    "weather_conditions": "Clear, 68°F",
    "created_at": "2025-08-15T10:00:00Z",
    "updated_at": "2025-10-06T23:45:00Z"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Game Team Stats

```http
GET /api/v2/games/{id}/stats
```

Retrieve comprehensive team statistics for a game.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Game ID |

**Response:**

```json
{
  "data": {
    "game_id": "990e8400-e29b-41d4-a716-446655440003",
    "home_team": {
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "first_downs": 24,
      "first_downs_passing": 16,
      "first_downs_rushing": 7,
      "first_downs_penalty": 1,
      "total_yards": 398,
      "passing_yards": 289,
      "rushing_yards": 109,
      "passing_attempts": 38,
      "passing_completions": 28,
      "rushing_attempts": 24,
      "third_down_conversions": 7,
      "third_down_attempts": 13,
      "fourth_down_conversions": 0,
      "fourth_down_attempts": 1,
      "turnovers": 1,
      "fumbles": 0,
      "interceptions": 1,
      "sacks_allowed": 2,
      "penalties": 5,
      "penalty_yards": 45,
      "possession_time": "32:18",
      "red_zone_attempts": 4,
      "red_zone_conversions": 3
    },
    "away_team": {
      "team_id": "880e8400-e29b-41d4-a716-446655440002",
      "first_downs": 19,
      "total_yards": 342,
      "passing_yards": 256,
      "rushing_yards": 86,
      "turnovers": 2,
      "possession_time": "27:42"
    }
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Game Scoring Plays

```http
GET /api/v2/games/{id}/scoring-plays
```

Retrieve chronological scoring plays for a game.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | UUID | Game ID |

**Response:**

```json
{
  "data": [
    {
      "id": "aa0e8400-e29b-41d4-a716-446655440004",
      "game_id": "990e8400-e29b-41d4-a716-446655440003",
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "quarter": 1,
      "time_remaining": "8:42",
      "play_type": "Passing Touchdown",
      "description": "Patrick Mahomes 23 yd pass to Travis Kelce",
      "scoring_player_id": "550e8400-e29b-41d4-a716-446655440005",
      "scoring_player_name": "Travis Kelce",
      "points": 6,
      "home_score": 6,
      "away_score": 0,
      "yard_line": 23,
      "play_sequence": 1
    },
    {
      "id": "bb0e8400-e29b-41d4-a716-446655440005",
      "game_id": "990e8400-e29b-41d4-a716-446655440003",
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "quarter": 1,
      "time_remaining": "8:42",
      "play_type": "Extra Point",
      "description": "Harrison Butker extra point",
      "scoring_player_id": "660e8400-e29b-41d4-a716-446655440006",
      "scoring_player_name": "Harrison Butker",
      "points": 1,
      "home_score": 7,
      "away_score": 0,
      "play_sequence": 2
    }
  ],
  "meta": {
    "total": 12,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

### Statistics

#### Stat Leaders

```http
GET /api/v2/stats/leaders
```

Retrieve league leaders for various statistical categories.

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `category` | string | Statistical category (required) | - |
| `season` | integer | Season year | Current season |
| `limit` | integer | Number of leaders to return (max 100) | 10 |

**Valid Categories:**
- `passing_yards`
- `passing_touchdowns`
- `rushing_yards`
- `rushing_touchdowns`
- `receiving_yards`
- `receiving_touchdowns`
- `receptions`
- `sacks`
- `interceptions`
- `tackles`

**Response:**

```json
{
  "data": [
    {
      "rank": 1,
      "player_id": "550e8400-e29b-41d4-a716-446655440000",
      "player_name": "Patrick Mahomes",
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "team_abbreviation": "KC",
      "position": "QB",
      "passing_yards": 1823,
      "games_played": 5
    }
  ],
  "meta": {
    "category": "passing_yards",
    "season": 2025,
    "limit": 10,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Cache TTL:** 10 minutes

---

#### Game Player Stats

```http
GET /api/v2/stats/game/{gameId}
```

Retrieve all player statistics for a specific game.

**Path Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `gameId` | UUID | Game ID |

**Response:**

```json
{
  "data": [
    {
      "player_id": "550e8400-e29b-41d4-a716-446655440000",
      "player_name": "Patrick Mahomes",
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "position": "QB",
      "passing_yards": 356,
      "passing_touchdowns": 3,
      "interceptions": 0,
      "completions": 28,
      "attempts": 38,
      "completion_percentage": 73.7,
      "rushing_yards": 23,
      "rushing_attempts": 4,
      "rushing_touchdowns": 0
    }
  ],
  "meta": {
    "game_id": "990e8400-e29b-41d4-a716-446655440003",
    "total": 48,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Auto-Fetch:** Enabled for completed games when no stats found.

---

### Standings

#### Team Standings

```http
GET /api/v2/standings
```

Retrieve current NFL standings with comprehensive team records.

**Query Parameters (Required):**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Season year (required) | - |
| `week` | integer | Week number | Latest week |
| `division` | string | Filter by division (e.g., "AFC West") | All |
| `conference` | string | Filter by conference (AFC/NFC) | All |

**Response:**

```json
{
  "data": [
    {
      "team_id": "770e8400-e29b-41d4-a716-446655440001",
      "team_name": "Kansas City Chiefs",
      "team_abbreviation": "KC",
      "conference": "AFC",
      "division": "West",
      "season": 2025,
      "week": 5,
      "wins": 4,
      "losses": 1,
      "ties": 0,
      "win_percentage": 0.800,
      "points_for": 142,
      "points_against": 98,
      "point_differential": 44,
      "home_wins": 2,
      "home_losses": 1,
      "away_wins": 2,
      "away_losses": 0,
      "division_wins": 2,
      "division_losses": 0,
      "conference_wins": 3,
      "conference_losses": 1,
      "current_streak": "W3",
      "division_rank": 1,
      "conference_rank": 2,
      "playoff_seed": 2,
      "last_updated": "2025-10-06T23:45:00Z"
    }
  ],
  "meta": {
    "season": 2025,
    "week": 5,
    "total": 32,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Cache TTL:** 10 minutes

---

### Defense

#### Defensive Rankings

```http
GET /api/v2/defense/rankings
```

Retrieve defensive rankings across various categories.

**Query Parameters (Required):**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Season year (required) | - |
| `category` | string | Ranking category | overall |

**Valid Categories:**
- `overall` - Overall defensive ranking
- `pass` - Pass defense ranking
- `rush` - Rush defense ranking
- `points_allowed` - Points allowed ranking

**Response:**

```json
{
  "data": [
    {
      "rank": 1,
      "team_id": "aa0e8400-e29b-41d4-a716-446655440007",
      "team_name": "San Francisco 49ers",
      "team_abbreviation": "SF",
      "games_played": 5,
      "points_allowed": 76,
      "yards_allowed": 1342,
      "passing_yards_allowed": 845,
      "rushing_yards_allowed": 497,
      "sacks": 21,
      "interceptions": 9,
      "fumbles_recovered": 5,
      "touchdowns": 3,
      "yards_per_game": 268.4,
      "points_per_game": 15.2
    }
  ],
  "meta": {
    "season": 2025,
    "category": "overall",
    "total": 32,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Cache TTL:** 10 minutes

---

### Weather

#### Current Weather

```http
GET /api/v2/weather/current
```

Retrieve current weather conditions for a location.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `lat` | float | Latitude |
| `lon` | float | Longitude |

**Response:**

```json
{
  "data": {
    "location": "Kansas City, MO",
    "latitude": 39.0489,
    "longitude": -94.4839,
    "temperature": 68,
    "feels_like": 66,
    "humidity": 52,
    "wind_speed": 8,
    "wind_direction": "SW",
    "conditions": "Clear",
    "precipitation": 0,
    "timestamp": "2025-10-02T14:30:00Z"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

**Rate Limit:** 60 requests/minute

**Cache TTL:** 15 minutes

---

#### Historical Weather

```http
GET /api/v2/weather/historical
```

Retrieve historical weather data for a location and date.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `lat` | float | Latitude |
| `lon` | float | Longitude |
| `date` | string | Date (YYYY-MM-DD) |

**Response:**

```json
{
  "data": {
    "location": "Kansas City, MO",
    "latitude": 39.0489,
    "longitude": -94.4839,
    "date": "2025-09-15",
    "temperature_high": 72,
    "temperature_low": 58,
    "avg_temperature": 65,
    "humidity": 48,
    "wind_speed": 12,
    "conditions": "Partly Cloudy",
    "precipitation": 0
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Weather Forecast

```http
GET /api/v2/weather/forecast
```

Retrieve weather forecast for a location.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `lat` | float | Latitude |
| `lon` | float | Longitude |
| `days` | integer | Number of days (1-7, default 3) |

**Response:**

```json
{
  "data": [
    {
      "date": "2025-10-03",
      "temperature_high": 71,
      "temperature_low": 59,
      "humidity": 55,
      "wind_speed": 10,
      "conditions": "Sunny",
      "precipitation_chance": 5
    },
    {
      "date": "2025-10-04",
      "temperature_high": 68,
      "temperature_low": 56,
      "humidity": 62,
      "wind_speed": 14,
      "conditions": "Cloudy",
      "precipitation_chance": 20
    }
  ],
  "meta": {
    "location": "Kansas City, MO",
    "latitude": 39.0489,
    "longitude": -94.4839,
    "days": 3,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

### Metrics

#### Database Metrics

```http
GET /api/v2/metrics/database
```

Retrieve database health and statistics.

**Response:**

```json
{
  "data": {
    "status": "healthy",
    "total_players": 14104,
    "total_teams": 32,
    "total_games": 4382,
    "total_game_stats": 77551,
    "latest_season": 2025,
    "latest_week": 5,
    "database_size": "2.4 GB",
    "last_sync": "2025-10-02T12:00:00Z"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Health Check

```http
GET /api/v2/health
```

Simple health check endpoint.

**Response:**

```json
{
  "data": {
    "status": "healthy",
    "version": "2.0.0",
    "uptime": 86400,
    "database": "connected",
    "cache": "connected"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

### Admin Endpoints

⚠️ **Authentication Required** - All admin endpoints require API key authentication.

#### Sync Teams

```http
POST /api/v1/admin/sync/teams
X-API-Key: your-admin-key
```

Sync all NFL teams from ESPN API.

**Response:**

```json
{
  "data": {
    "message": "Successfully synced 32 teams"
  },
  "meta": {
    "duration": "2.3s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Sync Rosters

```http
POST /api/v1/admin/sync/rosters
X-API-Key: your-admin-key
```

Sync all team rosters from ESPN API.

**Response:**

```json
{
  "data": {
    "message": "Successfully synced rosters for 32 teams",
    "total_players": 1696
  },
  "meta": {
    "duration": "45.2s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Sync Games

```http
POST /api/v1/admin/sync/games?season=2025&week=5
X-API-Key: your-admin-key
```

Sync games for specific season and week.

**Query Parameters:**

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `season` | integer | Season year | Current season |
| `week` | integer | Week number | Current week |

**Response:**

```json
{
  "data": {
    "message": "Successfully synced 16 games for 2025 week 5"
  },
  "meta": {
    "duration": "8.7s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Full Sync

```http
POST /api/v1/admin/sync/full
X-API-Key: your-admin-key
```

Perform full sync of teams, rosters, and current season games.

**Response:**

```json
{
  "data": {
    "message": "Full sync completed",
    "teams_synced": 32,
    "players_synced": 1696,
    "games_synced": 80
  },
  "meta": {
    "duration": "120.5s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Sync Historical Season

```http
POST /api/v1/admin/sync/historical/season?year=2024
X-API-Key: your-admin-key
```

Sync complete historical season from NFLverse.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `year` | integer | Season year (2010-2024) |

**Response:**

```json
{
  "data": {
    "message": "Successfully synced 2024 season",
    "games_synced": 272,
    "stats_synced": 5234
  },
  "meta": {
    "duration": "180.2s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Sync Multiple Historical Seasons

```http
POST /api/v1/admin/sync/historical/seasons?start=2010&end=2024
X-API-Key: your-admin-key
```

Sync multiple historical seasons from NFLverse.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `start` | integer | Start year (2010-2024) |
| `end` | integer | End year (2010-2024) |

**Response:**

```json
{
  "data": {
    "message": "Successfully synced seasons 2010-2024",
    "seasons_synced": 15,
    "total_games": 4110,
    "total_stats": 77551
  },
  "meta": {
    "duration": "2400.5s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Sync NFLverse Stats

```http
POST /api/v1/admin/sync/nflverse/stats?season=2024
X-API-Key: your-admin-key
```

Sync player statistics from NFLverse.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `season` | integer | Season year |

---

#### Sync NFLverse Schedule

```http
POST /api/v1/admin/sync/nflverse/schedule?season=2024
X-API-Key: your-admin-key
```

Sync game schedule from NFLverse.

---

#### Sync Next Gen Stats

```http
POST /api/v1/admin/sync/nextgen-stats?season=2024
X-API-Key: your-admin-key
```

Sync Next Gen Stats from NFLverse.

---

#### Sync Weather

```http
POST /api/v1/admin/sync/weather?season=2024&week=5
X-API-Key: your-admin-key
```

Enrich games with weather data.

---

#### Sync Team Stats

```http
POST /api/v1/admin/sync/team-stats?season=2024
X-API-Key: your-admin-key
```

Sync team statistics.

---

#### Sync Injuries

```http
POST /api/v1/admin/sync/injuries
X-API-Key: your-admin-key
```

Sync current injury reports from ESPN.

**Response:**

```json
{
  "data": {
    "message": "Successfully synced injuries for 32 teams",
    "total_injuries": 156
  },
  "meta": {
    "duration": "12.3s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Sync Scoring Plays

```http
POST /api/v1/admin/sync/scoring-plays?game_id={gameId}
X-API-Key: your-admin-key
```

Sync scoring plays for a specific game.

---

#### Sync Player Season Stats

```http
POST /api/v1/admin/sync/player-season-stats?season=2024
X-API-Key: your-admin-key
```

Sync aggregated season statistics.

---

#### Calculate Standings

```http
POST /api/v1/admin/calc/standings?season=2025&week=5
X-API-Key: your-admin-key
```

Recalculate standings for season/week.

**Query Parameters (Required):**

| Parameter | Type | Description |
|-----------|------|-------------|
| `season` | integer | Season year |
| `week` | integer | Week number |

**Response:**

```json
{
  "data": {
    "message": "Successfully calculated standings for 2025 week 5",
    "teams_updated": 32
  },
  "meta": {
    "duration": "3.2s",
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

#### Generate API Key

```http
POST /api/v1/admin/keys/generate
X-API-Key: your-admin-key
```

Generate new API key (requires existing admin key).

**Response:**

```json
{
  "data": {
    "api_key": "gim_live_abc123xyz789",
    "created_at": "2025-10-02T14:30:00Z"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

---

## Examples

### Get Active Quarterbacks

```bash
curl "https://nfl.wearemachina.com/api/v2/players?position=QB&status=active&limit=10"
```

### Get Team Roster

```bash
curl "https://nfl.wearemachina.com/api/v2/teams/770e8400-e29b-41d4-a716-446655440001/players"
```

### Get Current Week Games

```bash
curl "https://nfl.wearemachina.com/api/v2/games?season=2025&week=5"
```

### Get Passing Yards Leaders

```bash
curl "https://nfl.wearemachina.com/api/v2/stats/leaders?category=passing_yards&season=2025&limit=10"
```

### Get Current Standings

```bash
curl "https://nfl.wearemachina.com/api/v2/standings?season=2025&week=5"
```

### Get Defensive Rankings

```bash
curl "https://nfl.wearemachina.com/api/v2/defense/rankings?season=2025&category=overall"
```

### Get Player Career Stats

```bash
curl "https://nfl.wearemachina.com/api/v2/players/550e8400-e29b-41d4-a716-446655440000/career?season=2024"
```

### Get Player vs Defense Stats

```bash
curl "https://nfl.wearemachina.com/api/v2/players/550e8400-e29b-41d4-a716-446655440000/vs-defense/880e8400-e29b-41d4-a716-446655440002"
```

### Get Game Scoring Plays

```bash
curl "https://nfl.wearemachina.com/api/v2/games/990e8400-e29b-41d4-a716-446655440003/scoring-plays"
```

### Admin: Sync Current Week Games

```bash
curl -X POST \
  -H "X-API-Key: your-admin-key" \
  "https://nfl.wearemachina.com/api/v1/admin/sync/games?season=2025&week=5"
```

### JavaScript Fetch Example

```javascript
// Get active quarterbacks
const response = await fetch(
  'https://nfl.wearemachina.com/api/v2/players?position=QB&status=active',
  {
    headers: {
      'Accept': 'application/json'
    }
  }
);

const data = await response.json();
console.log(`Total QBs: ${data.meta.total}`);
console.log(`Players:`, data.data);

// Check cache status
console.log(`Cache: ${response.headers.get('X-Cache')}`);
console.log(`Rate Limit Remaining: ${response.headers.get('X-RateLimit-Remaining')}`);
```

### Python Example

```python
import requests

# Get current standings
response = requests.get(
    'https://nfl.wearemachina.com/api/v2/standings',
    params={'season': 2025, 'week': 5}
)

data = response.json()
for team in data['data']:
    print(f"{team['team_abbreviation']}: {team['wins']}-{team['losses']} ({team['win_percentage']:.3f})")
```

### TypeScript Example

```typescript
interface Player {
  id: string;
  name: string;
  position: string;
  team_id: string;
  jersey_number: number;
  status: string;
}

interface ApiResponse<T> {
  data: T;
  meta: {
    total?: number;
    limit?: number;
    offset?: number;
    timestamp: string;
  };
}

async function getPlayers(position: string): Promise<Player[]> {
  const response = await fetch(
    `https://nfl.wearemachina.com/api/v2/players?position=${position}&status=active`
  );

  const result: ApiResponse<Player[]> = await response.json();
  return result.data;
}

// Usage
const quarterbacks = await getPlayers('QB');
console.log(`Found ${quarterbacks.length} active quarterbacks`);
```

---

## Data Coverage

### Historical Data (2010-2024)
- ✅ **14,104** unique players
- ✅ **4,110** games across 15 seasons
- ✅ **77,551** game stat records
- ✅ Complete play-by-play data
- ✅ Next Gen Stats (2016+)
- ✅ Advanced metrics

### Current Season (2025)
- ✅ All 32 teams with current rosters
- ✅ Complete 18-week schedule (272 games)
- ✅ Live game updates (5-minute refresh during games)
- ✅ Real-time injury reports
- ✅ Weekly standings calculations
- ✅ Defensive rankings

### Auto-Fetch Capabilities
- ✅ Automatically fetches missing games when queried
- ✅ Fetches team rosters on-demand
- ✅ Retrieves stats for completed games
- ✅ Self-healing data layer

---

## Performance

### Response Times
- **Cache Hit:** ~50ms average
- **Cache Miss (Simple Query):** ~100-200ms
- **Cache Miss (Complex Query):** ~200-500ms
- **Auto-Fetch (First Request):** 2-10 seconds
- **Subsequent Requests:** Fast (cached)

### Optimization Tips
1. **Use V2 endpoints** - Latest optimizations
2. **Enable caching** - Check `X-Cache` header
3. **Paginate large results** - Use `limit` and `offset`
4. **Filter aggressively** - Reduce result set size
5. **Batch requests** - Combine filters when possible
6. **Monitor rate limits** - Check `X-RateLimit-*` headers

---

## API Changelog

### Version 2.0.0 (2025-10-02)
- ✅ Added V2 endpoints with improved performance
- ✅ Implemented auto-fetch system for missing data
- ✅ Added comprehensive standings endpoint
- ✅ Added defensive rankings and matchup analysis
- ✅ Added player vs defense statistics
- ✅ Added game scoring plays timeline
- ✅ Added team defensive stats
- ✅ Added advanced stats (Next Gen Stats)
- ✅ Improved caching strategy
- ✅ Enhanced error handling and validation
- ✅ Added database and health metrics
- ✅ Loaded 15 years of historical data (2010-2024)

### Version 1.0.0 (2025-01-15)
- Initial API release
- Basic player, team, game, and stats endpoints
- Admin sync endpoints

---

## Support

For issues, questions, or feature requests:
- **API Status:** https://nfl.wearemachina.com/health
- **Documentation:** This file
- **Project README:** ../README.md

---

**Last Updated:** October 2, 2025
**API Version:** 2.0.0
**Data Coverage:** 2010-2025 (15 seasons)
