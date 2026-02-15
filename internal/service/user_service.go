package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/services-api/internal/domain"
)

var userEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// UserService handles user management operations
type UserService struct {
	userRepo domain.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Create creates a new user (admin operation)
func (s *UserService) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	// Create user
	user := &domain.User{
		Email:     strings.ToLower(req.Email),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Role:      req.Role,
		Active:    true,
	}

	// Set password
	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// Update updates a user (admin operation)
func (s *UserService) Update(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if it already exists
	newEmail := strings.ToLower(req.Email)
	if newEmail != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, newEmail)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrEmailAlreadyExists
		}
	}

	// Update fields
	user.Email = newEmail
	user.FirstName = strings.TrimSpace(req.FirstName)
	user.LastName = strings.TrimSpace(req.LastName)
	user.Role = req.Role
	if req.Active != nil {
		user.Active = *req.Active
	}

	// Save user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete deletes a user
func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}

// List retrieves users with pagination
func (s *UserService) List(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult[domain.User], error) {
	return s.userRepo.List(ctx, params)
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID string, req domain.ChangePasswordRequest) error {
	// Validate request
	if req.CurrentPassword == "" {
		return domain.ErrPasswordRequired
	}
	if req.NewPassword == "" {
		return domain.ErrPasswordRequired
	}
	if len(req.NewPassword) < 8 {
		return domain.ErrPasswordTooShort
	}
	if len(req.NewPassword) > 72 {
		return domain.ErrPasswordTooLong
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return domain.ErrInvalidCredentials
	}

	// Set new password
	if err := user.SetPassword(req.NewPassword); err != nil {
		return err
	}

	// Save user
	return s.userRepo.Update(ctx, user)
}

// validateCreateRequest validates the create user request
func (s *UserService) validateCreateRequest(req domain.CreateUserRequest) error {
	// Validate email
	if req.Email == "" {
		return domain.ErrEmailRequired
	}
	if len(req.Email) > 255 {
		return domain.ErrEmailTooLong
	}
	if !userEmailRegex.MatchString(req.Email) {
		return domain.ErrEmailInvalid
	}

	// Validate password
	if req.Password == "" {
		return domain.ErrPasswordRequired
	}
	if len(req.Password) < 8 {
		return domain.ErrPasswordTooShort
	}
	if len(req.Password) > 72 {
		return domain.ErrPasswordTooLong
	}

	// Validate first name
	if strings.TrimSpace(req.FirstName) == "" {
		return domain.ErrFirstNameRequired
	}
	if len(req.FirstName) > 100 {
		return domain.ErrFirstNameTooLong
	}

	// Validate last name (optional but has max length)
	if len(req.LastName) > 100 {
		return domain.ErrLastNameTooLong
	}

	// Validate role
	if req.Role == "" {
		req.Role = domain.RoleUser
	}
	if !domain.IsValidRole(req.Role) {
		return domain.ErrInvalidRole
	}

	return nil
}

// validateUpdateRequest validates the update user request
func (s *UserService) validateUpdateRequest(req domain.UpdateUserRequest) error {
	// Validate email
	if req.Email == "" {
		return domain.ErrEmailRequired
	}
	if len(req.Email) > 255 {
		return domain.ErrEmailTooLong
	}
	if !userEmailRegex.MatchString(req.Email) {
		return domain.ErrEmailInvalid
	}

	// Validate first name
	if strings.TrimSpace(req.FirstName) == "" {
		return domain.ErrFirstNameRequired
	}
	if len(req.FirstName) > 100 {
		return domain.ErrFirstNameTooLong
	}

	// Validate last name (optional but has max length)
	if len(req.LastName) > 100 {
		return domain.ErrLastNameTooLong
	}

	// Validate role
	if !domain.IsValidRole(req.Role) {
		return domain.ErrInvalidRole
	}

	return nil
}
