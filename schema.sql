-- Grid Iron Mind Database Schema
-- NFL data lake with AI enrichment capabilities

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nfl_id INTEGER UNIQUE,
    name TEXT NOT NULL,
    abbreviation TEXT NOT NULL UNIQUE,
    city TEXT NOT NULL,
    conference TEXT NOT NULL,
    division TEXT NOT NULL,
    stadium TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Players table
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nfl_id INTEGER UNIQUE,
    name TEXT NOT NULL,
    position TEXT NOT NULL,
    team_id UUID REFERENCES teams(id),
    jersey_number INTEGER,
    height_inches INTEGER,
    weight_pounds INTEGER,
    birth_date DATE,
    college TEXT,
    draft_year INTEGER,
    draft_round INTEGER,
    draft_pick INTEGER,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Games table
CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    nfl_game_id TEXT UNIQUE NOT NULL,
    home_team_id UUID NOT NULL REFERENCES teams(id),
    away_team_id UUID NOT NULL REFERENCES teams(id),
    game_date TIMESTAMP NOT NULL,
    season INTEGER NOT NULL,
    week INTEGER NOT NULL,
    home_score INTEGER,
    away_score INTEGER,
    status TEXT DEFAULT 'scheduled',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Game Stats table
CREATE TABLE IF NOT EXISTS game_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id),
    game_id UUID NOT NULL REFERENCES games(id),
    season INTEGER NOT NULL,
    week INTEGER NOT NULL,
    passing_yards INTEGER DEFAULT 0,
    rushing_yards INTEGER DEFAULT 0,
    receiving_yards INTEGER DEFAULT 0,
    touchdowns INTEGER DEFAULT 0,
    interceptions INTEGER DEFAULT 0,
    completions INTEGER DEFAULT 0,
    attempts INTEGER DEFAULT 0,
    targets INTEGER DEFAULT 0,
    receptions INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, game_id)
);

-- Predictions table (AI)
CREATE TABLE IF NOT EXISTS predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    prediction_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    prediction_data JSONB NOT NULL,
    confidence_score DECIMAL(3,2) CHECK (confidence_score >= 0 AND confidence_score <= 1),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMP NOT NULL,
    actual_outcome JSONB,
    accuracy_score DECIMAL(3,2) CHECK (accuracy_score >= 0 AND accuracy_score <= 1)
);

-- AI Analysis table
CREATE TABLE IF NOT EXISTS ai_analysis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    analysis_type TEXT NOT NULL,
    subject_ids JSONB NOT NULL,
    analysis_result JSONB NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_players_nfl_id ON players(nfl_id);
CREATE INDEX IF NOT EXISTS idx_players_team_id ON players(team_id);
CREATE INDEX IF NOT EXISTS idx_players_position ON players(position);
CREATE INDEX IF NOT EXISTS idx_teams_abbreviation ON teams(abbreviation);
CREATE INDEX IF NOT EXISTS idx_game_stats_player_id ON game_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_game_stats_game_id ON game_stats(game_id);
CREATE INDEX IF NOT EXISTS idx_game_stats_season ON game_stats(season);
CREATE INDEX IF NOT EXISTS idx_games_game_date ON games(game_date);
CREATE INDEX IF NOT EXISTS idx_games_season_week ON games(season, week);
CREATE INDEX IF NOT EXISTS idx_predictions_entity_id ON predictions(entity_id);
CREATE INDEX IF NOT EXISTS idx_predictions_type ON predictions(prediction_type);
CREATE INDEX IF NOT EXISTS idx_predictions_valid_until ON predictions(valid_until);
CREATE INDEX IF NOT EXISTS idx_ai_analysis_type ON ai_analysis(analysis_type);
CREATE INDEX IF NOT EXISTS idx_ai_analysis_expires_at ON ai_analysis(expires_at);

-- Sample test data for development
-- Kansas City Chiefs
INSERT INTO teams (nfl_id, name, abbreviation, city, conference, division, stadium)
VALUES (12, 'Chiefs', 'KC', 'Kansas City', 'AFC', 'West', 'GEHA Field at Arrowhead Stadium')
ON CONFLICT (abbreviation) DO NOTHING;

-- Buffalo Bills
INSERT INTO teams (nfl_id, name, abbreviation, city, conference, division, stadium)
VALUES (2, 'Bills', 'BUF', 'Buffalo', 'AFC', 'East', 'Highmark Stadium')
ON CONFLICT (abbreviation) DO NOTHING;

-- San Francisco 49ers
INSERT INTO teams (nfl_id, name, abbreviation, city, conference, division, stadium)
VALUES (25, '49ers', 'SF', 'San Francisco', 'NFC', 'West', 'Levi''s Stadium')
ON CONFLICT (abbreviation) DO NOTHING;

-- Sample players (using fictional data for testing)
INSERT INTO players (nfl_id, name, position, team_id, jersey_number, height_inches, weight_pounds, status)
SELECT 3139477, 'Patrick Mahomes', 'QB', id, 15, 75, 230, 'active'
FROM teams WHERE abbreviation = 'KC'
ON CONFLICT (nfl_id) DO NOTHING;

INSERT INTO players (nfl_id, name, position, team_id, jersey_number, height_inches, weight_pounds, status)
SELECT 3918298, 'Josh Allen', 'QB', id, 17, 77, 237, 'active'
FROM teams WHERE abbreviation = 'BUF'
ON CONFLICT (nfl_id) DO NOTHING;