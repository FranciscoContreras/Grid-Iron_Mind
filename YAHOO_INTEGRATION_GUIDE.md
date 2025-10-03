# Yahoo Fantasy Sports API Integration Guide

## Overview

This guide explains how to complete the Yahoo Fantasy Sports API integration for Grid Iron Mind. The foundation has been built, but due to OAuth2 complexity and credential security requirements, final implementation requires careful setup.

---

## 🎯 What's Been Built

### ✅ Complete Foundation

1. **Yahoo API Client** (`internal/yahoo/client.go`)
   - OAuth2 authentication support
   - All major API endpoints (rankings, projections, matchups)
   - Retry logic and rate limiting
   - XML/JSON response parsing

2. **Database Schema** (`migrations/011_add_yahoo_fantasy_data.sql`)
   - Player rankings (weekly, by position)
   - Player projections (stats breakdown)
   - Ownership data (percent owned/started)
   - Fantasy leagues tracking
   - Matchup advice and ratings
   - Player news and notes
   - Transaction trends
   - OAuth token storage (encrypted)

3. **Ingestion Service** (`internal/yahoo/ingestion.go`)
   - Sync player rankings
   - Sync projections
   - Map Yahoo players to database
   - Batch operations support

4. **API Handlers** (`internal/handlers/yahoo_fantasy.go`)
   - GET `/api/v1/fantasy/rankings` - Player rankings
   - GET `/api/v1/fantasy/projections/:id` - Player projections
   - GET `/api/v1/fantasy/projections/top` - Top projected players
   - GET `/api/v1/fantasy/ownership` - Ownership data

### 📊 What Yahoo Data Adds

**Unique Data Not in ESPN:**
- ✅ Fantasy-specific player rankings
- ✅ Weekly fantasy point projections
- ✅ Ownership percentages (how many teams roster a player)
- ✅ Start/sit recommendations
- ✅ Matchup difficulty ratings
- ✅ Waiver wire trends (adds/drops)
- ✅ FAAB bid data
- ✅ Fantasy news and injury impact analysis

---

## 🔑 OAuth2 Setup Required

### Step 1: Create Yahoo Developer App

1. Go to https://developer.yahoo.com/apps/
2. Create a new app
3. Fill in details:
   - **App Name:** Grid Iron Mind
   - **App Type:** Web Application
   - **Callback Domain:** `https://nfl.wearemachina.com` (or your domain)
   - **API Permissions:** Fantasy Sports (Read)
4. Get your credentials:
   - **Client ID:** `your_yahoo_client_id_here`
   - **Client Secret:** `your_yahoo_client_secret_here`

### Step 2: OAuth2 Flow Options

Yahoo requires OAuth2 for API access. You have 3 options:

#### Option A: Server-to-Server (Recommended for your use case)

**Use Case:** Automated background sync without user interaction

**Implementation:**
```go
// 1. Set up OAuth config in environment
YAHOO_CLIENT_ID=your_client_id_here
YAHOO_CLIENT_SECRET=your_client_secret_here
YAHOO_REDIRECT_URL=https://nfl.wearemachina.com/auth/yahoo/callback

// 2. One-time manual auth to get refresh token
// Run this locally once to get a refresh token:
func setupYahooAuth() {
    client := yahoo.NewClient(yahoo.Config{
        ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
        ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
        RedirectURL:  os.Getenv("YAHOO_REDIRECT_URL"),
    })

    // Get auth URL
    authURL := client.GetAuthURL("random-state-string")
    fmt.Println("Visit:", authURL)

    // User visits URL, grants permission, gets redirected
    // Extract code from redirect URL: ?code=xxxxx

    var code string
    fmt.Print("Enter auth code: ")
    fmt.Scanln(&code)

    // Exchange for tokens
    token, err := client.ExchangeCode(context.Background(), code)
    if err != nil {
        log.Fatal(err)
    }

    // Store refresh token securely (environment variable or encrypted DB)
    fmt.Println("Refresh Token:", token.RefreshToken)
    fmt.Println("Access Token:", token.AccessToken)
    fmt.Println("Expires:", token.Expiry)
}

// 3. In production, use refresh token to get new access tokens
func getYahooClient() *yahoo.Client {
    client := yahoo.NewClient(config)

    // Load refresh token from secure storage
    refreshToken := os.Getenv("YAHOO_REFRESH_TOKEN")

    // Create token with refresh token
    token := &oauth2.Token{
        RefreshToken: refreshToken,
    }

    // Client will automatically refresh when needed
    client.SetToken(token)

    return client
}
```

