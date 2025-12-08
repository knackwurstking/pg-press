# PG Press - Project Handover Documentation

## Project Overview
PG Press is a Go-based application for managing press operations, tools, and related data. It includes a web interface built with Go templates and a command-line tool for managing various aspects of the system.

## Project Structure
```
.
├── cmd/                  # Main application entry points
│   └── pg-press/         # CLI application
├── services/             # Business logic and service layer
├── handlers/             # HTTP handlers
├── models/               # Data models
├── components/           # Go template components
├── assets/               # Static assets (images, CSS, JS, etc.)
├── docs/                 # Documentation files
├── pdf/                  # PDF generation functionality
├── utils/                # Utility functions
├── env/                  # Environment configuration
├── errors/               # Custom error types
└── scripts/              # Helper scripts
```

## Build/Run
- Build: `make build`
- Lint: `make lint (golangci-lint)`
- Test: `go test ./...`
- Single test: `go test -run ^TestX$`
- Development: `make dev` (uses gow for hot reloading)
- Run server: `make run`

## Code Style
- Imports: stdlib, external, local groups
- Naming: PascalCase (types), camelCase (variables)
- Errors: Explicit checks with if err != nil
- Formatting: gofmt -w
- Linting: golangci-lint run

## Guidelines
- No unused imports or unhandled errors
- Comment public APIs
- Use validation package for inputs
- Keep functions under 50 lines
- Always log errors with stack trace
- Commit messages: Always use semantic git commit message style

## Key Features
- Web-based UI for managing press operations
- Command-line interface for various administrative tasks
- Database integration with SQLite
- PDF generation capabilities
- Template-based UI using templ
- RESTful API endpoints
- User management
- Tools and feeds management
- Press cycles tracking and reporting
- Cookie management

## Dependencies
- github.com/SuperPaintman/nice v0.0.0-20211001214957-a29cd3367b17
- github.com/a-h/templ v0.3.960
- github.com/google/uuid v1.6.0
- github.com/jung-kurt/gofpdf/v2 v2.17.3
- github.com/knackwurstking/ui v1.1.2-0.20251206161601-92d6501fae80
- github.com/labstack/echo/v4 v4.13.4
- github.com/labstack/gommon v0.4.2
- github.com/lmittmann/tint v1.1.2
- github.com/mattn/go-sqlite3 v1.14.32
- github.com/williepotgieter/keymaker v1.0.0
- golang.org/x/net v0.42.0

## Development Commands
- `make build` - Build the application
- `make dev` - Run development server with hot reloading
- `make run` - Run the server
- `make test` - Run all tests
- `make lint` - Run linter
- `make generate` - Generate Go templates
- `make clean` - Clean project directory

## macOS Service Management
- `make macos-install` - Install as macOS service
- `make macos-start-service` - Start service
- `make macos-stop-service` - Stop service
- `make macos-restart-service` - Restart service
- `make macos-print-service` - Print service information
- `make macos-watch-service` - Watch service logs
- `make macos-update` - Update service

## Key Components
- **Services**: Business logic implementations (press cycles, tools, users, etc.)
- **Handlers**: HTTP request handlers
- **Models**: Data structures and database models
- **Components**: Reusable UI components built with templ
- **PDF**: PDF generation functionality
- **Utils**: Generic utility functions
- **Env**: Environment configuration management

## Database
The application uses SQLite for data storage. Database operations are handled through the service layer.

## API Endpoints
The application exposes RESTful API endpoints for managing various resources including:
- Press cycles
- Tools
- Users
- Feeds
- Cookies
- Reports

## CLI Commands
The CLI supports various commands for administrative tasks:
- apiKey
- user (list, show, add, remove, modify)
- cookies (remove, auto-clean)
- feeds (list, remove)
- cycles
- tools
- server

## Environment Variables
The application uses environment variables for configuration:
- SERVER_ADDR
- SERVER_PATH_PREFIX
- LOG_LEVEL
- LOG_FORMAT
- ADMINS

## Testing
Tests are written using Go's built-in testing framework. Run all tests with `go test ./...` or individual tests with `go test -run ^TestX$`.

## Documentation
- Database schema documentation: docs/database.md
- PG Press mon documentation: docs/pg-press-mon.md
- PG Press roadmap: docs/pg-press-roadmap.md
