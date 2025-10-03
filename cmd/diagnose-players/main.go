package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/espn"
)

// Top fantasy players to check (2024-2025 season leaders)
var topFantasyPlayers = []string{
	"Saquon Barkley",
	"Lamar Jackson",
	"Josh Allen",
	"Jalen Hurts",
	"Derrick Henry",
	"Joe Burrow",
	"Ja'Marr Chase",
	"Amon-Ra St. Brown",
	"Justin Jefferson",
	"CeeDee Lamb",
	"Tyreek Hill",
	"Travis Kelce",
	"Sam LaPorta",
	"Christian McCaffrey",
	"Bijan Robinson",
	"Breece Hall",
	"Jahmyr Gibbs",
	"De'Von Achane",
	"Patrick Mahomes",
	"Kyler Murray",
	"A.J. Brown",
	"Nico Collins",
	"Puka Nacua",
	"Cooper Kupp",
	"Mike Evans",
	"Garrett Wilson",
	"Drake London",
	"Deebo Samuel",
	"George Kittle",
	"Trey McBride",
}

func main() {
	log.Println("=== Grid Iron Mind - Player Diagnostic Tool ===")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dbConfig := db.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    cfg.DBMaxConns,
		MinConns:    cfg.DBMinConns,
	}

	if err := db.Connect(context.Background(), dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Database connected")

	// Create ESPN client
	client := espn.NewClient()
	ctx := context.Background()

	// Check each top fantasy player
	log.Println("\n=== Checking Top Fantasy Players ===")
	missing := []string{}
	found := []string{}

	for _, playerName := range topFantasyPlayers {
		// Check if player exists in database
		exists, dbPlayerName := checkPlayerInDB(ctx, playerName)

		if exists {
			found = append(found, fmt.Sprintf("✓ %s (DB: %s)", playerName, dbPlayerName))
		} else {
			missing = append(missing, playerName)
			// Try to find in ESPN
			espnPlayer := findPlayerInESPN(ctx, client, playerName)
			if espnPlayer != "" {
				log.Printf("✗ %s - MISSING in DB but FOUND in ESPN: %s", playerName, espnPlayer)
			} else {
				log.Printf("✗ %s - NOT FOUND in DB or ESPN", playerName)
			}
		}
	}

	// Print summary
	log.Println("\n=== SUMMARY ===")
	log.Printf("Total players checked: %d", len(topFantasyPlayers))
	log.Printf("Found in database: %d", len(found))
	log.Printf("Missing from database: %d", len(missing))

	if len(missing) > 0 {
		log.Println("\n=== MISSING PLAYERS ===")
		for _, name := range missing {
			log.Printf("  - %s", name)
		}
	}

	if len(found) > 0 {
		log.Println("\n=== FOUND PLAYERS ===")
		for _, name := range found {
			log.Printf("  %s", name)
		}
	}

	// Check total player count
	totalPlayers := getTotalPlayerCount(ctx)
	log.Printf("\n=== DATABASE STATS ===")
	log.Printf("Total players in database: %d", totalPlayers)

	// Exit with error code if players are missing
	if len(missing) > 0 {
		os.Exit(1)
	}
}

func checkPlayerInDB(ctx context.Context, name string) (bool, string) {
	pool := db.GetPool()

	// Try exact match first
	query := `SELECT name FROM players WHERE LOWER(name) = LOWER($1) LIMIT 1`
	var dbName string
	err := pool.QueryRow(ctx, query, name).Scan(&dbName)
	if err == nil {
		return true, dbName
	}

	// Try partial match (e.g., "Saquon Barkley" vs "S. Barkley")
	query = `SELECT name FROM players WHERE LOWER(name) LIKE LOWER($1) LIMIT 1`
	parts := strings.Split(name, " ")
	if len(parts) >= 2 {
		lastName := parts[len(parts)-1]
		pattern := fmt.Sprintf("%%%s%%", lastName)
		err = pool.QueryRow(ctx, query, pattern).Scan(&dbName)
		if err == nil {
			return true, dbName
		}
	}

	return false, ""
}

func findPlayerInESPN(ctx context.Context, client *espn.Client, name string) string {
	// Fetch active players from ESPN
	resp, err := client.FetchActivePlayers(ctx, 100)
	if err != nil {
		log.Printf("Failed to fetch ESPN players: %v", err)
		return ""
	}

	// Search for player in response
	nameLower := strings.ToLower(name)
	for _, athlete := range resp.Items {
		fullName := strings.ToLower(athlete.FullName)
		if fullName == nameLower || strings.Contains(fullName, nameLower) {
			data, _ := json.MarshalIndent(athlete, "", "  ")
			return string(data)
		}
	}

	return ""
}

func getTotalPlayerCount(ctx context.Context) int {
	pool := db.GetPool()

	query := `SELECT COUNT(*) FROM players`
	var count int
	err := pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		log.Printf("Failed to get player count: %v", err)
		return 0
	}

	return count
}
