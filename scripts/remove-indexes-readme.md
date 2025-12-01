# Remove Indexes Script

This is a Go-based script for removing all indexes from the SQLite database.

## Usage

```bash
# Build the script
go build -o remove-indexes scripts/remove-indexes.go

# Run the script
./remove-indexes /path/to/database.db
```

## Features

- Automatically detects all indexes in the database
- Safely drops each index using DROP INDEX IF EXISTS
- Provides clear status updates during execution
- Error handling for failed index removals

## Requirements

- Go 1.25 or higher
- SQLite3 database file