package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/francisco/gridironmind/internal/cache"
	"github.com/francisco/gridironmind/pkg/response"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// DefaultRateLimit is the standard rate limit (100 requests per minute)
var DefaultRateLimit = RateLimitConfig{
	RequestsPerMinute: 100,
	BurstSize:         10,
}

// AIRateLimit is the stricter rate limit for AI endpoints (10 requests per minute)
var AIRateLimit = RateLimitConfig{
	RequestsPerMinute: 10,
	BurstSize:         2,
}

// RateLimit middleware with configurable limits
func RateLimit(config RateLimitConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get client identifier (IP address or API key)
			clientID := getClientIdentifier(r)

			// Check rate limit
			allowed, remaining, resetTime, err := checkRateLimit(r.Context(), clientID, config)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

			if err != nil {
				// Redis error - allow request but log error
				log.Printf("Rate limit check error: %v", err)
				next(w, r)
				return
			}

			if !allowed {
				retryAfter := resetTime - time.Now().Unix()
				if retryAfter < 0 {
					retryAfter = 60
				}
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))

				log.Printf("Rate limit exceeded for %s on %s %s", clientID, r.Method, r.URL.Path)
				response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED",
					fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", retryAfter))
				return
			}

			next(w, r)
		}
	}
}

// getClientIdentifier returns a unique identifier for the client
func getClientIdentifier(r *http.Request) string {
	// Try API key first
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return "api:" + apiKey
	}

	// Fall back to IP address
	// Check X-Forwarded-For header (for proxies like Heroku)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return "ip:" + forwarded
	}

	return "ip:" + r.RemoteAddr
}

// checkRateLimit checks if a request is allowed based on rate limits
func checkRateLimit(ctx context.Context, clientID string, config RateLimitConfig) (allowed bool, remaining int, resetTime int64, err error) {
	// Generate rate limit key
	now := time.Now()
	minute := now.Truncate(time.Minute).Unix()
	key := fmt.Sprintf("ratelimit:%s:%d", clientID, minute)

	// Get current count from Redis
	countStr, err := cache.Get(ctx, key)
	if err != nil {
		return true, config.RequestsPerMinute, now.Add(time.Minute).Unix(), err
	}

	var count int
	if countStr == "" {
		count = 0
	} else {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			count = 0
		}
	}

	// Calculate reset time (end of current minute)
	resetTime = now.Truncate(time.Minute).Add(time.Minute).Unix()

	// Check if limit exceeded
	if count >= config.RequestsPerMinute {
		return false, 0, resetTime, nil
	}

	// Increment counter
	count++
	if err := cache.Set(ctx, key, strconv.Itoa(count), time.Minute); err != nil {
		// If Redis fails, allow the request
		log.Printf("Failed to increment rate limit counter: %v", err)
		return true, config.RequestsPerMinute - count, resetTime, err
	}

	remaining = config.RequestsPerMinute - count
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetTime, nil
}

// StandardRateLimit applies default rate limiting
func StandardRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return RateLimit(DefaultRateLimit)(next)
}

// StrictRateLimit applies strict rate limiting for AI endpoints
func StrictRateLimit(next http.HandlerFunc) http.HandlerFunc {
	return RateLimit(AIRateLimit)(next)
}