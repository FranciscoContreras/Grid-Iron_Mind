package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/ingestion"
	"github.com/francisco/gridironmind/internal/utils"
)

// Scheduler manages automatic data synchronization based on game schedule
type Scheduler struct {
	config       Config
	service      *ingestion.Service
	detector     *GameDetector
	currentMode  SyncMode
	running      bool
	mu           sync.RWMutex
	lastSync     time.Time
	lastError    error
	syncCount    int
	errorCount   int
	ctx          context.Context
	cancel       context.CancelFunc
}

// Status represents the current scheduler status
type Status struct {
	Enabled      bool      `json:"enabled"`
	Running      bool      `json:"running"`
	CurrentMode  SyncMode  `json:"current_mode"`
	NextSync     time.Time `json:"next_sync"`
	LastSync     time.Time `json:"last_sync"`
	LastError    string    `json:"last_error,omitempty"`
	SyncCount    int       `json:"sync_count"`
	ErrorCount   int       `json:"error_count"`
	Interval     string    `json:"interval"`
	SeasonInfo   string    `json:"season_info"`
	GamesSummary string    `json:"games_summary,omitempty"`
}

// NewScheduler creates a new scheduler instance
func NewScheduler(config Config) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		config:   config,
		service:  ingestion.NewService(config.WeatherAPIKey),
		detector: NewGameDetector(),
		running:  false,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the scheduler's sync loop
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		log.Println("[SCHEDULER] Already running")
		return
	}
	s.running = true
	s.mu.Unlock()

	log.Println("[SCHEDULER] Starting auto-sync scheduler...")
	log.Printf("[SCHEDULER] Configuration: enabled=%v, mode=%s",
		s.config.Enabled, s.config.Mode)

	// Run initial sync immediately
	go s.runSyncLoop()
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	log.Println("[SCHEDULER] Stopping scheduler...")
	s.cancel()
	s.running = false
	log.Println("[SCHEDULER] Scheduler stopped")
}

// runSyncLoop is the main sync loop that runs in a goroutine
func (s *Scheduler) runSyncLoop() {
	// Run initial sync
	s.runSync()

	for {
		// Determine current mode
		mode := s.determineMode()
		s.updateMode(mode)

		// Get interval for current mode
		interval := s.config.GetInterval(mode)

		log.Printf("[SCHEDULER] Next sync in %v (mode: %s)", interval, mode)

		// Wait for next sync or shutdown signal
		select {
		case <-s.ctx.Done():
			log.Println("[SCHEDULER] Shutdown signal received")
			return
		case <-time.After(interval):
			s.runSync()
		}
	}
}

// runSync executes a single sync iteration
func (s *Scheduler) runSync() {
	if !s.config.Enabled {
		log.Println("[SCHEDULER] Sync skipped (disabled)")
		return
	}

	log.Println("========================================")
	log.Printf("[SCHEDULER] Starting sync iteration at %s", time.Now().Format("15:04:05"))
	log.Println("========================================")

	startTime := time.Now()

	// Get season info
	seasonInfo := utils.GetCurrentSeason()
	log.Printf("[SCHEDULER] Season: %d, Week: %d, Active: %v",
		seasonInfo.Year, seasonInfo.CurrentWeek, !seasonInfo.IsOffseason)

	// Get games summary if possible
	if gamesSummary := s.detector.GetGamesSummary(s.ctx); gamesSummary != "" {
		log.Printf("[SCHEDULER] Today's games: %s", gamesSummary)
	}

	// Execute sync operations
	var syncError error

	// 1. Sync games (scores, status, details)
	if s.config.SyncGames {
		log.Println("[SCHEDULER] [1/4] Syncing games...")
		if err := s.syncGames(seasonInfo); err != nil {
			log.Printf("[SCHEDULER] ERROR syncing games: %v", err)
			syncError = fmt.Errorf("games sync failed: %w", err)
		} else {
			log.Println("[SCHEDULER] ✓ Games synced successfully")
		}
	}

	// 2. Sync team stats for completed games
	if s.config.SyncStats && !seasonInfo.IsOffseason {
		log.Println("[SCHEDULER] [2/4] Syncing team stats...")
		if err := s.syncTeamStats(seasonInfo); err != nil {
			log.Printf("[SCHEDULER] WARNING syncing team stats: %v", err)
			// Don't fail entire sync for stats errors
		} else {
			log.Println("[SCHEDULER] ✓ Team stats synced successfully")
		}
	}

	// 3. Sync injuries (once per day)
	if s.config.SyncInjuries && s.detector.ShouldSyncInjuries() {
		log.Println("[SCHEDULER] [3/4] Syncing injury reports...")
		if err := s.service.SyncAllTeamInjuries(s.ctx); err != nil {
			log.Printf("[SCHEDULER] WARNING syncing injuries: %v", err)
			// Don't fail for injury sync errors
		} else {
			log.Println("[SCHEDULER] ✓ Injuries synced successfully")
		}
	} else {
		log.Println("[SCHEDULER] [3/4] Skipping injury sync (not scheduled)")
	}

	// 4. Clear cache for updated data
	if s.config.ClearCache {
		log.Println("[SCHEDULER] [4/4] Clearing cache...")
		if err := s.clearRelevantCache(); err != nil {
			log.Printf("[SCHEDULER] WARNING clearing cache: %v", err)
		} else {
			log.Println("[SCHEDULER] ✓ Cache cleared successfully")
		}
	}

	duration := time.Since(startTime)

	// Update status
	s.mu.Lock()
	s.lastSync = time.Now()
	s.syncCount++
	if syncError != nil {
		s.lastError = syncError
		s.errorCount++
	} else {
		s.lastError = nil
	}
	s.mu.Unlock()

	log.Println("========================================")
	if syncError != nil {
		log.Printf("[SCHEDULER] Sync completed with ERRORS in %v", duration)
	} else {
		log.Printf("[SCHEDULER] Sync completed successfully in %v", duration)
	}
	log.Println("========================================")
}

