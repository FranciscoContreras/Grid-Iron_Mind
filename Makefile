.PHONY: help build sync-full sync-update sync-live sync-stats sync-injuries clean install-cron logs

# Default target
help:
	@echo "Grid Iron Mind - 2025 Season Sync"
	@echo ""
	@echo "Available commands:"
	@echo "  make build         - Build the sync2025 binary"
	@echo "  make sync-full     - Run full initial sync (30-60 min)"
	@echo "  make sync-update   - Run regular update (2-5 min)"
	@echo "  make sync-live     - Run live game day sync"
	@echo "  make sync-stats    - Sync player stats only"
	@echo "  make sync-injuries - Sync injury reports only"
	@echo "  make install-cron  - Install automated cron schedule"
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

# Clean built files
clean:
	@echo "Cleaning built files..."
	@rm -f bin/sync2025
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
