# PG Press

A comprehensive web application for press and tool management in manufacturing environments. Built with Go and designed for efficient tracking of manufacturing processes, tool lifecycle management, and maintenance reporting.

## Overview

PG Press is a manufacturing management system that provides real-time tracking and management of:

- **Manufacturing Tools**: Complete lifecycle management from creation to regeneration
- **Press Operations**: Multi-press environment support with cycle tracking
- **Maintenance Reports**: Trouble report generation with PDF export capabilities
- **Notes System**: Comprehensive documentation and issue tracking
- **Activity Feeds**: Real-time updates and audit trails
- **Metal Sheet Management**: Inventory and specification tracking

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

## Technology Stack

- **Backend**: Go 1.25+ with Echo web framework
- **Database**: SQLite with comprehensive schema
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **PDF Generation**: gofpdf for report exports
- **Authentication**: Cookie-based sessions with API key support
- **Real-time**: WebSocket for live updates
- **Architecture**: Server-rendered HTML with HTMX for dynamic updates (no REST API)

## Quick Start

### Prerequisites

- Go 1.25 or higher
- Make (for build automation)
- Git

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/knackwurstking/pg-press.git
   cd pg-press
   ```

2. **Build the application**

   ```bash
   make build
   ```

3. **Create your first user**

   ```bash
   ./bin/pg-press user add 123456789 "Your Name" "your-api-key-here"
   ```

   Note: Replace `123456789` with your Telegram ID, `"Your Name"` with your actual name, and `"your-api-key-here"` with a secure API key.

4. **Start the server**

   ```bash
   make run
   ```

5. **Access the application**
   Open your browser to `http://localhost:8080`

### Development Setup

For development with hot reloading:

```bash
make dev
```

This will start the server with automatic rebuilding on file changes.

## Configuration

### Environment Variables

| Variable             | Description               | Default        | Example          |
| -------------------- | ------------------------- | -------------- | ---------------- |
| `SERVER_ADDR`        | Server bind address       | `:8080`        | `localhost:3000` |
| `SERVER_PATH_PREFIX` | URL path prefix           | `/`            | `/pg-press`      |
| `DATABASE_PATH`      | SQLite database file      | `pg-press.db`  | `/data/app.db`   |
| `ASSET_VERSION`      | Asset versioning override | auto-generated | `v1.2.3`         |

### Production Deployment

1. **Build production binary**

   ```bash
   make build-prod
   ```

2. **Set production environment**

   ```bash
   export SERVER_ADDR=":8080"
   export SERVER_PATH_PREFIX="/pg-press"
   ```

3. **Run with systemd or supervisor**
   ```bash
   ./bin/pg-press server
   ```

## API Documentation

### Web Interface

PG Press uses an HTMX-based web interface that provides dynamic interactions without traditional REST APIs:

#### Main Pages

- `/` - Dashboard with system overview
- `/tools` - Tools management and press overview
- `/tools/tool/{id}` - Individual tool details
- `/tools/press/{number}` - Press-specific view
- `/trouble-reports` - Trouble reports management
- `/notes` - Notes and documentation system
- `/profile` - User profile and session management

#### Dynamic Features

- **Real-time Updates**: WebSocket-powered live data updates
- **Partial Page Updates**: HTMX endpoints for seamless interactions
- **Form Handling**: Server-rendered forms with validation
- **File Uploads**: Drag-and-drop attachment support

For complete endpoint documentation, see [docs/API.md](docs/API.md).

## Database Schema

The application uses SQLite with a comprehensive schema supporting:

- **Users & Authentication**: Secure user management with API keys
- **Tools & Equipment**: Complete tool lifecycle tracking
- **Press Operations**: Multi-press cycle management
- **Documentation**: Notes, reports, and attachments
- **Audit Trail**: Complete modification history

See [docs/DATABASE.md](docs/DATABASE.md) for detailed schema information.

## Documentation

- [🌟 Features Overview](docs/FEATURES.md) - Comprehensive feature documentation
- [🗄️ Database Schema](docs/DATABASE.md) - Complete database structure
- [🚀 Caching Strategy](docs/CACHING.md) - Asset optimization details
- [🛣️ HTMX Endpoints](docs/API.md) - HTMX architecture and endpoints
- [🛤️ Routing Table](docs/ROUTING.md) - All available routes
- [📝 Notes System](docs/NOTES_SYSTEM.md) - Documentation management
- [📝 Editor System](docs/EDITOR_SYSTEM.md) - Markdown editor implementation
- [❓ Help System](docs/HELP_SYSTEM.md) - Interactive markdown help and user documentation
- [📝 Shared Markdown](docs/SHARED_MARKDOWN_SYSTEM.md) - Shared markdown rendering system
- [📝 Markdown Implementation](docs/MARKDOWN_IMPLEMENTATION.md) - Detailed markdown feature documentation
- [🚀 Markdown Enhancements](docs/MARKDOWN_ENHANCEMENTS.md) - Recent markdown improvements and features
- [🧪 Path Prefix Testing](docs/PATH_PREFIX_TESTING.md) - Server path prefix testing and deployment guide
- [🔄 Migration Guide](docs/MIGRATION_GUIDE.md) - Database migration procedures
- [📋 Migration Completion](docs/MIGRATION_COMPLETION.md) - Migration summary and status
- [🧹 Scripts Cleanup](docs/SCRIPTS_CLEANUP.md) - Scripts directory reorganization summary
- [📚 Documentation Index](docs/README.md) - Complete documentation overview and navigation
- [📋 Documentation Organization](docs/DOCUMENTATION_ORGANIZATION_SUMMARY.md) - Validation and organization summary

## Development

### Project Structure

```
pg-press/
├── cmd/pg-press/          # Application entry points
├── internal/              # Internal application code
│   ├── database/          # Database connection management
│   ├── services/          # Business logic layer
│   ├── web/              # Web handlers and templates
│   └── constants/        # Application constants
├── pkg/                  # Reusable packages
│   ├── models/           # Data models
│   ├── utils/            # Utility functions
│   └── logger/           # Logging system
├── docs/                 # Documentation
├── scripts/              # Build and deployment scripts
└── bin/                  # Compiled binaries
```

### Make Commands

```bash
make build      # Build production binary
make build-dev  # Build development binary
make dev        # Run with hot reload
make run        # Run production build
make test       # Run all tests
make clean      # Clean build artifacts
make docs       # Generate documentation
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is proprietary software. All rights reserved.

## Support

For support and questions:

- Create an issue on GitHub
- Contact the development team
- Check the documentation in `/docs`

## Changelog

### v0.0.1 (Current)

- Initial release
- Core tool and press management
- Trouble reporting system
- Notes and documentation
- Real-time updates via WebSocket
- PDF export functionality
- Advanced caching system
