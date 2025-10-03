# 🔄 Yahoo Credentials Rotation - Visual Guide

## Quick Links
- **Yahoo Developer Portal**: https://developer.yahoo.com/apps/
- **Yahoo Login**: https://login.yahoo.com/

---

## Step-by-Step Screenshots Guide

### 1️⃣ Go to Yahoo Developer Portal

**URL**: https://developer.yahoo.com/apps/

**What you'll see**:
- A login page (if not logged in)
- OR a list of your apps (if already logged in)

**Action**: Log in with your Yahoo account

---

### 2️⃣ Find Your App

**What to look for**:
```
My Apps
├── Grid Iron Mind (App ID: eYC8I6iq) ← THIS ONE
└── (other apps if you have any)
```

**Action**: Click on "Grid Iron Mind"

---

### 3️⃣ View App Details

**You should see**:
```
Grid Iron Mind
App ID: eYC8I6iq

Client ID (Consumer Key): dj0yJmk9... (old one, needs rotation)
Client Secret (Consumer Secret): ******** (hidden)

API Permissions:
☑ Fantasy Sports (Read)
```

---

### 4️⃣ Regenerate Credentials

**Look for one of these buttons**:
- [ ] "Regenerate Secret"
- [ ] "Reset Client Secret"
- [ ] "Generate New Secret"
- [ ] "Delete App" (if no regenerate option)

---

## Method A: Regenerate Secret (If Available)

**Steps**:
1. Click **"Regenerate Secret"** button
2. Confirm the action
3. **NEW CLIENT SECRET APPEARS** - Copy it immediately!
4. Client ID stays the same

**Copy this**:
```
Client ID (unchanged): [copy the existing one]
Client Secret (NEW): [copy the new one that just appeared]
```

---

## Method B: Delete and Recreate App

**If no "Regenerate" button exists**:

### Delete Old App:
1. Scroll down to find "Delete App" button
2. Click it
3. Confirm deletion
4. App is now deleted

### Create New App:
1. Click **"Create an App"** button (usually top right)
2. Fill in the form:

   ```
   Application Name: Grid Iron Mind
   Description: NFL Fantasy Sports Data API Integration

   Application Type:
   ⚫ Web Application
   ⚪ Desktop Application
   ⚪ Mobile Application

   Home Page URL: https://nfl.wearemachina.com

   Redirect URI(s):
   https://grid-iron-mind-71cc9734eaf4.herokuapp.com/yahoo/callback

   Callback Domain:
   grid-iron-mind-71cc9734eaf4.herokuapp.com

   API Permissions:
   ☑ Fantasy Sports
      └── ☑ Read
      └── ☐ Write
   ```

3. Click **"Create App"**
4. **SUCCESS PAGE** - You'll see:
   ```
   App Created Successfully!

   App ID: [new app ID]
   Client ID (Consumer Key): [NEW - copy this]
   Client Secret (Consumer Secret): [NEW - copy this]

   ⚠️ Warning: Save your Client Secret now.
       You won't be able to see it again!
   ```

5. **COPY BOTH** immediately!

---

## What to Do Next

Once you have your new credentials:

### Option 1: Use the automated script
```bash
./complete-yahoo-oauth.sh
```
The script will ask you to paste the new credentials.

### Option 2: Set manually
```bash
heroku config:set YAHOO_CLIENT_ID="paste_new_client_id_here"
heroku config:set YAHOO_CLIENT_SECRET="paste_new_client_secret_here"
```

---

## Troubleshooting

### "I don't see my app in the list"

**Solution**:
- Make sure you're logged in with the correct Yahoo account
- Check if you created the app with a different Yahoo account
- The app might have been deleted - you'll need to create a new one

### "I don't see a Regenerate button"

**Solution**:
- Yahoo may not allow secret regeneration for this app type
- Use Method B: Delete and recreate the app

### "I forgot to copy the Client Secret"

**Solution**:
- You'll need to regenerate it again
- OR delete and recreate the app
- There's no way to view the secret after closing the page

### "The create app form looks different"

**Solution**:
- Yahoo updates their UI sometimes
- Look for similar fields
- The key fields you need:
  - App name
  - Callback/Redirect URL
  - API permissions (Fantasy Sports)

---

## Alternative: Use Yahoo's New App Registration (If Available)

Some accounts may see a newer interface:

1. Go to: https://developer.yahoo.com/oauth2/guide/openid_connect/getting_started.html
2. Click "Create an app"
3. Choose "Private app" or "Public app"
4. Follow similar steps as above

---

## Need More Help?

**Yahoo Developer Documentation**:
- https://developer.yahoo.com/oauth2/guide/
- https://developer.yahoo.com/fantasysports/

**Still stuck?**
1. Take a screenshot of what you see
2. Check Yahoo's developer forums
3. Or just delete and recreate - it's the safest option

---

## Security Reminder

⚠️ **After you rotate**:
- Old credentials are now invalid
- Anyone who found the old credentials can't use them
- New credentials are secure (as long as you keep them secret!)

🔒 **Best practices**:
- Never commit credentials to Git
- Always use environment variables
- Rotate regularly (every 6-12 months)
