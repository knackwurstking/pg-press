# PG Press - Project Handover Document

## Project Overview

PG Press is a web-based application built in Go that manages press-related data including tools, cycles, metal sheets, notes, and press regenerations. It provides both a command-line interface and a web interface for managing various aspects of press operations.

## Technologies Used

- **Go 1.25.3** - Main programming language
- **Echo Framework** - Web framework for building REST APIs and web applications
- **Templ** - Go templating engine for building HTML UI components
- **SQLite** - Local database storage
- **Custom UI Library** - Shared UI components using `knackwurstking/ui`
- **CLI Library** - Command-line interface using `SuperPaintman/nice`

## Project Structure

```
pg-press/
├── cmd/pg-press/           # Command-line interface
│   ├── main.go             # Main entry point
│   ├── commands-*.go       # CLI command implementations
│   └── middleware.go       # Custom middleware
├── handlers/               # Web handlers for different application sections
│   ├── auth/               # Authentication handlers
│   ├── home/               # Home page handler
│   ├── press/              # Press-related handlers
│   ├── tools/              # Tools management handlers
│   ├── notes/              # Notes handlers
│   ├── metalsheets/        # Metal sheets handlers
│   └── ...                 # Other handlers
├── components/             # UI components built with templ
├── services/               # Database services
│   ├── press/              # Press-related services
│   ├── shared/             # Shared model and service definitions
│   └── user/               # User-related services
├── pdf/                    # PDF generation functionality
├── env/                    # Environment configuration
├── utils/                  # Utility functions
├── assets/                 # Static assets (embedded)
├── templates/              # Page templates
└── Makefile                # Build and development commands
```

## Key Files and Directories

### Main Entry Points
- `cmd/pg-press/main.go` - Main CLI application with all commands
- `cmd/pg-press/commands-server.go` - Server command that launches the web app
- `embed.go` - Asset embedding for static files

### Core Services
- `services/shared/` - Shared models and interfaces for database entities
- `services/press/` - Press-related database services
- `services/user/` - User authentication and session management

### Web Handlers
- `handlers/home/` - Home page handler
- `handlers/auth/` - Authentication handlers
- `handlers/tool/` - Tool management handlers
- `handlers/press/` - Press-related handlers

### UI Components
- `components/` - Reusable UI components built with templ
- `handlers/*/templates/` - Page templates for different sections

## Running the Application

### Development
```bash
# Install dependencies and generate templates
make init

# Run in development mode with auto-reload (requires gow)
make dev

# Or run directly
make run
```

### Building
```bash
# Build the binary
make build

# Build and install for macOS service
make macos-install
```

## Database Schema

The application uses SQLite for data storage. Key tables are defined in the services modules:

- `press` - Main press information
- `tools` - Tool data
- `press_cycles` - Cycle tracking
- `tool_regenerations` - Tool regeneration history
- `press_regenerations` - Press regeneration history
- `users` - User accounts
- `user_sessions` - User session management
- `user_cookies` - Cookie data

## Adding New Features

### 1. Adding a New Handler
1. Create a new handler file in `handlers/`
2. Define the handler struct and route registration method
3. Add route registration to the main router (in `cmd/pg-press/router.go`)

### 2. Adding New UI Components
1. Create a new `.templ` file in the `components/` directory
2. Generate the Go code using `make generate`
3. Use the component in templates or other components

### 3. Adding New Services
1. Define the model in `services/shared/`
2. Implement service methods in the appropriate service directory
3. Add any necessary database migrations

### 4. Adding New CLI Commands
1. Add a new command function in `cmd/pg-press/commands-*.go`
2. Register it in the main CLI configuration in `main.go`

## Key Environment Variables

- `SERVER_ADDR` - Server address in format `<host>:<port>` (default: `:9020`)
- `SERVER_PATH_PREFIX` - Path prefix for server routes (default: `/pg-press`)
- `LOG_LEVEL` - Logging level (`debug`, `info`, `warn`, `error`, `fatal`)
- `LOG_FORMAT` - Logging format (`json`, `text`)
- `ADMINS` - Admin user identifiers

## Development Workflow

1. **Code Generation**: Run `make generate` after modifying `.templ` files
2. **Testing**: Run `make test` for unit tests
3. **Linting**: Run `make lint` for code linting
4. **Building**: Run `make build` to create a binary

## Deployment

### macOS Service Installation
```bash
make macos-install
make macos-start-service
```

### Configuration
The application stores configuration in:
- User config directory (`$HOME/Library/Application Support/pg-press/`)
- Environment variables for runtime settings

## Troubleshooting

### Common Issues
1. **Port in use**: Ensure no other process uses the configured port
2. **Missing dependencies**: Run `make init` to install all dependencies
3. **Template not found**: Run `make generate` to regenerate templates
4. **Database issues**: Check database file permissions and path

### Logging
- Logs are written to `~/Library/Application Support/pg-press/pg-press.log` on macOS
- Set `LOG_LEVEL=debug` for detailed logging during development

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes and add tests
4. Run `make test` and `make lint`
5. Submit a pull request

