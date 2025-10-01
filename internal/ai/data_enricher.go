package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/models"
	"github.com/google/uuid"
)

// DataEnricher uses AI to enhance and complete data
type DataEnricher struct {
	aiService     *Service
	playerQueries *db.PlayerQueries
}

// NewDataEnricher creates a new AI data enricher
func NewDataEnricher(aiService *Service) *DataEnricher {
	return &DataEnricher{
		aiService:     aiService,
		playerQueries: &db.PlayerQueries{},
	}
}

// EnrichmentSuggestion represents an AI-suggested data enhancement
type EnrichmentSuggestion struct {
	EntityType   string                 `json:"entity_type"` // player, team, game
	EntityID     uuid.UUID              `json:"entity_id"`
	Field        string                 `json:"field"`
	CurrentValue interface{}            `json:"current_value"`
	SuggestedValue interface{}          `json:"suggested_value"`
	Confidence   float64                `json:"confidence"`
	Reasoning    string                 `json:"reasoning"`
	Sources      []string               `json:"sources"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// EnrichPlayer uses AI to enhance player data with missing information
func (de *DataEnricher) EnrichPlayer(ctx context.Context, player *models.Player) ([]EnrichmentSuggestion, error) {
	log.Printf("[DATA ENRICHER] Enriching player: %s", player.Name)

	// Build context about what we know
	knownData := map[string]interface{}{
		"name":     player.Name,
		"position": player.Position,
		"college":  player.College,
	}

	if player.DraftYear != nil {
		knownData["draft_year"] = *player.DraftYear
	}
	if player.HeightInches != nil {
		knownData["height_inches"] = *player.HeightInches
	}
	if player.WeightPounds != nil {
		knownData["weight_pounds"] = *player.WeightPounds
	}

	// Identify missing fields
	missingFields := de.identifyMissingFields(player)

	if len(missingFields) == 0 {
		return []EnrichmentSuggestion{}, nil
	}

	// Ask AI to suggest enrichments
	prompt := fmt.Sprintf(`You are an NFL data expert. Help enrich player data by suggesting accurate values for missing fields.

Player Name: %s
Position: %s
Known Data:
%s

Missing Fields: %v

For each missing field, suggest the most likely accurate value based on typical NFL data for similar players.
Consider position-specific typical values (e.g., QB height/weight vs RB).

Respond with ONLY valid JSON in this exact format:
{
  "suggestions": [
    {
      "field": "field_name",
      "suggested_value": "value",
      "confidence": 0.0-1.0,
      "reasoning": "why this value makes sense",
      "sources": ["source1", "source2"]
    }
  ]
}

Be conservative - only suggest values you're confident about (>0.7 confidence).
Respond with ONLY the JSON, no other text.`, player.Name, player.Position, toJSON(knownData), missingFields)

	response, provider, err := de.aiService.AnswerQuery(ctx, prompt, "Player data enrichment")
	if err != nil {
		return nil, fmt.Errorf("AI enrichment failed: %w", err)
	}

	log.Printf("[DATA ENRICHER] Enrichment from %s", provider)

	// Parse AI response
	var aiResponse struct {
		Suggestions []struct {
			Field          string   `json:"field"`
			SuggestedValue interface{} `json:"suggested_value"`
			Confidence     float64  `json:"confidence"`
			Reasoning      string   `json:"reasoning"`
			Sources        []string `json:"sources"`
		} `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Convert to enrichment suggestions
	suggestions := []EnrichmentSuggestion{}
	for _, s := range aiResponse.Suggestions {
		suggestions = append(suggestions, EnrichmentSuggestion{
			EntityType:     "player",
			EntityID:       player.ID,
			Field:          s.Field,
			CurrentValue:   nil,
			SuggestedValue: s.SuggestedValue,
			Confidence:     s.Confidence,
			Reasoning:      s.Reasoning,
			Sources:        s.Sources,
			Metadata: map[string]interface{}{
				"player_name": player.Name,
				"position":    player.Position,
			},
		})
	}

	return suggestions, nil
}

