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

	// Service versions collection indexes
	versionsCollection := db.Collection("service_versions")

	// Compound index on service_id and revision for efficient lookups
	_, err = versionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "service_id", Value: 1},
			{Key: "revision", Value: -1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Println("Created compound unique index on service_versions(service_id, revision)")

	// Index on service_id for listing all versions of a service
	_, err = versionsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "service_id", Value: 1}},
	})
	if err != nil {
		return err
	}
	log.Println("Created index on service_versions.service_id")

	// Users collection indexes
	usersCollection := db.Collection("users")

	// Unique index on email for user lookup and preventing duplicates
	_, err = usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Println("Created unique index on users.email")

	return nil
}
