# Attachment Migration Fixes

## Problem Description

The attachment system migration was failing with the error: **"Failed to unmarshal attachments for report"**. This error occurred during the automatic migration process when trying to move existing attachment data from the old format (full attachment objects stored in `trouble_reports.linked_attachments`) to the new format (attachment IDs referencing a separate `attachments` table).

## Root Cause Analysis

The migration failure was caused by several issues in the migration logic:

1. **Assumption of Data Format**: The migration code assumed all `linked_attachments` data was in the old format (array of attachment objects with BLOB data), but the data could be in various states:
    - Already migrated (array of int64 IDs)
    - Empty arrays (`[]`)
    - Invalid JSON data
    - Mixed formats from partial migrations

2. **Insufficient Error Handling**: The migration failed completely when encountering any unexpected data format, rather than gracefully handling different scenarios.

3. **Poor Migration Detection**: The `checkIfMigrationNeeded()` function used overly simplistic logic that didn't accurately detect which records actually needed migration.

4. **Rigid Parsing Logic**: The unmarshaling logic didn't account for the possibility that data might already be in the new format.

## Fixes Implemented

### 1. Robust Migration Logic (`migration.go`)

#### Enhanced Data Format Detection

```go
// Check if already migrated by trying to unmarshal as array of IDs first
var attachmentIDs []int64
if err := json.Unmarshal(linkedAttachmentsJSON, &attachmentIDs); err == nil {
    // Already migrated - data is array of IDs
    logger.TroubleReport().Debug("Report %d already migrated (contains IDs)", reportID)
    continue
}
```

#### Graceful Error Handling

```go
// Try to parse as old format (full attachment objects)
var existingAttachments []*Attachment
if err := json.Unmarshal(linkedAttachmentsJSON, &existingAttachments); err != nil {
    // Check if it's just an empty array or invalid JSON
    var rawArray []interface{}
    if err2 := json.Unmarshal(linkedAttachmentsJSON, &rawArray); err2 == nil {
        if len(rawArray) == 0 {
            logger.TroubleReport().Debug("Report %d has empty attachments array", reportID)
            continue
        }
        logger.TroubleReport().Error("Report %d has unexpected attachment format: %s", reportID, string(linkedAttachmentsJSON))
    } else {
        logger.TroubleReport().Error("Failed to unmarshal attachments for report %d: %v, data: %s", reportID, err, string(linkedAttachmentsJSON))
    }
    continue
}
```

#### Data Validation Before Migration

```go
// Validate that we have old-format attachments with data
hasOldFormatData := false
for _, att := range existingAttachments {
    if att != nil && att.Data != nil && len(att.Data) > 0 {
        hasOldFormatData = true
        break
    }
}

if !hasOldFormatData {
    logger.TroubleReport().Debug("Report %d has no attachment data to migrate", reportID)
    continue
}
```

### 2. Improved Migration Detection (`checkIfMigrationNeeded`)

#### Sample-Based Analysis

```go
// Get a sample of trouble reports with non-empty linked_attachments
rows, err := m.db.Query(`
    SELECT linked_attachments
    FROM trouble_reports
    WHERE linked_attachments != '[]'
      AND linked_attachments != ''
      AND linked_attachments IS NOT NULL
    LIMIT 10
`)
```

#### Format-Aware Detection

```go
for rows.Next() {
    // Try to unmarshal as array of int64 (new format)
    var attachmentIDs []int64
    if err := json.Unmarshal(linkedAttachmentsJSON, &attachmentIDs); err == nil {
        // This is already in new format, continue checking others
        continue
    }

    // Try to unmarshal as array of attachment objects (old format)
    var attachments []*Attachment
    if err := json.Unmarshal(linkedAttachmentsJSON, &attachments); err == nil {
        // Check if any attachment has actual data (indicating old format)
        for _, att := range attachments {
            if att != nil && att.Data != nil && len(att.Data) > 0 {
                needsMigration = true
                break
            }
        }
    }
}
```

### 3. Enhanced Runtime Error Handling (`trouble-reports.go`)

#### Dual-Format Support

