-- Migration 003: Comprehensive Schema Enhancements
-- Purpose: Transform Grid Iron Mind into the most comprehensive NFL API
-- Created: 2025-09-30

-- ========================================
-- PART 1: ENHANCE EXISTING TABLES
-- ========================================

-- 1.1 Expand weather data in games table
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_wind_dir VARCHAR(10);
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_pressure INT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_visibility INT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_feels_like INT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_precipitation DECIMAL(4,2);
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_cloud_cover INT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS weather_uv_index DECIMAL(3,1);
ALTER TABLE games ADD COLUMN IF NOT EXISTS is_day_game BOOLEAN;

-- 1.2 Add game status details
ALTER TABLE games ADD COLUMN IF NOT EXISTS status_detail VARCHAR(100);
ALTER TABLE games ADD COLUMN IF NOT EXISTS current_period INT;
ALTER TABLE games ADD COLUMN IF NOT EXISTS game_clock VARCHAR(10);

-- 1.3 Enhance teams table
ALTER TABLE teams ADD COLUMN IF NOT EXISTS uid VARCHAR(100);
ALTER TABLE teams ADD COLUMN IF NOT EXISTS slug VARCHAR(100);
ALTER TABLE teams ADD COLUMN IF NOT EXISTS alternate_color VARCHAR(10);
ALTER TABLE teams ADD COLUMN IF NOT EXISTS logo_url TEXT;
ALTER TABLE teams ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- 1.4 Enhance players table
ALTER TABLE players ADD COLUMN IF NOT EXISTS short_name VARCHAR(100);
ALTER TABLE players ADD COLUMN IF NOT EXISTS display_name VARCHAR(100);
ALTER TABLE players ADD COLUMN IF NOT EXISTS espn_id VARCHAR(50);
ALTER TABLE players ADD COLUMN IF NOT EXISTS experience_years INT;
ALTER TABLE players ADD COLUMN IF NOT EXISTS status_detail VARCHAR(50);

-- ========================================
-- PART 2: CREATE NEW COMPREHENSIVE TABLES
-- ========================================

-- 2.1 Game Team Statistics (Box Score)
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

CREATE INDEX idx_game_team_stats_game ON game_team_stats(game_id);
CREATE INDEX idx_game_team_stats_team ON game_team_stats(team_id);
CREATE INDEX idx_game_team_stats_yards ON game_team_stats(total_yards);

-- 2.2 Player Season Statistics (Career History)
CREATE TABLE IF NOT EXISTS player_season_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    season INT NOT NULL,
    team_id UUID REFERENCES teams(id),
    position VARCHAR(10),

    -- Games
    games_played INT DEFAULT 0,
    games_started INT DEFAULT 0,

    -- Passing Stats (QB)
    passing_attempts INT DEFAULT 0,
    passing_completions INT DEFAULT 0,
    passing_yards INT DEFAULT 0,
    passing_tds INT DEFAULT 0,
    passing_ints INT DEFAULT 0,
    passing_rating DECIMAL(5,2),
    sacks INT DEFAULT 0,
    sack_yards INT DEFAULT 0,
    longest_pass INT DEFAULT 0,

    -- Rushing Stats (RB, QB)
    rushing_attempts INT DEFAULT 0,
    rushing_yards INT DEFAULT 0,
    rushing_tds INT DEFAULT 0,
    rushing_avg DECIMAL(4,2),
    longest_rush INT DEFAULT 0,
    fumbles INT DEFAULT 0,
    fumbles_lost INT DEFAULT 0,

    -- Receiving Stats (WR, TE, RB)
    receptions INT DEFAULT 0,
    receiving_yards INT DEFAULT 0,
    receiving_tds INT DEFAULT 0,
    receiving_avg DECIMAL(4,2),
    longest_reception INT DEFAULT 0,
    targets INT DEFAULT 0,

    -- Defensive Stats
    tackles_total INT DEFAULT 0,
    tackles_solo INT DEFAULT 0,
    tackles_assisted INT DEFAULT 0,
    sacks_defense DECIMAL(4,1) DEFAULT 0,
    tackles_for_loss INT DEFAULT 0,
    qb_hits INT DEFAULT 0,
    interceptions INT DEFAULT 0,
    passes_defended INT DEFAULT 0,
    forced_fumbles INT DEFAULT 0,
    fumble_recoveries INT DEFAULT 0,
    defensive_tds INT DEFAULT 0,

    -- Kicking Stats
    field_goals_made INT DEFAULT 0,
    field_goals_attempted INT DEFAULT 0,
    field_goal_pct DECIMAL(5,2),
    longest_field_goal INT DEFAULT 0,
    extra_points_made INT DEFAULT 0,
    extra_points_attempted INT DEFAULT 0,

    -- Punting Stats
    punts INT DEFAULT 0,
    punt_yards INT DEFAULT 0,
    punt_avg DECIMAL(4,2),
    longest_punt INT DEFAULT 0,
    punts_inside_20 INT DEFAULT 0,

    -- Return Stats
    kick_returns INT DEFAULT 0,
    kick_return_yards INT DEFAULT 0,
    kick_return_tds INT DEFAULT 0,
    punt_returns INT DEFAULT 0,
    punt_return_yards INT DEFAULT 0,
    punt_return_tds INT DEFAULT 0,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season)
);

