-- Migration 002: Add Historical Data Support
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