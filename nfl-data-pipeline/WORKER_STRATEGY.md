# Always-On Worker Strategy

## Overview

The Rust data pipeline runs as an **always-on worker dyno** that intelligently adapts its behavior based on day and time, providing maximum live data freshness during game days while being efficient during off-hours.

## Smart Update Strategy

### Game Day Mode (Live Updates)
**When:** Sunday 1PM-11PM, Monday 8PM-11PM, Thursday 8PM-11PM
- Runs `--mode live` for continuous real-time updates
- Updates every **5 minutes**
- Captures live stats, scores, injuries as games progress
- Optimized for minimal latency

### Off-Hours Mode (Regular Updates)
**When:** All other times
- Runs `--mode update` for current season refresh
- Updates every **30 minutes**
- Efficient resource usage
- Keeps data current without overwhelming the database

## Worker Script (`run-worker.sh`)

The worker script automatically:
1. Detects current day of week and hour
2. Chooses appropriate mode (live vs update)
3. Runs the Rust pipeline
4. Sleeps for appropriate interval
5. Loops continuously

### Logic Flow

```bash
while true; do
    if GAME_DAY and GAME_HOURS:
        run --mode live
        sleep 5 minutes
    else:
        run --mode update
        sleep 30 minutes
    fi
done
```

## Deployment

### 1. Deploy to Heroku
```bash
git push heroku main
```

### 2. Scale Up Worker Dyno
```bash
# Start the always-on worker
heroku ps:scale worker=1
```

### 3. Monitor Worker
```bash
# View worker logs
heroku logs --tail --dyno worker

# Check worker status
heroku ps
```

### 4. Stop Worker (if needed)
```bash
# Scale down to stop updates
heroku ps:scale worker=0
```

## Resource Usage

### Eco Dyno Hours (1000 hours/month)
- **Web dyno**: ~730 hours/month (always on)
- **Worker dyno**: ~270 hours/month (always on)
- **Total**: ~1000 hours/month ✅ Perfect fit!

### Database Impact
- Live updates: ~12 queries/min during game hours
- Regular updates: ~2 queries/min off hours
- Minimal load, well within Heroku PostgreSQL limits

### Network Usage
- Downloads ~2-5 MB per update cycle
- ~500-1000 MB/month total
- Negligible for most Heroku plans

## Update Frequency

| Day/Time | Mode | Frequency | Purpose |
|----------|------|-----------|---------|
| Sunday 1-11 PM | live | 5 min | Live game updates |
| Monday 8-11 PM | live | 5 min | Monday Night Football |
| Thursday 8-11 PM | live | 5 min | Thursday Night Football |
| All other times | update | 30 min | Keep current season fresh |

## Data Freshness Guarantees

- **Live games**: 5-minute freshness
- **Completed games**: Updated within 30 minutes
- **Injury reports**: Updated within 30 minutes
- **Roster changes**: Updated daily
- **Historical data**: Static (pulled once)

## Initial Full Import

Before starting the worker, run one-time full import:

```bash
# Import all historical data (2010-2025)
heroku run:detached "target/release/nfl-data-pipeline --mode full --start-year 2010 --end-year 2025"

# Monitor progress
heroku logs --dyno run.XXXX --tail
```

**Duration**: 30-60 minutes
**Data imported**:
- ~50,000 unique players
- ~4,300 games
- ~720,000 player stat records

After the full import, the worker maintains incremental updates.

## Monitoring & Alerts

### Heroku Dashboard
- Monitor dyno metrics
- Check memory/CPU usage
- View error rates

### Logs
```bash
# Real-time worker output
heroku logs --tail --dyno worker

# Filter for errors
heroku logs --tail --dyno worker | grep -i error

# Filter for successful updates
heroku logs --tail --dyno worker | grep "✅"
```

### Database Monitoring
```bash
# Check record counts
heroku pg:psql --command "SELECT COUNT(*) FROM players;"
heroku pg:psql --command "SELECT COUNT(*) FROM game_stats;"

# Check latest updates
heroku pg:psql --command "SELECT MAX(updated_at) FROM game_stats;"
```

## Troubleshooting

### Worker Not Starting
```bash
# Check worker logs
heroku logs --dyno worker

# Verify binary exists
heroku run "ls -la nfl-data-pipeline/target/release/"

# Manually test pipeline
heroku run "nfl-data-pipeline/target/release/nfl-data-pipeline --mode update"
```

### High Memory Usage
```bash
# Check dyno memory
heroku ps

# If memory issues, reduce batch size in Rust code
# Default: 500 records/batch
```

### Database Connection Issues
```bash
# Verify DATABASE_URL
heroku config:get DATABASE_URL

# Check database health
heroku pg:info
```

## Cost Optimization

### Current Setup (Eco Dynos)
- **Cost**: $5/month for Eco dynos plan
- **Includes**: 1000 dyno hours
- **Usage**: ~1000 hours (web + worker)

### Alternative: Scheduled Jobs (Lower Cost)
If you want to reduce costs, use Heroku Scheduler instead:
```bash
# Remove always-on worker
heroku ps:scale worker=0

# Add scheduler
heroku addons:create scheduler:standard

# Configure in dashboard:
# - Every 5 min during game hours
# - Every 30 min off hours
```

## Future Enhancements

Possible improvements:
- **Webhook integration**: Trigger updates on game events
- **Selective updates**: Only fetch changed data
- **Multi-region**: Deploy workers in multiple regions
- **Caching layer**: Add Redis for frequently accessed data
- **Metrics**: Track update latency and success rates
