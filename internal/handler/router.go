package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/services-api/pkg/auth"
	"github.com/services-api/pkg/config"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter creates and configures the Chi router with all routes
func NewRouter(
	cfg *config.Config,
	serviceHandler *ServiceHandler,
	healthHandler *HealthHandler,
) http.Handler {
	r := chi.NewRouter()

	// Configure Chi router with middleware (logging, recovery, CORS)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Register health endpoint (no auth)
	r.Get("/health", healthHandler.Check)

	// Swagger documentation endpoint (no auth)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Create auth middleware
	authMiddleware := auth.NewMiddleware(cfg)

	// Register /api/v1 routes with auth middleware
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth middleware to all /api/v1 routes
		r.Use(authMiddleware.Authenticate)

		// Service routes
		r.Route("/services", func(r chi.Router) {
			r.Post("/", serviceHandler.Create)
			r.Get("/", serviceHandler.List)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", serviceHandler.Get)
				r.Put("/", serviceHandler.Update)
				r.Patch("/", serviceHandler.Patch)
				r.Delete("/", serviceHandler.Delete)
			})
		})
	})

	return r
}
