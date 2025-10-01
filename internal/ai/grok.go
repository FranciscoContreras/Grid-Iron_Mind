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
	grokAPIURL   = "https://api.x.ai/v1/chat/completions"
	grokModel    = "grok-beta"
	grokMaxTokens = 4096
)

// GrokClient handles Grok AI API requests
type GrokClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewGrokClient creates a new Grok AI client
func NewGrokClient(apiKey string) *GrokClient {
	return &GrokClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GrokRequest represents a Grok API request
type GrokRequest struct {
	Model       string         `json:"model"`
	Messages    []GrokMessage  `json:"messages"`
	Stream      bool           `json:"stream"`
	Temperature float64        `json:"temperature"`
}

// GrokMessage represents a chat message for Grok
type GrokMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GrokResponse represents a Grok API response
type GrokResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// SendMessage sends a message to Grok and returns the response
func (g *GrokClient) SendMessage(ctx context.Context, prompt string) (string, error) {
	return g.SendMessageWithSystem(ctx, "You are an expert NFL analyst providing accurate, data-driven insights.", prompt)
}

// SendMessageWithSystem sends a message with a custom system prompt
func (g *GrokClient) SendMessageWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	req := GrokRequest{
		Model: grokModel,
		Messages: []GrokMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Stream:      false,
		Temperature: 0,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", grokAPIURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.apiKey))

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Grok API returned status %d: %s", resp.StatusCode, string(body))
	}

	var grokResp GrokResponse
	if err := json.Unmarshal(body, &grokResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(grokResp.Choices) == 0 {
		return "", fmt.Errorf("empty response from Grok")
	}

	return grokResp.Choices[0].Message.Content, nil
}

// PredictGameOutcome predicts the outcome of a game using Grok
func (g *GrokClient) PredictGameOutcome(ctx context.Context, homeTeam, awayTeam string, homeStats, awayStats string) (string, error) {
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

	return g.SendMessage(ctx, prompt)
}

// PredictPlayerPerformance predicts a player's performance in an upcoming game
func (g *GrokClient) PredictPlayerPerformance(ctx context.Context, playerName, position, opponent, recentStats string) (string, error) {
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

	return g.SendMessage(ctx, prompt)
}

// AnalyzePlayer provides in-depth analysis of a player
func (g *GrokClient) AnalyzePlayer(ctx context.Context, playerName, position, seasonStats, recentGames string) (string, error) {
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

	return g.SendMessage(ctx, prompt)
}

// AnswerQuery answers a general NFL question
func (g *GrokClient) AnswerQuery(ctx context.Context, query, context string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert NFL analyst with access to real-time data. Answer the following question:

Question: %s

Relevant Data:
%s

Provide a clear, concise answer based on the data provided. If you need more data to answer accurately, mention what additional information would be helpful.`, query, context)

	return g.SendMessage(ctx, prompt)
}
