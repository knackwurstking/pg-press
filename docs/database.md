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

## Shared Models Overview

This section provides a quick overview of all models defined in the shared library:

### Tool
Represents a tool used in a press machine, there are upper and lower tools. Each tool can have its own regeneration history. And the upper tool type has an optional cassette slot.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | EntityID | Unique identifier |
| `Width` | int | Width defines the tile width this tool can press |
| `Height` | int | Height defines the tile height this tool can press |
| `Type` | string | Type represents the tool type, e.g., "MASS", "FC", "GTC", etc. |
| `Code` | string | Code is the unique tool code/identifier, "G01", "12345", etc. |
| `CyclesOffset` | int64 | CyclesOffset is an offset added to the cycles count |
| `Cycles` | int64 | Cycles indicates how many cycles this tool has done |
| `LastRegeneration` | EntityID | ID of the last regeneration |
| `Regenerating` | bool | A regeneration resets the cycles counter, including the offset, back to zero |
| `Status` | string | Status represents the current state of the tool |
| `IsDead` | bool | IsDead indicates if the tool is dead/destroyed |

### ToolRegeneration
Tracks regeneration events for tools.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | EntityID | Unique identifier for the ToolRegeneration entity |
| `ToolID` | int64 | Indicates which tool has regenerated |
| `Start` | int64 | Start timestamp in milliseconds |
| `Stop` | int64 | Stop timestamp in milliseconds |
| `Cycles` | int64 | Cycles indicates the number cyles done before regeneration |

### Press
Represents a press machine with its associated tools and cassettes.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | PressNumber | Press number, required |
| `SlotUp` | EntityID | Upper tool entity ID, required |
| `SlotDown` | EntityID | Lower tool entity ID, required |
| `LastRegeneration` | EntityID | Tools last regeneration (entity) ID, optional |
| `StartCycles` | int64 | Press cycles since last regeneration, optional |
| `Cycles` | int64 | Current press cycles, required |
| `Type` | PressType | Type of press, e.g., "SACMI", "SITI" |

### Cycle
Tracks cycle events for press machines.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | EntityID | Unique identifier for the Cycle entity |
| `PressNumber` | PressNumber | Indicates which press machine performed the cycles |
| `Cycles` | int64 | Number of (partial) cycles |
| `Start` | int64 | Start timestamp in milliseconds |
| `Stop` | int64 | Stop timestamp in milliseconds |

### PressRegeneration
Tracks regeneration events for press machines.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | EntityID | Unique identifier for the PressRegeneration entity |
| `PressNumber` | PressNumber | Indicates which press has regenerated |
| `Start` | int64 | Start timestamp in milliseconds |
| `Stop` | int64 | Stop timestamp in milliseconds |
| `Cycles` | int64 | Cycles indicates the number cyles done before regeneration |

### User
Represents a user entity with relevant information.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | TelegramID | Unique Telegram ID for the user |
| `Name` | string | User's display name |
| `ApiKey` | string | Unique API key for the user |
| `LastFeed` | EntityID | ID of the last feed accessed by the user |

### Cookie
Represents a user session with authentication information.

| Field | Type | Description |
|-------|------|-------------|
| `UserAgent` | string | User agent string of the client |
| `Value` | string | Unique UUID cookie value |
| `UserID` | TelegramID | Associated Telegram ID |
| `LastLogin` | int64 | Last login timestamp in milliseconds |

### Session
Represents an in-memory session.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | EntityID | Unique session ID |

