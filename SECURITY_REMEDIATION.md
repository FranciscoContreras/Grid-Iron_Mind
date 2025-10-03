# 🚨 Yahoo OAuth Security Remediation Guide

## Status: PARTIALLY COMPLETED

✅ **Completed:**
- Removed exposed credentials from current codebase
- Removed credentials from Heroku production environment
- Committed sanitized versions of documentation files

⚠️ **YOU MUST COMPLETE:**
1. Rotate Yahoo credentials (requires manual login)
2. Update Heroku with new credentials
3. Complete OAuth flow to get refresh token

---

## STEP 1: Rotate Yahoo Credentials (MANUAL - REQUIRED)

**⚠️ DO THIS IMMEDIATELY - The old credentials are PUBLIC!**

### Option A: Regenerate Client Secret (Recommended)

1. Go to https://developer.yahoo.com/apps/
2. Log in with your Yahoo account
3. Find your app: **Grid Iron Mind** (App ID: eYC8I6iq)
4. Click on the app
5. Look for "Regenerate Secret" or "Client Secret" section
6. Click **Regenerate** or **Reset**
7. Copy the new Client Secret immediately (you won't see it again!)
8. Your Client ID stays the same

### Option B: Delete and Recreate App (Alternative)

1. Go to https://developer.yahoo.com/apps/
2. Delete the existing "Grid Iron Mind" app
3. Create a new app with these settings:
   - **App Name:** Grid Iron Mind
   - **App Type:** Web Application
   - **Callback Domain:** `grid-iron-mind-71cc9734eaf4.herokuapp.com`
   - **API Permissions:** Fantasy Sports (Read)
4. Copy both the new Client ID and Client Secret

---

## STEP 2: Set New Credentials on Heroku

Once you have your new credentials from Step 1, run these commands:

```bash
# Set new Yahoo credentials
heroku config:set YAHOO_CLIENT_ID="paste_your_new_client_id_here"
heroku config:set YAHOO_CLIENT_SECRET="paste_your_new_client_secret_here"

# Verify they're set
heroku config:get YAHOO_CLIENT_ID
heroku config:get YAHOO_CLIENT_SECRET
```

---

## STEP 3: Complete OAuth Flow to Get Refresh Token

### Method 1: Using the Web OAuth Helper (Easiest)

1. **Start the OAuth helper on Heroku:**
   ```bash
   heroku ps:scale oauth=1
   ```

2. **Open the OAuth helper in your browser:**
   ```bash
   heroku open
   # Then navigate to: /yahoo/auth
   # Or visit directly: https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/auth
   ```

3. **Follow the OAuth flow:**
   - You'll be redirected to Yahoo to authorize
   - Log in with your Yahoo account
   - Authorize the app
   - You'll be redirected back with your tokens displayed

4. **Copy the refresh token** from the success page

5. **Set the refresh token on Heroku:**
   ```bash
   heroku config:set YAHOO_REFRESH_TOKEN="paste_refresh_token_here"
   ```

6. **Stop the OAuth helper** (no longer needed):
   ```bash
   heroku ps:scale oauth=0
   ```

### Method 2: Manual OAuth Flow (Alternative)

If the web helper doesn't work, you can do it manually:

1. **Build the authorization URL:**
   ```
   https://api.login.yahoo.com/oauth2/request_auth?client_id=YOUR_CLIENT_ID&redirect_uri=https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/callback&response_type=code&state=test123
   ```

2. **Visit the URL** in your browser, authorize the app

3. **Copy the `code` parameter** from the redirect URL

4. **Exchange the code for tokens** using curl:
   ```bash
   curl -X POST https://api.login.yahoo.com/oauth2/get_token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -H "Authorization: Basic $(echo -n 'YOUR_CLIENT_ID:YOUR_CLIENT_SECRET' | base64)" \
     -d "grant_type=authorization_code&code=YOUR_CODE&redirect_uri=https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/callback"
   ```

5. **Extract the `refresh_token`** from the JSON response

6. **Set it on Heroku:**
   ```bash
   heroku config:set YAHOO_REFRESH_TOKEN="paste_refresh_token_here"
   ```

---

## STEP 4: Verify Everything Works

```bash
# Check all Yahoo config vars are set
heroku config | grep YAHOO

# Restart the app
heroku restart

# Check logs
heroku logs --tail

# Test a Yahoo Fantasy API endpoint (once OAuth is complete)
curl "https://grid-iron-mind-71cc9734eaf4.herokuapp.com/api/v1/fantasy/rankings?position=QB&limit=10"
```

---

## STEP 5: Clean Up Git History (Optional but Recommended)

The credentials still exist in your Git history. To remove them permanently:

### Option A: Using git-filter-repo (Recommended)

```bash
# Install git-filter-repo
pip3 install git-filter-repo

# Create a backup first
git clone . ../gridironmind-backup

# Remove credentials from history
git filter-repo --replace-text <(cat << 'EOF'
dj0yJmk9Z1l5eGU0T0FTczI4JmQ9WVdrOVpWbERPRWsyYVhFbWNHbzlNQT09JnM9Y29uc3VtZXJzZWNyZXQmc3Y9MCZ4PTU5==>***REMOVED***
d5a7529ea91e162ead045e86ec1ddcd8f3dd8157==>***REMOVED***
EOF
)

# Force push to remotes
git push origin --force --all
git push heroku --force --all

# Clean up
rm -rf .git/refs/original/
git reflog expire --expire=now --all
git gc --prune=now --aggressive
```

### Option B: Using BFG Repo-Cleaner (Alternative)

```bash
# Download BFG
wget https://repo1.maven.org/maven2/com/madgag/bfg/1.14.0/bfg-1.14.0.jar

# Create replacements file
cat > replacements.txt << 'EOF'
dj0yJmk9Z1l5eGU0T0FTczI4JmQ9WVdrOVpWbERPRWsyYVhFbWNHbzlNQT09JnM9Y29uc3VtZXJzZWNyZXQmc3Y9MCZ4PTU5==>***REMOVED***
d5a7529ea91e162ead045e86ec1ddcd8f3dd8157==>***REMOVED***
EOF

# Run BFG
java -jar bfg-1.14.0.jar --replace-text replacements.txt .

# Clean up and force push
git reflog expire --expire=now --all && git gc --prune=now --aggressive
git push origin --force --all
git push heroku --force --all
```

---

## Timeline Summary

1. **IMMEDIATE** (Do right now):
   - ✅ Old credentials removed from Heroku (DONE)
   - ⏳ Rotate credentials on Yahoo Developer Portal (YOU MUST DO)

2. **Within 1 hour**:
   - Set new credentials on Heroku
   - Complete OAuth flow
   - Test Yahoo API integration

3. **Within 24 hours**:
   - Clean Git history (optional but recommended)
   - Monitor for any suspicious activity on your Yahoo app

---

## Security Best Practices Going Forward

1. ✅ Never commit credentials to Git
2. ✅ Always use environment variables
3. ✅ Use `.env` files locally (already in .gitignore)
4. ✅ Enable GitHub secret scanning alerts
5. ✅ Rotate credentials immediately if exposed
6. ✅ Use separate credentials for dev/staging/prod

---

## Questions?

If you run into issues:
1. Check Heroku logs: `heroku logs --tail`
2. Verify config vars: `heroku config`
3. Test OAuth helper: `heroku open`

Once you've completed Steps 1-3, the Yahoo Fantasy integration will be fully functional with secure credentials!
