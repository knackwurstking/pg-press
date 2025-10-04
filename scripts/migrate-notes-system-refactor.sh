#!/bin/bash

# Migration script for Notes System Refactoring
# This script updates the database schema for the new generic notes linking system

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

# Create backup of database
backup_database() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local backup_file="${BACKUP_DIR}/pg-press_before_notes_refactor_${timestamp}.db"

    print_header "Creating database backup"
    cp "$DB_PATH" "$backup_file"
    print_success "Database backed up to: $backup_file"
    echo "Backup location: $backup_file"
}

# Check current schema and existing data
analyze_current_schema() {
    print_header "Analyzing current database schema"

    # Check notes table structure
    local notes_table_exists=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='notes';" | wc -l)
    if [ "$notes_table_exists" -eq 0 ]; then
        print_info "Notes table does not exist - will be created"
    else
        print_success "Notes table exists"
        local notes_columns=$(sqlite3 "$DB_PATH" "PRAGMA table_info(notes);")
        echo "Current notes table structure:"
        echo "$notes_columns" | while IFS='|' read -r cid name type notnull dflt_value pk; do
            echo "  - $name ($type)"
        done
    fi

    # Check tools table for notes column
    local tools_notes_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "notes" || true)
    if [ "$tools_notes_column" -gt 0 ]; then
        print_warning "Tools table has notes column (will be removed)"
    else
        print_info "Tools table has no notes column (already clean)"
    fi

    # Check metal_sheets table for notes column
    local metal_notes_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(metal_sheets);" | grep -c "notes" || true)
    if [ "$metal_notes_column" -gt 0 ]; then
        print_warning "Metal_sheets table has notes column (will be removed)"
    else
        print_info "Metal_sheets table has no notes column (already clean)"
    fi

    # Count existing data
    if [ "$notes_table_exists" -eq 1 ]; then
        local notes_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes;" 2>/dev/null || echo "0")
        print_info "Existing notes in database: $notes_count"
    fi
}

# Migrate notes table to new schema
migrate_notes_table() {
    print_header "Migrating notes table schema"

    # Check if notes table exists
    local notes_table_exists=$(sqlite3 "$DB_PATH" "SELECT name FROM sqlite_master WHERE type='table' AND name='notes';" | wc -l)

    if [ "$notes_table_exists" -eq 0 ]; then
        # Create new notes table with correct schema
        print_info "Creating new notes table with generic linking"
        sqlite3 "$DB_PATH" <<EOF
CREATE TABLE notes (
    id INTEGER NOT NULL,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked TEXT DEFAULT '',
    PRIMARY KEY("id" AUTOINCREMENT)
);
EOF
        print_success "Created new notes table"
    else
        # Check if we need to migrate existing notes table
        local has_linked=$(sqlite3 "$DB_PATH" "PRAGMA table_info(notes);" | grep -c "linked" || true)
        local has_linked_to_press=$(sqlite3 "$DB_PATH" "PRAGMA table_info(notes);" | grep -c "linked_to_press" || true)
        local has_linked_to_tool=$(sqlite3 "$DB_PATH" "PRAGMA table_info(notes);" | grep -c "linked_to_tool" || true)

        if [ "$has_linked" -eq 0 ]; then
            print_info "Adding linked column to notes table"
            sqlite3 "$DB_PATH" "ALTER TABLE notes ADD COLUMN linked TEXT;"
            print_success "Added linked column"

            # Convert NULL values to empty strings immediately after adding column
            print_info "Converting NULL linked values to empty strings"
            sqlite3 "$DB_PATH" "UPDATE notes SET linked = '' WHERE linked IS NULL;"
            print_success "Converted NULL values to empty strings"

            # Migrate data from old linking columns if they exist
            if [ "$has_linked_to_press" -gt 0 ] || [ "$has_linked_to_tool" -gt 0 ]; then
                print_info "Migrating existing linking data"

                # Migrate press links
                if [ "$has_linked_to_press" -gt 0 ]; then
                    sqlite3 "$DB_PATH" "UPDATE notes SET linked = 'press_' || linked_to_press WHERE linked_to_press IS NOT NULL;"
                    local migrated_press=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked LIKE 'press_%';" || echo "0")
                    print_success "Migrated $migrated_press press-linked notes"
                fi

                # Migrate tool links
                if [ "$has_linked_to_tool" -gt 0 ]; then
                    sqlite3 "$DB_PATH" "UPDATE notes SET linked = 'tool_' || linked_to_tool WHERE linked_to_tool IS NOT NULL AND linked = '';"
                    local migrated_tools=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked LIKE 'tool_%';" || echo "0")
                    print_success "Migrated $migrated_tools tool-linked notes"
                fi
            fi
        else
            print_success "Notes table already has linked column"

            # Even if column exists, ensure no NULL values remain
            print_info "Ensuring no NULL linked values exist"
            local null_count=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked IS NULL;" 2>/dev/null || echo "0")
            if [ "$null_count" -gt 0 ]; then
                print_info "Converting $null_count NULL linked values to empty strings"
                sqlite3 "$DB_PATH" "UPDATE notes SET linked = '' WHERE linked IS NULL;"
                print_success "Converted NULL values to empty strings"
            else
                print_success "No NULL linked values found"
            fi
        fi

        # Remove old linking columns if they exist
        if [ "$has_linked_to_press" -gt 0 ] || [ "$has_linked_to_tool" -gt 0 ]; then
            print_info "Removing old linking columns"
            # SQLite doesn't support DROP COLUMN easily, so we recreate the table
            sqlite3 "$DB_PATH" <<EOF
BEGIN TRANSACTION;

-- Create new table with correct schema
CREATE TABLE notes_new (
    id INTEGER NOT NULL,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked TEXT,
    PRIMARY KEY("id" AUTOINCREMENT)
);

-- Copy data from old table, ensuring linked is never NULL
INSERT INTO notes_new (id, level, content, created_at, linked)
SELECT id, level, content, created_at, COALESCE(linked, '') FROM notes;

-- Replace old table
DROP TABLE notes;
ALTER TABLE notes_new RENAME TO notes;

COMMIT;
EOF
            print_success "Removed old linking columns"
        fi
    fi
}

