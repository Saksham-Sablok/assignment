# Services API

A Go-based RESTful API for managing services with automatic revision tracking. This API powers a service dashboard widget, allowing users to view, search, and navigate to services in their organization.

## Features

- Full CRUD operations for services
- Automatic revision tracking (increments on every update)
- Filtering, sorting, and pagination for service listings
- API key authentication
- MongoDB persistence with proper indexing
- Swagger/OpenAPI documentation
- Clean architecture with dependency injection
- Comprehensive unit tests

## Prerequisites

- Docker and Docker Compose
- Go 1.24 or higher (only for local development without Docker)

## Project Structure

```
.
├── cmd/api/                    # Application entrypoint
├── docs/                       # Generated Swagger documentation
├── internal/
│   ├── domain/                 # Domain models and interfaces
│   ├── handler/                # HTTP handlers (presentation layer)
│   ├── repository/             # Data access layer (MongoDB)
│   │   └── mocks/             # Mock implementations for testing
│   └── service/               # Business logic layer
└── pkg/
    ├── auth/                   # Authentication middleware
    ├── config/                 # Configuration management
    └── response/              # HTTP response helpers
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | Database name | `services_db` |
| `PORT` | API server port | `8080` |
| `API_KEYS` | Comma-separated list of valid API keys | (none) |

## Quick Start with Docker Compose

The easiest way to run the application:

```bash
# Start the API and MongoDB
docker-compose up -d

# Check that services are running
docker-compose ps

# View logs
docker-compose logs -f api

# Stop the services
docker-compose down

# Stop and remove all data
docker-compose down -v
```

The API will be available at `http://localhost:8080`.

**Default API Keys:** `test-api-key-123`, `dev-key-456`

### Verify it's working

```bash
# Health check (no auth required)
curl http://localhost:8080/health

# Create a service
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{"name": "my-service", "description": "My first service"}'

# List services
curl http://localhost:8080/api/v1/services \
  -H "X-API-Key: test-api-key-123"
```

## Swagger Documentation

Interactive API documentation is available via Swagger UI:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json

The Swagger UI allows you to explore and test all API endpoints directly from your browser.

### Regenerating Swagger Docs

If you modify the API annotations, regenerate the documentation:

```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
```

## Local Development Setup

If you want to run without Docker:

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd assignment
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Start MongoDB:
   ```bash
   docker run -d -p 27017:27017 --name mongodb mongo:7.0
   ```

4. Set environment variables:
   ```bash
   export MONGODB_URI=mongodb://localhost:27017
   export DB_NAME=services_db
   export PORT=8080
   export API_KEYS=my-api-key
   ```

5. Run the API:
   ```bash
   go run cmd/api/main.go
   ```

## API Endpoints

### Health Check
```
GET /health
```
No authentication required.

### Services

All `/api/v1/*` endpoints require the `X-API-Key` header.

#### Create Service
```bash
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"name": "payment-service", "description": "Handles payment processing"}'
```

Response:
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "payment-service",
  "description": "Handles payment processing",
  "revision": 1,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### List Services
```bash
# Basic listing
curl http://localhost:8080/api/v1/services \
  -H "X-API-Key: your-api-key"

# With filtering and pagination
curl "http://localhost:8080/api/v1/services?search=payment&sort=name&order=asc&page=1&limit=10" \
  -H "X-API-Key: your-api-key"
```

Query Parameters:
- `search`: Search in name and description (case-insensitive)
- `name`: Filter by exact name (case-insensitive)
- `sort`: Sort field (`name`, `created_at`, `updated_at`)
- `order`: Sort order (`asc`, `desc`)
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

#### Get Service
```bash
curl http://localhost:8080/api/v1/services/{id} \
  -H "X-API-Key: your-api-key"
```

#### Update Service (Full)
```bash
curl -X PUT http://localhost:8080/api/v1/services/{id} \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"name": "updated-name", "description": "Updated description"}'
```

The revision is automatically incremented on every update.

#### Update Service (Partial)
```bash
curl -X PATCH http://localhost:8080/api/v1/services/{id} \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{"name": "new-name"}'
```

The revision is automatically incremented on every patch.

#### Delete Service
```bash
curl -X DELETE http://localhost:8080/api/v1/services/{id} \
  -H "X-API-Key: your-api-key"
```

## Automatic Revision Tracking

Each service has a `revision` field that tracks changes:

- **On creation**: Revision starts at `1`
- **On update (PUT/PATCH)**: Revision is atomically incremented using MongoDB's `$inc` operator

This provides a simple way to track how many times a service has been modified without maintaining a separate version history.

Example:
```bash
# Create a service (revision: 1)
curl -X POST http://localhost:8080/api/v1/services \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{"name": "my-service", "description": "Initial description"}'

# Update the service (revision: 2)
curl -X PUT http://localhost:8080/api/v1/services/{id} \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{"name": "my-service", "description": "Updated description"}'

# Patch the service (revision: 3)
curl -X PATCH http://localhost:8080/api/v1/services/{id} \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-123" \
  -d '{"description": "Another update"}'
```

## Testing

### Run Unit Tests
```bash
go test ./... -v
```

### Run with Coverage
```bash
go test ./... -cover
```

## Development

### Build
```bash
go build -o api ./cmd/api
```

### Run
```bash
./api
```

## Error Responses

All error responses follow this format:
```json
{
  "error": "error_code",
  "message": "Human readable message"
}
```

Error Codes:
- `bad_request` (400): Invalid input or validation error
- `unauthorized` (401): Missing or invalid API key
- `not_found` (404): Resource not found
- `internal_error` (500): Server error
