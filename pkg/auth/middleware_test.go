package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/services-api/pkg/auth"
	"github.com/services-api/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware_Authenticate(t *testing.T) {
	tests := []struct {
		name           string
		apiKeys        []string
		requestPath    string
		requestAPIKey  string
		expectedStatus int
		expectNext     bool
	}{
		{
			name:           "valid API key",
			apiKeys:        []string{"valid-key-1", "valid-key-2"},
			requestPath:    "/api/v1/services",
			requestAPIKey:  "valid-key-1",
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "second valid API key",
			apiKeys:        []string{"valid-key-1", "valid-key-2"},
			requestPath:    "/api/v1/services",
			requestAPIKey:  "valid-key-2",
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "invalid API key",
			apiKeys:        []string{"valid-key-1"},
			requestPath:    "/api/v1/services",
			requestAPIKey:  "invalid-key",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "missing API key",
			apiKeys:        []string{"valid-key-1"},
			requestPath:    "/api/v1/services",
			requestAPIKey:  "",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
		{
			name:           "health endpoint - no auth required",
			apiKeys:        []string{"valid-key-1"},
			requestPath:    "/health",
			requestAPIKey:  "",
			expectedStatus: http.StatusOK,
			expectNext:     true,
		},
		{
			name:           "no API keys configured",
			apiKeys:        []string{},
			requestPath:    "/api/v1/services",
			requestAPIKey:  "any-key",
			expectedStatus: http.StatusUnauthorized,
			expectNext:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := &config.Config{
				APIKeys: tt.apiKeys,
			}
			middleware := auth.NewMiddleware(cfg)

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.Authenticate(nextHandler)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			if tt.requestAPIKey != "" {
				req.Header.Set("X-API-Key", tt.requestAPIKey)
			}
			w := httptest.NewRecorder()

			// Execute
			handler.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectNext, nextCalled)
		})
	}
}

func TestMiddleware_APIKeyInContext(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"key-0", "key-1", "key-2"},
	}
	middleware := auth.NewMiddleware(cfg)

	var capturedKeyID int
	var keyFound bool

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedKeyID, keyFound = auth.GetAPIKeyID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	// Test with the second key (index 1)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("X-API-Key", "key-1")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, keyFound)
	assert.Equal(t, 1, capturedKeyID)
}

func TestGetAPIKeyID_NotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	keyID, ok := auth.GetAPIKeyID(req.Context())

	assert.False(t, ok)
	assert.Equal(t, 0, keyID)
}

func TestMiddleware_UnauthorizedResponseFormat(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"valid-key"},
	}
	middleware := auth.NewMiddleware(cfg)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "unauthorized")
	assert.Contains(t, w.Body.String(), "invalid credentials")
}
