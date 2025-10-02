package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	// Set up test API key
	os.Setenv("API_KEY", "test-api-key-123")
	defer os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := APIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-API-Key", "test-api-key-123")
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called with valid API key")
	}

	if w.Code != http.StatusOK {
		t.Errorf("APIKeyAuth() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPIKeyAuth_ValidKeyBearer(t *testing.T) {
	os.Setenv("API_KEY", "test-api-key-123")
	defer os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := APIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer test-api-key-123")
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called with valid Bearer token")
	}

	if w.Code != http.StatusOK {
		t.Errorf("APIKeyAuth() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	os.Setenv("API_KEY", "test-api-key-123")
	defer os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := APIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if handlerCalled {
		t.Error("Handler should not be called with invalid API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("APIKeyAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	os.Setenv("API_KEY", "test-api-key-123")
	defer os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := APIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if handlerCalled {
		t.Error("Handler should not be called with missing API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("APIKeyAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAPIKeyAuth_NoConfiguredKey(t *testing.T) {
	// Ensure no API key is set
	os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := APIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called when no API key configured (dev mode)")
	}

	if w.Code != http.StatusOK {
		t.Errorf("APIKeyAuth() status = %d, want %d (dev mode)", w.Code, http.StatusOK)
	}
}

func TestOptionalAPIKeyAuth_NoKey(t *testing.T) {
	os.Setenv("API_KEY", "test-api-key-123")
	defer os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := OptionalAPIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called even without API key (optional)")
	}

	if w.Code != http.StatusOK {
		t.Errorf("OptionalAPIKeyAuth() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestOptionalAPIKeyAuth_InvalidKey(t *testing.T) {
	os.Setenv("API_KEY", "test-api-key-123")
	defer os.Unsetenv("API_KEY")

	handlerCalled := false
	handler := OptionalAPIKeyAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if handlerCalled {
		t.Error("Handler should not be called with invalid API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("OptionalAPIKeyAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAdminAuth_ValidKey(t *testing.T) {
	os.Setenv("API_KEY", "admin-key-123")
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("ENVIRONMENT")

	handlerCalled := false
	handler := AdminAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/teams", nil)
	req.Header.Set("X-API-Key", "admin-key-123")
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Admin handler should be called with valid API key")
	}

	if w.Code != http.StatusOK {
		t.Errorf("AdminAuth() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAdminAuth_InvalidKey(t *testing.T) {
	os.Setenv("API_KEY", "admin-key-123")
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("ENVIRONMENT")

	handlerCalled := false
	handler := AdminAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/teams", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if handlerCalled {
		t.Error("Admin handler should not be called with invalid API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AdminAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAdminAuth_NoKeyProduction(t *testing.T) {
	os.Unsetenv("API_KEY")
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")

	handlerCalled := false
	handler := AdminAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/teams", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if handlerCalled {
		t.Error("Admin handler should not be called in production with no API key configured")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AdminAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAdminAuth_NoKeyDevelopment(t *testing.T) {
	os.Unsetenv("API_KEY")
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")

	handlerCalled := false
	handler := AdminAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/teams", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("Admin handler should be called in development mode with no API key")
	}

	if w.Code != http.StatusOK {
		t.Errorf("AdminAuth() status = %d, want %d (dev mode)", w.Code, http.StatusOK)
	}
}

func TestAdminAuth_MissingKey(t *testing.T) {
	os.Setenv("API_KEY", "admin-key-123")
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("API_KEY")
	defer os.Unsetenv("ENVIRONMENT")

	handlerCalled := false
	handler := AdminAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/teams", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if handlerCalled {
		t.Error("Admin handler should not be called with missing API key")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AdminAuth() status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{"Equal strings", "secret123", "secret123", true},
		{"Different strings", "secret123", "secret456", false},
		{"Different lengths", "secret", "secret123", false},
		{"Empty strings", "", "", true},
		{"One empty", "secret", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constantTimeCompare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("constantTimeCompare(%q, %q) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
