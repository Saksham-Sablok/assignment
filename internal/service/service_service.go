package service

import (
	"context"
	"errors"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServiceService handles business logic for services
type ServiceService struct {
	serviceRepo domain.ServiceRepository
}

// NewServiceService creates a new ServiceService
func NewServiceService(serviceRepo domain.ServiceRepository) *ServiceService {
	return &ServiceService{
		serviceRepo: serviceRepo,
	}
}

// Create creates a new service with validation
func (s *ServiceService) Create(ctx context.Context, req domain.CreateServiceRequest) (*domain.Service, error) {
	// Validate request
	if err := validateCreateServiceRequest(req); err != nil {
		return nil, err
	}

	service := &domain.Service{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.serviceRepo.Create(ctx, service); err != nil {
		return nil, err
	}

	return service, nil
}

// GetByID retrieves a service by its ID
func (s *ServiceService) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	return s.serviceRepo.GetByID(ctx, id)
}

// Update performs a full update of a service (increments revision)
func (s *ServiceService) Update(ctx context.Context, id string, req domain.UpdateServiceRequest) (*domain.Service, error) {
	// Validate request
	if err := validateUpdateServiceRequest(req); err != nil {
		return nil, err
	}

	// Get existing service
	service, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	service.Name = req.Name
	service.Description = req.Description

	if err := s.serviceRepo.Update(ctx, service); err != nil {
		return nil, err
	}

	// Re-fetch to get updated timestamps and revision
	return s.serviceRepo.GetByID(ctx, id)
}

// Patch performs a partial update of a service (increments revision)
func (s *ServiceService) Patch(ctx context.Context, id string, req domain.PatchServiceRequest) (*domain.Service, error) {
	// Get existing service
	service, err := s.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if req.Name != nil {
		if len(*req.Name) == 0 {
			return nil, domain.ErrNameRequired
		}
		if len(*req.Name) > 255 {
			return nil, domain.ErrNameTooLong
		}
		service.Name = *req.Name
	}

	if req.Description != nil {
		if len(*req.Description) == 0 {
			return nil, domain.ErrDescriptionRequired
		}
		if len(*req.Description) > 1000 {
			return nil, domain.ErrDescriptionTooLong
		}
		service.Description = *req.Description
	}

	if err := s.serviceRepo.Update(ctx, service); err != nil {
		return nil, err
	}

	// Re-fetch to get updated timestamps and revision
	return s.serviceRepo.GetByID(ctx, id)
}

// Delete deletes a service
func (s *ServiceService) Delete(ctx context.Context, id string) error {
	// Validate ID format
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return domain.ErrInvalidID
	}

	return s.serviceRepo.Delete(ctx, id)
}

// List retrieves services with filtering, sorting, and pagination
func (s *ServiceService) List(ctx context.Context, params domain.ListParams) (*domain.PaginatedResult[domain.Service], error) {
	// Validate sort field
	if params.Sort != "" && !domain.IsValidSortField(params.Sort) {
		return nil, domain.ErrInvalidSortField
	}

	// Apply defaults
	if params.Sort == "" {
		params.Sort = "created_at"
	}
	if params.Order == "" {
		params.Order = "desc"
	}
	if params.Pagination.Limit == 0 {
		params.Pagination.Limit = 20
	}
	if params.Pagination.Page == 0 {
		params.Pagination.Page = 1
	}

	// Cap limit at 100
	if params.Pagination.Limit > 100 {
		params.Pagination.Limit = 100
	}

	return s.serviceRepo.List(ctx, params)
}

// validateCreateServiceRequest validates a create service request
func validateCreateServiceRequest(req domain.CreateServiceRequest) error {
	if req.Name == "" {
		return domain.ErrNameRequired
	}
	if len(req.Name) > 255 {
		return domain.ErrNameTooLong
	}
	if req.Description == "" {
		return domain.ErrDescriptionRequired
	}
	if len(req.Description) > 1000 {
		return domain.ErrDescriptionTooLong
	}
	return nil
}

// validateUpdateServiceRequest validates an update service request
func validateUpdateServiceRequest(req domain.UpdateServiceRequest) error {
	if req.Name == "" {
		return domain.ErrNameRequired
	}
	if len(req.Name) > 255 {
		return domain.ErrNameTooLong
	}
	if req.Description == "" {
		return domain.ErrDescriptionRequired
	}
	if len(req.Description) > 1000 {
		return domain.ErrDescriptionTooLong
	}
	return nil
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, domain.ErrNameRequired) ||
		errors.Is(err, domain.ErrDescriptionRequired) ||
		errors.Is(err, domain.ErrNameTooLong) ||
		errors.Is(err, domain.ErrDescriptionTooLong) ||
		errors.Is(err, domain.ErrInvalidSortField) ||
		errors.Is(err, domain.ErrInvalidID)
}
