# Database Migration Scripts

This directory contains database migration scripts for the pg-press application.

## Available Migrations

### migrate-trouble-reports-markdown.sh

**Purpose**: Adds markdown support to trouble reports by adding a `use_markdown` column to the `trouble_reports` table.

**What it does**:

- Adds a `use_markdown BOOLEAN DEFAULT 0` column to the trouble_reports table
- Sets all existing records to `use_markdown = 0` for backward compatibility
- Creates a backup of your database before making changes
- Validates the migration was successful

**Prerequisites**:

- SQLite3 must be installed on your system
- The pg-press database must exist (`pg-press.db`)
- You must run the script from the project root directory
- **Stop your application before running this migration**

**Usage**:

```bash
# From the project root directory
./scripts/migrate-trouble-reports-markdown.sh

# For help
./scripts/migrate-trouble-reports-markdown.sh --help

# Dry-run to see what would be changed
./scripts/migrate-trouble-reports-markdown.sh --dry-run
```

**What gets changed**:

Before migration:

```sql
CREATE TABLE trouble_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    linked_attachments TEXT NOT NULL
);
```

After migration:

```sql
CREATE TABLE trouble_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    linked_attachments TEXT NOT NULL,
    use_markdown BOOLEAN DEFAULT 0  -- NEW FIELD
);
```

**New Features Enabled**:

- **Optional Markdown**: Users can opt-in to markdown formatting via checkbox
- **Rich Formatting**: Support for headers, bold, italic, lists, code blocks, links, tables
- **HTML Rendering**: Markdown content displays as formatted HTML in web interface
- **PDF Integration**: Markdown content is properly formatted in PDF exports
- **Security**: Built-in XSS prevention and HTML sanitization
- **Backward Compatibility**: All existing reports continue to work as plain text

**Supported Markdown Syntax**:

```markdown
# Header 1

## Header 2

### Header 3

**Bold text**
_Italic text_
~~Strikethrough~~

- Unordered list
- Another item

1. Ordered list
2. Another item

`Inline code`
```

Code block

```

[Link text](https://example.com)

| Table | Header |
|-------|--------|
| Cell  | Value  |
```

**Post-migration**:
After running the migration:

1. Restart your pg-press application
2. Users will see a "Markdown-Formatierung verwenden" checkbox in the edit dialog
3. When enabled, content will be rendered as HTML with proper formatting
4. PDF exports will include formatted content
5. All existing reports continue to display as plain text
6. New reports default to plain text unless markdown is explicitly enabled

**Security Features**:

- **XSS Prevention**: Dangerous HTML tags and scripts are automatically removed
- **Safe Rendering**: Uses Go's template.HTML for secure output
- **Input Sanitization**: Event handlers and dangerous protocols are filtered
- **Whitelist Approach**: Only safe HTML elements are allowed in output

**Backup and Recovery**:

- The script automatically creates a timestamped backup in `scripts/backups/`
- Backup format: `pg-press_before_markdown_migration_YYYYMMDD_HHMMSS.db`
- To restore from backup: `cp backups/backup_file.db pg-press.db`

---

### migrate-notes-system-refactor.sh

**Purpose**: Migrates the database to the new generic notes linking system, removing complex tool-specific linking.

**What it does**:

- Updates the notes table with a generic 'linked' column for flexible entity linking
- Removes notes columns from tools and metal_sheets tables
- Migrates existing linking data from old format to new format
- Creates a backup of your database before making changes
- Validates the migration was successful

**Prerequisites**:

- SQLite3 must be installed on your system
- The pg-press database must exist (`pg-press.db`)
- You must run the script from the project root directory
- **Stop your application before running this migration**

**Usage**:

```bash
# From the project root directory
./scripts/migrate-notes-system-refactor.sh

# For help
./scripts/migrate-notes-system-refactor.sh --help

# Dry-run to analyze current schema
./scripts/migrate-notes-system-refactor.sh --dry-run
```

**What gets changed**:

Before migration:

```sql
-- Old notes table (if exists)
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked_to_press INTEGER,
    linked_to_tool INTEGER
);

-- Tools table with notes column
CREATE TABLE tools (
    -- other fields
    notes BLOB NOT NULL
);

-- Metal sheets with notes column
CREATE TABLE metal_sheets (
    -- other fields
    notes BLOB NOT NULL
);
```

After migration:

