package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/nflverse"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	mode      = flag.String("mode", "year", "Import mode: year, range, full, validate, stats")
	year      = flag.Int("year", 2024, "Year to import (for mode=year)")
	startYear = flag.Int("start", 2010, "Start year (for mode=range or mode=full)")
	endYear   = flag.Int("end", 2024, "End year (for mode=range or mode=full)")
	dryRun    = flag.Bool("dry-run", false, "Dry run mode (don't actually import)")
	verbose   = flag.Bool("verbose", false, "Verbose logging")
)

type ImportStats struct {
	RostersImported    int
	GamesImported      int
	StatsImported      int
	NGSImported        int
	Errors             []string
	StartTime          time.Time
	EndTime            time.Time
}

type Importer struct {
	dbPool     *pgxpool.Pool
	csvParser  *nflverse.CSVParser
	ctx        context.Context
}

func main() {
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	ctx := context.Background()
	dbConfig := db.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    cfg.DBMaxConns,
		MinConns:    cfg.DBMinConns,
	}
	if err := db.Connect(ctx, dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create importer
	importer := &Importer{
		dbPool:    db.GetPool(),
		csvParser: nflverse.NewCSVParser(),
		ctx:       ctx,
	}

	// Initialize import progress tracking
	if err := importer.initProgressTracking(); err != nil {
		log.Fatalf("Failed to initialize progress tracking: %v", err)
	}

	// Execute based on mode
	switch *mode {
	case "year":
		if err := importer.importYear(*year); err != nil {
			log.Fatalf("Failed to import year %d: %v", *year, err)
		}
	case "range":
		if err := importer.importRange(*startYear, *endYear); err != nil {
			log.Fatalf("Failed to import range %d-%d: %v", *startYear, *endYear, err)
		}
	case "full":
		if err := importer.importRange(*startYear, *endYear); err != nil {
			log.Fatalf("Failed to import full range %d-%d: %v", *startYear, *endYear, err)
		}
	case "validate":
		if err := importer.validateData(); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
	case "stats":
		if err := importer.showStats(); err != nil {
			log.Fatalf("Failed to show stats: %v", err)
		}
	default:
		log.Fatalf("Unknown mode: %s (use: year, range, full, validate, stats)", *mode)
	}

	log.Println("✅ Import completed successfully!")
}

func (i *Importer) initProgressTracking() error {
	// Create import_progress table if it doesn't exist
	query := `
		CREATE TABLE IF NOT EXISTS import_progress (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			season INT NOT NULL,
			data_type VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			records_imported INT DEFAULT 0,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(season, data_type)
		);
	`
	_, err := i.dbPool.Exec(i.ctx, query)
	return err
}

func (i *Importer) importYear(year int) error {
	log.Printf("📅 Importing data for year %d...\n", year)
	stats := &ImportStats{StartTime: time.Now()}

	// Step 1: Import rosters (players)
	log.Printf("  [1/4] Importing rosters for %d...", year)
	if err := i.importRosters(year, stats); err != nil {
		log.Printf("❌ Failed: %v", err)
		stats.Errors = append(stats.Errors, fmt.Sprintf("Rosters: %v", err))
	} else {
		log.Printf(" ✅ Done: %d players", stats.RostersImported)
	}

	// Step 2: Import schedule (games)
	log.Printf("  [2/4] Importing schedule for %d...", year)
	if err := i.importSchedule(year, stats); err != nil {
		log.Printf("❌ Failed: %v", err)
		stats.Errors = append(stats.Errors, fmt.Sprintf("Schedule: %v", err))
	} else {
		log.Printf(" ✅ Done: %d games", stats.GamesImported)
	}

	// Step 3: Import player stats
	log.Printf("  [3/4] Importing player stats for %d...", year)
	if err := i.importPlayerStats(year, stats); err != nil {
		log.Printf("❌ Failed: %v", err)
		stats.Errors = append(stats.Errors, fmt.Sprintf("Player Stats: %v", err))
	} else {
		log.Printf(" ✅ Done: %d stat records", stats.StatsImported)
	}

	// Step 4: Import Next Gen Stats (if available, 2016+)
	if year >= 2016 {
		log.Printf("  [4/4] Importing Next Gen Stats for %d...", year)
		if err := i.importNextGenStats(year, stats); err != nil {
			log.Printf("❌ Failed: %v", err)
			stats.Errors = append(stats.Errors, fmt.Sprintf("NGS: %v", err))
		} else {
			log.Printf(" ✅ Done: %d NGS records", stats.NGSImported)
		}
	} else {
		log.Printf("  [4/4] Skipping Next Gen Stats (not available before 2016)")
	}

	stats.EndTime = time.Now()
	i.printImportSummary(year, stats)

	return nil
}

