package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockVersionRepository is a mock implementation of domain.VersionRepository
type MockVersionRepository struct {
	mu       sync.RWMutex
	versions map[string]*domain.Version

	// Hooks for customizing behavior
	CreateFunc                      func(ctx context.Context, version *domain.Version) error
	GetByIDFunc                     func(ctx context.Context, id string) (*domain.Version, error)
	DeleteFunc                      func(ctx context.Context, id string) error
	ListByServiceIDFunc             func(ctx context.Context, serviceID string) ([]domain.Version, error)
	CountByServiceIDFunc            func(ctx context.Context, serviceID string) (int, error)
	DeleteByServiceIDFunc           func(ctx context.Context, serviceID string) error
	ExistsByServiceIDAndVersionFunc func(ctx context.Context, serviceID, version string) (bool, error)
}

// NewMockVersionRepository creates a new MockVersionRepository
func NewMockVersionRepository() *MockVersionRepository {
	return &MockVersionRepository{
		versions: make(map[string]*domain.Version),
	}
}

// Create creates a new version
func (m *MockVersionRepository) Create(ctx context.Context, version *domain.Version) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, version)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if version.ID.IsZero() {
		version.ID = primitive.NewObjectID()
	}
	version.CreatedAt = time.Now()

	m.versions[version.ID.Hex()] = version
	return nil
}

// GetByID retrieves a version by its ID
func (m *MockVersionRepository) GetByID(ctx context.Context, id string) (*domain.Version, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	version, ok := m.versions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return version, nil
}

// Delete deletes a version by its ID
func (m *MockVersionRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.versions[id]; !ok {
		return domain.ErrNotFound
	}

	delete(m.versions, id)
	return nil
}

// ListByServiceID retrieves all versions for a specific service
func (m *MockVersionRepository) ListByServiceID(ctx context.Context, serviceID string) ([]domain.Version, error) {
	if m.ListByServiceIDFunc != nil {
		return m.ListByServiceIDFunc(ctx, serviceID)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var versions []domain.Version
	for _, v := range m.versions {
		if v.ServiceID.Hex() == serviceID {
			versions = append(versions, *v)
		}
	}
	return versions, nil
}

// CountByServiceID counts versions for a specific service
func (m *MockVersionRepository) CountByServiceID(ctx context.Context, serviceID string) (int, error) {
	if m.CountByServiceIDFunc != nil {
		return m.CountByServiceIDFunc(ctx, serviceID)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, v := range m.versions {
		if v.ServiceID.Hex() == serviceID {
			count++
		}
	}
	return count, nil
}

// DeleteByServiceID deletes all versions for a specific service
func (m *MockVersionRepository) DeleteByServiceID(ctx context.Context, serviceID string) error {
	if m.DeleteByServiceIDFunc != nil {
		return m.DeleteByServiceIDFunc(ctx, serviceID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, v := range m.versions {
		if v.ServiceID.Hex() == serviceID {
			delete(m.versions, id)
		}
	}
	return nil
}

// ExistsByServiceIDAndVersion checks if a version exists for a service
func (m *MockVersionRepository) ExistsByServiceIDAndVersion(ctx context.Context, serviceID, version string) (bool, error) {
	if m.ExistsByServiceIDAndVersionFunc != nil {
		return m.ExistsByServiceIDAndVersionFunc(ctx, serviceID, version)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, v := range m.versions {
		if v.ServiceID.Hex() == serviceID && v.Version == version {
			return true, nil
		}
	}
	return false, nil
}

// AddVersion adds a version directly to the mock (for test setup)
func (m *MockVersionRepository) AddVersion(version *domain.Version) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if version.ID.IsZero() {
		version.ID = primitive.NewObjectID()
	}
	m.versions[version.ID.Hex()] = version
}

// Reset clears all versions from the mock
func (m *MockVersionRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versions = make(map[string]*domain.Version)
}
