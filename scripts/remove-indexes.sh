#!/bin/bash

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
    BACKUP_FILE="$BACKUP_DIR/pg-press_before_indexes_removal_$TIMESTAMP.db"

    cp "$DB_PATH" "$BACKUP_FILE"
    print_success "Database backed up to: $BACKUP_FILE"
}

# Function to check if indexes exist
check_indexes_status() {
    print_header "Checking Indexes Status"
    
    # Get count of indexes
    INDEX_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_%';" 2>/dev/null || echo 0)
    
    if [ "$INDEX_COUNT" -eq 0 ]; then
        print_warning "No indexes found in database"
        return 1
    fi
    
    print_info "Found $INDEX_COUNT indexes in database"
    
    # Show the indexes
    print_info "Indexes to be removed:"
    sqlite3 "$DB_PATH" "SELECT name, tbl_name FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_%';" 2>/dev/null || true
    
    return 0
}

# Function to remove all indexes
remove_indexes() {
    print_header "Removing All Indexes"
    
    # Get list of indexes to remove
    INDEXES=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_%';" 2>/dev/null || echo "")
    
    if [ -z "$INDEXES" ]; then
        print_warning "No indexes found to remove"
        return
    fi
    
    # Remove each index
    IFS=$'\n'
    for INDEX in $INDEXES; do
        print_info "Removing index: $INDEX"
        sqlite3 "$DB_PATH" "DROP INDEX IF EXISTS $INDEX;" 2>/dev/null || true
        print_success "Removed index: $INDEX"
    done
    
    # Verify indexes were removed
    FINAL_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_%';" 2>/dev/null || echo 0)
    
    if [ "$FINAL_COUNT" -eq 0 ]; then
        print_success "All indexes successfully removed"
    else
        print_warning "Some indexes may still exist. Final count: $FINAL_COUNT"
    fi
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
    print_header "Database Indexes Removal Script"
    print_info "Database: $DB_PATH"
    print_info "Backup Directory: $BACKUP_DIR"

    # Check if database exists
    if [ ! -f "$DB_PATH" ]; then
        print_error "Database file not found: $DB_PATH"
        print_info "Please ensure the database exists before running this script"
        exit 1
    fi

    # Check if indexes exist
    if check_indexes_status; then
        # Create backup before making changes
        create_backup

        # Remove indexes
        remove_indexes

        print_header "Indexes Removal Completed"
        print_success "All indexes have been removed from the database"
        print_info "Note: Removing indexes may impact query performance. Consider re-adding them if needed."
    else
        print_header "No Indexes to Remove"
        print_info "Database is already without indexes or no database found"
    fi
}

# Run the script
main "$@"