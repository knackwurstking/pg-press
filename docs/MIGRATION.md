# Database Migration Guide

This guide explains how to migrate from the old modification system to the new centralized modification service in PG Press.

## Overview

The migration system helps transition from the legacy column-based modification tracking (using `mods` JSON columns) to a centralized modification service that provides:

- Centralized modification tracking across all entities
- Better performance and queryability
- Consistent modification history
- Enhanced audit trails
- Easier maintenance and debugging

## Prerequisites

- **Backup your database** before running any migration commands
- Ensure you have sufficient disk space for the migration process
- Stop any running PG Press services during migration
- Have database access permissions

## Migration Workflow

### 1. Check Migration Status

First, check if your database needs migration:

```bash
pgpress migration status
```

This will show:

- Whether the modification table exists
- Total modifications in the new system
- If old mods still exist
- Whether migration is recommended

### 2. Run the Migration

Execute the migration process:

```bash
# Interactive migration (recommended)
pgpress migration run

# Force migration without confirmation
pgpress migration run --force
```

The migration will:

- Process all entities with old modification data
- Convert old `mods` JSON to new modification records
- Preserve original timestamps and user information
- Display progress and statistics

**Expected output:**

```
=== Starting Migration Process ===
Starting migration...
Migrated 45 trouble reports with 127 total mods
Migrated 23 metal sheets
Migrated 67 tools
=== Migration Complete ===
Duration: 2.3s
Total modifications migrated: 234
✅ Migration completed successfully with no errors!
```

### 3. Verify Migration

Validate that the migration was successful:

```bash
pgpress migration verify
```

This compares old and new data counts to ensure accuracy:

```
=== Verifying Migration ===
Trouble Reports - Old: 45, New: 127, Match: ✓
Metal Sheets - Old: 23, New: 45, Match: ✓
Tools - Old: 67, New: 89, Match: ✓
✅ Verification successful!
```

### 4. Review Statistics

Check the migration statistics:

```bash
pgpress migration stats

# For JSON output (planned feature)
pgpress migration stats --output json
```

### 5. Test Your Application

**Important:** Thoroughly test your application with the new modification system:

- Verify all modification-related features work correctly
- Check that modification history displays properly
- Test creating new modifications
- Validate user permissions and access controls
- Ensure performance is acceptable

### 6. Export Migration Data (Optional)

Export the migration data for backup or analysis:

```bash
# Export all modifications
pgpress migration export

# Export specific entity type
pgpress migration export --entity trouble_reports --output tr_mods.json

# Export to custom location
pgpress migration export --output /backup/migration_data.json
```

### 7. Cleanup Old System (Optional)

**⚠️ WARNING: This step is DESTRUCTIVE and cannot be undone!**

After thorough testing, you can remove the old `mods` columns:

```bash
# Interactive cleanup (recommended)
pgpress migration cleanup

# Force cleanup without confirmation (dangerous)
pgpress migration cleanup --force
```

## Command Reference

### `pgpress migration status`

Shows current migration status and recommendations.

**Options:**

- `-d, --db <path>` - Custom database path

### `pgpress migration run`

Executes the migration from old to new system.

**Options:**

- `-d, --db <path>` - Custom database path
- `-f, --force` - Skip interactive confirmation

### `pgpress migration verify`

Verifies migration integrity by comparing data counts.

**Options:**

- `-d, --db <path>` - Custom database path

### `pgpress migration stats`

Displays modification system statistics.

**Options:**

- `-d, --db <path>` - Custom database path
- `-o, --output <format>` - Output format (text|json)

### `pgpress migration export`

Exports migration data to JSON file.

**Options:**

- `-d, --db <path>` - Custom database path
- `-o, --output <file>` - Output file path (default: migration_export.json)
- `-e, --entity <type>` - Entity type filter (trouble_reports|metal_sheets|tools|all)

### `pgpress migration cleanup`

Removes old mod columns (destructive operation).

**Options:**

- `-d, --db <path>` - Custom database path
- `-f, --force` - Skip safety checks and confirmation

### `pgpress migration help`

Shows detailed help and usage information.

## Troubleshooting

### Migration Fails with Errors

1. **Check database permissions**: Ensure read/write access
2. **Review error logs**: Check application logs for detailed error messages
3. **Verify database integrity**: Run database consistency checks
4. **Restore from backup**: If needed, restore and retry

### Verification Fails

1. **Don't panic**: Migration can be run multiple times safely
2. **Check specific entity types**: Use export command to investigate
3. **Manual verification**: Query database directly to compare counts
4. **Re-run migration**: The process is idempotent and safe to repeat

### Performance Issues

1. **Database size**: Large databases may take longer to migrate
2. **Index optimization**: Consider rebuilding indexes after migration
3. **Incremental approach**: For very large datasets, consider custom migration scripts

### Rollback Procedure

If you need to rollback after cleanup:

1. **Restore from backup**: This is the only way to recover old `mods` columns
2. **Stop application**: Prevent data corruption
3. **Verify backup**: Ensure backup is complete and accessible
4. **Coordinate downtime**: Plan for extended downtime during restore

## Best Practices

1. **Always backup first**: Cannot be emphasized enough
2. **Test in staging**: Run migration on staging environment first
3. **Plan for downtime**: Schedule migration during low-usage periods
4. **Monitor performance**: Watch database performance after migration
5. **Keep backups**: Maintain backups even after successful migration
6. **Document changes**: Record migration details for your team

## File Locations

- Migration logs: Check application log files
- Export files: Current working directory (unless specified otherwise)
- Database: Default location or specified with `--db` flag

## Support

For issues or questions:

1. Check application logs for detailed error messages
2. Review this documentation and command help
3. Verify database permissions and connectivity
4. Ensure sufficient disk space and resources

## Migration Data Structure

The new modification system stores data in the following format:

```json
{
    "id": 123,
    "user_id": 456,
    "entity_type": "trouble_reports",
    "entity_id": 789,
    "data": {
        "action": "create",
        "comment": "Initial creation",
        "changes": {
            "title": "New trouble report",
            "content": "Description of the issue"
        }
    },
    "created_at": "2024-01-15 10:30:00"
}
```

This provides a consistent structure across all entity types while preserving the specific data for each modification.