#### Option B: User OAuth (If you want league-specific data)

**Use Case:** Users connect their Yahoo accounts to see their league data

**Implementation:**
```go
// Add endpoints to cmd/server/main.go

// Initiate OAuth
mux.HandleFunc("/auth/yahoo/login", func(w http.ResponseWriter, r *http.Request) {
    yahooClient := getYahooClient()
    authURL := yahooClient.GetAuthURL("state-" + uuid.New().String())
    http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
})

// OAuth callback
mux.HandleFunc("/auth/yahoo/callback", func(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    if code == "" {
        http.Error(w, "No code", http.StatusBadRequest)
        return
    }

    yahooClient := getYahooClient()
    token, err := yahooClient.ExchangeCode(r.Context(), code)
    if err != nil {
        http.Error(w, "Auth failed", http.StatusInternalServerError)
        return
    }

    // Store token in database for this user
    // Use yahoo_oauth_tokens table

    http.Redirect(w, r, "/dashboard", http.StatusFound)
})
```

#### Option C: Public Data Only (Limited)

**Use Case:** Only access public Yahoo data

Yahoo Fantasy has some public APIs that don't require auth, but they're very limited.

---

## 📦 Integration Steps

### Step 1: Apply Database Migration

```bash
# Connect to your database
psql $DATABASE_URL -f migrations/011_add_yahoo_fantasy_data.sql
```

### Step 2: Set Environment Variables

```bash
# On Heroku
heroku config:set YAHOO_CLIENT_ID=your_client_id_here
heroku config:set YAHOO_CLIENT_SECRET=your_client_secret_here
heroku config:set YAHOO_REDIRECT_URL=https://nfl.wearemachina.com/auth/yahoo/callback

# After getting refresh token (one-time setup)
heroku config:set YAHOO_REFRESH_TOKEN=your_refresh_token_here
```

### Step 3: Add Routes to Server

Edit `cmd/server/main.go`:

```go
import (
    "github.com/francisco/gridironmind/internal/yahoo"
)

// In main(), after other handlers:

// Initialize Yahoo client
yahooConfig := yahoo.Config{
    ClientID:     cfg.YahooClientID,
    ClientSecret: cfg.YahooClientSecret,
    RedirectURL:  cfg.YahooRedirectURL,
}

yahooClient := yahoo.NewClient(yahooConfig)

// Set token from environment (for server-to-server)
if refreshToken := os.Getenv("YAHOO_REFRESH_TOKEN"); refreshToken != "" {
    token := &oauth2.Token{
        RefreshToken: refreshToken,
    }
    yahooClient.SetToken(token)
}

yahooIngestion := yahoo.NewIngestionService(yahooClient)
yahooHandler := handlers.NewYahooFantasyHandler(yahooIngestion)

// Add Yahoo Fantasy API endpoints
mux.HandleFunc("/api/v1/fantasy/rankings", applyGETMiddleware(yahooHandler.HandlePlayerRankings))
mux.HandleFunc("/api/v1/fantasy/projections/top", applyGETMiddleware(yahooHandler.HandleTopProjections))
mux.HandleFunc("/api/v1/fantasy/projections/", applyGETMiddleware(yahooHandler.HandlePlayerProjection))
mux.HandleFunc("/api/v1/fantasy/ownership", applyGETMiddleware(yahooHandler.HandleOwnershipData))
```

### Step 4: Add to Scheduler (Optional)

Edit `internal/scheduler/scheduler.go` to include Yahoo sync:

```go
// In runSync() method, add Yahoo sync:

// 5. Sync Yahoo Fantasy data (once per day)
if s.config.SyncYahoo && time.Now().Hour() == 4 { // 4am daily
    log.Println("[SCHEDULER] [5/5] Syncing Yahoo fantasy data...")
    if err := s.syncYahooData(seasonInfo); err != nil {
        log.Printf("[SCHEDULER] WARNING syncing Yahoo data: %v", err)
    } else {
        log.Println("[SCHEDULER] ✓ Yahoo data synced successfully")
    }
}

// Add method:
func (s *Scheduler) syncYahooData(seasonInfo utils.SeasonInfo) error {
    yahooClient := getYahooClient() // Your client initialization
    yahooService := yahoo.NewIngestionService(yahooClient)

    // Sync rankings for all positions
    return yahooService.SyncAllPositionRankings(s.ctx, seasonInfo.Year, seasonInfo.CurrentWeek)
}
```

---

## 🔒 Security Best Practices

