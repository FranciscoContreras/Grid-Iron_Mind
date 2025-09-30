package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/ingestion"
	"github.com/francisco/gridironmind/pkg/response"
)

type AdminHandler struct {
	ingestionService *ingestion.Service
}

func NewAdminHandler(weatherAPIKey string) *AdminHandler {
	return &AdminHandler{
		ingestionService: ingestion.NewService(weatherAPIKey),
	}
}

// HandleSyncTeams triggers a teams sync
func (h *AdminHandler) HandleSyncTeams(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Admin endpoint: Teams sync requested")

	ctx := r.Context()
	if err := h.ingestionService.SyncTeams(ctx); err != nil {
		log.Printf("Teams sync failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "SYNC_FAILED", "Failed to sync teams")
		return
	}

	// Invalidate teams cache
	if err := cache.DeletePattern(ctx, cache.InvalidateTeamsCache()); err != nil {
		log.Printf("Failed to invalidate teams cache: %v", err)
	}

	response.Success(w, map[string]interface{}{
		"message": "Teams sync completed successfully",
		"status":  "success",
	})
}

// HandleSyncRosters triggers a full roster sync for all teams
func (h *AdminHandler) HandleSyncRosters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Admin endpoint: Rosters sync requested")

	// Run sync in background for long operations
	go func() {
		// Use background context to prevent cancellation
		ctx := context.Background()
		if err := h.ingestionService.SyncAllRosters(ctx); err != nil {
			log.Printf("Rosters sync failed: %v", err)
		} else {
			log.Println("Rosters sync completed successfully")
			// Invalidate players cache on success
			if err := cache.DeletePattern(ctx, cache.InvalidatePlayersCache()); err != nil {
				log.Printf("Failed to invalidate players cache: %v", err)
			}
		}
	}()

	response.Success(w, map[string]interface{}{
		"message": "Rosters sync started in background",
		"status":  "processing",
	})
}

// HandleSyncGames triggers a games/scoreboard sync
func (h *AdminHandler) HandleSyncGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Admin endpoint: Games sync requested")

	ctx := r.Context()
	if err := h.ingestionService.SyncGames(ctx); err != nil {
		log.Printf("Games sync failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "SYNC_FAILED", "Failed to sync games")
		return
	}

	// Invalidate games and stats cache
	if err := cache.DeletePattern(ctx, cache.InvalidateGamesCache()); err != nil {
		log.Printf("Failed to invalidate games cache: %v", err)
	}
	if err := cache.DeletePattern(ctx, cache.InvalidateStatsCache()); err != nil {
		log.Printf("Failed to invalidate stats cache: %v", err)
	}

	response.Success(w, map[string]interface{}{
		"message": "Games sync completed successfully",
		"status":  "success",
	})
}

// HandleFullSync triggers a complete data sync (teams -> rosters -> games)
func (h *AdminHandler) HandleFullSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Admin endpoint: Full sync requested")

	ctx := r.Context()

	// Run sync in background for long operations
	go func() {
		if err := h.ingestionService.FullSync(ctx); err != nil {
			log.Printf("Full sync failed: %v", err)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message": "Full sync started in background",
		"status":  "processing",
	})
}

// HandleGenerateAPIKey generates a new API key
func (h *AdminHandler) HandleGenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reqBody struct {
		Unlimited bool   `json:"unlimited"`
		Label     string `json:"label"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Generate secure random API key
	apiKey, err := generateSecureAPIKey()
	if err != nil {
		log.Printf("Failed to generate API key: %v", err)
		response.Error(w, http.StatusInternalServerError, "GENERATION_FAILED", "Failed to generate API key")
		return
	}

	keyType := "standard"
	if reqBody.Unlimited {
		keyType = "unlimited"
	}

	log.Printf("Admin endpoint: Generated %s API key with label '%s'", keyType, reqBody.Label)

	response.Success(w, map[string]interface{}{
		"api_key":   apiKey,
		"type":      keyType,
		"label":     reqBody.Label,
		"unlimited": reqBody.Unlimited,
		"message":   "API key generated successfully. Store this key securely - it cannot be retrieved again.",
	})
}

// generateSecureAPIKey generates a cryptographically secure random API key
func generateSecureAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "gim_" + hex.EncodeToString(bytes), nil
}

// HandleSyncHistoricalGames handles POST /admin/sync/historical/season/:year
func (h *AdminHandler) HandleSyncHistoricalGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract year from URL
	var yearStr struct {
		Year int `json:"year"`
	}

	if err := json.NewDecoder(r.Body).Decode(&yearStr); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	log.Printf("Admin endpoint: Historical games sync requested for season %d", yearStr.Year)

	ctx := r.Context()

	// Run sync in background for long operations
	go func() {
		if err := h.ingestionService.SyncHistoricalGames(ctx, yearStr.Year); err != nil {
			log.Printf("Historical games sync failed: %v", err)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Historical games sync started for season %d", yearStr.Year),
		"season":  yearStr.Year,
		"status":  "processing",
	})
}

// HandleSyncMultipleSeasons handles POST /admin/sync/historical/seasons
func (h *AdminHandler) HandleSyncMultipleSeasons(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var reqBody struct {
		StartYear int `json:"start_year"`
		EndYear   int `json:"end_year"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	log.Printf("Admin endpoint: Multi-season sync requested from %d to %d", reqBody.StartYear, reqBody.EndYear)

	ctx := r.Context()

	// Run sync in background
	go func() {
		if err := h.ingestionService.SyncMultipleSeasons(ctx, reqBody.StartYear, reqBody.EndYear); err != nil {
			log.Printf("Multi-season sync failed: %v", err)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message":    fmt.Sprintf("Multi-season sync started from %d to %d", reqBody.StartYear, reqBody.EndYear),
		"start_year": reqBody.StartYear,
		"end_year":   reqBody.EndYear,
		"status":     "processing",
	})
}

