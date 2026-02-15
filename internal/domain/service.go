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
	ID          string    `json:"id" example:"507f1f77bcf86cd799439011"`
	Name        string    `json:"name" example:"payment-service"`
	Description string    `json:"description" example:"Handles payment processing"`
	Revision    int       `json:"revision" example:"1"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
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
	Name        string `json:"name" example:"payment-service"`
	Description string `json:"description" example:"Handles payment processing"`
}

// UpdateServiceRequest represents the request body for updating a service
type UpdateServiceRequest struct {
	Name        string `json:"name" example:"payment-service-v2"`
	Description string `json:"description" example:"Updated payment processing service"`
}

// PatchServiceRequest represents the request body for partially updating a service
type PatchServiceRequest struct {
	Name        *string `json:"name,omitempty" example:"new-service-name"`
	Description *string `json:"description,omitempty" example:"Updated description"`
}
