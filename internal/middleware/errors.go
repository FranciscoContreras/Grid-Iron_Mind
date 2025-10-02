package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/francisco/gridironmind/pkg/response"
	"github.com/google/uuid"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// StartTimeKey is the context key for request start time
	StartTimeKey ContextKey = "start_time"
)

// loggingResponseWriter wraps http.ResponseWriter to capture status code and bytes written
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (rw *loggingResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// RecoverPanic recovers from panics and returns a 500 error
func RecoverPanic(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestID(r)
				log.Printf("[ERROR] [%s] Panic recovered: %v", requestID, err)
				response.InternalError(w, "An unexpected error occurred")
			}
		}()
		next(w, r)
	}
}

// LogRequest logs all incoming requests with request ID and timing
func LogRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID
		requestID := uuid.New().String()
		startTime := time.Now()

		// Add request ID and start time to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		ctx = context.WithValue(ctx, StartTimeKey, startTime)
		r = r.WithContext(ctx)

		// Wrap response writer to capture status code
		rw := &loggingResponseWriter{ResponseWriter: w, statusCode: 0}

		// Add request ID to response headers
		rw.Header().Set("X-Request-ID", requestID)

		// Log request start
		log.Printf("[INFO] [%s] %s %s from %s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		// Execute handler
		next(rw, r)

		// Log request completion with timing and status
		duration := time.Since(startTime)
		statusCode := rw.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}

		logLevel := "INFO"
		if statusCode >= 500 {
			logLevel = "ERROR"
		} else if statusCode >= 400 {
			logLevel = "WARN"
		}

		log.Printf("[%s] [%s] %s %s %d %dms %dB",
			logLevel,
			requestID,
			r.Method,
			r.URL.Path,
			statusCode,
			duration.Milliseconds(),
			rw.written,
		)

		// Log slow requests (>100ms)
		if duration.Milliseconds() > 100 {
			log.Printf("[SLOW] [%s] Request took %dms: %s %s",
				requestID,
				duration.Milliseconds(),
				r.Method,
				r.URL.Path,
			)
		}
	}
}

// getRequestID retrieves the request ID from context
func getRequestID(r *http.Request) string {
	if id := r.Context().Value(RequestIDKey); id != nil {
		return id.(string)
	}
	return "unknown"
}