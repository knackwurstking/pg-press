# Database Schema Documentation

This document provides comprehensive documentation for the PG Press SQLite database schema, including table structures, relationships, constraints, and usage patterns.

## Overview

PG Press uses SQLite as its primary database system, chosen for its simplicity, reliability, and zero-configuration deployment. The database schema is designed to support:

- Manufacturing tool and press management
- User authentication and session management
- Trouble reporting with attachments
- Comprehensive notes system
- Activity feed and audit trails
- Metal sheet inventory management

## Database Architecture

### Connection Management

The database is managed through the `internal/database/DB` struct, which provides centralized access to all data access objects (DAOs). Each service in `internal/services/` manages a specific domain of the application.

```go
type DB struct {
    db *sql.DB

    // Core Services
    Users          *services.Users
    Tools          *services.Tools
    PressCycles    *services.PressCycles
    TroubleReports *services.TroubleReports
    Notes          *services.Notes
    MetalSheets    *services.MetalSheets

    // Supporting Services
    Attachments       *services.Attachments
    Cookies           *services.Cookies
    ToolRegenerations *services.ToolRegenerations
    Feeds             *services.Feeds
    Modifications     *services.Modifications
}
```

### Data Types

- **INTEGER**: Primary keys, foreign keys, numeric IDs, counters
- **TEXT**: String data, codes, names, descriptions
- **BLOB**: Binary data (JSON, file attachments, structured data)
- **DATETIME**: Timestamps, creation/update times
- **REAL**: Floating-point numbers for measurements
- **BOOLEAN**: Binary flags (stored as INTEGER 0/1 in SQLite)

## Core Tables

### users

Manages user authentication and profile information.

```sql
CREATE TABLE users (
    telegram_id INTEGER NOT NULL PRIMARY KEY,
    user_name   TEXT NOT NULL,
    api_key     TEXT NOT NULL UNIQUE,
    last_feed   TEXT NOT NULL DEFAULT ''
);
```

**Fields:**

- `telegram_id`: Unique identifier from Telegram integration (Primary Key)
- `user_name`: Display name for the user
- `api_key`: Generated API key for authentication
- `last_feed`: ID of the last feed entry the user has acknowledged

**Relationships:**

- Referenced by `press_cycles.performed_by`
- Referenced by `tool_regenerations.performed_by`
- Connected to `cookies` via `api_key`

**Indexes:**

- Primary key on `telegram_id`
- Unique index on `api_key`

### cookies

Manages user session cookies and authentication tokens.

```sql
CREATE TABLE cookies (
    user_agent TEXT NOT NULL,
    value      TEXT NOT NULL PRIMARY KEY,
    api_key    TEXT NOT NULL,
    last_login INTEGER NOT NULL
);
```

**Fields:**

- `user_agent`: Browser/client user agent string
- `value`: Session cookie value (Primary Key)
- `api_key`: References user's API key
- `last_login`: Unix timestamp of last login

**Relationships:**

- Links to `users.api_key`

**Security Notes:**

- Cookie values should be cryptographically secure random strings
- `last_login` enables session timeout functionality
- `user_agent` helps with session validation

### tools

Central table for manufacturing tool management.

