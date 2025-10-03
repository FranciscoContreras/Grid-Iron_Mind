release: psql $DATABASE_URL < schema.sql || true
web: server
# Always-on worker: Initial import + continuous updates
worker: bash nfl-data-pipeline/run-worker-with-init.sh
# Temporary OAuth helper - run manually with: heroku ps:scale oauth=1
oauth: yahoo_oauth_helper