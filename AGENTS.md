# AGENTS.md

This document provides coding conventions and guidance for agentic coding agents working on the pg-press codebase.

## Project Overview

**pg-press** is a comprehensive tool and press management system for industrial manufacturing environments. The application manages:

- **Tools**: Individual punch tools with metadata (dimensions, cycles, type)
- **Presses**: Manufacturing machines (SACMI, SITI) with upper/lower tool slots
- **Cycles**: Production records tracking tool usage over time
- **Metal Sheets**: Tool-specific metal sheet configurations (upper/lower)
- **Notes**: Tool/Press annotations with levels (normal, info, attention, broken)
- **Trouble Reports**: Issue documentation with markdown support and attachments
- **Users & Sessions**: API key-based authentication system

### Key Features
- Multi-database SQLite architecture (tools, presses, notes, users, reports)
- PDF generation for trouble reports and cycle summaries
- API key authentication with configurable access
- German localization for all display names

## Build/Lint/Test Commands

```bash
# Initialize dependencies and generate templ templates
make init

# Build the binary
make build

# Run all tests with verbose output
go test -v ./...

# Run linter
make lint

# Run development server with auto-reload (requires gow)
make dev

# Generate templ template code
make generate
```

To run a single test file: `go test -v ./path/to/file_test.go`
To run a single test function: `go test -v -run TestFunctionName ./path/to/`

## Code Style Guidelines

### General Structure
- Use lowercase package names matching directory structure
- Separate code into logical packages: `cmd/`, `internal/`, `scripts/`
- Keep files focused: entity definitions, handlers, database access, URL builders
- Use `internal/` for private packages that shouldn't be imported externally

### Imports
Group imports in order:
1. Standard library
2. Third-party dependencies (github.com/...)
3. Internal packages (alphabetically by module: `assets`, `db`, `errors`, `env`, `logger`, `shared`, `urlb`, `utils`)

Example:
```go
import (
    "fmt"

   "github.com/labstack/echo/v4"

    "github.com/knackwurstking/pg-press/internal/shared"
)
```

### Naming Conventions
- **Types (structs, interfaces)**: PascalCase (`User`, `Cycle`, `Tool`)
- **Variables/Functions**: camelCase (`userId`, `getToolByID`)
- **Constants**: PascalCase with prefixes for categories (`UserNameMinLength`, `ToolCyclesWarning`)
- **File names**: kebab-case for templates (`tool-page.templ`), snake_case for Go files
- **Package names**: single word, lowercase (`shared`, `utils`, `db`)

### Error Handling
- Custom error types in `internal/errors/`: `ValidationError`, `NotFoundError`, `AuthorizationError`, `ExistsError`
- Wrap errors with context using `errors.Wrap(err, "context")`
- HTTP errors: `*errors.HTTPError` with `.Echo()` conversion
- Always check for nil before accessing error values
- Use specific error types for database operations (sql.ErrNoRows, etc.)

### Entity Pattern
All entities implement the `Entity[T]` interface:
```go
type Entity[T any] interface {
    Validate() *errors.ValidationError
    Clone() T
    String() string
}
```
- Add validation logic in `Validate()` method
- Implement `Clone()` for deep copies
- Format `String()` for debugging (include all fields)

### Type Safety
- Use custom types instead of primitives: `EntityID`, `UnixMilli`, `TelegramID`, `MachineType`
- Define type constants: `SlotUpper`, `SlotLower`, `MachineTypeSACMI`
- Add methods on custom types for formatting: `German() string`

### Templ Templates
- Use `templ` package for HTML generation
- Template files: `/internal/handlers/*/templates/*.templ`
- Auto-generate Go code with `make generate`

### Logging
- Use structured logging via `logger.New("module-name")`
- Log in development with verbose mode: `export VERBOSE=true`
- Use placeholders for dynamic values: `log.Debug("message %s", value)`

## Database Schema

### Tables Overview
The application uses multiple SQLite databases with WAL mode enabled.

#### tools.db
- **tools**: Main tool inventory
  - `id`, `width`, `height`, `position` (Slot: upper/lower/cassette)
  - `type`, `code`, `cycles_offset`, `cycles`, `is_dead`, `cassette`
  - `min_thickness`, `max_thickness`

