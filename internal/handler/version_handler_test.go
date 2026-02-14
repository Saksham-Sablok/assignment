package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/services-api/internal/domain"
	"github.com/services-api/internal/handler"
	"github.com/services-api/internal/repository/mocks"
	"github.com/services-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupVersionHandler() (*handler.VersionHandler, *mocks.MockServiceRepository, *mocks.MockVersionRepository) {
	serviceRepo := mocks.NewMockServiceRepository()
	versionRepo := mocks.NewMockVersionRepository()
	svc := service.NewVersionService(versionRepo, serviceRepo)
	h := handler.NewVersionHandler(svc)
	return h, serviceRepo, versionRepo
}

func TestVersionHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				sRepo.AddService(svc)
				return svc.ID.Hex()
			},
			requestBody: map[string]string{
				"version": "1.0.0",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing version field",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)
				return svc.ID.Hex()
			},
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "version is required",
		},
		{
			name: "service not found",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				return primitive.NewObjectID().Hex()
			},
			requestBody: map[string]string{
				"version": "1.0.0",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "duplicate version",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)

				// Add existing version
				vRepo.AddVersion(&domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				})

				return svc.ID.Hex()
			},
			requestBody: map[string]string{
				"version": "1.0.0",
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo, versionRepo := setupVersionHandler()
			serviceID := tt.setupRepo(serviceRepo, versionRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/services/"+serviceID+"/versions", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("service_id", serviceID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestVersionHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) (string, string)
		expectedStatus int
	}{
		{
			name: "successful retrieval",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				sRepo.AddService(svc)

				ver := &domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				}
				vRepo.AddVersion(ver)

				return svc.ID.Hex(), ver.ID.Hex()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "version not found",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)
				return svc.ID.Hex(), primitive.NewObjectID().Hex()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo, versionRepo := setupVersionHandler()
			serviceID, versionID := tt.setupRepo(serviceRepo, versionRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+serviceID+"/versions/"+versionID, nil)
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("service_id", serviceID)
			rctx.URLParams.Add("version_id", versionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Get(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestVersionHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) (string, string)
		expectedStatus int
	}{
		{
			name: "successful deletion",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)

				ver := &domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				}
				vRepo.AddVersion(ver)

				return svc.ID.Hex(), ver.ID.Hex()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "version not found",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)
				return svc.ID.Hex(), primitive.NewObjectID().Hex()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo, versionRepo := setupVersionHandler()
			serviceID, versionID := tt.setupRepo(serviceRepo, versionRepo)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+serviceID+"/versions/"+versionID, nil)
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("service_id", serviceID)
			rctx.URLParams.Add("version_id", versionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Delete(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestVersionHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) string
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "list versions for service",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)

				for _, v := range []string{"1.0.0", "1.1.0", "2.0.0"} {
					vRepo.AddVersion(&domain.Version{
						ID:        primitive.NewObjectID(),
						ServiceID: svc.ID,
						Version:   v,
						CreatedAt: time.Now(),
					})
				}

				return svc.ID.Hex()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name: "empty version list",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				sRepo.AddService(svc)
				return svc.ID.Hex()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "service not found",
			setupRepo: func(sRepo *mocks.MockServiceRepository, vRepo *mocks.MockVersionRepository) string {
				return primitive.NewObjectID().Hex()
			},
			expectedStatus: http.StatusNotFound,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo, versionRepo := setupVersionHandler()
			serviceID := tt.setupRepo(serviceRepo, versionRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+serviceID+"/versions", nil)
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("service_id", serviceID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.List(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data, ok := resp["data"].([]interface{})
				require.True(t, ok)
				assert.Len(t, data, tt.expectedCount)
			}
		})
	}
}
