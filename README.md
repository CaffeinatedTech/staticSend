# staticSend

A self-hosted, secure contact form processing service for static websites. Enable functional contact forms on your static sites without backend code, with built-in spam protection and email forwarding.

[![Go](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com)

## ✨ Features

- **🔒 Cloudflare Turnstile Integration** - Bot protection with zero user friction
- **⏱️ Rate Limiting** - IP-based request limiting to prevent abuse
- **📧 Email Forwarding** - Send form submissions directly to your inbox
- **🖥️ Web Management UI** - HTMX-based interface for easy form management
- **🐳 Docker Ready** - Easy deployment with containerization
- **💾 SQLite Database** - Simple, file-based persistence
- **🔐 JWT Authentication** - Secure admin access
- **📱 Responsive Design** - Mobile-friendly management interface

## 🚀 Quick Start

### Prerequisites

- Go 1.21+ (for development)
- Docker (for production deployment)
- Cloudflare Turnstile keys ([get them here](https://dash.cloudflare.com/?to=/:account/turnstile))
- SMTP credentials (Gmail, SendGrid, etc.)

### Docker Deployment

```bash
# Create a directory for persistent data
mkdir -p /opt/staticsend/data

# Run with Docker
docker run -d \
  --name staticsend \
  -p 8080:8080 \
  -v /opt/staticsend/data:/data \
  -e STATICSEND_PORT=8080 \
  -e STATICSEND_JWT_SECRET=your-super-secret-jwt-key \
  -e STATICSEND_SMTP_HOST=smtp.gmail.com \
  -e STATICSEND_SMTP_PORT=587 \
  -e STATICSEND_SMTP_USER=your-email@gmail.com \
  -e STATICSEND_SMTP_PASS=your-app-password \
  ghcr.io/CaffeinatedTech/staticsend:latest
```

### Manual Installation

```bash
# Clone the repository
git clone https://github.com/CaffeinatedTech/staticsend.git
cd staticsend

# Build the binary
go build -o staticsend .

# Set environment variables
export STATICSEND_PORT=8080
export STATICSEND_DB_PATH=./staticsend.db
export STATICSEND_JWT_SECRET=your-jwt-secret
export STATICSEND_SMTP_HOST=smtp.gmail.com
export STATICSEND_SMTP_PORT=587
export STATICSEND_SMTP_USER=your-email@gmail.com
export STATICSEND_SMTP_PASS=your-app-password

# Run the application
./staticsend

# Or with custom port
./staticsend -port=3000
```

## 📋 Configuration

### Command Line Flags

| Flag | Description | Default | Environment Variable |
|------|-------------|---------|---------------------|
| `-port` | HTTP server port | `8080` | `STATICSEND_PORT` |

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `STATICSEND_PORT` | HTTP server port | `8080` | No |
| `STATICSEND_DB_PATH` | SQLite database path | `./staticsend.db` | No |
| `STATICSEND_JWT_SECRET` | JWT signing secret | - | Yes |
| `STATICSEND_SMTP_HOST` | SMTP server host | - | Yes |
| `STATICSEND_SMTP_PORT` | SMTP server port | - | Yes |
| `STATICSEND_SMTP_USER` | SMTP username | - | Yes |
| `STATICSEND_SMTP_PASS` | SMTP password | - | Yes |
| `STATICSEND_TURNSTILE_VERIFY_URL` | Turnstile verify URL | `https://challenges.cloudflare.com/turnstile/v0/siteverify` | No |

## 🛠️ Usage

### 1. Create a Contact Form

1. Access the web UI at `http://localhost:8080`
2. Register an account and log in
3. Create a new contact form with:
   - Form name and domain
   - Cloudflare Turnstile keys (public and secret)
   - Destination email address

### 2. Integrate with Your Static Site

Add this HTML to your static website:

```html
<form action="https://your-staticsend-instance.com/api/v1/submit/YOUR_FORM_KEY" 
      method="POST">
    <input type="text" name="name" placeholder="Your Name" required>
    <input type="email" name="email" placeholder="Your Email" required>
    <textarea name="message" placeholder="Your Message" required></textarea>
    
    <!-- Cloudflare Turnstile -->
    <div class="cf-turnstile" data-sitekey="YOUR_TURNSTILE_PUBLIC_KEY"></div>
    
    <button type="submit">Send Message</button>
</form>
<script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
```

### 3. Receive Submissions

Form submissions will be:
1. Validated by Cloudflare Turnstile
2. Rate-limited by IP address
3. Forwarded to your specified email address
4. Stored in the database for review

## 🔌 API Reference

### Public Endpoints

#### Submit Form
```http
POST /api/v1/submit/{form_key}
Content-Type: application/x-www-form-urlencoded

name=John&email=john@example.com&message=Hello&cf-turnstile-response=token
```

### Management Endpoints (Require Authentication)

- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/forms` - List all forms
- `POST /api/forms` - Create new form
- `GET /api/forms/{id}` - Get form details
- `PUT /api/forms/{id}` - Update form
- `DELETE /api/forms/{id}` - Delete form
- `GET /api/submissions` - List submissions (with optional form_id filter)

## 🧪 Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/CaffeinatedTech/staticsend.git
cd staticsend

# Install dependencies
go mod download

# Build
go build -o staticsend .

# Run tests
go test ./...

# Format code
go fmt ./...

# Run with hot reload (if using air)
air
```

### Project Structure

```
staticsend/
├── cmd/staticsend/     # Main application entry point
├── pkg/               # Go packages
│   ├── api/           # HTTP handlers and routing
│   ├── auth/          # Authentication logic
│   ├── database/      # Database operations
│   ├── email/         # Email sending service
│   ├── middleware/    # HTTP middleware
│   ├── models/        # Data models
│   ├── turnstile/     # Cloudflare Turnstile integration
│   └── utils/         # Utility functions
├── templates/         # HTML templates for HTMX UI
├── migrations/        # Database migrations
└── internal/          # Private application code
```

## 📦 Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  staticsend:
    image: ghcr.io/CaffeinatedTech/staticsend:latest
    ports:
      - "8080:8080"
    volumes:
      - staticsend_data:/data
    environment:
      - STATICSEND_PORT=8080
      - STATICSEND_JWT_SECRET=${JWT_SECRET}
      - STATICSEND_SMTP_HOST=${SMTP_HOST}
      - STATICSEND_SMTP_PORT=${SMTP_PORT}
      - STATICSEND_SMTP_USER=${SMTP_USER}
      - STATICSEND_SMTP_PASS=${SMTP_PASS}
    restart: unless-stopped

volumes:
  staticsend_data:
```

### Coolify Deployment

1. Create a new service in Coolify
2. Use the Docker image: `ghcr.io/CaffeinatedTech/staticsend:latest`
3. Configure environment variables
4. Add persistent volume for database
5. Deploy!

## 🤝 Contributing

We welcome contributions! Please feel free to submit issues, feature requests, and pull requests.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards
- Write tests for new functionality
- Update documentation for changes
- Use descriptive commit messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- 📖 [Documentation](https://github.com/CaffeinatedTech/staticsend/wiki)
- 🐛 [Issue Tracker](https://github.com/CaffeinatedTech/staticsend/issues)
- 💬 [Discussions](https://github.com/CaffeinatedTech/staticsend/discussions)

## 🙏 Acknowledgments

- [Cloudflare Turnstile](https://www.cloudflare.com/products/turnstile/) for bot protection
- [Chi Router](https://github.com/go-chi/chi) for HTTP routing
- [HTMX](https://htmx.org/) for the lightweight frontend approach
- [SQLite](https://sqlite.org/) for simple data persistence

---

**staticSend** - Making static websites more interactive, one form at a time. 🚀
