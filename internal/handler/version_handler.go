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

// VersionHandler handles HTTP requests for versions
type VersionHandler struct {
	service *service.VersionService
}

// NewVersionHandler creates a new VersionHandler
func NewVersionHandler(svc *service.VersionService) *VersionHandler {
	return &VersionHandler{
		service: svc,
	}
}

// Create handles POST /api/v1/services/{service_id}/versions
func (h *VersionHandler) Create(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "service_id")
	if serviceID == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	var req domain.CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	version, err := h.service.Create(r.Context(), serviceID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Created(w, version.ToResponse())
}

// List handles GET /api/v1/services/{service_id}/versions
func (h *VersionHandler) List(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "service_id")
	if serviceID == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	versions, err := h.service.ListByServiceID(r.Context(), serviceID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Convert to response format
	versionResponses := make([]domain.VersionResponse, len(versions))
	for i, v := range versions {
		versionResponses[i] = v.ToResponse()
	}

	response.OK(w, map[string]interface{}{
		"data": versionResponses,
	})
}

// Get handles GET /api/v1/services/{service_id}/versions/{version_id}
func (h *VersionHandler) Get(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "service_id")
	if serviceID == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	versionID := chi.URLParam(r, "version_id")
	if versionID == "" {
		response.BadRequest(w, "version id is required")
		return
	}

	version, err := h.service.GetByID(r.Context(), serviceID, versionID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.OK(w, version.ToResponse())
}

// Delete handles DELETE /api/v1/services/{service_id}/versions/{version_id}
func (h *VersionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "service_id")
	if serviceID == "" {
		response.BadRequest(w, "service id is required")
		return
	}

	versionID := chi.URLParam(r, "version_id")
	if versionID == "" {
		response.BadRequest(w, "version id is required")
		return
	}

	if err := h.service.Delete(r.Context(), serviceID, versionID); err != nil {
		h.handleError(w, err)
		return
	}

	response.NoContent(w)
}

// handleError handles errors from the service layer
func (h *VersionHandler) handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		response.NotFound(w, "resource not found")
		return
	}

	if errors.Is(err, domain.ErrDuplicateVersion) {
		response.Conflict(w, "version already exists for this service")
		return
	}

	if service.IsValidationError(err) {
		response.BadRequest(w, err.Error())
		return
	}

	if errors.Is(err, domain.ErrInvalidID) {
		response.BadRequest(w, "invalid id format")
		return
	}

	response.InternalServerError(w, "internal server error")
}
