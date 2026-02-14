package repository

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates necessary indexes for the database
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	// Services collection indexes
	servicesCollection := db.Collection("services")

	// Index on name for search queries
	_, err := servicesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
	})
	if err != nil {
		return err
	}
	log.Println("Created index on services.name")

	// Text index on name and description for search
	_, err = servicesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: "text"},
			{Key: "description", Value: "text"},
		},
	})
	if err != nil {
		return err
	}
	log.Println("Created text index on services.name and services.description")

	// Versions collection indexes
	versionsCollection := db.Collection("versions")

	// Index on service_id for efficient lookups
	_, err = versionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "service_id", Value: 1}},
	})
	if err != nil {
		return err
	}
	log.Println("Created index on versions.service_id")

	// Unique compound index on (service_id, version) to prevent duplicates
	_, err = versionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "service_id", Value: 1},
			{Key: "version", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Println("Created unique compound index on versions.(service_id, version)")

	return nil
}