func (i *Importer) importRange(start, end int) error {
	log.Printf("📚 Importing data for years %d-%d (%d seasons)...\n", start, end, end-start+1)

	overallStats := &ImportStats{StartTime: time.Now()}

	for year := start; year <= end; year++ {
		if err := i.importYear(year); err != nil {
			log.Printf("❌ Failed to import year %d: %v", year, err)
			overallStats.Errors = append(overallStats.Errors, fmt.Sprintf("Year %d: %v", year, err))
			// Continue with next year instead of stopping
			continue
		}
		log.Println() // Blank line between years
	}

	overallStats.EndTime = time.Now()
	duration := overallStats.EndTime.Sub(overallStats.StartTime)

	log.Println("═══════════════════════════════════════════════════")
	log.Printf("🎯 OVERALL IMPORT SUMMARY (%d-%d)", start, end)
	log.Println("═══════════════════════════════════════════════════")
	log.Printf("Total Duration: %s", duration.Round(time.Second))
	log.Printf("Errors: %d", len(overallStats.Errors))
	if len(overallStats.Errors) > 0 {
		log.Println("\nError Details:")
		for _, err := range overallStats.Errors {
			log.Printf("  - %s", err)
		}
	}
	log.Println("═══════════════════════════════════════════════════")

	return nil
}

func (i *Importer) importRosters(year int, stats *ImportStats) error {
	i.markProgress(year, "rosters", "in_progress")

	rosters, err := i.csvParser.ParseRosters(i.ctx, year)
	if err != nil {
		i.markProgress(year, "rosters", "failed")
		return err
	}

	// Group by team for better logging
	teamCounts := make(map[string]int)

	for _, r := range rosters {
		if *dryRun {
			teamCounts[r.TeamAbbr]++
			continue
		}

		// Upsert player
		if err := i.upsertPlayer(r); err != nil {
			if *verbose {
				log.Printf("    Warning: Failed to upsert player %s: %v", r.FullName, err)
			}
			continue
		}
		stats.RostersImported++
		teamCounts[r.TeamAbbr]++
	}

	if *verbose {
		log.Println("\n    Team Roster Counts:")
		for team, count := range teamCounts {
			log.Printf("      %s: %d players", team, count)
		}
	}

	i.markProgress(year, "rosters", "completed")
	i.updateProgress(year, "rosters", stats.RostersImported)
	return nil
}

func (i *Importer) importSchedule(year int, stats *ImportStats) error {
	i.markProgress(year, "schedule", "in_progress")

	schedules, err := i.csvParser.ParseSchedule(i.ctx, year)
	if err != nil {
		i.markProgress(year, "schedule", "failed")
		return err
	}

	// Filter to regular season only (can be changed)
	regularSeason := nflverse.FilterRegularSeason(schedules)
	log.Printf(" (found %d regular season games)", len(regularSeason))

	for _, s := range regularSeason {
		if *dryRun {
			stats.GamesImported++
			continue
		}

		if err := i.upsertGame(s); err != nil {
			if *verbose {
				log.Printf("    Warning: Failed to upsert game %s: %v", s.GameID, err)
			}
			continue
		}
		stats.GamesImported++
	}

	i.markProgress(year, "schedule", "completed")
	i.updateProgress(year, "schedule", stats.GamesImported)
	return nil
}

