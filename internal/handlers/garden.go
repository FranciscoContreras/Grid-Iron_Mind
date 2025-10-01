package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/francisco/gridironmind/internal/ai"
	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

// GardenHandler handles AI Data Garden endpoints
type GardenHandler struct {
	aiService     *ai.Service
	gardener      *ai.DataGardener
	enricher      *ai.DataEnricher
	translator    *ai.QueryTranslator
	scheduler     *ai.SyncScheduler
	playerQueries *db.PlayerQueries
}

// NewGardenHandler creates a new garden handler
func NewGardenHandler(cfg *config.Config) *GardenHandler {
	aiService := ai.NewService(cfg.ClaudeAPIKey, cfg.GrokAPIKey)

	return &GardenHandler{
		aiService:     aiService,
		gardener:      ai.NewDataGardener(aiService),
		enricher:      ai.NewDataEnricher(aiService),
		translator:    ai.NewQueryTranslator(aiService),
		scheduler:     ai.NewSyncScheduler(aiService),
		playerQueries: &db.PlayerQueries{},
	}
}

// HandleGarden routes garden requests
func (h *GardenHandler) HandleGarden(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/garden")

	switch {
	case path == "/health" && r.Method == http.MethodGet:
		h.handleHealthCheck(w, r)
	case path == "/health" && r.Method == http.MethodPost:
		h.handleHealthCheckWithHeal(w, r)
	case path == "/query" && r.Method == http.MethodPost:
		h.handleNaturalQuery(w, r)
	case strings.HasPrefix(path, "/enrich/player/"):
		h.handleEnrichPlayer(w, r)
	case path == "/schedule" && r.Method == http.MethodGet:
		h.handleSyncSchedule(w, r)
	case path == "/status" && r.Method == http.MethodGet:
		h.handleGardenStatus(w, r)
	default:
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "Garden endpoint not found")
	}
}

