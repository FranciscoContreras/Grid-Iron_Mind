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
ON CONFLICT (nfl_id) DO NOTHING;-- Migration 002: Add Historical Data Support
-- Enhances schema to support complete player career history, weather data, and location details

-- Add latitude/longitude to teams for weather tracking
ALTER TABLE teams ADD COLUMN IF NOT EXISTS stadium_lat DECIMAL(10, 8);
ALTER TABLE teams ADD COLUMN IF NOT EXISTS stadium_lon DECIMAL(11, 8);
ALTER TABLE teams ADD COLUMN IF NOT EXISTS stadium_type TEXT; -- outdoor, indoor, retractable
ALTER TABLE teams ADD COLUMN IF NOT EXISTS stadium_surface TEXT; -- grass, turf
ALTER TABLE teams ADD COLUMN IF NOT EXISTS stadium_capacity INTEGER;

-- Add more player career metadata
ALTER TABLE players ADD COLUMN IF NOT EXISTS rookie_year INTEGER;
ALTER TABLE players ADD COLUMN IF NOT EXISTS years_pro INTEGER;
ALTER TABLE players ADD COLUMN IF NOT EXISTS headshot_url TEXT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS birth_city TEXT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS birth_state TEXT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS birth_country TEXT;

-- Add weather and game conditions to games table
ALTER TABLE games ADD COLUMN IF NOT EXISTS venue_id TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS venue_name TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS venue_city TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS venue_state TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS venue_type TEXT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS attendance INTEGER;
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_temp INTEGER; -- Fahrenheit
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_condition TEXT; -- clear, rain, snow, etc.
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_wind_speed INTEGER; -- mph
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_humidity INTEGER; -- percentage
ALTER TABLE games ADD COLUMN IF NOT EXISTS game_time_et TEXT; -- kickoff time
ALTER TABLE games ADD COLUMN IF NOT EXISTS playoff_round TEXT; -- wild-card, divisional, conference, super-bowl

-- Enhance game_stats with more detailed statistics
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS passing_tds INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS fumbles INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS fumbles_lost INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS sacks INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS tackles INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS tackles_for_loss INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS qb_hits INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS passes_defended INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS interceptions_thrown INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS receiving_tds INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS rushing_tds INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS two_point_conversions INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS field_goals_made INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS field_goals_attempted INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS extra_points_made INTEGER DEFAULT 0;
ALTER TABLE game_stats ADD COLUMN IF NOT EXISTS extra_points_attempted INTEGER DEFAULT 0;

-- Create player_career_stats table for aggregated career statistics
CREATE TABLE IF NOT EXISTS player_career_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    season INTEGER NOT NULL,
    team_id UUID REFERENCES teams(id),
    games_played INTEGER DEFAULT 0,
    games_started INTEGER DEFAULT 0,

    -- Passing stats
    passing_yards INTEGER DEFAULT 0,
    passing_tds INTEGER DEFAULT 0,
    passing_ints INTEGER DEFAULT 0,
    passing_completions INTEGER DEFAULT 0,
    passing_attempts INTEGER DEFAULT 0,
    passing_rating DECIMAL(5, 2),

    -- Rushing stats
    rushing_yards INTEGER DEFAULT 0,
    rushing_tds INTEGER DEFAULT 0,
    rushing_attempts INTEGER DEFAULT 0,
    rushing_long INTEGER DEFAULT 0,

    -- Receiving stats
    receptions INTEGER DEFAULT 0,
    receiving_yards INTEGER DEFAULT 0,
    receiving_tds INTEGER DEFAULT 0,
    receiving_targets INTEGER DEFAULT 0,
    receiving_long INTEGER DEFAULT 0,

    -- Defensive stats
    tackles INTEGER DEFAULT 0,
    sacks DECIMAL(3, 1) DEFAULT 0,
    interceptions INTEGER DEFAULT 0,
    forced_fumbles INTEGER DEFAULT 0,
    fumble_recoveries INTEGER DEFAULT 0,
    passes_defended INTEGER DEFAULT 0,

    -- Kicking stats
    field_goals_made INTEGER DEFAULT 0,
    field_goals_attempted INTEGER DEFAULT 0,
    extra_points_made INTEGER DEFAULT 0,
    extra_points_attempted INTEGER DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(player_id, season, team_id)
);

-- Create player_team_history table to track which teams a player played for
CREATE TABLE IF NOT EXISTS player_team_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id),
    season_start INTEGER NOT NULL,
    season_end INTEGER,
    position TEXT NOT NULL,
    jersey_number INTEGER,
    is_current BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(player_id, team_id, season_start)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_game_stats_season_week ON game_stats(season, week);
CREATE INDEX IF NOT EXISTS idx_player_career_stats_player_id ON player_career_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_player_career_stats_season ON player_career_stats(season);
CREATE INDEX IF NOT EXISTS idx_player_career_stats_team_id ON player_career_stats(team_id);
CREATE INDEX IF NOT EXISTS idx_player_team_history_player_id ON player_team_history(player_id);
CREATE INDEX IF NOT EXISTS idx_player_team_history_team_id ON player_team_history(team_id);
CREATE INDEX IF NOT EXISTS idx_player_team_history_current ON player_team_history(is_current) WHERE is_current = true;
CREATE INDEX IF NOT EXISTS idx_games_venue_city ON games(venue_city);
CREATE INDEX IF NOT EXISTS idx_games_weather ON games(weather_condition) WHERE weather_condition IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_games_playoff ON games(playoff_round) WHERE playoff_round IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_players_rookie_year ON players(rookie_year);
-- Game Team Stats table (from migration 003)
CREATE TABLE IF NOT EXISTS game_team_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id),
    
    -- Offensive Stats
    first_downs INT DEFAULT 0,
    total_yards INT DEFAULT 0,
    passing_yards INT DEFAULT 0,
    rushing_yards INT DEFAULT 0,
    offensive_plays INT DEFAULT 0,
    yards_per_play DECIMAL(4,2),
    
    -- Efficiency
    third_down_attempts INT DEFAULT 0,
    third_down_conversions INT DEFAULT 0,
    third_down_pct DECIMAL(5,2),
    fourth_down_attempts INT DEFAULT 0,
    fourth_down_conversions INT DEFAULT 0,
    fourth_down_pct DECIMAL(5,2),
    red_zone_attempts INT DEFAULT 0,
    red_zone_scores INT DEFAULT 0,
    
    -- Turnovers & Penalties
    turnovers INT DEFAULT 0,
    fumbles_lost INT DEFAULT 0,
    interceptions_thrown INT DEFAULT 0,
    penalties INT DEFAULT 0,
    penalty_yards INT DEFAULT 0,
    
    -- Possession
    possession_time VARCHAR(10), -- MM:SS format
    possession_seconds INT,
    
    -- Passing Detail
    completions INT DEFAULT 0,
    pass_attempts INT DEFAULT 0,
    sacks_allowed INT DEFAULT 0,
    sack_yards INT DEFAULT 0,
    
    -- Rushing Detail
    rushing_attempts INT DEFAULT 0,
    rushing_avg DECIMAL(4,2),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, team_id)
);

CREATE INDEX IF NOT EXISTS idx_game_team_stats_game ON game_team_stats(game_id);
CREATE INDEX IF NOT EXISTS idx_game_team_stats_team ON game_team_stats(team_id);
CREATE INDEX IF NOT EXISTS idx_game_team_stats_yards ON game_team_stats(total_yards);
