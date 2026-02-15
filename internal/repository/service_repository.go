package repository

import (
	"context"
	"strings"
	"time"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const servicesCollection = "services"

// MongoServiceRepository implements domain.ServiceRepository using MongoDB
type MongoServiceRepository struct {
	db          *mongo.Database
	collection  *mongo.Collection
	versionColl *mongo.Collection
}

// NewMongoServiceRepository creates a new MongoServiceRepository
func NewMongoServiceRepository(db *mongo.Database) *MongoServiceRepository {
	return &MongoServiceRepository{
		db:          db,
		collection:  db.Collection(servicesCollection),
		versionColl: db.Collection(versionsCollection),
	}
}

// Create creates a new service
func (r *MongoServiceRepository) Create(ctx context.Context, service *domain.Service) error {
	service.ID = primitive.NewObjectID()
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, service)
	return err
}

// GetByID retrieves a service by its ID, including version count
func (r *MongoServiceRepository) GetByID(ctx context.Context, id string) (*domain.Service, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	var service domain.Service
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&service)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	// Get version count
	count, err := r.versionColl.CountDocuments(ctx, bson.M{"service_id": objectID})
	if err != nil {
		return nil, err
	}
	service.VersionCount = int(count)

	return &service, nil
}

// Update updates an existing service
func (r *MongoServiceRepository) Update(ctx context.Context, service *domain.Service) error {
	service.UpdatedAt = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": service.ID},
		bson.M{"$set": bson.M{
			"name":        service.Name,
			"description": service.Description,
			"updated_at":  service.UpdatedAt,
		}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a service by its ID and all associated versions
func (r *MongoServiceRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrInvalidID
	}

	// Delete all versions for this service
	_, err = r.versionColl.DeleteMany(ctx, bson.M{"service_id": objectID})
	if err != nil {
		return err
	}

	// Delete the service
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// List retrieves services with filtering, sorting, and pagination
func (r *MongoServiceRepository) List(ctx context.Context, params domain.ListParams) (*domain.PaginatedResult[domain.Service], error) {
	filter := bson.M{}

	// Apply name filter (exact match, case-insensitive)
	if params.Name != "" {
		filter["name"] = bson.M{"$regex": "^" + params.Name + "$", "$options": "i"}
	}

	// Apply search filter (partial match on name or description)
	if params.Search != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": params.Search, "$options": "i"}},
			{"description": bson.M{"$regex": params.Search, "$options": "i"}},
		}
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Determine sort order
	sortOrder := -1 // descending
	if strings.ToLower(params.Order) == "asc" {
		sortOrder = 1
	}

	sortField := params.Sort
	if sortField == "" {
		sortField = "created_at"
	}

	// Set up find options
	findOptions := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
		SetSkip(int64(params.Pagination.Offset())).
		SetLimit(int64(params.Pagination.Limit))

	// Execute query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var services []domain.Service
	if err := cursor.All(ctx, &services); err != nil {
		return nil, err
	}

	// Get version counts for each service
	for i := range services {
		count, err := r.versionColl.CountDocuments(ctx, bson.M{"service_id": services[i].ID})
		if err != nil {
			return nil, err
		}
		services[i].VersionCount = int(count)
	}

	return domain.NewPaginatedResult(services, total, params.Pagination), nil
}