```sql
CREATE TABLE tools (
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    position     TEXT NOT NULL,
    format       BLOB NOT NULL,
    type         TEXT NOT NULL,
    code         TEXT NOT NULL,
    regenerating BOOLEAN NOT NULL DEFAULT 0,
    press        INTEGER
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `position`: Tool position (e.g., 'top', 'bottom')
- `format`: JSON BLOB containing tool format specifications
- `type`: Tool classification/category
- `code`: Unique tool identifier code
- `regenerating`: Flag indicating if tool is currently being regenerated
- `press`: Press number (0-5) where tool is currently assigned

**Relationships:**

- Referenced by `press_cycles.tool_id`
- Referenced by `tool_regenerations.tool_id`
- Referenced by `metal_sheets.tool_id`
- Linked to `notes` via generic linking system (`tool_{id}`)

**Business Rules:**

- Tool codes should be unique within the system
- Press values are constrained to 0-5 range
- Format BLOB contains structured JSON data for tool specifications

### press_cycles

Records press cycle data for performance tracking and maintenance scheduling.

```sql
CREATE TABLE press_cycles (
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
    tool_id      INTEGER NOT NULL,
    tool_position TEXT NOT NULL,
    date         DATETIME NOT NULL,
    total_cycles INTEGER NOT NULL DEFAULT 0,
    performed_by INTEGER NOT NULL,
    FOREIGN KEY(tool_id) REFERENCES tools(id),
    FOREIGN KEY(performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `press_number`: Press identifier (0-5)
- `tool_id`: Foreign key to associated tool
- `tool_position`: Position of tool during cycle (top/bottom)
- `date`: Timestamp when cycles were recorded
- `total_cycles`: Total cycle count for this record
- `performed_by`: User who recorded the cycles (required)

**Constraints:**

- `press_number` must be between 0 and 5
- `tool_id` must reference existing tool
- `tool_position` must be specified
- `performed_by` must reference existing user, set to NULL if user is deleted

**Indexes Recommended:**

```sql
CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_position ON press_cycles(tool_position);
CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
```

### trouble_reports

Manages issue reporting and documentation system.

```sql
CREATE TABLE trouble_reports (
    id                 INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    title              TEXT NOT NULL,
    content            TEXT NOT NULL,
    linked_attachments TEXT NOT NULL DEFAULT '',
    use_markdown       BOOLEAN DEFAULT 0
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `title`: Report title/summary
- `content`: Detailed report content
- `linked_attachments`: JSON array of attachment IDs
- `use_markdown`: Boolean flag indicating if content should be rendered as markdown

**Relationships:**

- Links to `attachments` via `linked_attachments` JSON array
- Tracked in `feeds` for activity monitoring

**Data Formats:**

```json
// linked_attachments format
["123", "456", "789"]
```

### attachments

Stores file attachments for trouble reports.

```sql
CREATE TABLE attachments (
    id        INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    mime_type TEXT NOT NULL,
    data      BLOB NOT NULL
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `mime_type`: MIME type for proper file handling
- `data`: Binary file data

**Supported MIME Types:**

- Images: `image/jpeg`, `image/png`, `image/gif`
- Documents: `application/pdf`, `text/plain`
- Archives: `application/zip`

**Storage Considerations:**

- File size limits should be enforced at application level
- Consider external storage for large files in production

### notes

Flexible note system supporting multiple entity types.

```sql
CREATE TABLE notes (
    id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    level      INTEGER NOT NULL,
    content    TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked     TEXT DEFAULT ''
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `level`: Priority level (0=INFO, 1=ATTENTION, 2=BROKEN)
- `content`: Note content (supports Markdown)
- `created_at`: Creation timestamp
- `linked`: Generic entity reference

**Priority Levels:**

- `0` (INFO): General information and documentation
- `1` (ATTENTION): Important notices requiring attention
- `2` (BROKEN): Critical issues indicating equipment problems

**Linking System:**
The `linked` field uses a flexible string format to connect notes to any entity:

- `tool_123`: Links to tool with ID 123
- `press_5`: Links to press number 5
- `machine_42`: Could link to any future entity type
- `''`: Unlinked/global notes

**Query Examples:**

```sql
-- Get all notes for a specific tool
SELECT * FROM notes WHERE linked = 'tool_123' ORDER BY created_at DESC;

-- Get all critical notes across the system
SELECT * FROM notes WHERE level = 2 ORDER BY created_at DESC;

-- Get all notes for a press
SELECT * FROM notes WHERE linked = 'press_5' ORDER BY created_at DESC;
```

### metal_sheets

Manages metal sheet inventory and specifications.

```sql
CREATE TABLE metal_sheets (
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    tile_height  REAL NOT NULL,
    value        REAL NOT NULL,
    marke_height INTEGER NOT NULL,
    stf          REAL NOT NULL,
    stf_max      REAL NOT NULL,
    identifier   TEXT NOT NULL,
    tool_id      INTEGER NOT NULL,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `tile_height`: Physical tile height measurement
- `value`: Sheet value/cost
- `marke_height`: Mark height specification
- `stf`: STF measurement value
- `stf_max`: Maximum STF value
- `identifier`: Unique identifier for the metal sheet
- `tool_id`: Required assignment to specific tool
- `updated_at`: Last modification timestamp

**Relationships:**

- Required foreign key to `tools.id` with CASCADE delete

### modifications

Tracks modification history for various entities in the system.

```sql
CREATE TABLE modifications (
    id          INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id     INTEGER NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   INTEGER NOT NULL,
    data        BLOB NOT NULL,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE SET NULL
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `user_id`: User who made the modification
- `entity_type`: Type of entity modified (e.g., 'trouble_report')
- `entity_id`: ID of the modified entity
- `data`: JSON BLOB containing modification details
- `created_at`: When the modification was recorded

**Relationships:**

- `user_id` references `users.telegram_id`

**Indexes:**

```sql
CREATE INDEX IF NOT EXISTS idx_modifications_entity ON modifications(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_modifications_user ON modifications(user_id);
CREATE INDEX IF NOT EXISTS idx_modifications_created_at ON modifications(created_at);
```

### tool_regenerations

Tracks tool regeneration history and scheduling.

```sql
CREATE TABLE tool_regenerations (
    id           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    tool_id      INTEGER NOT NULL,
    cycle_id     INTEGER NOT NULL,
    reason       TEXT,
    performed_by INTEGER,
    FOREIGN KEY(tool_id) REFERENCES tools(id) ON DELETE CASCADE,
    FOREIGN KEY(cycle_id) REFERENCES press_cycles(id) ON DELETE CASCADE,
    FOREIGN KEY(performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `tool_id`: Tool being regenerated
- `cycle_id`: Press cycle record that triggered regeneration
- `reason`: Optional reason for regeneration
- `performed_by`: User who initiated the regeneration

**Cascade Behavior:**

- Deleting a tool removes all its regeneration records
- Deleting a cycle removes related regeneration records
- Deleting a user sets `performed_by` to NULL

### feeds

Activity feed system for audit trails and notifications.

```sql
CREATE TABLE feeds (
    id         INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    title      TEXT NOT NULL,
    content    TEXT NOT NULL,
    user_id    INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);
```

**Fields:**

- `id`: Auto-incrementing primary key
- `title`: Feed entry title
- `content`: Feed entry content/description
- `user_id`: ID of user who triggered the feed entry
- `created_at`: Unix timestamp when the feed entry was created

**Relationships:**

- `user_id` references `users.telegram_id`

**Indexes:**

```sql
CREATE INDEX IF NOT EXISTS idx_feeds_created_at ON feeds(created_at);
CREATE INDEX IF NOT EXISTS idx_feeds_user_id ON feeds(user_id);
```

## Indexes and Performance

### Recommended Indexes

```sql
-- Performance indexes for common queries
CREATE INDEX idx_press_cycles_tool_date ON press_cycles(tool_id, date);
CREATE INDEX idx_press_cycles_press ON press_cycles(press_number);
CREATE INDEX idx_notes_linked ON notes(linked);
CREATE INDEX idx_notes_level ON notes(level);
CREATE INDEX idx_feeds_created_at ON feeds(created_at);
CREATE INDEX idx_feeds_user_id ON feeds(user_id);
CREATE INDEX idx_tools_press ON tools(press);
CREATE INDEX idx_tools_code ON tools(code);
```

### Query Optimization

**Common Query Patterns:**

1. **Tool with Notes:**

```sql
SELECT t.*, n.id as note_id, n.level, n.content, n.created_at as note_created
FROM tools t
LEFT JOIN notes n ON n.linked = 'tool_' || t.id
WHERE t.id = ?
ORDER BY n.level DESC, n.created_at DESC;
```

2. **Press Performance Report:**

```sql
SELECT pc.press_number, COUNT(*) as cycle_count, SUM(pc.total_cycles) as total_cycles
FROM press_cycles pc
WHERE pc.date >= ? AND pc.date <= ?
GROUP BY pc.press_number
ORDER BY pc.press_number;
```

3. **Critical Notes Dashboard:**

```sql
SELECT n.*,
       CASE
         WHEN n.linked LIKE 'tool_%' THEN 'Tool'
         WHEN n.linked LIKE 'press_%' THEN 'Press'
         ELSE 'Other'
       END as entity_type
FROM notes n
WHERE n.level >= 1
ORDER BY n.level DESC, n.created_at DESC;
```

## Data Integrity and Constraints

### Foreign Key Constraints

The database enforces referential integrity through foreign key constraints:

- `press_cycles.tool_id` → `tools.id`
- `press_cycles.performed_by` → `users.telegram_id`
- `tool_regenerations.tool_id` → `tools.id`
- `tool_regenerations.cycle_id` → `press_cycles.id`
- `tool_regenerations.performed_by` → `users.telegram_id`
- `metal_sheets.tool_id` → `tools.id`

### Check Constraints

- `press_cycles.press_number` must be between 0 and 5
- Note levels must be 0, 1, or 2 (enforced at application level)

### Unique Constraints

- `users.api_key` must be unique
- `users.telegram_id` is unique (primary key)
- `cookies.value` must be unique (primary key)

## Migration Strategy

### Schema Evolution

Database migrations are handled through the application's table creation methods. When schema changes are needed:

1. Update the `createTable()` method in the relevant service
2. Add migration logic for existing data
3. Test with production data backup
4. Deploy with rollback plan

### Backup and Recovery

```bash
# Backup database
sqlite3 pg-press.db ".backup backup-$(date +%Y%m%d).db"

# Restore from backup
cp backup-20231201.db pg-press.db

# Export as SQL
sqlite3 pg-press.db ".dump" > pg-press-dump.sql
```

## Security Considerations

### Data Protection

- API keys are stored as secure random strings
- Session cookies should be HTTP-only and secure
- File uploads are validated by MIME type
- User input is sanitized before database insertion

### Access Control

- All database operations require authenticated user context
- Sensitive operations log user attribution
- Foreign key constraints prevent orphaned data
- Cascade deletes are carefully controlled

### Audit Trail

The feeds table provides comprehensive audit logging:

- All significant data changes are logged
- User attribution for all operations
- Timestamp-based event ordering
- JSON payload for detailed change tracking

## Troubleshooting

### Common Issues

**Database Lock Errors:**

```bash
# Check for existing connections
lsof pg-press.db

# Force unlock (use with caution)
sqlite3 pg-press.db ".timeout 1000"
```

**Integrity Constraint Violations:**

```sql
-- Check foreign key violations
PRAGMA foreign_key_check;

-- Verify data consistency
SELECT * FROM press_cycles pc
LEFT JOIN tools t ON pc.tool_id = t.id
WHERE t.id IS NULL;
```

**Performance Issues:**

```sql
-- Analyze query performance
EXPLAIN QUERY PLAN SELECT ...;

-- Update table statistics
ANALYZE;

-- Check index usage
.schema tools
```

### Maintenance Commands

```sql
-- Optimize database
VACUUM;

-- Update statistics
ANALYZE;

-- Check integrity
PRAGMA integrity_check;

-- Rebuild indexes
REINDEX;
```

## Future Enhancements

### Planned Schema Changes

1. **Audit Log Enhancements:**
   - Dedicated audit table with standardized format
   - Automatic trigger-based logging
   - Enhanced query capabilities

2. **File Storage Optimization:**
   - External file storage for large attachments
   - File metadata table
   - Automated cleanup of orphaned files

3. **Advanced Indexing:**
   - Full-text search indexes for notes and reports
   - Composite indexes for complex queries
   - Partial indexes for filtered queries

4. **Data Partitioning:**
   - Time-based partitioning for feeds table
   - Archive strategy for old data
   - Performance optimization for large datasets

This documentation serves as the definitive guide for understanding and working with the PG Press database schema. Keep it updated as the schema evolves to ensure accuracy and completeness.
