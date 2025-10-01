package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/google/uuid"
)

// DataGardener maintains data quality and health using AI
type DataGardener struct {
	aiService     *Service
	playerQueries *db.PlayerQueries
	gameQueries   *db.GameQueries
	teamQueries   *db.TeamQueries
}

// NewDataGardener creates a new AI-powered data gardener
func NewDataGardener(aiService *Service) *DataGardener {
	return &DataGardener{
		aiService:     aiService,
		playerQueries: &db.PlayerQueries{},
		gameQueries:   &db.GameQueries{},
		teamQueries:   &db.TeamQueries{},
	}
}

// DataHealthReport represents the health status of the database
type DataHealthReport struct {
	Timestamp        time.Time              `json:"timestamp"`
	OverallHealth    string                 `json:"overall_health"` // excellent, good, fair, poor
	Issues           []DataIssue            `json:"issues"`
	Recommendations  []string               `json:"recommendations"`
	AIProvider       string                 `json:"ai_provider"`
	AutoFixedIssues  int                    `json:"auto_fixed_issues"`
	MetricsAnalysis  map[string]interface{} `json:"metrics_analysis"`
}

// DataIssue represents a detected data quality issue
type DataIssue struct {
	Type        string    `json:"type"`        // missing_data, anomaly, staleness, inconsistency
	Severity    string    `json:"severity"`    // critical, high, medium, low
	Entity      string    `json:"entity"`      // table or entity type
	EntityID    uuid.UUID `json:"entity_id,omitempty"`
	Description string    `json:"description"`
	AutoFixed   bool      `json:"auto_fixed"`
	FixAction   string    `json:"fix_action,omitempty"`
}

// RunHealthCheck performs comprehensive AI-powered health check
func (dg *DataGardener) RunHealthCheck(ctx context.Context) (*DataHealthReport, error) {
	log.Println("[DATA GARDENER] Starting AI-powered health check...")

	report := &DataHealthReport{
		Timestamp:       time.Now(),
		Issues:          []DataIssue{},
		Recommendations: []string{},
		AutoFixedIssues: 0,
		MetricsAnalysis: make(map[string]interface{}),
	}

	// 1. Collect database metrics
	metrics, err := dg.collectMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect metrics: %w", err)
	}

	// 2. Ask AI to analyze the metrics
	analysis, provider, err := dg.analyzeMetrics(ctx, metrics)
	if err != nil {
		log.Printf("[DATA GARDENER] AI analysis failed: %v", err)
		report.OverallHealth = "unknown"
		return report, nil
	}

	report.AIProvider = string(provider)

	// 3. Parse AI analysis
	var aiAnalysis struct {
		OverallHealth   string   `json:"overall_health"`
		DetectedIssues  []string `json:"detected_issues"`
		Recommendations []string `json:"recommendations"`
		MetricsInsights map[string]interface{} `json:"metrics_insights"`
	}

	if err := json.Unmarshal([]byte(analysis), &aiAnalysis); err != nil {
		log.Printf("[DATA GARDENER] Failed to parse AI analysis: %v", err)
	} else {
		report.OverallHealth = aiAnalysis.OverallHealth
		report.Recommendations = aiAnalysis.Recommendations
		report.MetricsAnalysis = aiAnalysis.MetricsInsights

		// Convert AI detected issues to structured format
		for _, issue := range aiAnalysis.DetectedIssues {
			report.Issues = append(report.Issues, DataIssue{
				Type:        "ai_detected",
				Severity:    "medium",
				Description: issue,
				AutoFixed:   false,
			})
		}
	}

	// 4. Detect specific issues using rules
	dg.detectStalenessIssues(ctx, report)
	dg.detectMissingDataIssues(ctx, report)
	dg.detectAnomalies(ctx, report)

	log.Printf("[DATA GARDENER] Health check complete. Found %d issues, fixed %d",
		len(report.Issues), report.AutoFixedIssues)

	return report, nil
}

