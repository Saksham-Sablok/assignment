package domain

import "context"

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	// Create creates a new service with revision 1
	Create(ctx context.Context, service *Service) error

	// GetByID retrieves a service by its ID
	GetByID(ctx context.Context, id string) (*Service, error)

	// Update updates an existing service and increments revision
	Update(ctx context.Context, service *Service) error

	// Delete deletes a service by its ID
	Delete(ctx context.Context, id string) error

	// List retrieves services with filtering, sorting, and pagination
	List(ctx context.Context, params ListParams) (*PaginatedResult[Service], error)
}
