package response

import (
	"log"
	"net/http"
)

// LogAndError logs an error with request context and returns a structured error response.
//
// This function:
//   - Logs the error with HTTP method, path, and error code
//   - Returns a structured JSON error response
//   - Should be used for all error conditions that need logging
//
// Example:
//
//	if err := db.Query(ctx, id); err != nil {
//	    response.LogAndError(w, r, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to query database", err)
//	    return
//	}
func LogAndError(w http.ResponseWriter, r *http.Request, status int, code, message string, err error) {
	// Log with full context
	if err != nil {
		log.Printf("[ERROR] %s %s - %s: %v", r.Method, r.URL.Path, code, err)
	} else {
		log.Printf("[ERROR] %s %s - %s: %s", r.Method, r.URL.Path, code, message)
	}

	// Return structured error response
	Error(w, status, code, message)
}

// LogAndBadRequest logs a bad request error and returns a 400 response.
//
// Use this for validation errors, invalid parameters, malformed requests.
//
// Example:
//
//	id, err := uuid.Parse(idStr)
//	if err != nil {
//	    response.LogAndBadRequest(w, r, "Invalid UUID format", err)
//	    return
//	}
func LogAndBadRequest(w http.ResponseWriter, r *http.Request, message string, err error) {
	if err != nil {
		log.Printf("[WARN] %s %s - BAD_REQUEST: %s - %v", r.Method, r.URL.Path, message, err)
	} else {
		log.Printf("[WARN] %s %s - BAD_REQUEST: %s", r.Method, r.URL.Path, message)
	}
	BadRequest(w, message)
}

// LogAndNotFound logs a not found error and returns a 404 response.
//
// Use this when a requested resource doesn't exist.
//
// Example:
//
//	player, err := queries.GetPlayer(ctx, id)
//	if err == db.ErrNotFound {
//	    response.LogAndNotFound(w, r, "Player")
//	    return
//	}
func LogAndNotFound(w http.ResponseWriter, r *http.Request, resource string) {
	log.Printf("[INFO] %s %s - NOT_FOUND: %s", r.Method, r.URL.Path, resource)
	NotFound(w, resource)
}

// LogAndInternalError logs an internal server error and returns a 500 response.
//
// Use this for unexpected errors, database failures, external API errors.
//
// Example:
//
//	if err := syncService.SyncTeams(ctx); err != nil {
//	    response.LogAndInternalError(w, r, "Failed to sync teams", err)
//	    return
//	}
func LogAndInternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s %s - INTERNAL_ERROR: %s - %v", r.Method, r.URL.Path, message, err)
	} else {
		log.Printf("[ERROR] %s %s - INTERNAL_ERROR: %s", r.Method, r.URL.Path, message)
	}
	InternalError(w, message)
}

// LogAndUnauthorized logs an unauthorized access attempt and returns a 401 response.
//
// Use this for authentication failures.
//
// Example:
//
//	if !isValidAPIKey(apiKey) {
//	    response.LogAndUnauthorized(w, r, "Invalid API key", nil)
//	    return
//	}
func LogAndUnauthorized(w http.ResponseWriter, r *http.Request, message string, err error) {
	if err != nil {
		log.Printf("[WARN] %s %s - UNAUTHORIZED: %s - %v", r.Method, r.URL.Path, message, err)
	} else {
		log.Printf("[WARN] %s %s - UNAUTHORIZED: %s", r.Method, r.URL.Path, message)
	}
	Unauthorized(w, message)
}

// LogWarning logs a warning message with request context.
//
// Use this for non-error conditions that should be logged (empty results, deprecation warnings, etc.)
//
// Example:
//
//	if len(results) == 0 {
//	    response.LogWarning(r, "No games found for season %d week %d", season, week)
//	}
func LogWarning(r *http.Request, format string, args ...interface{}) {
	prefix := "[WARN] " + r.Method + " " + r.URL.Path + " - "
	log.Printf(prefix+format, args...)
}

// LogInfo logs an informational message with request context.
//
// Use this for important operations (cache hits, auto-fetch, etc.)
//
// Example:
//
//	response.LogInfo(r, "Auto-fetched %d games for season %d week %d", count, season, week)
func LogInfo(r *http.Request, format string, args ...interface{}) {
	prefix := "[INFO] " + r.Method + " " + r.URL.Path + " - "
	log.Printf(prefix+format, args...)
}
