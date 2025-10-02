-- Migration 007: Performance Indexes
-- Adds missing indexes based on query patterns and usage analysis

-- Players table indexes (frequently filtered columns)
CREATE INDEX IF NOT EXISTS idx_players_status ON players(status) WHERE status IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_players_position_status ON players(position, status);
CREATE INDEX IF NOT EXISTS idx_players_team_position ON players(team_id, position) WHERE team_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_players_name_lower ON players(LOWER(name));

-- Games table indexes (season/week filtering is very common)
CREATE INDEX IF NOT EXISTS idx_games_home_team ON games(home_team_id);
CREATE INDEX IF NOT EXISTS idx_games_away_team ON games(away_team_id);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_season ON games(season);
CREATE INDEX IF NOT EXISTS idx_games_season_week_date ON games(season, week, game_date);

-- Game stats table indexes (player/game lookups)
CREATE INDEX IF NOT EXISTS idx_game_stats_player_game ON game_stats(player_id, game_id);
CREATE INDEX IF NOT EXISTS idx_game_stats_season_week ON game_stats(season, week);
CREATE INDEX IF NOT EXISTS idx_game_stats_player_season ON game_stats(player_id, season);

-- Player career stats indexes
CREATE INDEX IF NOT EXISTS idx_player_career_stats_player ON player_career_stats(player_id);
CREATE INDEX IF NOT EXISTS idx_player_career_stats_player_season ON player_career_stats(player_id, season);
CREATE INDEX IF NOT EXISTS idx_player_career_stats_season ON player_career_stats(season);

-- Player team history indexes
CREATE INDEX IF NOT EXISTS idx_player_team_history_player ON player_team_history(player_id);
CREATE INDEX IF NOT EXISTS idx_player_team_history_team ON player_team_history(team_id);
CREATE INDEX IF NOT EXISTS idx_player_team_history_player_dates ON player_team_history(player_id, start_date, end_date);

-- Game team stats indexes
CREATE INDEX IF NOT EXISTS idx_game_team_stats_game ON game_team_stats(game_id);
CREATE INDEX IF NOT EXISTS idx_game_team_stats_team ON game_team_stats(team_id);
CREATE INDEX IF NOT EXISTS idx_game_team_stats_team_season ON game_team_stats(team_id, season);

-- Player injuries indexes
CREATE INDEX IF NOT EXISTS idx_player_injuries_player ON player_injuries(player_id);
CREATE INDEX IF NOT EXISTS idx_player_injuries_team ON player_injuries(team_id);
CREATE INDEX IF NOT EXISTS idx_player_injuries_status ON player_injuries(status) WHERE status IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_player_injuries_player_status ON player_injuries(player_id, status);

-- Defensive stats indexes
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_team ON team_defensive_stats(team_id);
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_season ON team_defensive_stats(season);
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_team_season ON team_defensive_stats(team_id, season);
CREATE INDEX IF NOT EXISTS idx_team_defensive_stats_season_week ON team_defensive_stats(season, week);

-- Defensive stats vs position indexes
CREATE INDEX IF NOT EXISTS idx_def_stats_vs_pos_team ON team_defensive_stats_vs_position(team_id);
CREATE INDEX IF NOT EXISTS idx_def_stats_vs_pos_season ON team_defensive_stats_vs_position(season);
CREATE INDEX IF NOT EXISTS idx_def_stats_vs_pos_position ON team_defensive_stats_vs_position(position);
CREATE INDEX IF NOT EXISTS idx_def_stats_vs_pos_team_season_pos ON team_defensive_stats_vs_position(team_id, season, position);

-- Player vs defense history indexes
CREATE INDEX IF NOT EXISTS idx_player_vs_def_player ON player_vs_defense_history(player_id);
CREATE INDEX IF NOT EXISTS idx_player_vs_def_defense ON player_vs_defense_history(defense_team_id);
CREATE INDEX IF NOT EXISTS idx_player_vs_def_season ON player_vs_defense_history(season);
CREATE INDEX IF NOT EXISTS idx_player_vs_def_player_defense ON player_vs_defense_history(player_id, defense_team_id);

-- Covering indexes for common queries (includes frequently selected columns)
CREATE INDEX IF NOT EXISTS idx_players_team_covering ON players(team_id)
    INCLUDE (name, position, jersey_number, status) WHERE team_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_games_season_week_covering ON games(season, week)
    INCLUDE (game_date, home_team_id, away_team_id, home_score, away_score, status);

CREATE INDEX IF NOT EXISTS idx_game_stats_player_covering ON game_stats(player_id)
    INCLUDE (game_id, season, week, passing_yards, rushing_yards, receiving_yards, touchdowns);

-- Partial indexes for active data (improve performance by excluding archived data)
CREATE INDEX IF NOT EXISTS idx_players_active ON players(team_id, position)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_games_recent_season ON games(season, week, game_date)
    WHERE season >= 2020;

CREATE INDEX IF NOT EXISTS idx_injuries_current ON player_injuries(player_id, status)
    WHERE status IN ('out', 'questionable', 'doubtful');

-- Text search indexes for name lookups
CREATE INDEX IF NOT EXISTS idx_players_name_trgm ON players USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_teams_name_trgm ON teams USING gin(name gin_trgm_ops);

-- Add trigram extension if not exists
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Analyze tables to update statistics for query planner
ANALYZE teams;
ANALYZE players;
ANALYZE games;
ANALYZE game_stats;
ANALYZE player_career_stats;
ANALYZE player_team_history;
ANALYZE game_team_stats;
ANALYZE player_injuries;
ANALYZE team_defensive_stats;
ANALYZE team_defensive_stats_vs_position;
ANALYZE player_vs_defense_history;
