package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	claudeAPIURL = "https://api.anthropic.com/v1/messages"
	modelName    = "claude-3-5-sonnet-20241022"
	maxTokens    = 4096
)

// Client handles Claude API requests
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request represents a Claude API request
type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents a Claude API response
type Response struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// SendMessage sends a message to Claude and returns the response
func (c *Client) SendMessage(ctx context.Context, prompt string) (string, error) {
	req := Request{
		Model:     modelName,
		MaxTokens: maxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Claude API returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp Response
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return claudeResp.Content[0].Text, nil
}

// PredictGameOutcome predicts the outcome of a game
func (c *Client) PredictGameOutcome(ctx context.Context, homeTeam, awayTeam string, homeStats, awayStats string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert NFL analyst. Analyze the upcoming game between %s (home) and %s (away).

Home Team Stats:
%s

Away Team Stats:
%s

Provide a detailed prediction including:
1. Predicted winner and score
2. Key factors that will determine the outcome
3. Player matchups to watch
4. Confidence level (0-100%%)

Format your response as JSON with the following structure:
{
  "winner": "team name",
  "predicted_score": {"home": 0, "away": 0},
  "confidence": 0,
  "key_factors": ["factor1", "factor2"],
  "analysis": "detailed analysis text"
}`, homeTeam, awayTeam, homeStats, awayStats)

	return c.SendMessage(ctx, prompt)
}

// PredictPlayerPerformance predicts a player's performance in an upcoming game
func (c *Client) PredictPlayerPerformance(ctx context.Context, playerName, position, opponent, recentStats string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert NFL fantasy analyst. Predict the performance of %s (%s) in their next game against %s.

Recent Performance:
%s

Provide a detailed prediction including:
1. Expected stats (yards, touchdowns, receptions, etc.)
2. Fantasy points projection
3. Key factors affecting performance
4. Confidence level (0-100%%)

Format your response as JSON with the following structure:
{
  "projected_stats": {"yards": 0, "touchdowns": 0, "receptions": 0},
  "fantasy_points": 0,
  "confidence": 0,
  "key_factors": ["factor1", "factor2"],
  "analysis": "detailed analysis text"
}`, playerName, position, opponent, recentStats)

	return c.SendMessage(ctx, prompt)
}

// AnalyzePlayer provides in-depth analysis of a player
func (c *Client) AnalyzePlayer(ctx context.Context, playerName, position, seasonStats, recentGames string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert NFL analyst. Provide a comprehensive analysis of %s (%s).

Season Stats:
%s

Recent Games:
%s

Provide analysis including:
1. Strengths and weaknesses
2. Performance trends
3. Comparison to position averages
4. Key insights for fantasy or betting

Format your response as JSON with the following structure:
{
  "strengths": ["strength1", "strength2"],
  "weaknesses": ["weakness1", "weakness2"],
  "trends": "trend analysis",
  "comparison": "position comparison",
  "insights": "key insights",
  "summary": "overall summary"
}`, playerName, position, seasonStats, recentGames)

	return c.SendMessage(ctx, prompt)
}

// AnswerQuery answers a general NFL question
func (c *Client) AnswerQuery(ctx context.Context, query, context string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert NFL analyst with access to real-time data. Answer the following question:

Question: %s

Relevant Data:
%s

Provide a clear, concise answer based on the data provided. If you need more data to answer accurately, mention what additional information would be helpful.`, query, context)

	return c.SendMessage(ctx, prompt)
}