# Remove notes column from tools table
migrate_tools_table() {
    print_header "Migrating tools table schema"

    local tools_notes_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "notes" || true)

    if [ "$tools_notes_column" -gt 0 ]; then
        print_info "Removing notes column from tools table"

        # Get current tools table schema (without notes column)
        sqlite3 "$DB_PATH" <<EOF
BEGIN TRANSACTION;

-- Create new tools table without notes column
CREATE TABLE tools_new (
    id INTEGER NOT NULL,
    position TEXT NOT NULL,
    format BLOB NOT NULL,
    type TEXT NOT NULL,
    code TEXT NOT NULL,
    regenerating BOOLEAN NOT NULL DEFAULT 0,
    press INTEGER,
    PRIMARY KEY("id" AUTOINCREMENT)
);

-- Copy data (excluding notes column)
INSERT INTO tools_new (id, position, format, type, code, regenerating, press)
SELECT id, position, format, type, code, regenerating, press FROM tools;

-- Replace old table
DROP TABLE tools;
ALTER TABLE tools_new RENAME TO tools;

COMMIT;
EOF
        print_success "Removed notes column from tools table"
    else
        print_success "Tools table already clean (no notes column)"
    fi
}

# Remove notes column from metal_sheets table
migrate_metal_sheets_table() {
    print_header "Migrating metal_sheets table schema"

    local metal_notes_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(metal_sheets);" | grep -c "notes" || true)

    if [ "$metal_notes_column" -gt 0 ]; then
        print_info "Removing notes column from metal_sheets table"

        # Check if identifier column exists (from previous migration)
        local has_identifier=$(sqlite3 "$DB_PATH" "PRAGMA table_info(metal_sheets);" | grep -c "identifier" || true)

        sqlite3 "$DB_PATH" <<EOF
BEGIN TRANSACTION;

-- Create new metal_sheets table without notes column
CREATE TABLE metal_sheets_new (
    id INTEGER NOT NULL,
    tile_height REAL NOT NULL,
    value REAL NOT NULL,
    marke_height INTEGER NOT NULL,
    stf REAL NOT NULL,
    stf_max REAL NOT NULL,
    $([ "$has_identifier" -gt 0 ] && echo "identifier TEXT NOT NULL,")
    tool_id INTEGER NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY("id" AUTOINCREMENT),
    FOREIGN KEY("tool_id") REFERENCES "tools"("id") ON DELETE CASCADE
);

-- Copy data (excluding notes column)
INSERT INTO metal_sheets_new (id, tile_height, value, marke_height, stf, stf_max, $([ "$has_identifier" -gt 0 ] && echo "identifier,") tool_id, updated_at)
SELECT id, tile_height, value, marke_height, stf, stf_max, $([ "$has_identifier" -gt 0 ] && echo "identifier,") tool_id, updated_at FROM metal_sheets;

-- Replace old table
DROP TABLE metal_sheets;
ALTER TABLE metal_sheets_new RENAME TO metal_sheets;

COMMIT;
EOF
        print_success "Removed notes column from metal_sheets table"
    else
        print_success "Metal_sheets table already clean (no notes column)"
    fi
}

