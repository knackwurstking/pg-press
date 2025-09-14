#!/bin/bash

# Migration script for pg-press: Migrate mods columns to modifications table
# This script orchestrates the complete migration from old mods system to new modification system

set -euo pipefail

# Default values
DB_PATH="./data.db"
ACTION=""
FORCE=false
DRY_RUN=false
VERBOSE=false
USE_GO_SCRIPT=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] ACTION

Migrate pg-press from old mods columns to new modifications table system.

ACTIONS:
    setup       - Set up database schema (add mods columns, create modifications table)
    migrate     - Migrate data from mods columns to modifications table
    verify      - Verify migration integrity
    cleanup     - Remove old mods columns (DESTRUCTIVE!)
    status      - Show current migration status
    full        - Run complete migration process (setup + migrate + verify)

OPTIONS:
    -d, --db PATH       Database path (default: ./data.db)
    -f, --force         Force operation without confirmation
    -n, --dry-run       Show what would be done without making changes
    -v, --verbose       Verbose output
    -g, --go-script     Use standalone Go script instead of pgpress commands
    -h, --help          Show this help

EXAMPLES:
    # Run complete migration
    $0 full

    # Dry run to see what would happen
    $0 --dry-run full

    # Use custom database path
    $0 --db /path/to/data.db migrate

    # Force cleanup without confirmation
    $0 --force cleanup

MIGRATION WORKFLOW:
    1. Backup your database first!
    2. Run: $0 setup
    3. Run: $0 migrate
    4. Run: $0 verify
    5. Test your application thoroughly
    6. Run: $0 cleanup (optional, destructive)

EOF
}

# Function to backup database
backup_database() {
    if [[ ! -f "$DB_PATH" ]]; then
        print_error "Database file not found: $DB_PATH"
        exit 1
    fi

    local backup_path="${DB_PATH}.backup.$(date +%Y%m%d_%H%M%S)"

    if [[ "$DRY_RUN" == true ]]; then
        print_info "DRY RUN: Would create backup at $backup_path"
        return 0
    fi

    print_info "Creating database backup..."
    cp "$DB_PATH" "$backup_path"
    print_success "Database backed up to: $backup_path"
}

# Function to check database exists and is accessible
check_database() {
    if [[ ! -f "$DB_PATH" ]]; then
        print_error "Database file not found: $DB_PATH"
        print_info "Please specify the correct path with --db option"
        exit 1
    fi

    # Test database connectivity
    if ! sqlite3 "$DB_PATH" "SELECT 1;" >/dev/null 2>&1; then
        print_error "Cannot connect to database: $DB_PATH"
        print_info "Please check the file permissions and that the database is not corrupted"
        exit 1
    fi

    print_success "Database connection verified: $DB_PATH"
}

# Function to set up database schema
setup_schema() {
    print_info "Setting up database schema..."

    local sql_script="$(dirname "$0")/migrate_mods_to_modifications.sql"

    if [[ ! -f "$sql_script" ]]; then
        print_error "SQL script not found: $sql_script"
        exit 1
    fi

    if [[ "$DRY_RUN" == true ]]; then
        print_info "DRY RUN: Would execute SQL setup script"
        print_info "SQL script location: $sql_script"
        return 0
    fi

    if sqlite3 "$DB_PATH" < "$sql_script"; then
        print_success "Database schema setup completed"
    else
        print_error "Failed to set up database schema"
        exit 1
    fi
}

# Function to run migration using pgpress command
run_pgpress_migration() {
    local cmd="migrate"
    [[ "$ACTION" == "verify" ]] && cmd="verify"
    [[ "$ACTION" == "cleanup" ]] && cmd="cleanup"
    [[ "$ACTION" == "status" ]] && cmd="status"

    local pgpress_cmd="pgpress migration $cmd"
    [[ -n "$DB_PATH" && "$DB_PATH" != "./data.db" ]] && pgpress_cmd="$pgpress_cmd --db '$DB_PATH'"
    [[ "$FORCE" == true && "$cmd" == "cleanup" ]] && pgpress_cmd="$pgpress_cmd --force"

    if [[ "$DRY_RUN" == true ]]; then
        print_info "DRY RUN: Would execute: $pgpress_cmd"
        return 0
    fi

    print_info "Executing: $pgpress_cmd"

    if eval "$pgpress_cmd"; then
        print_success "pgpress migration $cmd completed successfully"
    else
        print_error "pgpress migration $cmd failed"
        exit 1
    fi
}

