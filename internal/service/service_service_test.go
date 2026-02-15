package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/services-api/internal/domain"
	"github.com/services-api/internal/repository/mocks"
	"github.com/services-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestServiceService_Create(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.CreateServiceRequest
		wantErr error
	}{
		{
			name: "successful creation",
			req: domain.CreateServiceRequest{
				Name:        "test-service",
				Description: "Test service description",
			},
			wantErr: nil,
		},
		{
			name: "missing name",
			req: domain.CreateServiceRequest{
				Name:        "",
				Description: "Test description",
			},
			wantErr: domain.ErrNameRequired,
		},
		{
			name: "missing description",
			req: domain.CreateServiceRequest{
				Name:        "test-service",
				Description: "",
			},
			wantErr: domain.ErrDescriptionRequired,
		},
		{
			name: "name too long",
			req: domain.CreateServiceRequest{
				Name:        string(make([]byte, 256)),
				Description: "Test description",
			},
			wantErr: domain.ErrNameTooLong,
		},
		{
			name: "description too long",
			req: domain.CreateServiceRequest{
				Name:        "test-service",
				Description: string(make([]byte, 1001)),
			},
			wantErr: domain.ErrDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			versionRepo := mocks.NewMockServiceVersionRepository()
			svc := service.NewServiceService(serviceRepo, versionRepo)

			ctx := context.Background()
			result, err := svc.Create(ctx, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Description, result.Description)
				assert.False(t, result.ID.IsZero())
				assert.False(t, result.CreatedAt.IsZero())
				assert.False(t, result.UpdatedAt.IsZero())
			}
		})
	}
}

func TestServiceService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository)
		id        string
		wantErr   error
	}{
		{
			name: "successful retrieval",
			setupRepo: func(repo *mocks.MockServiceRepository) {
				repo.AddService(&domain.Service{
					ID:          primitive.NewObjectIDFromTimestamp(time.Now()),
					Name:        "test-service",
					Description: "Test description",
				})
			},
			wantErr: nil,
		},
		{
			name:      "service not found",
			setupRepo: func(repo *mocks.MockServiceRepository) {},
			id:        primitive.NewObjectID().Hex(),
			wantErr:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			tt.setupRepo(serviceRepo)
			versionRepo := mocks.NewMockServiceVersionRepository()
			svc := service.NewServiceService(serviceRepo, versionRepo)

			ctx := context.Background()

			// Get the ID to query
			var queryID string
			if tt.id != "" {
				queryID = tt.id
			} else {
				// Get the first service from the repo
				result, _ := serviceRepo.List(ctx, domain.DefaultListParams())
				if len(result.Data) > 0 {
					queryID = result.Data[0].ID.Hex()
				}
			}

			result, err := svc.GetByID(ctx, queryID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestServiceService_Update(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository) string
		req       domain.UpdateServiceRequest
		wantErr   error
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
			req: domain.UpdateServiceRequest{
				Name:        "updated-name",
				Description: "Updated description",
			},
			wantErr: nil,
		},
		{
			name: "missing name in update",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "original-name",
					Description: "Original description",
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			req: domain.UpdateServiceRequest{
				Name:        "",
				Description: "Updated description",
			},
			wantErr: domain.ErrNameRequired,
		},
		{
			name: "missing description in update",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "original-name",
					Description: "Original description",
				}
				repo.AddService(svc)
				return svc.ID.Hex()
			},
			req: domain.UpdateServiceRequest{
				Name:        "updated-name",
				Description: "",
			},
			wantErr: domain.ErrDescriptionRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			id := tt.setupRepo(serviceRepo)
			versionRepo := mocks.NewMockServiceVersionRepository()
			svc := service.NewServiceService(serviceRepo, versionRepo)

			ctx := context.Background()
			result, err := svc.Update(ctx, id, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Description, result.Description)
			}
		})
	}
}

func TestServiceService_Patch(t *testing.T) {
	newName := "patched-name"
	newDesc := "Patched description"

	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository) string
		req       domain.PatchServiceRequest
		wantErr   error
		wantName  string
		wantDesc  string
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
			req: domain.PatchServiceRequest{
				Name: &newName,
			},
			wantErr:  nil,
			wantName: "patched-name",
			wantDesc: "Original description",
		},
		{
			name: "patch description only",
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
			req: domain.PatchServiceRequest{
				Description: &newDesc,
			},
			wantErr:  nil,
			wantName: "original-name",
			wantDesc: "Patched description",
		},
		{
			name: "patch both fields",
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
			req: domain.PatchServiceRequest{
				Name:        &newName,
				Description: &newDesc,
			},
			wantErr:  nil,
			wantName: "patched-name",
			wantDesc: "Patched description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			id := tt.setupRepo(serviceRepo)
			versionRepo := mocks.NewMockServiceVersionRepository()
			svc := service.NewServiceService(serviceRepo, versionRepo)

			ctx := context.Background()
			result, err := svc.Patch(ctx, id, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantName, result.Name)
				assert.Equal(t, tt.wantDesc, result.Description)
			}
		})
	}
}

func TestServiceService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository) string
		wantErr   error
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
			wantErr: nil,
		},
		{
			name: "delete non-existent service",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				return primitive.NewObjectID().Hex()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "invalid id format",
			setupRepo: func(repo *mocks.MockServiceRepository) string {
				return "invalid-id"
			},
			wantErr: domain.ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			id := tt.setupRepo(serviceRepo)
			versionRepo := mocks.NewMockServiceVersionRepository()
			svc := service.NewServiceService(serviceRepo, versionRepo)

			ctx := context.Background()
			err := svc.Delete(ctx, id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				// Verify service is deleted
				_, err := serviceRepo.GetByID(ctx, id)
				assert.ErrorIs(t, err, domain.ErrNotFound)
			}
		})
	}
}

func TestServiceService_List(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository)
		params    domain.ListParams
		wantCount int
		wantErr   error
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
			params:    domain.DefaultListParams(),
			wantCount: 5,
			wantErr:   nil,
		},
		{
			name:      "empty list",
			setupRepo: func(repo *mocks.MockServiceRepository) {},
			params:    domain.DefaultListParams(),
			wantCount: 0,
			wantErr:   nil,
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
			params: domain.ListParams{
				Pagination: domain.PaginationParams{
					Page:  1,
					Limit: 5,
				},
			},
			wantCount: 5,
			wantErr:   nil,
		},
		{
			name: "invalid sort field",
			setupRepo: func(repo *mocks.MockServiceRepository) {
				repo.AddService(&domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Description",
				})
			},
			params: domain.ListParams{
				Sort: "invalid_field",
			},
			wantCount: 0,
			wantErr:   domain.ErrInvalidSortField,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			tt.setupRepo(serviceRepo)
			versionRepo := mocks.NewMockServiceVersionRepository()
			svc := service.NewServiceService(serviceRepo, versionRepo)

			ctx := context.Background()
			result, err := svc.List(ctx, tt.params)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Len(t, result.Data, tt.wantCount)
			}
		})
	}
}
