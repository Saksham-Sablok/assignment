# Services API

A Go-based RESTful API for managing services with automatic revision tracking. This API powers a service dashboard widget, allowing users to view, search, and navigate to services in their organization.

## Features

- Full CRUD operations for services
- Automatic revision tracking (increments on every update)
- Filtering, sorting, and pagination for service listings
- **Dual authentication support:**
  - JWT-based authentication (username/password) with access and refresh tokens
  - API key authentication for programmatic/service-to-service access
- User management with role-based access control (user/admin roles)
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
    ├── jwt/                    # JWT token management
    └── response/              # HTTP response helpers
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `DB_NAME` | Database name | `services_db` |
| `PORT` | API server port | `8080` |
| `API_KEYS` | Comma-separated list of valid API keys | (none) |
| `JWT_SECRET` | Secret key for signing JWT tokens | (required for JWT auth) |
| `JWT_ACCESS_EXPIRY` | Access token expiry duration | `15m` |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry duration | `168h` (7 days) |
| `JWT_ISSUER` | JWT issuer claim | `services-api` |

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
   export JWT_SECRET=your-secret-key-for-development
   export JWT_ACCESS_EXPIRY=15m
   export JWT_REFRESH_EXPIRY=168h
   export JWT_ISSUER=services-api
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

### Authentication

The API supports two authentication methods:
1. **JWT Authentication**: For user-based access with role-based permissions
2. **API Key Authentication**: For programmatic/service-to-service access

#### Register a New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
    "active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

#### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }'
```

#### Using JWT Authentication
Include the access token in the `Authorization` header:
```bash
curl http://localhost:8080/api/v1/services \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

#### Using API Key Authentication
Include the API key in the `X-API-Key` header:
```bash
curl http://localhost:8080/api/v1/services \
  -H "X-API-Key: your-api-key"
```

### User Management

#### Get Current User Profile
```bash
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <access_token>"
```

#### Change Password
```bash
curl -X POST http://localhost:8080/api/v1/users/me/password \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "oldpassword123",
    "new_password": "newpassword123"
  }'
```

#### Admin: Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "securepassword123",
    "first_name": "Jane",
    "last_name": "Smith",
    "role": "user"
  }'
```

#### Admin: List Users
```bash
curl "http://localhost:8080/api/v1/users?page=1&limit=20" \
  -H "Authorization: Bearer <admin_access_token>"
```

#### Admin: Update User
```bash
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H "Authorization: Bearer <admin_access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "updated@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "role": "admin",
    "active": true
  }'
```

#### Admin: Delete User
```bash
curl -X DELETE http://localhost:8080/api/v1/users/{id} \
  -H "Authorization: Bearer <admin_access_token>"
```

### Services

All `/api/v1/services/*` endpoints require authentication (JWT Bearer token or API key).

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
- `unauthorized` (401): Missing or invalid credentials (API key or JWT token)
- `forbidden` (403): Authenticated but not authorized for this action
- `not_found` (404): Resource not found
- `internal_error` (500): Server error

## User Roles

The API supports two user roles:

| Role | Permissions |
|------|-------------|
| `user` | Can manage their own profile, change password, access services |
| `admin` | All user permissions + create/read/update/delete any user |
