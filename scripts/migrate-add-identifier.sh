#!/bin/bash

# Migration script to add identifier column to metal_sheets table
# This script adds the identifier field to detect if stf_max is from type "SACMI" or "SITI"

set -e

# Configuration
DB_PATH="${DB_PATH:-./pg-press.db}"
BACKUP_DIR="./backups"

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

# Check if database exists
check_database() {
    if [ ! -f "$DB_PATH" ]; then
        print_error "Database file not found: $DB_PATH"
        print_warning "Make sure you're running this script from the project root directory"
        exit 1
    fi
    print_success "Database file found: $DB_PATH"
}

# Create backup directory if it doesn't exist
ensure_backup_dir() {
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        print_success "Created backup directory: $BACKUP_DIR"
    fi
}

# Create backup of database
backup_database() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="${BACKUP_DIR}/pg-press_before_identifier_migration_${timestamp}.db"

    print_header "Creating database backup"
    cp "$DB_PATH" "$backup_file"
    print_success "Database backed up to: $backup_file"
}

# Check if identifier column already exists
check_column_exists() {
    local column_exists=$(sqlite3 "$DB_PATH" "PRAGMA table_info(metal_sheets);" | grep -c "identifier" || true)

    if [ "$column_exists" -gt 0 ]; then
        print_warning "The 'identifier' column already exists in metal_sheets table"
        echo "Migration may have already been applied."
        read -p "Do you want to continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_warning "Migration cancelled by user"
            exit 0
        fi
    fi
}

# Add identifier column to metal_sheets table
add_identifier_column() {
    print_header "Adding identifier column to metal_sheets table"

    # Add the column with default value
    sqlite3 "$DB_PATH" "ALTER TABLE metal_sheets ADD COLUMN identifier TEXT NOT NULL DEFAULT 'SACMI';"

    if [ $? -eq 0 ]; then
        print_success "Successfully added identifier column"
    else
        print_error "Failed to add identifier column"
        exit 1
    fi
}

# Verify the migration
verify_migration() {
    print_header "Verifying migration"

    # Check if column exists and get its info
    local column_info=$(sqlite3 "$DB_PATH" "PRAGMA table_info(metal_sheets);" | grep identifier || true)

    if [ -n "$column_info" ]; then
        print_success "Identifier column successfully added"
        echo "Column info: $column_info"

        # Count total rows
        local total_rows=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM metal_sheets;")
        print_success "Total metal sheets in database: $total_rows"

        # Count rows with default SACMI identifier
        local sacmi_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM metal_sheets WHERE identifier = 'SACMI';")
        print_success "Metal sheets with SACMI identifier: $sacmi_count"

        if [ "$total_rows" -eq "$sacmi_count" ]; then
            print_success "All existing records have been set to SACMI (default)"
        else
            print_warning "Some records may have different identifier values"
        fi
    else
        print_error "Migration verification failed - identifier column not found"
        exit 1
    fi
}

# Show usage information
show_usage() {
    print_header "Post-Migration Information"
    echo "The identifier field has been added to track machine types:"
    echo "• SACMI - For SACMI type machines"
    echo "• SITI - For SITI type machines"
    echo ""
    echo "All existing metal sheets have been set to 'SACMI' by default."
    echo "You can update individual records as needed through the application."
    echo ""
    echo "To manually update a record:"
    echo "sqlite3 $DB_PATH \"UPDATE metal_sheets SET identifier='SITI' WHERE id=<sheet_id>;\""
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║              Metal Sheet Identifier Migration               ║"
    echo "║                                                              ║"
    echo "║ This script adds the 'identifier' field to the metal_sheets ║"
    echo "║ table to distinguish between SACMI and SITI machine types.  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Check if running from correct directory
    if [ ! -f "go.mod" ] || [ ! -d "scripts" ]; then
        print_error "Please run this script from the project root directory"
        exit 1
    fi

    # Change to scripts directory for relative paths
    cd scripts

    # Execute migration steps
    check_database
    ensure_backup_dir
    backup_database
    check_column_exists
    add_identifier_column
    verify_migration
    show_usage

    print_header "Migration Completed Successfully"
    print_success "The identifier field has been added to the metal_sheets table"
    print_warning "Remember to restart your application to pick up the database changes"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "This script adds the 'identifier' column to the metal_sheets table."
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo ""
        echo "The script will:"
        echo "1. Create a backup of your database"
        echo "2. Add the identifier column with default value 'SACMI'"
        echo "3. Verify the migration was successful"
        echo ""
        echo "Run this script from the project root directory."
        exit 0
        ;;
    --version|-v)
        echo "Metal Sheet Identifier Migration v1.0"
        exit 0
        ;;
    *)
        # No arguments or unknown arguments, proceed with migration
        main "$@"
        ;;
esac
