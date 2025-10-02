# Historical Data Import Guide

## 📚 Overview

This guide explains how to import 15 years of historical NFL data (2010-2024) into the Grid Iron Mind database using the automated import tool.

## 🎯 What Gets Imported

For each season from 2010 to 2024:
- **Rosters**: ~1,700 players per season (~25,000 total unique players)
- **Games**: ~267 games per season (~4,000 total games)
- **Player Statistics**: ~15,000 stat records per season (~225,000 total)
- **Next Gen Stats**: Available from 2016+ (~40,000 records)

**Estimated Database Growth**: 2-3 GB

## 🚀 Quick Start

### 1. Build the Import Tool

```bash
make build-importer
```

This creates `bin/import-historical` executable.

### 2. Test with Single Year (Recommended)

Start with a recent year to test the process:

```bash
make import-year YEAR=2024
```

**Duration**: ~5 minutes
**Records**: ~15,000 player stats, ~267 games, ~1,700 players

### 3. Validate the Import

```bash
make import-validate
```

This shows which seasons have data and identifies gaps.

### 4. Import Full Historical Range

Once you're confident, import all 15 years:

```bash
make import-full
```

**Duration**: 60-90 minutes
**Records**: ~225,000 player stats, ~4,000 games, ~25,000 players

## 📋 Available Commands

### Import Commands

| Command | Description | Duration |
|---------|-------------|----------|
| `make import-year YEAR=2020` | Import single year | ~5 min |
| `make import-range START=2010 END=2014` | Import year range | ~25 min (for 5 years) |
| `make import-full` | Import all 15 years (2010-2024) | 60-90 min |

### Validation Commands

| Command | Description |
|---------|-------------|
| `make import-validate` | Check data coverage by season |
| `make import-stats` | Show total counts and breakdown |

### Manual Execution

You can also run the tool directly for more control:

```bash
# Single year with verbose output
./bin/import-historical --mode=year --year=2023 --verbose

# Year range
./bin/import-historical --mode=range --start=2015 --end=2019 --verbose

# Dry run (test without importing)
./bin/import-historical --mode=year --year=2024 --dry-run

# Full import
./bin/import-historical --mode=full --start=2010 --end=2024
```

## 📊 Import Process

For each year, the tool follows this sequence:

```
1. Rosters (Players)
   ↓ Downloads roster CSV from NFLverse
   ↓ Upserts players into database
   ↓ Links players to teams

2. Schedule (Games)
   ↓ Downloads schedule CSV
   ↓ Filters to regular season games
   ↓ Inserts games with scores

3. Player Statistics
   ↓ Downloads player stats CSV
   ↓ Filters to regular season
   ↓ Batch inserts stats (500 at a time)

4. Next Gen Stats (2016+)
   ↓ Downloads passing/rushing/receiving NGS
   ↓ Inserts advanced metrics
```

## 🔍 Monitoring Progress

### View Live Logs

```bash
tail -f logs/import-historical.log
```

### Check Database Progress

```bash
# Count games by season
psql $DATABASE_URL -c "SELECT season, COUNT(*) FROM games GROUP BY season ORDER BY season DESC;"

# Count stats by season
psql $DATABASE_URL -c "SELECT season, COUNT(*) FROM game_stats GROUP BY season ORDER BY season DESC;"
```

### Import Progress Table

The tool tracks progress in the `import_progress` table:

```sql
SELECT season, data_type, status, records_imported,
       completed_at - started_at as duration
FROM import_progress
ORDER BY season DESC, data_type;
```

## ⚠️ Important Considerations

### Before You Start

1. **Database Space**: Ensure you have at least 5 GB free space
2. **Connection**: Stable internet connection required
3. **Time**: Full import takes 60-90 minutes
4. **API Limits**: NFLverse has no rate limits, but be respectful

### During Import

- **Don't Interrupt**: Interrupting mid-import can leave partial data
- **Resume Support**: Each year is independent, so you can resume from any year
- **Error Handling**: Tool continues even if individual records fail
- **Database Load**: Minimal impact on production API (uses batch inserts)

### After Import

1. **Validate Data**: Run `make import-validate` to check coverage
2. **Analyze Tables**: Run `ANALYZE` on large tables for query optimization
3. **Check Indexes**: Ensure all indexes are present
4. **Test Queries**: Verify API endpoints work with historical data

## 🛠️ Troubleshooting

### Import Fails with Connection Error

**Problem**: Database connection timeout

**Solution**:
```bash
# Check DATABASE_URL is set
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1;"

# Retry import for specific year
make import-year YEAR=2023
```

