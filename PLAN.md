# PLAN.md - staticSend Development Plan

**Current Task:** Build HTMX-based management UI
**Last Task:** Create authentication system

## Stage 1: Core Foundation
- [x] Initialize Go module structure
- [x] Set up basic HTTP server with chi router
- [x] Create initial test suite
- [x] Create Turnstile validation service
- [x] Implement rate limiting middleware
- [x] Set up basic email notification service
- [x] Create project documentation structure

## Stage 2: Data Persistence  
- [x] Design database schema
- [x] Set up SQLite database
- [x] Implement user model and storage
- [x] Implement contact form model and storage
- [x] Implement submission tracking
- [x] Create database migrations

## Stage 3: Web Interface
- [x] Create authentication system
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