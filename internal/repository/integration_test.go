//go:build integration
// +build integration

package repository_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/services-api/internal/domain"
	"github.com/services-api/internal/repository"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	testDB      *mongo.Database
	testClient  *mongo.Client
	serviceRepo domain.ServiceRepository
	versionRepo domain.VersionRepository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongo:7.0")
	if err != nil {
		log.Fatalf("Failed to start MongoDB container: %v", err)
	}

	// Get connection string
	connStr, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to MongoDB
	testClient, err = repository.ConnectMongoDB(ctx, connStr)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Get database
	testDB = testClient.Database("test_services_db")

	// Initialize indexes
	if err := repository.EnsureIndexes(ctx, testDB); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Initialize repositories
	serviceRepo = repository.NewMongoServiceRepository(testDB)
	versionRepo = repository.NewMongoVersionRepository(testDB)

	// Run tests
	code := m.Run()

	// Cleanup
	if err := testClient.Disconnect(ctx); err != nil {
		log.Printf("Error disconnecting from MongoDB: %v", err)
	}

	if err := testcontainers.TerminateContainer(mongoContainer); err != nil {
		log.Printf("Error terminating container: %v", err)
	}

	os.Exit(code)
}

func cleanupCollections(t *testing.T) {
	ctx := context.Background()
	if err := testDB.Collection("services").Drop(ctx); err != nil {
		t.Logf("Warning: failed to drop services collection: %v", err)
	}
	if err := testDB.Collection("versions").Drop(ctx); err != nil {
		t.Logf("Warning: failed to drop versions collection: %v", err)
	}
	// Re-create indexes
	if err := repository.EnsureIndexes(ctx, testDB); err != nil {
		t.Fatalf("Failed to re-create indexes: %v", err)
	}
}

// 9.2 Integration tests for service CRUD operations
func TestServiceRepository_CRUD(t *testing.T) {
	cleanupCollections(t)
	ctx := context.Background()

	// Test Create
	service := &domain.Service{
		Name:        "test-service",
		Description: "Test description",
	}

	err := serviceRepo.Create(ctx, service)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	if service.ID.IsZero() {
		t.Error("Service ID should not be zero after creation")
	}
	if service.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if service.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// Test GetByID
	fetched, err := serviceRepo.GetByID(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}
	if fetched.Name != service.Name {
		t.Errorf("Expected name %s, got %s", service.Name, fetched.Name)
	}
	if fetched.Description != service.Description {
		t.Errorf("Expected description %s, got %s", service.Description, fetched.Description)
	}

	// Test Update
	fetched.Name = "updated-service"
	fetched.Description = "Updated description"
	err = serviceRepo.Update(ctx, fetched)
	if err != nil {
		t.Fatalf("Failed to update service: %v", err)
	}

	updated, err := serviceRepo.GetByID(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to get updated service: %v", err)
	}
	if updated.Name != "updated-service" {
		t.Errorf("Expected updated name, got %s", updated.Name)
	}
	if !updated.UpdatedAt.After(fetched.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}

	// Test Delete
	err = serviceRepo.Delete(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to delete service: %v", err)
	}

	_, err = serviceRepo.GetByID(ctx, service.ID.Hex())
	if err != domain.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

// 9.3 Integration tests for version CRUD operations
func TestVersionRepository_CRUD(t *testing.T) {
	cleanupCollections(t)
	ctx := context.Background()

	// Create a service first
	service := &domain.Service{
		Name:        "test-service",
		Description: "Test description",
	}
	if err := serviceRepo.Create(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Test Create Version
	version := &domain.Version{
		ServiceID: service.ID,
		Version:   "1.0.0",
	}

	err := versionRepo.Create(ctx, version)
	if err != nil {
		t.Fatalf("Failed to create version: %v", err)
	}
	if version.ID.IsZero() {
		t.Error("Version ID should not be zero after creation")
	}

	// Test GetByID
	fetched, err := versionRepo.GetByID(ctx, version.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}
	if fetched.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", fetched.Version)
	}
	if fetched.ServiceID != service.ID {
		t.Error("ServiceID mismatch")
	}

	// Test ListByServiceID
	versions, err := versionRepo.ListByServiceID(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to list versions: %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(versions))
	}

	// Test CountByServiceID
	count, err := versionRepo.CountByServiceID(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to count versions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Test Delete
	err = versionRepo.Delete(ctx, version.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to delete version: %v", err)
	}

	_, err = versionRepo.GetByID(ctx, version.ID.Hex())
	if err != domain.ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

// 9.4 Integration tests for service listing with filters, sort, pagination
func TestServiceRepository_List(t *testing.T) {
	cleanupCollections(t)
	ctx := context.Background()

	// Create multiple services
	services := []struct {
		name string
		desc string
	}{
		{"api-gateway", "API Gateway service"},
		{"auth-service", "Authentication service"},
		{"payment-service", "Payment processing"},
		{"notification-service", "Notification sender"},
		{"user-service", "User management"},
	}

	for _, s := range services {
		svc := &domain.Service{
			Name:        s.name,
			Description: s.desc,
		}
		if err := serviceRepo.Create(ctx, svc); err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Test default listing
	params := domain.DefaultListParams()
	result, err := serviceRepo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list services: %v", err)
	}
	if len(result.Data) != 5 {
		t.Errorf("Expected 5 services, got %d", len(result.Data))
	}

	// Test pagination
	params.Pagination.Limit = 2
	params.Pagination.Page = 1
	result, err = serviceRepo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list with pagination: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("Expected 2 services, got %d", len(result.Data))
	}
	if result.Pagination.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Pagination.Total)
	}
	if result.Pagination.TotalPages != 3 {
		t.Errorf("Expected 3 pages, got %d", result.Pagination.TotalPages)
	}

	// Test sorting by name ascending
	params = domain.ListParams{
		Sort:       "name",
		Order:      "asc",
		Pagination: domain.PaginationParams{Page: 1, Limit: 10},
	}
	result, err = serviceRepo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list with sorting: %v", err)
	}
	if result.Data[0].Name != "api-gateway" {
		t.Errorf("Expected first service to be api-gateway, got %s", result.Data[0].Name)
	}

	// Test search filter
	params = domain.ListParams{
		Search:     "payment",
		Pagination: domain.PaginationParams{Page: 1, Limit: 10},
	}
	result, err = serviceRepo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list with search: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("Expected 1 service matching 'payment', got %d", len(result.Data))
	}

	// Test name filter
	params = domain.ListParams{
		Name:       "auth-service",
		Pagination: domain.PaginationParams{Page: 1, Limit: 10},
	}
	result, err = serviceRepo.List(ctx, params)
	if err != nil {
		t.Fatalf("Failed to list with name filter: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("Expected 1 service with name 'auth-service', got %d", len(result.Data))
	}
}

