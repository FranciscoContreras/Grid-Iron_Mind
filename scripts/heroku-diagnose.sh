#!/bin/bash

# Grid Iron Mind - Heroku Player Diagnostic
# Runs the player diagnostic script on Heroku PostgreSQL database

APP_NAME="${1:-grid-iron-mind}"

echo "=== Running Player Diagnostic on Heroku ==="
echo "App: $APP_NAME"
echo ""

# Check if Heroku CLI is installed
if ! command -v heroku &> /dev/null; then
    echo "❌ Heroku CLI not found. Install from: https://devcenter.heroku.com/articles/heroku-cli"
    exit 1
fi

# Check if user is logged in
if ! heroku whoami &> /dev/null; then
    echo "❌ Not logged in to Heroku. Run: heroku login"
    exit 1
fi

echo "✓ Heroku CLI ready"
echo ""

# Run diagnostic SQL script
echo "Running diagnostic script..."
echo ""

heroku pg:psql --app "$APP_NAME" --file scripts/diagnose-missing-players.sql

echo ""
echo "=== Diagnostic Complete ==="
echo ""
echo "Next steps:"
echo "  1. If players are missing, sync rosters:"
echo "     curl -X POST https://nfl.wearemachina.com/api/v1/admin/sync/rosters \\"
echo "       -H 'X-API-Key: YOUR_API_KEY'"
echo ""
echo "  2. Or use the admin dashboard:"
echo "     https://nfl.wearemachina.com/"
echo ""
echo "  3. Check sync logs:"
echo "     heroku logs --tail --app $APP_NAME | grep sync"
