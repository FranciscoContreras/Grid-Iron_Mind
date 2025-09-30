package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/francisco/gridironmind/internal/config"
	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/internal/handlers"
	"github.com/francisco/gridironmind/internal/middleware"
	"github.com/francisco/gridironmind/pkg/response"
)

var playersHandler *handlers.PlayersHandler
var dbInitialized bool

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	// Initialize database connection if not already done
	if !dbInitialized {
		if err := initDB(); err != nil {
			log.Printf("Failed to initialize database: %v", err)
			response.InternalError(w, "Database connection failed")
			return
		}
		dbInitialized = true
		playersHandler = handlers.NewPlayersHandler()
	}

	// Apply middleware chain
	handler := middleware.CORS(
		middleware.LogRequest(
			middleware.RecoverPanic(
				playersHandler.HandlePlayers,
			),
		),
	)

	handler(w, r)
}

func initDB() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbConfig := db.Config{
		DatabaseURL: cfg.DatabaseURL,
		MaxConns:    cfg.DBMaxConns,
		MinConns:    cfg.DBMinConns,
	}

	return db.Connect(context.Background(), dbConfig)
}