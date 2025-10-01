package autofetch

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/ingestion"
	"github.com/francisco/gridironmind/internal/models"
	"github.com/francisco/gridironmind/internal/utils"
	"github.com/google/uuid"
)

// Orchestrator handles automatic data fetching when data is missing
type Orchestrator struct {
	ingestionService *ingestion.Service
	mu               sync.Mutex
	fetchInProgress  map[string]bool
}

// NewOrchestrator creates a new auto-fetch orchestrator
func NewOrchestrator(weatherAPIKey string) *Orchestrator {
	return &Orchestrator{
		ingestionService: ingestion.NewService(weatherAPIKey),
		fetchInProgress:  make(map[string]bool),
	}
}

// FetchKey generates a unique key for tracking in-progress fetches
func fetchKey(resource, identifier string) string {
	return fmt.Sprintf("%s:%s", resource, identifier)
}

// isAlreadyFetching checks if a fetch is already in progress
func (o *Orchestrator) isAlreadyFetching(key string) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.fetchInProgress[key]
}

// markFetchStart marks a fetch as in progress
func (o *Orchestrator) markFetchStart(key string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.fetchInProgress[key] = true
}

// markFetchComplete marks a fetch as complete
func (o *Orchestrator) markFetchComplete(key string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.fetchInProgress, key)
}

// FetchGamesIfMissing fetches games for a specific season/week if they don't exist
func (o *Orchestrator) FetchGamesIfMissing(ctx context.Context, season, week int) error {
	// Check if we should auto-fetch for this season/week
	if !utils.ShouldFetchGames(season, week) {
		return fmt.Errorf("auto-fetch not allowed for season %d week %d", season, week)
	}

	key := fetchKey("games", fmt.Sprintf("%d-week-%d", season, week))

	// Check if already fetching
	if o.isAlreadyFetching(key) {
		return fmt.Errorf("fetch already in progress for season %d week %d", season, week)
	}

	// Mark as fetching
	o.markFetchStart(key)
	defer o.markFetchComplete(key)

	log.Printf("[AUTO-FETCH] Fetching games for season %d week %d", season, week)

	// Ensure teams exist first (games require teams)
	if err := o.ensureTeamsExist(ctx); err != nil {
		log.Printf("[AUTO-FETCH] Failed to ensure teams exist: %v", err)
		return err
	}

	// Fetch games from ESPN
	if err := o.ingestionService.SyncGames(ctx); err != nil {
		log.Printf("[AUTO-FETCH] Failed to sync games: %v", err)
		return err
	}

	log.Printf("[AUTO-FETCH] Successfully fetched games for season %d week %d", season, week)
	return nil
}

// FetchAllSeasonGames fetches all games for a given season if missing
func (o *Orchestrator) FetchAllSeasonGames(ctx context.Context, season int) error {
	key := fetchKey("games", fmt.Sprintf("%d-all", season))

	if o.isAlreadyFetching(key) {
		return fmt.Errorf("fetch already in progress for season %d", season)
	}

	o.markFetchStart(key)
	defer o.markFetchComplete(key)

	log.Printf("[AUTO-FETCH] Fetching all games for season %d", season)

	// Ensure teams exist
	if err := o.ensureTeamsExist(ctx); err != nil {
		return err
	}

	// Fetch all games
	if err := o.ingestionService.SyncGames(ctx); err != nil {
		return err
	}

	log.Printf("[AUTO-FETCH] Successfully fetched all games for season %d", season)
	return nil
}

// FetchPlayerIfMissing fetches a player by NFL ID if they don't exist
func (o *Orchestrator) FetchPlayerIfMissing(ctx context.Context, nflID int) (*models.Player, error) {
	key := fetchKey("player", fmt.Sprintf("%d", nflID))

	if o.isAlreadyFetching(key) {
		return nil, fmt.Errorf("fetch already in progress for player %d", nflID)
	}

	o.markFetchStart(key)
	defer o.markFetchComplete(key)

	log.Printf("[AUTO-FETCH] Fetching player with NFL ID %d", nflID)

	// Ensure teams exist (players need teams)
	if err := o.ensureTeamsExist(ctx); err != nil {
		return nil, err
	}

	// Sync rosters (will fetch all players)
	if err := o.ingestionService.SyncRosters(ctx); err != nil {
		log.Printf("[AUTO-FETCH] Failed to sync rosters: %v", err)
		return nil, err
	}

	// Query for the specific player
	queries := &db.PlayerQueries{}
	player, err := queries.GetByNFLID(ctx, nflID)
	if err != nil {
		return nil, fmt.Errorf("player not found after fetch: %w", err)
	}

	log.Printf("[AUTO-FETCH] Successfully fetched player %s", player.Name)
	return player, nil
}

