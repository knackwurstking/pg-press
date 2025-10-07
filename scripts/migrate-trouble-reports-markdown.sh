#!/bin/bash

# Migration script for Trouble Reports Markdown Feature
# This script adds the use_markdown column to the trouble_reports table

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

# Create backup of current database
create_backup() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="${BACKUP_DIR}/pg-press_before_markdown_migration_${timestamp}.db"

    cp "$DB_PATH" "$backup_file"
    print_success "Created backup: $backup_file"
    echo "Backup location: $backup_file"
}

# Check if migration is needed
check_migration_needed() {
    print_header "Checking if migration is needed"

    # Check if use_markdown column already exists
    local column_exists=$(sqlite3 "$DB_PATH" "PRAGMA table_info(trouble_reports);" | grep "use_markdown" || echo "")

    if [ -n "$column_exists" ]; then
        print_warning "Migration appears to already be complete"
        print_info "The use_markdown column already exists in the trouble_reports table"

        # Show current schema
        echo ""
        print_info "Current trouble_reports schema:"
        sqlite3 "$DB_PATH" "PRAGMA table_info(trouble_reports);"

        echo ""
        read -p "Do you want to continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_warning "Migration cancelled"
            exit 0
        fi
        return 1
    fi

    print_success "Migration is needed - use_markdown column does not exist"
    return 0
}

# Show current database schema
show_current_schema() {
    print_header "Current Database Schema"

    print_info "Current trouble_reports table structure:"
    sqlite3 "$DB_PATH" "PRAGMA table_info(trouble_reports);"

    print_info "Current trouble_reports record count:"
    local count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM trouble_reports;")
    echo "Total records: $count"
}

# Perform the migration
perform_migration() {
    print_header "Performing Migration"

    print_info "Adding use_markdown column to trouble_reports table..."

    # Add the use_markdown column with default value FALSE
    sqlite3 "$DB_PATH" "ALTER TABLE trouble_reports ADD COLUMN use_markdown BOOLEAN DEFAULT 0;"

    print_success "Successfully added use_markdown column"

    # Verify all existing records have use_markdown = 0
    print_info "Setting use_markdown = 0 for all existing records (backward compatibility)..."
    sqlite3 "$DB_PATH" "UPDATE trouble_reports SET use_markdown = 0 WHERE use_markdown IS NULL;"

    local updated_count=$(sqlite3 "$DB_PATH" "SELECT changes();")
    if [ "$updated_count" -gt 0 ]; then
        print_success "Updated $updated_count existing records to use_markdown = 0"
    else
        print_info "All records already have proper use_markdown values"
    fi
}

# Validate migration success
validate_migration() {
    print_header "Validating Migration"

    # Check if use_markdown column exists and has correct properties
    local column_info=$(sqlite3 "$DB_PATH" "PRAGMA table_info(trouble_reports);" | grep "use_markdown")

    if [ -z "$column_info" ]; then
        print_error "Migration validation failed: use_markdown column not found"
        return 1
    fi

    print_success "use_markdown column exists"
    print_info "Column details: $column_info"

    # Check data integrity
    local total_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM trouble_reports;")
    local markdown_true_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM trouble_reports WHERE use_markdown = 1;")
    local markdown_false_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM trouble_reports WHERE use_markdown = 0;")
    local null_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM trouble_reports WHERE use_markdown IS NULL;")

    print_success "Data integrity check:"
    echo "  Total records: $total_count"
    echo "  use_markdown = 1 (true): $markdown_true_count"
    echo "  use_markdown = 0 (false): $markdown_false_count"
    echo "  use_markdown = NULL: $null_count"

    if [ "$null_count" -gt 0 ]; then
        print_warning "Found $null_count records with NULL use_markdown values"
        print_info "Setting NULL values to 0 for consistency..."
        sqlite3 "$DB_PATH" "UPDATE trouble_reports SET use_markdown = 0 WHERE use_markdown IS NULL;"
        print_success "Fixed NULL values"
    fi

    # Verify schema matches expected structure
    print_info "Final schema verification:"
    sqlite3 "$DB_PATH" "PRAGMA table_info(trouble_reports);"

    print_success "Migration validation completed successfully"
}

