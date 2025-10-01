#!/bin/bash

# verify-setup.sh
# Verify that everything is ready for 2025 season sync

set -e

echo "========================================="
echo "Grid Iron Mind - Setup Verification"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SUCCESS=0
WARNINGS=0
ERRORS=0

# Check Go installation
echo -n "Checking Go installation... "
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} $GO_VERSION"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "${RED}✗${NC} Go not installed"
    echo "  Install from: https://golang.org/doc/install"
    ERRORS=$((ERRORS + 1))
fi

# Check .env file
echo -n "Checking .env file... "
if [ -f .env ]; then
    echo -e "${GREEN}✓${NC} Found"
    SUCCESS=$((SUCCESS + 1))

    # Check required variables
    source .env

    echo -n "  DATABASE_URL... "
    if [ -z "$DATABASE_URL" ]; then
        echo -e "${RED}✗${NC} Not set"
        ERRORS=$((ERRORS + 1))
    else
        echo -e "${GREEN}✓${NC}"
        SUCCESS=$((SUCCESS + 1))
    fi

    echo -n "  WEATHER_API_KEY... "
    if [ -z "$WEATHER_API_KEY" ]; then
        echo -e "${YELLOW}⚠${NC} Not set (optional)"
        WARNINGS=$((WARNINGS + 1))
    else
        echo -e "${GREEN}✓${NC}"
        SUCCESS=$((SUCCESS + 1))
    fi
else
    echo -e "${RED}✗${NC} Not found"
    echo "  Copy .env.example to .env and configure"
    ERRORS=$((ERRORS + 1))
fi

# Check database connection
if [ ! -z "$DATABASE_URL" ]; then
    echo -n "Checking database connection... "
    if psql "$DATABASE_URL" -c "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✓${NC} Connected"
        SUCCESS=$((SUCCESS + 1))

        # Check if tables exist
        echo -n "  Checking required tables... "
        TABLES=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('teams', 'players', 'games', 'game_team_stats', 'player_season_stats')")
        if [ "$TABLES" -eq 5 ]; then
            echo -e "${GREEN}✓${NC} All tables exist"
            SUCCESS=$((SUCCESS + 1))
        else
            echo -e "${RED}✗${NC} Missing tables ($TABLES/5 found)"
            echo "  Run migrations: psql \$DATABASE_URL -f migrations/003_enhance_comprehensive_schema.sql"
            ERRORS=$((ERRORS + 1))
        fi
    else
        echo -e "${RED}✗${NC} Cannot connect"
        echo "  Check DATABASE_URL in .env"
        ERRORS=$((ERRORS + 1))
    fi
fi

# Check directories
echo -n "Checking directories... "
if [ -d "bin" ] && [ -d "logs" ]; then
    echo -e "${GREEN}✓${NC} bin/ and logs/ exist"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "${YELLOW}⚠${NC} Creating directories"
    mkdir -p bin logs
    WARNINGS=$((WARNINGS + 1))
fi

# Check if sync tool is built
echo -n "Checking sync2025 binary... "
if [ -f "bin/sync2025" ]; then
    echo -e "${GREEN}✓${NC} Built"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "${YELLOW}⚠${NC} Not built"
    echo "  Run: make build"
    WARNINGS=$((WARNINGS + 1))
fi

# Check scripts are executable
echo -n "Checking script permissions... "
if [ -x "scripts/sync-2025-schedule.sh" ]; then
    echo -e "${GREEN}✓${NC} Executable"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "${YELLOW}⚠${NC} Fixing permissions"
    chmod +x scripts/*.sh
    WARNINGS=$((WARNINGS + 1))
fi

# Check ESPN API
echo -n "Testing ESPN API... "
if curl -s -f "https://site.api.espn.com/apis/site/v2/sports/football/nfl/teams" > /dev/null; then
    echo -e "${GREEN}✓${NC} Accessible"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "${RED}✗${NC} Cannot reach ESPN API"
    echo "  Check internet connection"
    ERRORS=$((ERRORS + 1))
fi

# Check NFLverse data
echo -n "Testing NFLverse data... "
if curl -s -f -I "https://github.com/nflverse/nflverse-data" > /dev/null; then
    echo -e "${GREEN}✓${NC} Accessible"
    SUCCESS=$((SUCCESS + 1))
else
    echo -e "${YELLOW}⚠${NC} Cannot reach NFLverse"
    WARNINGS=$((WARNINGS + 1))
fi

# Check PostgreSQL version
if [ ! -z "$DATABASE_URL" ] && psql "$DATABASE_URL" -c "SELECT 1" &> /dev/null; then
    echo -n "Checking PostgreSQL version... "
    PG_VERSION=$(psql "$DATABASE_URL" -t -c "SELECT version()" | head -1 | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo -e "${GREEN}✓${NC} PostgreSQL $PG_VERSION"
    SUCCESS=$((SUCCESS + 1))
fi

echo ""
echo "========================================="
echo "Verification Complete"
echo "========================================="
echo -e "${GREEN}✓ Passed: $SUCCESS${NC}"
if [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠ Warnings: $WARNINGS${NC}"
fi
if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}✗ Errors: $ERRORS${NC}"
fi
echo ""

if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✓ Ready to sync!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Build tool:  make build"
    echo "  2. Full sync:   make sync-full"
    echo "  3. Setup cron:  make install-cron"
    exit 0
else
    echo -e "${RED}✗ Please fix errors before syncing${NC}"
    exit 1
fi
