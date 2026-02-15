package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockServiceRepository is a mock implementation of domain.ServiceRepository
type MockServiceRepository struct {
	mu       sync.RWMutex
	services map[string]*domain.Service

	// Hooks for customizing behavior
	CreateFunc  func(ctx context.Context, service *domain.Service) error
	GetByIDFunc func(ctx context.Context, id string) (*domain.Service, error)
	UpdateFunc  func(ctx context.Context, service *domain.Service) error
	DeleteFunc  func(ctx context.Context, id string) error
	ListFunc    func(ctx context.Context, params domain.ListParams) (*domain.PaginatedResult[domain.Service], error)
}

// NewMockServiceRepository creates a new MockServiceRepository
func NewMockServiceRepository() *MockServiceRepository {
	return &MockServiceRepository{
		services: make(map[string]*domain.Service),
	}
}

// Create creates a new service
func (m *MockServiceRepository) Create(ctx context.Context, service *domain.Service) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, service)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if service.ID.IsZero() {
		service.ID = primitive.NewObjectID()
	}
	now := time.Now()
	service.CreatedAt = now
	service.UpdatedAt = now
	service.Revision = 1

	m.services[service.ID.Hex()] = service
	return nil
}

// GetByID retrieves a service by its ID
func (m *MockServiceRepository) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	service, ok := m.services[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return service, nil
}

// Update updates an existing service and increments revision
func (m *MockServiceRepository) Update(ctx context.Context, service *domain.Service) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, service)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	id := service.ID.Hex()
	if _, ok := m.services[id]; !ok {
		return domain.ErrNotFound
	}

	service.UpdatedAt = time.Now()
	service.Revision++
	m.services[id] = service
	return nil
}

// Delete deletes a service by its ID
func (m *MockServiceRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.services[id]; !ok {
		return domain.ErrNotFound
	}

	delete(m.services, id)
	return nil
}

// List retrieves services with filtering, sorting, and pagination
func (m *MockServiceRepository) List(ctx context.Context, params domain.ListParams) (*domain.PaginatedResult[domain.Service], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, params)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var services []domain.Service
	for _, s := range m.services {
		services = append(services, *s)
	}

	total := int64(len(services))

	// Apply pagination
	start := params.Pagination.Offset()
	end := start + params.Pagination.Limit
	if start >= len(services) {
		services = []domain.Service{}
	} else {
		if end > len(services) {
			end = len(services)
		}
		services = services[start:end]
	}

	return domain.NewPaginatedResult(services, total, params.Pagination), nil
}

// AddService adds a service directly to the mock (for test setup)
func (m *MockServiceRepository) AddService(service *domain.Service) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if service.ID.IsZero() {
		service.ID = primitive.NewObjectID()
	}
	m.services[service.ID.Hex()] = service
}

// Reset clears all services from the mock
func (m *MockServiceRepository) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = make(map[string]*domain.Service)
}