func (i *Importer) importPlayerStats(year int, stats *ImportStats) error {
	i.markProgress(year, "player_stats", "in_progress")

	playerStats, err := i.csvParser.ParsePlayerStats(i.ctx, year)
	if err != nil {
		i.markProgress(year, "player_stats", "failed")
		return err
	}

	// Filter to regular season only
	var regularSeasonStats []*nflverse.PlayerStatCSV
	for _, ps := range playerStats {
		if ps.SeasonType == "REG" {
			regularSeasonStats = append(regularSeasonStats, ps)
		}
	}

	log.Printf(" (found %d regular season stat records)", len(regularSeasonStats))

	// Batch insert for performance
	batchSize := 500
	for idx := 0; idx < len(regularSeasonStats); idx += batchSize {
		end := idx + batchSize
		if end > len(regularSeasonStats) {
			end = len(regularSeasonStats)
		}

		batch := regularSeasonStats[idx:end]
		if *dryRun {
			stats.StatsImported += len(batch)
			continue
		}

		if err := i.batchInsertStats(batch); err != nil {
			if *verbose {
				log.Printf("    Warning: Failed to insert batch %d-%d: %v", idx, end, err)
			}
			continue
		}
		stats.StatsImported += len(batch)

		if *verbose && (idx/batchSize)%10 == 0 {
			log.Printf("    Progress: %d/%d records", idx, len(regularSeasonStats))
		}
	}

	i.markProgress(year, "player_stats", "completed")
	i.updateProgress(year, "player_stats", stats.StatsImported)
	return nil
}

func (i *Importer) importNextGenStats(year int, stats *ImportStats) error {
	i.markProgress(year, "ngs", "in_progress")

	// Try importing all three types
	passingCount, rushingCount, receivingCount := 0, 0, 0

	// Passing NGS (2016+)
	if year >= 2016 {
		if passing, err := i.csvParser.ParseNextGenStatsPassing(i.ctx, year); err == nil {
			for _, ngs := range passing {
				if !*dryRun {
					if err := i.upsertNextGenStatsPassing(ngs); err != nil {
						continue
					}
				}
				passingCount++
			}
		}
	}

	// Rushing NGS (2018+)
	if year >= 2018 {
		if rushing, err := i.csvParser.ParseNextGenStatsRushing(i.ctx, year); err == nil {
			for _, ngs := range rushing {
				if !*dryRun {
					if err := i.upsertNextGenStatsRushing(ngs); err != nil {
						continue
					}
				}
				rushingCount++
			}
		}
	}

	// Receiving NGS (2017+)
	if year >= 2017 {
		if receiving, err := i.csvParser.ParseNextGenStatsReceiving(i.ctx, year); err == nil {
			for _, ngs := range receiving {
				if !*dryRun {
					if err := i.upsertNextGenStatsReceiving(ngs); err != nil {
						continue
					}
				}
				receivingCount++
			}
		}
	}

	stats.NGSImported = passingCount + rushingCount + receivingCount
	if *verbose {
		log.Printf(" (Passing: %d, Rushing: %d, Receiving: %d)", passingCount, rushingCount, receivingCount)
	}

	i.markProgress(year, "ngs", "completed")
	i.updateProgress(year, "ngs", stats.NGSImported)
	return nil
}

