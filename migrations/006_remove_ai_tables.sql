-- Migration 006: Remove AI Tables
-- Purpose: Remove AI prediction and analysis tables as AI features are being removed

-- Drop AI tables
DROP TABLE IF EXISTS predictions CASCADE;
DROP TABLE IF EXISTS ai_analysis CASCADE;

-- Drop AI indexes (if they still exist)
DROP INDEX IF EXISTS idx_predictions_entity_id;
DROP INDEX IF EXISTS idx_predictions_type;
DROP INDEX IF EXISTS idx_predictions_valid_until;
DROP INDEX IF EXISTS idx_ai_analysis_type;
DROP INDEX IF EXISTS idx_ai_analysis_expires_at;
