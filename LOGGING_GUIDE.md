# Request Logging & Tracing Guide

## Overview

Comprehensive request logging and tracing system with unique request IDs, timing metrics, and structured log levels.

## Features

- ✅ **Request ID Tracking** - Every request gets a unique UUID
- ✅ **Request Timing** - Automatic timing for all requests
- ✅ **Response Metrics** - Status codes, response sizes, durations
- ✅ **Slow Request Detection** - Automatic logging of requests >100ms
- ✅ **Structured Logging** - INFO, WARN, ERROR, DEBUG levels
- ✅ **Context Propagation** - Request ID available throughout request lifecycle
- ✅ **Specialized Loggers** - Cache, database, API call tracking

## Log Format

### Standard Request Log
```
[INFO] [550e8400-e29b-41d4-a716-446655440000] GET /api/v1/players from 127.0.0.1:54321
[INFO] [550e8400-e29b-41d4-a716-446655440000] GET /api/v1/players 200 45ms 1024B
```

### Error Request Log
```
[ERROR] [550e8400-e29b-41d4-a716-446655440000] GET /api/v1/players 500 120ms 256B
```

### Slow Request Log
```
[SLOW] [550e8400-e29b-41d4-a716-446655440000] Request took 150ms: GET /api/v1/players
```

## Response Headers

Every response includes:
```
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

Clients can use this ID to:
1. Correlate logs across systems
2. Report issues with specific requests
3. Debug distributed traces

## Usage in Handlers

### Basic Logging

```go
import "github.com/francisco/gridironmind/pkg/logging"

func (h *PlayersHandler) listPlayers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    logging.Info(ctx, "Listing players with filters: %v", filters)

    // Your handler logic...

    logging.Info(ctx, "Successfully returned %d players", len(players))
}
```

### Error Logging

```go
players, err := h.queries.ListPlayers(ctx, filters)
if err != nil {
    logging.Error(ctx, "Failed to query players: %v", err)
    response.InternalError(w, "Failed to retrieve players")
    return
}
```

### Warning Logging

```go
if total == 0 {
    logging.Warn(ctx, "No players found for filters: %v", filters)
}
```

### Debug Logging

```go
logging.Debug(ctx, "Cache key: %s, TTL: %d", cacheKey, ttl)
```

## Specialized Logging

### Database Queries

```go
import "github.com/francisco/gridironmind/pkg/logging"

start := time.Now()
rows, err := pool.Query(ctx, query, args...)
duration := time.Since(start).Milliseconds()

if duration > 100 {
    logging.SlowQuery(ctx, query, duration)
}
```

### External API Calls

```go
start := time.Now()
resp, err := http.Get(url)
duration := time.Since(start).Milliseconds()

logging.APICall(ctx, "ESPN", endpoint, duration, resp.StatusCode)
```

### Cache Operations

```go
cached, err := cache.Get(ctx, key)
if err == nil && cached != "" {
    logging.CacheHit(ctx, key)
    // Return cached data
} else {
    logging.CacheMiss(ctx, key)
    // Fetch from database
}
```

### Auto-Fetch Operations

```go
if total == 0 && shouldAutoFetch {
    logging.AutoFetch(ctx, "games", fmt.Sprintf("season %d week %d", season, week))

    // Perform fetch...
}
```

## Log Levels

### INFO
- Normal operations
- Successful requests (200-399)
- General information

```
[INFO] [request-id] Player retrieved: id=123
```

### WARN
- Client errors (400-499)
- Recoverable issues
- Missing data

```
[WARN] [request-id] Invalid position filter: INVALID
```

### ERROR
- Server errors (500-599)
- Database errors
- Panics
- Failed operations

```
[ERROR] [request-id] Database query failed: connection timeout
```

### DEBUG
- Development information
- Detailed execution flow
- Variable dumps

```
[DEBUG] [request-id] Cache key generated: player:123:2025
```

## Special Tags

### [SLOW]
Requests or queries taking >100ms

```
[SLOW] [request-id] Request took 150ms: GET /api/v1/players
```

### [SLOW-QUERY]
Database queries >100ms

```
[SLOW-QUERY] [request-id] Query took 200ms: SELECT * FROM players WHERE position = 'QB'
```

### [CACHE-HIT]
Cache hit operations

```
[CACHE-HIT] [request-id] players:list:QB:50:0
```

### [CACHE-MISS]
Cache miss operations

```
[CACHE-MISS] [request-id] players:list:WR:25:50
```

### [AUTO-FETCH]
Automatic data fetching

```
[AUTO-FETCH] [request-id] games: Fetching season 2025 week 5
```

## Request ID Retrieval

### From Context

```go
import "github.com/francisco/gridironmind/pkg/logging"

