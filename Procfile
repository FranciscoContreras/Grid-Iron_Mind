release: psql $DATABASE_URL < schema.sql || true
web: server
# Always-on worker: Runs live updates during game days, regular updates off-hours
worker: bash nfl-data-pipeline/run-worker.sh