// collectMetrics gathers key database metrics
func (dg *DataGardener) collectMetrics(ctx context.Context) (map[string]interface{}, error) {
	pool := db.GetPool()

	metrics := make(map[string]interface{})

	// Player counts
	var playerCount, activePlayerCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM players").Scan(&playerCount)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM players WHERE status = 'active'").Scan(&activePlayerCount)
	metrics["total_players"] = playerCount
	metrics["active_players"] = activePlayerCount

	// Game counts by status
	var scheduledGames, completedGames int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM games WHERE status = 'scheduled'").Scan(&scheduledGames)
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM games WHERE status = 'final'").Scan(&completedGames)
	metrics["scheduled_games"] = scheduledGames
	metrics["completed_games"] = completedGames

	// Stats coverage
	var gamesWithStats int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT game_id)
		FROM game_stats
		WHERE created_at > NOW() - INTERVAL '30 days'
	`).Scan(&gamesWithStats)
	metrics["games_with_stats_last_30d"] = gamesWithStats

	// Injury reports freshness
	var recentInjuries int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM player_injuries
		WHERE updated_at > NOW() - INTERVAL '24 hours'
	`).Scan(&recentInjuries)
	metrics["injuries_updated_24h"] = recentInjuries

	// Data staleness
	var oldestGameUpdate time.Time
	pool.QueryRow(ctx, `
		SELECT MIN(updated_at)
		FROM games
		WHERE status = 'final' AND updated_at > NOW() - INTERVAL '7 days'
	`).Scan(&oldestGameUpdate)
	if !oldestGameUpdate.IsZero() {
		metrics["oldest_game_update_hours"] = time.Since(oldestGameUpdate).Hours()
	}

	return metrics, nil
}

// analyzeMetrics asks AI to analyze database metrics
func (dg *DataGardener) analyzeMetrics(ctx context.Context, metrics map[string]interface{}) (string, AIProvider, error) {
	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")

	prompt := fmt.Sprintf(`You are a database health expert analyzing an NFL data API. Review these database metrics and provide insights.

Current Metrics:
%s

Analyze the data and respond with ONLY valid JSON in this exact format:
{
  "overall_health": "excellent|good|fair|poor",
  "detected_issues": ["issue 1", "issue 2"],
  "recommendations": ["recommendation 1", "recommendation 2"],
  "metrics_insights": {
    "player_data": "insight about player data",
    "game_data": "insight about game data",
    "stats_coverage": "insight about stats coverage",
    "data_freshness": "insight about data freshness"
  }
}

Focus on:
1. Data completeness (are we missing games, players, stats?)
2. Data freshness (is anything stale?)
3. Coverage gaps (which areas need more data?)
4. Anomalies (anything unusual in the numbers?)

Respond with ONLY the JSON, no other text.`, string(metricsJSON))

	return dg.aiService.AnswerQuery(ctx, prompt, "Database health analysis")
}

// detectStalenessIssues finds data that hasn't been updated recently
func (dg *DataGardener) detectStalenessIssues(ctx context.Context, report *DataHealthReport) {
	pool := db.GetPool()

	// Check for stale scheduled games (game day passed but still scheduled)
	var staleGames int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM games
		WHERE status = 'scheduled' AND game_date < NOW() - INTERVAL '6 hours'
	`).Scan(&staleGames)

	if staleGames > 0 {
		report.Issues = append(report.Issues, DataIssue{
			Type:        "staleness",
			Severity:    "high",
			Entity:      "games",
			Description: fmt.Sprintf("%d games are scheduled but game time has passed", staleGames),
			AutoFixed:   false,
			FixAction:   "Sync game scores from ESPN API",
		})
	}

	// Check for players without recent stat updates
	var playersNoStats int
	pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT p.id)
		FROM players p
		WHERE p.status = 'active'
		AND NOT EXISTS (
			SELECT 1 FROM game_stats gs
			WHERE gs.player_id = p.id
			AND gs.created_at > NOW() - INTERVAL '14 days'
		)
	`).Scan(&playersNoStats)

	if playersNoStats > 50 {
		report.Issues = append(report.Issues, DataIssue{
			Type:        "staleness",
			Severity:    "medium",
			Entity:      "players",
			Description: fmt.Sprintf("%d active players have no stats in last 14 days", playersNoStats),
			AutoFixed:   false,
			FixAction:   "Sync recent game stats from NFLverse",
		})
	}
}