### 1. Never Commit Credentials

❌ **DON'T:**
```go
const clientSecret = "hardcoded-secret-value-here" // NO!
```

✅ **DO:**
```go
clientSecret := os.Getenv("YAHOO_CLIENT_SECRET")
```

### 2. Encrypt Tokens in Database

The `yahoo_oauth_tokens` table should store encrypted tokens:

```go
import "crypto/aes"

func encryptToken(token string, key []byte) (string, error) {
    // Use AES encryption
    // Store encrypted value in DB
}

func decryptToken(encrypted string, key []byte) (string, error) {
    // Decrypt when loading
}
```

### 3. Rotate Tokens Regularly

OAuth tokens expire. Implement automatic refresh:

```go
func (c *Client) RefreshTokenIfNeeded() error {
    if time.Now().After(c.token.Expiry.Add(-5 * time.Minute)) {
        // Refresh token
        newToken, err := c.config.TokenSource(context.Background(), c.token).Token()
        if err != nil {
            return err
        }
        c.SetToken(newToken)
        // Save to database
    }
    return nil
}
```

---

## 📡 API Endpoints

Once integrated, you'll have these new endpoints:

### Get Player Rankings

```bash
GET /api/v1/fantasy/rankings?position=QB&week=5&season=2025&limit=25

Response:
{
  "data": {
    "season": 2025,
    "week": 5,
    "position": "QB",
    "count": 25,
    "players": [
      {
        "player_id": "uuid",
        "name": "Josh Allen",
        "position": "QB",
        "team": "BUF",
        "overall_rank": 1,
        "position_rank": 1,
        "percent_owned": 99.8,
        "yahoo_player_key": "nfl.p.12345"
      }
    ]
  }
}
```

### Get Player Projection

```bash
GET /api/v1/fantasy/projections/{player_id}?week=5&season=2025

Response:
{
  "data": {
    "player_id": "uuid",
    "season": 2025,
    "week": 5,
    "projection": {
      "projected_points": 24.5,
      "passing_yards": 285,
      "passing_tds": 2,
      "interceptions": 0,
      "rushing_yards": 35,
      "rushing_tds": 1,
      "receptions": 0,
      "receiving_yards": 0,
      "receiving_tds": 0
    }
  }
}
```

### Get Top Projections

```bash
GET /api/v1/fantasy/projections/top?position=RB&week=5&limit=20

Response:
{
  "data": {
    "season": 2025,
    "week": 5,
    "position": "RB",
    "count": 20,
    "players": [
      {
        "id": "uuid",
        "name": "Christian McCaffrey",
        "position": "RB",
        "team": "SF",
        "projected_points": 22.8,
        "projected_rushing_yards": 95,
        "projected_rushing_tds": 1,
        "projected_receptions": 5,
        "projected_receiving_yards": 45,
        "projected_receiving_tds": 0
      }
    ]
  }
}
```

### Get Ownership Data

```bash
GET /api/v1/fantasy/ownership?position=WR&min_owned=50&limit=50

Response:
{
  "data": {
    "season": 2025,
    "week": 5,
    "position": "WR",
    "count": 50,
    "players": [
      {
        "id": "uuid",
        "name": "Tyreek Hill",
        "position": "WR",
        "team": "MIA",
        "percent_owned": 99.5,
        "overall_rank": 5,
        "position_rank": 2
      }
    ]
  }
}
```

---

## 🧪 Testing

### 1. Test OAuth Flow (One-Time Setup)

```bash
# Create a test script
cat > test_yahoo_auth.go <<'EOF'
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/francisco/gridironmind/internal/yahoo"
)

func main() {
    client := yahoo.NewClient(yahoo.Config{
        ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
        ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
        RedirectURL:  "oob", // For command-line testing
    })

    fmt.Println("Visit this URL:", client.GetAuthURL("test"))
    fmt.Print("Enter code: ")

    var code string
    fmt.Scanln(&code)

    token, err := client.ExchangeCode(context.Background(), code)
    if err != nil {
        panic(err)
    }

    fmt.Println("✓ Access Token:", token.AccessToken)
    fmt.Println("✓ Refresh Token:", token.RefreshToken)
    fmt.Println("✓ Expires:", token.Expiry)
    fmt.Println("\nSet this in Heroku:")
    fmt.Printf("heroku config:set YAHOO_REFRESH_TOKEN=%s\n", token.RefreshToken)
}
EOF

go run test_yahoo_auth.go
```

### 2. Test API Client

