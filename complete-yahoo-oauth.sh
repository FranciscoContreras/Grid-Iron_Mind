#!/bin/bash

# Yahoo OAuth Setup - Quick Start Script
# This script will guide you through completing the Yahoo OAuth setup

set -e

echo "🏈 Yahoo Fantasy Sports OAuth Setup"
echo "===================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Check if credentials are rotated
echo -e "${YELLOW}STEP 1: Rotate Yahoo Credentials${NC}"
echo "⚠️  The old credentials are PUBLIC and MUST be rotated!"
echo ""
echo "Go to: https://developer.yahoo.com/apps/"
echo "  1. Log in with your Yahoo account"
echo "  2. Find app: Grid Iron Mind (App ID: eYC8I6iq)"
echo "  3. Regenerate the Client Secret OR delete & recreate the app"
echo ""
read -p "Have you rotated the credentials? (yes/no): " rotated

if [ "$rotated" != "yes" ]; then
    echo -e "${RED}❌ Please rotate the credentials first!${NC}"
    echo "Visit: https://developer.yahoo.com/apps/"
    exit 1
fi

echo -e "${GREEN}✓ Credentials rotated${NC}"
echo ""

# Step 2: Get new credentials
echo -e "${YELLOW}STEP 2: Enter New Credentials${NC}"
echo ""
read -p "Enter your new Yahoo Client ID: " client_id
read -p "Enter your new Yahoo Client Secret: " client_secret

if [ -z "$client_id" ] || [ -z "$client_secret" ]; then
    echo -e "${RED}❌ Both Client ID and Secret are required!${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Setting credentials on Heroku...${NC}"
heroku config:set YAHOO_CLIENT_ID="$client_id"
heroku config:set YAHOO_CLIENT_SECRET="$client_secret"

echo -e "${GREEN}✓ Credentials set on Heroku${NC}"
echo ""

# Step 3: Verify credentials
echo -e "${YELLOW}STEP 3: Verify Credentials${NC}"
echo ""
echo "Client ID set to: $(heroku config:get YAHOO_CLIENT_ID)"
echo "Client Secret set to: $(heroku config:get YAHOO_CLIENT_SECRET | head -c 20)..."
echo ""

# Step 4: Start OAuth flow
echo -e "${YELLOW}STEP 4: Complete OAuth Flow${NC}"
echo ""
echo "Starting OAuth helper on Heroku..."
heroku ps:scale oauth=1

sleep 3

echo ""
echo -e "${GREEN}OAuth helper is running!${NC}"
echo ""
echo "🌐 Opening OAuth flow in your browser..."
echo "URL: https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/auth"
echo ""

# Try to open in browser
if command -v open &> /dev/null; then
    open "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/auth"
elif command -v xdg-open &> /dev/null; then
    xdg-open "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/auth"
else
    echo "Please open this URL manually:"
    echo "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/auth"
fi

echo ""
echo "Follow these steps in your browser:"
echo "  1. You'll be redirected to Yahoo"
echo "  2. Log in with your Yahoo account"
echo "  3. Authorize the Grid Iron Mind app"
echo "  4. Copy the REFRESH TOKEN from the success page"
echo ""
read -p "Enter the refresh token you received: " refresh_token

if [ -z "$refresh_token" ]; then
    echo -e "${RED}❌ Refresh token is required!${NC}"
    echo "Please run the OAuth flow again."
    exit 1
fi

echo ""
echo -e "${BLUE}Setting refresh token on Heroku...${NC}"
heroku config:set YAHOO_REFRESH_TOKEN="$refresh_token"

echo -e "${GREEN}✓ Refresh token set${NC}"
echo ""

# Step 5: Stop OAuth helper
echo -e "${YELLOW}STEP 5: Cleanup${NC}"
echo ""
echo "Stopping OAuth helper (no longer needed)..."
heroku ps:scale oauth=0

echo -e "${GREEN}✓ OAuth helper stopped${NC}"
echo ""

# Step 6: Restart app
echo "Restarting application..."
heroku restart

echo -e "${GREEN}✓ Application restarted${NC}"
echo ""

# Step 7: Verify
echo -e "${YELLOW}STEP 6: Verification${NC}"
echo ""
echo "Checking configuration..."
heroku config | grep YAHOO

echo ""
echo -e "${GREEN}🎉 Yahoo OAuth setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Wait 30 seconds for app to restart"
echo "  2. Test the Yahoo Fantasy API:"
echo "     curl 'https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/fantasy/rankings?position=QB&limit=5'"
echo ""
echo "  3. Check logs for any errors:"
echo "     heroku logs --tail"
echo ""
echo "  4. (Optional) Clean Git history to remove old credentials:"
echo "     See SECURITY_REMEDIATION.md Step 5"
echo ""
