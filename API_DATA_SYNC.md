# Grid Iron Mind - API Data Synchronization Guide

This document explains how to populate the database with NFL data using the robust API endpoints.

## Base URL

Production: `https://grid-iron-mind-71cc9734eaf4.herokuapp.com`

## Admin Endpoints Overview

All admin endpoints require POST requests and run asynchronously in the background.

### 1. Sync Teams
Fetches all 32 NFL teams from ESPN API.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/teams
```

**Response:**
```json
{
  "data": {
    "message": "Teams sync completed successfully",
    "status": "success"
  }
}
```

### 2. Sync Rosters
Fetches all players for all teams from ESPN API.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/rosters
```

**Response:**
```json
{
  "data": {
    "message": "Rosters sync started in background",
    "status": "processing"
  }
}
```

**Note:** This operation takes several minutes. Check logs for completion:
```bash
heroku logs --tail -a grid-iron-mind
```

### 3. Sync Current Week Games
Fetches current week's games from ESPN scoreboard.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/games
```

### 4. Sync Historical Games (Single Season)
Fetches all games for a specific season.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/historical/season \
  -H "Content-Type: application/json" \
  -d '{"season": 2024}'
```

### 5. Sync Multiple Seasons
Fetches games for multiple seasons at once.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/historical/seasons \
  -H "Content-Type: application/json" \
  -d '{"start_season": 2020, "end_season": 2024}'
```

### 6. Sync NFLverse Player Stats
Fetches detailed player statistics from NFLverse (nflverse.com) data source.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/nflverse/stats \
  -H "Content-Type: application/json" \
  -d '{"season": 2024}'
```

**What it populates:**
- `game_stats` table - Individual player game statistics
- `player_career_stats` table - Aggregated career statistics by season

### 7. Sync NFLverse Schedule
Enhanced schedule data from NFLverse with additional metadata.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/nflverse/schedule \
  -H "Content-Type: application/json" \
  -d '{"season": 2024}'
```

### 8. Sync NFLverse Next Gen Stats
Advanced analytics and next-gen statistics.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/nflverse/nextgen \
  -H "Content-Type: application/json" \
  -d '{"season": 2024, "stat_type": "passing"}'
```

**Stat types:** `passing`, `rushing`, `receiving`

### 9. Sync Team Stats
Fetches team-level statistics for completed games.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/team-stats
```

### 10. Enrich Weather Data
Adds weather information to games based on location and date.

```bash
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/weather
```

**Note:** Requires `WEATHER_API_KEY` environment variable set.

## Complete Data Population Sequence

To fully populate the database, run these commands in order:

```bash
# 1. Sync teams (required first)
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/teams

# 2. Sync all player rosters (takes ~5 minutes)
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/rosters

# 3. Sync historical games for recent seasons
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/historical/seasons \
  -H "Content-Type: application/json" \
  -d '{"start_season": 2023, "end_season": 2024}'

# 4. Sync player statistics from NFLverse (CRITICAL for career stats)
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/nflverse/stats \
  -H "Content-Type: application/json" \
  -d '{"season": 2024}'

curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/nflverse/stats \
  -H "Content-Type: application/json" \
  -d '{"season": 2023}'

# 5. Sync team stats
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/team-stats

# 6. Optional: Enrich with weather data
curl -X POST https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/admin/sync/weather
```

## Query Endpoints

### Players

**List all players:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players?limit=50&offset=0
```

**Filter by position:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players?position=QB
```

**Filter by team:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players?team=TEAM_UUID"
```

**Get single player:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players/PLAYER_UUID
```

**Get player career stats:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players/PLAYER_UUID/career
```

**Response example:**
```json
{
  "data": {
    "player_id": "uuid",
    "total_seasons": 5,
    "career_stats": [
      {
        "season": 2024,
        "team_id": "uuid",
        "games_played": 17,
        "passing_yards": 4500,
        "passing_tds": 35,
        "rushing_yards": 350,
        "receiving_yards": 0
      }
    ],
    "team_history": [
      {
        "team_id": "uuid",
        "team_name": "Kansas City Chiefs",
        "start_season": 2020,
        "end_season": null,
        "is_current": true
      }
    ]
  }
}
```

**Get player team history:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/players/PLAYER_UUID/history
```

### Teams

**List all teams:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/teams
```

**Get single team:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/teams/TEAM_UUID
```

**Get team roster:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/teams/TEAM_UUID/players
```

### Games

**List games:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games?limit=20&offset=0
```

**Filter by team:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games?team=TEAM_UUID"
```

**Filter by season and week:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games?season=2024&week=10"
```

**Get single game:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games/GAME_UUID
```

**Get player stats for a game:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/stats/game/GAME_UUID
```

**Response example:**
```json
{
  "data": [
    {
      "id": "uuid",
      "player_id": "uuid",
      "game_id": "uuid",
      "team_id": "uuid",
      "season_year": 2024,
      "week": 10,
      "passing_yards": 350,
      "passing_touchdowns": 3,
      "passing_interceptions": 1,
      "passing_completions": 25,
      "passing_attempts": 35,
      "rushing_yards": 45,
      "rushing_touchdowns": 0,
      "rushing_attempts": 8,
      "receiving_yards": 0,
      "receiving_touchdowns": 0,
      "receiving_receptions": 0,
      "receiving_targets": 0,
      "player": {
        "id": "uuid",
        "name": "Patrick Mahomes",
        "position": "QB"
      }
    }
  ]
}
```

**Get team stats for a game:**
```bash
curl https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/games/GAME_UUID/stats
```

### Stats Leaders

**Get passing yards leaders:**
```bash
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/stats/leaders?category=passing_yards&season=2024&limit=10"
```

**Categories:**
- `passing_yards`
- `passing_touchdowns`
- `rushing_yards`
- `rushing_touchdowns`
- `receiving_yards`
- `receiving_touchdowns`

## Database Tables

### Core Tables
- `teams` - NFL teams (32 rows)
- `players` - NFL players (~2,284 rows after roster sync)
- `games` - NFL games (288 per season)
- `game_stats` - Individual player game statistics
- `player_career_stats` - Aggregated season statistics per player
- `player_team_history` - Player team changes over time
- `game_team_stats` - Team-level game statistics

### Supporting Tables
- `predictions` - AI predictions (future feature)
- `ai_analysis` - AI-generated insights (future feature)

## Monitoring Data Population

Check database record counts:

```bash
# Check teams
heroku pg:psql -a grid-iron-mind -c "SELECT COUNT(*) FROM teams;"

# Check players
heroku pg:psql -a grid-iron-mind -c "SELECT COUNT(*) FROM players;"

# Check games
heroku pg:psql -a grid-iron-mind -c "SELECT COUNT(*) FROM games;"

# Check game stats
heroku pg:psql -a grid-iron-mind -c "SELECT COUNT(*) FROM game_stats;"

# Check career stats
heroku pg:psql -a grid-iron-mind -c "SELECT COUNT(*) FROM player_career_stats;"
```

## Notes

1. **Order matters**: Always sync teams first, then rosters, then games, then stats.

2. **NFLverse stats are crucial**: The `/admin/sync/nflverse/stats` endpoint populates both `game_stats` and `player_career_stats` tables, which power the career statistics feature.

3. **All sync operations are asynchronous**: Check Heroku logs to monitor progress.

4. **Rate limiting**: Some endpoints may be rate-limited by source APIs. If syncs fail, wait a few minutes and retry.

5. **Weather API**: Weather enrichment requires a valid API key from weatherapi.com set as `WEATHER_API_KEY` environment variable.

6. **Cache invalidation**: Successful syncs automatically invalidate relevant caches.
