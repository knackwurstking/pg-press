# PG-Press Migration Scripts

This directory contains scripts to migrate from the old `mods` column-based modification tracking system to the new centralized `modifications` table system in pg-press.

## Overview

The migration process transforms data from individual `mods` JSON columns in `trouble_reports`, `metal_sheets`, and `tools` tables into a centralized `modifications` table. This provides better performance, consistency, and maintainability.

### What the Migration Does

1. **Adds missing schema**: Creates `mods` columns if they don't exist and sets up the `modifications` table
2. **Migrates data**: Converts JSON data from `mods` columns to structured records in the `modifications` table
3. **Preserves history**: Maintains original timestamps, user information, and modification details
4. **Enables verification**: Provides tools to verify migration accuracy
5. **Cleans up**: Optionally removes old `mods` columns after successful migration

## Files in This Directory

| File                                | Purpose                                                  |
| ----------------------------------- | -------------------------------------------------------- |
| `migrate_mods_to_modifications.sql` | SQL script to set up database schema                     |
| `migrate_mods.go`                   | Standalone Go migration tool                             |
| `run_migration.sh`                  | Shell script orchestrator for complete migration process |
| `README.md`                         | This documentation                                       |

## Prerequisites

### Before You Start

1. **Backup your database!** This cannot be stressed enough - create a full backup before starting
2. **Stop your application** to prevent conflicts during migration
3. **Verify database integrity** using SQLite's built-in tools
4. **Have sufficient disk space** for the migration and backups

### System Requirements

- SQLite 3.x
- Go 1.19+ (if using the standalone Go script)
- Bash shell (for the orchestrator script)
- `pgpress` binary (if using built-in migration commands)

## Quick Start

For most users, the shell script orchestrator provides the easiest migration path:

```bash
# Run complete migration with dry-run first to see what will happen
./scripts/run_migration.sh --dry-run full

# Run the actual migration
./scripts/run_migration.sh full

# If you need to specify a custom database path
./scripts/run_migration.sh --db /path/to/your/data.db full
```

## Detailed Migration Process

### Step 1: Preparation

```bash
# 1. Create a backup
cp data.db data.db.backup.$(date +%Y%m%d_%H%M%S)

# 2. Check current status
./scripts/run_migration.sh status

# 3. Test database connection
pgpress migration test-db
```

### Step 2: Schema Setup

```bash
# Option A: Using the orchestrator script
./scripts/run_migration.sh setup

# Option B: Using SQL directly
sqlite3 data.db < scripts/migrate_mods_to_modifications.sql

# Option C: Using pgpress (if the modifications table doesn't exist)
pgpress migration status
```

### Step 3: Data Migration

```bash
# Option A: Using the orchestrator script
./scripts/run_migration.sh migrate

# Option B: Using pgpress built-in migration
pgpress migration run

# Option C: Using standalone Go script
go run scripts/migrate_mods.go --db data.db --action migrate
```

### Step 4: Verification

```bash
# Verify migration completed successfully
./scripts/run_migration.sh verify

# Or using pgpress
pgpress migration verify

# Check statistics
pgpress migration stats
```

### Step 5: Testing

**Important**: Thoroughly test your application before proceeding to cleanup!

1. Start your application
2. Test all functionality that involves modifications
3. Verify data integrity in your application
4. Run your test suite if available

### Step 6: Cleanup (Optional)

⚠️ **WARNING: This step is destructive and cannot be undone!**

```bash
# Clean up old mods columns
./scripts/run_migration.sh cleanup

# Or using pgpress with force flag
pgpress migration cleanup --force
```

## Script Usage Details

### Shell Script Orchestrator (`run_migration.sh`)

The main orchestrator script that handles the complete migration process.

```bash
Usage: ./run_migration.sh [OPTIONS] ACTION

ACTIONS:
    setup       - Set up database schema
    migrate     - Migrate data from mods columns
    verify      - Verify migration integrity
    cleanup     - Remove old mods columns (DESTRUCTIVE!)
    status      - Show current migration status
    full        - Run complete migration process

OPTIONS:
    -d, --db PATH       Database path (default: ./data.db)
    -f, --force         Force operation without confirmation
    -n, --dry-run       Show what would be done without making changes
    -v, --verbose       Verbose output
    -g, --go-script     Use standalone Go script instead of pgpress commands
    -h, --help          Show help
```

#### Examples

```bash
# Dry run to see what would happen
./scripts/run_migration.sh --dry-run full

# Full migration with custom database
./scripts/run_migration.sh --db /path/to/data.db full

# Force cleanup without confirmation (dangerous!)
./scripts/run_migration.sh --force cleanup

# Use standalone Go script instead of pgpress
./scripts/run_migration.sh --go-script migrate
```

### Standalone Go Script (`migrate_mods.go`)

Direct Go implementation for environments where pgpress isn't available.

