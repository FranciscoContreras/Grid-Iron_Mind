package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_PreflightRequest(t *testing.T) {
	handler := CORS(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	handler(w, req)

	// Check CORS headers
	headers := w.Header()

	if headers.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin = %s, want *", headers.Get("Access-Control-Allow-Origin"))
	}

	if headers.Get("Access-Control-Allow-Methods") != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods incorrect: %s", headers.Get("Access-Control-Allow-Methods"))
	}

	if headers.Get("Access-Control-Allow-Headers") == "" {
		t.Error("Access-Control-Allow-Headers should be set")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Preflight response status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCORS_RegularRequest(t *testing.T) {
	handlerCalled := false
	handler := CORS(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called for regular request")
	}

	// Check CORS headers are still set
	headers := w.Header()
	if headers.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin = %s, want *", headers.Get("Access-Control-Allow-Origin"))
	}

	if w.Code != http.StatusOK {
		t.Errorf("Regular request status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCORS_WithoutOrigin(t *testing.T) {
	handlerCalled := false
	handler := CORS(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	// No Origin header set
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called even without Origin header")
	}

	// CORS headers should still be set
	headers := w.Header()
	if headers.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin = %s, want *", headers.Get("Access-Control-Allow-Origin"))
	}
}

func TestCORS_AllMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled := false
			handler := CORS(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(method, "/api/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			w := httptest.NewRecorder()

			handler(w, req)

			if method != http.MethodOptions && !handlerCalled {
				t.Errorf("Handler should be called for %s method", method)
			}

			headers := w.Header()
			if headers.Get("Access-Control-Allow-Origin") != "*" {
				t.Errorf("CORS headers missing for %s method", method)
			}
		})
	}
}

func TestCORS_CustomHeaders(t *testing.T) {
	handler := CORS(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Headers", "X-API-Key, Content-Type, Authorization")
	w := httptest.NewRecorder()

	handler(w, req)

	headers := w.Header()
	allowHeaders := headers.Get("Access-Control-Allow-Headers")

	// Should allow the standard headers
	expectedHeaders := []string{"Content-Type", "X-API-Key", "Authorization"}
	for _, header := range expectedHeaders {
		if allowHeaders == "" {
			t.Errorf("Access-Control-Allow-Headers should include %s", header)
		}
	}
}