// FetchTeamIfMissing fetches a team by NFL ID if it doesn't exist
func (o *Orchestrator) FetchTeamIfMissing(ctx context.Context, nflID int) (*models.Team, error) {
	key := fetchKey("team", fmt.Sprintf("%d", nflID))

	if o.isAlreadyFetching(key) {
		return nil, fmt.Errorf("fetch already in progress for team %d", nflID)
	}

	o.markFetchStart(key)
	defer o.markFetchComplete(key)

	log.Printf("[AUTO-FETCH] Fetching team with NFL ID %d", nflID)

	// Sync all teams
	if err := o.ingestionService.SyncTeams(ctx); err != nil {
		return nil, err
	}

	// Query for the specific team
	queries := &db.TeamQueries{}
	team, err := queries.GetByNFLID(ctx, nflID)
	if err != nil {
		return nil, fmt.Errorf("team not found after fetch: %w", err)
	}

	log.Printf("[AUTO-FETCH] Successfully fetched team %s", team.Name)
	return team, nil
}

// FetchStatsIfMissing fetches player stats for a game if they don't exist
func (o *Orchestrator) FetchStatsIfMissing(ctx context.Context, gameID uuid.UUID) error {
	key := fetchKey("stats", gameID.String())

	if o.isAlreadyFetching(key) {
		return fmt.Errorf("fetch already in progress for game %s", gameID)
	}

	o.markFetchStart(key)
	defer o.markFetchComplete(key)

	log.Printf("[AUTO-FETCH] Fetching stats for game %s", gameID)

	// Get game to determine season
	gameQueries := &db.GameQueries{}
	game, err := gameQueries.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Sync stats from NFLverse
	if err := o.ingestionService.SyncPlayerStats(ctx, game.Season); err != nil {
		log.Printf("[AUTO-FETCH] Failed to sync player stats: %v", err)
		return err
	}

	log.Printf("[AUTO-FETCH] Successfully fetched stats for game %s", gameID)
	return nil
}

// ensureTeamsExist checks if teams are in database, fetches if missing
func (o *Orchestrator) ensureTeamsExist(ctx context.Context) error {
	queries := &db.TeamQueries{}
	teams, err := queries.ListTeams(ctx)
	if err != nil {
		return err
	}

	// If we have all 32 teams, we're good
	if len(teams) >= 32 {
		return nil
	}

	log.Printf("[AUTO-FETCH] Only %d teams found, fetching all teams", len(teams))

	// Fetch teams
	key := fetchKey("teams", "all")
	if o.isAlreadyFetching(key) {
		// Wait a bit and retry
		time.Sleep(2 * time.Second)
		return nil
	}

	o.markFetchStart(key)
	defer o.markFetchComplete(key)

	return o.ingestionService.SyncTeams(ctx)
}

// AutoFetchGameData is a convenience method that fetches games and related data
func (o *Orchestrator) AutoFetchGameData(ctx context.Context, season, week int) error {
	log.Printf("[AUTO-FETCH] Auto-fetching game data for season %d week %d", season, week)

	// Fetch games
	if err := o.FetchGamesIfMissing(ctx, season, week); err != nil {
		return err
	}

	// Games are now in database
	return nil
}

// AutoFetchCurrentWeek fetches data for the current NFL week
func (o *Orchestrator) AutoFetchCurrentWeek(ctx context.Context) error {
	seasonInfo := utils.GetCurrentSeason()

	if seasonInfo.IsOffseason {
		return fmt.Errorf("currently in offseason, no games to fetch")
	}

	return o.AutoFetchGameData(ctx, seasonInfo.Year, seasonInfo.CurrentWeek)
}
