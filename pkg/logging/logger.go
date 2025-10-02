package logging

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/francisco/gridironmind/internal/middleware"
)

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id := ctx.Value(middleware.RequestIDKey); id != nil {
		return id.(string)
	}
	return "unknown"
}

// GetRequestIDFromRequest retrieves the request ID from http.Request
func GetRequestIDFromRequest(r *http.Request) string {
	return GetRequestID(r.Context())
}

// Info logs an info message with request ID
func Info(ctx context.Context, format string, args ...interface{}) {
	requestID := GetRequestID(ctx)
	message := fmt.Sprintf(format, args...)
	log.Printf("[INFO] [%s] %s", requestID, message)
}

// Warn logs a warning message with request ID
func Warn(ctx context.Context, format string, args ...interface{}) {
	requestID := GetRequestID(ctx)
	message := fmt.Sprintf(format, args...)
	log.Printf("[WARN] [%s] %s", requestID, message)
}

// Error logs an error message with request ID
func Error(ctx context.Context, format string, args ...interface{}) {
	requestID := GetRequestID(ctx)
	message := fmt.Sprintf(format, args...)
	log.Printf("[ERROR] [%s] %s", requestID, message)
}

// Debug logs a debug message with request ID
func Debug(ctx context.Context, format string, args ...interface{}) {
	requestID := GetRequestID(ctx)
	message := fmt.Sprintf(format, args...)
	log.Printf("[DEBUG] [%s] %s", requestID, message)
}

// SlowQuery logs a slow database query
func SlowQuery(ctx context.Context, query string, durationMs int64) {
	requestID := GetRequestID(ctx)
	log.Printf("[SLOW-QUERY] [%s] Query took %dms: %s", requestID, durationMs, query)
}

// APICall logs an external API call
func APICall(ctx context.Context, service string, endpoint string, durationMs int64, statusCode int) {
	requestID := GetRequestID(ctx)
	logLevel := "INFO"
	if statusCode >= 500 {
		logLevel = "ERROR"
	} else if statusCode >= 400 {
		logLevel = "WARN"
	}

	log.Printf("[%s] [%s] External API: %s %s - %d (%dms)",
		logLevel,
		requestID,
		service,
		endpoint,
		statusCode,
		durationMs,
	)
}

// CacheHit logs a cache hit
func CacheHit(ctx context.Context, key string) {
	requestID := GetRequestID(ctx)
	log.Printf("[CACHE-HIT] [%s] %s", requestID, key)
}

// CacheMiss logs a cache miss
func CacheMiss(ctx context.Context, key string) {
	requestID := GetRequestID(ctx)
	log.Printf("[CACHE-MISS] [%s] %s", requestID, key)
}

// AutoFetch logs an auto-fetch operation
func AutoFetch(ctx context.Context, resource string, details string) {
	requestID := GetRequestID(ctx)
	log.Printf("[AUTO-FETCH] [%s] %s: %s", requestID, resource, details)
}
