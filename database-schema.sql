-- ============================================================================
-- PG-VIS Database Schema Documentation
-- ============================================================================
-- This file documents the complete database schema for the PG Press
-- Visualization and Management System.
--
-- Database: SQLite
-- Version: 1.0.0
-- ============================================================================

-- ============================================================================
-- USERS & AUTHENTICATION
-- ============================================================================

-- Users table - System users with roles and permissions
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('admin', 'user', 'viewer')),
    active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Cookies table - Session management
CREATE TABLE IF NOT EXISTS cookies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_cookies_token ON cookies(token);
CREATE INDEX idx_cookies_expires ON cookies(expires_at);

-- ============================================================================
-- TOOLS MANAGEMENT
-- ============================================================================

-- Tools table - Manufacturing tools/dies
CREATE TABLE IF NOT EXISTS tools (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position TEXT NOT NULL CHECK(position IN ('top', 'bottom', 'left', 'right')),
    format BLOB NOT NULL,  -- JSON: {"width": number, "height": number}
    type TEXT NOT NULL,     -- e.g., 'MASS', 'FORM', etc.
    code TEXT NOT NULL,     -- Tool code/identifier
    status TEXT NOT NULL DEFAULT 'available'
        CHECK(status IN ('active', 'available', 'regenerating')),
    press INTEGER,          -- Press number (0-5) when active, NULL when not
    notes BLOB NOT NULL,    -- JSON: array of note IDs
    mods BLOB NOT NULL      -- JSON: modifications history
);
CREATE INDEX idx_tools_status ON tools(status);
CREATE INDEX idx_tools_press ON tools(press);
CREATE INDEX idx_tools_code ON tools(code);

-- Tool regenerations table - Tracks tool regeneration history
CREATE TABLE IF NOT EXISTS tool_regenerations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_id INTEGER NOT NULL,
    regenerated_at DATETIME NOT NULL,
    cycles_at_regeneration INTEGER NOT NULL DEFAULT 0,
    reason TEXT,
    performed_by TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE
);
CREATE INDEX idx_tool_regenerations_tool_id ON tool_regenerations(tool_id);
CREATE INDEX idx_tool_regenerations_date ON tool_regenerations(regenerated_at);

-- ============================================================================
-- PRESS MANAGEMENT
-- ============================================================================

-- Press cycles table - Tracks tool usage on presses
CREATE TABLE IF NOT EXISTS press_cycles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),  -- Press 0-5
    tool_id INTEGER NOT NULL,
    from_date DATETIME NOT NULL,           -- When tool was mounted
    to_date DATETIME,                      -- When tool was removed (NULL = still active)
    total_cycles INTEGER NOT NULL DEFAULT 0,
    partial_cycles INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tool_id) REFERENCES tools(id)
);
CREATE INDEX idx_press_cycles_tool_id ON press_cycles(tool_id);
CREATE INDEX idx_press_cycles_press_number ON press_cycles(press_number);
CREATE INDEX idx_press_cycles_dates ON press_cycles(from_date, to_date);

-- ============================================================================
-- METAL SHEETS CONFIGURATION
-- ============================================================================

-- Metal sheets table - Configuration for different metal sheet types
CREATE TABLE IF NOT EXISTS metal_sheets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tool_id INTEGER NOT NULL,
    tile_height REAL NOT NULL,             -- Thickness in mm
    value REAL NOT NULL,                   -- Sheet value in mm
    marke_height INTEGER,                  -- Mark height in mm (for bottom tools)
    stf REAL,                              -- STF value (for bottom tools)
    stf_max REAL,                          -- Maximum STF value (for bottom tools)
    notes BLOB,                            -- JSON: array of note IDs
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE CASCADE
);
CREATE INDEX idx_metal_sheets_tool_id ON metal_sheets(tool_id);

-- ============================================================================
-- NOTES SYSTEM
-- ============================================================================

-- Notes table - General notes that can be linked to various entities
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    level TEXT NOT NULL CHECK(level IN ('INFO', 'ATTENTION', 'BROKEN')),
    content TEXT NOT NULL,
    created_by INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX idx_notes_level ON notes(level);
CREATE INDEX idx_notes_created_at ON notes(created_at);

-- ============================================================================
-- TROUBLE REPORTS
-- ============================================================================

