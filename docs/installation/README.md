# Installation Guide

## Prerequisites

- Go 1.21 or later
- SQLite3 (for database)
- SMTP server credentials (for email notifications)
- Cloudflare Turnstile keys

## Quick Start

### From Source

```bash
# Clone the repository
git clone https://github.com/your-username/staticsend.git
cd staticsend

# Build the binary
go build -o staticsend ./cmd/staticsend/

# Set environment variables
export STATICSEND_PORT=8080
export STATICSEND_DB_PATH=./staticsend.db
export STATICSEND_JWT_SECRET=your-super-secret-jwt-key
export STATICSEND_SMTP_HOST=smtp.gmail.com
export STATICSEND_SMTP_PORT=587
export STATICSEND_SMTP_USER=your-email@gmail.com
export STATICSEND_SMTP_PASS=your-app-password

# Run the application
./staticsend
```

### Using Docker

```bash
# Create a directory for persistent data
mkdir -p /opt/staticsend/data

# Run with Docker
docker run -d \
  --name staticsend \
  -p 8080:8080 \
  -v /opt/staticsend/data:/data \
  -e STATICSEND_PORT=8080 \
  -e STATICSEND_DB_PATH=/data/staticsend.db \
  -e STATICSEND_JWT_SECRET=your-super-secret-jwt-key \
  -e STATICSEND_SMTP_HOST=smtp.gmail.com \
  -e STATICSEND_SMTP_PORT=587 \
  -e STATICSEND_SMTP_USER=your-email@gmail.com \
  -e STATICSEND_SMTP_PASS=your-app-password \
  your-username/staticsend:latest
```

## Development Setup

```bash
# Clone the repository
git clone https://github.com/your-username/staticsend.git
cd staticsend

# Install dependencies
go mod download

# Run tests
go test ./...

# Build and run
go run ./cmd/staticsend/

# Build for production
go build -o staticsend ./cmd/staticsend/
```