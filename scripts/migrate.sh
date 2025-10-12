#!/bin/bash

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

show_usage() {
    echo "Quick Migration Helper"
    echo
    echo "Usage: $0 [database-path]"
    echo
    echo "Examples:"
    echo "  $0                    # Migrate default database (./pg-press.db)"
    echo "  $0 /path/to/db.db     # Migrate specific database"
    echo
    echo "Advanced options:"
    echo "  --help               # Show detailed help"
}

# Parse arguments
case "$1" in
    --help|-h)
        show_usage
        exit 0
        ;;
esac

# Determine database path
if [ -n "$1" ]; then
    DB_PATH="$1"
else
    DB_PATH="${DB_PATH:-./pg-press.db}"
fi

print_info "Starting migration for database: $DB_PATH"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Database not found: $DB_PATH"
    echo
    echo "Options:"
    echo "1. Create database first by running the application"
    echo "2. Specify correct database path: $0 /path/to/pg-press.db"
    exit 1
fi

# Run the migration
print_info "Running migration script..."
./scripts/migrate-tools-binding.sh -d "$DB_PATH"

print_success "Migration completed!"
