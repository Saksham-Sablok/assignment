package repository

import (
	"context"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const serviceVersionsCollection = "service_versions"

// MongoServiceVersionRepository implements domain.ServiceVersionRepository using MongoDB
type MongoServiceVersionRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewMongoServiceVersionRepository creates a new MongoServiceVersionRepository
func NewMongoServiceVersionRepository(db *mongo.Database) *MongoServiceVersionRepository {
	return &MongoServiceVersionRepository{
		db:         db,
		collection: db.Collection(serviceVersionsCollection),
	}
}

// Create creates a new service version snapshot
func (r *MongoServiceVersionRepository) Create(ctx context.Context, version *domain.ServiceVersion) error {
	if version.ID.IsZero() {
		version.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, version)
	return err
}

// GetByID retrieves a service version by its ID
func (r *MongoServiceVersionRepository) GetByID(ctx context.Context, id string) (*domain.ServiceVersion, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	var version domain.ServiceVersion
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &version, nil
}

// GetByServiceIDAndRevision retrieves a specific revision of a service
func (r *MongoServiceVersionRepository) GetByServiceIDAndRevision(ctx context.Context, serviceID string, revision int) (*domain.ServiceVersion, error) {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	var version domain.ServiceVersion
	err = r.collection.FindOne(ctx, bson.M{
		"service_id": objectID,
		"revision":   revision,
	}).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &version, nil
}

// ListByServiceID retrieves all versions for a service with pagination
func (r *MongoServiceVersionRepository) ListByServiceID(ctx context.Context, serviceID string, params domain.PaginationParams) (*domain.PaginatedResult[domain.ServiceVersion], error) {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	filter := bson.M{"service_id": objectID}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Set up find options - sort by revision descending (newest first)
	findOptions := options.Find().
		SetSort(bson.D{{Key: "revision", Value: -1}}).
		SetSkip(int64(params.Offset())).
		SetLimit(int64(params.Limit))

	// Execute query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var versions []domain.ServiceVersion
	if err := cursor.All(ctx, &versions); err != nil {
		return nil, err
	}

	return domain.NewPaginatedResult(versions, total, params), nil
}

// DeleteByServiceID deletes all versions for a service
func (r *MongoServiceVersionRepository) DeleteByServiceID(ctx context.Context, serviceID string) error {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return domain.ErrInvalidID
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"service_id": objectID})
	return err
}
