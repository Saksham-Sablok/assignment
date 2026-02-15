package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/services-api/internal/domain"
	"github.com/services-api/internal/service"
	"github.com/services-api/pkg/response"
)

// ServiceHandler handles HTTP requests for services
type ServiceHandler struct {
	service *service.ServiceService
}

// NewServiceHandler creates a new ServiceHandler
func NewServiceHandler(svc *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{
		service: svc,
	}
}

// Create handles POST /api/v1/services
// @Summary Create a new service
// @Description Create a new service with name and description. Revision starts at 1.
// @Tags services
// @Accept json
// @Produce json
// @Param request body domain.CreateServiceRequest true "Service creation request"
// @Success 201 {object} domain.ServiceResponse "Created service"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /services [post]
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	svc, err := h.service.Create(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Created(w, svc.ToResponse())
}

// ServiceListResponse represents the response for listing services
type ServiceListResponse struct {
	Data       []domain.ServiceResponse  `json:"data"`
	Pagination domain.PaginationMetadata `json:"pagination"`
}

// List handles GET /api/v1/services
// @Summary List all services
// @Description Get a paginated list of services with optional filtering and sorting
// @Tags services
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page (max 100)" default(20)
// @Param search query string false "Search in name and description"
// @Param name query string false "Filter by exact name"
// @Param sort query string false "Sort field (name, created_at, updated_at)" default(created_at)
// @Param order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} ServiceListResponse "List of services with pagination"
// @Failure 400 {object} response.ErrorResponse "Invalid parameters"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /services [get]
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	params := ParseListParams(r)

	result, err := h.service.List(r.Context(), params)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response format
	serviceResponses := make([]domain.ServiceResponse, len(result.Data))
	for i, svc := range result.Data {
		serviceResponses[i] = svc.ToResponse()
	}

	response.OK(w, map[string]interface{}{
		"data":       serviceResponses,
		"pagination": result.Pagination,
	})
}

// Get handles GET /api/v1/services/{id}
// @Summary Get a service by ID
// @Description Get detailed information about a specific service
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID (MongoDB ObjectID)"
// @Success 200 {object} domain.ServiceResponse "Service details"
// @Failure 400 {object} response.ErrorResponse "Invalid ID format"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /services/{id} [get]
func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	svc, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, svc.ToResponse())
}

// Update handles PUT /api/v1/services/{id}
// @Summary Update a service
// @Description Full update of a service. All fields are required. Revision is automatically incremented.
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID (MongoDB ObjectID)"
// @Param request body domain.UpdateServiceRequest true "Service update request"
// @Success 200 {object} domain.ServiceResponse "Updated service"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /services/{id} [put]
func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	var req domain.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	svc, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, svc.ToResponse())
}

// Patch handles PATCH /api/v1/services/{id}
// @Summary Partially update a service
// @Description Partial update of a service. Only provided fields are updated. Revision is automatically incremented.
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID (MongoDB ObjectID)"
// @Param request body domain.PatchServiceRequest true "Service patch request"
// @Success 200 {object} domain.ServiceResponse "Updated service"
// @Failure 400 {object} response.ErrorResponse "Validation error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /services/{id} [patch]
func (h *ServiceHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	var req domain.PatchServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	svc, err := h.service.Patch(r.Context(), id, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, svc.ToResponse())
}

// Delete handles DELETE /api/v1/services/{id}
// @Summary Delete a service
// @Description Delete a service by ID
// @Tags services
// @Accept json
// @Produce json
// @Param id path string true "Service ID (MongoDB ObjectID)"
// @Success 204 "Service deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid ID format"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 404 {object} response.ErrorResponse "Service not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /services/{id} [delete]
func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// handleError handles errors from the service layer
func (h *ServiceHandler) handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		response.NotFound(w, "service not found")
		return
	}

	if service.IsValidationError(err) {
		response.BadRequest(w, err.Error())
		return
	}

	if errors.Is(err, domain.ErrInvalidID) {
		response.BadRequest(w, "invalid service id format")
		return
	}

	response.InternalServerError(w, "internal server error")
}
