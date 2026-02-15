package service

import (
	"context"
	"errors"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VersionService handles business logic for versions
type VersionService struct {
	versionRepo domain.VersionRepository
	serviceRepo domain.ServiceRepository
}

// NewVersionService creates a new VersionService
func NewVersionService(versionRepo domain.VersionRepository, serviceRepo domain.ServiceRepository) *VersionService {
	return &VersionService{
		versionRepo: versionRepo,
		serviceRepo: serviceRepo,
	}
}

// Create creates a new version for a service with validation
func (s *VersionService) Create(ctx context.Context, serviceID string, req domain.CreateVersionRequest) (*domain.Version, error) {
	// Validate request
	if req.Version == "" {
		return nil, domain.ErrVersionRequired
	}

	// Validate service ID
	serviceObjID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	// Check if service exists
	_, err = s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Check for duplicate version
	exists, err := s.versionRepo.ExistsByServiceIDAndVersion(ctx, serviceID, req.Version)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrDuplicateVersion
	}

	version := &domain.Version{
		ServiceID: serviceObjID,
		Version:   req.Version,
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		return nil, err
	}

	return version, nil
}

// GetByID retrieves a version by its ID
func (s *VersionService) GetByID(ctx context.Context, serviceID, versionID string) (*domain.Version, error) {
	// Validate service exists
	_, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// Get the version
	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, err
	}

	// Verify the version belongs to the service
	if version.ServiceID.Hex() != serviceID {
		return nil, domain.ErrNotFound
	}

	return version, nil
}

// Delete deletes a version
func (s *VersionService) Delete(ctx context.Context, serviceID, versionID string) error {
	// Validate service exists
	_, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return err
	}

	// Get version to verify it belongs to the service
	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}

	// Verify the version belongs to the service
	if version.ServiceID.Hex() != serviceID {
		return domain.ErrNotFound
	}

	return s.versionRepo.Delete(ctx, versionID)
}

// ListByServiceID retrieves all versions for a service
func (s *VersionService) ListByServiceID(ctx context.Context, serviceID string) ([]domain.Version, error) {
	// Validate service exists
	_, err := s.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	return s.versionRepo.ListByServiceID(ctx, serviceID)
}

// IsDuplicateVersionError checks if the error is a duplicate version error
func IsDuplicateVersionError(err error) bool {
	return errors.Is(err, domain.ErrDuplicateVersion)
}
