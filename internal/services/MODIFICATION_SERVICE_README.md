# Modification Service

A comprehensive replacement for the old in-memory mod system, providing centralized database-based modification tracking for all entities in the PG Press system.

## Overview

The Modification Service stores and manages change history for various entities like trouble reports, metal sheets, tools, etc. It provides a centralized way to track who made changes, when they were made, and what the changes were.

## Features

- **Centralized Storage**: All modifications are stored in a single `modifications` table
- **Typed Data**: Support for different data types per entity using generics
- **User Tracking**: Links modifications to users who made the changes
- **Timestamp Tracking**: Precise creation timestamps for all modifications
- **Query Support**: Various query methods (by entity, by user, by date range, etc.)
- **Migration Support**: Tools to migrate from the old mod system
- **Extended Context**: Optional context information (IP address, user agent, etc.)

## Architecture

### Core Components

1. **ModificationService** (`internal/services/modification.go`)
    - Main service for CRUD operations
    - Database management and queries
    - Helper methods for specific entity types

2. **Modification Model** (`pkg/models/modification/modification.go`)
    - Generic modification structure
    - Data marshaling/unmarshaling
    - Validation and utility methods

3. **Data Types** (`pkg/models/modification/data_types.go`)
    - Specific data structures for different entities
    - Action types and context information
    - Extended modification data with metadata

4. **Migration Tools** (`internal/services/modification_migration.go`)
    - Migration from old mod system
    - Verification and cleanup utilities
    - Statistics and reporting

5. **CLI Interface** (`internal/services/modification_cli.go`)
    - Command-line tools for migration management
    - Status checking and verification
    - Statistics and maintenance operations

## Database Schema

```sql
CREATE TABLE modifications (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    data BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
);

CREATE INDEX idx_modifications_entity ON modifications(entity_type, entity_id);
CREATE INDEX idx_modifications_created_at ON modifications(created_at);
CREATE INDEX idx_modifications_user_id ON modifications(user_id);
```

## Usage Examples

### Basic Service Usage

```go
// Initialize the service
modService := services.NewModificationService(db)

// Add a modification for a trouble report
err := modService.AddTroubleReportMod(userID, reportID, troubleReportData)
if err != nil {
    log.Printf("Failed to add modification: %v", err)
}

// Get modification history for an entity
mods, err := modService.List(services.ModificationTypeTroubleReport, reportID, 10, 0)
if err != nil {
    log.Printf("Failed to get modifications: %v", err)
}

// Get latest modification
latest, err := modService.GetLatest(services.ModificationTypeTroubleReport, reportID)
if err != nil {
    log.Printf("Failed to get latest modification: %v", err)
}
```

### Using Extended Data with Context

```go
// Create modification data with context
modData := modification.NewExtendedModificationData(
    modification.TroubleReportModData{
        Title:             "Updated Title",
        Content:           "Updated content",
        LinkedAttachments: []int64{1, 2, 3},
    },
    modification.ActionUpdate,
    "Updated title and content",
).WithIPAddress("192.168.1.100").WithUserAgent("Mozilla/5.0...")

// Add the modification
err := modService.AddTroubleReportMod(userID, reportID, modData)
```

### Integration with Existing Services

```go
// Example: Integrating with trouble report service
func (s *TroubleReport) Update(report *troublereport.TroubleReport, user *user.User) error {
    // Get current report for comparison
    current, err := s.Get(report.ID)
    if err != nil {
        return err
    }

    // Update the report in database
    if err := s.updateInDatabase(report); err != nil {
        return err
    }

    // Record the modification
    modData := modification.NewExtendedModificationData(
        modification.TroubleReportModData{
            Title:             report.Title,
            Content:           report.Content,
            LinkedAttachments: report.LinkedAttachments,
        },
        modification.ActionUpdate,
        s.buildChangeDescription(current, report),
    )

    return s.modifications.AddTroubleReportMod(user.ID, report.ID, modData)
}
```

## Migration from Old System

### Step 1: Check Migration Status

```bash
./app modification status
```

### Step 2: Run Migration

```bash
./app modification migrate
```

### Step 3: Verify Migration

```bash
./app modification verify
```

### Step 4: Cleanup (Optional)

```bash
./app modification cleanup
```

### Migration Process Details

The migration process:

1. **Reads old mod data** from existing tables (trouble_reports, metal_sheets, tools)
2. **Converts old format** to new modification format
3. **Preserves timestamps** from original modifications
4. **Maps users** from old mod entries to new user_id references
5. **Adds context** indicating the data was migrated
6. **Maintains order** of modifications by timestamp

### What Gets Migrated

- **Trouble Reports**: Title, content, linked attachments
- **Metal Sheets**: All dimensional and property data, tool assignments
- **Tools**: Position, format, type, code, regeneration status, press assignments
- **User Information**: Preserves who made each modification
- **Timestamps**: Maintains original modification times

## Entity Types

The service supports the following entity types:

```go
const (
    ModificationTypeTroubleReport ModificationType = "trouble_reports"
    ModificationTypeMetalSheet    ModificationType = "metal_sheets"
    ModificationTypeTool          ModificationType = "tools"
    ModificationTypePressCycle    ModificationType = "press_cycles"
    ModificationTypeUser          ModificationType = "users"
    ModificationTypeNote          ModificationType = "notes"
    ModificationTypeAttachment    ModificationType = "attachments"
)
```

