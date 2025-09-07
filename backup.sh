#!/bin/bash
set -euo pipefail

# StaticSend Database Backup Script
# Backs up SQLite database to S3-compatible storage with optional Cronivore monitoring

# --- Utility Functions ---

# URL-encodes a string for use in a query parameter.
rawurlencode() {
  local string="${1}"
  local encoded=""
  local c

  while [ -n "$string" ]; do
    c=${string%"${string#?}"}
    string=${string#?}
    case "$c" in
      [-_.~a-zA-Z0-9] )
        encoded="${encoded}${c}"
        ;;
      *)
        encoded="${encoded}$(printf '%%%02x' "'$c")"
        ;;
    esac
  done
  echo "$encoded"
}

# --- Cronivore Monitoring ---

# Sends a ping to a Cronivore check. Pings are only sent if CRONIVORE_CHECK_SLUG is set.
# usage: ping_cronivore <endpoint> [reason]
# e.g., ping_cronivore "start"
# e.g., ping_cronivore "fail" "Something went wrong"
ping_cronivore() {
  local endpoint=$1
  local reason_param=""
  if [[ -n "${2:-}" ]]; then
    encoded_reason=$(rawurlencode "$2")
    reason_param="?reason=${encoded_reason}"
  fi

  if [[ -n "${CRONIVORE_CHECK_SLUG:-}" ]]; then
    local url="${CRONIVORE_URL:-https://cronivore.com}/check-in/${CRONIVORE_CHECK_SLUG}/${endpoint}${reason_param}"
    echo "Pinging Cronivore: ${url}"
    # Use curl to send the request. The -fsS options make it fail silently on errors but show them if they occur.
    # --retry 3 makes it more robust against transient network issues.
    curl -fsS --retry 3 "${url}" > /dev/null || echo "Warning: Failed to ping Cronivore."
  fi
}

# Trap to catch errors and send a failure ping.
# This executes when the script exits with a non-zero status code (on any error).
trap 'ping_cronivore "fail" "StaticSend backup script failed on or near line $LINENO."' ERR

# Ping start
ping_cronivore "start"

# --- Configuration Validation ---
if [[ -z "${S3_ENDPOINT:-}" || -z "${S3_BUCKET:-}" || -z "${S3_ACCESS_KEY:-}" || -z "${S3_SECRET_KEY:-}" ]]; then
  echo "Error: Missing required S3 environment variables."
  echo "Please set: S3_ENDPOINT, S3_BUCKET, S3_ACCESS_KEY, S3_SECRET_KEY"
  exit 1
fi

# Set environment variables for aws-cli.
# It requires the AWS_ prefix, so we map our generic vars to the ones it expects.
export AWS_ACCESS_KEY_ID=${S3_ACCESS_KEY}
export AWS_SECRET_ACCESS_KEY=${S3_SECRET_KEY}
export AWS_DEFAULT_REGION=${S3_REGION:-"us-east-1"} # aws-cli requires a region, even if the service (like StorJ) ignores it.

# --- StaticSend Database Configuration ---
# Default database path - can be overridden with DATABASE_PATH environment variable
DATABASE_PATH=${DATABASE_PATH:-"/app/data/staticsend.db"}

echo "StaticSend backup starting..."
echo "Database path: ${DATABASE_PATH}"

# Create a temporary directory to stage the backup files.
BACKUP_DIR=$(mktemp -d)

# Setup a trap to automatically clean up the temporary directory on script exit (success or failure).
trap 'rm -rf -- "$BACKUP_DIR"' EXIT

echo "Staging database backups in ${BACKUP_DIR}..."

# Check if database file exists
if [[ ! -f "$DATABASE_PATH" ]]; then
  echo "Error: Database file not found at ${DATABASE_PATH}"
  echo "Please check the DATABASE_PATH environment variable or ensure the database exists."
  exit 1
fi

# Get database file info
DB_SIZE=$(du -h "$DATABASE_PATH" | cut -f1)
echo "Database size: ${DB_SIZE}"

