-- Migration 009: Add historical import support
-- Date: 2025-01-15
-- Purpose: Add import progress tracking and constraints for historical data import

-- Create import progress tracking table
CREATE TABLE IF NOT EXISTS import_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season INT NOT NULL,
    data_type VARCHAR(50) NOT NULL,  -- 'rosters', 'games', 'player_stats', 'ngs'
    status VARCHAR(20) NOT NULL,      -- 'pending', 'in_progress', 'completed', 'failed'
    records_imported INT DEFAULT 0,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(season, data_type)
);

-- Create data quality metrics table
CREATE TABLE IF NOT EXISTS data_quality_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    season INT NOT NULL,
    metric_name VARCHAR(100) NOT NULL,
    metric_value JSONB,
    measured_at TIMESTAMP DEFAULT NOW()
);

-- Add unique constraint to game_stats to prevent duplicate stat entries
-- This allows upsert operations during historical import
ALTER TABLE game_stats DROP CONSTRAINT IF EXISTS game_stats_player_season_week_unique;
ALTER TABLE game_stats ADD CONSTRAINT game_stats_player_season_week_unique
    UNIQUE (player_id, season, week);

-- Add index for faster import queries
CREATE INDEX IF NOT EXISTS idx_game_stats_player_season_week ON game_stats(player_id, season, week);
CREATE INDEX IF NOT EXISTS idx_import_progress_season ON import_progress(season);
CREATE INDEX IF NOT EXISTS idx_import_progress_status ON import_progress(status);

-- Add comments for documentation
COMMENT ON TABLE import_progress IS 'Tracks progress of historical data imports';
COMMENT ON TABLE data_quality_metrics IS 'Stores data quality metrics for validation';
COMMENT ON COLUMN import_progress.data_type IS 'Type of data being imported: rosters, games, player_stats, ngs';
COMMENT ON COLUMN import_progress.status IS 'Import status: pending, in_progress, completed, failed';