// HandleSyncNFLverseStats handles POST /admin/sync/nflverse/stats
func (h *AdminHandler) HandleSyncNFLverseStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Season int `json:"season"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	log.Printf("Admin endpoint: NFLverse stats sync requested for season %d", reqBody.Season)

	// Run sync in background with background context
	go func() {
		ctx := context.Background()
		if err := h.ingestionService.SyncNFLversePlayerStats(ctx, reqBody.Season); err != nil {
			log.Printf("NFLverse stats sync failed: %v", err)
		} else {
			log.Printf("NFLverse stats sync completed successfully for season %d", reqBody.Season)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("NFLverse stats sync started for season %d", reqBody.Season),
		"season":  reqBody.Season,
		"status":  "processing",
	})
}

// HandleSyncNFLverseSchedule handles POST /admin/sync/nflverse/schedule
func (h *AdminHandler) HandleSyncNFLverseSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Season int `json:"season"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	log.Printf("Admin endpoint: NFLverse schedule sync requested for season %d", reqBody.Season)

	// Run sync in background with background context
	go func() {
		ctx := context.Background()
		if err := h.ingestionService.SyncNFLverseSchedule(ctx, reqBody.Season); err != nil {
			log.Printf("NFLverse schedule sync failed: %v", err)
		} else {
			log.Printf("NFLverse schedule sync completed successfully for season %d", reqBody.Season)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("NFLverse schedule sync started for season %d", reqBody.Season),
		"season":  reqBody.Season,
		"status":  "processing",
	})
}

// HandleSyncNFLverseNextGen handles POST /admin/sync/nflverse/nextgen
func (h *AdminHandler) HandleSyncNFLverseNextGen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Season   int    `json:"season"`
		StatType string `json:"stat_type"` // passing, rushing, receiving
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if reqBody.StatType == "" {
		reqBody.StatType = "passing"
	}

	log.Printf("Admin endpoint: NFLverse Next Gen Stats sync requested for season %d, type: %s", reqBody.Season, reqBody.StatType)

	// Run sync in background with background context
	go func() {
		ctx := context.Background()
		if err := h.ingestionService.SyncNFLverseNextGenStats(ctx, reqBody.Season, reqBody.StatType); err != nil {
			log.Printf("NFLverse Next Gen Stats sync failed: %v", err)
		} else {
			log.Printf("NFLverse Next Gen Stats sync completed successfully for season %d, type: %s", reqBody.Season, reqBody.StatType)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message":   fmt.Sprintf("NFLverse Next Gen Stats sync started for season %d", reqBody.Season),
		"season":    reqBody.Season,
		"stat_type": reqBody.StatType,
		"status":    "processing",
	})
}

// HandleEnrichWeather handles POST /admin/sync/weather
func (h *AdminHandler) HandleEnrichWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Season int `json:"season"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if reqBody.Season == 0 {
		reqBody.Season = time.Now().Year()
	}

	log.Printf("Admin endpoint: Weather enrichment requested for season %d", reqBody.Season)

	ctx := r.Context()

	// Run enrichment in background
	go func() {
		if err := h.ingestionService.EnrichGamesWithWeather(ctx, reqBody.Season); err != nil {
			log.Printf("Weather enrichment failed: %v", err)
		}
	}()

	response.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Weather enrichment started for season %d", reqBody.Season),
		"season":  reqBody.Season,
		"status":  "processing",
	})
}

// HandleSyncTeamStats triggers team stats sync for completed games
func (h *AdminHandler) HandleSyncTeamStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Season int  `json:"season"`
		Week   int  `json:"week"`
		Debug  bool `json:"debug"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if reqBody.Season == 0 {
		reqBody.Season = time.Now().Year()
	}
	if reqBody.Week == 0 {
		response.Error(w, http.StatusBadRequest, "MISSING_WEEK", "Week parameter is required")
		return
	}

	log.Printf("Admin endpoint: Team stats sync requested for season %d, week %d (debug=%v)", reqBody.Season, reqBody.Week, reqBody.Debug)

	ctx := r.Context()
	if err := h.ingestionService.SyncGameTeamStats(ctx, reqBody.Season, reqBody.Week); err != nil {
		log.Printf("Team stats sync failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "SYNC_FAILED", fmt.Sprintf("Failed to sync team stats: %v", err))
		return
	}

	// Invalidate games cache since stats are now available
	if err := cache.DeletePattern(ctx, cache.InvalidateGamesCache()); err != nil {
		log.Printf("Failed to invalidate games cache: %v", err)
	}

	response.Success(w, map[string]interface{}{
		"message": fmt.Sprintf("Team stats sync completed for season %d, week %d", reqBody.Season, reqBody.Week),
		"season":  reqBody.Season,
		"week":    reqBody.Week,
		"status":  "success",
	})
}