# Show migration summary
show_migration_summary() {
    print_header "Migration Summary"

    echo "The following changes have been made to your database:"
    echo ""
    echo "✓ Added 'use_markdown' column to trouble_reports table"
    echo "✓ Set default value to FALSE (0) for backward compatibility"
    echo "✓ All existing trouble reports will display as plain text"
    echo "✓ New trouble reports default to plain text unless markdown is enabled"
    echo ""
    echo "Next steps:"
    echo "1. Restart your pg-press application"
    echo "2. Users can now enable markdown formatting when creating/editing reports"
    echo "3. Markdown content will be rendered as HTML in the web interface"
    echo "4. PDF exports will format markdown content appropriately"
    echo ""
    echo "Supported markdown features:"
    echo "• Headers (# ## ###)"
    echo "• Bold (**text**) and italic (*text*)"
    echo "• Code blocks and inline code"
    echo "• Lists (ordered and unordered)"
    echo "• Links and strikethrough"
    echo "• Tables"
    echo ""
    echo "Security features:"
    echo "• HTML sanitization prevents XSS attacks"
    echo "• Safe rendering using Go's template.HTML"
    echo "• Dangerous elements and scripts are automatically removed"
}

# Dry run mode
dry_run() {
    print_header "Migration Dry Run"

    check_database
    show_current_schema

    if check_migration_needed; then
        echo ""
        print_info "This migration would perform the following changes:"
        echo "1. Add 'use_markdown BOOLEAN DEFAULT 0' column to trouble_reports table"
        echo "2. Set use_markdown = 0 for all existing records"
        echo "3. Validate the migration was successful"
        echo ""
        print_warning "No actual changes will be made in dry-run mode"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║           Trouble Reports Markdown Migration Script         ║"
    echo "║                                                              ║"
    echo "║ This script adds markdown support to trouble reports by     ║"
    echo "║ adding a use_markdown column to the trouble_reports table.  ║"
    echo "║                                                              ║"
    echo "║ ⚠ IMPORTANT: Stop your application before running this!     ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -f "pg-press.db" ] 2>/dev/null; then
        if [ ! -f "go.mod" ]; then
            print_error "go.mod not found. Please run this script from the project root directory"
            print_info "Usage: ./scripts/migrate-trouble-reports-markdown.sh"
            exit 1
        fi
    fi

    # Check prerequisites
    command -v sqlite3 >/dev/null 2>&1 || {
        print_error "sqlite3 is required but not installed."
        print_info "Please install sqlite3 and try again."
        exit 1
    }

    check_database
    show_current_schema

    if ! check_migration_needed; then
        print_warning "Migration may already be complete"
        echo ""
    fi

    # Ask for confirmation
    print_warning "This migration will modify your database structure"
    print_info "A backup will be created automatically before making changes"
    echo ""
    read -p "Do you want to proceed with the migration? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Migration cancelled by user"
        exit 0
    fi

    # Perform migration
    ensure_backup_dir
    create_backup
    perform_migration
    validate_migration
    show_migration_summary

    print_header "Migration Completed Successfully"
    print_success "Trouble reports now support markdown formatting!"
    print_warning "Remember to restart your pg-press application"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "This script adds markdown support to trouble reports by adding"
        echo "a use_markdown column to the trouble_reports table."
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  --dry-run      Show what would be changed without making changes"
        echo ""
        echo "What this migration does:"
        echo "  • Adds use_markdown BOOLEAN DEFAULT 0 column"
        echo "  • Sets all existing records to use_markdown = 0"
        echo "  • Creates backup before making changes"
        echo "  • Validates migration success"
        echo ""
        echo "Prerequisites:"
        echo "  • SQLite3 must be installed"
        echo "  • pg-press.db must exist"
        echo "  • Run from project root directory"
        echo "  • Stop pg-press application before running"
        echo ""
        echo "After migration:"
        echo "  • Users can enable markdown in trouble report edit dialog"
        echo "  • Markdown content renders as HTML in web interface"
        echo "  • PDF exports format markdown content appropriately"
        echo "  • All existing reports continue to work as plain text"
        exit 0
        ;;
    --dry-run)
        dry_run
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
