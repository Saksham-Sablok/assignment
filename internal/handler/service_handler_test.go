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

func setupServiceHandler() (*handler.ServiceHandler, *mocks.MockServiceRepository) {
	serviceRepo := mocks.NewMockServiceRepository()
	svc := service.NewServiceService(serviceRepo)
	h := handler.NewServiceHandler(svc)
	return h, serviceRepo
}

func TestServiceHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			requestBody: map[string]string{
				"name":        "test-service",
				"description": "Test description",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			requestBody: map[string]string{
				"description": "Test description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "missing description",
			requestBody: map[string]string{
				"name": "test-service",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "description is required",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _ := setupServiceHandler()

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestServiceHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository) string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful retrieval",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				return primitive.NewObjectID().Hex()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo := setupServiceHandler()
			id := tt.setupRepo(serviceRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+id, nil)
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Get(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestServiceHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository) string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful update",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "original-name",
					Description: "Original description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			requestBody: map[string]string{
				"name":        "updated-name",
				"description": "Updated description",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				return primitive.NewObjectID().Hex()
			},
			requestBody: map[string]string{
				"name":        "updated-name",
				"description": "Updated description",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "validation error - missing name",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "original-name",
					Description: "Original description",
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			requestBody: map[string]string{
				"description": "Updated description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo := setupServiceHandler()
			id := tt.setupRepo(serviceRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/api/v1/services/"+id, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Update(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestServiceHandler_Patch(t *testing.T) {
	newName := "patched-name"

	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository) string
		requestBody    interface{}
		expectedStatus int
		expectedName   string
		expectedDesc   string
	}{
		{
			name: "patch name only",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "original-name",
					Description: "Original description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			requestBody: map[string]*string{
				"name": &newName,
			},
			expectedStatus: http.StatusOK,
			expectedName:   "patched-name",
			expectedDesc:   "Original description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo := setupServiceHandler()
			id := tt.setupRepo(serviceRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+id, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Patch(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp domain.ServiceResponse
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.expectedName, resp.Name)
				assert.Equal(t, tt.expectedDesc, resp.Description)
			}
		})
	}
}

func TestServiceHandler_Delete(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository) string
		expectedStatus int
	}{
		{
			name: "successful deletion",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "not found",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				return primitive.NewObjectID().Hex()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo := setupServiceHandler()
			id := tt.setupRepo(serviceRepo)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+id, nil)
			w := httptest.NewRecorder()

			// Set up chi context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Delete(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestServiceHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func(*mocks.MockServiceRepository)
		query          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "list all services",
			setupRepo: func(repo *mocks.MockServiceRepository) {
				for i := 0; i < 5; i++ {
					repo.AddService(&domain.Service{
						ID:          primitive.NewObjectID(),
						Name:        "service-" + string(rune('a'+i)),
						Description: "Description",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					})
				}
			},
			query:          "",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			name:           "empty list",
			setupRepo:      func(repo *mocks.MockServiceRepository) {},
			query:          "",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "with pagination",
			setupRepo: func(repo *mocks.MockServiceRepository) {
				for i := 0; i < 10; i++ {
					repo.AddService(&domain.Service{
						ID:          primitive.NewObjectID(),
						Name:        "service-" + string(rune('a'+i)),
						Description: "Description",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					})
				}
			},
			query:          "?page=1&limit=5",
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, serviceRepo := setupServiceHandler()
			tt.setupRepo(serviceRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/services"+tt.query, nil)
			w := httptest.NewRecorder()

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
