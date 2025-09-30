package middleware

import (
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

		if apiKey != validAPIKey {
			log.Printf("Invalid API key attempt for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
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
		if apiKey != "" && apiKey != validAPIKey {
			log.Printf("Invalid API key attempt for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			response.Error(w, http.StatusUnauthorized, "INVALID_API_KEY", "Invalid API key")
			return
		}

		next(w, r)
	}
}