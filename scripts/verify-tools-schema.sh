#!/bin/bash

# Verification script for Tools Schema
# This script verifies the tools table schema and dead status functionality

set -e

# Configuration
DB_PATH="${DB_PATH:-./pg-press.db}"

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

# Function to check if database exists
check_database() {
    print_header "Database Check"

    if [ ! -f "$DB_PATH" ]; then
        print_error "Database file not found: $DB_PATH"
        exit 1
    fi

    print_success "Database file exists: $DB_PATH"

    # Check if we can read the database
    if ! sqlite3 "$DB_PATH" "SELECT 1;" > /dev/null 2>&1; then
        print_error "Cannot read database file (permissions or corruption)"
        exit 1
    fi

    print_success "Database is readable"
}

# Function to check tools table exists
check_tools_table() {
    print_header "Tools Table Check"

    TABLE_EXISTS=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='tools';" | wc -l)

    if [ "$TABLE_EXISTS" -eq 0 ]; then
        print_error "Tools table does not exist"
        exit 1
    fi

    print_success "Tools table exists"
}

# Function to verify schema
verify_schema() {
    print_header "Schema Verification"

    # Get current schema
    SCHEMA=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);")

    # Expected columns and their properties
    declare -A EXPECTED_COLUMNS=(
        ["id"]="INTEGER"
        ["position"]="TEXT"
        ["format"]="BLOB"
        ["type"]="TEXT"
        ["code"]="TEXT"
        ["regenerating"]="INTEGER"
        ["is_dead"]="INTEGER"
        ["press"]="INTEGER"
    )

    print_info "Current tools table schema:"
    echo "$SCHEMA" | while read line; do
        echo "  $line"
    done

    echo

    # Check each expected column
    for column in "${!EXPECTED_COLUMNS[@]}"; do
        COLUMN_INFO=$(echo "$SCHEMA" | grep "|$column|" || true)

        if [ -z "$COLUMN_INFO" ]; then
            print_error "Missing column: $column"
            continue
        fi

        COLUMN_TYPE=$(echo "$COLUMN_INFO" | cut -d'|' -f3)
        EXPECTED_TYPE="${EXPECTED_COLUMNS[$column]}"

        if [[ "$COLUMN_TYPE" == "$EXPECTED_TYPE" ]]; then
            print_success "Column $column ($COLUMN_TYPE) ✓"
        else
            print_warning "Column $column type mismatch: got $COLUMN_TYPE, expected $EXPECTED_TYPE"
        fi
    done
}

# Function to verify is_dead column specifics
verify_is_dead_column() {
    print_header "Dead Status Column Verification"

    # Check if is_dead column exists
    IS_DEAD_EXISTS=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "is_dead" || true)

    if [ "$IS_DEAD_EXISTS" -eq 0 ]; then
        print_error "is_dead column is missing"
        print_info "Run the migration script: ./scripts/migrate-tools-dead-status.sh"
        return 1
    fi

    print_success "is_dead column exists"

    # Check column properties
    IS_DEAD_INFO=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep "is_dead")
    print_info "is_dead column info: $IS_DEAD_INFO"

    # Check if column has NOT NULL constraint
    if echo "$IS_DEAD_INFO" | grep -q "|1|"; then
        print_success "is_dead column has NOT NULL constraint"
    else
        print_warning "is_dead column missing NOT NULL constraint"
    fi

    # Check default value
    DEFAULT_VALUE=$(echo "$IS_DEAD_INFO" | cut -d'|' -f5)
    if [ "$DEFAULT_VALUE" = "0" ]; then
        print_success "is_dead column has correct default value (0)"
    else
        print_warning "is_dead column default value: $DEFAULT_VALUE (expected: 0)"
    fi
}

