# staticSend Makefile

.PHONY: help build test test-unit test-integration clean run dev

# Default target
help:
	@echo "Available commands:"
	@echo "  make test           - Run all tests (unit + integration)"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests only"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make dev            - Run in development mode with auto-reload"
	@echo "  make clean          - Clean build artifacts"

# Build the application
build:
	@echo "Building staticSend..."
	go build -o bin/staticsend ./cmd/staticsend

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test ./pkg/... -v -cover

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./integration_test.go

# Run all tests
test: test-unit test-integration

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Run the application
run: build
	@echo "Starting staticSend..."
	./bin/staticsend

# Development mode (requires air for auto-reload)
dev:
	@echo "Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		make run; \
	fi

# Install development dependencies
install-dev:
	@echo "Installing development dependencies..."
	go install github.com/cosmtrek/air@latest

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test ./pkg/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Lint the code (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Full check (format, lint, test)
check: fmt lint test

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t staticsend .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 staticsend
