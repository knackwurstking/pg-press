# PG-Press Migration Tool

A standalone Go-based tool for migrating from the old `mods` column system to the new centralized `modifications` table in pg-press.

## 🚀 Quick Start

**Most users should run:**

```bash
# Complete migration (recommended)
go run migrate_mods.go

# Or see what would happen first
go run migrate_mods.go -dry-run
```

That's it! The tool automatically handles backup, schema setup, data migration, and verification.

## 📋 Overview

The migration tool transforms data from individual `mods` JSON columns in `trouble_reports`, `metal_sheets`, and `tools` tables into a centralized `modifications` table, providing:

- ✅ **Better performance** with proper indexing
- ✅ **Consistency** across all entity types
- ✅ **Foreign key relationships** for data integrity
- ✅ **Easier reporting** and analytics

## 📁 What's in This Directory

| File              | Purpose                                             |
| ----------------- | --------------------------------------------------- |
| `migrate_mods.go` | **Main migration tool** (standalone Go application) |
| `Makefile`        | Convenient build and automation targets             |
| `README.md`       | This documentation                                  |

## 🔧 Installation & Usage

### Prerequisites

- **Go 1.19+** installed
- **SQLite database** file accessible
- **Backup** of your database (automatically created by tool)

### Basic Usage

```bash
# Show help
go run migrate_mods.go -help

# Complete migration (default action)
go run migrate_mods.go -db ./data.db

# Check status
go run migrate_mods.go -action status -db ./data.db

# Dry run (preview changes)
go run migrate_mods.go -dry-run -db ./data.db
```

### Available Actions

| Action    | Description                                                 |
| --------- | ----------------------------------------------------------- |
| `full`    | **Complete migration** (setup + migrate + verify) [DEFAULT] |
| `setup`   | Set up database schema (add columns, create tables)         |
| `migrate` | Migrate data from old mods columns                          |
| `verify`  | Verify migration integrity                                  |
| `cleanup` | Remove old mods columns (DESTRUCTIVE!)                      |
| `status`  | Show current migration status                               |

### Command Line Options

```
  -action string    Action to perform (default "full")
  -backup          Create backup before migration (default true)
  -db string       Path to SQLite database (default "./data.db")
  -dry-run         Show what would be done without making changes
  -force           Force operation without confirmation
  -help            Show help message
  -v               Verbose output
```

## 📝 Usage Examples

### Complete Migration Workflow

```bash
# 1. Run complete migration (recommended)
go run migrate_mods.go -db ./data.db -v

# 2. Verify everything looks good
go run migrate_mods.go -action verify -db ./data.db

# 3. Test your application thoroughly

# 4. Optionally clean up old columns (destructive!)
go run migrate_mods.go -action cleanup -db ./data.db
```

### Step-by-Step Migration

```bash
# 1. Check current status
go run migrate_mods.go -action status -db ./data.db

# 2. Set up schema
go run migrate_mods.go -action setup -db ./data.db

# 3. Migrate data
go run migrate_mods.go -action migrate -db ./data.db

# 4. Verify migration
go run migrate_mods.go -action verify -db ./data.db
```

### Advanced Usage

```bash
# Dry run to preview changes
go run migrate_mods.go -action full -dry-run -v

# Custom database path
go run migrate_mods.go -db /path/to/custom/database.db

# Force operations (skip confirmations)
go run migrate_mods.go -action cleanup -force

# Disable automatic backup
go run migrate_mods.go -backup=false

# Verbose output for troubleshooting
go run migrate_mods.go -action migrate -v
```

## 🏗️ Using the Makefile

For convenience, use the included Makefile:

```bash
# Show all available targets
make help

# Quick migration with default database
make full

# Custom database path
make full DB_PATH=/path/to/data.db

# Other useful targets
make status                    # Check migration status
make verify                    # Verify migration
make dry-run                   # Preview full migration
make clean                     # Clean build artifacts
make build                     # Build binary
make install                   # Install to system PATH
```

## 🔍 Understanding Migration Status

The tool provides clear status information:

### Before Migration

```
Database: ./data.db (15.32 MB)
Modifications table exists: false
trouble_reports: 5 records with old mods
metal_sheets: 12 records with old mods
tools: 8 records with old mods

ℹ️  Migration needed - found 25 records with old mods
ℹ️  Run: go run migrate_mods.go -action=full
```

### After Migration

```
Database: ./data.db (15.45 MB)
Modifications table exists: true
Total modifications: 78
trouble_reports: 0 records with old mods
metal_sheets: 0 records with old mods
tools: 0 records with old mods

✅ Migration completed successfully!
ℹ️  Ready for cleanup. Run: go run migrate_mods.go -action=cleanup
```

### After Cleanup

```
Database: ./data.db (15.32 MB)
Modifications table exists: true
Total modifications: 78
trouble_reports: Old mods column not found (cleanup completed)
metal_sheets: Old mods column not found (cleanup completed)
tools: Old mods column not found (cleanup completed)

✅ Migration cleanup completed - old mods columns removed
ℹ️  New modification system is active with 78 modifications
```

## 🛡️ Safety Features

The migration tool includes several safety mechanisms:

