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

// ServiceVersionRepository defines the interface for service version data access
type ServiceVersionRepository interface {
	// Create creates a new service version snapshot
	Create(ctx context.Context, version *ServiceVersion) error

	// GetByID retrieves a service version by its ID
	GetByID(ctx context.Context, id string) (*ServiceVersion, error)

	// GetByServiceIDAndRevision retrieves a specific revision of a service
	GetByServiceIDAndRevision(ctx context.Context, serviceID string, revision int) (*ServiceVersion, error)

	// ListByServiceID retrieves all versions for a service with pagination
	ListByServiceID(ctx context.Context, serviceID string, params PaginationParams) (*PaginatedResult[ServiceVersion], error)

	// DeleteByServiceID deletes all versions for a service
	DeleteByServiceID(ctx context.Context, serviceID string) error
}