- **cycles**: Production cycle records
  - `id`, `tool_id`, `press_id`
  - `press_cycles`, `partial_cycles`
  - `start`, `stop` (UnixMilli timestamps)

- **notes`: Tool/Press annotations
  - `id`, `level` (normal/info/attention/broken)
  - `content`, `created_at`
  - `linked` (JSON string with linked entity info)

- **metal_sheets**: Tool metal sheet configurations
  - `id`, `tool_id`, `tile_height`, `value`
  - `type` (upper/lower)
  - Additional fields for lower: `marke_height`, `stf`, `stf_max`, `identifier`

- **tool_regenerations**: Tool regeneration tracking
  - `id`, `tool_id`, `start`, `stop`

#### presses.db
- **presses**: Machine inventory
  - `id`, `number` (PressNumber)
  - `type` (MachineType: SACMI/SITI), `code`
  - `slot_up`, `slot_down` (referencing tools)
  - `cycles_offset`

#### users.db
- **users**: User accounts with API keys
  - `id` (TelegramID), `name`, `api_key`
  - Minimum API key length: 32 characters

- **sessions**: User sessions
  - `id`

#### reports.db
- **trouble_reports**: Issue documentation
  - `id`, `title`, `content`
  - `linked_attachments` (JSON array)
  - `use_markdown` (boolean)

### Database Connection
Opened in `internal/db/db.go`:
```go
sql.Open("sqlite3", "file:db/tools.db?_foreign_keys=true&journal=WAL&synchronous=1")
```
- Connection pooling: `SetMaxOpenConns(10)`, `SetMaxIdleConns(5)`
- WAL mode enabled for concurrent access

## Entity Types Reference

### Tool
```go
type Tool struct {
    ID           EntityID
    Width        int
    Height       int
    Position     Slot          // upper/lower/cassette
    Type         string        // tool type identifier
    Code         string        // unique code
    CyclesOffset int64
    Cycles       int64         // Current cycle count
    IsDead       bool          // Deprecated flag
    Cassette     EntityID      // Parent cassette (if applicable)
    MinThickness float32
    MaxThickness float32
}
```
- Constants: `ToolCyclesWarning`, `ToolCyclesError`

### Cycle
```go
type Cycle struct {
    ID          EntityID
    ToolID      EntityID
    PressID     EntityID
    PressCycles int64   // Total presses in cycle
    PartialCycles int64 // Partial presses
    Start       UnixMilli
    Stop        UnixMilli
}
```

### Press
```go
type Press struct {
    ID           EntityID
    Number       PressNumber
    Type         MachineType  // SACMI/SITI
    Code         string
    SlotUp       EntityID     // Upper tool reference
    SlotDown     EntityID     // Lower tool reference
    CyclesOffset int64
}
```

### Metal Sheets
Base type with specialized variants:

```go
type BaseMetalSheet struct {
    ID        EntityID
    ToolID    EntityID
    TileHeight float64
    Value     float64
}

type UpperMetalSheet struct {
    BaseMetalSheet
    // No additional fields
}

type LowerMetalSheet struct {
    BaseMetalSheet
    MarkeHeight int
    STF         float64
    STFMax      float64
    Identifier  MachineType
}
```

### Note
```go
type Note struct {
    ID        EntityID
    Level     NoteLevel   // normal/info/attention/broken
    Content   string
    CreatedAt UnixMilli
    Linked    string      // JSON with linked entity info
}
```

### Trouble Report
```go
type TroubleReport struct {
    ID                EntityID
    Title             string
    Content           string
    LinkedAttachments []string  // JSON array of file paths
    UseMarkdown       bool
}
```

### Tool Regeneration
```go
type ToolRegeneration struct {
    ID      EntityID
    ToolID  EntityID
    Start   UnixMilli
    Stop    UnixMilli
}
```

## Common Entity Methods

### Validate()
All entities implement `Validate() *errors.ValidationError` for data validation.

### German()
Returns localized display names:
```go
func (t Tool) German() string {
    return fmt.Sprintf("%s (%d x %d)", t.Code, t.Width, t.Height)
}
```

### Clone()
Returns deep copies for safe data manipulation.

### String()
Debug formatting including all fields.

## File Organization

```
cmd/pg-press/          # CLI entry points
  main.go             # Command registration
  commands-*.go       # Individual commands (user, cookies, tools, server)
  middleware.go       # Auth and error handling
