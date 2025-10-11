# Database Migration Scripts

This directory contains database migration scripts for the PG Press application.

## Overview

Migration scripts are used to safely update the database schema and data when new features are added or existing structures need to be modified.

## Current Migrations

### `migrate-tools-dead-status.sh`

**Purpose**: Adds support for marking tools as "dead" instead of deleting them, solving foreign key constraint issues.

**Changes Made**:

- Adds `is_dead INTEGER NOT NULL DEFAULT 0` column to the `tools` table
- All existing tools are automatically marked as alive (`is_dead = 0`)
- Preserves all historical data and relationships

**Why This Migration**:

- Solves FOREIGN KEY constraint errors when trying to delete tools
- Preserves historical data (press cycles, tool regenerations, etc.)
- Allows for audit trails and data recovery
- Provides better user experience with reversible "soft delete"

## Usage

### Running a Migration

```bash
# Run with default database location (./pg-press.db)
./scripts/migrate-tools-dead-status.sh

# Run with custom database path
./scripts/migrate-tools-dead-status.sh -d /path/to/your/database.db

# Using environment variable
DB_PATH=/custom/path/pg-press.db ./scripts/migrate-tools-dead-status.sh

# Show help
./scripts/migrate-tools-dead-status.sh --help
```

### Prerequisites

1. **SQLite3** must be installed and accessible via command line
2. **Database file** must exist before running the migration
3. **Backup space** should be available (scripts create automatic backups)
4. **Write permissions** to the database file and backup directory

### Safety Features

- **Automatic Backup**: Creates timestamped backup before any changes
- **Migration Detection**: Skips migration if already applied
- **Verification**: Confirms changes were applied correctly
- **Rollback Information**: Backup files allow manual rollback if needed

### Test Script

Before running the migration on your production database, you can test it safely:

```bash
# Run complete migration test with temporary database
./scripts/test-migration.sh

# Keep test database for inspection
./scripts/test-migration.sh --keep-db

# Show help
./scripts/test-migration.sh --help
```

The test script:

- ✅ Creates temporary database with realistic schema
- ✅ Populates test data including foreign key relationships
- ✅ Runs the complete migration process
- ✅ Verifies all functionality works correctly
- ✅ Tests both old and new query patterns
- ✅ Automatically cleans up (unless `--keep-db` is used)

### Verification Script

Use the verification script to check your database schema and migration status:

```bash
# Full verification of tools table schema
./scripts/verify-tools-schema.sh

# Verify custom database
./scripts/verify-tools-schema.sh -d /path/to/custom.db

# Show only statistics (quiet mode)
./scripts/verify-tools-schema.sh --stats-only

# Show help
./scripts/verify-tools-schema.sh --help
```

The verification script checks:

- ✅ Database accessibility and table existence
- ✅ Complete schema validation
- ✅ `is_dead` column presence and configuration
- ✅ Data statistics (alive vs dead tools)
- ✅ Basic query functionality

## After Migration

Once the tools dead status migration is complete, you can:

### CLI Commands

```bash
# List all tools (shows alive/dead status)
./pg-press tools list

# List only dead tools
./pg-press tools list-dead

# Mark a tool as dead (instead of deleting)
./pg-press tools mark-dead <tool-id>
```

### API Changes

- New endpoint: `PATCH /htmx/tools/mark-dead?id=<tool-id>`
- Tools list endpoints now exclude dead tools by default
- Dead tools are preserved with all historical relationships

### Code Changes

```go
// New service methods available:
err := registry.Tools.MarkAsDead(toolID, user)      // Mark as dead
err := registry.Tools.ReviveTool(toolID, user)      // Restore dead tool
tools, err := registry.Tools.ListActiveTools()      // Only alive tools
deadTools, err := registry.Tools.ListDeadTools()    // Only dead tools
```

## Quick Migration

For most users, the simplest approach is to use the quick migration helper:

```bash
# Test the migration safely first (recommended)
./scripts/migrate.sh --test

# Run migration on default database
./scripts/migrate.sh

# Run migration on specific database
./scripts/migrate.sh /path/to/your/database.db

# Just verify existing database
./scripts/migrate.sh --verify-only
```

The quick helper automatically:

- ✅ Checks database existence and accessibility
- ✅ Runs the complete migration process
- ✅ Verifies migration success
- ✅ Shows final statistics and usage instructions

## Migration File Structure

```
scripts/
├── README.md                           # This file
├── migrate.sh                         # Quick migration helper (recommended)
├── migrate-tools-dead-status.sh       # Full migration script
├── verify-tools-schema.sh             # Schema verification script
├── test-migration.sh                  # Migration testing script
└── backups/                           # Auto-created backup directory
    └── pg-press_before_*_TIMESTAMP.db # Automatic backups
```

## Troubleshooting

### Common Issues

**Database not found**:

```bash
# Error: Database file not found: ./pg-press.db
# Solution: Check database path or specify correct path
./scripts/migrate-tools-dead-status.sh -d /correct/path/to/pg-press.db
```

**Permission denied**:

```bash
# Error: Permission denied
# Solution: Ensure write access to database and backup directory
chmod +w pg-press.db
mkdir -p scripts/backups && chmod +w scripts/backups
```

**Migration already applied**:

```bash
# Warning: Migration already applied
# This is safe - the script detects existing migrations and skips them
```

### Manual Rollback

If you need to rollback a migration:

1. Stop the application
2. Replace current database with backup:
   ```bash
   cp scripts/backups/pg-press_before_tools_dead_migration_TIMESTAMP.db pg-press.db
   ```
3. Restart the application

> ⚠️ **Warning**: Manual rollback will lose any data created after the migration!

## Best Practices

1. **Always backup** before running migrations (automatic, but verify)
2. **Test in development** before running on production using `./scripts/test-migration.sh`
3. **Stop the application** during migration to prevent data corruption
4. **Verify results** after migration completes using the verification script
5. **Keep backups** for a reasonable period after migration

## Development

When creating new migrations:

1. Follow the existing script structure
2. Include comprehensive error handling
3. Add automatic backup creation
4. Implement migration detection
5. Provide clear success/failure feedback
6. Update this README with new migration information

## Summary

The tools dead status migration solves the **FOREIGN KEY constraint failed** error by implementing a soft delete system:

### Before Migration

- ❌ Deleting tools fails due to foreign key constraints
- ❌ Risk of data loss when trying to remove tools
- ❌ No historical preservation of tool relationships

### After Migration

- ✅ Tools can be marked as "dead" instead of deleted
- ✅ All historical data (press cycles, regenerations) preserved
- ✅ Reversible operations (tools can be "revived")
- ✅ Clean separation of active vs. dead tools in queries
- ✅ Audit trail maintained through feed entries

### Migration Workflow

1. **Test**: `./scripts/migrate.sh --test`
2. **Backup**: Automatic (or manual backup recommended)
3. **Migrate**: `./scripts/migrate.sh`
4. **Verify**: `./scripts/verify-tools-schema.sh`
5. **Use**: `./pg-press tools mark-dead <tool-id>`

---

For questions or issues with migrations, check the application logs or create an issue in the project repository.