-- Trouble reports table - Issue tracking system
CREATE TABLE IF NOT EXISTS trouble_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open'
        CHECK(status IN ('open', 'in_progress', 'resolved', 'closed')),
    priority TEXT NOT NULL DEFAULT 'medium'
        CHECK(priority IN ('low', 'medium', 'high', 'critical')),
    category TEXT,
    reported_by INTEGER NOT NULL,
    assigned_to INTEGER,
    tool_id INTEGER,                       -- Optional link to affected tool
    press_number INTEGER CHECK(press_number >= 0 AND press_number <= 5),  -- Optional link to affected press (0-5)
    attachments BLOB,                      -- JSON: array of attachment IDs
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    FOREIGN KEY (reported_by) REFERENCES users(id),
    FOREIGN KEY (assigned_to) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (tool_id) REFERENCES tools(id) ON DELETE SET NULL
);
CREATE INDEX idx_trouble_reports_status ON trouble_reports(status);
CREATE INDEX idx_trouble_reports_priority ON trouble_reports(priority);
CREATE INDEX idx_trouble_reports_tool_id ON trouble_reports(tool_id);

-- ============================================================================
-- ATTACHMENTS
-- ============================================================================

-- Attachments table - File attachments for various entities
CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    filepath TEXT NOT NULL UNIQUE,
    mimetype TEXT NOT NULL,
    size INTEGER NOT NULL,                 -- File size in bytes
    entity_type TEXT NOT NULL,             -- 'trouble_report', 'tool', 'note', etc.
    entity_id INTEGER NOT NULL,            -- ID of the related entity
    uploaded_by INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (uploaded_by) REFERENCES users(id)
);
CREATE INDEX idx_attachments_entity ON attachments(entity_type, entity_id);

-- ============================================================================
-- ACTIVITY FEEDS
-- ============================================================================

-- Feeds table - Activity log and audit trail
CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,                    -- Event type
    data BLOB NOT NULL,                    -- JSON: Event-specific data
    user_id INTEGER,                       -- User who triggered the event
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX idx_feeds_type ON feeds(type);
CREATE INDEX idx_feeds_created_at ON feeds(created_at DESC);
CREATE INDEX idx_feeds_user_id ON feeds(user_id);

-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

-- Insert default admin user (password should be changed on first login)
-- Password hash is for 'admin' using bcrypt
INSERT OR IGNORE INTO users (username, password_hash, role) VALUES
    ('admin', '$2a$10$YourHashHere', 'admin');

-- ============================================================================
-- VIEWS (Optional - for reporting)
-- ============================================================================

-- View: Currently active tools on presses
CREATE VIEW IF NOT EXISTS v_active_tools AS
SELECT
    t.id as tool_id,
    t.code as tool_code,
    t.type as tool_type,
    t.position as tool_position,
    t.press as press_number,
    pc.from_date as mounted_at,
    pc.total_cycles,
    pc.partial_cycles
FROM tools t
LEFT JOIN press_cycles pc ON t.id = pc.tool_id AND pc.to_date IS NULL
WHERE t.status = 'active';

-- View: Tool regeneration statistics
CREATE VIEW IF NOT EXISTS v_tool_regeneration_stats AS
SELECT
    t.id as tool_id,
    t.code as tool_code,
    t.type as tool_type,
    COUNT(tr.id) as regeneration_count,
    MAX(tr.regenerated_at) as last_regeneration,
    SUM(tr.cycles_at_regeneration) as total_cycles_all_time
FROM tools t
LEFT JOIN tool_regenerations tr ON t.id = tr.tool_id
GROUP BY t.id, t.code, t.type;

-- View: Press utilization
CREATE VIEW IF NOT EXISTS v_press_utilization AS
WITH press_numbers AS (
    SELECT 0 as number UNION ALL
    SELECT 1 UNION ALL
    SELECT 2 UNION ALL
    SELECT 3 UNION ALL
    SELECT 4 UNION ALL
    SELECT 5
)
SELECT
    pn.number as press_number,
    COUNT(DISTINCT pc.tool_id) as current_tools,
    SUM(CASE WHEN pc.to_date IS NULL THEN 1 ELSE 0 END) as active_tools
FROM press_numbers pn
LEFT JOIN press_cycles pc ON pn.number = pc.press_number
GROUP BY pn.number;

-- ============================================================================
-- TRIGGERS (Optional - for automatic timestamp updates)
-- ============================================================================

-- Update timestamp triggers
CREATE TRIGGER IF NOT EXISTS update_users_timestamp
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_tools_timestamp
AFTER UPDATE ON tools
BEGIN
    UPDATE tools SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_metal_sheets_timestamp
AFTER UPDATE ON metal_sheets
BEGIN
    UPDATE metal_sheets SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_trouble_reports_timestamp
AFTER UPDATE ON trouble_reports
BEGIN
    UPDATE trouble_reports SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_notes_timestamp
AFTER UPDATE ON notes
BEGIN
    UPDATE notes SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
