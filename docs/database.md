# Database Overview

This document describes the database structure and services of the PG Press application.

## Database Components

### Core Services
- **common/** - Main database entry points
  - `db.go` - Main entry point for database services

- **shared/** - Shared components across services
  - `interfaces.go` - Defines `Service` and `Entity` interfaces
  - `types.go` - Common type aliases used across services
  - `model-*.go` - Common data models used across services
  - `base-service.go` - Base service struct implementing common functionality
  - `base-service-config.go` - Configuration struct for base service configuration

### User Management
- **user/**
  - `cookie.go` - Cookie storage with indexes: `cookies(user_id)`, `cookies(value)`
  - `session.go` - In-memory session service (no database storage)
  - `user.go` - User storage with index: `users(api_key)`. Uses telegram ID as primary identifier.

### Press Management
- **press/**
  - `press.go` - Press definitions and data
  - `cycles.go` - Cycle tracking for presses
  - `regeneration.go` - Press regeneration history

### Tool Management
- **tool/**
  - `tool.go` - Tool definitions (ID, position, code, type, etc.). Tools are never deleted, only marked as dead. Cycles are tracked within the press service.
  - `regeneration.go` - Tool regeneration tracking. Each regeneration links to a tool and resets tool cycles to zero.

