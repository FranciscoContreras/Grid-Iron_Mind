package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/francisco/gridironmind/internal/middleware"
)

func TestGetRequestID_WithContext(t *testing.T) {
	requestID := "test-request-id-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	result := GetRequestID(ctx)

	if result != requestID {
		t.Errorf("GetRequestID() = %s, want %s", result, requestID)
	}
}

func TestGetRequestID_WithoutContext(t *testing.T) {
	ctx := context.Background()

	result := GetRequestID(ctx)

	if result != "unknown" {
		t.Errorf("GetRequestID() = %s, want 'unknown'", result)
	}
}

func TestGetRequestIDFromRequest(t *testing.T) {
	requestID := "test-request-id-456"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(ctx)

	result := GetRequestIDFromRequest(req)

	if result != requestID {
		t.Errorf("GetRequestIDFromRequest() = %s, want %s", result, requestID)
	}
}

func TestInfo(t *testing.T) {
	requestID := "info-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	Info(ctx, "Test info message: %s", "value")
}

func TestWarn(t *testing.T) {
	requestID := "warn-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	Warn(ctx, "Test warning: %d", 42)
}

func TestError(t *testing.T) {
	requestID := "error-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	Error(ctx, "Test error: %v", "error message")
}

func TestDebug(t *testing.T) {
	requestID := "debug-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	Debug(ctx, "Test debug: %s", "debug info")
}

func TestSlowQuery(t *testing.T) {
	requestID := "query-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	SlowQuery(ctx, "SELECT * FROM players", 150)
}

func TestAPICall(t *testing.T) {
	tests := []struct {
		name       string
		service    string
		endpoint   string
		statusCode int
		duration   int64
	}{
		{"Success call", "ESPN", "/teams", 200, 50},
		{"Client error", "ESPN", "/players", 404, 30},
		{"Server error", "ESPN", "/games", 500, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestID := "api-test-123"
			ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

			// Should not panic
			APICall(ctx, tt.service, tt.endpoint, tt.duration, tt.statusCode)
		})
	}
}

func TestCacheHit(t *testing.T) {
	requestID := "cache-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	CacheHit(ctx, "player:123")
}

func TestCacheMiss(t *testing.T) {
	requestID := "cache-test-456"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	CacheMiss(ctx, "player:456")
}

func TestAutoFetch(t *testing.T) {
	requestID := "autofetch-test-123"
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

	// Should not panic
	AutoFetch(ctx, "games", "Fetching season 2025 week 5")
}

func TestLogging_WithoutRequestID(t *testing.T) {
	ctx := context.Background()

	// All logging functions should handle missing request ID gracefully
	Info(ctx, "Test without request ID")
	Warn(ctx, "Test without request ID")
	Error(ctx, "Test without request ID")
	Debug(ctx, "Test without request ID")
	SlowQuery(ctx, "SELECT 1", 200)
	APICall(ctx, "Service", "/endpoint", 100, 200)
	CacheHit(ctx, "key")
	CacheMiss(ctx, "key")
	AutoFetch(ctx, "resource", "details")
}