// syncGames syncs current week's games
func (s *Scheduler) syncGames(seasonInfo utils.SeasonInfo) error {
	// During offseason, skip game sync
	if seasonInfo.IsOffseason {
		log.Println("[SCHEDULER] Skipping game sync (offseason)")
		return nil
	}

	// Sync current week's games
	log.Printf("[SCHEDULER] Fetching games for season %d, week %d",
		seasonInfo.Year, seasonInfo.CurrentWeek)

	// Use the existing SyncGames method which fetches current scoreboard
	if err := s.service.SyncGames(s.ctx); err != nil {
		return fmt.Errorf("failed to sync current games: %w", err)
	}

	return nil
}

// syncTeamStats syncs team statistics for completed games
func (s *Scheduler) syncTeamStats(seasonInfo utils.SeasonInfo) error {
	// Sync current week's team stats
	log.Printf("[SCHEDULER] Syncing team stats for season %d, week %d",
		seasonInfo.Year, seasonInfo.CurrentWeek)

	if err := s.service.SyncGameTeamStats(s.ctx, seasonInfo.Year, seasonInfo.CurrentWeek); err != nil {
		return fmt.Errorf("failed to sync team stats: %w", err)
	}

	return nil
}

// clearRelevantCache clears cache entries that may have been updated
func (s *Scheduler) clearRelevantCache() error {
	// Clear game-related cache entries
	patterns := []string{
		"games:*",
		"game:*",
		"teams:*",
		"team:*",
		"stats:*",
		"standings:*",
		"defense:*",
	}

	for _, pattern := range patterns {
		if err := cache.DeletePattern(s.ctx, pattern); err != nil {
			log.Printf("[SCHEDULER] Warning: Failed to clear cache pattern %s: %v", pattern, err)
			// Don't fail, just log
		}
	}

	return nil
}

// determineMode determines the sync mode based on current conditions
func (s *Scheduler) determineMode() SyncMode {
	// If mode is manually set in config, use that
	if s.config.Mode != "" {
		return s.config.Mode
	}

	// Otherwise, auto-detect based on game schedule
	return s.detector.DetermineMode(s.ctx)
}

// updateMode updates the current mode
func (s *Scheduler) updateMode(mode SyncMode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentMode != mode {
		log.Printf("[SCHEDULER] Mode changed: %s → %s", s.currentMode, mode)
		s.currentMode = mode
	}
}

// GetStatus returns the current scheduler status
func (s *Scheduler) GetStatus() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mode := s.currentMode
	if mode == "" {
		mode = s.determineMode()
	}

	interval := s.config.GetInterval(mode)
	nextSync := s.lastSync.Add(interval)
	if s.lastSync.IsZero() {
		nextSync = time.Now()
	}

	seasonInfo := utils.GetCurrentSeason()
	seasonInfoStr := fmt.Sprintf("Season %d, Week %d", seasonInfo.Year, seasonInfo.CurrentWeek)
	if seasonInfo.IsOffseason {
		seasonInfoStr += " (Offseason)"
	} else if seasonInfo.IsRegular {
		seasonInfoStr += " (Regular Season)"
	} else if seasonInfo.IsPostseason {
		seasonInfoStr += " (Postseason)"
	}

	var lastErrorStr string
	if s.lastError != nil {
		lastErrorStr = s.lastError.Error()
	}

	gamesSummary := ""
	if ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second); cancel != nil {
		defer cancel()
		gamesSummary = s.detector.GetGamesSummary(ctx)
	}

	return Status{
		Enabled:      s.config.Enabled,
		Running:      s.running,
		CurrentMode:  mode,
		NextSync:     nextSync,
		LastSync:     s.lastSync,
		LastError:    lastErrorStr,
		SyncCount:    s.syncCount,
		ErrorCount:   s.errorCount,
		Interval:     interval.String(),
		SeasonInfo:   seasonInfoStr,
		GamesSummary: gamesSummary,
	}
}

// TriggerSync manually triggers a sync (useful for admin endpoint)
func (s *Scheduler) TriggerSync() {
	log.Println("[SCHEDULER] Manual sync triggered")
	go s.runSync()
}

// UpdateConfig updates the scheduler configuration
func (s *Scheduler) UpdateConfig(newConfig Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[SCHEDULER] Updating configuration: enabled=%v, mode=%s",
		newConfig.Enabled, newConfig.Mode)
	s.config = newConfig
}
