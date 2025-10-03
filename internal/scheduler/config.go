package scheduler

import (
	"time"
)

// SyncMode defines the scheduler's operating mode
type SyncMode string

const (
	// SyncModeLive - Every 1 minute during live game times
	SyncModeLive SyncMode = "live"

	// SyncModeActive - Every 5 minutes on game days outside game hours
	SyncModeActive SyncMode = "active"

	// SyncModeStandard - Every 15 minutes on non-game days during season
	SyncModeStandard SyncMode = "standard"

	// SyncModeIdle - Every hour during offseason
	SyncModeIdle SyncMode = "idle"

	// SyncModeDisabled - No automatic syncing
	SyncModeDisabled SyncMode = "disabled"
)

// Config holds scheduler configuration
type Config struct {
	// Enabled - Whether auto-sync is enabled
	Enabled bool

	// Mode - Current sync mode (overrides automatic detection if set)
	Mode SyncMode

	// WeatherAPIKey - For weather enrichment during sync
	WeatherAPIKey string

	// Intervals for each mode
	LiveInterval     time.Duration
	ActiveInterval   time.Duration
	StandardInterval time.Duration
	IdleInterval     time.Duration

	// SyncGames - Whether to sync game scores/status
	SyncGames bool

	// SyncStats - Whether to sync team/player stats
	SyncStats bool

	// SyncInjuries - Whether to sync injury reports
	SyncInjuries bool

	// ClearCache - Whether to clear cache after sync
	ClearCache bool
}

// DefaultConfig returns the default scheduler configuration
func DefaultConfig(weatherAPIKey string) Config {
	return Config{
		Enabled:          true,
		Mode:             "", // Auto-detect
		WeatherAPIKey:    weatherAPIKey,
		LiveInterval:     1 * time.Minute,
		ActiveInterval:   5 * time.Minute,
		StandardInterval: 15 * time.Minute,
		IdleInterval:     1 * time.Hour,
		SyncGames:        true,
		SyncStats:        true,
		SyncInjuries:     true,
		ClearCache:       true,
	}
}

// GetInterval returns the sync interval for the current mode
func (c *Config) GetInterval(mode SyncMode) time.Duration {
	switch mode {
	case SyncModeLive:
		return c.LiveInterval
	case SyncModeActive:
		return c.ActiveInterval
	case SyncModeStandard:
		return c.StandardInterval
	case SyncModeIdle:
		return c.IdleInterval
	default:
		return c.StandardInterval
	}
}
