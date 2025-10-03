package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/francisco/gridironmind/internal/scheduler"
	"github.com/francisco/gridironmind/pkg/response"
)

// SchedulerHandler handles scheduler-related requests
type SchedulerHandler struct {
	scheduler *scheduler.Scheduler
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(sched *scheduler.Scheduler) *SchedulerHandler {
	return &SchedulerHandler{
		scheduler: sched,
	}
}

// HandleSchedulerStatus returns the current scheduler status
func (h *SchedulerHandler) HandleSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	status := h.scheduler.GetStatus()
	response.Success(w, status)
}

// HandleSchedulerTrigger manually triggers a sync
func (h *SchedulerHandler) HandleSchedulerTrigger(w http.ResponseWriter, r *http.Request) {
	h.scheduler.TriggerSync()

	response.Success(w, map[string]interface{}{
		"message": "Sync triggered successfully",
		"status":  "running",
	})
}

// HandleSchedulerConfigure updates scheduler configuration
func (h *SchedulerHandler) HandleSchedulerConfigure(w http.ResponseWriter, r *http.Request) {
	// Parse configuration from request body
	var configUpdate struct {
		Enabled      *bool   `json:"enabled,omitempty"`
		Mode         *string `json:"mode,omitempty"`
		SyncGames    *bool   `json:"sync_games,omitempty"`
		SyncStats    *bool   `json:"sync_stats,omitempty"`
		SyncInjuries *bool   `json:"sync_injuries,omitempty"`
		ClearCache   *bool   `json:"clear_cache,omitempty"`
	}

	if err := parseJSONBody(r, &configUpdate); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Get current config
	currentStatus := h.scheduler.GetStatus()

	// For now, just return a message about configuration
	// In a full implementation, you'd update the scheduler config
	response.Success(w, map[string]interface{}{
		"message":        "Configuration update received",
		"current_status": currentStatus,
		"update":         configUpdate,
	})
}

// parseJSONBody is a helper to parse JSON request body
func parseJSONBody(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