### Automatic Backups

- **Created automatically** before any destructive operations
- **Timestamped** backups in `backups/` directory
- **Can be disabled** with `-backup=false`

### Dry Run Mode

- **Preview changes** without modifying database
- **Shows exactly** what would be executed
- **Safe to run** multiple times

### Verification System

- **Compares old vs new** data counts
- **Validates migration** integrity
- **Required before cleanup** (unless forced)

### Interactive Confirmations

- **Confirms destructive** operations
- **Can be bypassed** with `-force` flag
- **Shows warnings** for dangerous actions

## 🚨 Important Notes

### Before Migration

1. **⚠️ BACKUP YOUR DATABASE** - Always create backups (done automatically)
2. **Stop your application** during migration to prevent conflicts
3. **Test on a copy** of production data first if possible
4. **Ensure sufficient disk space** for migration and backups

### Cleanup Warning

The `cleanup` action **permanently removes** old `mods` columns and **cannot be undone**!

- Only run after thorough testing
- Automatically runs verification first (unless `-force` used)
- Consider keeping old columns for a grace period

### Performance Notes

- Migration speed: ~1,000-10,000 records per second
- Uses transactions to prevent database corruption
- WAL mode enabled for better concurrent access
- Indexes created automatically for optimal performance

## 🔧 Building and Distribution

### Build Binary

```bash
# Build migration binary
go build -o migrate_mods migrate_mods.go

# Use built binary
./migrate_mods -db ./data.db -action full
```

### Install System-Wide

```bash
# Install to /usr/local/bin
make install

# Now available as pgpress-migrate
pgpress-migrate -db ./data.db -action status
```

### Create Distribution Package

```bash
# Create tarball with all necessary files
make package
```

## ❓ Troubleshooting

### Common Issues

**Database not found**

```bash
❌ Database validation failed: database file does not exist: ./data.db
```

→ Check the database path with `-db` flag

**Permission denied**

```bash
❌ Failed to connect to database: unable to open database file
```

→ Check file permissions and ensure database isn't locked

**Column already exists**

```bash
⚠️ Column already exists (this is normal): duplicate column name: mods
```

→ This is normal and can be ignored

**Verification failed**

```bash
❌ Verification failed - found discrepancies in migration data
```

→ Check logs with `-v` flag, may need to re-run migration

### Getting Help

**Check status first:**

```bash
go run migrate_mods.go -action status -v
```

**Run with verbose output:**

```bash
go run migrate_mods.go -action full -v
```

**Use dry run to preview:**

```bash
go run migrate_mods.go -action full -dry-run -v
```

### Database Recovery

**If migration fails partway:**

```bash
# Check what was completed
go run migrate_mods.go -action status -v

# Resume migration
go run migrate_mods.go -action migrate
```

**Restore from backup:**

```bash
# Find backup file
ls -la backups/

# Restore (replace with actual backup filename)
cp backups/data_backup_2024-12-19_10-30-45.db ./data.db
```

## 📊 Migration Data Validation

### Manual Verification Queries

```sql
-- Check migration status
SELECT
    'trouble_reports' as table_name,
    COUNT(*) as total_records,
    SUM(CASE WHEN mods IS NOT NULL AND mods != '[]' THEN 1 ELSE 0 END) as with_mods
FROM trouble_reports;

-- Check new modifications
SELECT entity_type, COUNT(*) as modifications
FROM modifications
GROUP BY entity_type;

-- Recent activity
SELECT datetime(created_at), entity_type, entity_id
FROM modifications
ORDER BY created_at DESC
LIMIT 10;
```

### Data Integrity Checks

```bash
# Verify all tables
go run migrate_mods.go -action verify -v

# Check specific migration statistics
make stats
```

## 🔄 Migration from Old Commands

If you previously used `pgpress migration` commands:

| Old Command                 | New Command                                |
| --------------------------- | ------------------------------------------ |
| `pgpress migration status`  | `go run migrate_mods.go -action status`    |
| `pgpress migration run`     | `go run migrate_mods.go -action migrate`   |
| `pgpress migration verify`  | `go run migrate_mods.go -action verify`    |
| `pgpress migration cleanup` | `go run migrate_mods.go -action cleanup`   |
| `pgpress migration stats`   | `go run migrate_mods.go -action status -v` |

The new tool provides **enhanced functionality** with better safety features, dry-run capabilities, and automatic backups.

## 📈 What's New in v2.0

- **🎨 Colorized output** for better readability
- **🔄 Automatic backups** before destructive operations
- **👀 Dry-run mode** to preview changes safely
- **📊 Enhanced status reporting** with detailed information
- **🛡️ Better error handling** and recovery options
- **⚡ Performance improvements** with optimized queries
- **🧹 Simplified workflow** - one tool does everything

## 🤝 Support

- **Documentation**: This README and inline help (`-help`)
- **Status checking**: Use `-action status -v` for detailed information
- **Dry runs**: Always available with `-dry-run` flag
- **Verbose mode**: Use `-v` for detailed output during troubleshooting

---

**The migration tool is designed to be safe, reliable, and user-friendly. Most users should simply run `go run migrate_mods.go` and let the tool handle everything automatically.**
