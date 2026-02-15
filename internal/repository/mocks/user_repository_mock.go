package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository is a mock implementation of domain.UserRepository
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User

	// Hooks for customizing behavior
	CreateFunc        func(ctx context.Context, user *domain.User) error
	GetByIDFunc       func(ctx context.Context, id string) (*domain.User, error)
	GetByEmailFunc    func(ctx context.Context, email string) (*domain.User, error)
	UpdateFunc        func(ctx context.Context, user *domain.User) error
	DeleteFunc        func(ctx context.Context, id string) error
	ListFunc          func(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult[domain.User], error)
	ExistsByEmailFunc func(ctx context.Context, email string) (bool, error)
}

// NewMockUserRepository creates a new MockUserRepository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

// Create creates a new user
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if email already exists
	for _, u := range m.users {
		if u.Email == user.Email {
			return domain.ErrEmailAlreadyExists
		}
	}

	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	m.users[user.ID.Hex()] = user
	return nil
}

// GetByID retrieves a user by their ID
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	user, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// GetByEmail retrieves a user by their email
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

// Update updates an existing user
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	id := user.ID.Hex()
	if _, ok := m.users[id]; !ok {
		return domain.ErrUserNotFound
	}

	// Check if email is taken by another user
	for _, u := range m.users {
		if u.Email == user.Email && u.ID.Hex() != id {
			return domain.ErrEmailAlreadyExists
		}
	}

	user.UpdatedAt = time.Now()
	m.users[id] = user
	return nil
}

// Delete deletes a user by their ID
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[id]; !ok {
		return domain.ErrUserNotFound
	}

	delete(m.users, id)
	return nil
}

// List retrieves users with pagination
func (m *MockUserRepository) List(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedResult[domain.User], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, params)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []domain.User
	for _, u := range m.users {
		users = append(users, *u)
	}

	total := int64(len(users))

	// Apply pagination
	start := params.Offset()
	end := start + params.Limit
	if start >= len(users) {
		users = []domain.User{}
	} else {
		if end > len(users) {
			end = len(users)
		}
		users = users[start:end]
	}

	return domain.NewPaginatedResult(users, total, params), nil
}

// ExistsByEmail checks if a user with the given email exists
func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.ExistsByEmailFunc != nil {
		return m.ExistsByEmailFunc(ctx, email)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// AddUser adds a user directly to the mock (for test setup)
func (m *MockUserRepository) AddUser(user *domain.User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	m.users[user.ID.Hex()] = user
}

// Reset clears all users from the mock
func (m *MockUserRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[string]*domain.User)
}

// GetAllUsers returns all users (for test assertions)
func (m *MockUserRepository) GetAllUsers() []*domain.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]*domain.User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users
}