## Data Structures

### Trouble Report Modifications

```go
type TroubleReportModData struct {
    Title             string  `json:"title"`
    Content           string  `json:"content"`
    LinkedAttachments []int64 `json:"linked_attachments"`
}
```

### Metal Sheet Modifications

```go
type MetalSheetModData struct {
    TileHeight  float64 `json:"tile_height"`
    Value       float64 `json:"value"`
    MarkeHeight int     `json:"marke_height"`
    STF         float64 `json:"stf"`
    STFMax      float64 `json:"stf_max"`
    ToolID      *int64  `json:"tool_id"`
    LinkedNotes []int64 `json:"linked_notes"`
}
```

### Tool Modifications

```go
type ToolModData struct {
    Position     ToolPosition `json:"position"`
    Format       ToolFormat   `json:"format"`
    Type         string       `json:"type"`
    Code         string       `json:"code"`
    Regenerating bool         `json:"regenerating"`
    Press        *int         `json:"press"`
    LinkedNotes  []int64      `json:"linked_notes"`
}
```

## API Methods

### Core Methods

- `Add(userID, entityType, entityID, data)` - Add a new modification
- `Get(id)` - Get a specific modification by ID
- `List(entityType, entityID, limit, offset)` - List modifications for an entity
- `Count(entityType, entityID)` - Count modifications for an entity
- `Delete(id)` - Delete a specific modification
- `DeleteAll(entityType, entityID)` - Delete all modifications for an entity

### Query Methods

- `GetLatest(entityType, entityID)` - Get most recent modification
- `GetOldest(entityType, entityID)` - Get oldest modification
- `GetByUser(userID, limit, offset)` - Get modifications by user
- `GetByDateRange(entityType, entityID, from, to)` - Get modifications in date range
- `ListWithUser(entityType, entityID, limit, offset)` - Get modifications with user info

### Helper Methods

- `AddTroubleReportMod(userID, reportID, data)` - Add trouble report modification
- `AddMetalSheetMod(userID, sheetID, data)` - Add metal sheet modification
- `AddToolMod(userID, toolID, data)` - Add tool modification
- `AddPressCycleMod(userID, cycleID, data)` - Add press cycle modification
- `AddUserMod(userID, targetUserID, data)` - Add user modification

## Best Practices

### 1. Always Record Modifications

```go
// ❌ Bad: Update without recording modification
s.updateDatabase(entity)

// ✅ Good: Update and record modification
s.updateDatabase(entity)
s.modifications.AddEntityMod(userID, entityID, modData)
```

### 2. Use Meaningful Descriptions

```go
// ❌ Bad: Generic description
modData := modification.NewExtendedModificationData(data, modification.ActionUpdate, "Updated")

// ✅ Good: Specific description
modData := modification.NewExtendedModificationData(data, modification.ActionUpdate, "Updated title from 'Old' to 'New' and added 2 attachments")
```

### 3. Handle Errors Gracefully

```go
// Record modification but don't fail the operation if modification fails
if err := s.modifications.AddEntityMod(userID, entityID, modData); err != nil {
    logger.Error("Failed to record modification: %v", err)
    // Continue with the operation
}
```

### 4. Use Transactions for Critical Operations

```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// Update entity
if err := updateEntity(tx, entity); err != nil {
    return err
}

// Commit transaction first
if err := tx.Commit(); err != nil {
    return err
}

// Record modification after successful commit
s.modifications.AddEntityMod(userID, entityID, modData)
```

## Performance Considerations

1. **Indexing**: The service creates indexes on commonly queried fields
2. **Pagination**: Use limit/offset for large result sets
3. **Data Size**: JSON data is compressed in the database
4. **Cleanup**: Regularly archive or clean old modifications if needed

## Monitoring and Maintenance

### View Statistics

```bash
./app modification stats
```

### Check System Health

```bash
./app modification status
```

### Monitor Recent Activity

```go
// Get modifications from last 24 hours
yesterday := time.Now().AddDate(0, 0, -1)
mods, err := modService.GetByDateRange(entityType, entityID, yesterday, time.Now())
```

## Error Handling

The service uses standard Go error handling patterns:

- `utils.NewNotFoundError()` for missing entities
- `utils.NewValidationError()` for invalid data
- Standard errors for database and system issues

## Security Considerations

1. **User Validation**: Always validate user permissions before recording modifications
2. **Data Sanitization**: Ensure modification data is properly validated
3. **Audit Trail**: Modifications provide a complete audit trail
4. **Foreign Key Constraints**: User references are enforced at database level

## Future Enhancements

Potential future improvements:

1. **Compression**: Compress modification data for large payloads
2. **Archiving**: Archive old modifications to separate tables
3. **Real-time Notifications**: WebSocket notifications for modification events
4. **Diff Visualization**: Visual diff tools for comparing modifications
5. **Batch Operations**: Bulk modification operations
6. **Export/Import**: Tools for exporting/importing modification history

## Troubleshooting

### Common Issues

1. **Migration Fails**: Check database permissions and foreign key constraints
2. **Large Data**: Consider pagination for large modification histories
3. **Performance**: Ensure proper indexing and query optimization
4. **Storage**: Monitor disk space usage for modification data

### Debug Commands

```bash
# Check migration status
./app modification status

# Verify data integrity
./app modification verify

# View system statistics
./app modification stats
```

For additional support, check the application logs and database query performance.
