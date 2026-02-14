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

// List handles GET /api/v1/services
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
