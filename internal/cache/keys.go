package cache

import (
	"fmt"
	"time"
)

// Cache TTL durations
const (
	TTLTeams     = 1 * time.Hour      // Teams change rarely
	TTLPlayers   = 15 * time.Minute   // Players update moderately
	TTLGames     = 5 * time.Minute    // Games update frequently during season
	TTLStats     = 5 * time.Minute    // Stats update with games
	TTLLeaders   = 10 * time.Minute   // Leaders update less frequently
)

// Cache key patterns
const (
	KeyPrefixTeams   = "teams"
	KeyPrefixPlayers = "players"
	KeyPrefixGames   = "games"
	KeyPrefixStats   = "stats"
	KeyPrefixLeaders = "leaders"
)

// TeamsCacheKey generates cache key for teams list
func TeamsCacheKey() string {
	return "teams:list"
}

// TeamCacheKey generates cache key for single team
func TeamCacheKey(teamID string) string {
	return fmt.Sprintf("teams:%s", teamID)
}

// TeamRosterCacheKey generates cache key for team roster
func TeamRosterCacheKey(teamID string) string {
	return fmt.Sprintf("teams:%s:players", teamID)
}

// PlayersCacheKey generates cache key for players list with filters
func PlayersCacheKey(position, team, status string, limit, offset int) string {
	return fmt.Sprintf("players:list:pos=%s:team=%s:status=%s:limit=%d:offset=%d",
		position, team, status, limit, offset)
}

// PlayerCacheKey generates cache key for single player
func PlayerCacheKey(playerID string) string {
	return fmt.Sprintf("players:%s", playerID)
}

// PlayerStatsCacheKey generates cache key for player stats
func PlayerStatsCacheKey(playerID string, season, week, limit, offset int) string {
	return fmt.Sprintf("stats:player:%s:season=%d:week=%d:limit=%d:offset=%d",
		playerID, season, week, limit, offset)
}

// GamesCacheKey generates cache key for games list with filters
func GamesCacheKey(season, week int, teamID, status string, limit, offset int) string {
	return fmt.Sprintf("games:list:season=%d:week=%d:team=%s:status=%s:limit=%d:offset=%d",
		season, week, teamID, status, limit, offset)
}

// GameCacheKey generates cache key for single game
func GameCacheKey(gameID string) string {
	return fmt.Sprintf("games:%s", gameID)
}

// GameStatsCacheKey generates cache key for game stats
func GameStatsCacheKey(gameID string) string {
	return fmt.Sprintf("stats:game:%s", gameID)
}

// StatsLeadersCacheKey generates cache key for stats leaders
func StatsLeadersCacheKey(category string, season, limit int) string {
	return fmt.Sprintf("leaders:%s:season=%d:limit=%d", category, season, limit)
}

// InvalidateTeamsCache invalidates all team-related cache keys
func InvalidateTeamsCache() string {
	return "teams:*"
}

// InvalidatePlayersCache invalidates all player-related cache keys
func InvalidatePlayersCache() string {
	return "players:*"
}

// InvalidateGamesCache invalidates all game-related cache keys
func InvalidateGamesCache() string {
	return "games:*"
}

// InvalidateStatsCache invalidates all stats-related cache keys
func InvalidateStatsCache() string {
	return "stats:*"
}

// InvalidateLeadersCache invalidates all leaders-related cache keys
func InvalidateLeadersCache() string {
	return "leaders:*"
}

// InvalidateAllCache returns pattern for all cache keys
func InvalidateAllCache() string {
	return "*"
}