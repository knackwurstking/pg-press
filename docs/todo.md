# Common Functions Pattern in Handlers

## Overview
Analysis of the handlers directory reveals a significant pattern of code duplication across multiple handler implementations. This document outlines the common functions and patterns found throughout the handler files.

## Identified Patterns

### 1. Route Registration Pattern
All handlers implement the same `RegisterRoutes` method structure:
```go
func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
    ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
        // ... route registrations 
    })
}
```

### 2. Handler Creation Pattern
All handlers follow the same `NewHandler` function pattern:
```go
func NewHandler(r *services.Registry) *Handler {
    return &Handler{
        registry: r,
    }
}
```

### 3. Similar Function Implementations
Several handlers implement very similar functions:
- `HTMXDeleteNote` - Handles note deletion with feed creation
- `HTMXGetNotesGrid` - Renders notes grid with tools data
- `HTMXDeleteRegeneration` - Handles regeneration deletion with feed creation
- Functions for getting data from services and rendering templates

### 4. Error Handling Pattern
All handlers implement similar error handling patterns:
- Getting user from context using `utils.GetUserFromContext(c)`
- Handling service errors with `errors.Handler(err, "description")`
- Consistent redirect patterns for authentication

## Recommendations
1. Create a shared base handler or utility functions for route registration
2. Implement common HTTP handler functions in a shared library 
3. Extract repeated error handling patterns into reusable functions