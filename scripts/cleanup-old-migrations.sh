#!/bin/bash

# Cleanup script for old migration files
# This script removes deprecated migration scripts that are no longer needed

set -e

# Configuration
BACKUP_DIR="./backups"
OLD_MIGRATIONS=(
    "migrate-add-identifier.sh"
    "test-migration.sh"
)

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

# Check which old migration files exist
check_old_migrations() {
    print_header "Checking for old migration files"

    local found_files=()

    for migration in "${OLD_MIGRATIONS[@]}"; do
        if [ -f "$migration" ]; then
            found_files+=("$migration")
            print_warning "Found deprecated migration: $migration"
        else
            print_info "Migration file not found: $migration (already removed)"
        fi
    done

    if [ ${#found_files[@]} -eq 0 ]; then
        print_success "No old migration files found to clean up"
        return 1
    fi

    echo ""
    print_info "Found ${#found_files[@]} deprecated migration file(s) to clean up"
    return 0
}

# Create backup directory if it doesn't exist
ensure_backup_dir() {
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        print_success "Created backup directory: $BACKUP_DIR"
    fi
}

# Backup old migration files before deletion
backup_old_migrations() {
    print_header "Backing up old migration files"

    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_subdir="${BACKUP_DIR}/old_migrations_${timestamp}"

    mkdir -p "$backup_subdir"

    local backed_up_count=0

    for migration in "${OLD_MIGRATIONS[@]}"; do
        if [ -f "$migration" ]; then
            cp "$migration" "$backup_subdir/"
            print_success "Backed up: $migration"
            ((backed_up_count++))
        fi
    done

    if [ $backed_up_count -gt 0 ]; then
        print_success "Backed up $backed_up_count file(s) to: $backup_subdir"
        echo "Backup location: $backup_subdir"
    fi
}

# Remove old migration files
remove_old_migrations() {
    print_header "Removing old migration files"

    local removed_count=0

    for migration in "${OLD_MIGRATIONS[@]}"; do
        if [ -f "$migration" ]; then
            rm "$migration"
            print_success "Removed: $migration"
            ((removed_count++))
        fi
    done

    if [ $removed_count -gt 0 ]; then
        print_success "Removed $removed_count old migration file(s)"
    fi
}

# Show information about why these migrations are deprecated
show_deprecation_info() {
    print_header "Why these migrations are deprecated"

    echo "The following migrations are no longer needed:"
    echo ""
    echo "• migrate-add-identifier.sh:"
    echo "  - Replaced by migrate-notes-system-refactor.sh"
    echo "  - The new migration handles all schema changes comprehensively"
    echo "  - The identifier field functionality is included in the new system"
    echo ""
    echo "• test-migration.sh:"
    echo "  - Was a testing script for the old migration system"
    echo "  - No longer relevant with the new migration approach"
    echo ""
    echo "The new migration system provides:"
    echo "• Generic notes linking instead of tool-specific linking"
    echo "• Cleaner database schema"
    echo "• Better semantic correctness for press notes"
    echo "• Extensibility for future entity types"
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║               Old Migration Cleanup Script                  ║"
    echo "║                                                              ║"
    echo "║ This script removes deprecated migration files that are     ║"
    echo "║ no longer needed after the notes system refactoring.        ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"

    # Check if running from correct directory
    if [ ! -f "../go.mod" ] || [ ! -d "." ]; then
        print_error "Please run this script from the scripts directory"
        print_info "Usage: cd scripts && ./cleanup-old-migrations.sh"
        exit 1
    fi

    # Check for old migrations
    if ! check_old_migrations; then
        print_success "No cleanup needed - all old migrations already removed"
        exit 0
    fi

    show_deprecation_info

    # Ask for confirmation
    echo ""
    print_warning "This will permanently remove the old migration files."
    print_info "Files will be backed up before removal."
    read -p "Do you want to continue with the cleanup? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Cleanup cancelled by user"
        exit 0
    fi

    ensure_backup_dir
    backup_old_migrations
    remove_old_migrations

    print_header "Cleanup Completed Successfully"
    print_success "Old migration files have been cleaned up"
    print_info "Backups are available in the backups directory if needed"
    print_info "The database migration system is now using migrate-notes-system-refactor.sh"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "This script removes deprecated migration files that are no longer needed."
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  --list         List files that would be removed (dry run)"
        echo "  --force        Skip confirmation prompts"
        echo ""
        echo "Deprecated files that will be removed:"
        for migration in "${OLD_MIGRATIONS[@]}"; do
            echo "  • $migration"
        done
        echo ""
        echo "Files are backed up before removal to scripts/backups/"
        echo "Run this script from the scripts directory."
        exit 0
        ;;
    --list)
        echo "Files that would be removed:"
        for migration in "${OLD_MIGRATIONS[@]}"; do
            if [ -f "$migration" ]; then
                echo "  ✓ $migration (exists)"
            else
                echo "  - $migration (not found)"
            fi
        done
        exit 0
        ;;
    --force)
        # Skip confirmation, but still show info
        main() {
            echo "Force cleanup mode - skipping confirmations"
            if ! check_old_migrations; then
                echo "No cleanup needed"
                exit 0
            fi
            ensure_backup_dir
            backup_old_migrations
            remove_old_migrations
            echo "Cleanup completed"
        }
        main
        ;;
    *)
        main "$@"
        ;;
esac
