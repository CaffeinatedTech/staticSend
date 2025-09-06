# Configuration Guide

## Environment Variables

### Server Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `STATICSEND_PORT` | HTTP server port | `8080` | No |
| `STATICSEND_DB_PATH` | SQLite database path | `./staticsend.db` | No |
| `STATICSEND_JWT_SECRET` | JWT signing secret | - | Yes |

### Email Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `STATICSEND_SMTP_HOST` | SMTP server host | - | Yes |
| `STATICSEND_SMTP_PORT` | SMTP server port | - | Yes |
| `STATICSEND_SMTP_USER` | SMTP username | - | Yes |
| `STATICSEND_SMTP_PASS` | SMTP password | - | Yes |
| `STATICSEND_SMTP_FROM` | From email address | - | Yes |
| `STATICSEND_SMTP_USE_TLS` | Use TLS for SMTP | `true` | No |

### Turnstile Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `STATICSEND_TURNSTILE_SECRET` | Cloudflare Turnstile secret key | - | Yes |
| `STATICSEND_TURNSTILE_VERIFY_URL` | Turnstile verify URL | `https://challenges.cloudflare.com/turnstile/v0/siteverify` | No |

### Rate Limiting Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `STATICSEND_RATE_LIMIT_RATE` | Rate limit duration | `1s` | No |
| `STATICSEND_RATE_LIMIT_BURST` | Rate limit burst capacity | `5` | No |

## Command Line Flags

| Flag | Description | Default | Environment Variable |
|------|-------------|---------|---------------------|
| `-port` | HTTP server port | `8080` | `STATICSEND_PORT` |
| `-help` | Show help information | `false` | - |

## Example Configuration

### Environment File (.env)

```bash
# Server
STATICSEND_PORT=8080
STATICSEND_DB_PATH=./staticsend.db
STATICSEND_JWT_SECRET=your-super-secret-jwt-key

# Email
STATICSEND_SMTP_HOST=smtp.gmail.com
STATICSEND_SMTP_PORT=587
STATICSEND_SMTP_USER=your-email@gmail.com
STATICSEND_SMTP_PASS=your-app-password
STATICSEND_SMTP_FROM=noreply@yourdomain.com
STATICSEND_SMTP_USE_TLS=true

# Turnstile
STATICSEND_TURNSTILE_SECRET=your-turnstile-secret-key

# Rate Limiting
STATICSEND_RATE_LIMIT_RATE=1s
STATICSEND_RATE_LIMIT_BURST=5
```

### Docker Compose

```yaml
version: '3.8'

services:
  staticsend:
    image: your-username/staticsend:latest
    ports:
      - "8080:8080"
    volumes:
      - staticsend_data:/data
    environment:
      - STATICSEND_PORT=8080
      - STATICSEND_DB_PATH=/data/staticsend.db
      - STATICSEND_JWT_SECRET=${JWT_SECRET}
      - STATICSEND_SMTP_HOST=${SMTP_HOST}
      - STATICSEND_SMTP_PORT=${SMTP_PORT}
      - STATICSEND_SMTP_USER=${SMTP_USER}
      - STATICSEND_SMTP_PASS=${SMTP_PASS}
      - STATICSEND_TURNSTILE_SECRET=${TURNSTILE_SECRET}
    restart: unless-stopped

volumes:
  staticsend_data:
```

## Security Considerations

- Always use strong JWT secrets
- Use TLS for SMTP connections
- Keep Turnstile secrets secure
- Regularly rotate credentials
- Use environment variables instead of hardcoded values