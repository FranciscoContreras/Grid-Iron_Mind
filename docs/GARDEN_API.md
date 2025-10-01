# 🌱 AI Data Garden API Endpoints

## Overview

The AI Data Garden provides intelligent data management, natural language queries, and automated health monitoring for the NFL API.

---

## Endpoints

### 1. **Garden Status**
Get overall status of the AI Data Garden system.

```bash
GET /api/v1/garden/status
```

**Response:**
```json
{
  "data": {
    "ai_enabled": true,
    "ai_provider": "grok",
    "data_counts": {
      "players": 2543,
      "games": 285,
      "injuries": 47
    },
    "last_updates": {
      "players": "2025-10-01T10:30:00Z",
      "games": "2025-10-01T11:15:00Z"
    },
    "garden_features": {
      "health_monitoring": true,
      "data_enrichment": true,
      "natural_queries": true,
      "smart_scheduling": true
    },
    "timestamp": "2025-10-01T12:00:00Z"
  }
}
```

---

### 2. **Health Check (Read-Only)**
Run AI-powered health check without auto-healing.

```bash
GET /api/v1/garden/health
```

**Response:**
```json
{
  "data": {
    "health_report": {
      "timestamp": "2025-10-01T12:00:00Z",
      "overall_health": "good",
      "issues": [
        {
          "type": "staleness",
          "severity": "medium",
          "entity": "games",
          "description": "3 completed games from last week have no stats",
          "auto_fixed": false,
          "fix_action": "Sync game stats from NFLverse"
        }
      ],
      "recommendations": [
        "Sync stats for completed games",
        "Update injury reports (24h old)"
      ],
      "ai_provider": "grok",
      "auto_fixed_issues": 0,
      "metrics_analysis": {
        "player_data": "2,543 players with 95% having complete profiles",
        "game_data": "285 games in system, 3 missing stats",
        "stats_coverage": "Good coverage for recent weeks",
        "data_freshness": "Most data updated within 24 hours"
      }
    },
    "message": "Health: good with 1 issues detected"
  }
}
```

---

### 3. **Health Check with Auto-Heal**
Run health check and automatically fix issues.

```bash
POST /api/v1/garden/health
```

**Response:**
```json
{
  "data": {
    "health_report": {
      "timestamp": "2025-10-01T12:00:00Z",
      "overall_health": "good",
      "issues": [...],
      "auto_fixed_issues": 2
    },
    "message": "Health: good - Fixed 2/3 issues"
  }
}
```

---

### 4. **Natural Language Query** 🔥
Query the database using plain English.

```bash
POST /api/v1/garden/query
Content-Type: application/json
X-API-Key: your-api-key
```

**Request:**
```json
{
  "query": "Who are the top 5 rushing leaders this season?"
}
```

**Response:**
```json
{
  "data": {
    "query": "Who are the top 5 rushing leaders this season?",
    "sql": "SELECT p.name, SUM(gs.rushing_yards) as total_yards FROM players p JOIN game_stats gs ON gs.player_id = p.id JOIN games g ON gs.game_id = g.id WHERE g.season_year = 2025 AND g.status = 'final' GROUP BY p.id, p.name ORDER BY total_yards DESC LIMIT 5",
    "explanation": "This query finds the top 5 players by total rushing yards in the 2025 season",
    "results": [
      {
        "name": "Saquon Barkley",
        "total_yards": 1234
      },
      {
        "name": "Christian McCaffrey",
        "total_yards": 1189
      },
      {
        "name": "Derrick Henry",
        "total_yards": 1156
      }
    ],
    "count": 5,
    "insights": "Saquon Barkley leads the league with 1,234 rushing yards through 10 games. CMC and Derrick Henry are close behind. All three are elite RB1 options for fantasy playoffs.",
    "ai_provider": "grok",
    "warnings": []
  }
}
```

**Example Queries:**
- "Show me all QBs with over 300 passing yards last week"
- "Which teams have the most injuries right now?"
- "Top 10 receivers by fantasy points this season"
- "Games between Chiefs and Bills in 2024"
- "Players drafted in 2023 who are starters"

---

### 5. **Enrich Player Data**
Use AI to enhance player data with missing information.

```bash
POST /api/v1/garden/enrich/player/{playerID}
X-API-Key: your-api-key
```

**Response:**
```json
{
  "data": {
    "player": {
      "id": "uuid",
      "name": "Brock Purdy",
      "position": "QB"
    },
    "enrichments": [
      {
        "entity_type": "player",
        "field": "college",
        "current_value": null,
        "suggested_value": "Iowa State",
        "confidence": 0.95,
        "reasoning": "Brock Purdy played college football at Iowa State",
        "sources": ["NFL Draft records", "ESPN"]
      },
      {
        "field": "draft_year",
        "suggested_value": 2022,
        "confidence": 0.99
      }
    ],
    "tags": [
      "game-manager",
      "efficient-passer",
      "qb2",
      "49ers-system",
      "rookie-contract"
    ],
    "summary": "Brock Purdy is the San Francisco 49ers' starting quarterback known for his efficiency and decision-making. He was the last pick in the 2022 NFL Draft (Mr. Irrelevant) and has exceeded expectations.",
    "similar": [
      "Jared Goff",
      "Tua Tagovailoa",
      "Kirk Cousins"
    ],
    "message": "Generated 2 enrichments and 5 tags"
  }
}
```