CREATE INDEX idx_player_season_stats_player ON player_season_stats(player_id);
CREATE INDEX idx_player_season_stats_season ON player_season_stats(season);
CREATE INDEX idx_player_season_stats_team ON player_season_stats(team_id);
CREATE INDEX idx_player_season_stats_position ON player_season_stats(position);

-- 2.3 Team Season Standings
CREATE TABLE IF NOT EXISTS team_standings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    season INT NOT NULL,
    week INT, -- NULL for season totals

    -- Record
    wins INT DEFAULT 0,
    losses INT DEFAULT 0,
    ties INT DEFAULT 0,
    win_pct DECIMAL(5,3),

    -- Scoring
    points_for INT DEFAULT 0,
    points_against INT DEFAULT 0,
    point_differential INT DEFAULT 0,

    -- Split Records
    home_wins INT DEFAULT 0,
    home_losses INT DEFAULT 0,
    away_wins INT DEFAULT 0,
    away_losses INT DEFAULT 0,
    division_wins INT DEFAULT 0,
    division_losses INT DEFAULT 0,
    conference_wins INT DEFAULT 0,
    conference_losses INT DEFAULT 0,

    -- Streak
    current_streak VARCHAR(10), -- "W3", "L2", etc.

    -- Ranking
    division_rank INT,
    conference_rank INT,
    playoff_seed INT,

    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, season, week)
);

CREATE INDEX idx_team_standings_season ON team_standings(season);
CREATE INDEX idx_team_standings_week ON team_standings(week);
CREATE INDEX idx_team_standings_team ON team_standings(team_id);

-- 2.4 Game Scoring Plays (Timeline)
CREATE TABLE IF NOT EXISTS game_scoring_plays (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id),

    -- When
    quarter INT NOT NULL,
    time_remaining VARCHAR(10), -- "MM:SS"
    sequence_number INT, -- Order within game

    -- What
    play_type VARCHAR(50), -- TD, FG, Safety, 2PT, XP
    scoring_type VARCHAR(50), -- Passing TD, Rushing TD, etc.
    points INT NOT NULL,
    description TEXT,

    -- Players involved
    scoring_player_id UUID REFERENCES players(id),
    assist_player_id UUID REFERENCES players(id), -- QB on passing TD

    -- Score after
    home_score INT NOT NULL,
    away_score INT NOT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scoring_plays_game ON game_scoring_plays(game_id);
CREATE INDEX idx_scoring_plays_team ON game_scoring_plays(team_id);
CREATE INDEX idx_scoring_plays_player ON game_scoring_plays(scoring_player_id);
CREATE INDEX idx_scoring_plays_sequence ON game_scoring_plays(game_id, sequence_number);

