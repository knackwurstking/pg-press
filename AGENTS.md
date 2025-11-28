# PG Press Handover Documentation

## Overview

PG Press is a manufacturing management system built with Go that provides real-time tracking and management of manufacturing tools, press operations, maintenance reporting, and documentation. It uses HTMX for dynamic interactions without traditional REST APIs.

## Project Structure

The project follows a typical Go web application structure:

- `cmd/` - Application entry points and CLI commands
- `services/` - Business logic organized by domain
- `models/` - Data models and database entities
- `handlers/` - HTTP request handlers
- `utils/` - Utility functions and helpers
- `pdf/` - PDF generation functionality
- `assets/` - Static assets and configuration files
- `docs/` - Documentation including roadmap

## Main Components

- Services Layer
- Models Layer
- Handlers Layer
- Utils Layer
- PDF Generation
- Database Schema
- Configuration
- CLI Commands

## Key Features

### 🔧 Tool Management
- Complete tool lifecycle tracking (position, format, type, code)
- Press assignment and tool regeneration workflows
- Multi-press support (Press 0-5)
- Tool-specific notes and maintenance history
- Real-time status updates via WebSocket

### 📊 Press Cycle Tracking
- Automated cycle counting and reporting
- Historical cycle data with user attribution
- Press-specific performance metrics
- Integration with tool regeneration schedules

### 📋 Trouble Reports
- Comprehensive issue reporting system
- File attachment support with preview functionality
- PDF export for sharing and documentation
- Modification history tracking
- Searchable report database

### 📝 Notes System
- Multi-level priority system (INFO, ATTENTION, BROKEN)
- Flexible linking to tools, presses, or any entity
- Real-time collaborative editing
- Advanced filtering and search capabilities

### 🔄 Real-time Updates
- WebSocket-powered live updates
- Activity feed with comprehensive audit trail
- Automatic UI refresh for collaborative environments
- Push notifications for critical events

### 🚀 Performance Optimizations
- Advanced asset caching with versioning
- Efficient static file delivery
- ETag and Last-Modified support
- Optimized database queries with SQLite

## Services Layer

The services layer contains business logic organized by domain:

- Tools - Tool lifecycle management and status tracking
- Press Cycles - Cycle counting and reporting for presses
- Trouble Reports - Issue reporting and management system
- Notes - Notes management with priority levels
- Feeds - Activity feed and audit trail
- Users - User authentication and management
- Cookies - Cookie handling for sessions
- Modifications - Change tracking system
- Scanner - Tool and press scanning functionality
- Metal Sheets - Metal sheet management
- Tool Regenerations - Tool regeneration workflows
- Press Regenerations - Press regeneration workflows

## Models Layer

The models layer defines the data structures used throughout the application:

- Tool - Main tool entity with position, format, type and code
- Cycle - Press cycle tracking data
- TroubleReport - Issue reporting system with attachments
- Note - Notes with priority levels and entity linking
- Feed - Activity feed items for audit trail
- User - Authentication and authorization data
- Cookie - Session cookie handling
- Attachment - File attachment support
- Modification - Change tracking data
- PressRegeneration - Press regeneration tracking
- ToolRegeneration - Tool regeneration tracking

## PDF Generation

PDF generation utilities:
- Cycle Summary - Generates cycle summary reports
- Trouble Report - Creates PDF versions of trouble reports

## Database Schema

The system uses SQLite for database storage with a comprehensive schema that includes:

- Tool tracking and lifecycle management
- Press cycle counting and reporting
- Trouble report system with file attachments
- Notes system with priority levels
- Activity feed and audit trail
- User authentication and session management

## CLI Commands

The project includes a Makefile with useful commands:
- `make init` - Initialize project dependencies
- `make build` - Build the application
- `make dev` - Run development server with hot reloading
- `make test` - Run tests
- `make lint` - Run linters

Additional CLI commands available through the application:
- `pg-press user` - User management (add, remove, modify, list)
- `pg-press api-key` - API key generation
- `pg-press cookies` - Cookie cleanup and maintenance
- `pg-press feeds` - Activity feed management
- `pg-press cycles` - Press cycle data management
- `pg-press tools` - Tool management operations
- `pg-press server` - Start the web server

## Development Setup

1. Clone the repository
2. Run `make init` to initialize project dependencies
3. Run `make build` to build the application
4. Run `make dev` for development with hot reloading
5. Access the application at http://localhost:9020

## Technology Stack

