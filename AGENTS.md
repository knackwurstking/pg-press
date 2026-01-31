# PG Press Project Documentation

## Overview

PG Press is a modern web application built with Go, featuring a comprehensive UI library and various backend services. This document serves as technical documentation for the project.

## Technology Stack

### Core Technologies

- **Language**: Go (1.25.3)
- **Web Framework**: Echo v4.13.4
- **Templating Engine**: A-H Templ v0.3.960
- **Database**: SQLite (via mattn/go-sqlite3 v1.14.32)
- **Authentication**: Keymaker v1.0.0
- **UI Library**: Custom UI library (Recursive font-based design)
- **Dynamic Content**: HTMX

### Key Dependencies

```go
github.com/a-h/templ v0.3.960
github.com/labstack/echo/v4 v4.13.4
github.com/mattn/go-sqlite3 v1.14.32
github.com/williepotgieter/keymaker v1.0.0
```

## Project Structure

### Root Directory

```
/pg-press/
├── bin/              # Compiled binaries
├── cmd/              # Main application entry points
├── data/             # Data files and assets
├── docs/             # Documentation
├── internal/         # Main application code
│   ├── assets/       # UI assets (CSS, JS, fonts)
│   │   └── css/ui.min.css  # Main UI stylesheet
│   ├── components/   # Shared UI components
│   ├── db/           # Database operations
│   ├── env/          # Environment configuration
│   ├── errors/       # Error handling
│   ├── handlers/     # HTTP request handlers
│   │   ├── auth/     # Authentication routes
│   │   ├── dashboard/# Dashboard functionality
│   │   └── ...       # Other route handlers
│   ├── logger/       # Logging infrastructure
│   ├── shared/       # Shared utilities
│   ├── urlb/         # URL building
│   └── utils/        # Utility functions
├── scripts/          # Development scripts
└── ...
```

### UI Architecture

- **CSS Framework**: Custom minified CSS (`ui.min.css`) with:
  - Recursive font family (Google Fonts)
  - CSS variables for theming
  - Responsive design patterns
  - Component-based styling

## Development Workflow

### Setup

1. Install Go 1.25.3+
2. Run `go mod download` to fetch dependencies
3. Generate templ files: `templ generate` or `make generate`

### Key Commands

```bash
# Build the application
make build

# Run development server
make dev

# Generate templ files
make generate

# Run tests
go test ./...
```

### Environment Variables

```env
# Database configuration
DB_PATH=data/pg_press.db

# Server settings
SERVER_ADDR=:8080
SERVER_PATH_PREFIX=/
VERBOSE=true

# Authentication
ADMINS=admin@example.com

SERVER_PATH_IMAGES=${HOME}/.pg-press/images
```

## Database Schema

The application uses SQLite with the following key tables:

- `users` - User accounts and profiles (see [internal/db/user_users.go](file:///Users/knackwurstking/Git/pg-press/internal/db/user_users.go))
- `cookies` - Authentication cookies (see [internal/db/user_cookies.go](file:///Users/knackwurstking/Git/pg-press/internal/db/user_cookies.go))
- `presses` - Press machines (see [internal/db/press_presses.go](file:///Users/knackwurstking/Git/pg-press/internal/db/press_presses.go))
- `tools` - Tools and equipment (see [internal/db/tool_tools.go](file:///Users/knackwurstking/Git/pg-press/internal/db/tool_tools.go))

## API Endpoints

The application uses Echo framework with handlers organized in [internal/handlers/](file:///Users/knackwurstking/Git/pg-press/internal/handlers). The endpoints are registered in [handlers.go](file:///Users/knackwurstking/Git/pg-press/internal/handlers/handlers.go) and include:

```
GET    /               # Home page (see [home/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/home))
GET    /auth/*          # Authentication routes (see [auth/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/auth))
GET    /profile/*       # Profile management (see [profile/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/profile))
GET    /tools/*         # Tools management (see [tools/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/tools))
GET    /tool/*          # Individual tool operations (see [tool/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/tool))
GET    /notes/*         # Notes management (see [notes/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/notes))
GET    /press/*         # Press operations (see [press/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/press))
GET    /umbau/*         # Umbau operations (see [umbau/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/umbau))
GET    /metal-sheets/*  # Metal sheets management (see [metalsheets/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/metalsheets))
GET    /trouble-reports/* # Trouble reports (see [troublereports/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/troublereports))
GET    /dialog/*        # Dialog operations (see [dialogs/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/dialogs))
GET    /editor/*        # Editor operations (see [editor/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/editor))
GET    /help/markdown   # Help documentation (see [help/](file:///Users/knackwurstking/Git/pg-press/internal/handlers/help))
```

## UI Components

The project uses a component-based architecture with:

- Shared components in `internal/components/`
- Page-specific templates in `internal/handlers/{handler}/templates/`
- HTMX for dynamic content loading

## Styling System

The UI library provides:

- Color themes (light/dark mode)
- Responsive grid system
- Component styles (cards, buttons, forms)
- Utility classes for spacing, alignment, etc.

## Best Practices

1. **Templating**: Use `.templ` files for HTML templates with Go logic
2. **Error Handling**: Centralized error handling in `internal/errors/`
3. **Database**: Use prepared statements and transactions
4. **Security**: Input validation, CSRF protection, JWT authentication
5. **Logging**: Structured logging with context

## Deployment

1. Build the application: `make build`
2. Configure environment variables
3. Run the binary or use Docker

## Troubleshooting

- Check `go.mod` for dependency issues
- Verify database migrations
- Review logs in `internal/logger/`

## Project Agents

### Agent Types

PG Press implements various agent patterns to handle different aspects of the application:

#### 1. User Management Agent

- Handles user authentication and session management
- Manages user profiles and permissions
- Integrates with Keymaker for secure authentication

#### 2. Database Agent

- Manages all database operations using SQLite
- Handles transactions and prepared statements
- Provides centralized access to database tables

#### 3. UI Agent

- Renders HTML templates using A-H Templ engine
- Manages component rendering and state
- Handles dynamic content via HTMX

#### 4. Logging Agent

- Provides structured logging capabilities
- Manages log levels and output formats
- Supports contextual logging

#### 5. Configuration Agent

- Handles environment variable parsing and validation
- Manages application configuration loading
- Provides centralized config access

### Agent Architecture

Agents are designed with the following principles:

1. **Single Responsibility**: Each agent handles a specific domain
2. **Loose Coupling**: Agents communicate through well-defined interfaces
3. **Reusability**: Agents can be reused across different parts of the application
4. **Testability**: Each agent is designed to be easily unit tested

### Agent Integration

Agents integrate through:

- Shared interfaces for consistent communication
- Context propagation for request-scoped data
- Error handling middleware for centralized error management

This document provides an overview of the PG Press project architecture and development workflow.
