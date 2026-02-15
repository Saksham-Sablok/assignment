# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Allow building with newer module requirements
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies - allow toolchain auto-download
RUN go mod download || true

# Copy source code
COPY . .

# Build the binary (skip tests, only build main)
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /api ./cmd/api

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /api /app/api

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/api"]
