# Notes System Documentation

The PG Press Notes System provides comprehensive note management functionality for documenting important information about tools and presses.

## Overview

The notes system allows users to create, manage, and link notes to tools and presses with different priority levels. Notes help track maintenance issues, operational instructions, and other important information.

## Features

### Note Priority Levels

Notes support three priority levels:

- **INFO** (Level 0): General information and documentation
- **ATTENTION** (Level 1): Important notices that require attention
- **BROKEN** (Level 2): Critical issues indicating equipment problems

### Note Linking System

Notes can be linked to:

1. **Individual Tools**: Notes appear only for the specific tool
   - Format: `tool_{ID}` (e.g., `tool_123`)
   - Use case: Tool-specific maintenance notes, calibration info

2. **Entire Presses**: Notes appear for all tools currently on that press
   - Format: `press_{ID}` (e.g., `press_5`)
   - Use case: Press-wide issues, operational procedures

### User Interface

#### Notes Management Page (`/notes`)

- **Comprehensive View**: Display all notes with filtering and sorting
- **Priority Filtering**: Filter notes by INFO, ATTENTION, or BROKEN levels
- **Real-time Search**: JavaScript-based filtering without page reloads
- **Tool Relationships**: See which tools are linked to each note
- **CRUD Operations**: Create, edit, and delete notes with confirmation dialogs

#### Tool Pages (`/tools/{id}`)

- **Tool-Specific Notes**: Display notes linked to the individual tool
- **Add Note Button**: Quick access to create notes for the tool
- **Edit Capabilities**: Modify existing notes directly from the tool page

#### Press Pages (`/tools/press/{id}`)

- **Press-Wide Notes**: Show all notes for tools currently on the press
- **Contextual Display**: Notes show which tools they affect
- **Add Press Note**: Create notes that apply to all tools on the press

## Technical Implementation

### Database Schema

```sql
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER NOT NULL,
    level INTEGER NOT NULL,        -- 0=INFO, 1=ATTENTION, 2=BROKEN
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY("id" AUTOINCREMENT)
);
```

### Data Models

#### Note Model

```go
type Note struct {
    ID        int64     `json:"id"`
    Level     Level     `json:"level"`     // 0, 1, or 2
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### Tool-Note Linking

Tools store note relationships in their `LinkedNotes` field:

```go
type Tool struct {
    // ... other fields
    LinkedNotes []int64 `json:"notes"` // Note IDs linked to this tool
}
```

### API Endpoints

#### Notes Management

- `GET /notes` - Notes management page
- `GET /htmx/notes/edit` - New note dialog
- `GET /htmx/notes/edit?id={id}` - Edit note dialog
- `POST /htmx/notes/edit` - Create note
- `PUT /htmx/notes/edit?id={id}` - Update note
- `DELETE /htmx/notes/delete?id={id}` - Delete note

#### Tool Notes

- `GET /htmx/tools/notes?tool_id={id}` - Get notes for specific tool

#### Press Notes

- `GET /htmx/tools/press/{id}/notes` - Get notes for all tools on press

### Linking Logic

#### Creating Notes with Links

When creating a note with `link_to_tables` parameter:

1. **Tool Linking** (`tool_{id}`):
   - Add note ID to tool's `LinkedNotes` array
   - Update tool record in database

2. **Press Linking** (`press_{id}`):
   - Find all active tools on the press
   - Add note ID to each tool's `LinkedNotes` array
   - Update all affected tool records

#### Updating Note Links

When updating notes:

1. Remove note ID from previously linked tools
2. Add note ID to newly linked tools
3. Handle press linking by updating all tools on the press

#### Deleting Notes

When deleting notes:

1. Find all tools with the note ID in their `LinkedNotes`
2. Remove note ID from all affected tools
3. Delete the note record
4. Update tool records

## Usage Examples

### Creating a Tool-Specific Note

```http
POST /htmx/notes/edit?link_to_tables=tool_123
Content-Type: application/x-www-form-urlencoded

level=1&content=Requires calibration every 100 cycles&linked_tables=tool_123
```

### Creating a Press-Wide Note

```http
POST /htmx/notes/edit?link_to_tables=press_5
Content-Type: application/x-www-form-urlencoded

level=2&content=Press hydraulics need maintenance&linked_tables=press_5
```

### Updating a Note

```http
PUT /htmx/notes/edit?id=42&link_to_tables=tool_123
Content-Type: application/x-www-form-urlencoded

level=0&content=Calibration completed - normal operation&linked_tables=tool_123
```

## Integration Points

### Feed System

All note operations automatically create feed entries:

- Note creation: "Neue Notiz erstellt"
- Note updates: "Notiz aktualisiert"
- Note deletion: "Notiz gelöscht"

### HTMX Integration

The system uses HTMX for dynamic updates:

- `HX-Trigger: noteCreated, pageLoaded` - Refresh sections after creation
- `HX-Trigger: noteUpdated, pageLoaded` - Refresh sections after updates
- `HX-Trigger: noteDeleted, pageLoaded` - Refresh sections after deletion

### Real-time Updates

Notes sections automatically reload when:

- New notes are created
- Existing notes are modified
- Notes are deleted
- Page loads or reloads occur

## Security Considerations

### User Authentication

- All note operations require authenticated user session
- User information from context used for audit trails and feed entries

### Input Validation

- Note content is required and validated
- Priority levels restricted to valid values (0, 1, 2)
- Table references validated for format and existence

### Data Integrity

- Automatic cleanup of orphaned note references
- Transactional updates for consistent linking
- Error handling with graceful degradation

## Performance Considerations

### Database Queries

- Indexed queries for note retrieval
- Batch operations for multi-tool updates
- Optimized JOIN queries for tool-note relationships

### Frontend Optimization

- Client-side filtering reduces server requests
- HTMX partial updates minimize page reloads
- Lazy loading for large note lists

## Troubleshooting

### Common Issues

**Notes not appearing on tool pages:**

- Check if note is properly linked to tool in `LinkedNotes` array
- Verify tool and note IDs are correct
- Check database consistency

**Press notes not showing:**

- Ensure tools are properly assigned to press
- Verify press number is valid (0-5)
- Check if tools have press assignment

**Linking failures:**

- Tool must exist before linking notes
- Press must have active tools for linking
- Database constraints prevent invalid references

### Debug Information

Enable debug logging to see:

- Note linking operations
- Tool updates during note operations
- Press-to-tool relationship resolution

## Future Enhancements

### Planned Features

- Note categories/tags for better organization
- Batch operations for multiple notes
- Note templates for common scenarios
- Advanced search and filtering options
- Note history and versioning
- Email notifications for critical notes

### Integration Opportunities

- Export notes to maintenance systems
- Import notes from external sources
- Integration with calendar for scheduled notes
- Mobile app support for field notes
