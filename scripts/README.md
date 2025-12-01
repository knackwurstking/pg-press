# Database Migration Scripts

This directory contains various migration scripts for managing database schema changes.

## Available Scripts

### add-indexes-press.go

This is a Go-based script for adding indexes to the SQLite database to improve query performance.

#### Usage

```bash
# Run the script
go run ./scripts/add-indexes-press -v /path/to/database.db
```

#### Features

- Creates indexes on press_cycles.tool_id column for faster tool-specific cycle queries
- Creates indexes on press_cycles.press_number column for faster press-specific cycle queries  
- Creates indexes on press_cycles.date column for faster date-based sorting
- Uses CREATE INDEX IF NOT EXISTS to avoid errors if indexes already exist
- Provides clear status updates during execution

#### Requirements

- Go 1.25 or higher
- SQLite3 database file


### remove-indexes.go

This is a Go-based script for removing all indexes from the SQLite database.

#### Usage

```bash
# Run the script
go run ./scripts/remove-indexes -v /path/to/database.db
```

#### Features

- Automatically detects all indexes in the database
- Safely drops each index using DROP INDEX IF EXISTS
- Provides clear status updates during execution
- Error handling for failed index removals

#### Requirements

- Go 1.25 or higher
- SQLite3 database file
