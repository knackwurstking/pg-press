# Notes System Documentation

The PG Press Notes System provides comprehensive note management functionality for documenting important information about tools and presses.

## Overview

The notes system allows users to create, manage, and link notes to any entity using a generic linking system. Notes help track maintenance issues, operational instructions, and other important information across tools, presses, and other system components.

## Features

### Note Priority Levels

Notes support three priority levels:

- **INFO** (Level 0): General information and documentation
- **ATTENTION** (Level 1): Important notices that require attention
- **BROKEN** (Level 2): Critical issues indicating equipment problems

### Note Linking System

Notes can be linked to any entity using a flexible string-based system:

1. **Individual Tools**: Notes specific to a single tool
   - Format: `tool_{ID}` (e.g., `tool_123`)
   - Use case: Tool-specific maintenance notes, calibration info

2. **Entire Presses**: Notes specific to a press
   - Format: `press_{ID}` (e.g., `press_5`)
   - Use case: Press-wide issues, operational procedures

3. **Any Other Entity**: The system supports linking to any entity type
   - Format: `{type}_{ID}` (e.g., `machine_42`, `line_7`)
   - Use case: Extensible for future system components

### User Interface

#### Notes Management Page (`/notes`)

- **Comprehensive View**: Display all notes with filtering and sorting
- **Priority Filtering**: Filter notes by INFO, ATTENTION, or BROKEN levels
- **Real-time Search**: JavaScript-based filtering without page reloads
- **Tool Relationships**: See which tools are linked to each note
- **CRUD Operations**: Create, edit, and delete notes with confirmation dialogs

#### Tool Pages (`/tools/{id}`)

- **Tool-Specific Notes**: Display notes directly linked to the individual tool
- **Add Note Button**: Quick access to create notes for the tool
- **Edit Capabilities**: Modify existing notes directly from the tool page

#### Press Pages (`/tools/press/{id}`)

- **Press-Specific Notes**: Show notes directly linked to the press
- **Contextual Display**: Notes are specific to the press, not dependent on current tools
- **Add Press Note**: Create notes that stay with the press regardless of tool changes

## Technical Implementation

### Database Schema

```sql
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER NOT NULL,
    level INTEGER NOT NULL,        -- 0=INFO, 1=ATTENTION, 2=BROKEN
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    linked TEXT,                   -- Generic entity link (e.g., "tool_123", "press_5")
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
    Linked    string    `json:"linked,omitempty"` // Generic entity link
}
```

#### Generic Linking System

Notes use a simple string field to link to any entity:

- `"tool_123"` - Links to tool with ID 123
- `"press_5"` - Links to press number 5
- `"machine_42"` - Could link to any future entity type
- `""` - Empty string for unlinked notes

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

- `GET /htmx/tools/press/{id}/notes` - Get notes for specific press

### Linking Logic

#### Creating Notes with Links

When creating a note with `link_to_tables` parameter:

1. **Direct Entity Linking**:
   - Set note's `Linked` field to the entity reference (e.g., `"tool_123"`)
   - Store note record in database
   - No additional table updates required

#### Updating Note Links

When updating notes:

1. Simply update the note's `Linked` field to new entity reference
2. No complex relationship management needed
3. Single database operation

#### Deleting Notes

When deleting notes:

1. Delete the note record directly
2. No cleanup of related tables required
3. Linking is self-contained within note record

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

- Indexed queries for note retrieval by entity
- Simple WHERE clauses on `linked` field
- No complex JOIN operations required

### Frontend Optimization

- Client-side filtering reduces server requests
- HTMX partial updates minimize page reloads
- Lazy loading for large note lists

## Troubleshooting

### Common Issues

**Notes not appearing on tool pages:**

- Check if note's `linked` field equals `"tool_{id}"`
- Verify tool and note IDs are correct
- Check database query for proper filtering

**Press notes not showing:**

- Check if note's `linked` field equals `"press_{number}"`
- Verify press number is valid (0-5)
- Ensure press notes query is working correctly

**Linking failures:**

- Verify `linked` field format is correct
- Check for typos in entity references
- Ensure proper form data submission

### Debug Information

Enable debug logging to see:

- Note creation and linking operations
- Entity reference resolution
- Simple linked field queries

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
