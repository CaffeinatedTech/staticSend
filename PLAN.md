# PLAN.md - staticSend Development Plan

**Current Task:** Create Turnstile validation service
**Last Task:** Set up basic HTTP server with chi router

## Stage 1: Core Foundation
- [x] Initialize Go module structure
- [x] Set up basic HTTP server with chi router
- [x] Create initial test suite
- [ ] Create Turnstile validation service
- [ ] Implement rate limiting middleware
- [ ] Set up basic email notification service
- [ ] Create project documentation structure

## Stage 2: Data Persistence  
- [ ] Design database schema
- [ ] Set up SQLite database
- [ ] Implement user model and storage
- [ ] Implement contact form model and storage
- [ ] Implement submission tracking
- [ ] Create database migrations

## Stage 3: Web Interface
- [ ] Create authentication system
- [ ] Build HTMX-based management UI
- [ ] Implement form management interface
- [ ] Create submission history view
- [ ] Add user settings and configuration
- [ ] Implement authentication UI (login/register)

## Stage 4: API Implementation
- [ ] Create form submission endpoint
- [ ] Implement Turnstile validation endpoint
- [ ] Add rate limiting to submission API
- [ ] Create management API endpoints
- [ ] Implement JWT authentication for API

## Stage 5: Deployment & Polish
- [ ] Create Docker configuration
- [ ] Set up environment configuration system
- [ ] Write comprehensive documentation
- [ ] Add testing suite
- [ ] Prepare for production deployment
- [ ] Create example integration code