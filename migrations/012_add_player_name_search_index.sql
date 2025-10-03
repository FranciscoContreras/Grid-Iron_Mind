-- Migration 012: Add player name search optimization indexes
-- Created: 2025-10-03
-- Purpose: Optimize player name search with case-insensitive LIKE queries and full-text search

-- Add index for case-insensitive name search (LIKE queries)
-- This speeds up LOWER(name) LIKE '%search%' queries
CREATE INDEX IF NOT EXISTS idx_players_name_lower ON players (LOWER(name));

-- Add index for prefix search (name starts with)
-- This speeds up LOWER(name) LIKE 'search%' queries
CREATE INDEX IF NOT EXISTS idx_players_name_lower_prefix ON players (LOWER(name) text_pattern_ops);

-- Add trigram index for fuzzy matching (optional, for future enhancements)
-- Requires pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_players_name_trgm ON players USING gin (name gin_trgm_ops);

-- Add full-text search index for advanced search (optional, for future)
CREATE INDEX IF NOT EXISTS idx_players_name_fts ON players USING gin (to_tsvector('english', name));

-- Add index for common filter combinations with search
CREATE INDEX IF NOT EXISTS idx_players_status_name ON players (status, LOWER(name));
CREATE INDEX IF NOT EXISTS idx_players_position_name ON players (position, LOWER(name));

-- Analyze table to update statistics for query planner
ANALYZE players;