func someFunction(ctx context.Context) {
    requestID := logging.GetRequestID(ctx)
    fmt.Printf("Processing request: %s\n", requestID)
}
```

### From HTTP Request

```go
func (h *Handler) someMethod(w http.ResponseWriter, r *http.Request) {
    requestID := logging.GetRequestIDFromRequest(r)
    // Use request ID...
}
```

## Timing Metrics

### Automatic Timing
All requests are automatically timed:

```
[INFO] [request-id] GET /api/v1/players 200 45ms 1024B
                                         ↑   ↑    ↑
                                    status  time  bytes
```

### Manual Timing

```go
import "time"

start := time.Now()
// Operation...
duration := time.Since(start).Milliseconds()

logging.Info(ctx, "Operation completed in %dms", duration)
```

## Integration with Middleware

### Middleware Stack

```go
func applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return middleware.CORS(
        middleware.LogRequest(        // Adds request ID + timing
            middleware.RecoverPanic(   // Uses request ID in panic logs
                middleware.StandardRateLimit(handler),
            ),
        ),
    )
}
```

### Order Matters

1. **CORS** - First to set headers
2. **LogRequest** - Generates request ID and timing
3. **RecoverPanic** - Uses request ID for panic logs
4. **Auth/RateLimit** - Can use request ID for tracking
5. **Handler** - Has access to request ID via context

## Log Analysis

### Finding Specific Request

```bash
# Heroku logs
heroku logs --tail | grep "550e8400-e29b-41d4-a716-446655440000"

# Local logs
cat app.log | grep "550e8400-e29b-41d4-a716-446655440000"
```

### Finding Slow Requests

```bash
heroku logs --tail | grep "\[SLOW\]"
```

### Finding Errors

```bash
heroku logs --tail | grep "\[ERROR\]"
```

### Finding Cache Performance

```bash
heroku logs --tail | grep -E "\[CACHE-(HIT|MISS)\]"
```

## Best Practices

### 1. Always Use Context Logging

✅ **Good:**
```go
logging.Info(ctx, "Player created: %s", player.Name)
```

❌ **Bad:**
```go
log.Printf("Player created: %s", player.Name)
```

### 2. Include Relevant Details

✅ **Good:**
```go
logging.Error(ctx, "Failed to create player: %v, data: %+v", err, playerData)
```

❌ **Bad:**
```go
logging.Error(ctx, "Error")
```

### 3. Use Appropriate Log Levels

✅ **Good:**
```go
if player == nil {
    logging.Warn(ctx, "Player not found: %s", id)
    response.NotFound(w, "Player")
    return
}
```

❌ **Bad:**
```go
if player == nil {
    logging.Error(ctx, "Player not found: %s", id)  // Not an error!
}
```

### 4. Log Before and After Critical Operations

```go
logging.Info(ctx, "Starting sync for season %d", season)

if err := syncGames(ctx, season); err != nil {
    logging.Error(ctx, "Sync failed: %v", err)
    return err
}

