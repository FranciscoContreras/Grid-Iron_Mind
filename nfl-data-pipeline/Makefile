.PHONY: help build test clean sync-full sync-year sync-update sync-live sync-stats sync-injuries install-cron logs db-status

help:
	@echo "NFL Data Pipeline - Rust Edition"
	@echo ""
	@echo "Available commands:"
	@echo "  make build          - Build release binary"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Data Sync Commands:"
	@echo "  make sync-full      - Full import (2010-2025) - TAKES 30-60 MIN"
	@echo "  make sync-year      - Import specific year (set YEAR=2024)"
	@echo "  make sync-update    - Update current week only"
	@echo "  make sync-live      - Live game day updates (continuous)"
	@echo "  make sync-stats     - Player statistics only"
	@echo "  make sync-injuries  - Injury reports only"
	@echo ""
	@echo "Automation:"
	@echo "  make install-cron   - Install automated cron schedule"
	@echo "  make logs           - View sync logs"
	@echo "  make db-status      - Check database status"

# Build release binary
build:
	@echo "🔨 Building release binary..."
	@cargo build --release
	@echo "✅ Binary: target/release/nfl-data-pipeline"

# Run tests
test:
	@echo "🧪 Running tests..."
	@cargo test

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@cargo clean

# Full import (2010-2025)
sync-full: build
	@echo "⚠️  WARNING: Full import takes 30-60 minutes"
	@echo "📦 Importing 16 seasons of data (2010-2025)..."
	@./target/release/nfl-data-pipeline --mode full --start-year 2010 --end-year 2025
	@echo "✅ Full import complete!"

# Import specific year
sync-year: build
	@echo "📅 Importing year $(YEAR)..."
	@./target/release/nfl-data-pipeline --mode year --year $(YEAR)

# Update current week
sync-update: build
	@echo "🔄 Updating current week..."
	@./target/release/nfl-data-pipeline --mode update

# Live game day sync (continuous)
sync-live: build
	@echo "🔴 LIVE: Starting continuous sync..."
	@./target/release/nfl-data-pipeline --mode live

# Player statistics only
sync-stats: build
	@echo "📊 Syncing player statistics..."
	@./target/release/nfl-data-pipeline --mode stats

# Injury reports only
sync-injuries: build
	@echo "🏥 Syncing injury reports..."
	@./target/release/nfl-data-pipeline --mode injuries

# Install cron schedule
install-cron:
	@echo "⏰ Installing cron schedule..."
	@cat scripts/crontab-2025.txt
	@echo ""
	@echo "Add the above to your crontab with: crontab -e"

# View logs
logs:
	@tail -f sync.log

# Database status
db-status:
	@echo "🗄️  Database Status:"
	@psql $$DATABASE_URL -c "SELECT COUNT(*) as players FROM players;"
	@psql $$DATABASE_URL -c "SELECT COUNT(*) as games FROM games;"
	@psql $$DATABASE_URL -c "SELECT COUNT(*) as stats FROM game_stats;"
	@psql $$DATABASE_URL -c "SELECT season, data_type, status, records_imported FROM import_progress ORDER BY season DESC, data_type;"

# Quick check
check:
	@cargo check
