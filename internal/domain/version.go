package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Version represents a version of a service
type Version struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceID primitive.ObjectID `bson:"service_id" json:"service_id"`
	Version   string             `bson:"version" json:"version"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// VersionResponse is the API response format for a version
type VersionResponse struct {
	ID        string    `json:"id"`
	ServiceID string    `json:"service_id"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a Version to its API response format
func (v *Version) ToResponse() VersionResponse {
	return VersionResponse{
		ID:        v.ID.Hex(),
		ServiceID: v.ServiceID.Hex(),
		Version:   v.Version,
		CreatedAt: v.CreatedAt,
	}
}

// CreateVersionRequest represents the request body for creating a version
type CreateVersionRequest struct {
	Version string `json:"version"`
}