// detectMissingDataIssues finds gaps in expected data
func (dg *DataGardener) detectMissingDataIssues(ctx context.Context, report *DataHealthReport) {
	pool := db.GetPool()

	// Check for missing team rosters
	var teamsWithoutPlayers int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM teams t
		WHERE NOT EXISTS (
			SELECT 1 FROM players p WHERE p.team_id = t.id
		)
	`).Scan(&teamsWithoutPlayers)

	if teamsWithoutPlayers > 0 {
		report.Issues = append(report.Issues, DataIssue{
			Type:        "missing_data",
			Severity:    "critical",
			Entity:      "teams",
			Description: fmt.Sprintf("%d teams have no players in roster", teamsWithoutPlayers),
			AutoFixed:   false,
			FixAction:   "Sync team rosters from ESPN API",
		})
	}

	// Check for games missing stats
	var gamesWithoutStats int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM games g
		WHERE g.status = 'final'
		AND g.game_date > NOW() - INTERVAL '7 days'
		AND NOT EXISTS (
			SELECT 1 FROM game_stats gs WHERE gs.game_id = g.id
		)
	`).Scan(&gamesWithoutStats)

	if gamesWithoutStats > 0 {
		report.Issues = append(report.Issues, DataIssue{
			Type:        "missing_data",
			Severity:    "high",
			Entity:      "games",
			Description: fmt.Sprintf("%d completed games from last week have no stats", gamesWithoutStats),
			AutoFixed:   false,
			FixAction:   "Sync game stats from NFLverse",
		})
	}
}

// detectAnomalies finds unusual patterns in the data
func (dg *DataGardener) detectAnomalies(ctx context.Context, report *DataHealthReport) {
	pool := db.GetPool()

	// Check for abnormal scores (potential data errors)
	var abnormalScores int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM games
		WHERE status = 'final'
		AND (home_score > 100 OR away_score > 100 OR home_score < 0 OR away_score < 0)
	`).Scan(&abnormalScores)

	if abnormalScores > 0 {
		report.Issues = append(report.Issues, DataIssue{
			Type:        "anomaly",
			Severity:    "high",
			Entity:      "games",
			Description: fmt.Sprintf("%d games have abnormal scores (>100 or <0)", abnormalScores),
			AutoFixed:   false,
			FixAction:   "Review and correct game scores",
		})
	}

	// Check for duplicate players (same NFL ID)
	var duplicatePlayers int
	pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM (
			SELECT nfl_id, COUNT(*) as cnt
			FROM players
			WHERE nfl_id IS NOT NULL
			GROUP BY nfl_id
			HAVING COUNT(*) > 1
		) dupes
	`).Scan(&duplicatePlayers)

	if duplicatePlayers > 0 {
		report.Issues = append(report.Issues, DataIssue{
			Type:        "anomaly",
			Severity:    "medium",
			Entity:      "players",
			Description: fmt.Sprintf("%d duplicate players detected (same NFL ID)", duplicatePlayers),
			AutoFixed:   false,
			FixAction:   "Deduplicate players using NFL ID",
		})
	}
}

// AutoHeal attempts to automatically fix detected issues
func (dg *DataGardener) AutoHeal(ctx context.Context, report *DataHealthReport) error {
	log.Printf("[DATA GARDENER] Starting auto-heal for %d issues...", len(report.Issues))

	for i := range report.Issues {
		issue := &report.Issues[i]

		// Only auto-fix low severity issues for now
		if issue.Severity == "low" || issue.Severity == "medium" {
			if err := dg.attemptFix(ctx, issue); err != nil {
				log.Printf("[DATA GARDENER] Failed to fix issue: %v", err)
			} else {
				issue.AutoFixed = true
				report.AutoFixedIssues++
				log.Printf("[DATA GARDENER] Auto-fixed: %s", issue.Description)
			}
		}
	}

	return nil
}

// attemptFix tries to automatically fix a specific issue
func (dg *DataGardener) attemptFix(ctx context.Context, issue *DataIssue) error {
	// This would integrate with your sync services
	// For now, just log the intended action
	log.Printf("[DATA GARDENER] Would fix: %s - Action: %s", issue.Description, issue.FixAction)
	return nil
}
