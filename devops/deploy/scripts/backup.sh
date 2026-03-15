#!/bin/bash

# Automated Database Backup Script for Sum-100 Game
# Place this at /opt/100sumgame/backend/scripts/backup.sh on the server
# Make executable: chmod +x backup.sh

set -e  # Exit on error

# Configuration
DB_CONTAINER_NAME="sum100-db-prod"
DB_USER="postgres"
DB_NAME="sum100game"
BACKUP_DIR="/opt/100sumgame/backups"
RETENTION_DAYS=7

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/backup_${DATE}.sql.gz"

echo "========================================="
echo "🗄️  Database Backup Started"
echo "========================================="
echo "Time: $(date)"
echo "Backup file: $BACKUP_FILE"
echo ""

# Perform backup
echo "📦 Creating backup..."
docker exec "$DB_CONTAINER_NAME" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"

# Verify backup was created
if [ -f "$BACKUP_FILE" ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "✅ Backup created successfully!"
    echo "Size: $BACKUP_SIZE"
else
    echo "❌ Backup failed!"
    exit 1
fi

# Clean up old backups
echo ""
echo "🧹 Cleaning up old backups (older than $RETENTION_DAYS days)..."
find "$BACKUP_DIR" -name "backup_*.sql.gz" -mtime +$RETENTION_DAYS -delete

# List remaining backups
echo ""
echo "📋 Current backups:"
ls -lh "$BACKUP_DIR" | grep "backup_" | awk '{print "  - " $9 " (" $5 ")"}'

echo ""
echo "========================================="
echo "✅ Backup Completed Successfully"
echo "========================================="
echo "Total backups: $(ls -1 "$BACKUP_DIR"/backup_*.sql.gz 2>/dev/null | wc -l)"
echo "Disk usage: $(du -sh "$BACKUP_DIR" | cut -f1)"
echo ""

# Log to system log (optional)
logger "Sum-100 Game Database Backup: $BACKUP_FILE ($BACKUP_SIZE)"