internal/
  assets/             # Static file bundling (templ)
  components/         # Reusable UI components
  db/                 # Database access layer
    db.go            # Database initialization
    tool_*.go        # Tool-related queries
    press_*.go       # Press-related queries
    note_*.go        # Note operations
    user_*.go        # User/session queries
    reports_*.go     # Trouble report operations
  errors/            # Custom error types
    base.go          # Base HTTPError
    validation.go    # ValidationError
    not_found.go     # NotFoundError
  env/               # Environment variables
  handlers/*/        # HTTP request handlers
    templates/       # templ files (auto-generated Go code)
  logger/            # Logging infrastructure
  pdf/               # PDF generation logic
  shared/            # Shared types and entities
    type_*.go        # Custom type definitions (Slot, MachineType, etc.)
  urlb/              # URL building functions
  utils/             # Generic utilities
scripts/             # Helper scripts
```

## Handler Structure

Each handler package follows this pattern:

### Register
```go
func Register(e *echo.Echo) {
    e.GET("/path", handler)
}
```

### Common Handlers
- **auth**: login, logout, register (API key based)
- **home**: Main dashboard
- **tools**: Tool listing CRUD operations
- **tool/**: Individual tool pages (cycles, metal-sheets, notes regeneration)
- **press**: Press management and cycle tracking
- **metalsheets**: Metal sheet CRUD
- **notes**: Note management
- **troublereports**: Issue reporting with PDF export
- **editor**: Markdown editor
- **profile**: User profile settings

### Template Rendering
```go
func (h Handler) Render(c echo.Context, statusCode int, name string, data any) *errors.HTTPError {
    t, err := templates.Get(name)
    if err != nil {
        return errors.NewNotFoundError(err).Wrap("failed to get template")
    }

    c.Response().Header().Set(echo.HeaderContentType, "text/html; charset=utf-8")
    c.Response().WriteHeader(statusCode)

    if err := t.Render(c.Request().Context(), c.Response()); err != nil {
        return errors.NewRenderError(err, name)
    }
    return nil
}
```

## Database Operations

### Standard Query Pattern
```go
// Constant definition
const sqlListTools = `SELECT id, code, width, height FROM tools ORDER BY code;`

// Handler function
func ListTools() ([]*shared.Tool, *errors.HTTPError) {
    rows, err := dbTool.Query(sqlListTools)
    if err != nil {
        return nil, errors.NewHTTPError(err)
    }
    defer rows.Close()

    var tools []*shared.Tool
    for rows.Next() {
        tool, herr := ScanTool(rows)
        if herr != nil {
            return nil, herr
        }
        tools = append(tools, tool)
    }
    return tools, nil
}

// Scanner function
func ScanTool(row Scannable) (*shared.Tool, *errors.HTTPError) {
    var t shared.Tool
    err := row.Scan(&t.ID, &t.Code, &t.Width, &t.Height)
    if err != nil {
        return nil, errors.NewHTTPError(err)
    }
    return &t, nil
}
```

### Specific Database Operations

#### Tools Database (tools.db)
- **Cycle Management**: `AddCycle`, `GetCyclesByTool`, `ListCyclesByPress`
- **Metal Sheets**: Upper/lower metal sheet CRUD operations
- **Notes**: Tool and press notes management
- **Regenerations**: Track tool regeneration periods

#### Presses Database (presses.db)
- **Press Operations**: CRUD for manufacturing machines
- **Slot Management**: Upper/lower tool assignments
- **Cycle Tracking**: Press cycle history

#### Users Database (users.db)
- **User Management**: CRUD for API key authentication
- **Session Handling**: User session tracking

#### Reports Database (reports.db)
- **Trouble Reports**: Create, edit, delete issue reports
- **PDF Export**: Generate PDFs with attachments

## URL Building

Use `urlb` package for consistent route generation:
```go
import "github.com/knackwurstking/pg-press/internal/urlb"

// Generate URLs
url := urlb.ToolPage(toolID)
url := urlb.EditCycle(cycleID, pressID)
url := urlb.DownloadAttachment(reportID, filename)
```

## Error Handling Hierarchy

```go
ValidationError   → HTTP 400 (Bad Request)
NotFoundError     → HTTP 404 (Not Found)
AuthorizationError → HTTP 401 (Unauthorized)
ExistsError       → HTTP 409 (Conflict)
RenderError       → HTTP 500 (Internal Server Error)
```

All errors implement `*errors.HTTPError` with methods:
- `.Echo()` - Convert to Echo HTTP error
- `.Wrap(format, args...)` - Add context
- `.HTTPStatusCode()` - Get HTTP status code

## API Key Authentication

Configuration in `middleware.go`:
- Required length: 32 characters minimum (`MinAPIKeyLength`)
- Validated from header, query param, or cookie
- Skipper list for public assets (images, static files)
- User name stored in context: `c.Set("user-name", userName)`

## German Localization

Use `German()` method on entities for display:
```go
func (t Tool) German() string {
    return fmt.Sprintf("%s (%d x %d)", t.Code, t.Width, t.Height)
}

func (p Press) German() string {
    return fmt.Sprintf("%s (%s)", p.Code, p.Type.German())
}
```

### Date/Time Formatting
- Use `UnixMilli` type for timestamps (int64 milliseconds)
- Format constants: `DateFormat = "02.01.2006"`, `TimeFormat = "15:04"`
- German date format: day.month.year
- Conversion: `UnixMilli.FormatDate()`, `FormatDateTime(unixMilli)`

## PDF Generation

Location: `internal/pdf/`
- Generate PDFs for trouble reports
- Include linked attachments
- Support markdown content rendering

## Static Assets

Static files served from:
- `/` - Public assets (bundled by templ)
- `/images` - User uploaded images

Configure in `internal/env/`:
```go
var (
    ServerPathImages string  // Image directory path
)
```

## Special Considerations

### Database Migrations
No automatic migrations exist. Schema changes require manual SQL updates or new table creation.

### Template Auto-Generation
Templates are auto-generated with `make generate`. Edit `.templ` files, not the generated Go files.

### Common Patterns

#### Adding a New Entity
1. Define struct in `internal/shared/entity_*.go`
2. Implement `Entity[T]` interface methods
3. Add SQL constants in `internal/db/*.go`
4. Create handler CRUD operations
5. Register routes in handler

#### Adding a New Handler
1. Create handler package in `internal/handlers/`
2. Implement `Register(e *echo.Echo)` function
3. Add template files in `templates/`
4. Register handler in `internal/handlers/handlers.go`

#### Error Message Conventions
- Validation: "field must be [requirement]"
- Not found: "resource [name] with id [id] not found"
- Authorization: "you must be logged in to access this resource"

## Testing

- Test files named `*_test.go`
- Use standard Go testing package
- No tests currently exist; add tests when implementing new features
- Run with: `go test -v ./path/to/...`

### Test Example
```go
func TestAddTool(t *testing.T) {
    tool := &shared.Tool{
        Width: 100,
        Height: 50,
        // ... other fields
    }

    err := db.AddTool(tool)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
}
```

## Development Workflow

1. **Start development server**: `make dev`
2. **Make changes** to handlers, templates, or entities
3. **Regenerate templates**: `make generate`
4. **Run linter**: `make lint`
5. **Build binary**: `make build`

## CLI Commands

Available commands (run with `-h` for help):
- `server` - Start HTTP server
- `tools` - Tool management commands
- `user` - User database operations
- `cookies` - Cookie management
- `api-key` - API key utilities

## Common Tasks

### Add a New Tool Field
1. Update `shared.Tool` struct
2. Update SQL queries in `internal/db/tool_tools.go`
3. Update scanner function
4. Update validation in `Validate()`

### Add a New Route
1. Create handler function in appropriate package
2. Register route in `Register()` or `handlers.go`
3. Add template if rendering HTML
4. Update navigation if needed

### Debugging Issues
1. Enable verbose logging: `export VERBOSE=true`
2. Check server logs for stack traces
3. Verify database connections in `internal/db/db.go`
4. Check template errors in browser console

## Troubleshooting

### Common Issues
- **Templates not updating**: Run `make generate`
- **Database locked**: Ensure WAL mode is enabled
- **401 Unauthorized**: Check API key format and length (32+ chars)
- **Static files not found**: Verify `ServerPathImages` configuration

## Version Control

- Current version: v0.0.1
- Branching strategy: Feature branches merged to main
- Commit messages should describe changes clearly