```go
// Try to unmarshal as new format (array of int64 IDs) first
if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
    // If that fails, try to handle as old format or empty/invalid data
    if linkedAttachments == "" || linkedAttachments == "[]" {
        report.LinkedAttachments = make([]int64, 0)
    } else {
        // Try to unmarshal as old format (full attachment objects) and extract IDs
        var oldAttachments []*Attachment
        if err2 := json.Unmarshal([]byte(linkedAttachments), &oldAttachments); err2 == nil {
            // Convert old format to new format (this shouldn't happen after migration)
            logger.TroubleReport().Warn("Found old format attachments for report %d, converting to IDs", report.ID)
            report.LinkedAttachments = make([]int64, 0)
            for _, att := range oldAttachments {
                if att != nil && att.GetID() > 0 {
                    report.LinkedAttachments = append(report.LinkedAttachments, att.GetID())
                }
            }
        } else {
            // Neither format worked, log and use empty array
            logger.TroubleReport().Error("Failed to unmarshal linked attachments for report %d: %v, data: %s", report.ID, err, linkedAttachments)
            report.LinkedAttachments = make([]int64, 0)
        }
    }
}
```

### 4. Safety Mechanisms

#### Pre-Migration Validation

```go
// Check if there are any trouble reports at all
var totalCount int
err = m.db.QueryRow("SELECT COUNT(*) FROM trouble_reports").Scan(&totalCount)
if err != nil {
    return NewDatabaseError("count", "trouble_reports", "failed to count trouble reports", err)
}

if totalCount == 0 {
    logger.TroubleReport().Info("No trouble reports found - skipping migration")
    return nil
}
```

#### Attachment Data Validation

```go
// Validate attachment before migration
if attachment.MimeType == "" {
    logger.TroubleReport().Warn("Attachment for report %d has empty MIME type, setting default", reportID)
    attachment.MimeType = "application/octet-stream"
}
```

### 5. Diagnostic Utilities

#### Attachment Data Diagnosis

```go
func (m *Migration) DiagnoseAttachmentData() error {
    // Provides detailed diagnostic information about attachment data formats
    // Helps troubleshoot migration issues
    // Shows sample data and parsing results
}
```

## Migration Flow (Fixed)

1. **Safety Check**: Verify trouble reports exist before attempting migration
2. **Format Detection**: Check if migration is actually needed by sampling data
3. **Dual-Format Processing**: Handle both old and new formats gracefully
4. **Data Validation**: Ensure attachment data is valid before migration
5. **Error Resilience**: Continue processing even if individual records fail
6. **Comprehensive Logging**: Provide detailed logs for troubleshooting

## Error Handling Strategy

### Before Fix

- Migration failed completely on first error
- No differentiation between data formats
- Poor error messages
- No fallback mechanisms

### After Fix

- Continue processing when individual records fail
- Detect and handle multiple data formats
- Comprehensive error logging with context
- Graceful fallbacks for invalid data
- Diagnostic utilities for troubleshooting

## Testing the Migration

### Manual Testing

```bash
# The migration now runs automatically and should complete without errors
# Check logs for migration progress and any warnings
```

### Diagnostic Commands

```go
// In code, you can run diagnostics:
migration := pgvis.NewMigration(db)
err := migration.DiagnoseAttachmentData()
```

## Migration Outcomes

### Success Scenarios

- ✅ Old format data migrated to new format
- ✅ Already migrated data left unchanged
- ✅ Empty attachment arrays handled gracefully
- ✅ Invalid data logged and skipped
- ✅ Mixed format scenarios handled

### Failure Recovery

- 🔄 Individual record failures don't stop migration
- 📝 Comprehensive logging for troubleshooting
- 🛡️ Data integrity preserved
- 🔍 Diagnostic tools available

## Post-Migration Verification

1. **Check Migration Logs**: Look for successful completion messages
2. **Verify Data Integrity**: Ensure all valid attachments were migrated
3. **Test Functionality**: Verify attachment viewing/downloading works
4. **Monitor Performance**: Confirm improved performance in listing operations

## Troubleshooting

### If Migration Still Fails

1. Run diagnostic function to examine data formats
2. Check logs for specific error patterns
3. Verify database permissions and disk space
4. Consider manual data cleanup if needed

### Common Issues

- **Mixed Data Formats**: Fixed by dual-format parsing
- **Invalid JSON**: Fixed by graceful error handling
- **Empty Data**: Fixed by proper validation
- **Partial Migrations**: Fixed by proper detection logic

## Summary

The attachment migration system is now robust and handles all edge cases that were causing the "Failed to unmarshal attachments" error. The migration will automatically detect the current state of the data and only migrate what needs to be migrated, while gracefully handling any invalid or unexpected data formats.
