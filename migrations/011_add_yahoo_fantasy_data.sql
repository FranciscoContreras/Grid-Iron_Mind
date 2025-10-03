-- Migration 011: Add Yahoo Fantasy Sports Data Support
-- Adds tables for Yahoo Fantasy player rankings, projections, and ownership data

-- Yahoo Player Fantasy Rankings
-- Stores weekly fantasy rankings from Yahoo for different scoring formats
CREATE TABLE IF NOT EXISTS yahoo_player_rankings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    yahoo_player_key VARCHAR(50), -- Yahoo's unique player identifier (e.g., "nfl.p.12345")
    season INT NOT NULL,
    week INT NOT NULL,
    position VARCHAR(10) NOT NULL,
    overall_rank INT,
    position_rank INT,
    percent_owned DECIMAL(5,2), -- Ownership percentage (0-100)
    percent_started DECIMAL(5,2), -- Started percentage (0-100)
    average_draft_position DECIMAL(6,2), -- ADP for draft rankings
    scoring_format VARCHAR(20) DEFAULT 'standard', -- standard, ppr, half_ppr
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season, week, scoring_format)
);

-- Yahoo Player Projections
-- Weekly fantasy point projections from Yahoo
CREATE TABLE IF NOT EXISTS yahoo_player_projections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    yahoo_player_key VARCHAR(50),
    season INT NOT NULL,
    week INT NOT NULL,
    projected_points DECIMAL(6,2),
    projected_passing_yards INT,
    projected_passing_tds INT,
    projected_interceptions INT,
    projected_rushing_yards INT,
    projected_rushing_tds INT,
    projected_receptions INT,
    projected_receiving_yards INT,
    projected_receiving_tds INT,
    projected_fumbles_lost INT,
    projected_two_point_conversions INT,
    scoring_format VARCHAR(20) DEFAULT 'standard',
    confidence_score DECIMAL(4,2), -- Optional confidence metric
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season, week, scoring_format)
);

-- Yahoo Fantasy Leagues (Optional - for tracking specific leagues)
-- Stores information about Yahoo Fantasy leagues for analysis
CREATE TABLE IF NOT EXISTS yahoo_fantasy_leagues (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    yahoo_league_key VARCHAR(50) UNIQUE NOT NULL, -- e.g., "nfl.l.123456"
    yahoo_league_id VARCHAR(20) NOT NULL,
    name VARCHAR(255),
    season INT NOT NULL,
    num_teams INT,
    scoring_type VARCHAR(50), -- head2head, points
    draft_status VARCHAR(50),
    current_week INT,
    start_week INT,
    end_week INT,
    is_finished BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Yahoo Fantasy Player Matchup Advice
-- Stores start/sit advice and matchup ratings
CREATE TABLE IF NOT EXISTS yahoo_player_matchup_advice (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    opponent_team_id UUID REFERENCES teams(id),
    season INT NOT NULL,
    week INT NOT NULL,
    game_id UUID REFERENCES games(id),
    matchup_rating VARCHAR(20), -- favorable, average, unfavorable
    matchup_score DECIMAL(4,2), -- 1-10 rating
    start_sit_advice TEXT, -- Analysis text
    defensive_rank_vs_position INT, -- How well defense performs vs this position
    target_share_projection DECIMAL(5,2), -- Projected target share %
    snap_count_projection INT, -- Projected snaps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season, week)
);

-- Yahoo Player News and Notes
-- Stores player news, injury updates, and notes from Yahoo
CREATE TABLE IF NOT EXISTS yahoo_player_news (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    yahoo_player_key VARCHAR(50),
    headline TEXT NOT NULL,
    full_text TEXT,
    analysis TEXT,
    source VARCHAR(255),
    published_at TIMESTAMP NOT NULL,
    impact_level VARCHAR(20), -- high, medium, low
    news_type VARCHAR(50), -- injury, trade, suspension, performance, etc.
    is_breaking BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Yahoo Transaction Data
-- Tracks add/drop trends and waiver wire activity
CREATE TABLE IF NOT EXISTS yahoo_transaction_trends (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    player_id UUID REFERENCES players(id) ON DELETE CASCADE,
    season INT NOT NULL,
    week INT NOT NULL,
    adds_count INT DEFAULT 0,
    drops_count INT DEFAULT 0,
    waiver_adds INT DEFAULT 0,
    free_agent_adds INT DEFAULT 0,
    average_faab_bid DECIMAL(6,2), -- Average FAAB bid for player
    max_faab_bid DECIMAL(6,2), -- Highest FAAB bid
    trade_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(player_id, season, week)
);

-- Yahoo OAuth Tokens (for storing user tokens securely)
-- Stores encrypted OAuth tokens for Yahoo API access
CREATE TABLE IF NOT EXISTS yahoo_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_identifier VARCHAR(255), -- Could be email, user ID, or 'system' for server-to-server
    access_token TEXT NOT NULL, -- Encrypted OAuth access token
    refresh_token TEXT, -- Encrypted OAuth refresh token
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    scope TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_identifier)
);

