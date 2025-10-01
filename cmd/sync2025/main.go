package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/espn"
	"github.com/francisco/gridironmind/internal/ingestion"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// sync2025 is a command-line tool to load and keep updated all 2025 NFL season data
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Validate required environment variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	weatherAPIKey := os.Getenv("WEATHER_API_KEY")

	// Initialize database connection
	if err := db.InitDB(context.Background(), dbURL); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Create ingestion service
	service := ingestion.NewService(weatherAPIKey)

	// Parse command line arguments
	mode := "full"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	ctx := context.Background()

	switch mode {
	case "full":
		// Full sync: Load all 2025 data from scratch
		log.Println("========================================")
		log.Println("FULL 2025 SEASON SYNC")
		log.Println("========================================")
		if err := fullSync2025(ctx, service); err != nil {
			log.Fatalf("Full sync failed: %v", err)
		}

	case "update":
		// Update sync: Refresh current week's data
		log.Println("========================================")
		log.Println("UPDATE 2025 SEASON DATA")
		log.Println("========================================")
		if err := updateSync2025(ctx, service); err != nil {
			log.Fatalf("Update sync failed: %v", err)
		}

	case "live":
		// Live sync: Continuous updates during game day
		log.Println("========================================")
		log.Println("LIVE 2025 SEASON SYNC")
		log.Println("========================================")
		if err := liveSync2025(ctx, service); err != nil {
			log.Fatalf("Live sync failed: %v", err)
		}

	case "stats":
		// Stats only: Just update player stats
		log.Println("========================================")
		log.Println("2025 PLAYER STATS SYNC")
		log.Println("========================================")
		if err := statsSync2025(ctx, service); err != nil {
			log.Fatalf("Stats sync failed: %v", err)
		}

	case "injuries":
		// Injuries only: Just update injury reports
		log.Println("========================================")
		log.Println("2025 INJURY REPORTS SYNC")
		log.Println("========================================")
		if err := injuriesSync2025(ctx, service); err != nil {
			log.Fatalf("Injuries sync failed: %v", err)
		}

	default:
		log.Fatalf("Unknown mode: %s. Use: full, update, live, stats, or injuries", mode)
	}

	log.Println("========================================")
	log.Println("SYNC COMPLETED SUCCESSFULLY")
	log.Println("========================================")
}

// fullSync2025 performs a complete initial load of all 2025 season data
func fullSync2025(ctx context.Context, service *ingestion.Service) error {
	season := 2025

	// Step 1: Sync teams (always start here)
	log.Println("\n[1/6] Syncing teams...")
	if err := service.SyncTeams(ctx); err != nil {
		return fmt.Errorf("teams sync failed: %w", err)
	}

	// Step 2: Sync all rosters
	log.Println("\n[2/6] Syncing rosters for all teams...")
	if err := service.SyncAllRosters(ctx); err != nil {
		return fmt.Errorf("rosters sync failed: %w", err)
	}

	// Step 3: Sync all games for 2025 season (all 18 weeks)
	log.Println("\n[3/6] Syncing all 2025 games...")
	if err := service.SyncHistoricalGames(ctx, season); err != nil {
		return fmt.Errorf("games sync failed: %w", err)
	}

	// Step 4: Sync team stats for completed games
	log.Println("\n[4/6] Syncing team stats for completed games...")
	if err := syncCompletedGameStats(ctx, service, season); err != nil {
		log.Printf("Warning: Some team stats sync failed: %v", err)
		// Don't fail the entire sync for stats errors
	}

	// Step 5: Sync player season stats
	log.Println("\n[5/6] Syncing player season stats from nflverse...")
	if err := service.SyncNFLversePlayerStats(ctx, season); err != nil {
		log.Printf("Warning: NFLverse player stats sync failed: %v", err)
		// Continue anyway - we can try ESPN stats
	}

	// Step 6: Sync injury reports
	log.Println("\n[6/6] Syncing injury reports...")
	if err := service.SyncAllTeamInjuries(ctx); err != nil {
		log.Printf("Warning: Injuries sync failed: %v", err)
		// Don't fail for injuries
	}

	return nil
}

// updateSync2025 refreshes data for the current week
func updateSync2025(ctx context.Context, service *ingestion.Service) error {
	season := 2025
	currentWeek := getCurrentNFLWeek()

	log.Printf("Updating data for 2025 Season, Week %d\n", currentWeek)

	// Step 1: Sync current rosters (players may have been traded/signed)
	log.Println("\n[1/5] Syncing current rosters...")
	if err := service.SyncAllRosters(ctx); err != nil {
		return fmt.Errorf("rosters sync failed: %w", err)
	}

	// Step 2: Sync games for current week
	log.Printf("\n[2/5] Syncing Week %d games...\n", currentWeek)
	if err := syncWeekGames(ctx, service, season, currentWeek); err != nil {
		return fmt.Errorf("games sync failed: %w", err)
	}

	// Step 3: Sync team stats for current week's completed games
	log.Printf("\n[3/5] Syncing Week %d team stats...\n", currentWeek)
	if err := service.SyncGameTeamStats(ctx, season, currentWeek); err != nil {
		log.Printf("Warning: Team stats sync failed: %v", err)
	}

	// Step 4: Update player season stats
	log.Println("\n[4/5] Updating player season stats...")
	if err := service.SyncNFLversePlayerStats(ctx, season); err != nil {
		log.Printf("Warning: Player stats sync failed: %v", err)
	}

	// Step 5: Update injury reports
	log.Println("\n[5/5] Updating injury reports...")
	if err := service.SyncAllTeamInjuries(ctx); err != nil {
		log.Printf("Warning: Injuries sync failed: %v", err)
	}

	return nil
}

