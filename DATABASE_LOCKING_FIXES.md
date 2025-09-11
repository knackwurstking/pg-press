# Database Locking Fixes for Migration Commands

## Problem Summary

The `pgpress migration run` command was experiencing "database is locked" errors when processing trouble reports. This issue prevented successful migration from the old mod system to the new modification service.

## Root Causes Identified

### 1. SQLite Connection Configuration Issues

- **Missing WAL mode**: Default SQLite journaling mode can cause extended locks
- **No busy timeout**: Connections would fail immediately on lock contention
- **Missing connection pool limits**: Multiple connections could cause resource conflicts
- **No connection lifecycle management**: Connections weren't properly closed

### 2. Transaction Management Problems

- **No transaction boundaries**: Large migration operations ran as individual statements
- **Extended lock periods**: Row-by-row processing held locks for long durations
- **No rollback handling**: Failed operations could leave database in inconsistent state

### 3. Resource Management Issues

- **Unclosed database connections**: Migration commands didn't properly close DB connections
- **No connection reuse**: Each migration step opened new connections

## Fixes Implemented

### 1. Enhanced Database Connection Configuration (`utils.go`)

```go
// Added SQLite-specific connection parameters
connectionString := path + "?_busy_timeout=30000&_journal_mode=WAL&_foreign_keys=on&_synchronous=NORMAL"

// Configured connection pool for SQLite
db.SetMaxOpenConns(1)    // SQLite works best with single writer
db.SetMaxIdleConns(1)    // Keep one connection alive
db.SetConnMaxLifetime(0) // No maximum lifetime

// Added connection validation
if err = db.Ping(); err != nil {
    db.Close()
    return nil, err
}
```

**Key improvements:**

- **WAL Mode**: Write-Ahead Logging allows concurrent readers with single writer
- **Busy Timeout**: 30-second timeout prevents immediate lock failures
- **Connection Limits**: Single writer prevents SQLite lock contention
- **Connection Testing**: Validates connectivity before use

### 2. Transaction-Based Migration (`modification_migration.go`)

```go
// Each migration method now uses transactions
tx, err := m.db.Begin()
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// Use transaction for queries and inserts
rows, err := tx.Query(query)
// ... processing ...
err = tx.Commit()
```

**Benefits:**

- **Atomic operations**: All changes in a migration step succeed or fail together
- **Reduced lock time**: Transaction boundaries minimize lock duration
- **Consistency**: Rollback on error prevents partial migrations
- **Performance**: Batched operations are more efficient

### 3. Proper Resource Management (`commands-migration.go`)

```go
// Added deferred cleanup to all migration commands
db, err := openDB(*customDBPath)
if err != nil {
    return err
}
defer db.GetDB().Close() // Ensures connection cleanup
```

### 4. Database Connection Testing

Added new `test-db` command to help diagnose connection issues:

```bash
pgpress migration test-db
```

**Features:**

- Tests basic connectivity
- Verifies WAL mode is enabled
- Performs transaction test
- Provides troubleshooting guidance

## Usage Guide

### 1. Test Database Connection First

Before running migrations, test your database connection:

```bash
# Test default database
pgpress migration test-db

# Test custom database path
pgpress migration test-db --db /path/to/custom.db
```

### 2. Updated Migration Workflow

```bash
# 1. Test database connectivity
pgpress migration test-db

# 2. Check migration status
pgpress migration status

# 3. Run migration (now with better transaction handling)
pgpress migration run

# 4. Verify migration success
pgpress migration verify

# 5. Review statistics
pgpress migration stats

# 6. Test application thoroughly

# 7. Clean up old columns (optional)
pgpress migration cleanup
```

### 3. Troubleshooting Database Locks

If you still encounter locking issues:

1. **Check for other processes**: Ensure no other applications are using the database
2. **File permissions**: Verify read/write access to database file and directory
3. **Disk space**: Ensure sufficient space for WAL files
4. **Use custom path**: Try with `--db` flag to use a different database location

## Configuration Details

### SQLite Connection Parameters

| Parameter       | Value   | Purpose                            |
| --------------- | ------- | ---------------------------------- |
| `_busy_timeout` | 30000ms | Wait 30 seconds for locks to clear |
| `_journal_mode` | WAL     | Enable Write-Ahead Logging         |
| `_foreign_keys` | on      | Enforce referential integrity      |
| `_synchronous`  | NORMAL  | Balance safety and performance     |

### Connection Pool Settings

- **Max Open Connections**: 1 (SQLite single writer limitation)
- **Max Idle Connections**: 1 (Keep connection alive)
- **Connection Lifetime**: Unlimited (Stable connection)

## Benefits of These Changes

1. **Eliminated Database Locking**: WAL mode and proper timeouts prevent most lock conflicts
2. **Improved Performance**: Transaction batching reduces I/O operations
3. **Better Error Handling**: Rollback capability prevents corrupted migrations
4. **Enhanced Debugging**: Test command helps identify connection issues
5. **Resource Efficiency**: Proper connection management prevents leaks
6. **Production Ready**: Configuration suitable for concurrent access patterns

## Backward Compatibility

These changes are fully backward compatible:

- Existing database files work without modification
- Migration commands have same interface
- No breaking changes to existing functionality
- WAL mode is automatically enabled for better performance

## Monitoring and Logging

The migration process now includes:

- Transaction start/commit/rollback logging
- Connection configuration logging
- Detailed error messages for troubleshooting
- Migration statistics with timing information

## Next Steps

1. **Backup Strategy**: Always backup database before migrations
2. **Testing**: Run `test-db` command before important migrations
3. **Monitoring**: Watch logs during migration for any warnings
4. **Performance**: Monitor migration performance with new transaction batching

For additional support or questions about these fixes, refer to the enhanced help command:

```bash
pgpress migration help
```
