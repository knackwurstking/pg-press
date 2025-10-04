# Database Migration Scripts

This directory contains database migration scripts for the pg-press application.

## Available Migrations

### migrate-notes-system-refactor.sh

**Purpose**: Migrates the database to the new generic notes linking system, removing complex tool-specific linking.

**What it does**:

- Updates the notes table with a generic 'linked' column for flexible entity linking
- Removes notes columns from tools and metal_sheets tables
- Migrates existing linking data from old format to new format
- Creates a backup of your database before making changes
- Validates the migration was successful

**Prerequisites**:

- SQLite3 must be installed on your system
- The pg-press database must exist (`pg-press.db`)
- You must run the script from the project root directory
- **Stop your application before running this migration**

**Usage**:

```bash
# From the project root directory
./scripts/migrate-notes-system-refactor.sh

# For help
./scripts/migrate-notes-system-refactor.sh --help

# Dry-run to analyze current schema
./scripts/migrate-notes-system-refactor.sh --dry-run
```

**What gets changed**:

Before migration:

```sql
-- Old notes table (if exists)
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked_to_press INTEGER,
    linked_to_tool INTEGER
);

-- Tools table with notes column
CREATE TABLE tools (
    -- other fields
    notes BLOB NOT NULL
);

-- Metal sheets with notes column
CREATE TABLE metal_sheets (
    -- other fields
    notes BLOB NOT NULL
);
```

After migration:

```sql
-- New notes table with generic linking
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked TEXT  -- Generic entity reference: "tool_123", "press_5", etc.
);

-- Tools table without notes column
CREATE TABLE tools (
    -- other fields, no notes column
);

-- Metal sheets without notes column
CREATE TABLE metal_sheets (
    -- other fields, no notes column
);
```

**New Notes Linking Format**:

- `"tool_123"` - Links note to tool with ID 123
- `"press_5"` - Links note to press number 5
- `"machine_42"` - Can link to any future entity type
- `""` or `NULL` - Unlinked notes

**Benefits of New System**:

- **Semantic Correctness**: Press notes stay with press, not tools
- **Maximum Flexibility**: Can link to any entity type without schema changes
- **Simplified Operations**: No complex relationship management
- **Better Performance**: Direct queries without JOINs
- **Easier Maintenance**: Single linking field to manage

**Post-migration**:
After running the migration:

1. Restart your application to pick up the database schema changes
2. Notes will use the new generic linking system
3. Press notes will stay with presses regardless of tool changes
4. Tool notes will stay with specific tools
5. The system is ready for linking to any new entity types

**Backup and Recovery**:

- The script automatically creates a timestamped backup in `scripts/backups/`
- Backup format: `pg-press_before_notes_refactor_YYYYMMDD_HHMMSS.db`
- To restore from backup: `cp backups/backup_file.db ../pg-press.db`

**Troubleshooting NULL Values**:

If you encounter an error like "converting NULL to string is unsupported" when accessing the /notes page, it means there are NULL values in the linked column that need to be fixed:

```bash
# Quick manual fix using SQLite
sqlite3 pg-press.db "UPDATE notes SET linked = '' WHERE linked IS NULL;"

# Or use the automated fix script
./scripts/fix-null-linked-values.sh
```

**Manual SQL Fix**:

```sql
-- Convert NULL linked values to empty strings
UPDATE notes SET linked = '' WHERE linked IS NULL;

-- Verify the fix
SELECT COUNT(*) as null_count FROM notes WHERE linked IS NULL;
-- Should return 0

-- Check distribution
SELECT
  CASE
    WHEN linked = '' THEN 'Unlinked'
    WHEN linked LIKE 'tool_%' THEN 'Tool Links'
    WHEN linked LIKE 'press_%' THEN 'Press Links'
    ELSE 'Other'
  END as type,
  COUNT(*) as count
FROM notes GROUP BY type;
```

---

### migrate-add-identifier.sh (DEPRECATED)

**Purpose**: Adds an `identifier` field to the `metal_sheets` table to distinguish between SACMI and SITI machine types.

**What it does**:

- Adds a new `identifier` column to the `metal_sheets` table
- Sets the default value to "SACMI" for all existing records
- Creates a backup of your database before making changes
- Validates the migration was successful

**Prerequisites**:

- SQLite3 must be installed on your system
- The pg-press database must exist (`pg-press.db`)
- You must run the script from the project root directory

**Usage**:

```bash
# From the project root directory
./scripts/migrate-add-identifier.sh

# For help
./scripts/migrate-add-identifier.sh --help
```

**What gets changed**:

Before migration:

```sql
CREATE TABLE metal_sheets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tile_height REAL NOT NULL,
    value REAL NOT NULL,
    marke_height INTEGER NOT NULL,
    stf REAL NOT NULL,
    stf_max REAL NOT NULL,
    tool_id INTEGER NOT NULL,
    notes BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

After migration:

```sql
CREATE TABLE metal_sheets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tile_height REAL NOT NULL,
    value REAL NOT NULL,
    marke_height INTEGER NOT NULL,
    stf REAL NOT NULL,
    stf_max REAL NOT NULL,
    identifier TEXT NOT NULL DEFAULT 'SACMI',  -- NEW FIELD
    tool_id INTEGER NOT NULL,
    notes BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Valid identifier values**:

- `"SACMI"` - For SACMI type machines
- `"SITI"` - For SITI type machines

**Post-migration**:
After running the migration:

1. Restart your application to pick up the database schema changes
2. All existing metal sheets will have identifier set to "SACMI"
3. New metal sheets created through the application will default to "SACMI"
4. You can update individual records through the application UI or directly in the database

**Press Filtering Feature**:
The identifier field enables automatic filtering of metal sheets on press pages:

- **Press 0 and Press 5**: Show only SACMI metal sheets
- **All other presses** (2, 3, 4): Show only SITI metal sheets

This ensures operators only see metal sheets that are compatible with their specific press/machine combination. The filtering happens automatically when viewing press pages - no additional configuration is needed.

**Manual updates**:
If you need to update specific records to SITI type:

```bash
# Update a specific metal sheet to SITI type
sqlite3 pg-press.db "UPDATE metal_sheets SET identifier='SITI' WHERE id=123;"

# Update multiple sheets based on some criteria
sqlite3 pg-press.db "UPDATE metal_sheets SET identifier='SITI' WHERE tool_id IN (1,2,3);"

# Check current identifier distribution
sqlite3 pg-press.db "SELECT identifier, COUNT(*) FROM metal_sheets GROUP BY identifier;"
```

**Backup and Recovery**:

- The script automatically creates a timestamped backup in `scripts/backups/`
- Backup format: `pg-press_before_identifier_migration_YYYYMMDD_HHMMSS.db`
- To restore from backup: `cp backups/backup_file.db ../pg-press.db`

**Troubleshooting**:

If the migration fails:

1. Check that you're running from the project root directory
2. Ensure SQLite3 is installed: `sqlite3 --version`
3. Verify database file exists: `ls -la pg-press.db`
4. Check database permissions are correct

If you need to rollback:

1. Stop the application
2. Restore from the backup: `cp scripts/backups/backup_file.db pg-press.db`
3. Restart the application

**Development Notes**:

- This migration is designed for development mode and uses simple ALTER TABLE statements
- For production environments, consider more robust migration strategies
- The migration script includes safety checks and creates backups automatically
- The identifier field is used by the application to determine STF_MAX calculation methods and automatic press filtering

## General Migration Guidelines

1. **Always backup your database** before running any migration
2. **Test migrations** on a copy of your data first
3. **Run from the project root** directory unless otherwise specified
4. **Stop the application** before running migrations
5. **Restart the application** after successful migrations

---

### cleanup-old-migrations.sh

**Purpose**: Removes deprecated migration scripts that are no longer needed after system refactoring.

**What it does**:

- Identifies and removes old migration files that have been superseded
- Creates backups of removed files before deletion
- Provides information about why migrations are deprecated
- Helps keep the scripts directory clean and up-to-date

**Prerequisites**:

- Must be run from the `scripts` directory
- Will create backups in `scripts/backups/` before removal

**Usage**:

```bash
# From the scripts directory
cd scripts
./cleanup-old-migrations.sh

# List files that would be removed (dry run)
./cleanup-old-migrations.sh --list

# Force cleanup without confirmation
./cleanup-old-migrations.sh --force

# For help
./cleanup-old-migrations.sh --help
```

**Files removed by this script**:

- `migrate-add-identifier.sh` - Superseded by `migrate-notes-system-refactor.sh`
- `test-migration.sh` - No longer relevant testing script

**When to use**:

- After successfully running the new notes system migration
- When cleaning up development environments
- Before deploying to ensure only current migrations are present

**Safety features**:

- Always creates backups before removing files
- Requires confirmation before proceeding
- Provides detailed information about why files are deprecated
- Can be run multiple times safely (idempotent)

## Future Migrations

When adding new migration scripts:

1. Use descriptive names: `migrate-description-of-change.sh`
2. Include backup functionality
3. Add verification steps
4. Document in this README
5. Make scripts executable: `chmod +x script-name.sh`
6. Update cleanup script if superseding old migrations
