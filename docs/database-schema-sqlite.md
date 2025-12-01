# SQLite Database Schema

This document describes the SQLite database schema used in the pg-press application.

## Tables

### tools
Stores information about press tools.
```sql
CREATE TABLE tools (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position TEXT NOT NULL,
    format BLOB NOT NULL,
    type TEXT NOT NULL,
    code TEXT NOT NULL,
    regenerating BOOLEAN NOT NULL DEFAULT 0,
    is_dead BOOLEAN NOT NULL DEFAULT 0,
    press INTEGER,
    binding INTEGER,
    FOREIGN KEY (binding) REFERENCES tools(id)
);
```

### press_cycles
Tracks cycle counts for tools on presses.
```sql
CREATE TABLE press_cycles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    press_number INTEGER NOT NULL,
    tool_id INTEGER NOT NULL,
    tool_position TEXT NOT NULL,
    total_cycles INTEGER NOT NULL,
    date DATETIME NOT NULL,
    performed_by INTEGER NOT NULL,
    FOREIGN KEY (tool_id) REFERENCES tools(id)
);
```

### users
Manages system users with Telegram integration.
```sql
CREATE TABLE users (
    telegram_id INTEGER PRIMARY KEY,
    user_name TEXT NOT NULL,
    api_key TEXT NOT NULL,
    last_feed INTEGER
);
```

### trouble_reports
Stores trouble reports submitted by users.
```sql
CREATE TABLE trouble_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    use_markdown BOOLEAN NOT NULL DEFAULT 0
);
```

### feeds
Stores feed entries (notifications/news).
```sql
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);
```

### attachments
Manages file attachments for trouble reports.
```sql
CREATE TABLE attachments (
    id TEXT PRIMARY KEY,
    mime_type TEXT NOT NULL,
    data BLOB NOT NULL
);
```

### notes
Stores system notes with different levels.
```sql
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    linked TEXT
);
```

### cookies
Manages user session cookies.
```sql
CREATE TABLE cookies (
    user_agent TEXT NOT NULL,
    value TEXT PRIMARY KEY,
    api_key TEXT NOT NULL,
    last_login INTEGER NOT NULL
);
```

### press_regenerations
Stores information about press regenerations.
```sql
CREATE TABLE press_regenerations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    press_number INTEGER NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME,
    status TEXT NOT NULL
);
```

### tool_regenerations
Tracks individual tool regenerations.
```sql
CREATE TABLE tool_regenerations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_id INTEGER NOT NULL,
    cycle_id INTEGER NOT NULL,
    start_date DATETIME NOT NULL,
    end_date DATETIME,
    FOREIGN KEY (tool_id) REFERENCES tools(id),
    FOREIGN KEY (cycle_id) REFERENCES press_cycles(id)
);
```

### modifications
Tracks modifications to tools.
```sql
CREATE TABLE modifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_id INTEGER NOT NULL,
    date DATETIME NOT NULL,
    description TEXT NOT NULL,
    performed_by INTEGER NOT NULL,
    FOREIGN KEY (tool_id) REFERENCES tools(id)
);
```

### metal_sheets
Stores information about metal sheets.
```sql
CREATE TABLE metal_sheets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    thickness REAL NOT NULL,
    material TEXT NOT NULL
);
```

## Indexes

### tools
```sql
CREATE INDEX idx_tools_position ON tools(position);
CREATE INDEX idx_tools_binding ON tools(binding);
```

### press_cycles
```sql
CREATE INDEX idx_press_cycles_tool_id ON press_cycles(tool_id);
CREATE INDEX idx_press_cycles_date ON press_cycles(date);
```

### feeds
```sql
CREATE INDEX idx_feeds_user_id ON feeds(user_id);
CREATE INDEX idx_feeds_created_at ON feeds(created_at);
```

### trouble_reports
```sql
CREATE INDEX idx_trouble_reports_date ON trouble_reports(id DESC);
```

## Relationships

- **tools** ↔ **press_cycles**: One-to-many (tool → cycles)
- **tools** ↔ **modifications**: One-to-many (tool → modifications)  
- **tools** ↔ **tool_regenerations**: One-to-many (tool → regenerations)
- **users** ↔ **feeds**: One-to-many (user → feeds)
- **users** ↔ **press_cycles**: One-to-many (user → cycles)
- **tools** ↔ **tool_regenerations**: One-to-many (tool → regenerations)
- **press_cycles** ↔ **tool_regenerations**: One-to-many (cycle → regenerations)
- **trouble_reports** ↔ **attachments**: Many-to-many through junction table (not shown, managed by application logic)

## Constraints

- All foreign key relationships are enforced through application logic
- Primary keys are auto-incremented where appropriate
- Required fields are marked as NOT NULL
- Data types follow the Go model definitions:
  - INTEGER for IDs and numbers
  - TEXT for strings
  - BLOB for binary data (formats, attachments)
  - DATETIME for date/time fields