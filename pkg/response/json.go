package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Meta contains metadata about the response
type Meta struct {
	Timestamp string `json:"timestamp"`
	Total     *int   `json:"total,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
	Offset    *int   `json:"offset,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// JSON writes a JSON response with the given status code
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success writes a successful response with data
func Success(w http.ResponseWriter, data interface{}) {
	response := SuccessResponse{
		Data: data,
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	JSON(w, http.StatusOK, response)
}

// SuccessWithPagination writes a successful response with pagination metadata
func SuccessWithPagination(w http.ResponseWriter, data interface{}, total, limit, offset int) {
	response := SuccessResponse{
		Data: data,
		Meta: Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Total:     &total,
			Limit:     &limit,
			Offset:    &offset,
		},
	}
	JSON(w, http.StatusOK, response)
}

// Error writes an error response
func Error(w http.ResponseWriter, status int, code, message string) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
		},
	}
	JSON(w, status, response)
}

// NotFound writes a 404 error response
func NotFound(w http.ResponseWriter, resource string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", resource+" not found")
}

// BadRequest writes a 400 error response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// InternalError writes a 500 error response
func InternalError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// Unauthorized writes a 401 error response
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// TooManyRequests writes a 429 error response with Retry-After header
func TooManyRequests(w http.ResponseWriter, retryAfter int) {
	w.Header().Set("Retry-After", string(rune(retryAfter)))
	Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")
}