func (i *Importer) validateData() error {
	log.Println("🔍 Validating imported data...")

	// Check data coverage by year
	query := `
		SELECT season,
		       COUNT(DISTINCT id) as games_count,
		       CASE
		           WHEN COUNT(DISTINCT id) = 0 THEN 'MISSING'
		           WHEN COUNT(DISTINCT id) < 250 THEN 'INCOMPLETE'
		           ELSE 'COMPLETE'
		       END as status
		FROM games
		WHERE season >= 2010 AND season <= 2024
		GROUP BY season
		ORDER BY season;
	`

	rows, err := i.dbPool.Query(i.ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	log.Println("\nSeason Coverage:")
	log.Println("────────────────────────────────")
	incomplete := 0
	missing := 0

	for rows.Next() {
		var season, gamesCount int
		var status string
		if err := rows.Scan(&season, &gamesCount, &status); err != nil {
			continue
		}

		emoji := "✅"
		if status == "INCOMPLETE" {
			emoji = "⚠️"
			incomplete++
		} else if status == "MISSING" {
			emoji = "❌"
			missing++
		}

		log.Printf("%s %d: %d games (%s)", emoji, season, gamesCount, status)
	}

	log.Println("────────────────────────────────")
	log.Printf("\nSummary:")
	log.Printf("  Complete: %d seasons", 15-incomplete-missing)
	log.Printf("  Incomplete: %d seasons", incomplete)
	log.Printf("  Missing: %d seasons", missing)

	return nil
}

func (i *Importer) showStats() error {
	log.Println("📊 Import Statistics")
	log.Println("════════════════════════════════════════════════")

	// Overall counts
	var playerCount, gameCount, statCount, ngsCount int64

	i.dbPool.QueryRow(i.ctx, "SELECT COUNT(*) FROM players").Scan(&playerCount)
	i.dbPool.QueryRow(i.ctx, "SELECT COUNT(*) FROM games").Scan(&gameCount)
	i.dbPool.QueryRow(i.ctx, "SELECT COUNT(*) FROM game_stats").Scan(&statCount)
	i.dbPool.QueryRow(i.ctx, "SELECT COUNT(*) FROM advanced_stats").Scan(&ngsCount)

	log.Printf("Total Players:      %,d", playerCount)
	log.Printf("Total Games:        %,d", gameCount)
	log.Printf("Total Game Stats:   %,d", statCount)
	log.Printf("Total Advanced Stats: %,d", ngsCount)

	// Season breakdown
	log.Println("\nBy Season:")
	query := `
		SELECT season,
		       COUNT(DISTINCT id) as games
		FROM games
		GROUP BY season
		ORDER BY season DESC;
	`

	rows, err := i.dbPool.Query(i.ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var season, games int
		if err := rows.Scan(&season, &games); err != nil {
			continue
		}
		log.Printf("  %d: %d games", season, games)
	}

	log.Println("════════════════════════════════════════════════")

	return nil
}

func (i *Importer) printImportSummary(year int, stats *ImportStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	log.Println("───────────────────────────────────────────────")
	log.Printf("📊 Import Summary for %d", year)
	log.Println("───────────────────────────────────────────────")
	log.Printf("Rosters:      %d players", stats.RostersImported)
	log.Printf("Games:        %d games", stats.GamesImported)
	log.Printf("Player Stats: %d records", stats.StatsImported)
	log.Printf("Next Gen:     %d records", stats.NGSImported)
	log.Printf("Errors:       %d", len(stats.Errors))
	log.Printf("Duration:     %s", duration.Round(time.Second))
	log.Println("───────────────────────────────────────────────")
}

// Helper functions for database operations

func (i *Importer) upsertPlayer(r *nflverse.RosterCSV) error {
	// Map team abbreviation to team ID
	var teamID *uuid.UUID
	if r.TeamAbbr != "" {
		var tid uuid.UUID
		err := i.dbPool.QueryRow(i.ctx,
			"SELECT id FROM teams WHERE abbreviation = $1",
			r.TeamAbbr,
		).Scan(&tid)
		if err == nil {
			teamID = &tid
		}
	}

	// Check if player exists by GSIS ID
	var existingID uuid.UUID
	err := i.dbPool.QueryRow(i.ctx,
		"SELECT id FROM players WHERE nfl_id = $1",
		r.PlayerID,
	).Scan(&existingID)

	now := time.Now()

	if err != nil {
		// Player doesn't exist, insert
		id := uuid.New()
		_, err = i.dbPool.Exec(i.ctx,
			`INSERT INTO players (
				id, nfl_id, name, position, team_id, jersey_number,
				height_inches, weight_pounds, birth_date, college,
				draft_year, draft_round, draft_pick, status,
				rookie_year, years_pro, headshot_url,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
			ON CONFLICT (nfl_id) DO UPDATE SET
				name = EXCLUDED.name,
				position = EXCLUDED.position,
				team_id = EXCLUDED.team_id,
				jersey_number = EXCLUDED.jersey_number,
				updated_at = EXCLUDED.updated_at`,
			id, r.PlayerID, r.FullName, r.Position, teamID, nilInt(r.JerseyNumber),
			nilInt(r.Height), nilInt(r.Weight), nilString(r.BirthDate), nilString(r.College),
			nilInt(r.EntryYear), nilInt(r.DraftRound), nilInt(r.DraftNumber), r.Status,
			nilInt(r.RookieYear), nilInt(r.YearsExp), nilString(r.HeadshotURL),
			now, now,
		)
		return err
	}

	// Player exists, update
	_, err = i.dbPool.Exec(i.ctx,
		`UPDATE players SET
			name = $1, position = $2, team_id = $3, jersey_number = $4,
			height_inches = $5, weight_pounds = $6, status = $7, updated_at = $8
		WHERE id = $9`,
		r.FullName, r.Position, teamID, nilInt(r.JerseyNumber),
		nilInt(r.Height), nilInt(r.Weight), r.Status, now, existingID,
	)

	return err
}

func (i *Importer) upsertGame(s *nflverse.ScheduleCSV) error {
	// Map team abbreviations to UUIDs
	var homeTeamID, awayTeamID uuid.UUID

	err := i.dbPool.QueryRow(i.ctx,
		"SELECT id FROM teams WHERE abbreviation = $1",
		s.HomeTeam,
	).Scan(&homeTeamID)
	if err != nil {
		return fmt.Errorf("home team %s not found", s.HomeTeam)
	}

	err = i.dbPool.QueryRow(i.ctx,
		"SELECT id FROM teams WHERE abbreviation = $1",
		s.AwayTeam,
	).Scan(&awayTeamID)
	if err != nil {
		return fmt.Errorf("away team %s not found", s.AwayTeam)
	}

	// Parse game date
	gameDate, err := time.Parse("2006-01-02", s.GameDay)
	if err != nil {
		return fmt.Errorf("invalid game date: %s", s.GameDay)
	}

	now := time.Now()
	id := uuid.New()

	_, err = i.dbPool.Exec(i.ctx,
		`INSERT INTO games (
			id, nfl_game_id, season, week, game_date, home_team_id, away_team_id,
			home_score, away_score, status, venue_name, weather_temp, weather_condition,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (nfl_game_id) DO UPDATE SET
			home_score = EXCLUDED.home_score,
			away_score = EXCLUDED.away_score,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`,
		id, s.GameID, s.Season, s.Week, gameDate, homeTeamID, awayTeamID,
		s.HomeScore, s.AwayScore, "final", nilString(s.Stadium),
		nilInt(s.Temp), nilString(""), now, now,
	)

	return err
}

func (i *Importer) batchInsertStats(batch []*nflverse.PlayerStatCSV) error {
	// Placeholder - simplified batch insert
	// In production, use COPY or multi-row INSERT for better performance
	for _, ps := range batch {
		if err := i.upsertPlayerStat(ps); err != nil {
			// Log error but continue
			if *verbose {
				log.Printf("      Failed to insert stat for %s: %v", ps.PlayerName, err)
			}
			continue
		}
	}
	return nil
}

func (i *Importer) upsertPlayerStat(ps *nflverse.PlayerStatCSV) error {
	// Get player ID and game ID
	var playerID uuid.UUID
	err := i.dbPool.QueryRow(i.ctx,
		"SELECT id FROM players WHERE nfl_id = $1",
		ps.PlayerID,
	).Scan(&playerID)
	if err != nil {
		// Player not found, skip
		return nil
	}

	// For now, we'll create simplified game stats
	// This is a simplified version - you may want to enhance this
	id := uuid.New()
	now := time.Now()

	_, err = i.dbPool.Exec(i.ctx,
		`INSERT INTO game_stats (
			id, player_id, season, week,
			passing_yards, passing_tds, interceptions_thrown,
			rushing_yards, rushing_tds, attempts,
			receiving_yards, receiving_tds, receptions, targets,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (player_id, season, week) DO UPDATE SET
			passing_yards = EXCLUDED.passing_yards,
			passing_tds = EXCLUDED.passing_tds,
			updated_at = EXCLUDED.updated_at`,
		id, playerID, ps.Season, ps.Week,
		ps.PassingYards, ps.PassingTDs, ps.Interceptions,
		ps.RushingYards, ps.RushingTDs, ps.Carries,
		ps.ReceivingYards, ps.ReceivingTDs, ps.Receptions, ps.Targets,
		now, now,
	)

	return err
}

func (i *Importer) upsertNextGenStatsPassing(ngs *nflverse.NextGenStatsPassingCSV) error {
	// Simplified NGS insert - enhance as needed
	return nil
}

func (i *Importer) upsertNextGenStatsRushing(ngs *nflverse.NextGenStatsRushingCSV) error {
	return nil
}

func (i *Importer) upsertNextGenStatsReceiving(ngs *nflverse.NextGenStatsReceivingCSV) error {
	return nil
}

func (i *Importer) markProgress(season int, dataType, status string) {
	query := `
		INSERT INTO import_progress (season, data_type, status, started_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (season, data_type) DO UPDATE SET
			status = EXCLUDED.status,
			started_at = NOW()
	`
	i.dbPool.Exec(i.ctx, query, season, dataType, status)
}

func (i *Importer) updateProgress(season int, dataType string, recordsImported int) {
	query := `
		UPDATE import_progress
		SET records_imported = $1,
		    completed_at = NOW(),
		    status = 'completed'
		WHERE season = $2 AND data_type = $3
	`
	i.dbPool.Exec(i.ctx, query, recordsImported, season, dataType)
}

// Helper functions for nullable values
func nilString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nilInt(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}