// identifyMissingFields finds which important fields are missing for a player
func (de *DataEnricher) identifyMissingFields(player *models.Player) []string {
	missing := []string{}

	if player.College == nil || *player.College == "" {
		missing = append(missing, "college")
	}
	if player.HeightInches == nil {
		missing = append(missing, "height_inches")
	}
	if player.WeightPounds == nil {
		missing = append(missing, "weight_pounds")
	}
	if player.DraftYear == nil {
		missing = append(missing, "draft_year")
	}
	if player.DraftRound == nil {
		missing = append(missing, "draft_round")
	}
	if player.DraftPick == nil {
		missing = append(missing, "draft_pick")
	}
	if player.BirthDate == nil {
		missing = append(missing, "birth_date")
	}

	return missing
}

// GeneratePlayerTags creates AI-generated tags for better searchability
func (de *DataEnricher) GeneratePlayerTags(ctx context.Context, player *models.Player, recentStats string) ([]string, error) {
	prompt := fmt.Sprintf(`You are an NFL analyst. Create descriptive tags for this player to improve searchability.

Player: %s
Position: %s
Recent Stats:
%s

Generate 5-10 relevant tags that describe:
- Playing style (e.g., "deep-threat", "dual-threat-qb", "power-runner")
- Strengths (e.g., "elite-speed", "strong-arm", "route-running")
- Fantasy relevance (e.g., "rb1", "flex-option", "waiver-pickup")
- Situation (e.g., "bellcow", "committee-back", "handcuff")

Respond with ONLY valid JSON:
{
  "tags": ["tag1", "tag2", "tag3"]
}

Keep tags lowercase, hyphenated, and descriptive.
Respond with ONLY the JSON, no other text.`, player.Name, player.Position, recentStats)

	response, _, err := de.aiService.AnswerQuery(ctx, prompt, "Player tag generation")
	if err != nil {
		return nil, err
	}

	var result struct {
		Tags []string `json:"tags"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	return result.Tags, nil
}

// GeneratePlayerSummary creates an AI-written summary of a player
func (de *DataEnricher) GeneratePlayerSummary(ctx context.Context, player *models.Player, seasonStats string) (string, error) {
	prompt := fmt.Sprintf(`Write a concise 2-3 sentence player summary for %s (%s).

Season Stats:
%s

Focus on:
- Current role and team situation
- Key strengths and playing style
- Fantasy football relevance

Write in present tense, factual tone. Be concise and informative.
Respond with ONLY the summary text, no JSON or extra formatting.`, player.Name, player.Position, seasonStats)

	summary, _, err := de.aiService.AnswerQuery(ctx, prompt, "Player summary generation")
	if err != nil {
		return "", err
	}

	return summary, nil
}

// SuggestRelatedPlayers finds similar players for comparison
func (de *DataEnricher) SuggestRelatedPlayers(ctx context.Context, player *models.Player) ([]string, error) {
	prompt := fmt.Sprintf(`You are an NFL analyst. Suggest 5 players who are comparable to %s (%s) for comparison purposes.

Consider:
- Same position
- Similar playing style
- Similar production level
- Useful for fantasy football comparisons

Respond with ONLY valid JSON:
{
  "related_players": [
    {"name": "Player Name", "reason": "why they're comparable"}
  ]
}

Respond with ONLY the JSON, no other text.`, player.Name, player.Position)

	response, _, err := de.aiService.AnswerQuery(ctx, prompt, "Related players suggestion")
	if err != nil {
		return nil, err
	}

	var result struct {
		RelatedPlayers []struct {
			Name   string `json:"name"`
			Reason string `json:"reason"`
		} `json:"related_players"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	names := []string{}
	for _, p := range result.RelatedPlayers {
		names = append(names, p.Name)
	}

	return names, nil
}

// toJSON helper
func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
