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
	"github.com/francisco/gridironmind/internal/scheduler"
	"github.com/francisco/gridironmind/internal/weather"
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

	// Initialize and start auto-sync scheduler
	schedulerConfig := scheduler.DefaultConfig(cfg.WeatherAPIKey)

	// Check for scheduler override from environment
	if os.Getenv("ENABLE_AUTO_SYNC") == "false" {
		schedulerConfig.Enabled = false
		log.Println("Auto-sync scheduler disabled via ENABLE_AUTO_SYNC=false")
	}

	autoScheduler := scheduler.NewScheduler(schedulerConfig)
	autoScheduler.Start()
	defer autoScheduler.Stop()

	// Initialize handlers
	playersHandler := handlers.NewPlayersHandler()
	teamsHandler := handlers.NewTeamsHandler()
	gamesHandler := handlers.NewGamesHandler()
	statsHandler := handlers.NewStatsHandler()
	defensiveHandler := handlers.NewDefensiveHandler()
	standingsHandler := handlers.NewStandingsHandler()
	adminHandler := handlers.NewAdminHandler(cfg.WeatherAPIKey)
	weatherHandler := handlers.NewWeatherHandler(weather.NewClient(cfg.WeatherAPIKey))
	styleAgentHandler := handlers.NewStyleAgentHandler()
	metricsHandler := handlers.NewMetricsHandler()
	schedulerHandler := handlers.NewSchedulerHandler(autoScheduler)

	// Setup router
	mux := http.NewServeMux()

	// API endpoints (all GET methods)
	mux.HandleFunc("/api/v1/players", applyGETMiddleware(playersHandler.HandlePlayers))
	mux.HandleFunc("/api/v1/players/", applyGETMiddleware(playersHandler.HandlePlayers))
	mux.HandleFunc("/api/v1/teams", applyGETMiddleware(teamsHandler.HandleTeams))
	mux.HandleFunc("/api/v1/teams/", applyGETMiddleware(teamsHandler.HandleTeams))
	mux.HandleFunc("/api/v1/games", applyGETMiddleware(gamesHandler.HandleGames))
	mux.HandleFunc("/api/v1/games/", applyGETMiddleware(gamesHandler.HandleGames))
	mux.HandleFunc("/api/v1/stats/leaders", applyGETMiddleware(statsHandler.HandleStatsLeaders))
	mux.HandleFunc("/api/v1/stats/game/", applyGETMiddleware(statsHandler.HandleGameStats))

	// Standings endpoint (GET)
	mux.HandleFunc("/api/v1/standings", applyGETMiddleware(standingsHandler.HandleStandings))

	// Defensive stats endpoints (GET)
	mux.HandleFunc("/api/v1/defense/rankings", applyGETMiddleware(defensiveHandler.HandleDefensiveRankings))

	// Admin endpoints for data ingestion (POST methods, require API key authentication)
	mux.HandleFunc("/api/v1/admin/sync/teams", applyPOSTAdminMiddleware(adminHandler.HandleSyncTeams))
	mux.HandleFunc("/api/v1/admin/sync/rosters", applyPOSTAdminMiddleware(adminHandler.HandleSyncRosters))
	mux.HandleFunc("/api/v1/admin/sync/games", applyPOSTAdminMiddleware(adminHandler.HandleSyncGames))
	mux.HandleFunc("/api/v1/admin/sync/full", applyPOSTAdminMiddleware(adminHandler.HandleFullSync))
	mux.HandleFunc("/api/v1/admin/sync/historical/season", applyPOSTAdminMiddleware(adminHandler.HandleSyncHistoricalGames))
	mux.HandleFunc("/api/v1/admin/sync/historical/seasons", applyPOSTAdminMiddleware(adminHandler.HandleSyncMultipleSeasons))

	// NFLverse enrichment endpoints (POST)
	mux.HandleFunc("/api/v1/admin/sync/nflverse/stats", applyPOSTAdminMiddleware(adminHandler.HandleSyncNFLverseStats))
	mux.HandleFunc("/api/v1/admin/sync/nflverse/schedule", applyPOSTAdminMiddleware(adminHandler.HandleSyncNFLverseSchedule))
	mux.HandleFunc("/api/v1/admin/sync/nflverse/nextgen", applyPOSTAdminMiddleware(adminHandler.HandleSyncNFLverseNextGen))

	// Next Gen Stats sync endpoint (POST)
	mux.HandleFunc("/api/v1/admin/sync/nextgen-stats", applyPOSTAdminMiddleware(adminHandler.HandleSyncNextGenStats))

	// Weather enrichment endpoint (POST)
	mux.HandleFunc("/api/v1/admin/sync/weather", applyPOSTAdminMiddleware(adminHandler.HandleEnrichWeather))

	// Team stats sync endpoint (POST)
	mux.HandleFunc("/api/v1/admin/sync/team-stats", applyPOSTAdminMiddleware(adminHandler.HandleSyncTeamStats))

	// Injury sync endpoint (POST)
	mux.HandleFunc("/api/v1/admin/sync/injuries", applyPOSTAdminMiddleware(adminHandler.HandleSyncInjuries))

	// Scoring plays sync endpoint (POST)
	mux.HandleFunc("/api/v1/admin/sync/scoring-plays", applyPOSTAdminMiddleware(adminHandler.HandleSyncScoringPlays))

	// Player season stats sync endpoint (POST)
	mux.HandleFunc("/api/v1/admin/sync/player-season-stats", applyPOSTAdminMiddleware(adminHandler.HandleSyncPlayerSeasonStats))

	// Standings calculation endpoint (POST)
	mux.HandleFunc("/api/v1/admin/calc/standings", applyPOSTAdminMiddleware(adminHandler.HandleCalculateStandings))

	// Scheduler control endpoints (GET and POST)
	mux.HandleFunc("/api/v1/admin/scheduler/status", applyGETAdminMiddleware(schedulerHandler.HandleSchedulerStatus))
	mux.HandleFunc("/api/v1/admin/scheduler/trigger", applyPOSTAdminMiddleware(schedulerHandler.HandleSchedulerTrigger))
	mux.HandleFunc("/api/v1/admin/scheduler/configure", applyPOSTAdminMiddleware(schedulerHandler.HandleSchedulerConfigure))

	// Weather API endpoints (GET)
	mux.HandleFunc("/api/v1/weather/current", applyGETMiddleware(weatherHandler.HandleCurrentWeather))
	mux.HandleFunc("/api/v1/weather/historical", applyGETMiddleware(weatherHandler.HandleHistoricalWeather))
	mux.HandleFunc("/api/v1/weather/forecast", applyGETMiddleware(weatherHandler.HandleForecastWeather))

	// Generate API Key (POST)
	mux.HandleFunc("/api/v1/admin/keys/generate", applyPOSTAdminMiddleware(adminHandler.HandleGenerateAPIKey))

	// ========================================
	// API v2 - Enhanced endpoints with new features
	// ========================================
	mux.HandleFunc("/api/v2/players", applyGETMiddleware(playersHandler.HandlePlayers))
	mux.HandleFunc("/api/v2/players/", applyGETMiddleware(playersHandler.HandlePlayers))
	mux.HandleFunc("/api/v2/teams", applyGETMiddleware(teamsHandler.HandleTeams))
	mux.HandleFunc("/api/v2/teams/", applyGETMiddleware(teamsHandler.HandleTeams))
	mux.HandleFunc("/api/v2/games", applyGETMiddleware(gamesHandler.HandleGames))
	mux.HandleFunc("/api/v2/games/", applyGETMiddleware(gamesHandler.HandleGames))
	mux.HandleFunc("/api/v2/stats/leaders", applyGETMiddleware(statsHandler.HandleStatsLeaders))
	mux.HandleFunc("/api/v2/stats/game/", applyGETMiddleware(statsHandler.HandleGameStats))
	mux.HandleFunc("/api/v2/standings", applyGETMiddleware(standingsHandler.HandleStandings))
	mux.HandleFunc("/api/v2/defense/rankings", applyGETMiddleware(defensiveHandler.HandleDefensiveRankings))
	mux.HandleFunc("/api/v2/weather/current", applyGETMiddleware(weatherHandler.HandleCurrentWeather))
	mux.HandleFunc("/api/v2/weather/historical", applyGETMiddleware(weatherHandler.HandleHistoricalWeather))
	mux.HandleFunc("/api/v2/weather/forecast", applyGETMiddleware(weatherHandler.HandleForecastWeather))

	// Health check endpoints (GET)
	mux.HandleFunc("/health", applyGETMiddleware(healthCheck))
	mux.HandleFunc("/api/v1/health", applyGETMiddleware(healthCheck))
	mux.HandleFunc("/api/v2/health", applyGETMiddleware(healthCheck))

	// Metrics endpoints (GET)
	mux.HandleFunc("/api/v1/metrics/database", applyGETMiddleware(metricsHandler.HandleDatabaseMetrics))
	mux.HandleFunc("/api/v1/metrics/health", applyGETMiddleware(metricsHandler.HandleHealthMetrics))
	mux.HandleFunc("/api/v2/metrics/database", applyGETMiddleware(metricsHandler.HandleDatabaseMetrics))
	mux.HandleFunc("/api/v2/metrics/health", applyGETMiddleware(metricsHandler.HandleHealthMetrics))

	// API Documentation endpoints
	mux.HandleFunc("/api-docs.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./dashboard/api-docs.html")
	})
	mux.HandleFunc("/api-v2-docs.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./dashboard/api-v2-docs.html")
	})
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api-v2-docs.html", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api-v2-docs.html", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api-v2-docs.html", http.StatusMovedPermanently)
	})

	// Dashboard redirect
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/dashboard/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	// UI System Documentation endpoint
	mux.HandleFunc("/ui-system.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./dashboard/ui-system.html")
	})

	// Style Agent endpoints
	mux.HandleFunc("/api/v1/style/check", applyMiddleware(styleAgentHandler.HandleStyleCheck))
	mux.HandleFunc("/api/v1/style/rules", applyMiddleware(styleAgentHandler.HandleStyleRules))
	mux.HandleFunc("/api/v1/style/example", applyMiddleware(styleAgentHandler.HandleStyleExample))
	mux.HandleFunc("/style-guide.html", styleAgentHandler.HandleStyleGuide)

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

func applyAdminMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				middleware.AdminAuth(
					middleware.StandardRateLimit(handler),
				),
			),
		),
	)
}

// applyGETMiddleware applies standard middleware + GET method validation
func applyGETMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				middleware.GET(
					middleware.StandardRateLimit(handler),
				),
			),
		),
	)
}

// applyPOSTAdminMiddleware applies admin middleware + POST method validation
func applyPOSTAdminMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				middleware.AdminAuth(
					middleware.POST(
						middleware.StandardRateLimit(handler),
					),
				),
			),
		),
	)
}

// applyGETAdminMiddleware applies admin middleware + GET method validation
func applyGETAdminMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				middleware.AdminAuth(
					middleware.GET(
						middleware.StandardRateLimit(handler),
					),
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