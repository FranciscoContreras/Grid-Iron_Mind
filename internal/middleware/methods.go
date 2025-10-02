package middleware

import (
	"net/http"

	"github.com/francisco/gridironmind/pkg/response"
)

// MethodValidator creates middleware that validates HTTP methods.
//
// This middleware:
//   - Validates the incoming request method against allowed methods
//   - Returns 405 Method Not Allowed for invalid methods
//   - Eliminates duplicate method validation code across handlers
//
// Example usage:
//
//	// Allow only GET requests
//	mux.HandleFunc("/api/v1/players", middleware.MethodValidator(http.MethodGet)(handler))
//
//	// Allow GET and POST requests
//	mux.HandleFunc("/api/v1/admin/sync", middleware.MethodValidator(http.MethodGet, http.MethodPost)(handler))
func MethodValidator(allowedMethods ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if request method is in allowed methods
			for _, method := range allowedMethods {
				if r.Method == method {
					next(w, r)
					return
				}
			}

			// Method not allowed
			response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
	}
}

// GET creates middleware that allows only GET requests.
// Shorthand for MethodValidator(http.MethodGet).
//
// Example:
//
//	mux.HandleFunc("/api/v1/players", middleware.GET(handler))
func GET(next http.HandlerFunc) http.HandlerFunc {
	return MethodValidator(http.MethodGet)(next)
}

// POST creates middleware that allows only POST requests.
// Shorthand for MethodValidator(http.MethodPost).
//
// Example:
//
//	mux.HandleFunc("/api/v1/admin/sync", middleware.POST(handler))
func POST(next http.HandlerFunc) http.HandlerFunc {
	return MethodValidator(http.MethodPost)(next)
}

// PUT creates middleware that allows only PUT requests.
// Shorthand for MethodValidator(http.MethodPut).
func PUT(next http.HandlerFunc) http.HandlerFunc {
	return MethodValidator(http.MethodPut)(next)
}

// DELETE creates middleware that allows only DELETE requests.
// Shorthand for MethodValidator(http.MethodDelete).
func DELETE(next http.HandlerFunc) http.HandlerFunc {
	return MethodValidator(http.MethodDelete)(next)
}
