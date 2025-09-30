# Heroku Deployment Guide

## Prerequisites

1. [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) installed
2. Heroku account created
3. Git repository initialized

## Quick Deploy

```bash
# Login to Heroku
heroku login

# Create a new Heroku app
heroku create gridironmind

# Add PostgreSQL addon (Hobby Basic recommended)
heroku addons:create heroku-postgresql:essential-0

# Add Redis addon
heroku addons:create heroku-redis:mini

# Set environment variables
heroku config:set ENVIRONMENT=production
heroku config:set API_KEY=your-api-key-here
heroku config:set CLAUDE_API_KEY=your-claude-api-key-here
heroku config:set DB_MAX_CONNS=25
heroku config:set DB_MIN_CONNS=5

# Deploy
git push heroku main

# Open the app
heroku open
```

## Configuration Details

### Environment Variables

Heroku automatically sets:
- `DATABASE_URL` - Set by heroku-postgresql addon
- `REDIS_URL` - Set by heroku-redis addon
- `PORT` - Set by Heroku (don't override)

You need to set:
- `ENVIRONMENT` - Set to "production"
- `API_KEY` - Your API authentication key
- `CLAUDE_API_KEY` - Your Claude API key for AI features
- `DB_MAX_CONNS` - Maximum database connections (25 recommended)
- `DB_MIN_CONNS` - Minimum database connections (5 recommended)

### Database Setup

After first deployment, run the schema:

```bash
# Get database URL
heroku config:get DATABASE_URL

# Run schema (local psql)
psql $(heroku config:get DATABASE_URL) -f schema.sql

# Or via Heroku CLI
heroku pg:psql < schema.sql
```

### Buildpack

Heroku automatically detects Go projects and uses the Go buildpack. The `go.mod` file tells Heroku to:
- Use Go 1.21
- Install all dependencies
- Build the binary from `cmd/server/main.go`

### Procfile

The `Procfile` tells Heroku how to run your app:
```
web: bin/server
```

This runs the compiled binary from `cmd/server/main.go`.

## Monitoring

```bash
# View logs
heroku logs --tail

# Check app status
heroku ps

# Check database
heroku pg:info

# Check Redis
heroku redis:info
```

## Scaling

```bash
# Scale web dynos (Hobby dyno is 1 instance)
heroku ps:scale web=1

# Upgrade to Standard dynos for better performance
heroku dyno:type web=standard-1x
```

## Testing

```bash
# Test health endpoint
curl https://your-app.herokuapp.com/health

# Test API endpoints
curl https://your-app.herokuapp.com/api/v1/players?limit=5

# Open dashboard
open https://your-app.herokuapp.com
```

## Continuous Deployment

### Option 1: GitHub Integration
1. Go to Heroku Dashboard
2. Select your app
3. Navigate to "Deploy" tab
4. Connect to GitHub repository
5. Enable automatic deploys from main branch

### Option 2: Git Push
```bash
# Every commit and push to heroku remote deploys
git add .
git commit -m "Update"
git push heroku main
```

## Troubleshooting

### Build Fails
```bash
# Check build logs
heroku logs --tail

# Ensure go.mod is correct
cat go.mod

# Try clearing build cache
heroku plugins:install heroku-repo
heroku repo:purge_cache -a gridironmind
git commit --allow-empty -m "Rebuild"
git push heroku main
```

### Database Connection Issues
```bash
# Check DATABASE_URL is set
heroku config:get DATABASE_URL

# Test connection
heroku pg:psql
\dt  # List tables

# Verify schema was applied
heroku pg:psql -c "SELECT COUNT(*) FROM players;"
```

### App Crashes
```bash
# Check logs
heroku logs --tail

# Check dyno status
heroku ps

# Restart dynos
heroku restart
```

### Performance Issues
```bash
# Check response times
heroku logs --tail | grep "Response"

# Upgrade database
heroku addons:upgrade heroku-postgresql:standard-0

# Scale dynos
heroku ps:scale web=2
```

## Cost Estimate

**Hobby Tier (Development):**
- Hobby Dyno: $7/month
- Heroku Postgres Essential-0: $5/month
- Heroku Redis Mini: $3/month
- **Total: ~$15/month**

**Production Tier:**
- Standard-1X Dyno: $25/month
- Heroku Postgres Standard-0: $50/month
- Heroku Redis Premium-0: $15/month
- **Total: ~$90/month**

## Best Practices

1. **Use environment-specific configs** - Different settings for dev/staging/prod
2. **Enable automatic deploys** - Deploy on every push to main
3. **Monitor logs regularly** - Set up log drains or use Heroku logging
4. **Backup database** - Use `heroku pg:backups:schedule`
5. **Set up alerts** - Configure Heroku metrics alerts
6. **Use connection pooling** - Already configured in `db.Config`
7. **Implement rate limiting** - Use middleware (Phase 7)

## Useful Commands

```bash
# View config vars
heroku config

# Set config var
heroku config:set KEY=value

# Run database migrations
heroku run go run cmd/migrate/main.go

# Access Rails console (if needed)
heroku run bash

# View addon info
heroku addons

# Open Heroku dashboard
heroku dashboard
```

## Local Testing

Test the Heroku build locally:

```bash
# Install Heroku Local
heroku local web

# This runs the Procfile locally
# Access at http://localhost:5000
```

## Additional Resources

- [Heroku Go Support](https://devcenter.heroku.com/articles/go-support)
- [Heroku Postgres](https://devcenter.heroku.com/articles/heroku-postgresql)
- [Heroku Redis](https://devcenter.heroku.com/articles/heroku-redis)
- [Heroku CLI Commands](https://devcenter.heroku.com/articles/heroku-cli-commands)