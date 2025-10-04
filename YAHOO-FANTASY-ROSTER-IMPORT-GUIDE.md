# Yahoo Fantasy Roster Import Implementation Guide

This guide walks you through implementing Yahoo Fantasy API integration to allow users to import their fantasy rosters into Grid Iron Mind.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [OAuth 2.0 Flow](#oauth-20-flow)
4. [Backend Implementation](#backend-implementation)
5. [Frontend Implementation](#frontend-implementation)
6. [API Endpoints](#api-endpoints)
7. [Database Schema](#database-schema)
8. [Testing](#testing)
9. [Deployment](#deployment)
10. [Common Issues](#common-issues)

---

## Overview

### What This Does

Allows users to:
1. **Authenticate** with their Yahoo account (OAuth 2.0)
2. **Select** their fantasy league
3. **Import** their fantasy roster into Grid Iron Mind
4. **Sync** roster updates automatically

### Architecture

```
User Browser
    ↓
[1. Click "Import from Yahoo"]
    ↓
[2. Redirect to Yahoo Login]
    ↓
Yahoo OAuth (User approves)
    ↓
[3. Yahoo Callback with auth code]
    ↓
Grid Iron Mind API
    ↓
[4. Exchange code for access token]
    ↓
[5. Fetch user's leagues]
    ↓
[6. User selects league]
    ↓
[7. Fetch team roster]
    ↓
[8. Import players to database]
    ↓
User's Custom Roster ✅
```

---

## Prerequisites

### 1. Yahoo Developer App Setup

**Create Yahoo App:** https://developer.yahoo.com/apps/create/

```yaml
Application Name: Grid Iron Mind
Description: NFL Fantasy Football Analytics
Redirect URI: https://nfl.wearemachina.com/yahoo/callback
API Permissions:
  - Fantasy Sports (read)
```

**Get Credentials:**
- Client ID: `abc123...`
- Client Secret: `xyz789...`

### 2. Environment Variables

Add to `.env` and Heroku config:

```bash
# Yahoo OAuth
YAHOO_CLIENT_ID=your_client_id_here
YAHOO_CLIENT_SECRET=your_client_secret_here

# Optional: Override redirect URL (defaults to Heroku URL)
YAHOO_REDIRECT_URL=https://nfl.wearemachina.com/yahoo/callback
```

Deploy to Heroku:

```bash
heroku config:set YAHOO_CLIENT_ID="your_client_id" --app grid-iron-mind
heroku config:set YAHOO_CLIENT_SECRET="your_secret" --app grid-iron-mind
```

---

## OAuth 2.0 Flow

### Step-by-Step OAuth Process

#### Step 1: User Clicks "Import from Yahoo"

**Frontend** (`dashboard/index.html`):

```html
<button onclick="startYahooImport()">
    🏈 Import My Yahoo Fantasy Roster
</button>
```

```javascript
function startYahooImport() {
    // Redirect to Yahoo authorization
    window.location.href = '/yahoo/auth';
}
```

#### Step 2: Backend Generates OAuth URL

**Backend** (`internal/handlers/yahoo_oauth.go`):

```go
func HandleYahooAuth(w http.ResponseWriter, r *http.Request) {
    // Generate OAuth URL with state parameter (CSRF protection)
    authURL := yahooOAuthConfig.AuthCodeURL("grid-iron-mind-state")

    // Redirect user to Yahoo login
    http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}
```

**Yahoo Authorization URL Format:**
```
https://api.login.yahoo.com/oauth2/request_auth
    ?client_id=YOUR_CLIENT_ID
    &redirect_uri=https://nfl.wearemachina.com/yahoo/callback
    &response_type=code
    &state=grid-iron-mind-state
```

#### Step 3: User Approves on Yahoo

User sees Yahoo login page:
1. Logs in with Yahoo credentials
2. Reviews permissions request
3. Clicks "Agree" or "Allow"

#### Step 4: Yahoo Redirects Back with Code

**Yahoo callback URL:**
```
https://nfl.wearemachina.com/yahoo/callback
    ?code=AUTH_CODE_HERE
    &state=grid-iron-mind-state
```

#### Step 5: Backend Exchanges Code for Access Token

**Backend** (`internal/handlers/yahoo_oauth.go`):

```go
func HandleYahooCallback(w http.ResponseWriter, r *http.Request) {
    // 1. Validate state (CSRF protection)
    state := r.URL.Query().Get("state")
    if state != "grid-iron-mind-state" {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }

    // 2. Get authorization code
    code := r.URL.Query().Get("code")
    if code == "" {
        http.Error(w, "No code provided", http.StatusBadRequest)
        return
    }

    // 3. Exchange code for access token
    token, err := yahooOAuthConfig.Exchange(r.Context(), code)
    if err != nil {
        http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
        return
    }

    // 4. Store token in session or database
    // (For demo, we'll use a simple in-memory store)
    userID := "user-123" // Get from session/auth
    tokenStore[userID] = token

    // 5. Redirect to league selection page
    http.Redirect(w, r, "/yahoo/leagues", http.StatusSeeOther)
}
```

**Token Response from Yahoo:**
```json
{
  "access_token": "abc123...",
  "token_type": "bearer",
  "expires_in": 3600,
  "refresh_token": "xyz789...",
  "xoauth_yahoo_guid": "USER_GUID"
}
```

---

## Backend Implementation

### 1. Yahoo API Client

**File:** `internal/yahoo/roster_client.go`

```go
package yahoo

import (
    "context"
    "encoding/json"
    "fmt"
    "golang.org/x/oauth2"
)

type RosterClient struct {
    httpClient *http.Client
}

// NewRosterClient creates a client with OAuth token
func NewRosterClient(token *oauth2.Token, config *oauth2.Config) *RosterClient {
    return &RosterClient{
        httpClient: config.Client(context.Background(), token),
    }
}

// GetUserLeagues fetches all leagues for the authenticated user
func (c *RosterClient) GetUserLeagues(ctx context.Context) ([]League, error) {
    url := "https://fantasysports.yahooapis.com/fantasy/v2/users;use_login=1/games;game_keys=nfl/leagues"

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Accept", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result LeaguesResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Fantasy.Users[0].Games[0].Leagues, nil
}

// GetTeamRoster fetches roster for a specific team
func (c *RosterClient) GetTeamRoster(ctx context.Context, teamKey string) (*Roster, error) {
    url := fmt.Sprintf("https://fantasysports.yahooapis.com/fantasy/v2/team/%s/roster", teamKey)

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Accept", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result RosterResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result.Fantasy.Team.Roster, nil
}
```

### 2. Data Models

**File:** `internal/yahoo/models.go`

```go
package yahoo

// League represents a Yahoo Fantasy league
type League struct {
    LeagueKey  string `json:"league_key"`
    LeagueID   string `json:"league_id"`
    Name       string `json:"name"`
    Season     string `json:"season"`
    NumTeams   int    `json:"num_teams"`
}

// Team represents a fantasy team
type Team struct {
    TeamKey  string `json:"team_key"`
    TeamID   string `json:"team_id"`
    Name     string `json:"name"`
    Managers []Manager `json:"managers"`
}

// Manager represents a team manager
type Manager struct {
    ManagerID string `json:"manager_id"`
    Nickname  string `json:"nickname"`
    Guid      string `json:"guid"`
}

// Roster represents a team's roster
type Roster struct {
    Players []RosterPlayer `json:"players"`
}

// RosterPlayer represents a player on a fantasy roster
type RosterPlayer struct {
    PlayerKey       string `json:"player_key"`
    PlayerID        string `json:"player_id"`
    Name            string `json:"name"`
    Position        string `json:"position"`
    Team            string `json:"team"`
    SelectedPosition string `json:"selected_position"` // QB, RB, WR, TE, FLEX, BN, etc.
}

// LeaguesResponse is the API response structure
type LeaguesResponse struct {
    Fantasy struct {
        Users []struct {
            Games []struct {
                Leagues []League `json:"leagues"`
            } `json:"games"`
        } `json:"users"`
    } `json:"fantasy_content"`
}

// RosterResponse is the roster API response structure
type RosterResponse struct {
    Fantasy struct {
        Team struct {
            Roster Roster `json:"roster"`
        } `json:"team"`
    } `json:"fantasy_content"`
}
```

### 3. Roster Import Service

**File:** `internal/yahoo/roster_import.go`

```go
package yahoo

import (
    "context"
    "fmt"
    "log"

    "github.com/francisco/gridironmind/internal/db"
    "github.com/francisco/gridironmind/internal/models"
    "github.com/google/uuid"
)

type RosterImportService struct {
    client       *RosterClient
    playerQueries *db.PlayerQueries
}

func NewRosterImportService(client *RosterClient) *RosterImportService {
    return &RosterImportService{
        client:       client,
        playerQueries: &db.PlayerQueries{},
    }
}

// ImportRoster imports a Yahoo fantasy roster
func (s *RosterImportService) ImportRoster(ctx context.Context, userID, teamKey string) (*models.UserRoster, error) {
    // 1. Fetch roster from Yahoo
    roster, err := s.client.GetTeamRoster(ctx, teamKey)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch roster: %w", err)
    }

    // 2. Create user roster record
    userRoster := &models.UserRoster{
        ID:           uuid.New(),
        UserID:       userID,
        YahooTeamKey: teamKey,
        Players:      []models.Player{},
    }

    // 3. Match Yahoo players to Grid Iron Mind players
    for _, yahooPlayer := range roster.Players {
        player, err := s.matchPlayer(ctx, yahooPlayer)
        if err != nil {
            log.Printf("Warning: Could not match player %s: %v", yahooPlayer.Name, err)
            continue
        }

        userRoster.Players = append(userRoster.Players, *player)
    }

    // 4. Save to database
    if err := s.saveUserRoster(ctx, userRoster); err != nil {
        return nil, fmt.Errorf("failed to save roster: %w", err)
    }

    return userRoster, nil
}

// matchPlayer matches a Yahoo player to a Grid Iron Mind player
func (s *RosterImportService) matchPlayer(ctx context.Context, yahooPlayer RosterPlayer) (*models.Player, error) {
    // Strategy 1: Search by exact name match
    filters := db.PlayerFilters{
        Search: yahooPlayer.Name,
        Status: "active",
        Limit:  10,
    }

    players, _, err := s.playerQueries.ListPlayers(ctx, filters)
    if err != nil {
        return nil, err
    }

    // Find exact match
    for _, player := range players {
        if player.Name == yahooPlayer.Name && player.Position == yahooPlayer.Position {
            return &player, nil
        }
    }

    // Strategy 2: Fuzzy match (close enough)
    if len(players) > 0 {
        return &players[0], nil // Best match
    }

    return nil, fmt.Errorf("no match found for %s", yahooPlayer.Name)
}

// saveUserRoster saves the roster to database
func (s *RosterImportService) saveUserRoster(ctx context.Context, roster *models.UserRoster) error {
    pool := db.GetPool()

    // Insert user roster
    query := `
        INSERT INTO user_rosters (id, user_id, yahoo_team_key, imported_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (user_id, yahoo_team_key) DO UPDATE
        SET imported_at = NOW()
    `

    _, err := pool.Exec(ctx, query, roster.ID, roster.UserID, roster.YahooTeamKey)
    if err != nil {
        return err
    }

    // Insert roster players
    for _, player := range roster.Players {
        playerQuery := `
            INSERT INTO user_roster_players (user_roster_id, player_id)
            VALUES ($1, $2)
            ON CONFLICT DO NOTHING
        `
        _, err := pool.Exec(ctx, playerQuery, roster.ID, player.ID)
        if err != nil {
            log.Printf("Warning: Failed to add player %s to roster: %v", player.Name, err)
        }
    }

    return nil
}
```

### 4. HTTP Handlers

**File:** `internal/handlers/yahoo_roster.go`

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/francisco/gridironmind/internal/yahoo"
    "github.com/francisco/gridironmind/pkg/response"
)

type YahooRosterHandler struct {
    importService *yahoo.RosterImportService
}

func NewYahooRosterHandler(importService *yahoo.RosterImportService) *YahooRosterHandler {
    return &YahooRosterHandler{
        importService: importService,
    }
}

// HandleGetLeagues returns user's Yahoo fantasy leagues
func (h *YahooRosterHandler) HandleGetLeagues(w http.ResponseWriter, r *http.Request) {
    // Get user ID from session/auth
    userID := getUserIDFromSession(r)
    if userID == "" {
        response.Unauthorized(w, "Not authenticated")
        return
    }

    // Get OAuth token from store
    token, ok := tokenStore[userID]
    if !ok {
        response.Unauthorized(w, "No Yahoo token found - please authorize first")
        return
    }

    // Create Yahoo client
    client := yahoo.NewRosterClient(token, yahooOAuthConfig)

    // Fetch leagues
    leagues, err := client.GetUserLeagues(r.Context())
    if err != nil {
        response.InternalError(w, "Failed to fetch leagues")
        return
    }

    response.Success(w, map[string]interface{}{
        "leagues": leagues,
        "count":   len(leagues),
    })
}

// HandleImportRoster imports a fantasy roster
func (h *YahooRosterHandler) HandleImportRoster(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        response.Error(w, 405, "METHOD_NOT_ALLOWED", "Only POST allowed")
        return
    }

    // Parse request body
    var req struct {
        TeamKey string `json:"team_key"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body")
        return
    }

    // Get user ID
    userID := getUserIDFromSession(r)
    if userID == "" {
        response.Unauthorized(w, "Not authenticated")
        return
    }

    // Import roster
    roster, err := h.importService.ImportRoster(r.Context(), userID, req.TeamKey)
    if err != nil {
        response.InternalError(w, "Failed to import roster")
        return
    }

    response.Success(w, map[string]interface{}{
        "roster": roster,
        "message": fmt.Sprintf("Successfully imported %d players", len(roster.Players)),
    })
}

// getUserIDFromSession extracts user ID from session
// TODO: Implement proper session management
func getUserIDFromSession(r *http.Request) string {
    // For demo purposes, return a static ID
    // In production, use proper session/JWT authentication
    return "demo-user-123"
}
```

---

## Frontend Implementation

### 1. OAuth Flow UI

**File:** `dashboard/yahoo-import.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Import Yahoo Fantasy Roster - Grid Iron Mind</title>
    <link rel="stylesheet" href="styles.css">
</head>
<body>
    <div class="container">
        <h1>🏈 Import Your Yahoo Fantasy Roster</h1>

        <!-- Step 1: Authorize -->
        <div id="step-authorize" class="step-card">
            <h2>Step 1: Authorize with Yahoo</h2>
            <p>Connect your Yahoo account to import your fantasy roster.</p>
            <button onclick="startYahooAuth()" class="btn-primary">
                Connect Yahoo Account
            </button>
        </div>

        <!-- Step 2: Select League (hidden initially) -->
        <div id="step-leagues" class="step-card" style="display: none;">
            <h2>Step 2: Select Your League</h2>
            <div id="leagues-list"></div>
        </div>

        <!-- Step 3: Confirm Import (hidden initially) -->
        <div id="step-import" class="step-card" style="display: none;">
            <h2>Step 3: Import Roster</h2>
            <div id="roster-preview"></div>
            <button onclick="confirmImport()" class="btn-primary">
                Import Roster
            </button>
        </div>

        <!-- Step 4: Success (hidden initially) -->
        <div id="step-success" class="step-card success" style="display: none;">
            <h2>✅ Roster Imported Successfully!</h2>
            <p>Your Yahoo fantasy roster has been imported.</p>
            <a href="/dashboard" class="btn-primary">Go to Dashboard</a>
        </div>
    </div>

    <script src="yahoo-import.js"></script>
</body>
</html>
```

### 2. JavaScript Logic

**File:** `dashboard/yahoo-import.js`

```javascript
const API_BASE = 'https://nfl.wearemachina.com/api/v1';
let selectedTeamKey = null;

// Step 1: Start OAuth flow
function startYahooAuth() {
    window.location.href = '/yahoo/auth';
}

// Step 2: Fetch leagues after OAuth callback
async function fetchLeagues() {
    try {
        const response = await fetch(`${API_BASE}/yahoo/leagues`, {
            credentials: 'include' // Include cookies/session
        });

        if (!response.ok) {
            throw new Error('Failed to fetch leagues');
        }

        const data = await response.json();
        displayLeagues(data.data.leagues);

        // Show leagues step, hide authorize step
        document.getElementById('step-authorize').style.display = 'none';
        document.getElementById('step-leagues').style.display = 'block';
    } catch (error) {
        console.error('Error fetching leagues:', error);
        alert('Failed to fetch leagues. Please try again.');
    }
}

// Display leagues as selectable cards
function displayLeagues(leagues) {
    const container = document.getElementById('leagues-list');
    container.innerHTML = '';

    leagues.forEach(league => {
        const card = document.createElement('div');
        card.className = 'league-card';
        card.innerHTML = `
            <h3>${league.name}</h3>
            <p>Season: ${league.season} • ${league.num_teams} teams</p>
            <button onclick="selectLeague('${league.league_key}')">
                Select This League
            </button>
        `;
        container.appendChild(card);
    });
}

// Step 3: Select league and fetch roster preview
async function selectLeague(leagueKey) {
    selectedTeamKey = leagueKey; // Simplified - normally you'd get team_key

    try {
        const response = await fetch(`${API_BASE}/yahoo/team/${selectedTeamKey}/roster`, {
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to fetch roster');
        }

        const data = await response.json();
        displayRosterPreview(data.data.roster);

        // Show import step, hide leagues step
        document.getElementById('step-leagues').style.display = 'none';
        document.getElementById('step-import').style.display = 'block';
    } catch (error) {
        console.error('Error fetching roster:', error);
        alert('Failed to fetch roster. Please try again.');
    }
}

// Display roster preview
function displayRosterPreview(roster) {
    const container = document.getElementById('roster-preview');
    container.innerHTML = '<h3>Your Roster:</h3>';

    const table = document.createElement('table');
    table.innerHTML = `
        <thead>
            <tr>
                <th>Player</th>
                <th>Position</th>
                <th>Team</th>
            </tr>
        </thead>
        <tbody>
            ${roster.players.map(player => `
                <tr>
                    <td>${player.name}</td>
                    <td>${player.position}</td>
                    <td>${player.team}</td>
                </tr>
            `).join('')}
        </tbody>
    `;

    container.appendChild(table);
}

// Step 4: Confirm and import roster
async function confirmImport() {
    try {
        const response = await fetch(`${API_BASE}/yahoo/roster/import`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
            body: JSON.stringify({
                team_key: selectedTeamKey
            })
        });

        if (!response.ok) {
            throw new Error('Failed to import roster');
        }

        // Show success step
        document.getElementById('step-import').style.display = 'none';
        document.getElementById('step-success').style.display = 'block';
    } catch (error) {
        console.error('Error importing roster:', error);
        alert('Failed to import roster. Please try again.');
    }
}

// On page load, check if we're returning from OAuth callback
window.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.get('oauth_callback') === 'success') {
        fetchLeagues();
    }
});
```

---

## API Endpoints

### 1. Register Routes

**File:** `cmd/server/main.go`

```go
// Yahoo OAuth routes
http.HandleFunc("/yahoo/auth", handlers.HandleYahooAuth)
http.HandleFunc("/yahoo/callback", handlers.HandleYahooCallback)

// Yahoo roster routes
yahooRosterHandler := handlers.NewYahooRosterHandler(yahooImportService)
http.HandleFunc("/api/v1/yahoo/leagues", yahooRosterHandler.HandleGetLeagues)
http.HandleFunc("/api/v1/yahoo/roster/import", yahooRosterHandler.HandleImportRoster)
http.HandleFunc("/api/v1/yahoo/team/", yahooRosterHandler.HandleGetTeamRoster)
```

### 2. Endpoint Summary

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/yahoo/auth` | GET | Redirects to Yahoo OAuth |
| `/yahoo/callback` | GET | OAuth callback handler |
| `/api/v1/yahoo/leagues` | GET | Get user's fantasy leagues |
| `/api/v1/yahoo/team/:key/roster` | GET | Get team roster |
| `/api/v1/yahoo/roster/import` | POST | Import roster to database |

---

## Database Schema

### Migration File

**File:** `migrations/013_add_user_rosters.sql`

```sql
-- Migration 013: Add Yahoo Fantasy roster tables
-- Created: 2025-10-03

-- User rosters table (links users to their Yahoo fantasy teams)
CREATE TABLE IF NOT EXISTS user_rosters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    yahoo_team_key VARCHAR(100) NOT NULL,
    league_name VARCHAR(255),
    team_name VARCHAR(255),
    imported_at TIMESTAMP DEFAULT NOW(),
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_id, yahoo_team_key)
);

-- User roster players (many-to-many: rosters <-> players)
CREATE TABLE IF NOT EXISTS user_roster_players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_roster_id UUID NOT NULL REFERENCES user_rosters(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    roster_position VARCHAR(20), -- QB, RB, WR, TE, FLEX, BN, etc.
    added_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(user_roster_id, player_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_rosters_user_id ON user_rosters(user_id);
CREATE INDEX IF NOT EXISTS idx_user_rosters_yahoo_key ON user_rosters(yahoo_team_key);
CREATE INDEX IF NOT EXISTS idx_user_roster_players_roster ON user_roster_players(user_roster_id);
CREATE INDEX IF NOT EXISTS idx_user_roster_players_player ON user_roster_players(player_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at
CREATE TRIGGER update_user_rosters_updated_at
    BEFORE UPDATE ON user_rosters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Apply Migration

```bash
# Locally
psql $DATABASE_URL -f migrations/013_add_user_rosters.sql

# Heroku
heroku pg:psql --app grid-iron-mind < migrations/013_add_user_rosters.sql
```

---

## Testing

### 1. Manual Testing Flow

#### Step 1: Test OAuth Flow

```bash
# Open browser
open https://nfl.wearemachina.com/yahoo-import.html

# Click "Connect Yahoo Account"
# → Should redirect to Yahoo login
# → Log in with Yahoo credentials
# → Approve app
# → Should redirect back to /yahoo/callback
# → Should show leagues list
```

#### Step 2: Test League Fetching

```bash
# Test leagues endpoint
curl -X GET "https://nfl.wearemachina.com/api/v1/yahoo/leagues" \
  -H "Cookie: session=YOUR_SESSION_ID" \
  | jq '.'
```

**Expected Response:**
```json
{
  "data": {
    "leagues": [
      {
        "league_key": "nfl.l.12345",
        "league_id": "12345",
        "name": "My Fantasy League",
        "season": "2025",
        "num_teams": 12
      }
    ],
    "count": 1
  }
}
```

#### Step 3: Test Roster Import

```bash
# Import roster
curl -X POST "https://nfl.wearemachina.com/api/v1/yahoo/roster/import" \
  -H "Content-Type: application/json" \
  -H "Cookie: session=YOUR_SESSION_ID" \
  -d '{
    "team_key": "nfl.l.12345.t.1"
  }' \
  | jq '.'
```

**Expected Response:**
```json
{
  "data": {
    "roster": {
      "id": "abc-123-uuid",
      "user_id": "demo-user-123",
      "yahoo_team_key": "nfl.l.12345.t.1",
      "players": [
        {
          "id": "player-uuid-1",
          "name": "Patrick Mahomes",
          "position": "QB"
        },
        {
          "id": "player-uuid-2",
          "name": "Saquon Barkley",
          "position": "RB"
        }
      ]
    },
    "message": "Successfully imported 15 players"
  }
}
```

### 2. Database Verification

```sql
-- Check imported rosters
SELECT * FROM user_rosters WHERE user_id = 'demo-user-123';

-- Check roster players
SELECT
    urp.roster_position,
    p.name,
    p.position,
    p.team_id
FROM user_roster_players urp
JOIN players p ON p.id = urp.player_id
WHERE urp.user_roster_id = 'YOUR_ROSTER_UUID';
```

### 3. End-to-End Test Script

**File:** `scripts/test-yahoo-import.sh`

```bash
#!/bin/bash

# Test Yahoo Fantasy Roster Import

BASE_URL="https://nfl.wearemachina.com"

echo "1. Testing OAuth URL generation..."
curl -I "${BASE_URL}/yahoo/auth"

echo "\n2. Testing leagues endpoint (requires auth)..."
# Note: You'll need a valid session cookie
curl -X GET "${BASE_URL}/api/v1/yahoo/leagues" \
  -H "Cookie: session=YOUR_SESSION" \
  -s | jq '.data.leagues[] | {name, league_key}'

echo "\n3. Testing roster import..."
curl -X POST "${BASE_URL}/api/v1/yahoo/roster/import" \
  -H "Content-Type: application/json" \
  -H "Cookie: session=YOUR_SESSION" \
  -d '{"team_key": "nfl.l.12345.t.1"}' \
  -s | jq '.data.message'

echo "\n✅ Tests complete!"
```

---

## Deployment

### 1. Deploy Backend Changes

```bash
# Add all files
git add migrations/013_add_user_rosters.sql
git add internal/yahoo/roster_client.go
git add internal/yahoo/roster_import.go
git add internal/handlers/yahoo_roster.go
git add dashboard/yahoo-import.html
git add dashboard/yahoo-import.js

# Commit
git commit -m "Add Yahoo Fantasy roster import functionality

- Implement OAuth 2.0 flow for Yahoo
- Add roster import service
- Create user_rosters database tables
- Add frontend UI for roster import
- Support league selection and player matching"

# Deploy to Heroku
git push origin main
git push heroku main
```

### 2. Set Heroku Config

```bash
# Set Yahoo credentials
heroku config:set YAHOO_CLIENT_ID="your_client_id" --app grid-iron-mind
heroku config:set YAHOO_CLIENT_SECRET="your_secret" --app grid-iron-mind

# Verify
heroku config --app grid-iron-mind | grep YAHOO
```

### 3. Run Migration

```bash
# Run database migration on Heroku
heroku run "psql \$DATABASE_URL -f migrations/013_add_user_rosters.sql" --app grid-iron-mind

# Or use release command (already configured in Procfile)
# Migration will run automatically on next deploy
```

### 4. Verify Deployment

```bash
# Check web dyno is running
heroku ps --app grid-iron-mind

# View logs
heroku logs --tail --app grid-iron-mind

# Test OAuth endpoint
curl -I https://nfl.wearemachina.com/yahoo/auth
# Should return 307 Temporary Redirect to Yahoo
```

---

## Common Issues

### Issue 1: "invalid_client" Error

**Symptom:** OAuth returns `{"error": "invalid_client"}`

**Causes:**
- Wrong Client ID or Secret
- Redirect URI mismatch

**Fix:**
```bash
# Verify credentials
heroku config:get YAHOO_CLIENT_ID --app grid-iron-mind
heroku config:get YAHOO_CLIENT_SECRET --app grid-iron-mind

# Check redirect URI in Yahoo Developer Console
# Must match exactly: https://nfl.wearemachina.com/yahoo/callback
```

### Issue 2: "Invalid Redirect URI"

**Symptom:** Yahoo shows error: "The redirect_uri doesn't match"

**Fix:**
1. Go to Yahoo Developer Console
2. Edit your app
3. Set Redirect URI to: `https://nfl.wearemachina.com/yahoo/callback`
4. Save and retry

### Issue 3: Token Expired

**Symptom:** API returns 401 Unauthorized after some time

**Cause:** Access tokens expire after 1 hour

**Fix:** Implement token refresh:

```go
func refreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
    token := &oauth2.Token{
        RefreshToken: refreshToken,
    }

    tokenSource := yahooOAuthConfig.TokenSource(ctx, token)
    newToken, err := tokenSource.Token()
    if err != nil {
        return nil, err
    }

    return newToken, nil
}
```

### Issue 4: Player Matching Fails

**Symptom:** Many players show "Could not match player" warnings

**Causes:**
- Name variations (e.g., "Pat Mahomes" vs "Patrick Mahomes")
- Missing players in database

**Fix:**
1. Improve matching algorithm with fuzzy search:

```go
func matchPlayerFuzzy(ctx context.Context, yahooPlayer RosterPlayer) (*models.Player, error) {
    // Extract last name
    nameParts := strings.Split(yahooPlayer.Name, " ")
    lastName := nameParts[len(nameParts)-1]

    // Search by last name + position
    filters := db.PlayerFilters{
        Search:   lastName,
        Position: yahooPlayer.Position,
        Status:   "active",
        Limit:    5,
    }

    players, _, err := playerQueries.ListPlayers(ctx, filters)
    if err != nil {
        return nil, err
    }

    // Use Levenshtein distance for best match
    bestMatch := findBestMatch(yahooPlayer.Name, players)
    return bestMatch, nil
}
```

2. Ensure player database is populated:
```bash
# Run data sync to import all players
heroku run "./nfl-data-pipeline/target/release/nfl-data-pipeline --mode update" --app grid-iron-mind
```

### Issue 5: CORS Errors

**Symptom:** Browser console shows CORS errors

**Fix:** Add CORS headers in middleware:

```go
func (h *YahooRosterHandler) HandleImportRoster(w http.ResponseWriter, r *http.Request) {
    // Add CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    // ... rest of handler
}
```

---

## Next Steps

### Phase 1: Basic Import ✅
- [x] OAuth 2.0 flow
- [x] League selection
- [x] Roster import
- [x] Player matching

### Phase 2: Enhancements
- [ ] Auto-sync rosters (weekly refresh)
- [ ] Support multiple rosters per user
- [ ] Trade notifications
- [ ] Waiver wire tracking

### Phase 3: Advanced Features
- [ ] Roster optimization suggestions
- [ ] Start/sit recommendations
- [ ] Matchup analysis
- [ ] Trade analyzer

---

## Resources

- **Yahoo Fantasy API Docs:** https://developer.yahoo.com/fantasysports/guide/
- **OAuth 2.0 Spec:** https://oauth.net/2/
- **Go OAuth2 Library:** https://pkg.go.dev/golang.org/x/oauth2
- **Yahoo Developer Console:** https://developer.yahoo.com/apps/

---

## Support

If you encounter issues:

1. Check Heroku logs: `heroku logs --tail --app grid-iron-mind`
2. Verify environment variables: `heroku config --app grid-iron-mind`
3. Test endpoints with curl (see Testing section)
4. Review Yahoo API docs for endpoint changes

---

**Last Updated:** October 3, 2025
**Version:** 1.0
**Author:** Grid Iron Mind Team
