#!/bin/bash
# TODO: Added "binding", need migration script for this

set -e

# Configuration
DB_PATH="${DB_PATH:-./pg-press.db}"
BACKUP_DIR="./scripts/backups"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions for colored output
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Function to create backup
create_backup() {
    print_header "Creating Database Backup"

    if [ ! -f "$DB_PATH" ]; then
        print_error "Database file not found: $DB_PATH"
        exit 1
    fi

    # Create backup directory if it doesn't exist
    mkdir -p "$BACKUP_DIR"

    # Create backup with timestamp
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    BACKUP_FILE="$BACKUP_DIR/pg-press_before_tools_binding_migration_$TIMESTAMP.db"

    cp "$DB_PATH" "$BACKUP_FILE"
    print_success "Database backed up to: $BACKUP_FILE"
}

# Function to check if migration is needed
check_migration_status() {
    print_header "Checking Migration Status"

    # Check if binding column already exists
    COLUMN_EXISTS=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "binding" || true)

    if [ "$COLUMN_EXISTS" -gt 0 ]; then
        print_warning "Migration already applied: binding column exists in tools table"
        print_info "Checking current schema..."
        sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep "binding"
        return 1
    fi

    print_success "Migration needed: binding column not found in tools table"
    return 0
}

# Function to run the migration
run_migration() {
    print_header "Running Migration"

    print_info "Adding binding column to tools table..."

    # Add the binding column with default value 0 (false/alive)
    sqlite3 "$DB_PATH" "ALTER TABLE tools ADD COLUMN binding INTEGER DEFAULT NULL;"

    print_success "Successfully added binding column to tools table"

    # Verify the column was added correctly
    print_info "Verifying migration..."
    COLUMN_COUNT=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "binding" || true)

    if [ "$COLUMN_COUNT" -eq 1 ]; then
        print_success "Migration verification passed"
    else
        print_error "Migration verification failed"
        exit 1
    fi

    # Show updated schema
    print_info "Updated tools table schema:"
    sqlite3 "$DB_PATH" "PRAGMA table_info(tools);"
}

# Function to show usage information
show_usage() {
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -d, --database PATH    Specify database path (default: ./pg-press.db)"
    echo "  -h, --help            Show this help message"
    echo
    echo "Environment Variables:"
    echo "  DB_PATH               Database path (default: ./pg-press.db)"
    echo
    echo "Examples:"
    echo "  $0                                    # Use default database path"
    echo "  $0 -d /path/to/custom.db             # Use custom database path"
    echo "  DB_PATH=/custom/path.db $0           # Use environment variable"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--database)
            DB_PATH="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_header "Tools Binding Migration Script"
    print_info "Database: $DB_PATH"
    print_info "Backup Directory: $BACKUP_DIR"

    # Check if database exists
    if [ ! -f "$DB_PATH" ]; then
        print_error "Database file not found: $DB_PATH"
        print_info "Please ensure the database exists before running this migration"
        exit 1
    fi

    # Check if migration is needed
    if check_migration_status; then
        # Create backup before making changes
        create_backup

        # Run the migration
        run_migration

        print_header "Migration Completed Successfully"
        print_success "Tools table now supports binding"

    else
        print_header "Migration Skipped"
        print_info "No migration needed - database is already up to date"
    fi
}

# Run the script
main "$@"
