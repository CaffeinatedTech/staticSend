# Coolify Setup Guide for StaticSend

This guide walks you through deploying StaticSend on Coolify using GitHub Container Registry.

## Prerequisites

- Coolify instance running
- GitHub repository with StaticSend code
- Domain name (optional but recommended)

## Step 1: GitHub Repository Setup

1. **Push your code** to GitHub if not already done
2. **Enable GitHub Actions** (should be automatic)
3. **Verify workflow** runs successfully in Actions tab
4. **Check packages** tab for published Docker images at `ghcr.io/yourusername/staticsend`

## Step 2: Coolify Project Setup

### Create New Resource
1. In Coolify dashboard, click **"+ New Resource"**
2. Select **"Public Repository"**
3. Enter your GitHub repository URL
4. Choose **"Dockerfile"** as build pack
5. Set **branch** to `main`

### Basic Configuration
- **Name**: `staticsend`
- **Port**: `8080`
- **Health Check Path**: `/health`
- **Health Check Port**: `8080`

## Step 3: Environment Variables

Add these environment variables in Coolify:

### Security (Required)
```
JWT_SECRET_KEY=your-very-secure-random-string-change-this-immediately
```

### Email Configuration (Required for notifications)
```
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-specific-password
EMAIL_FROM=noreply@yourdomain.com
EMAIL_USE_TLS=true
```

### Turnstile (Required for bot protection)
```
TURNSTILE_PUBLIC_KEY=your-cloudflare-turnstile-public-key
TURNSTILE_SECRET_KEY=your-cloudflare-turnstile-secret-key
```

### Optional Settings
```
PORT=8080
DATABASE_PATH=/app/data/staticsend.db
REGISTRATION_ENABLED=true
```

## Step 4: Persistent Storage

1. Go to **"Storages"** tab in your Coolify resource
2. Click **"+ Add Storage"**
3. Configure:
   - **Name**: `staticsend-data`
   - **Mount Path**: `/app/data`
   - **Host Path**: `/var/lib/coolify/staticsend/data` (or your preferred path)

## Step 5: Domain Configuration (Optional)

1. Go to **"Domains"** tab
2. Click **"+ Add Domain"**
3. Enter your domain name
4. Enable **"Generate SSL Certificate"**
5. Configure DNS to point to your Coolify server

## Step 6: Deploy

1. Click **"Deploy"** button
2. Monitor deployment logs
3. Wait for health check to pass
4. Access your application at the configured domain

## Step 7: Verification

### Test Health Endpoint
```bash
curl https://yourdomain.com/health
# Should return: OK
```

### Test Application
1. Visit your domain
2. Try registering a new account
3. Create a test form
4. Submit a test form submission
5. Check email notifications

## Automatic Updates

Once configured, Coolify will automatically:
1. Watch for changes to `main` branch
2. Pull new Docker images from GitHub Container Registry
3. Deploy updates with zero downtime
4. Run health checks before switching traffic

## Troubleshooting

### Deployment Fails
- Check build logs in Coolify
- Verify Dockerfile syntax
- Ensure all required files are in repository

### Health Check Fails
- Verify port 8080 is exposed
- Check application startup logs
- Ensure `/health` endpoint is accessible

### Database Issues
- Verify persistent storage is mounted to `/app/data`
- Check file permissions on host system
- Review migration logs in application logs

### Email Not Working
- Test SMTP credentials manually
- Check firewall rules for outbound SMTP
- Verify email service configuration

## Security Notes

- Always use strong, unique JWT secret keys
- Use app-specific passwords for email accounts
- Keep Turnstile keys secure and domain-specific
- Regularly update Docker images for security patches
- Monitor application logs for suspicious activity

## Backup Strategy

### Database Backup
```bash
# On Coolify host
cp /var/lib/coolify/staticsend/data/staticsend.db /backup/location/
```

### Automated Backups
Consider setting up automated backups of the `/app/data` directory using your preferred backup solution.