# Function to run migration using Go script
run_go_migration() {
    local go_script="$(dirname "$0")/migrate_mods.go"

    if [[ ! -f "$go_script" ]]; then
        print_error "Go migration script not found: $go_script"
        exit 1
    fi

    local go_cmd="go run '$go_script' --db '$DB_PATH' --action '$ACTION'"
    [[ "$FORCE" == true ]] && go_cmd="$go_cmd --force"
    [[ "$DRY_RUN" == true ]] && go_cmd="$go_cmd --dry-run"
    [[ "$VERBOSE" == true ]] && go_cmd="$go_cmd --v"

    if [[ "$DRY_RUN" == true ]]; then
        print_info "DRY RUN: Would execute: $go_cmd"
        return 0
    fi

    print_info "Executing Go migration script..."

    if eval "$go_cmd"; then
        print_success "Go migration script completed successfully"
    else
        print_error "Go migration script failed"
        exit 1
    fi
}

# Function to run full migration process
run_full_migration() {
    print_info "Starting full migration process..."

    # Step 1: Setup schema
    setup_schema

    # Step 2: Migrate data
    ACTION="migrate"
    if [[ "$USE_GO_SCRIPT" == true ]]; then
        run_go_migration
    else
        run_pgpress_migration
    fi

    # Step 3: Verify migration
    ACTION="verify"
    if [[ "$USE_GO_SCRIPT" == true ]]; then
        run_go_migration
    else
        run_pgpress_migration
    fi

    print_success "Full migration completed successfully!"
    print_warning "Important: Test your application thoroughly before running cleanup!"
    print_info "To remove old mods columns, run: $0 cleanup"
}

# Function to confirm destructive operations
confirm_destructive_operation() {
    if [[ "$FORCE" == true || "$DRY_RUN" == true ]]; then
        return 0
    fi

    print_warning "This operation will make changes to your database!"
    print_warning "Make sure you have a backup before proceeding."

    if [[ "$ACTION" == "cleanup" ]]; then
        print_error "CLEANUP IS DESTRUCTIVE AND CANNOT BE UNDONE!"
        print_warning "It will permanently remove the old mods columns."
    fi

    read -p "Are you sure you want to continue? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_info "Operation cancelled."
        exit 0
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--db)
            DB_PATH="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -g|--go-script)
            USE_GO_SCRIPT=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        setup|migrate|verify|cleanup|status|full)
            ACTION="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate action
if [[ -z "$ACTION" ]]; then
    print_error "No action specified!"
    show_usage
    exit 1
fi

# Main execution
main() {
    print_info "PG-Press Migration Tool"
    print_info "Action: $ACTION"
    print_info "Database: $DB_PATH"
    [[ "$DRY_RUN" == true ]] && print_warning "DRY RUN MODE - No changes will be made"
    [[ "$FORCE" == true ]] && print_warning "FORCE MODE - Skipping confirmations"

    # Check database
    check_database

    # Create backup (except for status/dry-run)
    if [[ "$ACTION" != "status" && "$DRY_RUN" != true ]]; then
        backup_database
    fi

    # Confirm destructive operations
    if [[ "$ACTION" =~ ^(migrate|cleanup|full)$ ]]; then
        confirm_destructive_operation
    fi

    # Execute action
    case "$ACTION" in
        setup)
            setup_schema
            ;;
        migrate|verify|cleanup|status)
            if [[ "$USE_GO_SCRIPT" == true ]]; then
                run_go_migration
            else
                run_pgpress_migration
            fi
            ;;
        full)
            run_full_migration
            ;;
    esac

    print_success "Migration script completed successfully!"
}

# Run main function
main "$@"
