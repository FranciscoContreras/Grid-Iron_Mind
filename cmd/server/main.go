package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/handlers"
	"github.com/francisco/gridironmind/internal/middleware"
	"github.com/francisco/gridironmind/pkg/response"
)

func main() {
	log.Println("Starting Grid Iron Mind API server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dbConfig := db.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    cfg.DBMaxConns,
		MinConns:    cfg.DBMinConns,
	}

	if err := db.Connect(context.Background(), dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Connect to Redis (optional - continue if not available)
	if cfg.RedisURL != "" {
		cacheConfig := cache.Config{
			RedisURL: cfg.RedisURL,
		}
		if err := cache.Connect(cacheConfig); err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v (caching disabled)", err)
		} else {
			defer cache.Close()
		}
	} else {
		log.Println("Redis URL not configured (caching disabled)")
	}

	// Initialize handlers
	playersHandler := handlers.NewPlayersHandler()
	teamsHandler := handlers.NewTeamsHandler()
	gamesHandler := handlers.NewGamesHandler()
	statsHandler := handlers.NewStatsHandler()
	careerHandler := handlers.NewCareerHandler()
	aiHandler := handlers.NewAIHandler(cfg)
	adminHandler := handlers.NewAdminHandler()

	// Setup router
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/v1/players", applyMiddleware(playersHandler.HandlePlayers))
	mux.HandleFunc("/api/v1/players/", applyMiddleware(playersHandler.HandlePlayers))
	mux.HandleFunc("/api/v1/teams", applyMiddleware(teamsHandler.HandleTeams))
	mux.HandleFunc("/api/v1/teams/", applyMiddleware(teamsHandler.HandleTeams))
	mux.HandleFunc("/api/v1/games", applyMiddleware(gamesHandler.HandleGames))
	mux.HandleFunc("/api/v1/games/", applyMiddleware(gamesHandler.HandleGames))
	mux.HandleFunc("/api/v1/stats/leaders", applyMiddleware(statsHandler.HandleStatsLeaders))

	// Career and history endpoints
	mux.HandleFunc("/api/v1/players/:id/career", applyMiddleware(careerHandler.HandlePlayerCareerStats))
	mux.HandleFunc("/api/v1/players/:id/history", applyMiddleware(careerHandler.HandlePlayerTeamHistory))

	// AI endpoints (require API key and stricter rate limiting)
	mux.HandleFunc("/api/v1/ai/predict/game/", applyAIMiddleware(aiHandler.HandlePredictGame))
	mux.HandleFunc("/api/v1/ai/predict/player/", applyAIMiddleware(aiHandler.HandlePredictPlayer))
	mux.HandleFunc("/api/v1/ai/insights/player/", applyAIMiddleware(aiHandler.HandleAnalyzePlayer))
	mux.HandleFunc("/api/v1/ai/query", applyAIMiddleware(aiHandler.HandleAIQuery))

	// Admin endpoints for data ingestion
	mux.HandleFunc("/api/v1/admin/sync/teams", applyMiddleware(adminHandler.HandleSyncTeams))
	mux.HandleFunc("/api/v1/admin/sync/rosters", applyMiddleware(adminHandler.HandleSyncRosters))
	mux.HandleFunc("/api/v1/admin/sync/games", applyMiddleware(adminHandler.HandleSyncGames))
	mux.HandleFunc("/api/v1/admin/sync/full", applyMiddleware(adminHandler.HandleFullSync))
	mux.HandleFunc("/api/v1/admin/sync/historical/season", applyMiddleware(adminHandler.HandleSyncHistoricalGames))
	mux.HandleFunc("/api/v1/admin/sync/historical/seasons", applyMiddleware(adminHandler.HandleSyncMultipleSeasons))
	mux.HandleFunc("/api/v1/admin/keys/generate", applyMiddleware(adminHandler.HandleGenerateAPIKey))

	// Health check endpoint
	mux.HandleFunc("/health", applyMiddleware(healthCheck))
	mux.HandleFunc("/api/v1/health", applyMiddleware(healthCheck))

	// Serve dashboard static files
	fs := http.FileServer(http.Dir("./dashboard"))
	mux.Handle("/", fs)

	// Get port from environment (Heroku sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				middleware.StandardRateLimit(handler),
			),
		),
	)
}

func applyAIMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				middleware.APIKeyAuth(
					middleware.StrictRateLimit(handler),
				),
			),
		),
	)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := db.HealthCheck(r.Context()); err != nil {
		response.Error(w, http.StatusServiceUnavailable, "UNHEALTHY", "Database connection failed")
		return
	}

	response.Success(w, map[string]interface{}{
		"status":  "healthy",
		"service": "Grid Iron Mind API",
		"version": "1.0.0",
	})
}