package middleware

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/francisco/gridironmind/pkg/response"
)

// APIKeyAuth middleware validates API key from X-API-Key header
func APIKeyAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API key from environment
		validAPIKey := os.Getenv("API_KEY")

		// If no API key configured, allow all requests (development mode)
		if validAPIKey == "" {
			next(w, r)
			return
		}

		// Extract API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header as Bearer token
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Validate API key
		if apiKey == "" {
			log.Printf("API key missing for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			response.Error(w, http.StatusUnauthorized, "MISSING_API_KEY", "API key is required")
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if !constantTimeCompare(apiKey, validAPIKey) {
			log.Printf("Invalid API key attempt for %s %s", r.Method, r.URL.Path)
			response.Error(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
			return
		}

		// Valid API key, continue
		next(w, r)
	}
}

// OptionalAPIKeyAuth allows requests with or without API key, but validates if present
func OptionalAPIKeyAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		validAPIKey := os.Getenv("API_KEY")

		// If no API key configured, allow all
		if validAPIKey == "" {
			next(w, r)
			return
		}

		// Extract API key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// If API key provided, validate it
		if apiKey != "" && !constantTimeCompare(apiKey, validAPIKey) {
			log.Printf("Invalid API key attempt for %s %s", r.Method, r.URL.Path)
			response.Error(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
			return
		}

		next(w, r)
	}
}

// Admin Auth middleware - requires API key for admin operations
func AdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API key from environment
		validAPIKey := os.Getenv("API_KEY")

		// If no API key configured, BLOCK in production, WARN in development
		if validAPIKey == "" {
			env := os.Getenv("ENVIRONMENT")
			if env == "production" {
				log.Printf("SECURITY: Admin endpoint accessed with no API key configured")
				response.Error(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication required")
				return
			}
			log.Printf("WARNING: Admin endpoint accessed with no auth (development mode)")
		}

		// Extract API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			// Also check Authorization header as Bearer token
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Validate API key
		if apiKey == "" {
			log.Printf("Admin endpoint accessed without API key: %s %s", r.Method, r.URL.Path)
			response.Error(w, http.StatusUnauthorized, "MISSING_API_KEY", "Admin API key is required")
			return
		}

		// Use constant-time comparison
		if validAPIKey != "" && !constantTimeCompare(apiKey, validAPIKey) {
			log.Printf("Invalid admin API key attempt for %s %s", r.Method, r.URL.Path)
			response.Error(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid admin API key")
			return
		}

		// Valid API key, continue
		next(w, r)
	}
}

// constantTimeCompare compares two strings in constant time to prevent timing attacks
func constantTimeCompare(a, b string) bool {
	// Convert strings to byte slices for subtle.ConstantTimeCompare
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}