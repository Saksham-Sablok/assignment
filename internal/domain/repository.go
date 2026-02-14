package domain

import "context"

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	// Create creates a new service
	Create(ctx context.Context, service *Service) error

	// GetByID retrieves a service by its ID, including version count
	GetByID(ctx context.Context, id string) (*Service, error)

	// Update updates an existing service
	Update(ctx context.Context, service *Service) error

	// Delete deletes a service by its ID
	Delete(ctx context.Context, id string) error

	// List retrieves services with filtering, sorting, and pagination
	List(ctx context.Context, params ListParams) (*PaginatedResult[Service], error)
}

// VersionRepository defines the interface for version data access
type VersionRepository interface {
	// Create creates a new version
	Create(ctx context.Context, version *Version) error

	// GetByID retrieves a version by its ID
	GetByID(ctx context.Context, id string) (*Version, error)

	// Delete deletes a version by its ID
	Delete(ctx context.Context, id string) error

	// ListByServiceID retrieves all versions for a specific service
	ListByServiceID(ctx context.Context, serviceID string) ([]Version, error)

	// CountByServiceID counts versions for a specific service
	CountByServiceID(ctx context.Context, serviceID string) (int, error)

	// DeleteByServiceID deletes all versions for a specific service
	DeleteByServiceID(ctx context.Context, serviceID string) error

	// ExistsByServiceIDAndVersion checks if a version exists for a service
	ExistsByServiceIDAndVersion(ctx context.Context, serviceID, version string) (bool, error)
}
