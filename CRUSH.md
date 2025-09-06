# CRUSH.md - staticSend Codebase Guidelines

## Build & Development Commands
- `go build` - Build the binary
- `go run .` - Run the application
- `go test ./...` - Run all tests
- `go test -v ./pkg/...` - Run package tests with verbose output
- `go test -run TestName` - Run specific test
- `go mod tidy` - Clean up dependencies
- `go fmt ./...` - Format all Go code
- `go vet ./...` - Run static analysis

## Code Style & Conventions

### Go Style
- Use Go 1.21+ features
- Follow standard Go formatting (gofmt)
- Use `camelCase` for variables and functions
- Use `PascalCase` for exported identifiers
- Prefer short, descriptive names

### Imports
- Group imports: stdlib, third-party, local
- Use absolute import paths for local packages
- Avoid dot imports

### Error Handling
- Always handle errors explicitly
- Use `fmt.Errorf("context: %w", err)` for wrapped errors
- Return early on errors
- Use sentinel errors for expected error conditions

### Types & Structs
- Use meaningful type names
- Add JSON tags for API structs
- Include comments for exported types
- Use `time.Time` for timestamps

### HTTP & API
- Use chi router for HTTP routing
- JSON responses should use consistent structure
- Validate all incoming requests
- Use middleware for common functionality (auth, logging)

### Testing
- Table-driven tests for complex logic
- Use testify/assert for assertions
- Mock external dependencies
- Test both success and error cases

### Database
- Use sqlc for type-safe SQL queries
- SQLite for development and production
- Use transactions for write operations
- Handle database migrations with goose

## Project Structure
```
cmd/
  staticsend/     # Main application
pkg/
  api/           # HTTP handlers and routing
  auth/          # Authentication logic
  database/      # Database operations
  email/         # Email sending service
  middleware/    # HTTP middleware
  models/        # Data models
  turnstile/     # Cloudflare Turnstile integration
  utils/         # Utility functions
internal/
  # Private application code
templates/       # HTML templates for HTMX UI
migrations/      # Database migrations
```

## Key Dependencies
- chi: HTTP router
- sqlc: SQL code generation
- testify: Testing utilities
- go-sqlite3: SQLite driver
- jwt-go: JWT authentication
- mail: SMTP email sending