### Import Runs But No Data Appears

**Problem**: Dry-run mode or permission issues

**Solution**:
```bash
# Remove --dry-run flag
./bin/import-historical --mode=year --year=2024

# Check table permissions
psql $DATABASE_URL -c "SELECT COUNT(*) FROM players;"
```

### CSV Download Fails

**Problem**: NFLverse GitHub unavailable

**Solution**:
- Wait a few minutes and retry
- Check https://github.com/nflverse/nflverse-data/releases
- Import different year range while waiting

### Duplicate Key Errors

**Problem**: Data already exists for that season

**Solution**:
- This is normal! The tool uses UPSERT operations
- Existing records are updated, not duplicated
- Safe to re-run import for any year

## 📈 Expected Results

### After Full Import (2010-2024)

```
📊 Import Statistics
════════════════════════════════════════════════
Total Players:      ~25,000
Total Games:        ~4,000
Total Game Stats:   ~225,000
Total Advanced Stats: ~40,000

By Season:
  2024: 267 games
  2023: 272 games
  2022: 272 games
  ...
  2011: 256 games
  2010: 256 games
════════════════════════════════════════════════
```

### Data Coverage Validation

```
🔍 Validating imported data...

Season Coverage:
────────────────────────────────
✅ 2024: 267 games (COMPLETE)
✅ 2023: 272 games (COMPLETE)
✅ 2022: 272 games (COMPLETE)
...
✅ 2011: 256 games (COMPLETE)
✅ 2010: 256 games (COMPLETE)
────────────────────────────────

Summary:
  Complete: 15 seasons
  Incomplete: 0 seasons
  Missing: 0 seasons
```

## 🎓 Advanced Usage

### Import Specific Data Types Only

Modify `cmd/import_historical/main.go` to skip certain data types:

```go
// In importYear function, comment out steps you don't want:

// Skip Next Gen Stats
// if year >= 2016 {
//     i.importNextGenStats(year, stats)
// }
```

### Custom Year Ranges

Import only playoff seasons:

```bash
# 2010s decade
make import-range START=2010 END=2019

# 2020s so far
make import-range START=2020 END=2024
```

### Parallel Imports (Advanced)

Import different year ranges in parallel:

```bash
# Terminal 1
make import-range START=2010 END=2014

# Terminal 2
make import-range START=2015 END=2019

# Terminal 3
make import-range START=2020 END=2024
```

**Note**: Ensure each range doesn't overlap to avoid database conflicts.

## 🔄 Incremental Updates

### Updating Existing Data

Re-run import for a specific year to update with latest corrections:

```bash
make import-year YEAR=2024
```

This will:
- Update existing players
- Update game scores
- Upsert statistics (existing records updated, new records added)

### Adding New Seasons

When 2026 season data becomes available:

```bash
make import-year YEAR=2026
```

## 📝 Data Sources

All data imported from [NFLverse](https://github.com/nflverse/nflverse-data):

- **Player Stats**: `player_stats/player_stats_YYYY.csv`
- **Schedule**: `schedules/sched_YYYY.csv`
- **Rosters**: `rosters/roster_YYYY.csv`
- **Next Gen Stats**: `nextgen_stats/ngs_YYYY_[type].csv`

Data is community-maintained, open-source, and validated against official NFL data.

## 🤝 Contributing

Found an issue with the import process? Submit a bug report with:

1. Year being imported
2. Error message
3. Log file (`logs/import-historical.log`)
4. Database state (`make import-stats`)

## 📚 Related Documentation

- [Historical Data Import Plan](./HISTORICAL_DATA_IMPORT_PLAN.md) - Detailed technical plan
- [API Documentation](./API_DOCUMENTATION.md) - API endpoints for historical data
- [Database Schema](../schema.sql) - Complete database structure
- [NFLverse Data Dictionary](https://nflverse.github.io/nflverse-data/) - Source data documentation

## ✅ Checklist

Before starting full import:

- [ ] Database has 5+ GB free space
- [ ] Stable internet connection
- [ ] DATABASE_URL environment variable set
- [ ] Test import completed successfully (`make import-year YEAR=2024`)
- [ ] Validation shows correct results (`make import-validate`)
- [ ] 60-90 minutes available for full import
- [ ] Logs directory exists (`logs/`)

After full import:

- [ ] Validation shows all 15 seasons complete
- [ ] Statistics show expected counts (`make import-stats`)
- [ ] API endpoints return historical data
- [ ] Database size increased by 2-3 GB
- [ ] Query performance is acceptable

---

**Ready to import?** Start with `make import-year YEAR=2024` to test! 🚀
