package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/services-api/pkg/auth"
	"github.com/services-api/pkg/config"
	"github.com/services-api/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a JWT manager for testing
func newTestJWTManager() *jwt.Manager {
	return jwt.NewManager("test-secret-key-for-testing-only", 15*time.Minute, 7*24*time.Hour, "test-issuer")
}

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
			jwtManager := newTestJWTManager()
			middleware := auth.NewMiddleware(cfg, jwtManager)

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
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

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
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

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

func TestMiddleware_JWTAuthentication(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"api-key"},
	}
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

	// Generate a valid access token
	token, err := jwtManager.GenerateAccessToken("user123", "test@example.com", "user")
	assert.NoError(t, err)

	var capturedUserID string
	var capturedEmail string
	var capturedRole string
	var userIDFound, emailFound, roleFound bool

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, userIDFound = auth.GetUserID(r.Context())
		capturedEmail, emailFound = auth.GetUserEmail(r.Context())
		capturedRole, roleFound = auth.GetUserRole(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, userIDFound)
	assert.Equal(t, "user123", capturedUserID)
	assert.True(t, emailFound)
	assert.Equal(t, "test@example.com", capturedEmail)
	assert.True(t, roleFound)
	assert.Equal(t, "user", capturedRole)
}

func TestMiddleware_JWTInvalidToken(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"api-key"},
	}
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestMiddleware_JWTRefreshTokenNotAllowed(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"api-key"},
	}
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

	// Generate a refresh token (not an access token)
	refreshToken, err := jwtManager.GenerateRefreshToken("user123", "test@example.com", "user")
	assert.NoError(t, err)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Refresh tokens should not be accepted for authentication
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserID_NotSet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	userID, ok := auth.GetUserID(req.Context())

	assert.False(t, ok)
	assert.Equal(t, "", userID)
}

func TestIsJWTAuth(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"api-key"},
	}
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

	token, err := jwtManager.GenerateAccessToken("user123", "test@example.com", "user")
	assert.NoError(t, err)

	var isJWT bool

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isJWT = auth.IsJWTAuth(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, isJWT)
}

func TestIsAPIKeyAuth(t *testing.T) {
	cfg := &config.Config{
		APIKeys: []string{"api-key"},
	}
	jwtManager := newTestJWTManager()
	middleware := auth.NewMiddleware(cfg, jwtManager)

	var isAPIKey bool

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAPIKey = auth.IsAPIKeyAuth(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Authenticate(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/services", nil)
	req.Header.Set("X-API-Key", "api-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, isAPIKey)
}