```go
// Test fetching rankings
func testYahooRankings() {
    client := getYahooClient() // Your setup
    rankings, err := client.FetchPlayerRankings(context.Background(), "QB", 5)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Found %d QB rankings\n", len(rankings.Players))
    for i, player := range rankings.Players[:5] {
        fmt.Printf("%d. %s (%s)\n", i+1, player.Name.Full, player.EditorialTeamAbbr)
    }
}
```

### 3. Test Database Integration

```bash
# After syncing data
psql $DATABASE_URL <<EOF
SELECT COUNT(*) FROM yahoo_player_rankings WHERE season = 2025 AND week = 5;
SELECT p.name, r.overall_rank, r.percent_owned
FROM yahoo_player_rankings r
JOIN players p ON r.player_id = p.id
WHERE r.season = 2025 AND r.week = 5
ORDER BY r.overall_rank
LIMIT 10;
EOF
```

---

## 🚀 Deployment Checklist

- [ ] Apply database migration (011)
- [ ] Set up Yahoo Developer App
- [ ] Get OAuth credentials
- [ ] Complete one-time OAuth flow to get refresh token
- [ ] Set environment variables on Heroku
- [ ] Add routes to `cmd/server/main.go`
- [ ] Deploy to Heroku
- [ ] Test API endpoints
- [ ] (Optional) Add to scheduler for daily sync
- [ ] Monitor logs for errors
- [ ] Set up alerts for token expiration

---

## 💡 Usage Examples

### Fantasy Dashboard Integration

```javascript
// Fetch weekly rankings for QB
fetch('/api/v1/fantasy/rankings?position=QB&week=5&limit=10')
  .then(res => res.json())
  .then(data => {
    data.data.players.forEach(player => {
      console.log(`${player.name}: Rank ${player.overall_rank}, ${player.percent_owned}% owned`);
    });
  });

// Get player projection
fetch('/api/v1/fantasy/projections/PLAYER_UUID?week=5')
  .then(res => res.json())
  .then(data => {
    const proj = data.data.projection;
    console.log(`Projected: ${proj.projected_points} points`);
  });
```

### Start/Sit Decision Tool

```sql
-- Find best waiver wire pickups
SELECT
    p.name,
    p.position,
    t.abbreviation as team,
    r.percent_owned,
    proj.projected_points,
    g.opponent_team_id
FROM yahoo_player_rankings r
JOIN yahoo_player_projections proj ON r.player_id = proj.player_id
    AND r.season = proj.season AND r.week = proj.week
JOIN players p ON r.player_id = p.id
LEFT JOIN teams t ON p.team_id = t.id
LEFT JOIN games g ON g.home_team_id = t.id OR g.away_team_id = t.id
WHERE r.season = 2025
AND r.week = 5
AND r.percent_owned < 50  -- Available in most leagues
AND proj.projected_points > 10  -- Decent projection
ORDER BY proj.projected_points DESC
LIMIT 20;
```

---

## 🎯 Next Steps

1. **Complete OAuth Setup** - Get refresh token
2. **Deploy Database Migration** - Add Yahoo tables
3. **Integrate Routes** - Add to main server
4. **Test Endpoints** - Verify data flow
5. **Add to Dashboard** - Display fantasy data
6. **Monitor Usage** - Track API quota

---

## 📚 Resources

- **Yahoo Fantasy API Docs:** https://developer.yahoo.com/fantasysports/guide/
- **OAuth2 Go Library:** https://pkg.go.dev/golang.org/x/oauth2
- **Yahoo Developer Network:** https://developer.yahoo.com/
- **Rate Limits:** Check Yahoo API documentation for current limits

---

## ⚠️ Important Notes

1. **Yahoo API has rate limits** - Be respectful, cache aggressively
2. **OAuth tokens expire** - Implement auto-refresh
3. **User privacy** - If storing user tokens, comply with privacy laws
4. **Yahoo TOS** - Read and follow Yahoo's terms of service
5. **Data attribution** - Credit Yahoo where required

---

## 🎉 What You Get

Once integrated, your API will have:

✅ **Weekly fantasy rankings** for all positions
✅ **Fantasy point projections** with stat breakdowns
✅ **Ownership percentages** for waiver wire strategy
✅ **Start/sit recommendations** based on matchups
✅ **Trending players** (adds/drops tracking)
✅ **FAAB insights** for auction waivers
✅ **Fantasy news** with impact ratings

This data complements ESPN's real-world stats with fantasy-specific insights, making your API **the most comprehensive NFL data source available**! 🏈
