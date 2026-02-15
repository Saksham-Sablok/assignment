package repository

import (
	"context"
	"time"

	"github.com/services-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const versionsCollection = "versions"

// MongoVersionRepository implements domain.VersionRepository using MongoDB
type MongoVersionRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
}

// NewMongoVersionRepository creates a new MongoVersionRepository
func NewMongoVersionRepository(db *mongo.Database) *MongoVersionRepository {
	return &MongoVersionRepository{
		db:         db,
		collection: db.Collection(versionsCollection),
	}
}

// Create creates a new version
func (r *MongoVersionRepository) Create(ctx context.Context, version *domain.Version) error {
	version.ID = primitive.NewObjectID()
	version.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, version)
	if mongo.IsDuplicateKeyError(err) {
		return domain.ErrDuplicateVersion
	}
	return err
}

// GetByID retrieves a version by its ID
func (r *MongoVersionRepository) GetByID(ctx context.Context, id string) (*domain.Version, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	var version domain.Version
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &version, nil
}

// Delete deletes a version by its ID
func (r *MongoVersionRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrInvalidID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ListByServiceID retrieves all versions for a specific service
func (r *MongoVersionRepository) ListByServiceID(ctx context.Context, serviceID string) ([]domain.Version, error) {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	// Sort by created_at descending (newest first)
	findOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"service_id": objectID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var versions []domain.Version
	if err := cursor.All(ctx, &versions); err != nil {
		return nil, err
	}

	return versions, nil
}

// CountByServiceID counts versions for a specific service
func (r *MongoVersionRepository) CountByServiceID(ctx context.Context, serviceID string) (int, error) {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return 0, domain.ErrInvalidID
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"service_id": objectID})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// DeleteByServiceID deletes all versions for a specific service
func (r *MongoVersionRepository) DeleteByServiceID(ctx context.Context, serviceID string) error {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return domain.ErrInvalidID
	}

	_, err = r.collection.DeleteMany(ctx, bson.M{"service_id": objectID})
	return err
}

// ExistsByServiceIDAndVersion checks if a version exists for a service
func (r *MongoVersionRepository) ExistsByServiceIDAndVersion(ctx context.Context, serviceID, version string) (bool, error) {
	objectID, err := primitive.ObjectIDFromHex(serviceID)
	if err != nil {
		return false, domain.ErrInvalidID
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{
		"service_id": objectID,
		"version":    version,
	})
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
