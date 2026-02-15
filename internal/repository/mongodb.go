package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	maxRetries    = 5
	initialDelay  = 1 * time.Second
	maxDelay      = 30 * time.Second
	backoffFactor = 2
)

// ConnectMongoDB establishes a connection to MongoDB with retry logic
func ConnectMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
	var client *mongo.Client
	var err error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		clientOptions := options.Client().ApplyURI(uri)

		// Set connection pool options
		clientOptions.SetMaxPoolSize(100)
		clientOptions.SetMinPoolSize(10)
		clientOptions.SetMaxConnIdleTime(30 * time.Second)

		client, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Printf("MongoDB connection attempt %d failed: %v", attempt, err)
			if attempt < maxRetries {
				log.Printf("Retrying in %v...", delay)
				time.Sleep(delay)
				delay = min(delay*backoffFactor, maxDelay)
				continue
			}
			return nil, fmt.Errorf("failed to connect to MongoDB after %d attempts: %w", maxRetries, err)
		}

		// Ping to verify connection
		if err = client.Ping(ctx, nil); err != nil {
			log.Printf("MongoDB ping attempt %d failed: %v", attempt, err)
			if attempt < maxRetries {
				log.Printf("Retrying in %v...", delay)
				time.Sleep(delay)
				delay = min(delay*backoffFactor, maxDelay)
				continue
			}
			return nil, fmt.Errorf("failed to ping MongoDB after %d attempts: %w", maxRetries, err)
		}

		log.Println("Successfully connected to MongoDB")
		return client, nil
	}

	return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