# Create a safe backup of the StaticSend database
BACKUP_FILENAME="staticsend-$(date +%Y-%m-%d-%H%M%S).db"
echo "Creating backup: ${BACKUP_FILENAME}..."
sqlite3 "$DATABASE_PATH" ".backup '${BACKUP_DIR}/${BACKUP_FILENAME}'"

# Also backup any additional files in the data directory (like logs, etc.)
DATA_DIR=$(dirname "$DATABASE_PATH")
if [[ -d "$DATA_DIR" ]]; then
  echo "Backing up additional data directory contents..."
  # Copy any other files in the data directory (excluding the main db which we already backed up)
  find "$DATA_DIR" -type f ! -name "$(basename "$DATABASE_PATH")" -exec cp {} "$BACKUP_DIR/" \; 2>/dev/null || true
fi

# Create a compressed tarball of the backups.
ARCHIVE_NAME="staticsend-backup-$(date +%Y-%m-%d-%H%M%S).tar.gz"
ARCHIVE_PATH="/tmp/${ARCHIVE_NAME}" # Create archive outside of the backup dir.
trap 'rm -rf -- "$BACKUP_DIR" "${ARCHIVE_PATH}"' EXIT # Add archive to cleanup trap.

echo "Creating compressed archive: ${ARCHIVE_NAME}..."
tar -czf "${ARCHIVE_PATH}" -C "${BACKUP_DIR}" .

# Get archive size for reporting
ARCHIVE_SIZE=$(du -h "$ARCHIVE_PATH" | cut -f1)
echo "Archive size: ${ARCHIVE_SIZE}"

# --- Upload ---
echo "Uploading ${ARCHIVE_NAME} to S3 bucket '${S3_BUCKET}'..."
# Use aws-cli to upload the archive to the S3-compatible storage.
# Configure AWS CLI to handle Content-Length properly
export AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY}"
export AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY}"
export AWS_DEFAULT_REGION="${S3_REGION:-us-east-1}"

# Configure AWS CLI to disable new integrity checks that cause Content-Length issues
# This fixes the breaking change introduced in AWS CLI v2.17+ and boto3 v1.36.0+
export AWS_CONFIG_FILE="/tmp/aws_config"
cat > "$AWS_CONFIG_FILE" << EOF
[default]
request_checksum_calculation = when_required
response_checksum_validation = when_required
EOF

# Use standard s3 cp command which works reliably
aws s3 cp \
  "${ARCHIVE_PATH}" \
  "s3://${S3_BUCKET}/${ARCHIVE_NAME}" \
  --endpoint-url "${S3_ENDPOINT}"

echo "Backup upload complete: ${ARCHIVE_NAME} (${ARCHIVE_SIZE})"

# Optional: Clean up old backups (keep last 30 days)
if [[ "${CLEANUP_OLD_BACKUPS:-true}" == "true" ]]; then
  echo "Cleaning up old backups (keeping last 30 days)..."
  # Alpine Linux uses busybox date which doesn't support -d flag
  # Calculate cutoff date using arithmetic (30 days = 30 * 24 * 60 * 60 = 2592000 seconds)
  CUTOFF_TIMESTAMP=$(($(date +%s) - 2592000))
  CUTOFF_DATE=$(date -d "@${CUTOFF_TIMESTAMP}" +%Y-%m-%d 2>/dev/null || date -r "${CUTOFF_TIMESTAMP}" +%Y-%m-%d)
  
  aws s3 ls "s3://${S3_BUCKET}/" --endpoint-url "${S3_ENDPOINT}" | \
    grep "staticsend-backup-" | \
    awk '{print $4}' | \
    while read -r backup_file; do
      backup_date=$(echo "$backup_file" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}')
      if [[ "$backup_date" < "$CUTOFF_DATE" ]]; then
        echo "Deleting old backup: $backup_file"
        aws s3 rm "s3://${S3_BUCKET}/$backup_file" --endpoint-url "${S3_ENDPOINT}" || echo "Warning: Failed to delete $backup_file"
      fi
    done
fi

# Ping success now that the upload is finished.
ping_cronivore "success"

echo "StaticSend backup script finished successfully."
