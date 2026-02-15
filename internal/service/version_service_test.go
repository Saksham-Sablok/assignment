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

func TestVersionService_Create(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) string
		req       domain.CreateVersionRequest
		wantErr   error
	}{
		{
			name: "successful creation",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				serviceRepo.AddService(svc)
				return svc.ID.Hex()
			},
			req: domain.CreateVersionRequest{
				Version: "1.0.0",
			},
			wantErr: nil,
		},
		{
			name: "missing version",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				serviceRepo.AddService(svc)
				return svc.ID.Hex()
			},
			req: domain.CreateVersionRequest{
				Version: "",
			},
			wantErr: domain.ErrVersionRequired,
		},
		{
			name: "service not found",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				return primitive.NewObjectID().Hex()
			},
			req: domain.CreateVersionRequest{
				Version: "1.0.0",
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "invalid service ID",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				return "invalid-id"
			},
			req: domain.CreateVersionRequest{
				Version: "1.0.0",
			},
			wantErr: domain.ErrInvalidID,
		},
		{
			name: "duplicate version",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				serviceRepo.AddService(svc)

				// Add existing version
				versionRepo.AddVersion(&domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				})

				return svc.ID.Hex()
			},
			req: domain.CreateVersionRequest{
				Version: "1.0.0",
			},
			wantErr: domain.ErrDuplicateVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			versionRepo := mocks.NewMockVersionRepository()
			serviceID := tt.setupRepo(serviceRepo, versionRepo)
			svc := service.NewVersionService(versionRepo, serviceRepo)

			ctx := context.Background()
			result, err := svc.Create(ctx, serviceID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Version, result.Version)
				assert.False(t, result.ID.IsZero())
				assert.False(t, result.CreatedAt.IsZero())
			}
		})
	}
}

func TestVersionService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) (string, string)
		wantErr   error
	}{
		{
			name: "successful retrieval",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				serviceRepo.AddService(svc)

				ver := &domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				}
				versionRepo.AddVersion(ver)

				return svc.ID.Hex(), ver.ID.Hex()
			},
			wantErr: nil,
		},
		{
			name: "version not found",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				serviceRepo.AddService(svc)
				return svc.ID.Hex(), primitive.NewObjectID().Hex()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "service not found",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				return primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "version belongs to different service",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				svc1 := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "service-1",
					Description: "Test description",
				}
				svc2 := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "service-2",
					Description: "Test description",
				}
				serviceRepo.AddService(svc1)
				serviceRepo.AddService(svc2)

				// Version belongs to svc1 but we'll query with svc2
				ver := &domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc1.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				}
				versionRepo.AddVersion(ver)

				return svc2.ID.Hex(), ver.ID.Hex()
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			versionRepo := mocks.NewMockVersionRepository()
			serviceID, versionID := tt.setupRepo(serviceRepo, versionRepo)
			svc := service.NewVersionService(versionRepo, serviceRepo)

			ctx := context.Background()
			result, err := svc.GetByID(ctx, serviceID, versionID)

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

func TestVersionService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) (string, string)
		wantErr   error
	}{
		{
			name: "successful deletion",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				serviceRepo.AddService(svc)

				ver := &domain.Version{
					ID:        primitive.NewObjectID(),
					ServiceID: svc.ID,
					Version:   "1.0.0",
					CreatedAt: time.Now(),
				}
				versionRepo.AddVersion(ver)

				return svc.ID.Hex(), ver.ID.Hex()
			},
			wantErr: nil,
		},
		{
			name: "version not found",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				serviceRepo.AddService(svc)
				return svc.ID.Hex(), primitive.NewObjectID().Hex()
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "service not found",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) (string, string) {
				return primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			versionRepo := mocks.NewMockVersionRepository()
			serviceID, versionID := tt.setupRepo(serviceRepo, versionRepo)
			svc := service.NewVersionService(versionRepo, serviceRepo)

			ctx := context.Background()
			err := svc.Delete(ctx, serviceID, versionID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				// Verify version is deleted
				_, err := versionRepo.GetByID(ctx, versionID)
				assert.ErrorIs(t, err, domain.ErrNotFound)
			}
		})
	}
}

func TestVersionService_ListByServiceID(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func(*mocks.MockServiceRepository, *mocks.MockVersionRepository) string
		wantCount int
		wantErr   error
	}{
		{
			name: "list versions for service",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				serviceRepo.AddService(svc)

				// Add multiple versions
				for _, v := range []string{"1.0.0", "1.1.0", "2.0.0"} {
					versionRepo.AddVersion(&domain.Version{
						ID:        primitive.NewObjectID(),
						ServiceID: svc.ID,
						Version:   v,
						CreatedAt: time.Now(),
					})
				}

				return svc.ID.Hex()
			},
			wantCount: 3,
			wantErr:   nil,
		},
		{
			name: "empty version list",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				svc := &domain.Service{
					ID:          primitive.NewObjectID(),
					Name:        "test-service",
					Description: "Test description",
				}
				serviceRepo.AddService(svc)
				return svc.ID.Hex()
			},
			wantCount: 0,
			wantErr:   nil,
		},
		{
			name: "service not found",
			setupRepo: func(serviceRepo *mocks.MockServiceRepository, versionRepo *mocks.MockVersionRepository) string {
				return primitive.NewObjectID().Hex()
			},
			wantCount: 0,
			wantErr:   domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceRepo := mocks.NewMockServiceRepository()
			versionRepo := mocks.NewMockVersionRepository()
			serviceID := tt.setupRepo(serviceRepo, versionRepo)
			svc := service.NewVersionService(versionRepo, serviceRepo)

			ctx := context.Background()
			result, err := svc.ListByServiceID(ctx, serviceID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}
