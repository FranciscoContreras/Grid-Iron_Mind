release: psql $DATABASE_URL < schema.sql || true
web: server
# Run Rust pipeline with: heroku run worker
worker: nfl-data-pipeline/target/release/nfl-data-pipeline --mode update