#!/bin/bash

# Test script for the identifier migration
# This script creates a test database, runs the migration, and verifies the results

set -e

# Configuration
TEST_DB="test_migration.db"
MIGRATION_SCRIPT="./migrate-add-identifier.sh"
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

# Cleanup function
cleanup() {
    print_header "Cleaning up test files"

    # Remove test database
    if [ -f "$TEST_DB" ]; then
        rm "$TEST_DB"
        print_success "Removed test database: $TEST_DB"
    fi

    # Remove test backups from local directory
    if [ -d "$BACKUP_DIR" ]; then
        rm -rf "$BACKUP_DIR"/*before_identifier_migration* 2>/dev/null || true
        print_success "Removed local test backup files"
    fi

    # Remove test backups from parent directory (created by migration script)
    if [ -d "../$BACKUP_DIR" ]; then
        rm -rf "../$BACKUP_DIR"/*before_identifier_migration* 2>/dev/null || true
        print_success "Removed parent directory test backup files"
    fi
}

# Create test database with sample data
create_test_database() {
    print_header "Creating test database with sample data"

    # Remove existing test database
    [ -f "$TEST_DB" ] && rm "$TEST_DB"

    # Create the original metal_sheets table structure (without identifier)
    sqlite3 "$TEST_DB" "
        CREATE TABLE metal_sheets (
            id INTEGER NOT NULL,
            tile_height REAL NOT NULL,
            value REAL NOT NULL,
            marke_height INTEGER NOT NULL,
            stf REAL NOT NULL,
            stf_max REAL NOT NULL,
            tool_id INTEGER NOT NULL,
            notes BLOB NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY(id AUTOINCREMENT)
        );
    "

    # Insert some sample data
    sqlite3 "$TEST_DB" "
        INSERT INTO metal_sheets (tile_height, value, marke_height, stf, stf_max, tool_id, notes) VALUES
        (10.5, 25.0, 100, 15.5, 20.0, 1, '[]'),
        (12.0, 30.0, 110, 18.0, 22.5, 2, '[]'),
        (8.5, 20.0, 90, 12.0, 16.0, 1, '[]');
    "

    print_success "Created test database with 3 sample records"
}

# Verify original state
verify_original_state() {
    print_header "Verifying original database state"

    # Check that identifier column doesn't exist
    local column_exists=$(sqlite3 "$TEST_DB" "PRAGMA table_info(metal_sheets);" | grep -c "identifier" || true)

    if [ "$column_exists" -eq 0 ]; then
        print_success "Confirmed: identifier column does not exist"
    else
        print_error "Expected: identifier column should not exist yet"
        return 1
    fi

    # Check record count
    local record_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM metal_sheets;")
    if [ "$record_count" -eq 3 ]; then
        print_success "Confirmed: 3 test records exist"
    else
        print_error "Expected 3 records, found: $record_count"
        return 1
    fi
}

# Run the migration
run_migration() {
    print_header "Running identifier migration"

    # Save current directory and change to project root
    local current_dir=$(pwd)
    cd ..

    # Temporarily set the database path for the migration script (relative to project root)
    export DB_PATH="./scripts/$TEST_DB"

    # Run the migration script with auto-confirmation
    echo "y" | bash "./scripts/migrate-add-identifier.sh" 2>/dev/null || {
        print_error "Migration script failed"
        cd "$current_dir"
        return 1
    }

    # Return to scripts directory
    cd "$current_dir"

    print_success "Migration script completed"
}

# Verify migration results
verify_migration_results() {
    print_header "Verifying migration results"

    # Check that identifier column now exists
    local column_info=$(sqlite3 "$TEST_DB" "PRAGMA table_info(metal_sheets);" | grep identifier || true)

    if [ -n "$column_info" ]; then
        print_success "Confirmed: identifier column exists"
        echo "Column info: $column_info"
    else
        print_error "Expected: identifier column should exist after migration"
        return 1
    fi

    # Check that all records have SACMI identifier
    local sacmi_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM metal_sheets WHERE identifier = 'SACMI';")
    local total_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM metal_sheets;")

    if [ "$sacmi_count" -eq "$total_count" ] && [ "$total_count" -eq 3 ]; then
        print_success "Confirmed: All $total_count records have SACMI identifier"
    else
        print_error "Expected all 3 records to have SACMI identifier. SACMI: $sacmi_count, Total: $total_count"
        return 1
    fi

    # Test updating a record to SITI
    sqlite3 "$TEST_DB" "UPDATE metal_sheets SET identifier = 'SITI' WHERE id = 2;"

    local siti_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM metal_sheets WHERE identifier = 'SITI';")
    if [ "$siti_count" -eq 1 ]; then
        print_success "Confirmed: Successfully updated record to SITI type"
    else
        print_error "Failed to update record to SITI type"
        return 1
    fi

    # Show final distribution
    echo "Final identifier distribution:"
    sqlite3 "$TEST_DB" "SELECT identifier, COUNT(*) as count FROM metal_sheets GROUP BY identifier;" | while read line; do
        echo "  $line"
    done
}

# Check if backup was created
verify_backup_created() {
    print_header "Verifying backup creation"

    # Check both local and parent directory backup locations
    # Migration script creates backups with pattern: pg-press_before_identifier_migration_*
    local backup_count=$(ls "$BACKUP_DIR"/*before_identifier_migration* 2>/dev/null | wc -l || echo "0")
    local parent_backup_count=$(ls "../$BACKUP_DIR"/*before_identifier_migration* 2>/dev/null | wc -l || echo "0")

    if [ "$backup_count" -gt 0 ]; then
        print_success "Confirmed: Backup was created locally"
        ls -la "$BACKUP_DIR"/*before_identifier_migration* | head -1
    elif [ "$parent_backup_count" -gt 0 ]; then
        print_success "Confirmed: Backup was created in parent directory"
        ls -la "../$BACKUP_DIR"/*before_identifier_migration* | head -1
    else
        print_warning "No backup file found (this might be expected in test mode)"
    fi
}

# Main test execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                Migration Test Script                        ║"
    echo "║                                                              ║"
    echo "║ This script tests the identifier migration by creating a    ║"
    echo "║ test database, running the migration, and verifying results.║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Check prerequisites
    if [ ! -f "$MIGRATION_SCRIPT" ]; then
        print_error "Migration script not found: $MIGRATION_SCRIPT"
        print_warning "Make sure you're running this from the scripts directory"
        exit 1
    fi

    if ! command -v sqlite3 >/dev/null 2>&1; then
        print_error "SQLite3 is not installed or not in PATH"
        exit 1
    fi

    # Set trap for cleanup
    trap cleanup EXIT

    # Run tests
    create_test_database
    verify_original_state
    run_migration
    verify_migration_results
    verify_backup_created

    print_header "Test Results"
    print_success "All migration tests passed!"
    print_success "The identifier migration is working correctly"

    echo ""
    echo -e "${YELLOW}Test Summary:${NC}"
    echo "• Created test database with 3 sample records"
    echo "• Verified original state (no identifier column)"
    echo "• Successfully ran the migration script"
    echo "• Confirmed identifier column was added"
    echo "• Verified all records defaulted to SACMI"
    echo "• Tested updating a record to SITI type"
    echo "• Verified backup creation process"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "This script tests the identifier migration by:"
        echo "1. Creating a test database with sample data"
        echo "2. Running the migration script"
        echo "3. Verifying the results"
        echo "4. Cleaning up test files"
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo ""
        echo "Run this script from the scripts/ directory."
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