// liveSync2025 runs continuous updates during game day (every 5 minutes)
func liveSync2025(ctx context.Context, service *ingestion.Service) error {
	season := 2025
	currentWeek := getCurrentNFLWeek()

	log.Printf("Starting live sync for 2025 Season, Week %d\n", currentWeek)
	log.Println("Will update every 5 minutes. Press Ctrl+C to stop.")

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	if err := runLiveSyncIteration(ctx, service, season, currentWeek); err != nil {
		log.Printf("Error in sync iteration: %v", err)
	}

	// Then run every 5 minutes
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping live sync")
			return ctx.Err()
		case <-ticker.C:
			if err := runLiveSyncIteration(ctx, service, season, currentWeek); err != nil {
				log.Printf("Error in sync iteration: %v", err)
				// Don't exit, keep trying
			}
		}
	}
}

// runLiveSyncIteration performs one iteration of live sync
func runLiveSyncIteration(ctx context.Context, service *ingestion.Service, season, week int) error {
	log.Printf("\n=== Live Sync at %s ===\n", time.Now().Format("15:04:05"))

	// Sync current week's games (scores, status updates)
	if err := syncWeekGames(ctx, service, season, week); err != nil {
		return fmt.Errorf("games sync failed: %w", err)
	}

	// Sync team stats for newly completed games
	if err := service.SyncGameTeamStats(ctx, season, week); err != nil {
		log.Printf("Warning: Team stats sync failed: %v", err)
	}

	log.Println("Live sync iteration completed")
	return nil
}

// statsSync2025 syncs only player statistics for 2025
func statsSync2025(ctx context.Context, service *ingestion.Service) error {
	season := 2025

	log.Println("Syncing player stats from nflverse...")
	if err := service.SyncNFLversePlayerStats(ctx, season); err != nil {
		return fmt.Errorf("nflverse stats sync failed: %w", err)
	}

	// Also sync from ESPN for each player
	log.Println("Syncing individual player career stats from ESPN...")
	if err := syncAllPlayerCareerStats(ctx, service); err != nil {
		log.Printf("Warning: ESPN player stats sync had errors: %v", err)
	}

	return nil
}

// injuriesSync2025 syncs only injury reports
func injuriesSync2025(ctx context.Context, service *ingestion.Service) error {
	return service.SyncAllTeamInjuries(ctx)
}

// Helper functions

// syncWeekGames syncs games for a specific week
func syncWeekGames(ctx context.Context, service *ingestion.Service, season, week int) error {
	// The SyncHistoricalGames method handles syncing for all weeks
	// For a single week, we can fetch and process games directly

	// Create ESPN client
	espnClient := espn.NewClient()
	scoreboard, err := espnClient.FetchSeasonGames(ctx, season, week)
	if err != nil {
		return fmt.Errorf("failed to fetch week %d games: %w", week, err)
	}

	log.Printf("Found %d games for Week %d", len(scoreboard.Events), week)

	// Sync games using the service
	if err := service.SyncGames(ctx); err != nil {
		return fmt.Errorf("failed to sync games: %w", err)
	}

	return nil
}

// syncCompletedGameStats syncs team stats for all completed games in a season
func syncCompletedGameStats(ctx context.Context, service *ingestion.Service, season int) error {
	// Sync stats for each week
	for week := 1; week <= 18; week++ {
		log.Printf("Syncing team stats for Week %d...", week)
		if err := service.SyncGameTeamStats(ctx, season, week); err != nil {
			log.Printf("Warning: Failed to sync team stats for week %d: %v", week, err)
			// Continue with other weeks
		}
		// Rate limiting
		time.Sleep(2 * time.Second)
	}
	return nil
}

// syncAllPlayerCareerStats syncs career stats for all players from ESPN
func syncAllPlayerCareerStats(ctx context.Context, service *ingestion.Service) error {
	// Get all players from database
	pool := db.GetPool()
	rows, err := pool.Query(ctx, "SELECT id, nfl_id FROM players WHERE status = 'active' LIMIT 100")
	if err != nil {
		return fmt.Errorf("failed to fetch players: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var playerID uuid.UUID
		var nflID int
		if err := rows.Scan(&playerID, &nflID); err != nil {
			log.Printf("Error scanning player: %v", err)
			continue
		}

		espnID := strconv.Itoa(nflID)
		if err := service.SyncPlayerCareerStats(ctx, playerID, espnID); err != nil {
			log.Printf("Failed to sync career stats for player %s: %v", playerID, err)
			continue
		}

		count++
		if count%10 == 0 {
			log.Printf("Synced %d players...", count)
		}

		// Rate limiting - be nice to ESPN
		time.Sleep(1 * time.Second)
	}

	log.Printf("Synced career stats for %d players", count)
	return nil
}

// getCurrentNFLWeek determines the current week of the NFL season
// This is a simplified version - you may want to make this more sophisticated
func getCurrentNFLWeek() int {
	now := time.Now()
	// NFL season typically starts first Thursday in September
	// Week 1 is usually around Sept 5-11
	seasonStart := time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC)

	if now.Before(seasonStart) {
		return 1 // Preseason/week 1
	}

	daysSinceStart := now.Sub(seasonStart).Hours() / 24
	week := int(daysSinceStart/7) + 1

	if week > 18 {
		return 18 // Regular season is 18 weeks
	}

	return week
}