-- 2.5 Advanced Stats (Next Gen Stats)
CREATE TABLE IF NOT EXISTS advanced_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    season INT NOT NULL,
    week INT, -- NULL for season totals
    stat_type VARCHAR(50) NOT NULL, -- passing, rushing, receiving

    -- Passing Advanced
    avg_time_to_throw DECIMAL(4,2),
    avg_completed_air_yards DECIMAL(5,2),
    avg_intended_air_yards DECIMAL(5,2),
    avg_air_yards_differential DECIMAL(5,2),
    max_completed_air_distance INT,
    avg_air_yards_to_sticks DECIMAL(5,2),
    attempts INT,
    pass_yards INT,
    pass_touchdowns INT,
    interceptions INT,
    passer_rating DECIMAL(5,2),
    completions INT,
    completion_percentage DECIMAL(5,2),
    expected_completion_percentage DECIMAL(5,2),
    completion_percentage_above_expectation DECIMAL(5,2),

    -- Rushing Advanced
    efficiency DECIMAL(5,2),
    percent_attempts_gte_eight_defenders DECIMAL(5,2),
    avg_time_to_los DECIMAL(4,2), -- line of scrimmage
    rush_attempts INT,
    rush_yards INT,
    expected_rush_yards INT,
    rush_yards_over_expected INT,
    avg_rush_yards DECIMAL(4,2),
    rush_touchdowns INT,

    -- Receiving Advanced
    avg_cushion DECIMAL(4,2),
    avg_separation DECIMAL(4,2),
    avg_intended_air_yards_receiving DECIMAL(5,2),
    percent_share_of_intended_air_yards DECIMAL(5,2),
    receptions INT,
    targets INT,
    catch_percentage DECIMAL(5,2),
    yards INT,
    rec_touchdowns INT,
    avg_yac DECIMAL(4,2),
    avg_expected_yac DECIMAL(4,2),
    avg_yac_above_expectation DECIMAL(4,2),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season, week, stat_type)
);

CREATE INDEX idx_advanced_stats_player ON advanced_stats(player_id);
CREATE INDEX idx_advanced_stats_season ON advanced_stats(season);
CREATE INDEX idx_advanced_stats_type ON advanced_stats(stat_type);

-- 2.6 Game Broadcasts Information
CREATE TABLE IF NOT EXISTS game_broadcasts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    network VARCHAR(50), -- CBS, FOX, NBC, ESPN, etc.
    broadcast_type VARCHAR(50), -- National, Regional
    announcers TEXT[], -- Array of announcer names
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(game_id, network)
);

CREATE INDEX idx_broadcasts_game ON game_broadcasts(game_id);
CREATE INDEX idx_broadcasts_network ON game_broadcasts(network);

-- ========================================
-- PART 3: QUERY OPTIMIZATION INDEXES
-- ========================================

-- Games indexes
CREATE INDEX IF NOT EXISTS idx_games_season_week ON games(season, week);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_date ON games(game_date);

-- Players indexes
CREATE INDEX IF NOT EXISTS idx_players_team ON players(team_id);
CREATE INDEX IF NOT EXISTS idx_players_position ON players(position);
CREATE INDEX IF NOT EXISTS idx_players_status ON players(status);
CREATE INDEX IF NOT EXISTS idx_players_espn_id ON players(espn_id);

-- Game stats indexes
CREATE INDEX IF NOT EXISTS idx_game_stats_player_season ON game_stats(player_id, season);
CREATE INDEX IF NOT EXISTS idx_game_stats_position ON game_stats(season, week);

-- ========================================
-- PART 4: HELPER FUNCTIONS
-- ========================================

-- Function to calculate win percentage
CREATE OR REPLACE FUNCTION calculate_win_pct(w INT, l INT, t INT)
RETURNS DECIMAL(5,3) AS $$
BEGIN
    IF (w + l + t) = 0 THEN
        RETURN 0.000;
    END IF;
    RETURN ROUND((w + (t * 0.5)) / (w + l + t), 3);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to calculate possession in seconds
