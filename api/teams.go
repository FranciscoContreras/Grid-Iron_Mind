package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/francisco/gridironmind/pkg/config"
	"github.com/francisco/gridironmind/pkg/database"
	"github.com/francisco/gridironmind/pkg/handlers"
	"github.com/francisco/gridironmind/pkg/middleware"
	"github.com/francisco/gridironmind/pkg/response"
)

var teamsHandler *handlers.TeamsHandler
var teamsDBInitialized bool

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database connection if not already done
	if !teamsDBInitialized {
		if err := initTeamsDB(); err != nil {
			log.Printf("Failed to initialize database: %v", err)
			response.InternalError(w, "Database connection failed")
			return
		}
		teamsDBInitialized = true
		teamsHandler = handlers.NewTeamsHandler()
	}

	// Apply middleware chain
	handler := middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				teamsHandler.HandleTeams,
			),
		),
	)

	handler(w, r)
}

func initTeamsDB() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbConfig := database.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    cfg.DBMaxConns,
		MinConns:    cfg.DBMinConns,
	}

	return database.Connect(context.Background(), dbConfig)
}