# AGENTS.md

## Build/Lint/Test Commands

```bash
make build          # Build production binary
make run            # Generate templates and run server
make dev            # Run with hot reload using gow
make generate       # Generate templ templates
make init           # Initialize project (tidy modules, update submodules)
go test ./...       # Run all tests
go test -v ./...    # Run all tests with verbose output
go test -run TestName ./...  # Run a specific test
```

## Code Style Guidelines

### Imports

- Group imports in order: standard library, external libraries, internal packages
- Use descriptive aliases for long package names (e.g., `echo "github.com/labstack/echo/v4"`)

### Formatting

- Use Go's standard formatting (gofmt) for all Go files
- Maintain consistent indentation with tabs (not spaces)

### Types

- Use specific types instead of `interface{}` where possible
- Prefer explicit type declarations over type inference for readability
- Use pointer types for large structs to avoid copying

### Naming Conventions

- Use camelCase for Go identifiers (e.g., `toolName`, `serverAddr`)
- Use PascalCase for exported names (e.g., `Tool`, `Server`)
- Use lowercase for package names (e.g., `services`, `utils`)

### Error Handling

- Handle errors explicitly and provide meaningful error messages
- Use custom error types for better error categorization (e.g., `errors.AlreadyExists`)
- Log errors with context using structured logging
- Do not use "failed" in error messages

## Project Structure

The pg-press project follows a structured organization:

```
pg-press/
├── cmd/pg-press/          # Application entry points and CLI commands
├── handlers/              # Web handlers and UI components
│   ├── auth/              # Authentication handlers
│   ├── components/        # Shared UI components
│   ├── dialogs/           # Dialog handlers
│   ├── editor/            # Editor handlers
│   ├── feed/              # Activity feed handlers
│   ├── help/              # Help page handlers
│   ├── home/              # Home page handlers
│   ├── metalsheets/       # Metal sheet handlers
│   ├── nav/               # Navigation handlers
│   ├── notes/             # Notes handlers
│   ├── press/             # Press-related handlers
│   ├── profile/           # Profile handlers
│   ├── tool/              # Individual tool handlers
│   ├── tools/             # Tools list handlers
│   ├── troublereports/    # Trouble reports handlers
│   ├── umbau/             # Umbau-related handlers
│   └── wsfeed/            # WebSocket feed handlers
├── models/                # Data models
├── services/              # Business logic layer
├── env/                   # Environment configuration
├── errors/                # Custom error types
├── utils/                 # Utility functions and helpers
├── assets/                # Static assets (CSS, JS, images)
├── docs/                  # Documentation files
├── scripts/               # Build and deployment scripts
└── bin/                   # Compiled binaries
```

## CLI Commands

The application includes a comprehensive CLI for management tasks:

### User Management

- `pg-press user add` - Add a new user
- `pg-press user remove` - Remove a user
- `pg-press user list` - List all users
- `pg-press user show` - Show user details
- `pg-press user modify` - Modify user data

### API Key Management

- `pg-press api-key` - Generate API keys

### Cookie Management

- `pg-press cookies remove` - Remove cookie data
- `pg-press cookies auto-clean` - Auto-clean expired cookies

### Feed Management

- `pg-press feeds list` - List activity feeds
- `pg-press feeds remove` - Remove feed data

### Cycle Management

- `pg-press cycles` - Manage press cycle data

### Tool Management

- `pg-press tools list` - List all tools with optional ID filtering
- `pg-press tools list-dead` - List all dead tools
- `pg-press tools mark-dead` - Mark a tool as dead by ID
- `pg-press tools revive` - Revive a dead tool by ID
- `pg-press tools delete` - Delete a tool by ID
- `pg-press tools list-cycles` - List press cycles for a tool by ID
- `pg-press tools list-regenerations` - List regenerations for a tool by ID

### Server Management

- `pg-press server` - Start the HTTP server for the web application

## Server Features

The pg-press server provides:

- Real-time updates via WebSocket
- HTMX-powered dynamic web interface
- Server-rendered HTML with dynamic interactions
- Cookie-based authentication
- API key support for secure access
- Activity feed with audit trail
- PDF export capabilities for trouble reports
- Multi-press environment support

## Database Schema

The application uses SQLite with a comprehensive schema supporting:

- **Users & Authentication**: Secure user management with API keys
- **Tools & Equipment**: Complete tool lifecycle tracking
- **Press Operations**: Multi-press cycle management
- **Documentation**: Notes, reports, and attachments
- **Audit Trail**: Complete modification history

The database is automatically initialized on first run with proper SQLite optimizations including WAL mode and foreign key constraints.

## Technology Stack

- **Backend**: Go 1.25.3 with Echo web framework
- **Database**: SQLite with comprehensive schema
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **PDF Generation**: gofpdf for report exports
- **Authentication**: Cookie-based sessions with API key support
- **Real-time**: WebSocket for live updates
- **Architecture**: Server-rendered HTML with HTMX for dynamic updates (no REST API)

## Environment Variables

| Variable             | Description          | Default       | Example          |
| -------------------- | -------------------- | ------------- | ---------------- |
| `SERVER_ADDR`        | Server bind address  | `:9020`       | `localhost:3000` |
| `SERVER_PATH_PREFIX` | URL path prefix      | `/pg-press`   | `/`              |
| `DATABASE_PATH`      | SQLite database file | `pg-press.db` | `/data/app.db`   |
| `LOG_LEVEL`          | Logging level        | `INFO`        | `DEBUG`          |

## Development Workflow

1. **Initialize project**: `make init`
2. **Generate templates**: `make generate`
3. **Run in development mode**: `make dev` (with hot reload)
4. **Build for production**: `make build`
5. **Run with server**: `make run`

## Cursor/Copilot Rules

No specific Cursor or Copilot rules configured.