-- Indexes for performance

-- Rankings indexes
CREATE INDEX IF NOT EXISTS idx_yahoo_rankings_player ON yahoo_player_rankings(player_id);
CREATE INDEX IF NOT EXISTS idx_yahoo_rankings_season_week ON yahoo_player_rankings(season, week);
CREATE INDEX IF NOT EXISTS idx_yahoo_rankings_position ON yahoo_player_rankings(position, season, week);
CREATE INDEX IF NOT EXISTS idx_yahoo_rankings_rank ON yahoo_player_rankings(overall_rank) WHERE overall_rank IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_yahoo_rankings_ownership ON yahoo_player_rankings(percent_owned DESC) WHERE percent_owned IS NOT NULL;

-- Projections indexes
CREATE INDEX IF NOT EXISTS idx_yahoo_projections_player ON yahoo_player_projections(player_id);
CREATE INDEX IF NOT EXISTS idx_yahoo_projections_season_week ON yahoo_player_projections(season, week);
CREATE INDEX IF NOT EXISTS idx_yahoo_projections_points ON yahoo_player_projections(projected_points DESC);

-- Matchup advice indexes
CREATE INDEX IF NOT EXISTS idx_yahoo_matchup_player ON yahoo_player_matchup_advice(player_id);
CREATE INDEX IF NOT EXISTS idx_yahoo_matchup_season_week ON yahoo_player_matchup_advice(season, week);
CREATE INDEX IF NOT EXISTS idx_yahoo_matchup_game ON yahoo_player_matchup_advice(game_id);
CREATE INDEX IF NOT EXISTS idx_yahoo_matchup_rating ON yahoo_player_matchup_advice(matchup_rating);

-- News indexes
CREATE INDEX IF NOT EXISTS idx_yahoo_news_player ON yahoo_player_news(player_id);
CREATE INDEX IF NOT EXISTS idx_yahoo_news_published ON yahoo_player_news(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_yahoo_news_impact ON yahoo_player_news(impact_level);
CREATE INDEX IF NOT EXISTS idx_yahoo_news_type ON yahoo_player_news(news_type);
CREATE INDEX IF NOT EXISTS idx_yahoo_news_breaking ON yahoo_player_news(is_breaking) WHERE is_breaking = TRUE;

-- Transaction trends indexes
CREATE INDEX IF NOT EXISTS idx_yahoo_transactions_player ON yahoo_transaction_trends(player_id);
CREATE INDEX IF NOT EXISTS idx_yahoo_transactions_season_week ON yahoo_transaction_trends(season, week);
CREATE INDEX IF NOT EXISTS idx_yahoo_transactions_adds ON yahoo_transaction_trends(adds_count DESC);
CREATE INDEX IF NOT EXISTS idx_yahoo_transactions_waiver ON yahoo_transaction_trends(waiver_adds DESC);

-- OAuth tokens indexes
CREATE INDEX IF NOT EXISTS idx_yahoo_oauth_user ON yahoo_oauth_tokens(user_identifier) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_yahoo_oauth_expires ON yahoo_oauth_tokens(expires_at);

-- Comments for documentation
COMMENT ON TABLE yahoo_player_rankings IS 'Yahoo Fantasy player rankings by week and position';
COMMENT ON TABLE yahoo_player_projections IS 'Yahoo Fantasy weekly player projections';
COMMENT ON TABLE yahoo_fantasy_leagues IS 'Yahoo Fantasy league metadata for tracking';
COMMENT ON TABLE yahoo_player_matchup_advice IS 'Start/sit advice and matchup analysis';
COMMENT ON TABLE yahoo_player_news IS 'Player news and injury updates from Yahoo';
COMMENT ON TABLE yahoo_transaction_trends IS 'Add/drop and waiver wire activity trends';
COMMENT ON TABLE yahoo_oauth_tokens IS 'Encrypted OAuth tokens for Yahoo API access';