CREATE OR REPLACE FUNCTION possession_to_seconds(possession VARCHAR(10))
RETURNS INT AS $$
DECLARE
    parts TEXT[];
    minutes INT;
    seconds INT;
BEGIN
    IF possession IS NULL OR possession = '' THEN
        RETURN 0;
    END IF;

    parts := string_to_array(possession, ':');
    IF array_length(parts, 1) = 2 THEN
        minutes := parts[1]::INT;
        seconds := parts[2]::INT;
        RETURN (minutes * 60) + seconds;
    END IF;

    RETURN 0;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ========================================
-- PART 5: MATERIALIZED VIEWS (OPTIONAL)
-- ========================================

-- View: Current Season Team Stats
CREATE MATERIALIZED VIEW IF NOT EXISTS vw_current_season_team_stats AS
SELECT
    t.id as team_id,
    t.name as team_name,
    t.abbreviation,
    COUNT(*) as games_played,
    SUM(CASE WHEN (g.home_team_id = t.id AND g.home_score > g.away_score) OR
                   (g.away_team_id = t.id AND g.away_score > g.home_score) THEN 1 ELSE 0 END) as wins,
    SUM(CASE WHEN (g.home_team_id = t.id AND g.home_score < g.away_score) OR
                   (g.away_team_id = t.id AND g.away_score < g.home_score) THEN 1 ELSE 0 END) as losses,
    SUM(CASE WHEN g.home_score = g.away_score THEN 1 ELSE 0 END) as ties,
    SUM(CASE WHEN g.home_team_id = t.id THEN g.home_score ELSE g.away_score END) as points_for,
    SUM(CASE WHEN g.home_team_id = t.id THEN g.away_score ELSE g.home_score END) as points_against,
    SUM(CASE WHEN g.home_team_id = t.id THEN g.home_score ELSE g.away_score END) -
    SUM(CASE WHEN g.home_team_id = t.id THEN g.away_score ELSE g.home_score END) as point_differential
FROM teams t
JOIN games g ON g.home_team_id = t.id OR g.away_team_id = t.id
WHERE g.status = 'completed'
  AND g.season = EXTRACT(YEAR FROM CURRENT_DATE)
GROUP BY t.id, t.name, t.abbreviation;

CREATE INDEX IF NOT EXISTS idx_vw_current_stats_team ON vw_current_season_team_stats(team_id);

-- ========================================
-- PART 6: COMMENTS FOR DOCUMENTATION
-- ========================================

COMMENT ON TABLE game_team_stats IS 'Team-level statistics for each game (box score data)';
COMMENT ON TABLE player_season_stats IS 'Player career statistics aggregated by season';
COMMENT ON TABLE team_standings IS 'Team standings and records by week/season';
COMMENT ON TABLE game_scoring_plays IS 'Timeline of scoring plays within games';
COMMENT ON TABLE advanced_stats IS 'Next Gen Stats and advanced analytics';
COMMENT ON TABLE game_broadcasts IS 'TV/streaming broadcast information';

COMMENT ON COLUMN games.weather_wind_dir IS 'Wind direction (N, NE, E, SE, S, SW, W, NW)';
COMMENT ON COLUMN games.weather_pressure IS 'Barometric pressure in millibars';
COMMENT ON COLUMN games.weather_visibility IS 'Visibility in miles';
COMMENT ON COLUMN games.weather_feels_like IS 'Feels-like temperature in Fahrenheit';
COMMENT ON COLUMN games.weather_precipitation IS 'Precipitation in inches';
COMMENT ON COLUMN games.is_day_game IS 'True if game played during daylight';

-- ========================================
-- MIGRATION COMPLETE
-- ========================================

-- Update schema version
INSERT INTO schema_migrations (version, description, applied_at)
VALUES ('003', 'Comprehensive schema enhancements for beautiful NFL API', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO NOTHING;