```sql
-- New notes table with generic linking
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked TEXT  -- Generic entity reference: "tool_123", "press_5", etc.
);

-- Tools table without notes column
CREATE TABLE tools (
    -- other fields, no notes column
);

-- Metal sheets without notes column
CREATE TABLE metal_sheets (
    -- other fields, no notes column
);
```

**New Notes Linking Format**:

- `"tool_123"` - Links note to tool with ID 123
- `"press_5"` - Links note to press number 5
- `"machine_42"` - Can link to any future entity type
- `""` or `NULL` - Unlinked notes

**Benefits of New System**:

- **Semantic Correctness**: Press notes stay with press, not tools
- **Maximum Flexibility**: Can link to any entity type without schema changes
- **Simplified Operations**: No complex relationship management
- **Better Performance**: Direct queries without JOINs
- **Easier Maintenance**: Single linking field to manage

**Post-migration**:
After running the migration:

1. Restart your application to pick up the database schema changes
2. Notes will use the new generic linking system
3. Press notes will stay with presses regardless of tool changes
4. Tool notes will stay with specific tools
5. The system is ready for linking to any new entity types

**Backup and Recovery**:

- The script automatically creates a timestamped backup in `scripts/backups/`
- Backup format: `pg-press_before_notes_refactor_YYYYMMDD_HHMMSS.db`
- To restore from backup: `cp backups/backup_file.db pg-press.db`

**Troubleshooting NULL Values**:

If you encounter an error like "converting NULL to string is unsupported" when accessing the /notes page, it means there are NULL values in the linked column that need to be fixed:

```bash
# Quick manual fix using SQLite
sqlite3 pg-press.db "UPDATE notes SET linked = '' WHERE linked IS NULL;"
```

**Manual SQL Fix**:

```sql
-- Convert NULL linked values to empty strings
UPDATE notes SET linked = '' WHERE linked IS NULL;

-- Verify the fix
SELECT COUNT(*) as null_count FROM notes WHERE linked IS NULL;
-- Should return 0

-- Check distribution
SELECT
  CASE
    WHEN linked = '' THEN 'Unlinked'
    WHEN linked LIKE 'tool_%' THEN 'Tool Links'
    WHEN linked LIKE 'press_%' THEN 'Press Links'
    ELSE 'Other'
  END as type,
  COUNT(*) as count
FROM notes GROUP BY type;
```

## Migration Order

If you're setting up a fresh database or need to run multiple migrations, run them in this order:

1. `migrate-notes-system-refactor.sh` - Updates notes system (if needed)
2. `migrate-trouble-reports-markdown.sh` - Adds markdown support to trouble reports

## General Migration Guidelines

1. **Always backup your database** before running any migration
2. **Test migrations** on a copy of your data first
3. **Run from the project root** directory unless otherwise specified
4. **Stop the application** before running migrations
5. **Restart the application** after successful migrations
6. **Check logs** after restart to ensure everything works correctly

## Safety Features

All migration scripts include:

- **Automatic Backups**: Database is backed up before any changes
- **Validation**: Migration success is verified before completion
- **Dry-run Mode**: See what would be changed without making changes
- **Error Handling**: Scripts stop on first error to prevent data corruption
- **Detailed Logging**: Color-coded output shows progress and results

## Backup Directory Structure

```
scripts/backups/
├── pg-press_before_notes_refactor_20240101_120000.db
└── pg-press_before_markdown_migration_20240101_130000.db
```

## Rollback Procedures

To rollback a migration:

1. **Stop the application**
2. **Identify the backup** file you want to restore from
3. **Restore the database**: `cp scripts/backups/backup_file.db pg-press.db`
4. **Restart the application**
5. **Verify** the application works with the restored database

## Getting Help

- Use `--help` flag with any migration script for detailed usage information
- Use `--dry-run` flag to see what would be changed without making changes
- Check the backup directory if you need to restore previous state
- Ensure SQLite3 is installed: `sqlite3 --version`
- Verify you're in the correct directory (project root)

## Development Notes

- All migration scripts follow the same pattern for consistency
- Scripts are designed to be idempotent where possible
- Color-coded output helps distinguish between different types of messages
- Comprehensive error checking prevents partial migrations
- Scripts create detailed logs of all changes made

## Future Migrations

When adding new migration scripts:

1. Use descriptive names: `migrate-description-of-change.sh`
2. Include backup functionality
3. Add validation steps
4. Document in this README
5. Make scripts executable: `chmod +x script-name.sh`
6. Follow the established pattern for consistency
