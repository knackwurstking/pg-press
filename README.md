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

## Overview

PG Press is a manufacturing management system that provides real-time tracking and management of:

- **Manufacturing Tools**: Complete lifecycle management from creation to regeneration
- **Press Operations**: Multi-press environment support with cycle tracking
- **Maintenance Reports**: Trouble report generation with PDF export capabilities
- **Notes System**: Comprehensive documentation and issue tracking
- **Activity Feeds**: Real-time updates and audit trails
- **Metal Sheet Management**: Inventory and specification tracking
- **Tool Regeneration**: Automated tool regeneration scheduling and tracking

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
- Template pre-generation for faster rendering

## Technology Stack

- **Backend**: Go 1.25.3 with Echo web framework
- **Database**: SQLite with comprehensive schema
- **Frontend**: HTMX for dynamic interactions, vanilla JavaScript
- **Templates**: Templ for type-safe HTML generation
- **PDF Generation**: gofpdf for report exports
- **Authentication**: Cookie-based sessions with API key support
- **Real-time**: WebSocket for live updates
- **Architecture**: Server-rendered HTML with HTMX for dynamic updates (no REST API)
- **Build Tools**: Make, templ for template generation

## Quick Start

### Prerequisites

- Go 1.25.3 or higher
- Make (for build automation)
- Git
- Node.js (for development tools)

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

### Testing

To run tests:

```bash
make test
```

This will execute all unit and integration tests in the project.

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

2. **Install as macOS service (if on macOS)**

   ```bash
   make macos-install
   ```

3. **Start the service**

   ```bash
   make macos-start-service
   ```

4. **Alternative: Run directly**

   ```bash
   ./bin/pg-press server
   ```

   Or with custom environment variables:

   ```bash
   SERVER_ADDR=":9020" SERVER_PATH_PREFIX="/pg-press" ./bin/pg-press server
   ```

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
- `/tool/{id}` - Individual tool details
- `/press/{number}` - Press-specific view
- `/trouble-reports` - Trouble reports management
- `/notes` - Notes and documentation system
- `/profile` - User profile and session management
- `/feed` - Activity feed
- `/metal-sheets` - Metal sheets management
- `/umbau` - Umbau (conversion) management
- `/help` - Help and documentation
- `/editor` - Editor for content creation
- `/press-regeneration` - Press regeneration scheduling

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
├── components/            # Reusable UI components
├── docs/                  # Documentation files
├── handlers/              # HTTP handlers and templates
├── models/                # Data models
├── services/              # Business logic layer
├── utils/                 # Utility functions
├── scripts/               # Build and deployment scripts
└── bin/                   # Compiled binaries
```

### Make Commands

```bash
make init       # Initialize project (tidy modules, update submodules, generate templates)
make build      # Build production binary
make run        # Generate templates and run
make dev        # Run with hot reload using gow
make generate   # Generate templ templates
make clean      # Clean build artifacts
make count      # Count lines of code
make test       # Run tests (Golang)
make macos-install     # Install pg-press as a macOS service
make macos-start-service   # Start the pg-press service
make macos-stop-service    # Stop the pg-press service
make macos-restart-service # Restart the pg-press service
make macos-print-service   # Print information about the pg-press service
make macos-watch-service   # Watch the pg-press service logs
make macos-update     # Update the pg-press service
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
