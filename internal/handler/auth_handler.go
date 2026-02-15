package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/services-api/internal/domain"
	"github.com/services-api/internal/service"
	"github.com/services-api/pkg/response"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles POST /api/v1/auth/register
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Registration request"
// @Success 201 {object} domain.AuthResponse "Successfully registered"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 409 {object} response.ErrorResponse "Email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	authResp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Created(w, authResp)
}

// Login handles POST /api/v1/auth/login
// @Summary Login to the application
// @Description Authenticate with email and password to get access tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login request"
// @Success 200 {object} domain.AuthResponse "Successfully authenticated"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Invalid credentials"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	authResp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, authResp)
}

// Refresh handles POST /api/v1/auth/refresh
// @Summary Refresh access token
// @Description Get new access and refresh tokens using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} domain.AuthResponse "Successfully refreshed tokens"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "refresh_token is required")
		return
	}

	authResp, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, authResp)
}

// handleError handles errors from the auth service
func (h *AuthHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailRequired),
		errors.Is(err, domain.ErrEmailInvalid),
		errors.Is(err, domain.ErrEmailTooLong),
		errors.Is(err, domain.ErrPasswordRequired),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordTooLong),
		errors.Is(err, domain.ErrFirstNameRequired),
		errors.Is(err, domain.ErrFirstNameTooLong),
		errors.Is(err, domain.ErrLastNameTooLong):
		response.BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		response.Conflict(w, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Unauthorized(w, "invalid credentials")
	default:
		response.InternalServerError(w, "internal server error")
	}
}
