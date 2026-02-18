# AGENTS.md

This document provides coding conventions and guidance for agentic coding agents working on the pg-press codebase.

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
3. Internal packages (alphabetically by module: `db`, `errors`, `env`, `logger`, `shared`, `urlb`, `utils`)

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

### Database
- Multiple SQLite databases: tools, presses, notes, users, reports
- Open databases in `internal/db/db.go`
- Enable WAL mode: `journal=WAL&synchronous=1`
- Configure connection pooling for concurrent access

### Error Messages
- Validation errors: "field must be [requirement]"
- Not found: "resource [name] with id [id] not found"
- Authorization: "you must be logged in to access this resource"

### File Organization
```
cmd/pg-press/     # CLI entry points (main.go, commands-*.go)
internal/
  components/     # Reusable UI components
  db/             # Database access layer
  errors/         # Custom error types
  env/            # Environment variables
  handlers/*/     # HTTP request handlers
    templates/    # templ files
  logger/         # Logging infrastructure
  pdf/            # PDF generation logic
  shared/         # Shared types, entities, utilities
    type_*.go     # Custom type definitions
  urlb/           # URL building functions
  utils/          # Generic utilities
scripts/          # Helper scripts
```

### SQL Conventions
- Use named databases: `tool`, `press`, `note`, `user`, `reports`
- Table names pluralized: `tools`, `cycles`, `users`
- Use WAL mode for SQLite connections
- Set connection limits: `SetMaxOpenConns(10)`, `SetMaxIdleConns(5)`
- Close connections properly

### Date/Time
- Use `UnixMilli` type for timestamps (int64 milliseconds)
- Format constants: `DateFormat = "02.01.2006"`, `TimeFormat = "15:04"`
- German date format: day.month.year
- Convert with `UnixMilli.FormatDate()`, `FormatDateTime()`

### Testing
- No tests currently exist; add tests when implementing new features
- Test files named `*_test.go`
- Use standard Go testing package

## Special Considerations

### Error Types Hierarchy
```go
ValidationError → HTTP 400
NotFoundError   → HTTP 404
AuthorizationError → HTTP 401
ExistsError     → HTTP 409
```

### Template Rendering
- Render templates in handlers: `t.Render(c.Request().Context(), c.Response())`
- Wrap render errors: `errors.NewRenderError(err, "Template Name")`

### API Key Authentication
- Required length: 32 characters minimum (`MinAPIKeyLength`)
- Validated from header, query param, or cookie
- Skipper list in `middleware.go` for public assets

### German Localization
- Use `German()` method on entities for display names
- Translate PDF output with translator function
