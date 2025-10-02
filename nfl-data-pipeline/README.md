# NFL Data Pipeline (Rust)

High-performance local data import pipeline that fetches NFL data from nflfastr, processes it locally, and syncs to Heroku PostgreSQL.

## Architecture

```
nflfastr Data (GitHub) → Rust Pipeline → Local Processing → Heroku PostgreSQL → API
```

**Benefits:**
- **Fast**: Rust's performance handles large datasets efficiently
- **Reliable**: Process and validate data locally before pushing to production
- **Comprehensive**: nflfastr has 10+ years of rich NFL data
- **Flexible**: Easy to add new data sources or transformations

## Prerequisites

### Install Rust

```bash
# Install Rust via rustup
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Restart shell or run:
source $HOME/.cargo/env

# Verify installation
cargo --version
```

### Environment Setup

Create `.env` file in `nfl-data-pipeline/`:

```env
DATABASE_URL=postgresql://username:password@host:port/database
# Or use Heroku database URL:
# DATABASE_URL=postgres://...@...ec2-...amazonaws.com:5432/...
```

## Installation

```bash
cd nfl-data-pipeline

# Build in release mode (optimized)
cargo build --release

# Or run directly
cargo run --release
```

## Usage

### Full Import (10 years: 2015-2024)

```bash
cargo run --release -- --mode full --start-year 2015 --end-year 2024
```

### Single Season Import

```bash
cargo run --release -- --mode year --year 2024
```

### Update Current Season

```bash
cargo run --release -- --mode update
```

### Dry Run (Test without writing to database)

```bash
cargo run --release -- --mode full --start-year 2024 --end-year 2024 --dry-run
```

### Validate Existing Data

```bash
cargo run --release -- --mode validate
```

## Data Sources

The pipeline fetches from nflfastr GitHub releases:

- **Play-by-play data**: `https://github.com/nflverse/nflverse-data/releases/download/pbp/play_by_play_{year}.csv.gz`
- **Player stats**: `https://github.com/nflverse/nflverse-data/releases/download/player_stats/player_stats_{year}.csv`
- **Rosters**: `https://github.com/nflverse/nflverse-data/releases/download/rosters/roster_{year}.csv`
- **Schedule**: `https://github.com/nflverse/nflverse-data/releases/download/schedules/sched_{year}.csv`
- **Next Gen Stats**: `https://github.com/nflverse/nflverse-data/releases/download/nextgen_stats/ngs_{year}_passing.csv`

## Pipeline Stages

### 1. Download (Parallel)
- Downloads CSV files from nflfastr GitHub
- Supports gzip compression
- Concurrent downloads per season

### 2. Parse & Transform
- CSV → Rust structs
- Data validation and cleaning
- Team abbreviation mapping (historical → current)
- Player ID normalization

### 3. Local Storage (Optional)
- Saves processed data to local SQLite/Parquet
- Enables incremental updates
- Faster subsequent runs

### 4. Database Sync
- Batch UPSERT operations
- Transaction-safe (rollback on error)
- Progress tracking
- Handles duplicates gracefully

### 5. Validation
- Row count verification
- Data integrity checks
- Missing data reports

## Performance

**Expected import times (M1 Mac):**

| Dataset | Size | Time |
|---------|------|------|
| Single season | ~15K players, 267 games, 40K stats | 30-60s |
| 10 years (2015-2024) | ~150K players, 2.7K games, 400K stats | 5-10 min |

**Optimizations:**
- Parallel downloads (rayon)
- Batch inserts (500 rows)
- Connection pooling
- Compiled release mode

## Project Structure

```
nfl-data-pipeline/
├── Cargo.toml           # Dependencies
├── README.md            # This file
├── .env                 # Database credentials (gitignored)
├── src/
│   ├── main.rs          # CLI entry point
│   ├── config.rs        # Configuration management
│   ├── downloader.rs    # HTTP client for nflfastr data
│   ├── parser.rs        # CSV parsing and data models
│   ├── transformer.rs   # Data cleaning and normalization
│   ├── database.rs      # PostgreSQL connection and queries
│   ├── sync.rs          # Database sync logic
│   └── validator.rs     # Data validation
└── data/                # Local cache (gitignored)
    ├── raw/             # Downloaded CSVs
    └── processed/       # Transformed data
```

## Error Handling

- **Network errors**: Automatic retries with exponential backoff
- **Parse errors**: Skip malformed rows, log warnings
- **Database errors**: Rollback transaction, exit cleanly
- **Validation failures**: Generate report, optionally abort

## Logging

```bash
# Enable debug logging
RUST_LOG=debug cargo run --release

# Info level (default)
RUST_LOG=info cargo run --release

# Only errors
RUST_LOG=error cargo run --release
```

## Incremental Updates

The pipeline supports incremental updates:

```bash
# First run: Full import (slow)
cargo run --release -- --mode full --start-year 2015 --end-year 2024

# Subsequent runs: Only fetch new/updated data
cargo run --release -- --mode update
```

This checks `import_progress` table and only re-imports seasons with updates.

## Team Abbreviation Mapping

Handles historical team changes:

```rust
STL → LA   // St. Louis Rams → Los Angeles Rams
SD  → LAC  // San Diego Chargers → Los Angeles Chargers
OAK → LV   // Oakland Raiders → Las Vegas Raiders
```

## Contributing

1. Add new data source in `downloader.rs`
2. Define CSV models in `parser.rs`
3. Add transformation logic in `transformer.rs`
4. Update database sync in `sync.rs`
5. Test with `--dry-run` flag

## Troubleshooting

### "Failed to connect to database"

Check `DATABASE_URL` in `.env` file. For Heroku:

```bash
# Get Heroku database URL
heroku config:get DATABASE_URL --app grid-iron-mind

# Test connection
psql $DATABASE_URL -c "SELECT 1;"
```

### "Download failed: 404"

nflfastr data for very recent seasons may not be available yet. Try previous year.

### "Out of memory"

Reduce batch size in `sync.rs` or process one season at a time.

## Roadmap

- [ ] Parquet output format for faster local caching
- [ ] Streaming CSV parser for large files
- [ ] Delta sync (only changed rows)
- [ ] Web UI for monitoring imports
- [ ] Docker containerization
- [ ] CI/CD integration

## License

MIT
