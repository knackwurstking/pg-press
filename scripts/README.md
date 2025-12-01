# Database Migration Scripts

This directory contains various migration scripts for managing database schema changes.

## Available Scripts

### remove-indexes.go

This is a Go-based script for removing all indexes from the SQLite database.

#### Usage

```bash
# Run the script
go run scripts/remove-indexes.go /path/to/database.db
```

#### Features

- Automatically detects all indexes in the database
- Safely drops each index using DROP INDEX IF EXISTS
- Provides clear status updates during execution
- Error handling for failed index removals

#### Requirements

- Go 1.25 or higher
- SQLite3 database file
