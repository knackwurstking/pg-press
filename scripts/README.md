# Database Migration Scripts

This directory contains various migration scripts for managing database schema changes.

## Available Scripts

### remove-indexes.sh
This script removes all indexes from the SQLite database to improve performance during bulk operations.

**Usage:**
```bash
# Use default database
./scripts/remove-indexes.sh

# Specify custom database path
./scripts/remove-indexes.sh -d /path/to/database.db

# Or use environment variable
DB_PATH=/path/to/database.db ./scripts/remove-indexes.sh
```

**Features:**
- Creates automatic backup before making changes
- Shows all indexes that will be removed
- Handles errors gracefully
- Provides clear status updates during execution

**Note:** Removing indexes can impact query performance. Re-add indexes after bulk operations are complete.

### migrate-tools-binding.sh
Migration script to add the binding column to tools table.

**Usage:**
```bash
# Use default database
./scripts/migrate-tools-binding.sh

# Specify custom database path
./scripts/migrate-tools-binding.sh -d /path/to/database.db
```

### migrate.sh
Main migration script that runs all available migrations.

**Usage:**
```bash
# Use default database
./scripts/migrate.sh

# Specify custom database path
./scripts/migrate.sh /path/to/database.db
```