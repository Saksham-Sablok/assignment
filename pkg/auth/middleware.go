package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/services-api/pkg/config"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// APIKeyHeader is the header name for API key authentication
	APIKeyHeader = "X-API-Key"
	// APIKeyContextKey is the context key for storing API key identifier
	APIKeyContextKey ContextKey = "api_key_id"
)

// Middleware creates an authentication middleware
type Middleware struct {
	config *config.Config
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{
		config: cfg,
	}
}

// Authenticate is the middleware function that validates API keys
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key
		apiKey := r.Header.Get(APIKeyHeader)
		if apiKey == "" {
			writeUnauthorizedResponse(w, "authentication required")
			return
		}

		// Validate API key
		if !m.config.HasAPIKeys() {
			// No API keys configured, reject all requests
			writeUnauthorizedResponse(w, "no API keys configured")
			return
		}

		// Find matching key
		keyIndex := m.findValidKeyIndex(apiKey)
		if keyIndex == -1 {
			writeUnauthorizedResponse(w, "invalid credentials")
			return
		}

		// Add key identifier to context for audit logging
		ctx := context.WithValue(r.Context(), APIKeyContextKey, keyIndex)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// findValidKeyIndex returns the index of the matching key, or -1 if not found
// Uses constant-time comparison to prevent timing attacks
func (m *Middleware) findValidKeyIndex(providedKey string) int {
	for i, configuredKey := range m.config.APIKeys {
		if secureCompare(providedKey, configuredKey) {
			return i
		}
	}
	return -1
}

// secureCompare performs a constant-time comparison of two strings
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// writeUnauthorizedResponse writes a 401 response with proper format
func writeUnauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	// Don't include any sensitive information in error response
	w.Write([]byte(`{"error":"unauthorized","message":"` + escapeJSON(message) + `"}`))
}

// escapeJSON escapes special characters in JSON strings
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// GetAPIKeyID retrieves the API key identifier from the request context
func GetAPIKeyID(ctx context.Context) (int, bool) {
	keyID, ok := ctx.Value(APIKeyContextKey).(int)
	return keyID, ok
}
