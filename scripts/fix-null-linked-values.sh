#!/bin/bash

# Quick fix script for NULL linked values in notes table
# This script converts NULL values in the linked column to empty strings

set -e

# Configuration
DB_PATH="${DB_PATH:-./pg-press.db}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions for colored output
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

# Check if notes table exists
check_notes_table() {
    local notes_table_exists=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='notes';" | wc -l)
    if [ "$notes_table_exists" -eq 0 ]; then
        print_error "Notes table does not exist in database"
        exit 1
    fi
    print_success "Notes table found"
}

# Check for NULL values
check_null_values() {
    print_info "Checking for NULL linked values..."

    local null_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked IS NULL;" 2>/dev/null || echo "0")

    if [ "$null_count" -eq 0 ]; then
        print_success "No NULL linked values found - no fix needed"
        return 1
    else
        print_warning "Found $null_count notes with NULL linked values"
        return 0
    fi
}

# Apply the fix
apply_fix() {
    print_info "Converting NULL linked values to empty strings..."

    sqlite3 "$DB_PATH" "UPDATE notes SET linked = '' WHERE linked IS NULL;"

    if [ $? -eq 0 ]; then
        print_success "Successfully updated NULL linked values"
    else
        print_error "Failed to update NULL linked values"
        exit 1
    fi
}

# Verify the fix
verify_fix() {
    print_info "Verifying the fix..."

    local null_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked IS NULL;" 2>/dev/null || echo "0")
    local total_notes=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes;" 2>/dev/null || echo "0")
    local empty_linked=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked = '';" 2>/dev/null || echo "0")

    print_success "Verification results:"
    echo "  - Total notes: $total_notes"
    echo "  - NULL linked values: $null_count"
    echo "  - Empty linked values: $empty_linked"

    if [ "$null_count" -eq 0 ]; then
        print_success "Fix verified successfully - no more NULL values"
    else
        print_error "Fix verification failed - still have NULL values"
        exit 1
    fi
}

# Show current distribution
show_distribution() {
    print_info "Current linked values distribution:"

    sqlite3 "$DB_PATH" "
    SELECT
        CASE
            WHEN linked IS NULL THEN 'NULL'
            WHEN linked = '' THEN 'Empty/Unlinked'
            WHEN linked LIKE 'tool_%' THEN 'Tool Links'
            WHEN linked LIKE 'press_%' THEN 'Press Links'
            ELSE 'Other'
        END as link_type,
        COUNT(*) as count
    FROM notes
    GROUP BY
        CASE
            WHEN linked IS NULL THEN 'NULL'
            WHEN linked = '' THEN 'Empty/Unlinked'
            WHEN linked LIKE 'tool_%' THEN 'Tool Links'
            WHEN linked LIKE 'press_%' THEN 'Press Links'
            ELSE 'Other'
        END
    ORDER BY count DESC;
    " | while IFS='|' read -r type count; do
        echo "  - $type: $count"
    done
}

# Main execution
main() {
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                NULL Linked Values Fix                       ║${NC}"
    echo -e "${BLUE}║                                                              ║${NC}"
    echo -e "${BLUE}║ This script fixes NULL values in the notes.linked column   ║${NC}"
    echo -e "${BLUE}║ to prevent scanning errors in the Go application.           ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Check if running from correct directory
    if [ ! -f "go.mod" ] || [ ! -d "scripts" ]; then
        print_error "Please run this script from the project root directory"
        exit 1
    fi

    # Execute fix steps
    check_database
    check_notes_table

    if ! check_null_values; then
        show_distribution
        echo ""
        print_success "No action needed - database is already in good state"
        exit 0
    fi

    echo ""
    print_warning "This will update NULL linked values to empty strings"
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Fix cancelled by user"
        exit 0
    fi

    apply_fix
    verify_fix
    echo ""
    show_distribution

    echo ""
    print_success "Fix completed successfully!"
    print_info "You can now access the /notes page without scanning errors"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0"
        echo ""
        echo "This script fixes NULL values in the notes.linked column."
        echo ""
        echo "The script will:"
        echo "1. Check for NULL values in notes.linked column"
        echo "2. Convert NULL values to empty strings"
        echo "3. Verify the fix was applied correctly"
        echo ""
        echo "Run this script from the project root directory."
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
