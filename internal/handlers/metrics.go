package handlers

import (
	"net/http"

	"github.com/francisco/gridironmind/internal/db"
	"github.com/francisco/gridironmind/pkg/response"
)

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// HandleDatabaseMetrics returns database connection pool metrics
func (h *MetricsHandler) HandleDatabaseMetrics(w http.ResponseWriter, r *http.Request) {
	// Get pool metrics
	poolMetrics := db.PoolMetrics()

	// Get health status
	healthy := db.IsHealthy()

	// Combine into response
	metrics := map[string]interface{}{
		"database": poolMetrics,
		"healthy":  healthy,
	}

	response.Success(w, metrics)
}

// HandleHealthMetrics returns combined health metrics
func (h *MetricsHandler) HandleHealthMetrics(w http.ResponseWriter, r *http.Request) {
	// Check database health
	dbHealthy := db.IsHealthy()

	// Get database metrics
	poolMetrics := db.PoolMetrics()

	// Combine all metrics
	metrics := map[string]interface{}{
		"status":       getStatus(dbHealthy),
		"database": map[string]interface{}{
			"healthy": dbHealthy,
			"pool":    poolMetrics,
		},
		"service": map[string]interface{}{
			"name":    "Grid Iron Mind API",
			"version": "2.0.0",
		},
	}

	// Set status code based on health
	if !dbHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	response.Success(w, metrics)
}

func getStatus(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}
