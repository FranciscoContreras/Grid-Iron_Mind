-- Migration 005: Add Defensive Statistics Tables
-- Purpose: Support defensive rankings and matchup analysis for fantasy football

-- Team Defensive Statistics (Season/Week Level)
CREATE TABLE IF NOT EXISTS team_defensive_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    season INTEGER NOT NULL,
    week INTEGER, -- NULL for season-long stats

    -- Points & Yards Allowed
    points_allowed INTEGER DEFAULT 0,
    points_allowed_per_game DECIMAL(5,2),
    yards_allowed INTEGER DEFAULT 0,
    yards_allowed_per_game DECIMAL(6,2),
    pass_yards_allowed INTEGER DEFAULT 0,
    pass_yards_allowed_per_game DECIMAL(6,2),
    rush_yards_allowed INTEGER DEFAULT 0,
    rush_yards_allowed_per_game DECIMAL(6,2),

    -- Defensive Plays
    sacks INTEGER DEFAULT 0,
    sacks_yards INTEGER DEFAULT 0,
    interceptions INTEGER DEFAULT 0,
    interception_yards INTEGER DEFAULT 0,
    interception_touchdowns INTEGER DEFAULT 0,
    forced_fumbles INTEGER DEFAULT 0,
    fumble_recoveries INTEGER DEFAULT 0,
    fumble_recovery_touchdowns INTEGER DEFAULT 0,

    -- Third Down & Red Zone Defense
    third_down_attempts INTEGER DEFAULT 0,
    third_down_conversions_allowed INTEGER DEFAULT 0,
    third_down_percentage DECIMAL(5,2),
    red_zone_attempts INTEGER DEFAULT 0,
    red_zone_touchdowns_allowed INTEGER DEFAULT 0,
    red_zone_percentage DECIMAL(5,2),

    -- Additional Metrics
    pass_attempts_allowed INTEGER DEFAULT 0,
    pass_completions_allowed INTEGER DEFAULT 0,
    pass_touchdowns_allowed INTEGER DEFAULT 0,
    rush_attempts_allowed INTEGER DEFAULT 0,
    rush_touchdowns_allowed INTEGER DEFAULT 0,
    penalties INTEGER DEFAULT 0,
    penalty_yards INTEGER DEFAULT 0,

    -- Rankings (calculated)
    defensive_rank INTEGER,
    pass_defense_rank INTEGER,
    rush_defense_rank INTEGER,
    points_allowed_rank INTEGER,

    games_played INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    -- Ensure unique combination of team/season/week
    UNIQUE(team_id, season, week)
);

-- Position-Specific Defensive Stats (e.g., vs QB, vs RB, vs WR)
CREATE TABLE IF NOT EXISTS team_defensive_stats_vs_position (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    season INTEGER NOT NULL,
    week INTEGER, -- NULL for season-long stats
    position VARCHAR(10) NOT NULL, -- QB, RB, WR, TE

    -- Fantasy Points Allowed
    fantasy_points_allowed_standard DECIMAL(6,2) DEFAULT 0,
    fantasy_points_allowed_ppr DECIMAL(6,2) DEFAULT 0,
    fantasy_points_allowed_half_ppr DECIMAL(6,2) DEFAULT 0,
    fantasy_points_per_game_standard DECIMAL(5,2),
    fantasy_points_per_game_ppr DECIMAL(5,2),
    fantasy_points_per_game_half_ppr DECIMAL(5,2),

    -- Position-Specific Stats
    passing_yards_allowed INTEGER DEFAULT 0, -- vs QB
    passing_tds_allowed INTEGER DEFAULT 0,
    interceptions_forced INTEGER DEFAULT 0,
    sacks_recorded INTEGER DEFAULT 0,

    rushing_yards_allowed INTEGER DEFAULT 0, -- vs RB/QB
    rushing_tds_allowed INTEGER DEFAULT 0,

    receptions_allowed INTEGER DEFAULT 0, -- vs WR/TE/RB
    receiving_yards_allowed INTEGER DEFAULT 0,
    receiving_tds_allowed INTEGER DEFAULT 0,

    -- Rankings
    rank_vs_position INTEGER,

    games_played INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(team_id, season, week, position)
);

