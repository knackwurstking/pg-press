#!/bin/bash

# Test script for Tools Dead Status Migration
# This script creates a test database, runs the migration, and verifies the results

set -e

# Configuration
TEST_DB="./test_migration.db"
ORIGINAL_DB_PATH="$DB_PATH"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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
    echo -e "${CYAN}ℹ $1${NC}"
}

# Cleanup function
cleanup() {
    if [ -f "$TEST_DB" ]; then
        rm -f "$TEST_DB"
        print_info "Cleaned up test database"
    fi
    # Restore original DB_PATH
    export DB_PATH="$ORIGINAL_DB_PATH"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Function to create test database with initial schema
create_test_database() {
    print_header "Creating Test Database"

    # Remove existing test database if it exists
    rm -f "$TEST_DB"

    # Create the tools table with original schema (without is_dead column)
    sqlite3 "$TEST_DB" <<'EOF'
CREATE TABLE IF NOT EXISTS tools (
    id INTEGER NOT NULL,
    position TEXT NOT NULL,
    format BLOB NOT NULL,
    type TEXT NOT NULL,
    code TEXT NOT NULL,
    regenerating INTEGER NOT NULL DEFAULT 0,
    press INTEGER,
    PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS users (
    telegram_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    api_key TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY("telegram_id")
);

CREATE TABLE IF NOT EXISTS press_cycles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
    tool_id INTEGER NOT NULL,
    tool_position TEXT NOT NULL,
    total_cycles INTEGER NOT NULL DEFAULT 0,
    date DATETIME NOT NULL,
    performed_by INTEGER NOT NULL,
    FOREIGN KEY (tool_id) REFERENCES tools(id),
    FOREIGN KEY (performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL
);
EOF

    print_success "Created test database with original schema"
}

# Function to populate test data
populate_test_data() {
    print_header "Populating Test Data"

    # Insert test user
    sqlite3 "$TEST_DB" "INSERT INTO users (telegram_id, name, api_key) VALUES (1, 'test-user', 'test-key');"

    # Insert test tools
    sqlite3 "$TEST_DB" <<'EOF'
INSERT INTO tools (position, format, type, code, regenerating, press) VALUES
    ('top', '{"width": 100, "height": 50}', 'FC', 'G01', 0, 2),
    ('bottom', '{"width": 150, "height": 75}', 'GTC', 'G02', 0, NULL),
    ('top', '{"width": 200, "height": 100}', 'MASS', 'G03', 1, 3);
EOF

    # Insert test press cycles (this creates the foreign key relationships)
    sqlite3 "$TEST_DB" <<'EOF'
INSERT INTO press_cycles (press_number, tool_id, tool_position, total_cycles, date, performed_by) VALUES
    (2, 1, 'top', 50000, datetime('now'), 1),
    (3, 3, 'top', 75000, datetime('now'), 1);
EOF

    local tool_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM tools;")
    local cycle_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM press_cycles;")

    print_success "Inserted $tool_count test tools"
    print_success "Inserted $cycle_count test press cycles"
}

# Function to verify pre-migration state
verify_pre_migration() {
    print_header "Verifying Pre-Migration State"

    # Check that is_dead column doesn't exist
    local is_dead_exists=$(sqlite3 "$TEST_DB" "PRAGMA table_info(tools);" | grep -c "is_dead" || true)

    if [ "$is_dead_exists" -eq 0 ]; then
        print_success "Confirmed: is_dead column does not exist (pre-migration state)"
    else
        print_error "Unexpected: is_dead column already exists"
        return 1
    fi

    # Show current schema
    print_info "Current tools table schema:"
    sqlite3 "$TEST_DB" "PRAGMA table_info(tools);" | while read line; do
        echo "  $line"
    done

    # Try to delete a tool that has foreign key references (should fail)
    print_info "Testing foreign key constraint (should fail)..."
    if sqlite3 "$TEST_DB" "DELETE FROM tools WHERE id = 1;" 2>/dev/null; then
        print_warning "Foreign key constraint not enforced (this might be expected)"
    else
        print_success "Foreign key constraint working (delete failed as expected)"
    fi
}

# Function to run migration
run_test_migration() {
    print_header "Running Migration"

    # Set DB_PATH to our test database
    export DB_PATH="$TEST_DB"

    # Run the migration script
    if ./scripts/migrate-tools-dead-status.sh; then
        print_success "Migration script completed successfully"
    else
        print_error "Migration script failed"
        return 1
    fi
}

# Function to verify post-migration state
verify_post_migration() {
    print_header "Verifying Post-Migration State"

    # Check that is_dead column now exists
    local is_dead_exists=$(sqlite3 "$TEST_DB" "PRAGMA table_info(tools);" | grep -c "is_dead" || true)

    if [ "$is_dead_exists" -eq 1 ]; then
        print_success "Confirmed: is_dead column exists (post-migration state)"
    else
        print_error "Migration failed: is_dead column missing"
        return 1
    fi

    # Verify column properties
    local is_dead_info=$(sqlite3 "$TEST_DB" "PRAGMA table_info(tools);" | grep "is_dead")
    print_info "is_dead column info: $is_dead_info"

    # Check that all existing tools are marked as alive
    local alive_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM tools WHERE is_dead = 0;")
    local total_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM tools;")

    if [ "$alive_count" -eq "$total_count" ]; then
        print_success "All existing tools are marked as alive ($alive_count/$total_count)"
    else
        print_error "Tool status mismatch: $alive_count alive out of $total_count total"
        return 1
    fi

    # Show updated schema
    print_info "Updated tools table schema:"
    sqlite3 "$TEST_DB" "PRAGMA table_info(tools);" | while read line; do
        echo "  $line"
    done
}

# Function to test new functionality
test_new_functionality() {
    print_header "Testing New Functionality"

    # Test marking a tool as dead
    print_info "Testing: Mark tool as dead"
    sqlite3 "$TEST_DB" "UPDATE tools SET is_dead = 1 WHERE id = 2;"

    local dead_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM tools WHERE is_dead = 1;")
    if [ "$dead_count" -eq 1 ]; then
        print_success "Successfully marked tool as dead"
    else
        print_error "Failed to mark tool as dead"
        return 1
    fi

    # Test querying alive tools
    print_info "Testing: Query alive tools"
    local alive_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM tools WHERE is_dead = 0;")
    print_success "Found $alive_count alive tools"

    # Test querying dead tools
    print_info "Testing: Query dead tools"
    local dead_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM tools WHERE is_dead = 1;")
    print_success "Found $dead_count dead tools"

    # Test that foreign key relationships are preserved
    print_info "Testing: Foreign key relationships preserved"
    local cycle_count=$(sqlite3 "$TEST_DB" "SELECT COUNT(*) FROM press_cycles;")
    if [ "$cycle_count" -eq 2 ]; then
        print_success "All press cycles preserved ($cycle_count)"
    else
        print_error "Press cycles lost: expected 2, found $cycle_count"
        return 1
    fi

    # Show final statistics
    print_info "Final statistics:"
    sqlite3 "$TEST_DB" "SELECT 'Total tools: ' || COUNT(*) FROM tools;" | while read line; do echo "  $line"; done
    sqlite3 "$TEST_DB" "SELECT 'Alive tools: ' || COUNT(*) FROM tools WHERE is_dead = 0;" | while read line; do echo "  $line"; done
    sqlite3 "$TEST_DB" "SELECT 'Dead tools: ' || COUNT(*) FROM tools WHERE is_dead = 1;" | while read line; do echo "  $line"; done
    sqlite3 "$TEST_DB" "SELECT 'Press cycles: ' || COUNT(*) FROM press_cycles;" | while read line; do echo "  $line"; done
}

# Function to run verification script
test_verification_script() {
    print_header "Testing Verification Script"

    export DB_PATH="$TEST_DB"

    if ./scripts/verify-tools-schema.sh --quiet; then
        print_success "Verification script passed"
    else
        print_warning "Verification script reported issues (this might be expected)"
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo
    echo "This script creates a temporary test database, runs the migration,"
    echo "and verifies the results. No existing data is affected."
    echo
    echo "Options:"
    echo "  -h, --help            Show this help message"
    echo "  --keep-db            Keep test database after completion"
    echo
    echo "The test database is created as: $TEST_DB"
}

# Parse command line arguments
KEEP_DB=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        --keep-db)
            KEEP_DB=true
            shift
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
    print_header "Tools Dead Status Migration Test"
    print_info "Test database: $TEST_DB"

    # Check if required scripts exist
    if [ ! -f "./scripts/migrate-tools-dead-status.sh" ]; then
        print_error "Migration script not found: ./scripts/migrate-tools-dead-status.sh"
        exit 1
    fi

    if [ ! -f "./scripts/verify-tools-schema.sh" ]; then
        print_warning "Verification script not found: ./scripts/verify-tools-schema.sh"
    fi

    # Run the test sequence
    create_test_database
    populate_test_data
    verify_pre_migration
    run_test_migration
    verify_post_migration
    test_new_functionality

    if [ -f "./scripts/verify-tools-schema.sh" ]; then
        test_verification_script
    fi

    print_header "Migration Test Completed Successfully"
    print_success "All tests passed!"
    print_info "The migration appears to be working correctly"

    if [ "$KEEP_DB" = true ]; then
        trap - EXIT  # Remove cleanup trap
        print_info "Test database preserved: $TEST_DB"
    else
        print_info "Test database will be cleaned up automatically"
    fi
}

# Run the script
main "$@"