// handleHealthCheck performs AI-powered health check
func (h *GardenHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if !h.aiService.IsAvailable() {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	log.Println("[GARDEN] Running health check...")

	report, err := h.gardener.RunHealthCheck(r.Context())
	if err != nil {
		log.Printf("[GARDEN] Health check failed: %v", err)
		response.InternalError(w, "Health check failed")
		return
	}

	log.Printf("[GARDEN] Health check complete: %s (%d issues)",
		report.OverallHealth, len(report.Issues))

	response.Success(w, map[string]interface{}{
		"health_report": report,
		"message":       fmt.Sprintf("Health: %s with %d issues detected", report.OverallHealth, len(report.Issues)),
	})
}

// handleHealthCheckWithHeal performs health check and auto-heals issues
func (h *GardenHandler) handleHealthCheckWithHeal(w http.ResponseWriter, r *http.Request) {
	if !h.aiService.IsAvailable() {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	log.Println("[GARDEN] Running health check with auto-heal...")

	report, err := h.gardener.RunHealthCheck(r.Context())
	if err != nil {
		log.Printf("[GARDEN] Health check failed: %v", err)
		response.InternalError(w, "Health check failed")
		return
	}

	// Attempt auto-healing
	if err := h.gardener.AutoHeal(r.Context(), report); err != nil {
		log.Printf("[GARDEN] Auto-heal failed: %v", err)
	}

	log.Printf("[GARDEN] Health check complete: %s (%d issues, %d auto-fixed)",
		report.OverallHealth, len(report.Issues), report.AutoFixedIssues)

	response.Success(w, map[string]interface{}{
		"health_report": report,
		"message":       fmt.Sprintf("Health: %s - Fixed %d/%d issues", report.OverallHealth, report.AutoFixedIssues, len(report.Issues)),
	})
}

// handleNaturalQuery translates and executes natural language queries
func (h *GardenHandler) handleNaturalQuery(w http.ResponseWriter, r *http.Request) {
	if !h.aiService.IsAvailable() {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	var reqBody struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if reqBody.Query == "" {
		response.BadRequest(w, "Query is required")
		return
	}

	log.Printf("[GARDEN] Natural query: %s", reqBody.Query)

	// Translate query
	translated, err := h.translator.TranslateQuery(r.Context(), reqBody.Query)
	if err != nil {
		log.Printf("[GARDEN] Query translation failed: %v", err)
		response.InternalError(w, "Failed to translate query")
		return
	}

	// Check safety
	if translated.SafetyLevel == "unsafe" {
		response.Error(w, http.StatusBadRequest, "UNSAFE_QUERY", "Query contains unsafe operations")
		return
	}

	if translated.SafetyLevel == "review_required" {
		response.Error(w, http.StatusBadRequest, "QUERY_NEEDS_REVIEW",
			fmt.Sprintf("Query needs review: %v", translated.Warnings))
		return
	}

	// Execute SQL
	pool := db.GetPool()
	rows, err := pool.Query(r.Context(), translated.SQL)
	if err != nil {
		log.Printf("[GARDEN] Query execution failed: %v", err)
		response.InternalError(w, "Failed to execute query")
		return
	}
	defer rows.Close()

	// Parse results
	results := []map[string]interface{}{}
	columns := rows.FieldDescriptions()

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("[GARDEN] Row scan failed: %v", err)
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[string(col.Name)] = values[i]
		}
		results = append(results, row)
	}

	// Generate insights from results
	sampleJSON, _ := json.MarshalIndent(results[:min(5, len(results))], "", "  ")
	insights, err := h.translator.GenerateDataInsights(r.Context(),
		reqBody.Query,
		string(sampleJSON),
		len(results))

	if err != nil {
		log.Printf("[GARDEN] Insights generation failed: %v", err)
		insights = "Unable to generate insights"
	}

	log.Printf("[GARDEN] Query executed: %d results", len(results))

	response.Success(w, map[string]interface{}{
		"query":       reqBody.Query,
		"sql":         translated.SQL,
		"explanation": translated.Explanation,
		"results":     results,
		"count":       len(results),
		"insights":    insights,
		"ai_provider": translated.AIProvider,
		"warnings":    translated.Warnings,
	})
}

// handleEnrichPlayer enriches a player with AI-generated data
func (h *GardenHandler) handleEnrichPlayer(w http.ResponseWriter, r *http.Request) {
	if !h.aiService.IsAvailable() {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST allowed")
		return
	}

	// Extract player ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/garden/enrich/player/")
	playerID, err := uuid.Parse(path)
	if err != nil {
		response.BadRequest(w, "Invalid player ID")
		return
	}

	log.Printf("[GARDEN] Enriching player: %s", playerID)

	// Get player
	player, err := h.playerQueries.GetPlayerByID(r.Context(), playerID)
	if err != nil {
		response.NotFound(w, "Player")
		return
	}

	// Generate enrichments
	suggestions, err := h.enricher.EnrichPlayer(r.Context(), player)
	if err != nil {
		log.Printf("[GARDEN] Enrichment failed: %v", err)
		response.InternalError(w, "Enrichment failed")
		return
	}

	// Generate tags
	tags, err := h.enricher.GeneratePlayerTags(r.Context(), player, "Recent stats placeholder")
	if err != nil {
		log.Printf("[GARDEN] Tag generation failed: %v", err)
		tags = []string{}
	}

	// Generate summary
	summary, err := h.enricher.GeneratePlayerSummary(r.Context(), player, "Season stats placeholder")
	if err != nil {
		log.Printf("[GARDEN] Summary generation failed: %v", err)
		summary = ""
	}

	// Find similar players
	similar, err := h.enricher.SuggestRelatedPlayers(r.Context(), player)
	if err != nil {
		log.Printf("[GARDEN] Similar players failed: %v", err)
		similar = []string{}
	}

	log.Printf("[GARDEN] Enrichment complete: %d suggestions, %d tags",
		len(suggestions), len(tags))

	response.Success(w, map[string]interface{}{
		"player":      player,
		"enrichments": suggestions,
		"tags":        tags,
		"summary":     summary,
		"similar":     similar,
		"message":     fmt.Sprintf("Generated %d enrichments and %d tags", len(suggestions), len(tags)),
	})
}

// handleSyncSchedule generates intelligent sync schedule
func (h *GardenHandler) handleSyncSchedule(w http.ResponseWriter, r *http.Request) {
	if !h.aiService.IsAvailable() {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	log.Println("[GARDEN] Generating sync schedule...")

	plan, err := h.scheduler.GenerateSyncPlan(r.Context())
	if err != nil {
		log.Printf("[GARDEN] Schedule generation failed: %v", err)
		response.InternalError(w, "Schedule generation failed")
		return
	}

	log.Printf("[GARDEN] Schedule generated: %d recommendations (game day: %v)",
		len(plan.Recommendations), plan.GameDayMode)

	response.Success(w, map[string]interface{}{
		"schedule": plan,
		"message":  fmt.Sprintf("Generated schedule with %d sync recommendations", len(plan.Recommendations)),
	})
}

// handleGardenStatus provides overall garden status
func (h *GardenHandler) handleGardenStatus(w http.ResponseWriter, r *http.Request) {
	pool := db.GetPool()

	// Collect quick stats
	var playerCount, gameCount, injuryCount int
	pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM players").Scan(&playerCount)
	pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM games").Scan(&gameCount)
	pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM player_injuries").Scan(&injuryCount)

	// Check last update times
	var lastPlayerUpdate, lastGameUpdate time.Time
	pool.QueryRow(r.Context(), "SELECT MAX(updated_at) FROM players").Scan(&lastPlayerUpdate)
	pool.QueryRow(r.Context(), "SELECT MAX(updated_at) FROM games").Scan(&lastGameUpdate)

	status := map[string]interface{}{
		"ai_enabled":   h.aiService.IsAvailable(),
		"ai_provider":  h.aiService.GetProvider(),
		"data_counts": map[string]interface{}{
			"players":  playerCount,
			"games":    gameCount,
			"injuries": injuryCount,
		},
		"last_updates": map[string]interface{}{
			"players": lastPlayerUpdate,
			"games":   lastGameUpdate,
		},
		"garden_features": map[string]bool{
			"health_monitoring": h.aiService.IsAvailable(),
			"data_enrichment":   h.aiService.IsAvailable(),
			"natural_queries":   h.aiService.IsAvailable(),
			"smart_scheduling":  h.aiService.IsAvailable(),
		},
		"timestamp": time.Now(),
	}

	response.Success(w, status)
}

// min helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
