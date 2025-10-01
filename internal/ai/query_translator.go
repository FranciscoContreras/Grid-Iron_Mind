package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// QueryTranslator converts natural language queries to SQL
type QueryTranslator struct {
	aiService *Service
	schema    string
}

// NewQueryTranslator creates a new query translator
func NewQueryTranslator(aiService *Service) *QueryTranslator {
	schema := `
Available Tables:
- teams (id, nfl_id, name, abbreviation, city, conference, division, stadium)
- players (id, nfl_id, name, position, team_id, jersey_number, height_inches, weight_pounds, college, draft_year, status)
- games (id, espn_game_id, season_year, week, game_date, home_team_id, away_team_id, home_score, away_score, status)
- game_stats (id, game_id, player_id, passing_yards, passing_tds, rushing_yards, rushing_tds, receptions, receiving_yards, receiving_tds)
- player_injuries (id, player_id, team_id, injury_status, injury_type, injury_location, return_date)
- player_season_stats (player_id, season_year, games_played, passing_yards, rushing_yards, receiving_yards, total_tds)

Common Patterns:
- Use JOINs to get player names with stats
- Filter by season_year for recent data
- Use status='active' for current players
- Use status='final' for completed games
`

	return &QueryTranslator{
		aiService: aiService,
		schema:    schema,
	}
}

// TranslatedQuery represents a natural language query translated to SQL
type TranslatedQuery struct {
	NaturalQuery  string   `json:"natural_query"`
	SQL           string   `json:"sql"`
	Explanation   string   `json:"explanation"`
	SafetyLevel   string   `json:"safety_level"` // safe, review_required, unsafe
	Warnings      []string `json:"warnings"`
	AIProvider    string   `json:"ai_provider"`
	EstimatedRows int      `json:"estimated_rows,omitempty"`
}

// TranslateQuery converts natural language to SQL using AI
func (qt *QueryTranslator) TranslateQuery(ctx context.Context, naturalQuery string) (*TranslatedQuery, error) {
	log.Printf("[QUERY TRANSLATOR] Translating: %s", naturalQuery)

	prompt := fmt.Sprintf(`You are an expert SQL translator for an NFL database. Convert natural language queries to safe, optimized PostgreSQL queries.

Database Schema:
%s

User Query: "%s"

Generate a safe SQL query following these rules:
1. Only use SELECT statements (no INSERT, UPDATE, DELETE)
2. Include appropriate JOINs when needed
3. Add LIMIT clause (max 1000 rows)
4. Use proper indexes (id, team_id, player_id, season_year, week)
5. Include relevant WHERE clauses for performance
6. Return current/recent data by default (current season or last 30 days)

Respond with ONLY valid JSON in this exact format:
{
  "sql": "SELECT ... FROM ... WHERE ... LIMIT ...",
  "explanation": "brief explanation of what the query does",
  "safety_level": "safe|review_required|unsafe",
  "warnings": ["warning1 if any"],
  "estimated_rows": 100
}

If the query asks for something not in the schema, set safety_level to "unsafe" and explain why.
Respond with ONLY the JSON, no other text.`, qt.schema, naturalQuery)

	response, provider, err := qt.aiService.AnswerQuery(ctx, prompt, "SQL query translation")
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}

	var result struct {
		SQL           string   `json:"sql"`
		Explanation   string   `json:"explanation"`
		SafetyLevel   string   `json:"safety_level"`
		Warnings      []string `json:"warnings"`
		EstimatedRows int      `json:"estimated_rows"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Additional safety checks
	if qt.isUnsafe(result.SQL) {
		result.SafetyLevel = "unsafe"
		result.Warnings = append(result.Warnings, "Query contains potentially unsafe operations")
	}

	translatedQuery := &TranslatedQuery{
		NaturalQuery:  naturalQuery,
		SQL:           result.SQL,
		Explanation:   result.Explanation,
		SafetyLevel:   result.SafetyLevel,
		Warnings:      result.Warnings,
		AIProvider:    string(provider),
		EstimatedRows: result.EstimatedRows,
	}

	log.Printf("[QUERY TRANSLATOR] Translated successfully (safety: %s)", result.SafetyLevel)

	return translatedQuery, nil
}

// isUnsafe performs additional safety validation
func (qt *QueryTranslator) isUnsafe(sql string) bool {
	sqlLower := strings.ToLower(sql)

	// Block dangerous operations
	dangerousKeywords := []string{
		"drop ", "delete ", "update ", "insert ", "truncate ",
		"alter ", "create ", "grant ", "revoke ",
		"exec ", "execute ", "xp_", "sp_",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(sqlLower, keyword) {
			return true
		}
	}

	// Ensure it's a SELECT statement
	if !strings.HasPrefix(strings.TrimSpace(sqlLower), "select") {
		return true
	}

	return false
}

// SuggestOptimizations suggests query optimizations
func (qt *QueryTranslator) SuggestOptimizations(ctx context.Context, sql string) ([]string, error) {
	prompt := fmt.Sprintf(`You are a PostgreSQL performance expert. Analyze this query and suggest optimizations.

Query:
%s

Database Schema:
%s

Suggest optimizations for:
1. Index usage
2. JOIN efficiency
3. WHERE clause filtering
4. Result set size

Respond with ONLY valid JSON:
{
  "optimizations": ["optimization 1", "optimization 2"]
}

Respond with ONLY the JSON, no other text.`, sql, qt.schema)

	response, _, err := qt.aiService.AnswerQuery(ctx, prompt, "Query optimization")
	if err != nil {
		return nil, err
	}

	var result struct {
		Optimizations []string `json:"optimizations"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Optimizations, nil
}

// GenerateDataInsights creates insights from query results
func (qt *QueryTranslator) GenerateDataInsights(ctx context.Context, query string, results string, rowCount int) (string, error) {
	prompt := fmt.Sprintf(`You are an NFL data analyst. Generate insights from these query results.

User Query: %s
Number of Results: %d

Sample Results:
%s

Provide:
1. Key findings (2-3 sentences)
2. Interesting patterns or trends
3. Actionable insights for fantasy football or analysis

Be concise and focus on valuable insights.
Respond with plain text insights, no JSON.`, query, rowCount, results)

	insights, _, err := qt.aiService.AnswerQuery(ctx, prompt, "Data insights generation")
	if err != nil {
		return "", err
	}

	return insights, nil
}

// ExplainQueryResults explains query results in natural language
func (qt *QueryTranslator) ExplainQueryResults(ctx context.Context, naturalQuery string, resultCount int, sampleData string) (string, error) {
	prompt := fmt.Sprintf(`You asked: "%s"

I found %d results. Here's a sample:
%s

Explain these results in 2-3 sentences in a way that's easy to understand. Focus on what the data shows.
Respond with plain text explanation, no JSON.`, naturalQuery, resultCount, sampleData)

	explanation, _, err := qt.aiService.AnswerQuery(ctx, prompt, "Query explanation")
	if err != nil {
		return "", err
	}

	return explanation, nil
}
