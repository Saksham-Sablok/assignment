package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ServiceVersion represents a historical snapshot of a service at a specific revision
type ServiceVersion struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceID   primitive.ObjectID `bson:"service_id" json:"service_id"`
	Revision    int                `bson:"revision" json:"revision"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"` // When this version was created
}

// ServiceVersionResponse is the API response format for a service version
type ServiceVersionResponse struct {
	ID          string    `json:"id" example:"507f1f77bcf86cd799439011"`
	ServiceID   string    `json:"service_id" example:"507f1f77bcf86cd799439012"`
	Revision    int       `json:"revision" example:"2"`
	Name        string    `json:"name" example:"payment-service"`
	Description string    `json:"description" example:"Handles payment processing"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// ToResponse converts a ServiceVersion to its API response format
func (sv *ServiceVersion) ToResponse() ServiceVersionResponse {
	return ServiceVersionResponse{
		ID:          sv.ID.Hex(),
		ServiceID:   sv.ServiceID.Hex(),
		Revision:    sv.Revision,
		Name:        sv.Name,
		Description: sv.Description,
		CreatedAt:   sv.CreatedAt,
	}
}

// NewServiceVersion creates a new ServiceVersion from a Service
func NewServiceVersion(service *Service) *ServiceVersion {
	return &ServiceVersion{
		ID:          primitive.NewObjectID(),
		ServiceID:   service.ID,
		Revision:    service.Revision,
		Name:        service.Name,
		Description: service.Description,
		CreatedAt:   time.Now(),
	}
}
