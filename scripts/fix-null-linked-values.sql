-- Quick fix for NULL linked values in notes table
-- This SQL script converts NULL values in the linked column to empty strings
-- to prevent scanning errors in the Go application

-- Update NULL linked values to empty strings
UPDATE notes
SET linked = ''
WHERE linked IS NULL;

-- Verify the fix
SELECT
    COUNT(*) as total_notes,
    COUNT(linked) as notes_with_linked,
    SUM(CASE WHEN linked = '' THEN 1 ELSE 0 END) as empty_linked,
    SUM(CASE WHEN linked IS NULL THEN 1 ELSE 0 END) as null_linked
FROM notes;

-- Show distribution of linked values
SELECT
    CASE
        WHEN linked IS NULL THEN 'NULL'
        WHEN linked = '' THEN 'Empty'
        WHEN linked LIKE 'tool_%' THEN 'Tool Links'
        WHEN linked LIKE 'press_%' THEN 'Press Links'
        ELSE 'Other'
    END as link_type,
    COUNT(*) as count
FROM notes
GROUP BY
    CASE
        WHEN linked IS NULL THEN 'NULL'
        WHEN linked = '' THEN 'Empty'
        WHEN linked LIKE 'tool_%' THEN 'Tool Links'
        WHEN linked LIKE 'press_%' THEN 'Press Links'
        ELSE 'Other'
    END
ORDER BY count DESC;
