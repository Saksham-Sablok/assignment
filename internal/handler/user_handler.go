package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/services-api/internal/domain"
	"github.com/services-api/internal/service"
	"github.com/services-api/pkg/auth"
	"github.com/services-api/pkg/response"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// UserListResponse represents the response for listing users
type UserListResponse struct {
	Data       []domain.UserResponse     `json:"data"`
	Pagination domain.PaginationMetadata `json:"pagination"`
}

// Create handles POST /api/v1/users
// @Summary Create a new user (Admin only)
// @Description Create a new user account. Requires admin role.
// @Tags users
// @Accept json
// @Produce json
// @Param request body domain.CreateUserRequest true "User creation request"
// @Success 201 {object} domain.UserResponse "Created user"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - Admin only"
// @Failure 409 {object} response.ErrorResponse "Email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	if !h.isAdmin(r) {
		response.Forbidden(w, "admin access required")
		return
	}

	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.userService.Create(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Created(w, user.ToResponse())
}

// List handles GET /api/v1/users
// @Summary List all users (Admin only)
// @Description Get a paginated list of all users. Requires admin role.
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 100)" default(20)
// @Success 200 {object} UserListResponse "List of users with pagination"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - Admin only"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	if !h.isAdmin(r) {
		response.Forbidden(w, "admin access required")
		return
	}

	params := ParsePaginationParams(r)

	result, err := h.userService.List(r.Context(), params)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response format
	userResponses := make([]domain.UserResponse, len(result.Data))
	for i, u := range result.Data {
		userResponses[i] = u.ToResponse()
	}

	response.OK(w, map[string]interface{}{
		"data":       userResponses,
		"pagination": result.Pagination,
	})
}

// Get handles GET /api/v1/users/{id}
// @Summary Get a user by ID
// @Description Get detailed information about a user. Users can view their own profile, admins can view any user.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (MongoDB ObjectID)"
// @Success 200 {object} domain.UserResponse "User details"
// @Failure 400 {object} response.ErrorResponse "Invalid ID format"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "user id is required")
		return
	}

	// Check if user can access this resource
	if !h.canAccessUser(r, id) {
		response.Forbidden(w, "access denied")
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, user.ToResponse())
}

// Update handles PUT /api/v1/users/{id}
// @Summary Update a user (Admin only)
// @Description Update a user's information. Requires admin role.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (MongoDB ObjectID)"
// @Param request body domain.UpdateUserRequest true "User update request"
// @Success 200 {object} domain.UserResponse "Updated user"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - Admin only"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 409 {object} response.ErrorResponse "Email already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	if !h.isAdmin(r) {
		response.Forbidden(w, "admin access required")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "user id is required")
		return
	}

	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	user, err := h.userService.Update(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, user.ToResponse())
}

// Delete handles DELETE /api/v1/users/{id}
// @Summary Delete a user (Admin only)
// @Description Delete a user by ID. Requires admin role.
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (MongoDB ObjectID)"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid ID format"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden - Admin only"
// @Failure 404 {object} response.ErrorResponse "User not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Check admin role
	if !h.isAdmin(r) {
		response.Forbidden(w, "admin access required")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "user id is required")
		return
	}

	if err := h.userService.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// GetMe handles GET /api/v1/users/me
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} domain.UserResponse "Current user details"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "authentication required")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, user.ToResponse())
}

// ChangePassword handles POST /api/v1/users/me/password
// @Summary Change current user's password
// @Description Change the password of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Param request body domain.ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized or wrong current password"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /users/me/password [post]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "authentication required")
		return
	}

	var req domain.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.userService.ChangePassword(r.Context(), userID, req); err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "password changed successfully"})
}

// isAdmin checks if the current user has admin role
func (h *UserHandler) isAdmin(r *http.Request) bool {
	role, ok := auth.GetUserRole(r.Context())
	if !ok {
		return false
	}
	return role == domain.RoleAdmin
}

// canAccessUser checks if the current user can access the requested user resource
func (h *UserHandler) canAccessUser(r *http.Request, targetUserID string) bool {
	// Admins can access any user
	if h.isAdmin(r) {
		return true
	}

	// Users can only access their own profile
	currentUserID, ok := auth.GetUserID(r.Context())
	if !ok {
		return false
	}
	return currentUserID == targetUserID
}

// handleError handles errors from the user service
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		response.NotFound(w, "user not found")
	case errors.Is(err, domain.ErrInvalidID):
		response.BadRequest(w, "invalid user id format")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		response.Conflict(w, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Unauthorized(w, "invalid credentials")
	case errors.Is(err, domain.ErrEmailRequired),
		errors.Is(err, domain.ErrEmailInvalid),
		errors.Is(err, domain.ErrEmailTooLong),
		errors.Is(err, domain.ErrPasswordRequired),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordTooLong),
		errors.Is(err, domain.ErrFirstNameRequired),
		errors.Is(err, domain.ErrFirstNameTooLong),
		errors.Is(err, domain.ErrLastNameTooLong),
		errors.Is(err, domain.ErrInvalidRole):
		response.BadRequest(w, err.Error())
	default:
		response.InternalServerError(w, "internal server error")
	}
}
