# StaticSend Backup Setup Guide

This guide covers setting up automated backups for StaticSend using S3-compatible storage and Coolify cron jobs.

## Overview

The backup script (`backup.sh`) creates compressed backups of your StaticSend database and uploads them to S3-compatible storage. It includes:

- SQLite database backup using safe `.backup` command
- Automatic compression and timestamping
- S3-compatible storage upload (AWS S3, DigitalOcean Spaces, etc.)
- Optional Cronivore monitoring integration
- Automatic cleanup of old backups (30+ days)

## Prerequisites

- S3-compatible storage bucket
- AWS CLI installed in your container/server
- Coolify cron job capability

## Environment Variables

Add these environment variables to your Coolify deployment:

### Required S3 Configuration
```bash
S3_ENDPOINT=https://s3.amazonaws.com  # or your S3-compatible endpoint
S3_BUCKET=your-backup-bucket
S3_ACCESS_KEY=your-s3-access-key
S3_SECRET_KEY=your-s3-secret-key
S3_REGION=us-east-1
```

### Optional Configuration
```bash
DATABASE_PATH=/app/data/staticsend.db  # Default path
CLEANUP_OLD_BACKUPS=true              # Auto-delete backups older than 30 days
```

### Optional Cronivore Monitoring
```bash
CRONIVORE_CHECK_SLUG=your-check-slug  # Get from cronivore.com
CRONIVORE_URL=https://cronivore.com   # Default URL
```

## S3 Storage Providers

### AWS S3
```bash
S3_ENDPOINT=https://s3.amazonaws.com
S3_REGION=us-east-1  # Your actual region
```

### DigitalOcean Spaces
```bash
S3_ENDPOINT=https://nyc3.digitaloceanspaces.com  # Your region
S3_REGION=us-east-1  # Required by AWS CLI
```

### Backblaze B2
```bash
S3_ENDPOINT=https://s3.us-west-002.backblazeb2.com  # Your region
S3_REGION=us-west-002
```

### Wasabi
```bash
S3_ENDPOINT=https://s3.wasabisys.com
S3_REGION=us-east-1
```

## Coolify Cron Job Setup

1. **Go to your StaticSend resource in Coolify**
2. **Navigate to "Scheduled Tasks" or "Cron Jobs"**
3. **Create a new cron job with:**

### Daily Backup (Recommended)
```bash
# Schedule: 0 2 * * * (2 AM daily)
# Command:
/app/backup.sh
```

### Weekly Backup
```bash
# Schedule: 0 2 * * 0 (2 AM every Sunday)
# Command:
/app/backup.sh
```

## Docker Image Requirements

The backup script requires these tools in your container:
- `sqlite3` - For database backup
- `aws` CLI - For S3 upload
- `curl` - For Cronivore monitoring
- Standard Unix tools (`tar`, `gzip`, `find`, etc.)

These are already included in the StaticSend Docker image.

## Backup File Format

Backups are stored with this naming convention:
```
staticsend-backup-YYYY-MM-DD-HHMMSS.tar.gz
```

Example: `staticsend-backup-2025-09-07-020000.tar.gz`

## Monitoring with Cronivore

[Cronivore](https://cronivore.com) provides monitoring for your backup jobs:

1. **Create account** at cronivore.com
2. **Create a new check** for your backup job
3. **Copy the check slug** from the URL
4. **Set environment variables:**
   ```bash
   CRONIVORE_CHECK_SLUG=your-check-slug-here
   ```

The script will automatically:
- Ping "start" when backup begins
- Ping "success" when backup completes
- Ping "fail" if any error occurs

## Testing the Backup

Test your backup configuration:

```bash
# SSH into your Coolify server or container
docker exec -it your-staticsend-container /app/backup.sh
```

Expected output:
```
StaticSend backup starting...
Database path: /app/data/staticsend.db
Database size: 2.1M
Creating backup: staticsend-2025-09-07-140000.db
Creating compressed archive: staticsend-backup-2025-09-07-140000.tar.gz
Archive size: 512K
Uploading staticsend-backup-2025-09-07-140000.tar.gz to S3 bucket 'your-bucket'...
Backup upload complete: staticsend-backup-2025-09-07-140000.tar.gz (512K)
StaticSend backup script finished successfully.
```

## Restoring from Backup

To restore from a backup:

1. **Download backup from S3:**
   ```bash
   aws s3 cp s3://your-bucket/staticsend-backup-YYYY-MM-DD-HHMMSS.tar.gz . \
     --endpoint-url https://your-s3-endpoint
   ```

2. **Extract the backup:**
   ```bash
   tar -xzf staticsend-backup-YYYY-MM-DD-HHMMSS.tar.gz
   ```

3. **Stop StaticSend application**

4. **Replace database file:**
   ```bash
   cp staticsend-YYYY-MM-DD-HHMMSS.db /app/data/staticsend.db
   ```

5. **Restart StaticSend application**

## Troubleshooting

### Common Issues

**"Missing required S3 environment variables"**
- Verify all S3_* environment variables are set in Coolify
- Check for typos in variable names

**"Database file not found"**
- Verify DATABASE_PATH points to correct location
- Ensure persistent storage is properly mounted

**"AWS CLI command not found"**
- Ensure you're using the official StaticSend Docker image
- AWS CLI is pre-installed in the container

**"Permission denied"**
- Check that backup.sh is executable: `chmod +x backup.sh`
- Verify container has write access to temporary directories

**S3 Upload Fails**
- Test S3 credentials manually with AWS CLI
- Verify S3_ENDPOINT URL is correct
- Check bucket permissions and policies

### Debug Mode

For detailed debugging, run the script manually:
```bash
set -x  # Enable debug mode
/app/backup.sh
```

## Security Considerations

- **S3 Credentials**: Use IAM users with minimal required permissions
- **Bucket Policies**: Restrict access to backup bucket
- **Encryption**: Enable S3 bucket encryption
- **Access Logs**: Monitor S3 access logs for unauthorized access
- **Rotation**: Regularly rotate S3 access keys

## Backup Retention

The script automatically cleans up backups older than 30 days. To modify retention:

```bash
# Keep backups for 90 days (modify the script)
CUTOFF_DATE=$(date -d '90 days ago' +%Y-%m-%d)

# Disable cleanup entirely
CLEANUP_OLD_BACKUPS=false
```

## Cost Optimization

- Use S3 Intelligent Tiering for automatic cost optimization
- Consider using cheaper storage classes for older backups
- Monitor storage costs and adjust retention as needed
- Use compression (already enabled in the script)
