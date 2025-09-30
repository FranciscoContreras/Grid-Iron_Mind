# NFLverse Data Loading Scripts

These scripts fetch player statistics from [nflverse](https://nflverse.com) and load them into your PostgreSQL database.

## Why Use These Scripts?

NFLverse provides comprehensive NFL statistics as CSV/Parquet files on GitHub. Instead of implementing complex CSV parsing in Go, these scripts handle the ETL (Extract, Transform, Load) process externally:

1. **Download** CSV data from nflverse GitHub releases
2. **Transform** weekly stats into season aggregates
3. **Load** into PostgreSQL `player_career_stats` table

This approach keeps your Go API focused on serving data, not data ingestion.

## Available Scripts

### Python Version (Recommended)

**File:** `load_nflverse_data.py`

**Prerequisites:**
```bash
pip install pandas psycopg2-binary requests
```

**Usage:**
```bash
# Set DATABASE_URL environment variable
export DATABASE_URL="postgres://user:pass@host:5432/dbname"

# Run the script
python scripts/load_nflverse_data.py
```

**What it does:**
- Downloads player stats CSVs from nflverse for 2023-2024 seasons
- Aggregates weekly stats into season totals
- Matches players by name to your database
- Upserts into `player_career_stats` table
- Shows progress and error reporting

### R Version (Alternative)

**File:** `load_nflverse_data.R`

**Prerequisites:**
```R
install.packages(c("nflreadr", "DBI", "RPostgres", "dplyr", "tidyr"))
```

**Usage:**
```bash
# Set DATABASE_URL environment variable
export DATABASE_URL="postgres://user:pass@host:5432/dbname"

# Run the script
Rscript scripts/load_nflverse_data.R
```

**Advantages:**
- Uses official `nflreadr` R package
- Native nflverse tooling
- Handles more complex data transformations

## Loading Data to Heroku

To populate your Heroku production database:

```bash
# Get Heroku database URL
heroku config:get DATABASE_URL -a grid-iron-mind

# Set it locally
export DATABASE_URL="<paste_heroku_url_here>"

# Run the Python script
python scripts/load_nflverse_data.py
```

## What Gets Loaded

The scripts populate the `player_career_stats` table with:

- **Games played** - Number of games in season
- **Passing stats** - Yards, TDs, INTs, attempts, completions
- **Rushing stats** - Yards, TDs, attempts
- **Receiving stats** - Yards, TDs, receptions, targets

## Player Matching

Players are matched by name (case-insensitive) between:
- NFLverse `player_display_name` field
- Your database `players.name` field

Common name variations (Jr., Sr., III) are handled automatically.

## Customization

### Load Different Seasons

**Python:**
```python
# Edit line 172 in load_nflverse_data.py
seasons = [2020, 2021, 2022, 2023, 2024]
```

**R:**
```R
# Edit line 148 in load_nflverse_data.R
stats <- load_player_stats(seasons = 2020:2024)
```

### Filter by Position

**Python:**
```python
# After line 52, add:
df = df[df['position'].isin(['QB', 'RB', 'WR', 'TE'])]
```

## Troubleshooting

### Player Not Found Errors

If many players aren't matching:
- Ensure you've run `/api/v1/admin/sync/rosters` to populate players
- Check player names match ESPN format
- Consider fuzzy matching for edge cases

### Database Connection Issues

Verify your DATABASE_URL format:
```
postgres://username:password@hostname:port/database
```

For Heroku, always use SSL: `?sslmode=require` (handled automatically by scripts)

### CSV Download Failures

NFLverse data is hosted on GitHub releases. If downloads fail:
- Check internet connectivity
- Verify the season year exists (2020+)
- Check GitHub isn't rate-limiting

## Integration with API

After loading data:

1. **Test the API:**
   ```bash
   curl "https://nfl.wearemachina.com/api/v1/players/{player_id}/career"
   ```

2. **Clear cache:**
   Career stats are cached for 30 minutes. To force refresh, either:
   - Wait 30 minutes
   - Restart the Heroku dyno: `heroku restart -a grid-iron-mind`

3. **Verify in dashboard:**
   - Go to Players tab
   - Click any player
   - Career stats should now display

## Scheduled Updates

To keep stats current, run these scripts:
- **Weekly during season** - After each week's games complete
- **Off-season** - Once after season ends for final stats

Consider setting up a cron job or scheduled task.

## Future Enhancements

Possible additions:
- Next Gen Stats (advanced metrics)
- Play-by-play data
- Injury reports
- Depth charts
- Automated scheduling

For now, focus on career stats as the foundation.
