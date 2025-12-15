# PG Press Handover Document

## Overview

PG Press is a Go-based web application for managing press shop operations, including tools, presses, cycles, and user management. It's designed to track tool usage, press cycles, regenerations, and troubleshooting reports.

## Project Structure

```
pg-press/
├── .models/                # Data models (database entities)
├── .services/              # Service layer implementations
├── .utils/                 # Utility functions
├── cmd/pg-press/           # Main application entry point
├── internal/
│   ├── .pdf/              # PDF generation utilities
│   ├── assets/            # Static asset management
│   ├── common/            # Common utilities and database setup
│   ├── components/        # HTML template components
│   ├── env/               # Environment configuration
│   ├── errors/            # Error handling
│   ├── handlers/          # HTTP handlers and routes
│   └── urlb/              # URL building utilities
├── scripts/                # Database migration and utility scripts
└── internal/shared/       # Shared types and interfaces
```

## Key Components

### Technology Stack
- **Language**: Go 1.25.3
- **Web Framework**: Echo v4 (HTTP server)
- **Templates**: A-H Templ (HTML templating)
- **Database**: SQLite (via mattn/go-sqlite3)
- **CLI Framework**: SuperPaintman/nice
- **Authentication**: Custom cookie-based system

### Architecture Layers

1. **Data Models** (`.models/`)
   - `tool.go` - Tool management (stamping tools)
   - `press.go` - Press machine tracking
   - `cycle.go` - Press cycles and production data
   - `user.go` - User authentication and permissions
   - `cookie.go` - Session management
   - `modification.go` - Tool modifications and repairs
   - `note.go` - Notes and attachments
   - `troublereport.go` - Trouble reporting system

2. **Service Layer** (`.services/`)
   - Database operations and business logic
   - Service registry for dependency injection
   - SQL query builders and executors

3. **Handlers** (`internal/handlers/`)
   - HTTP route handlers
   - Web interface controllers
   - Template rendering

4. **Common Utilities** (`internal/common/`)
   - Database setup and management
   - Service orchestration
   - Background task coordination

## Setup and Development

### Prerequisites
- Go 1.25.3 or higher
- Git
- Make
- SQLite3

### Initial Setup

```bash
# Clone the repository
git clone <repository-url>
cd pg-press

# Initialize dependencies and submodules
make init

# Install templ CLI tool (for template generation)
go install github.com/a-h/templ/cmd/templ@latest
```

### Development Workflow

```bash
# Generate template code
templ generate

# Install gow for live reload (recommended for development)
go install github.com/mitranim/gow@latest

# Run in development mode with live reload
make dev

# Run tests
go test ./...

# Lint code
make lint

# Build the application
make build
```

### Common Make Commands

| Command | Description |
|---------|-------------|
| `make init` | Initialize dependencies and submodules |
| `make dev` | Run in development mode with live reload |
| `make build` | Build the binary |
| `make lint` | Run linter |
| `make test` | Run tests |
| `make clean` | Clean build artifacts |

## Database Structure

The application uses SQLite and has the following main entities:

### Core Entities

1. **Tools** - Stamping tools with:
   - Tool number and name
   - Current cycle count
   - Tool type and specification
   - Status (active/inactive/retired)
   - Attached presses

2. **Presses** - Press machines with:
   - Press number and name
   - Location and capacity
   - Status and availability

3. **Cycles** - Production cycles with:
   - Press and tool associations
   - Start/end times
   - Quantity produced
   - Quality metrics

4. **Users** - System users with:
   - Telegram ID for authentication
   - Roles and permissions
   - Session management

5. **Notes** - Attachments and documentation
6. **Trouble Reports** - Issue tracking and resolution
7. **Regenerations** - Tool regeneration tracking

### Database Initialization

The database is automatically initialized when the application starts. Each service sets up its own tables and indexes.

### Database Indexes

There are migration scripts for adding/removing database indexes:

```bash
# Add indexes (improves query performance)
go run ./scripts/add-indexes-press -v /path/to/database.db

# Remove all indexes (for testing or migrations)
go run ./scripts/remove-indexes -v /path/to/database.db
```