-- Player vs Defense Historical Performance Cache
CREATE TABLE IF NOT EXISTS player_vs_defense_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    defense_team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    game_id UUID REFERENCES games(id) ON DELETE CASCADE,
    season INTEGER NOT NULL,
    week INTEGER NOT NULL,

    -- Game Stats (from game_stats table, cached for quick lookup)
    passing_yards INTEGER DEFAULT 0,
    passing_tds INTEGER DEFAULT 0,
    interceptions_thrown INTEGER DEFAULT 0,
    rushing_yards INTEGER DEFAULT 0,
    rushing_tds INTEGER DEFAULT 0,
    receptions INTEGER DEFAULT 0,
    receiving_yards INTEGER DEFAULT 0,
    receiving_tds INTEGER DEFAULT 0,

    -- Fantasy Points
    fantasy_points_standard DECIMAL(5,2),
    fantasy_points_ppr DECIMAL(5,2),
    fantasy_points_half_ppr DECIMAL(5,2),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(player_id, defense_team_id, game_id)
);

-- Indexes for Performance
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_team_season ON team_defensive_stats(team_id, season);
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_season_week ON team_defensive_stats(season, week);
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_rankings ON team_defensive_stats(season, defensive_rank) WHERE defensive_rank IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_defensive_vs_position_team_season ON team_defensive_stats_vs_position(team_id, season, position);
CREATE INDEX IF NOT EXISTS idx_defensive_vs_position_season_position ON team_defensive_stats_vs_position(season, position, rank_vs_position);

CREATE INDEX IF NOT EXISTS idx_player_vs_defense_player ON player_vs_defense_history(player_id, defense_team_id);
CREATE INDEX IF NOT EXISTS idx_player_vs_defense_season ON player_vs_defense_history(season, week);
CREATE INDEX IF NOT EXISTS idx_player_vs_defense_game ON player_vs_defense_history(game_id);

-- Function to calculate fantasy points
CREATE OR REPLACE FUNCTION calculate_fantasy_points(
    p_passing_yards INTEGER,
    p_passing_tds INTEGER,
    p_interceptions INTEGER,
    p_rushing_yards INTEGER,
    p_rushing_tds INTEGER,
    p_receptions INTEGER,
    p_receiving_yards INTEGER,
    p_receiving_tds INTEGER,
    p_scoring_type VARCHAR -- 'standard', 'ppr', 'half_ppr'
) RETURNS DECIMAL(5,2) AS $$
DECLARE
    points DECIMAL(5,2) := 0;
    reception_bonus DECIMAL(3,1) := 0;
BEGIN
    -- Set reception bonus based on scoring type
    IF p_scoring_type = 'ppr' THEN
        reception_bonus := 1.0;
    ELSIF p_scoring_type = 'half_ppr' THEN
        reception_bonus := 0.5;
    END IF;

    -- Passing: 0.04 per yard, 4 per TD, -2 per INT
    points := points + (COALESCE(p_passing_yards, 0) * 0.04);
    points := points + (COALESCE(p_passing_tds, 0) * 4);
    points := points - (COALESCE(p_interceptions, 0) * 2);

    -- Rushing: 0.1 per yard, 6 per TD
    points := points + (COALESCE(p_rushing_yards, 0) * 0.1);
    points := points + (COALESCE(p_rushing_tds, 0) * 6);

    -- Receiving: reception bonus + 0.1 per yard + 6 per TD
    points := points + (COALESCE(p_receptions, 0) * reception_bonus);
    points := points + (COALESCE(p_receiving_yards, 0) * 0.1);
    points := points + (COALESCE(p_receiving_tds, 0) * 6);

    RETURN points;
END;
$$ LANGUAGE plpgsql IMMUTABLE;
