# PG Press

A comprehensive web application for press and tool management in manufacturing environments. Built with Go and designed for efficient tracking of manufacturing processes, tool lifecycle management, and maintenance reporting.

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
  - [🔧 Tool Management](#-tool-management)
  - [📊 Press Cycle Tracking](#-press-cycle-tracking)
  - [📋 Trouble Reports](#-trouble-reports)
  - [📝 Notes System](#-notes-system)
  - [🔄 Real-time Updates](#-real-time-updates)
  - [🚀 Performance Optimizations](#-performance-optimizations)
- [Technology Stack](#technology-stack)
- [Quick Start](#quick-start)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Development Setup](#development-setup)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Production Deployment](#production-deployment)
- [API Documentation](#api-documentation)
  - [Web Interface](#web-interface)
  - [Main Pages](#main-pages)
  - [Dynamic Features](#dynamic-features)
- [Database Schema](#database-schema)
- [Development](#development)
  - [Project Structure](#project-structure)
  - [Make Commands](#make-commands)
  - [CLI Commands](#cli-commands)
  - [Contributing](#contributing)
- [License](#license)
- [Support](#support)
- [Changelog](#changelog)

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

- **Backend**: Go 1.25.0 with Echo web framework
- **Database**: SQLite with comprehensive schema
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **PDF Generation**: gofpdf for report exports
- **Authentication**: Cookie-based sessions with API key support
- **Real-time**: WebSocket for live updates
- **Architecture**: Server-rendered HTML with HTMX for dynamic updates (no REST API)

## Quick Start

### Prerequisites

- Go 1.25.0 or higher
- Make (for build automation)
- Git

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/knackwurstking/pg-press.git
   cd pg-press
   ```

2. **Initialize and build the application**

   ```bash
   make init
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
   Open your browser to `http://localhost:9020`

### Development Setup

For development with hot reloading:

```bash
make dev
```

This will start the server with automatic rebuilding on file changes using `gow`.

## Configuration

### Environment Variables

| Variable             | Description          | Default       | Example          |
| -------------------- | -------------------- | ------------- | ---------------- |
| `SERVER_ADDR`        | Server bind address  | `:9020`       | `localhost:3000` |
| `SERVER_PATH_PREFIX` | URL path prefix      | `/pg-press`   | `/`              |
| `DATABASE_PATH`      | SQLite database file | `pg-press.db` | `/data/app.db`   |
| `LOG_LEVEL`          | Logging level        | `INFO`        | `DEBUG`          |

### Production Deployment

1. **Build production binary**

   ```bash
   make build
   ```

2. **Set production environment**

   ```bash
   export SERVER_ADDR=":9020"
   export SERVER_PATH_PREFIX="/pg-press"
   ```

3. **Run with systemd**

   A systemd service file is provided at `cmd/pg-press/pg-press.service`. Configure and install it:

   ```bash
   sudo cp cmd/pg-press/pg-press.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable pg-press
   sudo systemctl start pg-press
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

## Database Schema

The application uses SQLite with a comprehensive schema supporting:

- **Users & Authentication**: Secure user management with API keys
- **Tools & Equipment**: Complete tool lifecycle tracking
- **Press Operations**: Multi-press cycle management
- **Documentation**: Notes, reports, and attachments
- **Audit Trail**: Complete modification history

The database is automatically initialized on first run with proper SQLite optimizations including WAL mode and foreign key constraints.

## Development

### Project Structure

```
pg-press/
├── cmd/pg-press/          # Application entry points and CLI commands
├── internal/              # Internal application code
│   ├── constants/         # Application constants
│   ├── env/              # Environment configuration
│   ├── interfaces/       # Interface definitions
│   ├── pdf/              # PDF generation utilities
│   ├── services/         # Business logic layer
│   └── web/              # Web handlers and templates
├── pkg/                  # Reusable packages
│   ├── constants/        # Shared constants
│   ├── logger/           # Logging system
│   ├── models/           # Data models
│   ├── modification/     # Modification tracking
│   └── utils/            # Utility functions
├── scripts/              # Build and deployment scripts
└── bin/                  # Compiled binaries
```

### Make Commands

```bash
make init       # Initialize project (tidy modules, update submodules)
make build      # Build production binary
make run        # Generate templates and run
make dev        # Run with hot reload using gow
make generate   # Generate templ templates
make clean      # Clean build artifacts
make count      # Count lines of code
```

### CLI Commands

The application includes a comprehensive CLI for management tasks:

- `pg-press user` - User management (add, remove, modify, list)
- `pg-press api-key` - API key generation
- `pg-press cookies` - Cookie cleanup and maintenance
- `pg-press feeds` - Activity feed management
- `pg-press cycles` - Press cycle data management
- `pg-press tools` - Tool management operations
- `pg-press server` - Start the web server

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

## Changelog

### v0.0.1 (Current)

- Initial release
- Core tool and press management
- Trouble reporting system
- Notes and documentation
- Real-time updates via WebSocket
- PDF export functionality
- Advanced caching system
- CLI management tools
- Systemd service integration
