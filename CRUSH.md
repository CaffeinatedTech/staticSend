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

## Project Planning with PLAN.md

### Purpose
Simple project tracking with checkable tasks and clear current focus.

### Structure
```
# PLAN.md - staticSend Development Plan

**Current Task:** [Brief description of what's being worked on now]
**Last Task:** [Brief description of the most recently completed task]

## Stage 1: Core Foundation
- [ ] Initialize Go module structure
- [ ] Set up basic HTTP server with chi router
- [ ] Create Turnstile validation service
- [ ] Implement rate limiting middleware
- [ ] Set up basic email notification service

## Stage 2: Data Persistence  
- [ ] Design database schema
- [ ] Set up SQLite database
- [ ] Implement user model and storage
- [ ] Implement contact form model and storage
- [ ] Implement submission tracking

## Stage 3: Web Interface
- [ ] Create authentication system
- [ ] Build HTMX-based management UI
- [ ] Implement form management interface
- [ ] Create submission history view
- [ ] Add user settings and configuration

## Stage 4: Deployment & Polish
- [ ] Create Docker configuration
- [ ] Set up environment configuration
- [ ] Write comprehensive documentation
- [ ] Add testing suite
- [ ] Prepare for production deployment
```

### Usage Guidelines
1. **AI Priority:** Maintaining PLAN.md is a top priority - update it before starting any work
2. **Update Current Task:** AI must update this line when beginning a new task
3. **Update Last Task:** Move completed task from Current to Last when work is finished
4. **Check off tasks:** Mark completed tasks with [x] in the appropriate stage
5. **Sequential Work:** Complete one task at a time - do not move ahead without user instruction
6. **Add New Tasks:** If new work is discussed, add it to the appropriate stage in PLAN.md first
7. **User Control:** Wait for explicit user instruction before moving to next task
8. **Commit Changes:** Always update PLAN.md alongside code changes
9. **Communication:** Inform user when a task is completed and wait for next instructions
