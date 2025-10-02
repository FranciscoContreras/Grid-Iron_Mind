# Grid Iron Mind API Documentation

## Overview

Comprehensive NFL data API providing real-time and historical statistics, player information, team data, game schedules, injury reports, and performance metrics.

**Base URL:** `https://nfl.wearemachina.com/api/v1`

**Version:** 2.0.0

---

## Table of Contents

1. [Authentication](#authentication)
2. [Rate Limiting](#rate-limiting)
3. [Response Format](#response-format)
4. [Error Handling](#error-handling)
5. [Endpoints](#endpoints)
6. [Examples](#examples)

---

## Authentication

### API Key Authentication

Some endpoints require API key authentication via header:

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
| Public | Players, Teams, Games, Stats | Optional |
| Admin | /admin/sync/* | **Required** |

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
    "position": "QB"
  },
  "meta": {
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

### Success Response (Collection)

```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "limit": 50,
    "offset": 0,
    "timestamp": "2025-10-02T14:30:00Z"
  }
}
```

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
| 400 | BAD_REQUEST | Invalid parameters |
| 401 | UNAUTHORIZED | Invalid API key |
| 404 | NOT_FOUND | Resource not found |
| 429 | RATE_LIMIT_EXCEEDED | Too many requests |
| 500 | INTERNAL_ERROR | Server error |

---

## Endpoints

### Players

#### List Players
```http
GET /api/v1/players?position=QB&status=active&limit=50
```

#### Get Player
```http
GET /api/v1/players/{id}
```

#### Player Career Stats
```http
GET /api/v1/players/{id}/career?season=2025
```

### Teams

#### List Teams
```http
GET /api/v1/teams
```

#### Get Team
```http
GET /api/v1/teams/{id}
```

#### Team Roster
```http
GET /api/v1/teams/{id}/players?position=QB
```

### Games

#### List Games
```http
GET /api/v1/games?season=2025&week=5
```

#### Get Game
```http
GET /api/v1/games/{id}
```

### Statistics

#### Stat Leaders
```http
GET /api/v1/stats/leaders?type=passing&season=2025&limit=10
```

#### Game Stats
```http
GET /api/v1/stats/game/{gameID}
```

### Admin

#### Sync Teams
```http
POST /api/v1/admin/sync/teams
X-API-Key: admin-key
```

#### Sync Games
```http
POST /api/v1/admin/sync/games?season=2025&week=5
X-API-Key: admin-key
```

---

## Examples

### Get Active QBs

```bash
curl "https://nfl.wearemachina.com/api/v1/players?position=QB&status=active"
```

### Get Game Stats

```bash
curl "https://nfl.wearemachina.com/api/v1/stats/game/770e8400-e29b-41d4-a716-446655440001"
```

### JavaScript Example

```javascript
const response = await fetch(
  'https://nfl.wearemachina.com/api/v1/players?position=QB',
  { headers: { 'X-API-Key': 'your-key' } }
);
const data = await response.json();
```

---

For complete documentation, see the project README.
