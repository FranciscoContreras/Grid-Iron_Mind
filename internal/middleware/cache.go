package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
)

// responseWriter wraps http.ResponseWriter to capture the response
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Cache middleware with custom TTL
func Cache(ttl time.Duration, keyFunc func(*http.Request) string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next(w, r)
				return
			}

			// Generate cache key
			cacheKey := keyFunc(r)
			if cacheKey == "" {
				next(w, r)
				return
			}

			// Try to get from cache
			cached, err := cache.Get(r.Context(), cacheKey)
			if err != nil {
				log.Printf("Cache get error for key %s: %v", cacheKey, err)
				next(w, r)
				return
			}

			// Cache hit
			if cached != "" {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(cached))
				log.Printf("Cache HIT: %s", cacheKey)
				return
			}

			// Cache miss - capture response
			rw := newResponseWriter(w)
			next(rw, r)

			// Only cache successful responses
			if rw.statusCode == http.StatusOK {
				// Verify it's valid JSON before caching
				var js json.RawMessage
				if err := json.Unmarshal(rw.body.Bytes(), &js); err == nil {
					if err := cache.Set(r.Context(), cacheKey, rw.body.String(), ttl); err != nil {
						log.Printf("Cache set error for key %s: %v", cacheKey, err)
					} else {
						log.Printf("Cache MISS (now cached): %s", cacheKey)
					}
				}
			}

			rw.ResponseWriter.Header().Set("X-Cache", "MISS")
		}
	}
}

// CacheWithDefault creates a cache middleware with the provided key function
func CacheWithDefault(keyFunc func(*http.Request) string, ttl time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return Cache(ttl, keyFunc)
}