# Error Handling Improvements - Phase 3

## Current Issues

### 1. Inconsistent Error Response Format

**Problem:** Mix of `http.Error()` and `response.Error()` usage:

```go
// Inconsistent - admin.go still uses http.Error
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

// Consistent - uses structured response
response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
```

**Impact:**
- Admin endpoints return plain text errors
- Other endpoints return structured JSON errors
- Client applications need different error parsing logic

### 2. Missing Error Logging

**Current State:**
- Only 15 error log statements across 4 handler files
- 97 error responses, but only ~15% are logged
- Database errors, validation errors often go unlogged

**Example (Good):**
```go
if err := h.queries.GetGameStats(ctx, gameID); err != nil {
    log.Printf("Error getting game stats for %s: %v", gameID, err)
    response.InternalError(w, "Failed to fetch game stats")
    return
}
```

**Example (Missing Logging):**
```go
if err != nil {
    response.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid player ID")
    return
}
```

### 3. Inconsistent Error Codes

**Examples:**
- `"QUERY_FAILED"` vs `"INTERNAL_ERROR"` for database errors
- `"INVALID_PLAYER_ID"` vs `"INVALID_ID"` vs `"BAD_REQUEST"` for validation
- `"SYNC_FAILED"` vs `"INTERNAL_ERROR"` for sync operations

### 4. No Error Context

**Current:** Generic error messages without details
```go
response.InternalError(w, "Failed to fetch game stats")
```

**Missing:** Request context (endpoint, params, user info)

### 5. Error Response Inconsistency

Some handlers use different helper functions:
- `response.Error()` - Full control
- `response.BadRequest()` - Convenience helper
- `response.NotFound()` - Convenience helper
- `response.InternalError()` - Convenience helper
- `http.Error()` - Legacy plain text (admin.go)

## Proposed Solutions

### 1. Standardize Error Responses

**Action:** Replace all `http.Error()` with `response.Error()` or helpers

**Before:**
```go
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
```

**After:**
```go
response.Error(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
```

### 2. Add Structured Error Logging

**Create Error Logging Helper:**

```go
// pkg/response/logging.go
func LogAndError(w http.ResponseWriter, r *http.Request, status int, code, message string, err error) {
    // Log with context
    log.Printf("[ERROR] %s %s - %s: %v", r.Method, r.URL.Path, code, err)

    // Return structured error
    Error(w, status, code, message)
}

func LogAndBadRequest(w http.ResponseWriter, r *http.Request, message string, err error) {
    log.Printf("[ERROR] %s %s - BAD_REQUEST: %v", r.Method, r.URL.Path, err)
    BadRequest(w, message)
}

func LogAndNotFound(w http.ResponseWriter, r *http.Request, resource string) {
    log.Printf("[WARN] %s %s - NOT_FOUND: %s", r.Method, r.URL.Path, resource)
    NotFound(w, resource)
}

func LogAndInternalError(w http.ResponseWriter, r *http.Request, message string, err error) {
    log.Printf("[ERROR] %s %s - INTERNAL_ERROR: %v", r.Method, r.URL.Path, err)
    InternalError(w, message)
}
```

### 3. Standardize Error Codes

**Error Code Guidelines:**

| Category | Code | Usage |
|----------|------|-------|
| **Validation** | `INVALID_PARAMETER` | Query param validation failed |
| | `INVALID_ID` | UUID/ID parsing failed |
| | `MISSING_PARAMETER` | Required param missing |
| | `INVALID_REQUEST_BODY` | JSON decode failed |
| **Resources** | `NOT_FOUND` | Resource doesn't exist |
| | `ALREADY_EXISTS` | Resource conflict |
| **Database** | `DATABASE_ERROR` | Database query failed |
| | `QUERY_TIMEOUT` | Query exceeded timeout |
| **External** | `EXTERNAL_API_ERROR` | External API call failed |
| | `SYNC_FAILED` | Data sync operation failed |
| **Auth** | `UNAUTHORIZED` | Missing/invalid API key |
| | `FORBIDDEN` | Insufficient permissions |
| **Rate Limiting** | `RATE_LIMIT_EXCEEDED` | Too many requests |
| **Server** | `INTERNAL_ERROR` | Unexpected server error |
| | `SERVICE_UNAVAILABLE` | Service temporarily unavailable |

### 4. Error Context Enhancement

**Approach:** Include request context in logs

```go
// Before
log.Printf("Error getting game stats: %v", err)

// After
log.Printf("[ERROR] GET /api/v1/stats/game/%s - DATABASE_ERROR: %v", gameID, err)
```

### 5. Centralized Error Handling Middleware

**Optional Enhancement:**

```go
// internal/middleware/errors.go
func ErrorRecovery(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("[PANIC] %s %s - %v\n%s", r.Method, r.URL.Path, err, debug.Stack())
                response.InternalError(w, "An unexpected error occurred")
            }
        }()
        next(w, r)
    }
}
```

## Implementation Plan

### Phase 1: Add Logging Helpers (High Priority)
- [ ] Create `pkg/response/logging.go`
- [ ] Add `LogAndError()`, `LogAndBadRequest()`, etc.
- [ ] Add tests for logging functions

### Phase 2: Fix admin.go Inconsistency (High Priority)
- [ ] Replace 9 instances of `http.Error()` with `response.Error()`
- [ ] Ensure all admin errors return structured JSON

### Phase 3: Add Error Logging (Medium Priority)
- [ ] Update handlers to use `LogAndError()` helpers
- [ ] Focus on database errors first
- [ ] Then validation errors
- [ ] Then external API errors

### Phase 4: Standardize Error Codes (Medium Priority)
- [ ] Document error code standards
- [ ] Update existing handlers to use standard codes
- [ ] Create error code constants (optional)

### Phase 5: Enhanced Error Context (Low Priority)
- [ ] Add request ID tracking
- [ ] Include user/API key in logs (if present)
- [ ] Add error correlation IDs

## Success Metrics

- **0** instances of `http.Error()` in handlers (currently 9)
- **100%** error logging coverage for database errors
- **100%** error logging coverage for external API errors
- **Consistent** error code usage across all endpoints
- **Structured** JSON error responses on all endpoints

## Benefits

1. **Better Debugging:** All errors logged with context
2. **Consistent API:** All endpoints return same error format
3. **Easier Monitoring:** Structured logs enable alerting
4. **Better UX:** Clients get predictable error responses
5. **Audit Trail:** All errors tracked for analysis