```bash
Usage: go run migrate_mods.go [OPTIONS]

OPTIONS:
    -db string          Path to SQLite database (default "./data.db")
    -action string      Action: migrate, verify, cleanup, status (default "migrate")
    -force              Force operation without confirmation
    -dry-run            Show what would be done without making changes
    -v                  Verbose output
```

#### Examples

```bash
# Run migration
go run scripts/migrate_mods.go --db data.db --action migrate

# Verify with verbose output
go run scripts/migrate_mods.go --db data.db --action verify -v

# Dry run cleanup
go run scripts/migrate_mods.go --db data.db --action cleanup --dry-run
```

### SQL Schema Script (`migrate_mods_to_modifications.sql`)

Pure SQL script for manual database setup.

```bash
# Apply schema changes
sqlite3 data.db < scripts/migrate_mods_to_modifications.sql

# Check migration views
sqlite3 data.db "SELECT * FROM migration_status;"
sqlite3 data.db "SELECT * FROM modification_stats;"
```

## Migration Workflow Recommendations

### For Production Systems

1. **Maintenance Window**: Schedule migration during low-traffic periods
2. **Staged Approach**: Test on a copy of production data first
3. **Rollback Plan**: Keep backups and have a rollback procedure ready
4. **Monitoring**: Monitor system performance after migration
5. **Gradual Cleanup**: Wait several days before running cleanup

### For Development Systems

1. **Test First**: Run migrations on development data
2. **Version Control**: Commit any schema changes
3. **Team Notification**: Inform team members about the migration
4. **Documentation**: Update application documentation

## Troubleshooting

### Common Issues

#### "Column already exists" errors

```bash
# This is usually harmless during schema setup
# The scripts handle this gracefully
```

#### "Failed to drop column" during cleanup

```bash
# Older SQLite versions don't support DROP COLUMN
# Consider upgrading SQLite or manually recreating tables
```

#### Migration verification fails

```bash
# Check for data inconsistencies
sqlite3 data.db "SELECT * FROM migration_status;"

# Re-run migration if needed
./scripts/run_migration.sh migrate
```

#### Database locked errors

```bash
# Stop all applications using the database
# Check for zombie processes
ps aux | grep pgpress

# Force unlock (use with caution)
sqlite3 data.db "BEGIN IMMEDIATE; ROLLBACK;"
```

### Recovery Procedures

#### Restore from Backup

```bash
# If migration fails, restore from backup
cp data.db.backup.YYYYMMDD_HHMMSS data.db
```

#### Partial Migration Recovery

```bash
# Check what was migrated
./scripts/run_migration.sh status

# Continue from where it left off
./scripts/run_migration.sh migrate
```

#### Manual Cleanup

```bash
# If automatic cleanup fails, manually drop columns
sqlite3 data.db "
  -- Create new tables without mods columns
  CREATE TABLE trouble_reports_new AS
  SELECT id, title, content, linked_attachments
  FROM trouble_reports;

  DROP TABLE trouble_reports;
  ALTER TABLE trouble_reports_new RENAME TO trouble_reports;
"
```

## Performance Considerations

### During Migration

- Migration time depends on the amount of modification data
- Expect 1000-10000 records per second on modern hardware
- Monitor disk space as the process creates additional data

### After Migration

- The new system provides better query performance
- Indexes are automatically created for optimal access patterns
- Consider running `VACUUM` after cleanup to reclaim space

## Validation Queries

Use these queries to validate the migration:

```sql
-- Check migration status
SELECT * FROM migration_status;

-- Compare old vs new counts
SELECT
  'trouble_reports' as table_name,
  (SELECT COUNT(*) FROM trouble_reports WHERE mods IS NOT NULL AND mods != '[]') as old_count,
  (SELECT COUNT(DISTINCT entity_id) FROM modifications WHERE entity_type = 'trouble_reports') as new_count;

-- View recent modifications
SELECT * FROM modifications ORDER BY created_at DESC LIMIT 10;

-- Check for orphaned modifications (should return 0)
SELECT COUNT(*) FROM modifications m
WHERE NOT EXISTS (
  SELECT 1 FROM trouble_reports t WHERE t.id = m.entity_id AND m.entity_type = 'trouble_reports'
);
```

## Support and Reporting Issues

If you encounter issues:

1. **Check logs**: Enable verbose mode for detailed output
2. **Verify prerequisites**: Ensure all requirements are met
3. **Review documentation**: Check this README and inline help
4. **Create minimal reproduction**: Isolate the problem
5. **Report issues**: Include error messages, system info, and steps to reproduce

## Security Notes

- **Backup Security**: Store backups securely and encrypt if necessary
- **Access Control**: Ensure migration scripts have appropriate permissions
- **Data Integrity**: The migration preserves all original data and metadata
- **Audit Trail**: All modifications maintain original timestamps and user information

## Version Compatibility

This migration system is designed for:

- pg-press v0.9.x to v1.0.x
- SQLite 3.8.0 and newer
- Go 1.19 and newer

For older versions, manual migration may be required.
