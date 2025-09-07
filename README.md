# staticSend

A self-hosted, secure contact form processing service for static websites. Enable functional contact forms on your static sites without backend code, with built-in spam protection and email forwarding.

[![Go](https://img.shields.io/badge/Go-1.21%2B-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://docker.com)

## ✨ Features

- **🔒 Cloudflare Turnstile Integration** - Bot protection with zero user friction
- **🛡️ Authentication Bot Protection** - Optional Turnstile protection for login/register pages
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
  -e TURNSTILE_PUBLIC_KEY=your-turnstile-public-key \
  -e TURNSTILE_SECRET_KEY=your-turnstile-secret-key \
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
export TURNSTILE_PUBLIC_KEY=your-turnstile-public-key
export TURNSTILE_SECRET_KEY=your-turnstile-secret-key

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

#### Core Application
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | HTTP server port | `8080` | No |
| `DATABASE_PATH` | SQLite database path | `./data/staticsend.db` | No |
| `JWT_SECRET_KEY` | JWT signing secret | - | Yes |
| `REGISTRATION_ENABLED` | Enable user registration | `true` | No |

#### Email Configuration
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `EMAIL_HOST` | SMTP server host | - | Yes |
| `EMAIL_PORT` | SMTP server port | `587` | No |
| `EMAIL_USERNAME` | SMTP username | - | Yes |
| `EMAIL_PASSWORD` | SMTP password | - | Yes |
| `EMAIL_FROM` | From email address | - | Yes |
| `EMAIL_USE_TLS` | Use TLS for SMTP | `true` | No |

#### Turnstile Bot Protection
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `TURNSTILE_PUBLIC_KEY` | Turnstile public key for login/register pages | - | No |
| `TURNSTILE_SECRET_KEY` | Turnstile secret key for login/register pages | - | No |

#### S3 Backup Configuration (Optional)
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `S3_ENDPOINT` | S3-compatible storage endpoint | - | For backups |
| `S3_BUCKET` | S3 bucket name | - | For backups |
| `S3_ACCESS_KEY` | S3 access key | - | For backups |
| `S3_SECRET_KEY` | S3 secret key | - | For backups |
| `S3_REGION` | S3 region | `us-east-1` | No |
| `CLEANUP_OLD_BACKUPS` | Auto-delete backups older than 30 days | `true` | No |

#### Cronivore Monitoring (Optional)
| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `CRONIVORE_CHECK_SLUG` | Cronivore check slug for backup monitoring | - | No |
| `CRONIVORE_URL` | Cronivore service URL | `https://cronivore.com` | No |

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
      - staticsend_data:/app/data
    environment:
      - PORT=8080
      - DATABASE_PATH=/app/data/staticsend.db
      - JWT_SECRET_KEY=${JWT_SECRET}
      - EMAIL_HOST=${EMAIL_HOST}
      - EMAIL_PORT=${EMAIL_PORT}
      - EMAIL_USERNAME=${EMAIL_USERNAME}
      - EMAIL_PASSWORD=${EMAIL_PASSWORD}
      - EMAIL_FROM=${EMAIL_FROM}
      - TURNSTILE_PUBLIC_KEY=${TURNSTILE_PUBLIC_KEY}
      - TURNSTILE_SECRET_KEY=${TURNSTILE_SECRET_KEY}
    restart: unless-stopped

volumes:
  staticsend_data:
```

### Coolify Deployment

1. Create a new service in Coolify using GitHub repository
2. Select "Dockerfile" as build pack
3. Configure environment variables (see Configuration section above)
4. Add persistent volume: `/app/data` → `/var/lib/coolify/staticsend/data`
5. Deploy with health check on `/health` endpoint

For detailed Coolify setup instructions, see [docs/deployment/coolify-setup.md](docs/deployment/coolify-setup.md)

## 💾 Automated Backups

StaticSend includes an automated backup system that uploads database backups to S3-compatible storage.

### Features
- **SQLite database backup** using safe `.backup` command
- **S3-compatible storage** (AWS S3, DigitalOcean Spaces, Backblaze B2, etc.)
- **Automatic compression** and timestamping
- **Old backup cleanup** (configurable retention period)
- **Cronivore monitoring** integration for backup job monitoring
- **Coolify cron job** integration

### Quick Setup

1. **Configure S3 environment variables** in Coolify:
   ```bash
   S3_ENDPOINT=https://s3.amazonaws.com
   S3_BUCKET=your-backup-bucket
   S3_ACCESS_KEY=your-access-key
   S3_SECRET_KEY=your-secret-key
   ```

2. **Create Coolify cron job**:
   - Schedule: `0 2 * * *` (daily at 2 AM)
   - Command: `/app/backup.sh`

3. **Optional Cronivore monitoring**:
   ```bash
   CRONIVORE_CHECK_SLUG=your-check-slug
   ```

For complete backup setup instructions, see [docs/deployment/backup-setup.md](docs/deployment/backup-setup.md)

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
