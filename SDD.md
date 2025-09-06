# Software Design Document: staticSend

## 1. Introduction

### 1.1. Purpose
staticSend is a self-hosted, secure contact form processing service that enables static websites to have functional contact forms without backend code. It validates submissions using Cloudflare Turnstile, applies rate limiting, and forwards legitimate messages to specified email addresses.

### 1.2. Problem Statement
Static websites lack server-side processing capabilities, making traditional contact forms impossible. staticSend solves this by providing a centralized service that validates, processes, and forwards form submissions while protecting against spam and abuse.

### 1.3. Goals
- Provide a secure endpoint for contact form submissions
- Validate submissions using Cloudflare Turnstile
- Implement robust rate limiting
- Offer a web UI for management
- Easy deployment via Docker

## 2. Architecture Overview

### 2.1. System Architecture
```
Static Website (HTML/CSS/JS)
        |
        | POST (Form data + Turnstile token)
        |
+-------------------------------+
|        staticSend API        |
|  +-------------------------+  |
|  |  Rate Limiting Middleware |  |
|  +-------------------------+  |
|  |  Turnstile Validation   |  |
|  +-------------------------+  |
|  |  Request Processing     |  |
|  +-------------------------+  |
+-------------------------------+
        |
        | SMTP or Email API
        |
    Email Inbox
```

### 2.2. Component Diagram
1. **Web UI**: HTMX-based interface for managing contact forms
2. **API Server**: Go HTTP server with endpoints for form processing and management
3. **Authentication**: JWT-based auth for the management UI
4. **Rate Limiter**: IP-based request limiting
5. **Turnstile Validator**: Cloudflare Turnstile verification service
6. **Email Service**: SMTP or transactional email API integration
7. **Data Storage**: SQLite for simplicity (can be extended to other databases)

## 3. Detailed Design

### 3.1. Core Components

#### 3.1.1. Authentication Service
- Handles user registration/login
- Issues JWT tokens for API access
- Password hashing with bcrypt

#### 3.1.2. Form Management Service
- CRUD operations for contact forms
- Generates unique keys for each form
- Stores destination email and Turnstile secrets

#### 3.1.3. Submission Processing Service
- Validates Turnstile tokens
- Applies rate limiting
- Forwards valid submissions via email

#### 3.1.4. Web UI Service
- Serves HTMX-based management interface
- Requires authentication for access

### 3.2. API Endpoints

#### Public Endpoints (form submission):
```
POST /api/v1/submit/{form_key}
Content-Type: application/x-www-form-urlencoded
Body: name=John&email=john@example.com&message=Hello&cf-turnstile-response=token
```

#### Management Endpoints (require auth):
```
POST    /api/auth/register
POST    /api/auth/login
GET     /api/forms
POST    /api/forms
GET     /api/forms/{id}
PUT     /api/forms/{id}
DELETE  /api/forms/{id}
GET     /api/submissions?form_id={id}
```

### 3.3. Data Models

