package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	Success(w, data)

	if w.Code != http.StatusOK {
		t.Errorf("Success() status = %d, want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Success() Content-Type = %s, want application/json", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"message":"test"`) {
		t.Errorf("Success() body = %s, want to contain message:test", body)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()

	Error(w, http.StatusBadRequest, "TEST_ERROR", "Test error message")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Error() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"code":"TEST_ERROR"`) {
		t.Errorf("Error() body should contain error code")
	}
	if !strings.Contains(body, `"message":"Test error message"`) {
		t.Errorf("Error() body should contain error message")
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()

	NotFound(w, "Player")

	if w.Code != http.StatusNotFound {
		t.Errorf("NotFound() status = %d, want %d", w.Code, http.StatusNotFound)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Player not found") {
		t.Errorf("NotFound() body should contain 'Player not found'")
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()

	BadRequest(w, "Invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("BadRequest() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestInternalError(t *testing.T) {
	w := httptest.NewRecorder()

	InternalError(w, "Database error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("InternalError() status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()

	Unauthorized(w, "Invalid token")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Unauthorized() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestSuccessWithPagination(t *testing.T) {
	w := httptest.NewRecorder()
	data := []map[string]string{{"id": "1"}, {"id": "2"}}

	SuccessWithPagination(w, data, 100, 25, 0)

	if w.Code != http.StatusOK {
		t.Errorf("SuccessWithPagination() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"total":100`) {
		t.Errorf("SuccessWithPagination() should contain total count")
	}
	if !strings.Contains(body, `"limit":25`) {
		t.Errorf("SuccessWithPagination() should contain limit")
	}
	if !strings.Contains(body, `"offset":0`) {
		t.Errorf("SuccessWithPagination() should contain offset")
	}
}
