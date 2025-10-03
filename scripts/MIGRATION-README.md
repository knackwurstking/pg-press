# Database Migration Scripts

This directory contains database migration scripts for the pg-press application.

## Available Migrations

### migrate-add-identifier.sh

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

## Future Migrations

When adding new migration scripts:

1. Use descriptive names: `migrate-description-of-change.sh`
2. Include backup functionality
3. Add verification steps
4. Document in this README
5. Make scripts executable: `chmod +x script-name.sh`
