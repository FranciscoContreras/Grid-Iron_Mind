package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/utils"
)

// GameDetector determines sync mode based on game schedule and current time
type GameDetector struct {
}

// NewGameDetector creates a new game detector
func NewGameDetector() *GameDetector {
	return &GameDetector{}
}

// DetermineMode determines the appropriate sync mode based on current conditions
func (gd *GameDetector) DetermineMode(ctx context.Context) SyncMode {
	seasonInfo := utils.GetCurrentSeason()

	// Offseason - idle mode
	if seasonInfo.IsOffseason {
		return SyncModeIdle
	}

	// Check if there are games scheduled today
	hasGamesToday, err := gd.HasGamesToday(ctx)
	if err != nil {
		log.Printf("[SCHEDULER] Error checking games today: %v", err)
		return SyncModeStandard
	}

	// No games today - standard mode
	if !hasGamesToday {
		return SyncModeStandard
	}

	// Games scheduled today - check if we're in game hours
	if gd.IsGameTime() {
		return SyncModeLive
	}

	// Game day but outside game hours - active mode
	return SyncModeActive
}

// HasGamesToday checks if there are any games scheduled for today
func (gd *GameDetector) HasGamesToday(ctx context.Context) (bool, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT COUNT(*)
		FROM games
		WHERE game_date >= $1 AND game_date < $2
	`

	var count int
	err := db.GetPool().QueryRow(ctx, query, startOfDay, endOfDay).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasLiveGames checks if there are currently live (in-progress) games
func (gd *GameDetector) HasLiveGames(ctx context.Context) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM games
		WHERE status = 'in_progress'
	`

	var count int
	err := db.GetPool().QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// IsGameTime determines if current time is within typical NFL game hours
func (gd *GameDetector) IsGameTime() bool {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()

	switch weekday {
	case time.Thursday:
		// Thursday Night Football: 8pm-11pm ET (5pm-8pm PT)
		return hour >= 17 && hour <= 23

	case time.Sunday:
		// Sunday games: 1pm-11pm ET (10am-8pm PT)
		// Early games: 1pm ET
		// Late games: 4:05pm/4:25pm ET
		// Sunday Night Football: 8:20pm ET
		return hour >= 10 && hour <= 23

	case time.Monday:
		// Monday Night Football: 8pm-11pm ET (5pm-8pm PT)
		return hour >= 17 && hour <= 23

	case time.Saturday:
		// Saturday games (late season): 1pm-8pm ET
		seasonInfo := utils.GetCurrentSeason()
		if seasonInfo.CurrentWeek >= 15 || seasonInfo.IsPostseason {
			return hour >= 10 && hour <= 23
		}
		return false

	default:
		return false
	}
}

// GetNextGameTime returns the next scheduled game time for smarter scheduling
func (gd *GameDetector) GetNextGameTime(ctx context.Context) (*time.Time, error) {
	now := time.Now()

	query := `
		SELECT game_date
		FROM games
		WHERE game_date > $1
		AND status = 'scheduled'
		ORDER BY game_date ASC
		LIMIT 1
	`

	var gameDate time.Time
	err := db.GetPool().QueryRow(ctx, query, now).Scan(&gameDate)
	if err != nil {
		return nil, err
	}

	return &gameDate, nil
}

// ShouldSyncInjuries determines if we should sync injuries (once per day)
func (gd *GameDetector) ShouldSyncInjuries() bool {
	now := time.Now()
	hour := now.Hour()

	// Sync injuries once per day at 6am ET (3am PT)
	return hour == 3
}

// GetGamesSummary returns a summary of today's games for logging
func (gd *GameDetector) GetGamesSummary(ctx context.Context) string {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'scheduled') as scheduled,
			COUNT(*) FILTER (WHERE status = 'in_progress') as live,
			COUNT(*) FILTER (WHERE status = 'completed') as completed
		FROM games
		WHERE game_date >= $1 AND game_date < $2
	`

	var total, scheduled, live, completed int
	err := db.GetPool().QueryRow(ctx, query, startOfDay, endOfDay).Scan(&total, &scheduled, &live, &completed)
	if err != nil {
		return "unknown"
	}

	return fmt.Sprintf("%d total (%d live, %d scheduled, %d completed)", total, live, scheduled, completed)
}