# Function to show statistics
show_statistics() {
    print_header "Tools Statistics"

    # Total tools count
    TOTAL_TOOLS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools;" 2>/dev/null || echo "0")
    print_info "Total tools: $TOTAL_TOOLS"

    if [ "$TOTAL_TOOLS" -eq 0 ]; then
        print_warning "No tools found in database"
        return
    fi

    # Check if is_dead column exists for statistics
    IS_DEAD_EXISTS=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "is_dead" || true)

    if [ "$IS_DEAD_EXISTS" -gt 0 ]; then
        # Alive tools count
        ALIVE_TOOLS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools WHERE is_dead = 0;" 2>/dev/null || echo "0")
        print_success "Alive tools: $ALIVE_TOOLS"

        # Dead tools count
        DEAD_TOOLS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools WHERE is_dead = 1;" 2>/dev/null || echo "0")
        if [ "$DEAD_TOOLS" -gt 0 ]; then
            print_warning "Dead tools: $DEAD_TOOLS"
        else
            print_success "Dead tools: $DEAD_TOOLS"
        fi

        # Tools with press assignments
        ACTIVE_TOOLS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools WHERE is_dead = 0 AND press IS NOT NULL;" 2>/dev/null || echo "0")
        print_info "Tools assigned to press: $ACTIVE_TOOLS"

        # Regenerating tools
        REGEN_TOOLS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools WHERE is_dead = 0 AND regenerating = 1;" 2>/dev/null || echo "0")
        print_info "Regenerating tools: $REGEN_TOOLS"
    else
        print_warning "Cannot show detailed statistics - is_dead column missing"
    fi

    # Show sample tools if any exist
    if [ "$TOTAL_TOOLS" -gt 0 ]; then
        print_info "Sample tools (first 5):"
        if [ "$IS_DEAD_EXISTS" -gt 0 ]; then
            sqlite3 "$DB_PATH" "SELECT id, code, type, CASE WHEN is_dead = 1 THEN 'DEAD' ELSE 'ALIVE' END as status FROM tools ORDER BY id LIMIT 5;" | while read line; do
                echo "  $line"
            done
        else
            sqlite3 "$DB_PATH" "SELECT id, code, type FROM tools ORDER BY id LIMIT 5;" | while read line; do
                echo "  $line"
            done
        fi
    fi
}

# Function to test basic functionality
test_functionality() {
    print_header "Functionality Test"

    # Test basic queries
    if sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools;" > /dev/null 2>&1; then
        print_success "Basic SELECT query works"
    else
        print_error "Basic SELECT query failed"
        return 1
    fi

    # Test is_dead column queries if it exists
    IS_DEAD_EXISTS=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "is_dead" || true)

    if [ "$IS_DEAD_EXISTS" -gt 0 ]; then
        if sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools WHERE is_dead = 0;" > /dev/null 2>&1; then
            print_success "is_dead column queries work"
        else
            print_error "is_dead column queries failed"
            return 1
        fi

        if sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM tools WHERE is_dead = 1;" > /dev/null 2>&1; then
            print_success "Dead tools queries work"
        else
            print_error "Dead tools queries failed"
            return 1
        fi
    else
        print_warning "Skipping is_dead column functionality tests"
    fi
}

# Function to show usage information
show_usage() {
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -d, --database PATH    Specify database path (default: ./pg-press.db)"
    echo "  -h, --help            Show this help message"
    echo "  --stats-only          Show only statistics"
    echo "  --quiet               Show only errors and warnings"
    echo
    echo "Environment Variables:"
    echo "  DB_PATH               Database path (default: ./pg-press.db)"
    echo
    echo "Examples:"
    echo "  $0                                    # Full verification with default database"
    echo "  $0 -d /path/to/custom.db             # Verify custom database"
    echo "  $0 --stats-only                      # Show only statistics"
    echo "  DB_PATH=/custom/path.db $0           # Use environment variable"
}

# Parse command line arguments
STATS_ONLY=false
QUIET=false

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
        --stats-only)
            STATS_ONLY=true
            shift
            ;;
        --quiet)
            QUIET=true
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
    if [ "$QUIET" = false ]; then
        print_header "Tools Schema Verification"
        print_info "Database: $DB_PATH"

        check_database
        check_tools_table

        if [ "$STATS_ONLY" = false ]; then
            verify_schema
            verify_is_dead_column
            test_functionality
        fi
    fi

    show_statistics

    if [ "$QUIET" = false ]; then
        print_header "Verification Complete"

        IS_DEAD_EXISTS=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "is_dead" || true)

        if [ "$IS_DEAD_EXISTS" -gt 0 ]; then
            print_success "Database schema is up to date"
            print_info "Tools dead status feature is available"
        else
            print_warning "Database schema needs migration"
            print_info "Run: ./scripts/migrate-tools-dead-status.sh"
        fi
    fi
}

# Run the script
main "$@"
