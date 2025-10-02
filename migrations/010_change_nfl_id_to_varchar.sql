-- Migration 010: Change nfl_id from integer to varchar
-- Date: 2025-10-02
-- Purpose: Support NFLverse player IDs which are strings like "00-0022531"

-- Drop existing index on nfl_id
DROP INDEX IF EXISTS idx_players_nfl_id;

-- Change nfl_id column type to VARCHAR
ALTER TABLE players ALTER COLUMN nfl_id TYPE VARCHAR(20) USING nfl_id::VARCHAR;

-- Recreate index
CREATE INDEX idx_players_nfl_id ON players(nfl_id);

-- Add comment
COMMENT ON COLUMN players.nfl_id IS 'NFL player ID from NFLverse (format: XX-XXXXXXX)';
