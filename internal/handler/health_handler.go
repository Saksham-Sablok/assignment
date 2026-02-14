package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/services-api/pkg/response"
	"go.mongodb.org/mongo-driver/mongo"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	mongoClient *mongo.Client
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(client *mongo.Client) *HealthHandler {
	return &HealthHandler{
		mongoClient: client,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Check handles GET /health
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	healthResp := HealthResponse{
		Status:   "healthy",
		Database: "connected",
	}

	// Check MongoDB connection
	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		healthResp.Status = "unhealthy"
		healthResp.Database = "disconnected"
		response.JSON(w, http.StatusServiceUnavailable, healthResp)
		return
	}

	response.OK(w, healthResp)
}
