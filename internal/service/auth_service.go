package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/services-api/internal/domain"
	"github.com/services-api/pkg/jwt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// AuthService handles authentication operations
type AuthService struct {
	userRepo   domain.UserRepository
	jwtManager *jwt.Manager
}

// NewAuthService creates a new AuthService
func NewAuthService(userRepo domain.UserRepository, jwtManager *jwt.Manager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Validate request
	if err := s.validateRegisterRequest(req); err != nil {
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
		Role:      domain.RoleUser, // Default role
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

	// Generate tokens
	return s.generateAuthResponse(user)
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	// Validate request
	if req.Email == "" {
		return nil, domain.ErrEmailRequired
	}
	if req.Password == "" {
		return nil, domain.ErrPasswordRequired
	}

	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.Active {
		return nil, domain.ErrInvalidCredentials
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Get user to ensure they still exist and are active
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.Active {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate new tokens
	return s.generateAuthResponse(user)
}

// generateAuthResponse creates tokens and builds the auth response
func (s *AuthService) generateAuthResponse(user *domain.User) (*domain.AuthResponse, error) {
	userID := user.ID.Hex()

	accessToken, err := s.jwtManager.GenerateAccessToken(userID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(userID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtManager.GetAccessTokenExpiry(),
		User:         user.ToResponse(),
	}, nil
}

// validateRegisterRequest validates the registration request
func (s *AuthService) validateRegisterRequest(req domain.RegisterRequest) error {
	// Validate email
	if req.Email == "" {
		return domain.ErrEmailRequired
	}
	if len(req.Email) > 255 {
		return domain.ErrEmailTooLong
	}
	if !emailRegex.MatchString(req.Email) {
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

	return nil
}