// 9.5 Integration tests for cascade delete (service with versions)
func TestServiceRepository_CascadeDelete(t *testing.T) {
	cleanupCollections(t)
	ctx := context.Background()

	// Create service
	service := &domain.Service{
		Name:        "test-service",
		Description: "Test description",
	}
	if err := serviceRepo.Create(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Create versions
	for _, v := range []string{"1.0.0", "1.1.0", "2.0.0"} {
		version := &domain.Version{
			ServiceID: service.ID,
			Version:   v,
		}
		if err := versionRepo.Create(ctx, version); err != nil {
			t.Fatalf("Failed to create version: %v", err)
		}
	}

	// Verify versions exist
	count, err := versionRepo.CountByServiceID(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to count versions: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 versions, got %d", count)
	}

	// Delete service (should cascade to versions)
	if err := serviceRepo.Delete(ctx, service.ID.Hex()); err != nil {
		t.Fatalf("Failed to delete service: %v", err)
	}

	// Verify versions are deleted
	count, err = versionRepo.CountByServiceID(ctx, service.ID.Hex())
	if err != nil {
		t.Fatalf("Failed to count versions after cascade: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 versions after cascade delete, got %d", count)
	}
}

// 9.6 Integration tests for duplicate version prevention
func TestVersionRepository_DuplicatePrevention(t *testing.T) {
	cleanupCollections(t)
	ctx := context.Background()

	// Create service
	service := &domain.Service{
		Name:        "test-service",
		Description: "Test description",
	}
	if err := serviceRepo.Create(ctx, service); err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Create first version
	version1 := &domain.Version{
		ServiceID: service.ID,
		Version:   "1.0.0",
	}
	if err := versionRepo.Create(ctx, version1); err != nil {
		t.Fatalf("Failed to create first version: %v", err)
	}

	// Test ExistsByServiceIDAndVersion
	exists, err := versionRepo.ExistsByServiceIDAndVersion(ctx, service.ID.Hex(), "1.0.0")
	if err != nil {
		t.Fatalf("Failed to check version existence: %v", err)
	}
	if !exists {
		t.Error("Expected version to exist")
	}

	// Check non-existent version
	exists, err = versionRepo.ExistsByServiceIDAndVersion(ctx, service.ID.Hex(), "2.0.0")
	if err != nil {
		t.Fatalf("Failed to check version existence: %v", err)
	}
	if exists {
		t.Error("Expected version 2.0.0 to not exist")
	}

	// Different service with same version should not conflict
	service2 := &domain.Service{
		Name:        "another-service",
		Description: "Another description",
	}
	if err := serviceRepo.Create(ctx, service2); err != nil {
		t.Fatalf("Failed to create second service: %v", err)
	}

	version2 := &domain.Version{
		ServiceID: service2.ID,
		Version:   "1.0.0", // Same version string as service1
	}
	if err := versionRepo.Create(ctx, version2); err != nil {
		t.Fatalf("Should be able to create same version for different service: %v", err)
	}
}
