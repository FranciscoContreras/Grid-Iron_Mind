package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverPanic_NoPanic(t *testing.T) {
	handlerCalled := false
	handler := RecoverPanic(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called when no panic occurs")
	}

	if w.Code != http.StatusOK {
		t.Errorf("RecoverPanic() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRecoverPanic_WithPanic(t *testing.T) {
	handler := RecoverPanic(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	// Should not panic - should recover gracefully
	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("RecoverPanic() status = %d, want %d after panic", w.Code, http.StatusInternalServerError)
	}

	// Check that error response was written
	body := w.Body.String()
	if body == "" {
		t.Error("RecoverPanic() should write error response after panic")
	}
}

func TestRecoverPanic_WithNilPanic(t *testing.T) {
	handler := RecoverPanic(func(w http.ResponseWriter, r *http.Request) {
		panic(nil)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	// Should handle nil panic
	handler(w, req)

	// Should still return 500 error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("RecoverPanic() status = %d, want %d for nil panic", w.Code, http.StatusInternalServerError)
	}
}

func TestLogRequest(t *testing.T) {
	handlerCalled := false
	handler := LogRequest(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called after logging")
	}

	if w.Code != http.StatusOK {
		t.Errorf("LogRequest() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Check request ID header is set
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestLogRequest_AllMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			handlerCalled := false
			handler := LogRequest(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(method, "/api/test", nil)
			w := httptest.NewRecorder()

			handler(w, req)

			if !handlerCalled {
				t.Errorf("Handler should be called for %s method", method)
			}
		})
	}
}

func TestMiddlewareChaining(t *testing.T) {
	// Test that middleware can be chained together
	handlerCalled := false
	handler := RecoverPanic(
		LogRequest(
			func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			},
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called through middleware chain")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Chained middleware status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMiddlewareChaining_PanicRecovery(t *testing.T) {
	// Test that panic recovery works when chained
	handler := RecoverPanic(
		LogRequest(
			func(w http.ResponseWriter, r *http.Request) {
				panic("chained panic")
			},
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	// Should not panic
	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Chained panic recovery status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
