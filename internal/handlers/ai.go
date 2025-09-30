package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/francisco/gridironmind/internal/ai"
	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

type AIHandler struct {
	aiClient      *ai.Client
	gameQueries   *db.GameQueries
	playerQueries *db.PlayerQueries
	teamQueries   *db.TeamQueries
}

func NewAIHandler(cfg *config.Config) *AIHandler {
	var aiClient *ai.Client
	if cfg.ClaudeAPIKey != "" {
		aiClient = ai.NewClient(cfg.ClaudeAPIKey)
	}

	return &AIHandler{
		aiClient:      aiClient,
		gameQueries:   &db.GameQueries{},
		playerQueries: &db.PlayerQueries{},
		teamQueries:   &db.TeamQueries{},
	}
}

// HandlePredictGame handles POST /ai/predict/game/:gameID
func (h *AIHandler) HandlePredictGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	if h.aiClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	// Extract game ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/predict/game/")
	gameID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_GAME_ID", "Game ID must be a valid UUID")
		return
	}

	log.Printf("AI prediction requested for game %s", gameID)

	// Check cache first
	cacheKey := fmt.Sprintf("ai:predict:game:%s", gameID)
	if cached, err := cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Get game details
	game, err := h.gameQueries.GetGameByID(r.Context(), gameID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "GAME_NOT_FOUND", "Game not found")
		return
	}

	// Get team details
	homeTeam, _ := h.teamQueries.GetTeamByID(r.Context(), game.HomeTeamID)
	awayTeam, _ := h.teamQueries.GetTeamByID(r.Context(), game.AwayTeamID)

	if homeTeam == nil || awayTeam == nil {
		response.Error(w, http.StatusInternalServerError, "TEAM_NOT_FOUND", "Failed to retrieve team information")
		return
	}

	// Build stats context (simplified - in production would include more data)
	homeStats := fmt.Sprintf("Team: %s, Current Season Record", homeTeam.Name)
	awayStats := fmt.Sprintf("Team: %s, Current Season Record", awayTeam.Name)

	// Get AI prediction
	prediction, err := h.aiClient.PredictGameOutcome(r.Context(), homeTeam.Name, awayTeam.Name, homeStats, awayStats)
	if err != nil {
		log.Printf("AI prediction failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "PREDICTION_FAILED", "Failed to generate prediction")
		return
	}

	// Build response
	respData := map[string]interface{}{
		"game_id":    gameID,
		"home_team":  homeTeam.Name,
		"away_team":  awayTeam.Name,
		"prediction": prediction,
		"generated_at": getCurrentTimestamp(),
	}

	respJSON, _ := json.Marshal(map[string]interface{}{
		"data": respData,
		"meta": map[string]string{
			"timestamp": getCurrentTimestamp(),
		},
	})

	// Cache for 15 minutes
	cache.Set(r.Context(), cacheKey, string(respJSON), 15*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(respJSON)
}

// HandlePredictPlayer handles POST /ai/predict/player/:playerID/next-game
func (h *AIHandler) HandlePredictPlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	if h.aiClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	// Extract player ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/predict/player/")
	path = strings.TrimSuffix(path, "/next-game")
	playerID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PLAYER_ID", "Player ID must be a valid UUID")
		return
	}

	log.Printf("AI prediction requested for player %s", playerID)

	// Check cache
	cacheKey := fmt.Sprintf("ai:predict:player:%s", playerID)
	if cached, err := cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Get player details
	player, err := h.playerQueries.GetPlayerByID(r.Context(), playerID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "PLAYER_NOT_FOUND", "Player not found")
		return
	}

	// Build context (simplified)
	recentStats := fmt.Sprintf("Player: %s, Position: %s, Recent Performance Data", player.Name, player.Position)
	opponent := "Upcoming Opponent"

	// Get AI prediction
	prediction, err := h.aiClient.PredictPlayerPerformance(r.Context(), player.Name, player.Position, opponent, recentStats)
	if err != nil {
		log.Printf("AI prediction failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "PREDICTION_FAILED", "Failed to generate prediction")
		return
	}

	// Build response
	respData := map[string]interface{}{
		"player_id":    playerID,
		"player_name":  player.Name,
		"position":     player.Position,
		"prediction":   prediction,
		"generated_at": getCurrentTimestamp(),
	}

	respJSON, _ := json.Marshal(map[string]interface{}{
		"data": respData,
		"meta": map[string]string{
			"timestamp": getCurrentTimestamp(),
		},
	})

	// Cache for 30 minutes
	cache.Set(r.Context(), cacheKey, string(respJSON), 30*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(respJSON)
}

// HandleAnalyzePlayer handles POST /ai/insights/player/:playerID
func (h *AIHandler) HandleAnalyzePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	if h.aiClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	// Extract player ID
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ai/insights/player/")
	playerID, err := uuid.Parse(path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_PLAYER_ID", "Player ID must be a valid UUID")
		return
	}

	log.Printf("AI analysis requested for player %s", playerID)

	// Check cache
	cacheKey := fmt.Sprintf("ai:insights:player:%s", playerID)
	if cached, err := cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Get player details
	player, err := h.playerQueries.GetPlayerByID(r.Context(), playerID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "PLAYER_NOT_FOUND", "Player not found")
		return
	}

	// Build context
	seasonStats := fmt.Sprintf("Player: %s, Position: %s, Season Statistics", player.Name, player.Position)
	recentGames := "Recent game performances"

	// Get AI analysis
	analysis, err := h.aiClient.AnalyzePlayer(r.Context(), player.Name, player.Position, seasonStats, recentGames)
	if err != nil {
		log.Printf("AI analysis failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "ANALYSIS_FAILED", "Failed to generate analysis")
		return
	}

	// Build response
	respData := map[string]interface{}{
		"player_id":    playerID,
		"player_name":  player.Name,
		"position":     player.Position,
		"analysis":     analysis,
		"generated_at": getCurrentTimestamp(),
	}

	respJSON, _ := json.Marshal(map[string]interface{}{
		"data": respData,
		"meta": map[string]string{
			"timestamp": getCurrentTimestamp(),
		},
	})

	// Cache for 1 hour
	cache.Set(r.Context(), cacheKey, string(respJSON), 1*time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(respJSON)
}

// HandleAIQuery handles POST /ai/query
func (h *AIHandler) HandleAIQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST method is allowed")
		return
	}

	if h.aiClient == nil {
		response.Error(w, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI service not configured")
		return
	}

	// Parse request body
	var reqBody struct {
		Query string `json:"query"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if reqBody.Query == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_QUERY", "Query is required")
		return
	}

	log.Printf("AI query: %s", reqBody.Query)

	// Get AI response
	answer, err := h.aiClient.AnswerQuery(r.Context(), reqBody.Query, "Current NFL season data")
	if err != nil {
		log.Printf("AI query failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "QUERY_FAILED", "Failed to process query")
		return
	}

	// Build response
	respData := map[string]interface{}{
		"query":        reqBody.Query,
		"answer":       answer,
		"generated_at": getCurrentTimestamp(),
	}

	response.Success(w, respData)
}