logging.Info(ctx, "Sync completed successfully")
```

### 5. Never Log Sensitive Data

❌ **Never log:**
- API keys
- Passwords
- Authentication tokens
- Personal information (PII)
- Credit card numbers

✅ **Safe logging:**
```go
logging.Info(ctx, "User authenticated: %s", userID)  // Log ID, not credentials
```

## Performance Considerations

### Logging Overhead

- Each log entry: ~0.1ms
- Request ID generation: ~0.01ms
- Response wrapping: ~0.001ms

**Total middleware overhead: <1ms per request**

### Production Optimization

For high-traffic scenarios:

```go
// Only log errors and warnings in production
if os.Getenv("ENVIRONMENT") == "production" {
    if statusCode >= 400 {
        // Log only errors/warnings
    }
} else {
    // Log everything in development
}
```

## Monitoring Integration

### Log Aggregation Services

The structured log format works with:
- **Papertrail** - Heroku add-on
- **Loggly** - External service
- **Datadog** - APM platform
- **Splunk** - Enterprise logging

### Example Papertrail Search

```
[ERROR] [*] *           # All errors
[SLOW] [*] * >100ms     # All slow requests
[*] [request-id] *      # Trace single request
```

## Testing

### Unit Tests

```go
func TestLoggingWithRequestID(t *testing.T) {
    requestID := "test-id-123"
    ctx := context.WithValue(context.Background(), middleware.RequestIDKey, requestID)

    // Logging should not panic
    logging.Info(ctx, "Test message")
    logging.Error(ctx, "Test error")
}
```

### Integration Tests

```go
func TestRequestIDPropagation(t *testing.T) {
    handler := middleware.LogRequest(func(w http.ResponseWriter, r *http.Request) {
        requestID := logging.GetRequestIDFromRequest(r)
        if requestID == "" {
            t.Error("Request ID should be set")
        }
    })

    req := httptest.NewRequest(http.MethodGet, "/test", nil)
    w := httptest.NewRecorder()

    handler(w, req)

    responseID := w.Header().Get("X-Request-ID")
    if responseID == "" {
        t.Error("X-Request-ID header should be set")
    }
}
```

## Migration from Old Logging

### Before (Old Code)

```go
log.Printf("Error getting player: %v", err)
log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
```

### After (New Code)

```go
logging.Error(ctx, "Failed to get player: %v", err)
// Request logging handled automatically by middleware
```

### Search and Replace

```bash
# Find old logging patterns
grep -r "log.Printf" internal/

# Replace with context logging
# log.Printf("Error: %v", err)
# → logging.Error(ctx, "Error: %v", err)
```

## Configuration

### Environment Variables

```bash
# Set log level (future enhancement)
LOG_LEVEL=INFO  # DEBUG, INFO, WARN, ERROR

# Enable/disable slow request logging
LOG_SLOW_REQUESTS=true
SLOW_REQUEST_THRESHOLD_MS=100

# Enable/disable request ID in logs
LOG_REQUEST_ID=true
```

## Example Complete Handler

```go
package handlers

import (
    "net/http"
    "time"

    "github.com/francisco/gridironmind/pkg/logging"
    "github.com/francisco/gridironmind/pkg/response"
)

func (h *PlayersHandler) listPlayers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Parse filters
    filters := parseFilters(r)
    logging.Debug(ctx, "Parsed filters: %+v", filters)

    // Check cache
    cacheKey := generateCacheKey(filters)
    if cached, err := cache.Get(ctx, cacheKey); err == nil {
        logging.CacheHit(ctx, cacheKey)
        w.Write([]byte(cached))
        return
    }
    logging.CacheMiss(ctx, cacheKey)

    // Query database
    start := time.Now()
    players, total, err := h.queries.ListPlayers(ctx, filters)
    queryDuration := time.Since(start).Milliseconds()

    if queryDuration > 100 {
        logging.SlowQuery(ctx, "ListPlayers", queryDuration)
    }

    if err != nil {
        logging.Error(ctx, "Database query failed: %v", err)
        response.InternalError(w, "Failed to retrieve players")
        return
    }

    logging.Info(ctx, "Retrieved %d players (total: %d)", len(players), total)

    // Cache result
    cache.Set(ctx, cacheKey, responseJSON, 15*time.Minute)

    response.SuccessWithPagination(w, players, total, filters.Limit, filters.Offset)
}
```

## Conclusion

The request logging and tracing system provides:
- ✅ Complete request lifecycle visibility
- ✅ Easy debugging with request IDs
- ✅ Performance monitoring (timing, slow requests)
- ✅ Structured logs for analysis
- ✅ Context propagation throughout the stack

Use the logging utilities consistently for better observability and easier debugging in production.
