-- Migration script to add mods columns, create modifications table, and prepare for data migration
-- This script should be run before using the Go migration tools
--
-- Usage:
-- 1. Run this script to add missing columns and create tables
-- 2. Use: pgpress migration run (to migrate data)
-- 3. Use: pgpress migration verify (to verify migration)
-- 4. Use: pgpress migration cleanup (to remove old columns)

-- Start transaction to ensure atomicity
BEGIN TRANSACTION;

-- Add mods column to trouble_reports table if it doesn't exist
-- This column stores JSON data for old modification tracking
ALTER TABLE trouble_reports ADD COLUMN mods TEXT DEFAULT '[]';

-- Add mods column to metal_sheets table if it doesn't exist
ALTER TABLE metal_sheets ADD COLUMN mods TEXT DEFAULT '[]';

-- Add mods column to tools table if it doesn't exist
ALTER TABLE tools ADD COLUMN mods TEXT DEFAULT '[]';

-- Create the modifications table if it doesn't exist
-- This is the new centralized modification tracking system
CREATE TABLE IF NOT EXISTS modifications (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    data BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_modifications_entity ON modifications(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_modifications_user ON modifications(user_id);
CREATE INDEX IF NOT EXISTS idx_modifications_created_at ON modifications(created_at);

-- Create a view to help with migration verification
CREATE VIEW IF NOT EXISTS migration_status AS
SELECT
    'trouble_reports' as table_name,
    COUNT(*) as total_records,
    SUM(CASE WHEN mods IS NOT NULL AND mods != '[]' AND mods != '' THEN 1 ELSE 0 END) as records_with_mods
FROM trouble_reports
UNION ALL
SELECT
    'metal_sheets' as table_name,
    COUNT(*) as total_records,
    SUM(CASE WHEN mods IS NOT NULL AND mods != '[]' AND mods != '' THEN 1 ELSE 0 END) as records_with_mods
FROM metal_sheets
UNION ALL
SELECT
    'tools' as table_name,
    COUNT(*) as total_records,
    SUM(CASE WHEN mods IS NOT NULL AND mods != '[]' AND mods != '' THEN 1 ELSE 0 END) as records_with_mods
FROM tools;

-- Create a view to show modification statistics
CREATE VIEW IF NOT EXISTS modification_stats AS
SELECT
    entity_type,
    COUNT(*) as total_modifications,
    COUNT(DISTINCT entity_id) as unique_entities,
    MIN(created_at) as earliest_modification,
    MAX(created_at) as latest_modification
FROM modifications
GROUP BY entity_type;

COMMIT;

-- Instructions for completing the migration:
--
-- After running this script, use the following commands:
--
-- 1. Check migration status:
--    pgpress migration status
--
-- 2. Run the data migration:
--    pgpress migration run
--
-- 3. Verify the migration completed successfully:
--    pgpress migration verify
--
-- 4. View migration statistics:
--    pgpress migration stats
--
-- 5. Export migration data (optional backup):
--    pgpress migration export --output migration_backup.json
--
-- 6. Clean up old mods columns (DESTRUCTIVE - only after verification):
--    pgpress migration cleanup
--
-- You can also check the migration status with:
--    SELECT * FROM migration_status;
--    SELECT * FROM modification_stats;

-- Note: The actual data migration from mods columns to the modifications table
-- is handled by the Go migration tool (pgpress migration run) which provides:
-- - Proper JSON parsing and validation
-- - User ID resolution
-- - Timestamp preservation
-- - Error handling and rollback
-- - Progress tracking
-- - Verification capabilities
