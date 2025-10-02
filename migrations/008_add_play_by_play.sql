-- Migration 008: Add Play-by-Play Data
-- This migration adds tables for storing detailed play-by-play data from NFLverse

-- Play-by-Play table for detailed game events
CREATE TABLE IF NOT EXISTS play_by_play (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Game/Play Identifiers
    play_id VARCHAR(50) NOT NULL,
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    nfl_game_id VARCHAR(50),

    -- Teams
    home_team_id UUID REFERENCES teams(id),
    away_team_id UUID REFERENCES teams(id),
    possession_team_id UUID REFERENCES teams(id),
    defensive_team_id UUID REFERENCES teams(id),

    -- Game Context
    season INT NOT NULL,
    week INT NOT NULL,
    quarter INT,
    down INT,
    yards_to_go INT,
    yard_line INT, -- yards from opponent's end zone (100 = own goal line)
    game_seconds_remaining INT,

    -- Play Information
    play_type VARCHAR(50), -- pass, run, punt, field_goal, kickoff, etc.
    play_type_nfl VARCHAR(50), -- NFL's official play type
    description TEXT,
    yards_gained INT,

    -- Players Involved
    passer_player_id UUID REFERENCES players(id),
    receiver_player_id UUID REFERENCES players(id),
    rusher_player_id UUID REFERENCES players(id),

    -- Pass Details
    pass_length VARCHAR(20), -- short, deep
    pass_location VARCHAR(20), -- left, middle, right
    air_yards DECIMAL(5,2),
    yards_after_catch DECIMAL(5,2),

    -- Run Details
    run_location VARCHAR(20), -- left, middle, right
    run_gap VARCHAR(20), -- guard, tackle, end

    -- Advanced Metrics
    epa DECIMAL(8,4), -- Expected Points Added
    wpa DECIMAL(8,4), -- Win Probability Added
    success_play INT, -- 1 if successful, 0 if not

    -- Play Outcomes
    first_down INT DEFAULT 0, -- 1 if resulted in first down
    touchdown INT DEFAULT 0,
    pass_touchdown INT DEFAULT 0,
    rush_touchdown INT DEFAULT 0,
    interception INT DEFAULT 0,
    fumble INT DEFAULT 0,
    completed_pass INT DEFAULT 0,
    sack INT DEFAULT 0,
    penalty INT DEFAULT 0,

    -- Scores After Play
    possession_team_score INT,
    defensive_team_score INT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(play_id, game_id)
);

-- Indexes for common queries
CREATE INDEX idx_pbp_game ON play_by_play(game_id);
CREATE INDEX idx_pbp_season_week ON play_by_play(season, week);
CREATE INDEX idx_pbp_passer ON play_by_play(passer_player_id) WHERE passer_player_id IS NOT NULL;
CREATE INDEX idx_pbp_receiver ON play_by_play(receiver_player_id) WHERE receiver_player_id IS NOT NULL;
CREATE INDEX idx_pbp_rusher ON play_by_play(rusher_player_id) WHERE rusher_player_id IS NOT NULL;
CREATE INDEX idx_pbp_play_type ON play_by_play(play_type);
CREATE INDEX idx_pbp_touchdown ON play_by_play(touchdown) WHERE touchdown = 1;
CREATE INDEX idx_pbp_first_down ON play_by_play(first_down) WHERE first_down = 1;

-- Comments
COMMENT ON TABLE play_by_play IS 'Detailed play-by-play data for every play in every game';
COMMENT ON COLUMN play_by_play.yard_line IS 'Yards from opponent end zone (100 = own goal line, 50 = midfield)';
COMMENT ON COLUMN play_by_play.epa IS 'Expected Points Added - change in expected points from start to end of play';
COMMENT ON COLUMN play_by_play.wpa IS 'Win Probability Added - change in win probability from start to end of play';
COMMENT ON COLUMN play_by_play.success_play IS 'Binary indicator if play was successful (gained significant yards relative to down/distance)';

-- Materialized view for play-by-play summary by game
CREATE MATERIALIZED VIEW IF NOT EXISTS game_play_summary AS
SELECT
    game_id,
    COUNT(*) as total_plays,
    SUM(CASE WHEN play_type = 'pass' THEN 1 ELSE 0 END) as pass_plays,
    SUM(CASE WHEN play_type = 'run' THEN 1 ELSE 0 END) as run_plays,
    SUM(CASE WHEN touchdown = 1 THEN 1 ELSE 0 END) as touchdowns,
    SUM(CASE WHEN interception = 1 THEN 1 ELSE 0 END) as interceptions,
    SUM(CASE WHEN fumble = 1 THEN 1 ELSE 0 END) as fumbles,
    SUM(CASE WHEN sack = 1 THEN 1 ELSE 0 END) as sacks,
    SUM(CASE WHEN penalty = 1 THEN 1 ELSE 0 END) as penalties,
    AVG(epa) as avg_epa_per_play,
    SUM(yards_gained) as total_yards
FROM play_by_play
GROUP BY game_id;

CREATE UNIQUE INDEX idx_game_play_summary_game ON game_play_summary(game_id);

-- Materialized view for player play-by-play stats
CREATE MATERIALIZED VIEW IF NOT EXISTS player_play_stats AS
SELECT
    passer_player_id as player_id,
    season,
    week,
    'passing' as stat_type,
    COUNT(*) as plays,
    SUM(completed_pass) as completions,
    SUM(yards_gained) as yards,
    SUM(pass_touchdown) as touchdowns,
    SUM(interception) as interceptions,
    AVG(epa) as avg_epa,
    SUM(air_yards) as total_air_yards,
    SUM(yards_after_catch) as total_yac
FROM play_by_play
WHERE passer_player_id IS NOT NULL
GROUP BY passer_player_id, season, week

UNION ALL

SELECT
    rusher_player_id as player_id,
    season,
    week,
    'rushing' as stat_type,
    COUNT(*) as plays,
    NULL as completions,
    SUM(yards_gained) as yards,
    SUM(rush_touchdown) as touchdowns,
    NULL as interceptions,
    AVG(epa) as avg_epa,
    NULL as total_air_yards,
    NULL as total_yac
FROM play_by_play
WHERE rusher_player_id IS NOT NULL
GROUP BY rusher_player_id, season, week

UNION ALL

SELECT
    receiver_player_id as player_id,
    season,
    week,
    'receiving' as stat_type,
    COUNT(*) as plays,
    SUM(completed_pass) as completions,
    SUM(yards_gained) as yards,
    SUM(pass_touchdown) as touchdowns,
    NULL as interceptions,
    AVG(epa) as avg_epa,
    SUM(air_yards) as total_air_yards,
    SUM(yards_after_catch) as total_yac
FROM play_by_play
WHERE receiver_player_id IS NOT NULL
GROUP BY receiver_player_id, season, week;

CREATE INDEX idx_player_play_stats_player ON player_play_stats(player_id);
CREATE INDEX idx_player_play_stats_season ON player_play_stats(season, week);
CREATE INDEX idx_player_play_stats_type ON player_play_stats(stat_type);
