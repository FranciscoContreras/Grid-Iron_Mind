release: psql $DATABASE_URL < schema.sql || true
web: server
# Always-on worker: Runs live updates during game days, regular updates off-hours
worker: bash nfl-data-pipeline/run-worker.sh
# Temporary OAuth helper - run manually with: heroku ps:scale oauth=1
oauth: yahoo_oauth_helper