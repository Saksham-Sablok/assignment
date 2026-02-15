package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockServiceVersionRepository is a mock implementation of domain.ServiceVersionRepository
type MockServiceVersionRepository struct {
	mu       sync.RWMutex
	versions map[string]*domain.ServiceVersion

	// Hooks for customizing behavior
	CreateFunc                    func(ctx context.Context, version *domain.ServiceVersion) error
	GetByIDFunc                   func(ctx context.Context, id string) (*domain.ServiceVersion, error)
	GetByServiceIDAndRevisionFunc func(ctx context.Context, serviceID string, revision int) (*domain.ServiceVersion, error)
	ListByServiceIDFunc           func(ctx context.Context, serviceID string, params domain.PaginationParams) (*domain.PaginatedResult[domain.ServiceVersion], error)
	DeleteByServiceIDFunc         func(ctx context.Context, serviceID string) error
}

// NewMockServiceVersionRepository creates a new MockServiceVersionRepository
func NewMockServiceVersionRepository() *MockServiceVersionRepository {
	return &MockServiceVersionRepository{
		versions: make(map[string]*domain.ServiceVersion),
	}
}

// Create creates a new service version
func (m *MockServiceVersionRepository) Create(ctx context.Context, version *domain.ServiceVersion) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, version)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if version.ID.IsZero() {
		version.ID = primitive.NewObjectID()
	}
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now()
	}

	m.versions[version.ID.Hex()] = version
	return nil
}

// GetByID retrieves a service version by its ID
func (m *MockServiceVersionRepository) GetByID(ctx context.Context, id string) (*domain.ServiceVersion, error) {
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

// GetByServiceIDAndRevision retrieves a specific revision of a service
func (m *MockServiceVersionRepository) GetByServiceIDAndRevision(ctx context.Context, serviceID string, revision int) (*domain.ServiceVersion, error) {
	if m.GetByServiceIDAndRevisionFunc != nil {
		return m.GetByServiceIDAndRevisionFunc(ctx, serviceID, revision)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, v := range m.versions {
		if v.ServiceID.Hex() == serviceID && v.Revision == revision {
			return v, nil
		}
	}
	return nil, domain.ErrNotFound
}

// ListByServiceID retrieves all versions for a service with pagination
func (m *MockServiceVersionRepository) ListByServiceID(ctx context.Context, serviceID string, params domain.PaginationParams) (*domain.PaginatedResult[domain.ServiceVersion], error) {
	if m.ListByServiceIDFunc != nil {
		return m.ListByServiceIDFunc(ctx, serviceID, params)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var versions []domain.ServiceVersion
	for _, v := range m.versions {
		if v.ServiceID.Hex() == serviceID {
			versions = append(versions, *v)
		}
	}

	total := int64(len(versions))

	// Apply pagination
	start := params.Offset()
	end := start + params.Limit
	if start >= len(versions) {
		versions = []domain.ServiceVersion{}
	} else {
		if end > len(versions) {
			end = len(versions)
		}
		versions = versions[start:end]
	}

	return domain.NewPaginatedResult(versions, total, params), nil
}

// DeleteByServiceID deletes all versions for a service
func (m *MockServiceVersionRepository) DeleteByServiceID(ctx context.Context, serviceID string) error {
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

// AddVersion adds a version directly to the mock (for test setup)
func (m *MockServiceVersionRepository) AddVersion(version *domain.ServiceVersion) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if version.ID.IsZero() {
		version.ID = primitive.NewObjectID()
	}
	m.versions[version.ID.Hex()] = version
}

// Reset clears all versions from the mock
func (m *MockServiceVersionRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versions = make(map[string]*domain.ServiceVersion)
}
