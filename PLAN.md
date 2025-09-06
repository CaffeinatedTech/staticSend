# PLAN.md - staticSend Development Plan

**Current Task:** Add rate limiting to submission API
**Last Task:** Create submission history view

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
- [x] Build HTMX-based management UI
- [x] Implement form management interface
    - [x] Create forms
    - [x] View form details
    - [x] Edit forms
    - [x] Delete forms
- [x] Add user settings and configuration
    - [x] Registration disable/enable setting
    - [x] Database persistence for settings
    - [x] Registration route protection
- [x] Implement authentication UI (login/register)
- [x] Create submission history view

## Stage 4: Submission API & Testing
- [x] Create form submission endpoint
- [ ] Implement Turnstile validation endpoint
- [ ] Add rate limiting to submission API
- [ ] Add comprehensive test coverage for all packages
- [ ] Implement integration tests for submission flow

## Stage 5: Deployment & Polish
- [x] Add favicon and static assets
- [ ] Create Docker configuration
- [ ] Set up environment configuration system
- [ ] Write comprehensive documentation
- [ ] Prepare for production deployment
- [ ] Create example integration code