#### User
```go
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`
}
```

#### ContactForm
```go
type ContactForm struct {
    ID             string    `json:"id"`
    UserID         string    `json:"user_id"`
    Name           string    `json:"name"`
    Domain         string    `json:"domain"`
    TurnstileKey   string    `json:"turnstile_key"`    // Public key
    TurnstileSecret string   `json:"turnstile_secret"` // Private key
    ForwardEmail   string    `json:"forward_email"`
    FormKey        string    `json:"form_key"`         // Generated unique key
    CreatedAt      time.Time `json:"created_at"`
}
```

#### FormSubmission
```go
type FormSubmission struct {
    ID        string    `json:"id"`
    FormID    string    `json:"form_id"`
    IPAddress string    `json:"ip_address"`
    Data      string    `json:"data"` // JSON of form fields
    CreatedAt time.Time `json:"created_at"`
    Status    string    `json:"status"` // pending, processed, failed
}
```

### 3.4. Key Algorithms & Logic

#### Turnstile Validation
1. Receive token from form submission
2. POST to Cloudflare's verification endpoint with secret key
3. Validate response including hostname matching
4. Process or reject based on result

#### Rate Limiting
- Token bucket algorithm per IP address
- Configurable limits (e.g., 5 requests per minute)
- Separate limits for authentication endpoints

#### Email Processing
- Template-based email formatting
- SMTP fallback with TLS
- Optional integration with transactional email services

## 4. User Interface Design

### 4.1. Web UI Components

#### Authentication Views
- Login form
- Registration form

#### Dashboard View
- Overview of forms and recent submissions
- Quick actions to create new forms

#### Form Management View
- List of all contact forms
- CRUD operations for forms
- Display of form key and integration code

#### Submission History View
- Paginated list of form submissions
- Filtering by form and date
- Export functionality

### 4.2. HTMX Implementation Strategy
- Server-side rendering with Go templates
- HTMX for dynamic interactions
- Alpine.js for client-side interactions where needed
- RESTful API endpoints for data operations

## 5. Security Considerations

### 5.1. Authentication & Authorization
- JWT tokens with secure signing
- Password hashing with work factor of 12
- HTTPS enforcement in production

### 5.2. Input Validation
- Strict validation of all inputs
- SQL injection prevention
- XSS protection through output encoding

### 5.3. Rate Limiting
- IP-based rate limiting on submission endpoint
- Account-level rate limiting on management API

### 5.4. Data Protection
- Encryption of sensitive data (Turnstile secrets)
- Secure email transmission

## 6. Deployment Strategy

### 6.1. Docker Configuration
```dockerfile
FROM golang:alpine AS builder
# Build steps

FROM alpine:latest
# Runtime configuration
EXPOSE 8080
CMD ["./staticSend"]
```

### 6.2. Environment Variables
```
STATICSEND_PORT=8080
STATICSEND_DB_PATH=/data/staticSend.db
STATICSEND_JWT_SECRET=your-secret-key
STATICSEND_SMTP_HOST=smtp.gmail.com
STATICSEND_SMTP_PORT=587
STATICSEND_SMTP_USER=your-email@gmail.com
STATICSEND_SMTP_PASS=your-app-password
```

### 6.3. Coolify Deployment
- Docker-based deployment
- Persistent volume for database
- Environment variable configuration

## 7. Implementation Plan

### Phase 1: Core API
- Basic HTTP server with routing
- Turnstile validation integration
- Rate limiting middleware
- Email notification service

### Phase 2: Data Persistence
- SQLite database setup
- User management
- Form configuration storage

### Phase 3: Web UI
- Authentication UI
- Form management interface
- Submission history view

### Phase 4: Deployment Packaging
- Docker containerization
- Documentation
- Example configurations

## 8. Testing Strategy

### 8.1. Unit Tests
- Individual component testing
- Mock external dependencies

### 8.2. Integration Tests
- API endpoint testing
- Database interaction tests
- Turnstile validation tests

### 8.3. End-to-End Tests
- Form submission workflow
- UI interaction tests

## 9. Future Enhancements

### 9.1. Potential Extensions
- Webhook support for form submissions
- Multiple email backend support (SendGrid, Mailgun, etc.)
- Advanced analytics and reporting
- Team collaboration features
- API keys for programmatic access

### 9.2. Scalability Considerations
- Database connection pooling
- Caching strategies
- Horizontal scaling possibilities

## 10. Appendix

### 10.1. Example Integration Code
```html
<!-- Example form HTML for static websites -->
<form action="https://your-safecontact-instance.com/api/v1/submit/FORM_KEY" 
      method="POST">
    <input type="text" name="name" placeholder="Your Name" required>
    <input type="email" name="email" placeholder="Your Email" required>
    <textarea name="message" placeholder="Your Message" required></textarea>
    
    <!-- Cloudflare Turnstile -->
    <div class="cf-turnstile" data-sitekey="TURNSTILE_PUBLIC_KEY"></div>
    
    <button type="submit">Send Message</button>
</form>
<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
```

### 10.2. Technology Choices
- **Go**: Performance, simplicity, strong standard library
- **SQLite**: Simple persistence, no external dependencies
- **HTMX**: Lightweight interactivity without client-side complexity
- **JWT**: Stateless authentication suitable for microservices

This document provides a comprehensive overview of the staticSend system design. The implementation should follow these specifications to create a secure, efficient, and easy-to-deploy solution for handling contact form submissions from static websites.