# Verify the migration
verify_migration() {
    print_header "Verifying migration"

    # Verify notes table structure
    local notes_linked_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(notes);" | grep -c "linked" || true)
    local notes_old_columns=$(sqlite3 "$DB_PATH" "PRAGMA table_info(notes);" | grep -E "(linked_to_press|linked_to_tool)" | wc -l)

    if [ "$notes_linked_column" -gt 0 ] && [ "$notes_old_columns" -eq 0 ]; then
        print_success "Notes table schema is correct"
    else
        print_error "Notes table migration failed"
        return 1
    fi

    # Verify tools table
    local tools_notes_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(tools);" | grep -c "notes" || true)
    if [ "$tools_notes_column" -eq 0 ]; then
        print_success "Tools table schema is correct"
    else
        print_error "Tools table migration failed"
        return 1
    fi

    # Verify metal_sheets table
    local metal_notes_column=$(sqlite3 "$DB_PATH" "PRAGMA table_info(metal_sheets);" | grep -c "notes" || true)
    if [ "$metal_notes_column" -eq 0 ]; then
        print_success "Metal_sheets table schema is correct"
    else
        print_error "Metal_sheets table migration failed"
        return 1
    fi

    # Count migrated data
    local total_notes=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes;" 2>/dev/null || echo "0")
    local linked_notes=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked != '';" 2>/dev/null || echo "0")
    local null_linked=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM notes WHERE linked IS NULL;" 2>/dev/null || echo "0")

    print_success "Total notes in database: $total_notes"
    print_success "Notes with linking information: $linked_notes"

    if [ "$null_linked" -gt 0 ]; then
        print_error "Found $null_linked notes with NULL linked values - this will cause scanning errors"
        return 1
    else
        print_success "No NULL linked values found"
    fi

    if [ "$linked_notes" -gt 0 ]; then
        # Show distribution of linked notes
        print_info "Linked notes distribution:"
        sqlite3 "$DB_PATH" "SELECT
            CASE
                WHEN linked LIKE 'tool_%' THEN 'Tools'
                WHEN linked LIKE 'press_%' THEN 'Presses'
                WHEN linked = '' OR linked IS NULL THEN 'Unlinked'
                ELSE 'Other'
            END as link_type,
            COUNT(*) as count
        FROM notes
        GROUP BY link_type;" | while IFS='|' read -r type count; do
            echo "  - $type: $count"
        done
    fi
}

# Show post-migration information
show_post_migration_info() {
    print_header "Post-Migration Information"
    echo "The database has been successfully migrated to the new notes system:"
    echo ""
    echo "Changes made:"
    echo "• Notes table: Added 'linked' column for generic entity linking"
    echo "• Tools table: Removed 'notes' column (linking is now done via notes table)"
    echo "• Metal_sheets table: Removed 'notes' column"
    echo ""
    echo "Notes linking format:"
    echo "• Tool notes: 'tool_123' (where 123 is the tool ID)"
    echo "• Press notes: 'press_5' (where 5 is the press number)"
    echo "• Other entities: 'type_id' (extensible for future use)"
    echo ""
    echo "The new system provides:"
    echo "• Semantic correctness (press notes stay with press)"
    echo "• Generic linking to any entity type"
    echo "• Simplified database schema and operations"
    echo "• Better maintainability and extensibility"
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                Notes System Refactor Migration              ║"
    echo "║                                                              ║"
    echo "║ This script migrates the database to the new generic notes  ║"
    echo "║ linking system, removing tool-specific linking complexity.  ║"
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
    analyze_current_schema

    echo ""
    print_warning "This migration will modify your database schema."
    print_warning "A backup has been created, but please ensure you have additional backups."
    read -p "Do you want to continue with the migration? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Migration cancelled by user"
        exit 0
    fi

    migrate_notes_table
    migrate_tools_table
    migrate_metal_sheets_table
    verify_migration
    show_post_migration_info

    print_header "Migration Completed Successfully"
    print_success "Database has been migrated to the new generic notes system"
    print_warning "Remember to restart your application to pick up the database changes"
    print_info "The old migration script (migrate-add-identifier.sh) is no longer needed"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "This script migrates the database to the new generic notes linking system."
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  --version      Show version information"
        echo ""
        echo "The script will:"
        echo "1. Create a backup of your database"
        echo "2. Add/update the notes table with generic 'linked' column"
        echo "3. Remove notes columns from tools and metal_sheets tables"
        echo "4. Migrate existing linking data to the new format"
        echo "5. Verify the migration was successful"
        echo ""
        echo "Run this script from the project root directory."
        echo "Make sure your application is stopped before running the migration."
        exit 0
        ;;
    --version|-v)
        echo "Notes System Refactor Migration v1.0"
        echo "Migrates to generic notes linking system"
        exit 0
        ;;
    --dry-run)
        echo "Dry-run mode: analyzing current schema without making changes"
        if [ ! -f "go.mod" ] || [ ! -d "scripts" ]; then
            print_error "Please run this script from the project root directory"
            exit 1
        fi
        cd scripts
        check_database
        analyze_current_schema
        exit 0
        ;;
    *)
        # No arguments or unknown arguments, proceed with migration
        main "$@"
        ;;
esac
