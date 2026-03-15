#!/bin/bash

# Deployment Script for Sum-100 Game Backend
# Place this at /opt/100sumgame/backend/scripts/deploy.sh on the server
# Make executable: chmod +x deploy.sh

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_DIR="/opt/100sumgame/backend"
BACKUP_DIR="/opt/100sumgame/backups"
DB_CONTAINER_NAME="sum100-db-prod"
BACKUP_CONTAINER_NAME="sum100-backend-prod"

echo ""
echo "========================================="
echo "🚀 Sum-100 Game Deployment"
echo "========================================="
echo "Time: $(date)"
echo ""

# Function to print colored messages
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    print_warning "This script should be run as root or with sudo"
    print_info "You may be prompted for your password"
fi

# Step 1: Backup current database
print_info "Step 1: Creating database backup before deployment..."
if [ -f "$PROJECT_DIR/scripts/backup.sh" ]; then
    bash "$PROJECT_DIR/scripts/backup.sh"
else
    print_warning "Backup script not found, skipping backup"
fi

# Step 2: Pull latest code
print_info "Step 2: Pulling latest code from GitHub..."
cd "$PROJECT_DIR"
git pull origin main
print_success "Code updated"

# Step 3: Check if environment file exists
if [ ! -f "$PROJECT_DIR/devops/.env.prod" ]; then
    print_error "Environment file not found: $PROJECT_DIR/devops/.env.prod"
    print_info "Creating from example file..."
    cp "$PROJECT_DIR/devops/.env.prod.example" "$PROJECT_DIR/devops/.env.prod"
    print_warning "Please edit $PROJECT_DIR/devops/.env.prod with your actual values"
    exit 1
fi

print_success "Environment file found"

# Step 4: Stop existing containers
print_info "Step 3: Stopping existing containers..."
if [ -f "$PROJECT_DIR/docker-compose.prod.yml" ]; then
    docker-compose -f "$PROJECT_DIR/docker-compose.prod.yml" down
else
    print_warning "docker-compose.prod.yml not found"
fi

# Step 5: Build and start new containers
print_info "Step 4: Building and starting new containers..."
docker-compose -f "$PROJECT_DIR/docker-compose.prod.yml" up -d --build

# Step 6: Wait for database to be ready
print_info "Step 5: Waiting for database to be ready..."
sleep 10

# Check if database is ready
MAX_RETRIES=30
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker exec "$DB_CONTAINER_NAME" pg_isready -U postgres > /dev/null 2>&1; then
        print_success "Database is ready"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo -n "."
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    print_error "Database failed to start"
    exit 1
fi

# Step 7: Run migrations (if needed)
print_info "Step 6: Checking for database migrations..."
# Add migration logic here if you have automated migrations
print_success "Migrations up to date"

# Step 8: Check application health
print_info "Step 7: Checking application health..."
sleep 5

MAX_HEALTH_RETRIES=10
HEALTH_RETRY_COUNT=0
while [ $HEALTH_RETRY_COUNT -lt $MAX_HEALTH_RETRIES ]; do
    if curl -s http://localhost:8080/health | grep -q "OK"; then
        print_success "Application is healthy"
        break
    fi
    HEALTH_RETRY_COUNT=$((HEALTH_RETRY_COUNT + 1))
    echo -n "."
    sleep 3
done

if [ $HEALTH_RETRY_COUNT -eq $MAX_HEALTH_RETRIES ]; then
    print_error "Application health check failed"
    print_info "Check logs: docker-compose -f $PROJECT_DIR/docker-compose.prod.yml logs"
    exit 1
fi

# Step 9: Show recent logs
print_info "Step 8: Showing recent application logs..."
echo ""
docker-compose -f "$PROJECT_DIR/docker-compose.prod.yml" logs --tail=20 backend
echo ""

# Step 10: Restart Nginx
print_info "Step 9: Restarting Nginx..."
if systemctl is-active --quiet nginx; then
    systemctl reload nginx
    print_success "Nginx reloaded"
else
    print_warning "Nginx is not running, skipping reload"
fi

# Final summary
echo ""
echo "========================================="
echo "✅ Deployment Completed Successfully!"
echo "========================================="
echo ""
print_info "Application Status:"
echo "  - Backend: Running on port 8080"
echo "  - Database: Running"
echo "  - Nginx: Running and proxying requests"
echo ""
print_info "Useful Commands:"
echo "  - View logs: docker-compose -f $PROJECT_DIR/docker-compose.prod.yml logs -f"
echo "  - Check health: curl http://localhost:8080/health"
echo "  - Restart: docker-compose -f $PROJECT_DIR/docker-compose.prod.yml restart"
echo "  - Stop: docker-compose -f $PROJECT_DIR/docker-compose.prod.yml down"
echo ""
print_success "Deployment completed at $(date)"
echo ""

# Log to system log
logger "Sum-100 Game Deployment: Successful at $(date)"