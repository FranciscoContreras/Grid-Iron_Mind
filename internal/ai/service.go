package ai

import (
	"context"
	"fmt"
	"log"
)

// AIProvider represents an AI provider type
type AIProvider string

const (
	ProviderClaude AIProvider = "claude"
	ProviderGrok   AIProvider = "grok"
)

// Service manages multiple AI providers with automatic fallback
type Service struct {
	claudeClient *Client
	grokClient   *GrokClient
	primaryProvider AIProvider
}

// NewService creates a new AI service with fallback support
func NewService(claudeAPIKey, grokAPIKey string) *Service {
	var claudeClient *Client
	var grokClient *GrokClient
	var primaryProvider AIProvider

	// Initialize available clients
	if claudeAPIKey != "" {
		claudeClient = NewClient(claudeAPIKey)
		primaryProvider = ProviderClaude
		log.Println("Claude AI client initialized (primary)")
	}

	if grokAPIKey != "" {
		grokClient = NewGrokClient(grokAPIKey)
		if primaryProvider == "" {
			primaryProvider = ProviderGrok
			log.Println("Grok AI client initialized (primary)")
		} else {
			log.Println("Grok AI client initialized (fallback)")
		}
	}

	if claudeClient == nil && grokClient == nil {
		log.Println("Warning: No AI providers configured")
	}

	return &Service{
		claudeClient:    claudeClient,
		grokClient:      grokClient,
		primaryProvider: primaryProvider,
	}
}

// IsAvailable returns true if at least one AI provider is configured
func (s *Service) IsAvailable() bool {
	return s.claudeClient != nil || s.grokClient != nil
}

// GetProvider returns the name of the primary provider
func (s *Service) GetProvider() string {
	return string(s.primaryProvider)
}

// sendMessageWithFallback sends a message using primary provider, with automatic fallback
func (s *Service) sendMessageWithFallback(ctx context.Context, fn func(provider AIProvider) (string, error)) (string, AIProvider, error) {
	if !s.IsAvailable() {
		return "", "", fmt.Errorf("no AI providers configured")
	}

	// Try primary provider first
	result, err := fn(s.primaryProvider)
	if err == nil {
		return result, s.primaryProvider, nil
	}

	log.Printf("Primary AI provider (%s) failed: %v, attempting fallback...", s.primaryProvider, err)

	// Determine fallback provider
	var fallbackProvider AIProvider
	if s.primaryProvider == ProviderClaude && s.grokClient != nil {
		fallbackProvider = ProviderGrok
	} else if s.primaryProvider == ProviderGrok && s.claudeClient != nil {
		fallbackProvider = ProviderClaude
	} else {
		return "", s.primaryProvider, fmt.Errorf("primary provider failed and no fallback available: %w", err)
	}

	// Try fallback provider
	result, fallbackErr := fn(fallbackProvider)
	if fallbackErr == nil {
		log.Printf("Fallback AI provider (%s) succeeded", fallbackProvider)
		return result, fallbackProvider, nil
	}

	log.Printf("Fallback AI provider (%s) also failed: %v", fallbackProvider, fallbackErr)
	return "", fallbackProvider, fmt.Errorf("all AI providers failed - primary: %w, fallback: %v", err, fallbackErr)
}

// PredictGameOutcome predicts game outcome with automatic fallback
func (s *Service) PredictGameOutcome(ctx context.Context, homeTeam, awayTeam, homeStats, awayStats string) (string, AIProvider, error) {
	return s.sendMessageWithFallback(ctx, func(provider AIProvider) (string, error) {
		switch provider {
		case ProviderClaude:
			if s.claudeClient == nil {
				return "", fmt.Errorf("Claude client not available")
			}
			return s.claudeClient.PredictGameOutcome(ctx, homeTeam, awayTeam, homeStats, awayStats)
		case ProviderGrok:
			if s.grokClient == nil {
				return "", fmt.Errorf("Grok client not available")
			}
			return s.grokClient.PredictGameOutcome(ctx, homeTeam, awayTeam, homeStats, awayStats)
		default:
			return "", fmt.Errorf("unknown provider: %s", provider)
		}
	})
}

// PredictPlayerPerformance predicts player performance with automatic fallback
func (s *Service) PredictPlayerPerformance(ctx context.Context, playerName, position, opponent, recentStats string) (string, AIProvider, error) {
	return s.sendMessageWithFallback(ctx, func(provider AIProvider) (string, error) {
		switch provider {
		case ProviderClaude:
			if s.claudeClient == nil {
				return "", fmt.Errorf("Claude client not available")
			}
			return s.claudeClient.PredictPlayerPerformance(ctx, playerName, position, opponent, recentStats)
		case ProviderGrok:
			if s.grokClient == nil {
				return "", fmt.Errorf("Grok client not available")
			}
			return s.grokClient.PredictPlayerPerformance(ctx, playerName, position, opponent, recentStats)
		default:
			return "", fmt.Errorf("unknown provider: %s", provider)
		}
	})
}

// AnalyzePlayer analyzes player with automatic fallback
func (s *Service) AnalyzePlayer(ctx context.Context, playerName, position, seasonStats, recentGames string) (string, AIProvider, error) {
	return s.sendMessageWithFallback(ctx, func(provider AIProvider) (string, error) {
		switch provider {
		case ProviderClaude:
			if s.claudeClient == nil {
				return "", fmt.Errorf("Claude client not available")
			}
			return s.claudeClient.AnalyzePlayer(ctx, playerName, position, seasonStats, recentGames)
		case ProviderGrok:
			if s.grokClient == nil {
				return "", fmt.Errorf("Grok client not available")
			}
			return s.grokClient.AnalyzePlayer(ctx, playerName, position, seasonStats, recentGames)
		default:
			return "", fmt.Errorf("unknown provider: %s", provider)
		}
	})
}

// AnswerQuery answers query with automatic fallback
func (s *Service) AnswerQuery(ctx context.Context, query, contextData string) (string, AIProvider, error) {
	return s.sendMessageWithFallback(ctx, func(provider AIProvider) (string, error) {
		switch provider {
		case ProviderClaude:
			if s.claudeClient == nil {
				return "", fmt.Errorf("Claude client not available")
			}
			return s.claudeClient.AnswerQuery(ctx, query, contextData)
		case ProviderGrok:
			if s.grokClient == nil {
				return "", fmt.Errorf("Grok client not available")
			}
			return s.grokClient.AnswerQuery(ctx, query, contextData)
		default:
			return "", fmt.Errorf("unknown provider: %s", provider)
		}
	})
}
