package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/services-api/pkg/config"
	"github.com/services-api/pkg/jwt"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// APIKeyHeader is the header name for API key authentication
	APIKeyHeader = "X-API-Key"
	// AuthorizationHeader is the header name for Bearer token authentication
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for Bearer tokens
	BearerPrefix = "Bearer "

	// Context keys
	APIKeyContextKey ContextKey = "api_key_id"
	UserIDContextKey ContextKey = "user_id"
	UserEmailKey     ContextKey = "user_email"
	UserRoleKey      ContextKey = "user_role"
	AuthTypeKey      ContextKey = "auth_type"
)

// AuthType represents the type of authentication used
type AuthType string

const (
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeJWT    AuthType = "jwt"
)

// Middleware creates an authentication middleware
type Middleware struct {
	config     *config.Config
	jwtManager *jwt.Manager
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(cfg *config.Config, jwtManager *jwt.Manager) *Middleware {
	return &Middleware{
		config:     cfg,
		jwtManager: jwtManager,
	}
}

// Authenticate is the middleware function that validates API keys or JWT tokens
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health endpoint
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Try JWT authentication first (Authorization: Bearer <token>)
		authHeader := r.Header.Get(AuthorizationHeader)
		if strings.HasPrefix(authHeader, BearerPrefix) {
			token := strings.TrimPrefix(authHeader, BearerPrefix)
			if ctx, ok := m.authenticateJWT(r.Context(), token); ok {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			writeUnauthorizedResponse(w, "invalid or expired token")
			return
		}

		// Try API key authentication (X-API-Key header)
		apiKey := r.Header.Get(APIKeyHeader)
		if apiKey != "" {
			if ctx, ok := m.authenticateAPIKey(r.Context(), apiKey); ok {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			writeUnauthorizedResponse(w, "invalid credentials")
			return
		}

		// No authentication provided
		writeUnauthorizedResponse(w, "authentication required")
	})
}

// authenticateJWT validates a JWT token and returns an updated context
func (m *Middleware) authenticateJWT(ctx context.Context, token string) (context.Context, bool) {
	if m.jwtManager == nil {
		return ctx, false
	}

	claims, err := m.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return ctx, false
	}

	// Add user info to context
	ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
	ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
	ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
	ctx = context.WithValue(ctx, AuthTypeKey, AuthTypeJWT)

	return ctx, true
}

// authenticateAPIKey validates an API key and returns an updated context
func (m *Middleware) authenticateAPIKey(ctx context.Context, apiKey string) (context.Context, bool) {
	if !m.config.HasAPIKeys() {
		return ctx, false
	}

	keyIndex := m.findValidKeyIndex(apiKey)
	if keyIndex == -1 {
		return ctx, false
	}

	// Add API key info to context
	ctx = context.WithValue(ctx, APIKeyContextKey, keyIndex)
	ctx = context.WithValue(ctx, AuthTypeKey, AuthTypeAPIKey)

	return ctx, true
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

// GetUserID retrieves the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

// GetUserEmail retrieves the user email from the request context
func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUserRole retrieves the user role from the request context
func GetUserRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}

// GetAuthType retrieves the authentication type from the request context
func GetAuthType(ctx context.Context) (AuthType, bool) {
	authType, ok := ctx.Value(AuthTypeKey).(AuthType)
	return authType, ok
}

// IsJWTAuth checks if the request was authenticated via JWT
func IsJWTAuth(ctx context.Context) bool {
	authType, ok := GetAuthType(ctx)
	return ok && authType == AuthTypeJWT
}

// IsAPIKeyAuth checks if the request was authenticated via API key
func IsAPIKeyAuth(ctx context.Context) bool {
	authType, ok := GetAuthType(ctx)
	return ok && authType == AuthTypeAPIKey
}
