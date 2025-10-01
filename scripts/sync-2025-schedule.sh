#!/bin/bash

# sync-2025-schedule.sh
# Automated script to keep 2025 NFL season data up-to-date
# Run this script via cron for automated updates

set -e

# Change to project directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"
cd "$PROJECT_DIR"

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Determine current day of week
DAY_OF_WEEK=$(date +%u) # 1=Monday, 7=Sunday
HOUR=$(date +%H)

echo "========================================="
echo "Grid Iron Mind - 2025 Season Sync"
echo "Time: $(date)"
echo "========================================="

# Build the sync tool if not already built
if [ ! -f "./bin/sync2025" ]; then
    echo "Building sync2025 tool..."
    go build -o ./bin/sync2025 ./cmd/sync2025
fi

# Sync strategy:
# - Monday morning: Full roster update (trades, signings)
# - Tuesday-Saturday: Daily stats update
# - Sunday-Monday (game days): Frequent updates every 5 minutes during games
# - Daily: Injury report updates

if [ "$DAY_OF_WEEK" -eq 1 ] && [ "$HOUR" -lt 12 ]; then
    # Monday morning - full roster refresh
    echo "Running FULL roster sync (Monday morning)..."
    ./bin/sync2025 update

elif [ "$DAY_OF_WEEK" -eq 7 ] || ([ "$DAY_OF_WEEK" -eq 1 ] && [ "$HOUR" -ge 12 ]); then
    # Sunday or Monday afternoon (game days) - frequent updates
    if [ "$HOUR" -ge 13 ] && [ "$HOUR" -le 23 ]; then
        echo "Running LIVE sync (game day)..."
        # Live mode will run continuously for 5 minutes
        timeout 5m ./bin/sync2025 live || true
    else
        echo "Running UPDATE sync (non-game hours on game day)..."
        ./bin/sync2025 update
    fi

else
    # Tuesday-Saturday - regular daily update
    echo "Running UPDATE sync (regular day)..."
    ./bin/sync2025 update
fi

# Always update injuries (quick operation)
echo ""
echo "Updating injury reports..."
./bin/sync2025 injuries

echo ""
echo "========================================="
echo "Sync completed successfully"
echo "========================================="
