# Trouble Reports Modifications System

## Overview

The trouble reports modifications system provides comprehensive version control and rollback functionality for trouble reports. This feature allows administrators to view the complete history of changes made to any trouble report and restore previous versions when needed.

## Features

### 1. Modification Tracking

- **Automatic History**: Every create, update, and attachment change is automatically recorded
- **User Attribution**: All modifications are linked to the user who made the change
- **Timestamping**: Precise timestamps for all modifications using Unix milliseconds
- **Data Preservation**: Complete snapshots of title, content, and linked attachments

### 2. Modifications Viewer

- **Chronological Display**: Modifications shown in reverse chronological order (newest first)
- **Current Version Highlight**: Latest version clearly marked with green styling
- **Detailed View**: Each modification shows user, timestamp, and complete data snapshot
- **Responsive Design**: Mobile-friendly layout using ui.min.css design system with proper spacing and typography

### 3. Rollback System

- **Administrator Only**: Rollback functionality restricted to users with admin privileges
- **Confirmation Required**: JavaScript confirmation dialog prevents accidental rollbacks
- **HTMX Integration**: Smooth, AJAX-based rollbacks without full page reloads
- **Success Feedback**: Visual confirmation with automatic page refresh after rollback

## Technical Architecture

### Database Schema

#### Modifications Table

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
```

#### Indexes

- `idx_modifications_entity` - Fast lookups by entity type and ID
- `idx_modifications_created_at` - Chronological ordering
- `idx_modifications_user_id` - User-based queries

### Data Structure

#### TroubleReportModData

```go
type TroubleReportModData struct {
    Title             string  `json:"title"`
    Content           string  `json:"content"`
    LinkedAttachments []int64 `json:"linked_attachments"`
}
```

## Implementation Details

### Service Layer

#### ModificationService

Located in `internal/services/modification.go`

**Key Methods:**

- `AddTroubleReportMod(userID, reportID int64, data interface{})` - Record new modification
- `ListAll(entityType, entityID int64)` - Get all modifications for an entity
- `ListWithUser(entityType, entityID int64, limit, offset int)` - Get modifications with user details

#### Integration Points

- **TroubleReport Service**: Automatically saves modifications on Add/Update operations
- **Database Layer**: Utilizes existing database connection and transaction management
- **User Authentication**: Leverages existing user system for attribution

### Web Layer

#### Routes

- `GET /trouble-reports/modifications/:id` - Display modifications page
- `POST /htmx/trouble-reports/rollback?id=:id` - Perform rollback via HTMX

#### Handlers

- **handleModificationsGET**: Main page handler with modification history
- **handleRollbackPOST**: HTMX rollback handler with admin security checks

### Template System

#### Generic Modifications Template

Located in `internal/web/templates/modificationspage/page.templ`

**Features:**

- Generic template supporting any modification type `T`
- Breadcrumb navigation with ui.min.css styling
- Empty state handling with proper iconography
- Current vs. previous version distinction using color-coded cards
- Help text and usage instructions with consistent design system

## Security Considerations

### Access Control

- **View Permissions**: Any authenticated user can view modification history
- **Rollback Permissions**: Only administrators can perform rollbacks
- **User Validation**: All operations validate user context and permissions

### Data Integrity

- **Immutable History**: Modifications are never deleted, only new ones added
- **Rollback Tracking**: Rollbacks create new modification entries
- **Atomic Operations**: Database operations use transactions where appropriate

### Input Validation

- **Parameter Validation**: All URL parameters properly parsed and validated
- **Data Sanitization**: Modification data properly escaped in templates
- **Error Handling**: Comprehensive error handling with proper HTTP status codes

## Usage Instructions

### Viewing Modifications

1. Navigate to any trouble report
2. Click the "Modifications" link or button (if available in UI)
3. Or directly access: `/trouble-reports/modifications/{report_id}`

### Performing Rollbacks (Admin Only)

1. Access the modifications page for a trouble report
2. Locate the desired previous version
3. Click the "Rollback" button next to that modification
4. Confirm the action in the dialog that appears
5. Wait for the success message and automatic page refresh

### Understanding the Interface

#### Current Version

- Highlighted with success styling and "Current Version" badge using ui.min.css
- Shows the most recent state of the trouble report
- Distinguished with left border accent color

#### Previous Versions

- Listed chronologically (newest to oldest) in card-based layout
- Each entry shows:
    - User who made the modification
    - Timestamp of the change
    - Version number badge
    - Complete data snapshot (title, content, attachments count)
    - Rollback button (admin only) with proper button styling

## Maintenance and Monitoring

### Performance Considerations

- **Pagination**: Modifications list supports pagination for large histories
- **Indexing**: Proper database indexes ensure fast queries
- **Memory Usage**: JSON data stored efficiently in BLOB fields

### Cleanup and Archival

- **Data Retention**: No automatic cleanup implemented (consider for future)
- **Storage Growth**: Monitor modifications table size over time
- **Archival Strategy**: Consider implementing archival for very old modifications

## Future Enhancements

### Potential Improvements

1. **Diff Viewer**: Show what changed between versions
2. **Batch Rollback**: Roll back multiple trouble reports at once
3. **Modification Comments**: Allow users to add notes when making changes
4. **Export History**: Download modification history as CSV/PDF
5. **Notification System**: Notify relevant users of rollbacks
6. **Audit Reports**: Administrative reports on modification patterns

### Technical Debt

1. **Generic Templates**: Make modification templates more reusable across entities
2. **Caching**: Implement caching for frequently accessed modification histories
3. **API Endpoints**: Create REST API endpoints for programmatic access
4. **CSS Framework**: Fully migrate all hardcoded styles to use ui.min.css variables and classes

## Troubleshooting

### Common Issues

#### "No Modifications Found"

- **Cause**: New trouble reports or reports created before modification tracking
- **Solution**: Normal behavior, modifications will appear after updates

#### Rollback Button Not Visible

- **Cause**: User lacks administrator privileges
- **Solution**: Contact system administrator for permission elevation

#### Rollback Failed

- **Cause**: Various - check server logs for specific error
- **Common Fixes**:
    - Verify trouble report still exists
    - Check database connectivity
    - Ensure user has valid session

### Logging

- **Location**: Standard application logs
- **Levels**: INFO for successful operations, ERROR for failures
- **Context**: All log entries include user information and trouble report IDs

## Configuration

### Environment Variables

- `ADMINS`: Comma-separated list of Telegram IDs with admin privileges
- Standard database configuration applies

### Database Configuration

- No specific configuration required beyond standard setup
- Modifications table created automatically on first use

## Testing

### Manual Testing Checklist

- [ ] Create trouble report and verify modification recorded
- [ ] Update trouble report and verify new modification
- [ ] View modifications page as regular user
- [ ] View modifications page as administrator
- [ ] Perform rollback and verify data restored
- [ ] Verify rollback creates new modification entry
- [ ] Test with trouble reports having many modifications
- [ ] Test with trouble reports having attachments

### Test Data

Consider creating test trouble reports with multiple modifications to verify the complete workflow.

## Design System Integration

### UI Framework

The modifications system uses **ui.min.css** as the primary design framework, providing:

- **Consistent Styling**: All components use ui.min.css classes and CSS variables
- **Dark/Light Theme Support**: Automatic theme switching based on user preferences
- **Responsive Design**: Built-in responsive utilities and mobile-friendly layouts
- **Accessibility**: Proper ARIA labels and semantic HTML structure

### Key UI Components

- **Cards**: Used for modification entries and page sections
- **Badges**: Version indicators and status labels
- **Buttons**: Rollback actions with proper hover and focus states
- **Breadcrumbs**: Navigation using ui.min.css navigation patterns
- **Typography**: Consistent text sizing and weight using ui.min.css font utilities

### CSS Variables Used

- `--ui-primary`: Primary accent color for important elements
- `--ui-success`: Success states and current version highlighting
- `--ui-muted`: Secondary text and subtle elements
- `--ui-spacing`: Consistent spacing throughout the interface
- `--ui-radius`: Consistent border radius for cards and buttons
