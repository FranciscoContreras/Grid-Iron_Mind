.PHONY: help build build-importer build-diagnose sync-full sync-update sync-live sync-stats sync-injuries diagnose-players clean install-cron logs

# Default target
help:
	@echo "Grid Iron Mind - 2025 Season Sync & Historical Import"
	@echo ""
	@echo "2025 Season Sync Commands:"
	@echo "  make build         - Build the sync2025 binary"
	@echo "  make sync-full     - Run full initial sync (30-60 min)"
	@echo "  make sync-update   - Run regular update (2-5 min)"
	@echo "  make sync-live     - Run live game day sync"
	@echo "  make sync-stats    - Sync player stats only"
	@echo "  make sync-injuries - Sync injury reports only"
	@echo "  make install-cron  - Install automated cron schedule"
	@echo ""
	@echo "Diagnostic Commands:"
	@echo "  make diagnose-players    - Check for missing players (SQL, local)"
	@echo "  make diagnose-heroku     - Check for missing players (Heroku)"
	@echo "  make diagnose-players-go - Check for missing players (Go binary)"
	@echo ""
	@echo "Historical Import Commands (2010-2024):"
	@echo "  make build-importer       - Build historical import tool"
	@echo "  make import-year YEAR=2024 - Import single year"
	@echo "  make import-range START=2010 END=2014 - Import year range"
	@echo "  make import-full          - Import all 15 years (2010-2024)"
	@echo "  make import-validate      - Validate imported data"
	@echo "  make import-stats         - Show import statistics"
	@echo ""
	@echo "Other Commands:"
	@echo "  make logs          - View sync logs"
	@echo "  make clean         - Remove built binaries"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make build"
	@echo "  2. make sync-full"
	@echo "  3. make install-cron"

# Build the sync tool
build:
	@echo "Building sync2025..."
	@mkdir -p bin logs
	@go build -o bin/sync2025 cmd/sync2025/main.go
	@echo "✓ Build complete: bin/sync2025"

# Build diagnostic tool
build-diagnose:
	@echo "Building player diagnostic tool..."
	@mkdir -p bin logs
	@go build -o bin/diagnose-players cmd/diagnose-players/main.go
	@echo "✓ Build complete: bin/diagnose-players"

# Run player diagnostic (SQL-based - works without Go)
diagnose-players:
	@echo "Running player diagnostic (SQL)..."
	@psql $(DATABASE_URL) -f scripts/diagnose-missing-players.sql 2>&1 | tee logs/diagnose-players.log

# Run player diagnostic (Go binary - requires Go installed)
diagnose-players-go: build-diagnose
	@echo "Running player diagnostic (Go)..."
	@./bin/diagnose-players 2>&1 | tee -a logs/diagnose-players-go.log

# Run player diagnostic on Heroku production database
diagnose-heroku:
	@echo "Running player diagnostic on Heroku..."
	@bash scripts/heroku-diagnose.sh

# Run full initial sync
sync-full: build
	@echo "Running full 2025 season sync..."
	@echo "This will take 30-60 minutes"
	@./bin/sync2025 full 2>&1 | tee -a logs/sync-2025.log

# Run regular update
sync-update: build
	@echo "Running update sync..."
	@./bin/sync2025 update 2>&1 | tee -a logs/sync-2025.log

# Run live game day sync
sync-live: build
	@echo "Running live sync (Press Ctrl+C to stop)..."
	@./bin/sync2025 live 2>&1 | tee -a logs/sync-2025.log

# Sync player stats only
sync-stats: build
	@echo "Syncing player statistics..."
	@./bin/sync2025 stats 2>&1 | tee -a logs/sync-2025.log

# Sync injuries only
sync-injuries: build
	@echo "Syncing injury reports..."
	@./bin/sync2025 injuries 2>&1 | tee -a logs/sync-2025.log

# Install cron schedule
install-cron:
	@echo "Installing cron schedule for automated syncs..."
	@echo "Current crontab will be backed up to crontab.backup"
	@crontab -l > crontab.backup 2>/dev/null || true
	@crontab scripts/crontab-2025.txt
	@echo "✓ Cron schedule installed"
	@echo ""
	@echo "Scheduled jobs:"
	@crontab -l | grep -v "^#" | grep -v "^$$"

# View logs
logs:
	@tail -f logs/sync-2025.log

# Build historical import tool
build-importer:
	@echo "Building historical import tool..."
	@mkdir -p bin logs
	@go build -o bin/import-historical cmd/import_historical/main.go
	@echo "✓ Build complete: bin/import-historical"

# Import single year
import-year: build-importer
	@echo "Importing data for year $(YEAR)..."
	@./bin/import-historical --mode=year --year=$(YEAR) --verbose 2>&1 | tee -a logs/import-historical.log

# Import year range
import-range: build-importer
	@echo "Importing data for years $(START)-$(END)..."
	@./bin/import-historical --mode=range --start=$(START) --end=$(END) --verbose 2>&1 | tee -a logs/import-historical.log

# Import full 15 years
import-full: build-importer
	@echo "Importing 15 years of historical data (2010-2024)..."
	@echo "⚠️  This will take 60-90 minutes"
	@echo "Press Ctrl+C within 5 seconds to cancel..."
	@sleep 5
	@./bin/import-historical --mode=full --start=2010 --end=2024 --verbose 2>&1 | tee -a logs/import-historical.log

# Validate imported data
import-validate: build-importer
	@echo "Validating imported historical data..."
	@./bin/import-historical --mode=validate

# Show import statistics
import-stats: build-importer
	@echo "Showing import statistics..."
	@./bin/import-historical --mode=stats

# Clean built files
clean:
	@echo "Cleaning built files..."
	@rm -f bin/sync2025 bin/import-historical bin/diagnose-players
	@echo "✓ Clean complete"

# Development helpers
.PHONY: test db-status check-env

# Check database status
db-status:
	@echo "Checking database status..."
	@psql $(DATABASE_URL) -c "SELECT season, COUNT(*) as game_count FROM games GROUP BY season ORDER BY season DESC LIMIT 5;"
	@echo ""
	@psql $(DATABASE_URL) -c "SELECT COUNT(*) as player_count FROM players WHERE status = 'active';"
	@echo ""
	@psql $(DATABASE_URL) -c "SELECT season, COUNT(*) as stat_count FROM player_season_stats GROUP BY season ORDER BY season DESC LIMIT 5;"

# Check environment variables
check-env:
	@echo "Checking environment configuration..."
	@if [ -z "$(DATABASE_URL)" ]; then echo "❌ DATABASE_URL not set"; else echo "✓ DATABASE_URL configured"; fi
	@if [ -z "$(WEATHER_API_KEY)" ]; then echo "⚠️  WEATHER_API_KEY not set (optional)"; else echo "✓ WEATHER_API_KEY configured"; fi
	@if [ -z "$(REDIS_URL)" ]; then echo "⚠️  REDIS_URL not set (optional)"; else echo "✓ REDIS_URL configured"; fi

# Test connection to APIs
test-apis:
	@echo "Testing ESPN API connection..."
	@curl -s -o /dev/null -w "ESPN: %{http_code}\n" "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams"
	@echo "Testing NFLverse data..."
	@curl -s -o /dev/null -w "NFLverse: %{http_code}\n" "https://github.com/nflverse/nflverse-data/releases/latest/download/player_stats.csv"