## Running the Application

### Development Mode

```bash
# Start the server on port 8888
make dev

# Access the web interface at http://localhost:8888
```

### Production Mode

```bash
# Build the binary
make build

# Run the binary
./bin/pg-press server -a :9020

# Or install as a macOS service
make macos-install
make macos-start-service
```

### Command Line Interface

The application provides a comprehensive CLI with these main commands:

```bash
pg-press [command]

Commands:
  server          Start the HTTP server
  user            Handle users database table
  cookies         Handle cookies database table
  cycles          Press cycles management
  tools           Tools management
  api-key         API key management
  completion      Generate completion script

Server Flags:
  -a, --addr     Set server address (e.g., localhost:8080)
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_ADDR` | Server address to bind to | `:9020` |
| `SERVER_PATH_PREFIX` | URL path prefix | `/pg-press` |
| `ADMINS` | Comma-separated list of admin Telegram IDs | Empty |
| `VERBOSE` | Enable verbose logging | `false` |

## Deployment

### macOS Service

The application includes scripts for installing as a macOS service:

```bash
# Install the service
make macos-install

# Start the service
make macos-start-service

# Stop the service
make macos-stop-service

# Restart the service
make macos-restart-service

# Watch service logs
make macos-watch-service
```

### Deployment Checklist

1. Build the application (`make build`)
2. Install the binary (`make macos-install`)
3. Configure environment variables
4. Start the service (`make macos-start-service`)
5. Verify the service is running (`make macos-print-service`)
6. Monitor logs (`make macos-watch-service`)

## Key Workflows

### Tool Management

1. Add new tools with specifications
2. Track tool cycles and usage
3. Schedule and track regenerations
4. Manage tool modifications and repairs

### Press Cycle Tracking

1. Record press cycles with tool associations
2. Track production quantities and quality
3. Calculate tool utilization
4. Generate cycle summaries and reports

### User Management

1. Add users via Telegram ID
2. Manage user permissions
3. Track user sessions
4. Generate and manage API keys

### Trouble Reporting

1. Create trouble reports
2. Track resolution status
3. Generate reports and analytics

## Code Organization Patterns

### Service Pattern

Each domain has a dedicated service with CRUD operations:
- `NewService()` - Service factory
- `Setup()` - Initialize database tables
- `Close()` - Clean up resources
- `Create/Read/Update/Delete()` - Core operations

### Handler Pattern

Handlers follow MVC pattern:
- `NewHandler()` - Handler factory
- `RegisterRoutes()` - Route registration
- Handler methods for each route

### Error Handling

Custom error wrapper system:
- `MasterError` - Main error type
- Wrap errors with context
- Consistent error responses

### Database Access

- SQLite with connection pooling
- Prepared statements
- Transactions for complex operations
- Index management scripts

## Important Configuration Files

### `.env/constants.go`

Contains application constants:
- Cycle warnings/errors (800,000/1,000,000 cycles)
- Cookie expiration (6 months)
- Attachment size limits (10MB)
- Date/time formats

### `go.mod`

Go module configuration with dependencies:
- Echo web framework
- Templ HTML templates
- SQLite driver
- UUID generation
- CLI framework

## Troubleshooting

### Common Issues

1. **Port in use**: Change SERVER_ADDR or stop conflicting service
2. **Database errors**: Run `make init` to recreate database
3. **Template errors**: Run `templ generate` to regenerate templates
4. **Permission errors**: Ensure proper file permissions on database

### Log Files

- Console logs during development
- `~/Library/Application Support/pg-press/pg-press.log` for service mode

### Debugging Tips

1. Enable verbose mode: `export VERBOSE=true`
2. Check database integrity with SQLite tools
3. Review service status: `make macos-print-service`
4. Tail logs: `make macos-watch-service`

## Future Enhancements

Check `docs/TODO.md` for planned features:
- Enhanced feed system
- Editor functionality
- Metal sheet management
- Press regeneration tracking
- Improved navigation
- Help system

## Contact and Support

For questions or issues, refer to:
- Code comments and documentation
- Git history for context
- `docs/` directory for additional documentation
- TODO items for planned features