---

### 6. **Smart Sync Schedule**
Get AI-generated sync schedule based on current context.

```bash
GET /api/v1/garden/schedule
```

**Response:**
```json
{
  "data": {
    "schedule": {
      "timestamp": "2025-10-01T12:00:00Z",
      "game_day_mode": true,
      "reasoning": "It's Sunday at 1:00 PM - active game day. High-frequency syncs needed for live score updates.",
      "recommendations": [
        {
          "sync_type": "games",
          "priority": "critical",
          "reason": "Games in progress - scores updating",
          "estimated_time": "5-10 minutes",
          "next_sync_in": "15m"
        },
        {
          "sync_type": "injuries",
          "priority": "high",
          "reason": "Game day - injury updates common",
          "estimated_time": "2-5 minutes",
          "next_sync_in": "1h"
        },
        {
          "sync_type": "stats",
          "priority": "high",
          "reason": "Games completing soon - stats will be available",
          "estimated_time": "10-20 minutes",
          "next_sync_in": "30m"
        },
        {
          "sync_type": "rosters",
          "priority": "low",
          "reason": "Roster changes unlikely during game day",
          "estimated_time": "10-15 minutes",
          "next_sync_in": "24h"
        }
      ],
      "ai_provider": "grok"
    },
    "message": "Generated schedule with 4 sync recommendations"
  }
}
```

**Schedule adapts to**:
- Day of week (game days vs off days)
- Time of day (business hours, night, morning)
- NFL season phase (preseason, regular, playoffs, offseason)
- Current game status (pre-game, in-progress, post-game)

---

## Usage Examples

### Health Monitoring

**Morning Health Check**:
```bash
curl https://nfl.wearemachina.com/api/v1/garden/health
```

**Auto-Heal Issues**:
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/garden/health
```

### Natural Language Queries

**Fantasy Football Query**:
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/garden/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key" \
  -d '{"query": "Best wide receivers for fantasy this week"}'
```

**Team Analysis**:
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/garden/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key" \
  -d '{"query": "Which teams have winning records at home?"}'
```

**Injury Report**:
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/garden/query \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key" \
  -d '{"query": "Show all injured players for the Eagles"}'
```

### Data Enrichment

**Enrich Player**:
```bash
curl -X POST https://nfl.wearemachina.com/api/v1/garden/enrich/player/{uuid} \
  -H "X-API-Key: your-key"
```

### Schedule Planning

**Get Sync Plan**:
```bash
curl https://nfl.wearemachina.com/api/v1/garden/schedule
```

---

## Authentication

- **Garden Status**: Public (no auth required)
- **Health Check (GET)**: Public (no auth required)
- **Health Check (POST)**: Public (no auth required)
- **Natural Query**: Requires API key
- **Enrich Player**: Requires API key
- **Schedule**: Public (no auth required)

API key can be provided via:
- Header: `X-API-Key: your-key`
- Header: `Authorization: Bearer your-key`

---

## Rate Limits

- **Public endpoints**: 100 requests/minute
- **AI endpoints** (query, enrich): 10 requests/minute
- **Unlimited key**: No rate limits

---

## Error Responses

**AI Not Available**:
```json
{
  "error": {
    "code": "AI_UNAVAILABLE",
    "message": "AI service not configured",
    "status": 503
  }
}
```

**Unsafe Query**:
```json
{
  "error": {
    "code": "UNSAFE_QUERY",
    "message": "Query contains unsafe operations",
    "status": 400
  }
}
```

**Invalid Request**:
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Query is required",
    "status": 400
  }
}
```

---

## Tips

### Writing Good Natural Language Queries

**Good**:
- "Top 5 rushing leaders this season"
- "QBs with over 300 passing yards last week"
- "Players on the Chiefs roster"
- "Games in week 5 of 2024 season"

**Too Vague**:
- "Good players" (define "good")
- "Recent games" (specify timeframe)
- "Stats" (which stats, for whom?)

**Pro Tip**: Be specific about:
- Time periods (this season, last week, 2024)
- Metrics (passing yards, TDs, fantasy points)
- Filters (position, team, status)

---

## Integration Example

```javascript
// JavaScript client
async function askGarden(question) {
  const response = await fetch('https://nfl.wearemachina.com/api/v1/garden/query', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': 'your-key'
    },
    body: JSON.stringify({ query: question })
  });

  const data = await response.json();

  return {
    results: data.data.results,
    insights: data.data.insights,
    sql: data.data.sql  // See what SQL was generated
  };
}

// Use it
const result = await askGarden("Who leads in touchdowns?");
console.log(result.insights);
// "Patrick Mahomes leads with 32 TDs..."
```

---

## Future Enhancements

Coming soon:
- **Conversational Mode**: Follow-up questions with context
- **Saved Queries**: Save common natural language queries
- **Auto-Enrichment**: Scheduled enrichment runs
- **Trend Detection**: AI alerts for unusual patterns
- **Custom Health Rules**: Define your own health checks
