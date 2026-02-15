package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service represents a service in the organization
type Service struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Revision    int                `bson:"revision" json:"revision"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ServiceResponse is the API response format for a service
type ServiceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Revision    int       `json:"revision"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts a Service to its API response format
func (s *Service) ToResponse() ServiceResponse {
	return ServiceResponse{
		ID:          s.ID.Hex(),
		Name:        s.Name,
		Description: s.Description,
		Revision:    s.Revision,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

// CreateServiceRequest represents the request body for creating a service
type CreateServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateServiceRequest represents the request body for updating a service
type UpdateServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PatchServiceRequest represents the request body for partially updating a service
type PatchServiceRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
