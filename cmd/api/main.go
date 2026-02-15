package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/services-api/internal/handler"
	"github.com/services-api/internal/repository"
	"github.com/services-api/internal/service"
	"github.com/services-api/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to MongoDB
	mongoClient, err := repository.ConnectMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Get database and initialize indexes
	db := mongoClient.Database(cfg.DBName)
	if err := repository.EnsureIndexes(ctx, db); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Initialize repositories
	serviceRepo := repository.NewMongoServiceRepository(db)

	// Initialize services
	serviceSvc := service.NewServiceService(serviceRepo)

	// Initialize handlers
	serviceHandler := handler.NewServiceHandler(serviceSvc)
	healthHandler := handler.NewHealthHandler(mongoClient)

	// Setup router
	router := handler.NewRouter(cfg, serviceHandler, healthHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
