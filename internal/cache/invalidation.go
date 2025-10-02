package cache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// InvalidationStrategy defines cache invalidation strategies
type InvalidationStrategy string

const (
	// InvalidateAll removes all cached data
	InvalidateAll InvalidationStrategy = "all"
	// InvalidatePlayers removes player-related cache
	InvalidatePlayers InvalidationStrategy = "players"
	// InvalidateTeams removes team-related cache
	InvalidateTeams InvalidationStrategy = "teams"
	// InvalidateGames removes game-related cache
	InvalidateGames InvalidationStrategy = "games"
	// InvalidateStats removes stats-related cache
	InvalidateStats InvalidationStrategy = "stats"
)

// InvalidationManager handles cache invalidation
type InvalidationManager struct{}

// NewInvalidationManager creates a new invalidation manager
func NewInvalidationManager() *InvalidationManager {
	return &InvalidationManager{}
}

// InvalidateByStrategy invalidates cache based on strategy
func (m *InvalidationManager) InvalidateByStrategy(ctx context.Context, strategy InvalidationStrategy) error {
	switch strategy {
	case InvalidateAll:
		return m.invalidateAll(ctx)
	case InvalidatePlayers:
		return m.invalidateByPattern(ctx, "player*")
	case InvalidateTeams:
		return m.invalidateByPattern(ctx, "team*")
	case InvalidateGames:
		return m.invalidateByPattern(ctx, "game*")
	case InvalidateStats:
		return m.invalidateByPattern(ctx, "stats*")
	default:
		return fmt.Errorf("unknown invalidation strategy: %s", strategy)
	}
}

// InvalidatePlayer invalidates all cache for a specific player
func (m *InvalidationManager) InvalidatePlayer(ctx context.Context, playerID uuid.UUID) error {
	patterns := []string{
		fmt.Sprintf("player:%s*", playerID),
		"players:list*", // List queries might include this player
	}

	for _, pattern := range patterns {
		if err := m.invalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Error invalidating pattern %s: %v", pattern, err)
		}
	}

	log.Printf("[CACHE] Invalidated player: %s", playerID)
	return nil
}

// InvalidateTeam invalidates all cache for a specific team
func (m *InvalidationManager) InvalidateTeam(ctx context.Context, teamID uuid.UUID) error {
	patterns := []string{
		fmt.Sprintf("team:%s*", teamID),
		"teams:list*",
		fmt.Sprintf("players:*team=%s*", teamID), // Players filtered by team
	}

	for _, pattern := range patterns {
		if err := m.invalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Error invalidating pattern %s: %v", pattern, err)
		}
	}

	log.Printf("[CACHE] Invalidated team: %s", teamID)
	return nil
}

// InvalidateGame invalidates all cache for a specific game
func (m *InvalidationManager) InvalidateGame(ctx context.Context, gameID uuid.UUID) error {
	patterns := []string{
		fmt.Sprintf("game:%s*", gameID),
		"games:list*",
		fmt.Sprintf("stats:game:%s*", gameID),
	}

	for _, pattern := range patterns {
		if err := m.invalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Error invalidating pattern %s: %v", pattern, err)
		}
	}

	log.Printf("[CACHE] Invalidated game: %s", gameID)
	return nil
}

// InvalidateSeasonWeek invalidates cache for a specific season/week
func (m *InvalidationManager) InvalidateSeasonWeek(ctx context.Context, season int, week int) error {
	patterns := []string{
		fmt.Sprintf("games:*season=%d*week=%d*", season, week),
		fmt.Sprintf("stats:*season=%d*week=%d*", season, week),
	}

	for _, pattern := range patterns {
		if err := m.invalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Error invalidating pattern %s: %v", pattern, err)
		}
	}

	log.Printf("[CACHE] Invalidated season %d week %d", season, week)
	return nil
}

// InvalidateAfterSync invalidates cache after data sync operations
func (m *InvalidationManager) InvalidateAfterSync(ctx context.Context, syncType string) error {
	var patterns []string

	switch syncType {
	case "teams":
		patterns = []string{"team*"}
	case "players", "rosters":
		patterns = []string{"player*", "team*"} // Teams include roster info
	case "games":
		patterns = []string{"game*"}
	case "stats":
		patterns = []string{"stats*", "game*"} // Stats affect game details
	case "injuries":
		patterns = []string{"player*", "injuries*"}
	case "full":
		return m.invalidateAll(ctx)
	}

	for _, pattern := range patterns {
		if err := m.invalidateByPattern(ctx, pattern); err != nil {
			log.Printf("Error invalidating pattern %s: %v", pattern, err)
		}
	}

	log.Printf("[CACHE] Invalidated after sync: %s", syncType)
	return nil
}

// invalidateByPattern removes all keys matching a pattern
func (m *InvalidationManager) invalidateByPattern(ctx context.Context, pattern string) error {
	if redisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	// Get all keys matching pattern
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		log.Printf("[CACHE] No keys to invalidate for pattern: %s", pattern)
		return nil
	}

	// Delete all matching keys
	deleted, err := redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	log.Printf("[CACHE] Invalidated %d keys for pattern: %s", deleted, pattern)
	return nil
}

// invalidateAll removes all cache
func (m *InvalidationManager) invalidateAll(ctx context.Context) error {
	if redisClient == nil {
		return fmt.Errorf("redis not initialized")
	}

	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush cache: %w", err)
	}

	log.Printf("[CACHE] Invalidated all cache")
	return nil
}

// WarmCache pre-loads frequently accessed data into cache
func (m *InvalidationManager) WarmCache(ctx context.Context, warmType string) error {
	log.Printf("[CACHE] Warming cache: %s", warmType)

	switch warmType {
	case "teams":
		// In a real implementation, you would fetch and cache all teams
		log.Printf("[CACHE] Teams cache warming not implemented")
	case "current_week":
		// Cache current week's games
		log.Printf("[CACHE] Current week cache warming not implemented")
	}

	return nil
}

// CacheMetrics returns cache statistics
func (m *InvalidationManager) CacheMetrics(ctx context.Context) (map[string]interface{}, error) {
	if redisClient == nil {
		return map[string]interface{}{
			"error": "redis not initialized",
		}, nil
	}

	// Get cache info
	info, err := redisClient.Info(ctx, "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache info: %w", err)
	}

	// Parse info string
	metrics := make(map[string]interface{})
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			metrics[parts[0]] = parts[1]
		}
	}

	// Get key count
	dbSize, err := redisClient.DBSize(ctx).Result()
	if err == nil {
		metrics["total_keys"] = dbSize
	}

	return metrics, nil
}

// ScheduledInvalidation runs periodic cache invalidation
func (m *InvalidationManager) ScheduledInvalidation(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("[CACHE] Running scheduled invalidation...")

			// Invalidate old game caches (games >7 days old)
			// This is a placeholder - implement based on your needs

		case <-ctx.Done():
			log.Printf("[CACHE] Stopping scheduled invalidation")
			return
		}
	}
}
