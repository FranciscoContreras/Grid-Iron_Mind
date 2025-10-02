#!/bin/bash

# Database Backup Script
# Creates compressed backups with 30-day retention

set -e

# Configuration
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="backups"
RETENTION_DAYS=30

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🗄️  Grid Iron Mind - Database Backup"
echo "===================================="
echo ""

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}❌ ERROR${NC}: DATABASE_URL environment variable not set"
    echo "Usage: DATABASE_URL=postgres://... ./scripts/backup-database.sh"
    exit 1
fi

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup filename
BACKUP_FILE="$BACKUP_DIR/gridironmind_$DATE.sql"
COMPRESSED_FILE="$BACKUP_FILE.gz"

echo "📦 Creating backup..."
echo "   Database: $DATABASE_URL"
echo "   File: $BACKUP_FILE"
echo ""

# Create backup
if pg_dump "$DATABASE_URL" > "$BACKUP_FILE" 2>/dev/null; then
    echo -e "${GREEN}✅ Backup created${NC}: $BACKUP_FILE"
else
    echo -e "${RED}❌ Backup failed${NC}"
    rm -f "$BACKUP_FILE"
    exit 1
fi

# Get backup size
BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
echo "   Size: $BACKUP_SIZE"

# Compress backup
echo ""
echo "🗜️  Compressing backup..."
if gzip "$BACKUP_FILE"; then
    COMPRESSED_SIZE=$(du -h "$COMPRESSED_FILE" | cut -f1)
    echo -e "${GREEN}✅ Compressed${NC}: $COMPRESSED_FILE"
    echo "   Size: $COMPRESSED_SIZE"
else
    echo -e "${RED}❌ Compression failed${NC}"
    exit 1
fi

# Clean up old backups
echo ""
echo "🧹 Cleaning old backups (>$RETENTION_DAYS days)..."
DELETED_COUNT=$(find "$BACKUP_DIR" -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete -print | wc -l | tr -d ' ')

if [ "$DELETED_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}🗑️  Deleted${NC}: $DELETED_COUNT old backup(s)"
else
    echo "   No old backups to delete"
fi

# List recent backups
echo ""
echo "📋 Recent backups:"
ls -lh "$BACKUP_DIR"/*.sql.gz 2>/dev/null | tail -5 | awk '{print "   "$9" - "$5}' || echo "   No backups found"

# Summary
echo ""
echo "===================================="
echo -e "${GREEN}✅ Backup complete!${NC}"
echo ""
echo "To restore this backup:"
echo "  gunzip -c $COMPRESSED_FILE | psql \$DATABASE_URL"
echo ""
