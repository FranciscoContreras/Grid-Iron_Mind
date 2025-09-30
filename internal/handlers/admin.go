package handlers

import (
	"log"
	"net/http"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/internal/ingestion"
	"github.com/francisco/gridironmind/pkg/response"
)

type AdminHandler struct {
	ingestionService *ingestion.Service
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		ingestionService: ingestion.NewService(),
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

	ctx := r.Context()
	if err := h.ingestionService.SyncAllRosters(ctx); err != nil {
		log.Printf("Rosters sync failed: %v", err)
		response.Error(w, http.StatusInternalServerError, "SYNC_FAILED", "Failed to sync rosters")
		return
	}

	// Invalidate players cache
	if err := cache.DeletePattern(ctx, cache.InvalidatePlayersCache()); err != nil {
		log.Printf("Failed to invalidate players cache: %v", err)
	}

	response.Success(w, map[string]interface{}{
		"message": "Rosters sync completed successfully",
		"status":  "success",
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