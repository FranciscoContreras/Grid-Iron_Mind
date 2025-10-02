# Deploying Rust Pipeline to Heroku

## Architecture

The app uses **multi-buildpack** support to run both Go and Rust:
- **Go buildpack**: Compiles API server (`cmd/server`)
- **Rust buildpack**: Compiles data pipeline (`nfl-data-pipeline`)
- **Shared DATABASE_URL**: Both connect to same PostgreSQL database

## Deployment Steps

### 1. Configure Multi-Buildpack

The app uses `.buildpacks` file to specify build order:
```
heroku/rust (first - builds nfl-data-pipeline)
heroku/go (second - builds API server)
```

Heroku will automatically detect and use both buildpacks.

### 2. Deploy to Heroku

```bash
# Commit all changes
git add .
git commit -m "Add Rust data pipeline to Heroku deployment"

# Push to Heroku (builds both Go and Rust)
git push heroku main
```

**Build process:**
1. Rust buildpack compiles `nfl-data-pipeline/target/release/nfl-data-pipeline`
2. Go buildpack compiles `bin/server`
3. Both binaries available in dyno

### 3. Run Data Import on Heroku

**One-time full import (2010-2025):**
```bash
# Run full import (30-60 min)
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode full --start-year 2010 --end-year 2025"
```

**Single year import:**
```bash
# Import specific year
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode year --year 2024"
```

**Daily update (current week):**
```bash
# Update current season
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode update"
```

### 4. Set Up Automated Scheduling

Use Heroku Scheduler addon for automatic imports:

```bash
# Add scheduler addon (free)
heroku addons:create scheduler:standard

# Open scheduler dashboard
heroku addons:open scheduler
```

**Recommended schedule:**

| Time | Frequency | Command |
|------|-----------|---------|
| Daily 3 AM | Every day | `nfl-data-pipeline/target/release/nfl-data-pipeline --mode update` |
| Sunday 1 PM | Weekly (game day) | `nfl-data-pipeline/target/release/nfl-data-pipeline --mode live` |

### 5. Monitor Imports

**View logs:**
```bash
heroku logs --tail --source app
```

**Check database after import:**
```bash
heroku run psql --command "SELECT COUNT(*) FROM players;"
heroku run psql --command "SELECT COUNT(*) FROM game_stats;"
heroku run psql --command "SELECT season, COUNT(*) FROM game_stats GROUP BY season ORDER BY season;"
```

## Performance Benefits

**Running on Heroku vs Local:**
- ✅ **No network latency** - Same datacenter as database
- ✅ **Faster execution** - Heroku dynos have good CPU
- ✅ **Automated** - Scheduler runs without intervention
- ✅ **No local machine** - Runs in cloud 24/7
- ✅ **Better error handling** - Heroku logs and monitoring

**Speed comparison:**
- Local → Heroku DB: ~10 min per season (network overhead)
- Heroku → Heroku DB: ~2-3 min per season (co-located)

**Cost:**
- Free tier: Sufficient for imports
- Scheduler: Free addon
- Database: Existing plan (no additional cost)

## Troubleshooting

### Build fails with "Cargo.toml not found"

Check `RustConfig` file points to correct path:
```
RUST_CARGO_BUILD_FLAGS="--release --manifest-path=nfl-data-pipeline/Cargo.toml"
```

### Binary not found at runtime

Verify Procfile path:
```
worker: nfl-data-pipeline/target/release/nfl-data-pipeline --mode update
```

### Database connection fails

Rust pipeline automatically uses $DATABASE_URL from environment. No configuration needed.

### TLS/SSL errors

The pipeline includes native-TLS support for Heroku PostgreSQL. If you see SSL errors, ensure `postgres-native-tls` is in Cargo.toml dependencies.

## Architecture Diagram

```
┌─────────────────────────────────────────┐
│         Heroku App (gridironmind)       │
├─────────────────────────────────────────┤
│                                         │
│  Web Dyno                               │
│  ├─ Go API Server (bin/server)         │
│  └─ Serves HTTP on $PORT               │
│                                         │
│  Worker/One-off Dyno (on demand)       │
│  ├─ Rust Pipeline (nfl-data-pipeline)  │
│  └─ Imports from nflfastr              │
│                                         │
├─────────────────────────────────────────┤
│      Heroku PostgreSQL Database         │
│      (shared by both processes)         │
└─────────────────────────────────────────┘
         ▲
         │
    Downloads from
         │
         ▼
┌─────────────────────────────────────────┐
│   GitHub: nflverse-data releases        │
│   ├─ player_stats_{year}.csv           │
│   ├─ roster_{year}.csv                  │
│   └─ nextgen_stats/ngs_{year}_*.csv    │
└─────────────────────────────────────────┘
```

## Commands Cheat Sheet

```bash
# Deploy
git push heroku main

# Full import (run once)
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode full --start-year 2010 --end-year 2025"

# Daily update (schedule this)
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode update"

# Import single year
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode year --year 2024"

# Validate data
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode validate"

# Check database size
heroku pg:info

# View logs
heroku logs --tail

# Check scheduler jobs
heroku addons:open scheduler
```