- **Backend**: Go 1.25.0 with Echo web framework
- **Database**: SQLite with comprehensive schema
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **PDF Generation**: gofpdf for report exports
- **Authentication**: Cookie-based sessions with API key support
- **Real-time**: WebSocket for live updates
- **Architecture**: Server-rendered HTML with HTMX for dynamic updates (no REST API)

## Deployment Notes

The application is designed to run in a production environment with proper security headers and SSL support. The build process includes optimizations for static assets and proper error handling.

## Known Issues & TODOs

- Tools list does not reload when navigating backwards in history
- Notes management page needs improvements after editing or deleting a note
- UI: Add press regeneration section to the press page
- Fix the `getTotalCycles` tool handler method to check the last press regeneration
- Create a `ResolvedTroubleReport` type and replace `TroubleReportWithAttachments` with this
- Migrate `Attachment.ID` from string to int64, also need to migrate the database table

## Performance Considerations

- The application uses ETags and Last-Modified headers for efficient static file delivery
- Database queries are optimized with proper indexing
- HTMX is used to minimize the need for traditional REST APIs, improving performance

## Project Directory Structure Details

### cmd/pg-press/
This directory contains the main application entry points and CLI command implementations:
- `main.go` - Main application entry point
- `router.go` - HTTP routing configuration
- `middleware.go` - Application middleware
- Various command files that implement the CLI functionality for user, API key, cookies, feeds, cycles, tools, and server management

### handlers/
The handlers directory contains HTTP request handlers organized by functionality:
- `tools/` - Tool management endpoints
- `troublereports/` - Trouble report handling
- `press/` - Press operation endpoints
- `notes/` - Notes system endpoints
- `components/` - Reusable UI component templates

### services/
The services directory contains the core business logic organized by domain:
- `tools.go` - Core tool management functionality
- `press-cycles.go` - Press cycle tracking and calculations
- `trouble-reports.go` - Trouble report processing and management
- `notes.go` - Notes system implementation
- `feeds.go` - Activity feed and audit trail functionality

### assets/
Static assets directory containing:
- CSS, JavaScript, and image files
- PWA (Progressive Web App) assets
- Bootstrap icon files

### pdf/
PDF generation utilities for creating cycle summaries and trouble reports in PDF format

### models/
Data models that define the structure of data throughout the application, including:
- Tool, Cycle, TroubleReport, Note, Feed, User, Cookie, Attachment, Modification, PressRegeneration, ToolRegeneration

## HTMX Usage Pattern

PG Press leverages HTMX for dynamic UI interactions without traditional REST APIs:
- Partial page updates using HTMX attributes
- Server-rendered HTML with client-side interactivity
- WebSocket integration for real-time updates
- Drag-and-drop file attachment support
- Form handling with server-rendered validation

## Authentication & Session Management

The system uses cookie-based session management:
- User authentication with API key support
- Secure session handling with proper cookie configuration
- Session validation and expiration management

## Development Workflow

1. Make changes to Go source files, templates, or assets
2. Run `make dev` for hot reloading during development
3. Use `make build` to create production binaries
4. Test with `make test` before committing changes
5. Follow the existing code patterns and directory structure

## Testing Approach

The project uses Go's built-in testing framework. Tests should be written in the same directory as the code they're testing, following Go conventions.

## Security Considerations

- All user authentication is handled through secure cookies
- API keys are used for secure access to management commands
- Proper validation and sanitization of all user inputs
- Secure session handling with proper cookie attributes

## Additional Information

### Press Regenerations System
The application includes a press regeneration system that works similarly to tool regenerations but resets press total cycles back to zero. Key aspects:
- The current `Regeneration` model was renamed to `ToolRegeneration`
- A new `PressRegeneration` model was added
- A `PressRegenerations` service was implemented
- The system handles sorting by date instead of total cycles for press cycles
- UI elements for submitting press regenerations are implemented but not fully completed

### Tool Management Details
- Tools can be assigned to presses (0-5)
- Tools can be marked as dead or regenerating
- Tools have position, format, type and code attributes
- The system supports tool binding between top/bottom positions
- Real-time status updates are provided via WebSocket

### Trouble Report Features
- Comprehensive issue reporting with attachments
- File preview functionality
- PDF export capabilities
- Modification history tracking
- Searchable database of reports

### Notes System
- Multi-level priority system (INFO, ATTENTION, BROKEN)
- Flexible linking to tools, presses, or any entity
- Real-time collaborative editing capabilities
- Advanced filtering and search features

### Activity Feed
- Comprehensive audit trail of all system activities
- Real-time updates for collaborative environments
- Push notifications for critical events

### Cycle Tracking
- Automated cycle counting and reporting for presses
- Historical data with user attribution
- Integration with tool regeneration schedules

