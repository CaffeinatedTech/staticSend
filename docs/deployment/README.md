# StaticSend Deployment Guide

This guide covers deploying StaticSend using Docker with Coolify and GitHub Container Registry.

## Quick Start

### 1. GitHub Container Registry Setup

The project includes automated Docker image building via GitHub Actions. Images are published to `ghcr.io/yourusername/staticsend`.

**Required GitHub Secrets:**
- No additional secrets needed - uses `GITHUB_TOKEN` automatically

### 2. Coolify Deployment

StaticSend is designed for Coolify deployment with live updates using the Dockerfile approach.

**Coolify Configuration:**
- **Source**: GitHub repository
- **Build Pack**: Dockerfile
- **Port**: 8080
- **Health Check**: `/health` endpoint (built-in)

## Environment Variables

Configure these environment variables in Coolify:

### Required Variables
```bash
# Database
DATABASE_PATH=/app/data/staticsend.db

# JWT Security (CHANGE THIS!)
JWT_SECRET_KEY=your-very-secure-jwt-secret-key-change-this

# Email Configuration
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_USE_TLS=true

# Turnstile (Cloudflare Bot Protection)
TURNSTILE_PUBLIC_KEY=your-turnstile-public-key
TURNSTILE_SECRET_KEY=your-turnstile-secret-key
```

### Optional Variables
```bash
# Server
PORT=8080

# Application Settings
REGISTRATION_ENABLED=true
```

## Docker Images

### GitHub Container Registry
Images are automatically built and pushed to:
```
ghcr.io/yourusername/staticsend:latest
ghcr.io/yourusername/staticsend:main
ghcr.io/yourusername/staticsend:v1.0.0  # for tagged releases
```

### Manual Build
```bash
# Build locally
docker build -t staticsend .

# Run locally
docker run -p 8080:8080 \
  -e JWT_SECRET_KEY=your-secret \
  -e EMAIL_HOST=smtp.gmail.com \
  -e EMAIL_USERNAME=your-email \
  -e EMAIL_PASSWORD=your-password \
  staticsend
```

## Health Checks

The application includes a built-in health check endpoint:
- **URL**: `/health`
- **Response**: `OK` (200 status)
- **Docker Health Check**: Configured automatically

## Persistent Data

StaticSend uses SQLite for data persistence. In Coolify:
1. Create a persistent volume mounted to `/app/data`
2. Database file: `/app/data/staticsend.db`
3. Automatic migrations run on startup

## Security Considerations

### JWT Secret Key
- **CRITICAL**: Change `JWT_SECRET_KEY` from default
- Use a strong, random 32+ character string
- Keep this secret secure

### Email Configuration
- Use app-specific passwords for Gmail
- Enable 2FA on email accounts
- Consider using dedicated SMTP services

### Turnstile Setup
1. Get keys from Cloudflare Turnstile
2. Add your domain to Turnstile configuration
3. Set both public and secret keys

## Monitoring

### Health Check
```bash
curl http://your-domain.com/health
# Should return: OK
```

### Logs
Monitor application logs in Coolify for:
- Database connection status
- Email service status
- Migration completion
- Authentication events

## Troubleshooting

### Common Issues

**Database Connection Failed**
- Check `/app/data` directory permissions
- Ensure persistent volume is mounted
- Verify SQLite is available in container

**Email Not Sending**
- Verify SMTP credentials
- Check email service logs
- Test with simple SMTP settings first

**Turnstile Validation Failing**
- Verify public/secret key pair
- Check domain configuration in Cloudflare
- Ensure keys match your domain

**Health Check Failing**
- Verify application is running on port 8080
- Check if `/health` endpoint is accessible
- Review application startup logs

### Debug Mode
For development, you can run without email/Turnstile:
```bash
docker run -p 8080:8080 \
  -e JWT_SECRET_KEY=debug-secret \
  -e EMAIL_HOST= \
  -e TURNSTILE_PUBLIC_KEY= \
  -e TURNSTILE_SECRET_KEY= \
  staticsend
```

## Production Checklist

- [ ] Change JWT_SECRET_KEY from default
- [ ] Configure proper SMTP settings
- [ ] Set up Turnstile keys
- [ ] Configure persistent volume for database
- [ ] Set up domain and SSL in Coolify
- [ ] Test health check endpoint
- [ ] Verify form submissions work
- [ ] Test email notifications
- [ ] Check user registration/login

## Updates and Rollbacks

### Automatic Updates (Coolify)
- Push to `main` branch triggers new build
- Coolify automatically deploys new image
- Health checks ensure successful deployment

### Manual Rollback
In Coolify dashboard:
1. Go to deployment history
2. Select previous working version
3. Click "Redeploy"

### Database Migrations
- Migrations run automatically on startup
- No manual intervention required
- Backup database before major updates
