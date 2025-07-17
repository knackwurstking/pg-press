# Go Code Style Guide for pg-vis

## Table of Contents

- [General Principles](#general-principles)
- [Naming Conventions](#naming-conventions)
- [File Organization](#file-organization)
- [Error Handling](#error-handling)
- [Documentation](#documentation)
- [Database Operations](#database-operations)
- [HTTP Handlers](#http-handlers)
- [Testing](#testing)
- [Security](#security)

## General Principles

### 1. Follow Go Standards

- Use `go fmt` and `goimports` for all code formatting
- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `go vet` to catch common mistakes
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### 2. Code Organization

- Group related functionality into packages
- Keep functions small and focused on a single responsibility
- Use interfaces to define contracts between components
- Avoid circular dependencies between packages

## Naming Conventions

### Variables and Functions

```go
// Use camelCase for private variables and functions
var userCount int
func getUserByID(id int64) (*User, error) {}

// Use PascalCase for exported variables and functions
var MinAPIKeyLength = 32
func NewUser(telegramID int64, userName string) *User {}
```

### Constants

```go
// Use descriptive names for constants
const (
    MinUserNameLength = 1
    MaxUserNameLength = 100
    DefaultTimeout    = 30 * time.Second
)

// Group related constants
const (
    // HTTP status codes
    StatusOK           = 200
    StatusNotFound     = 404
    StatusServerError  = 500
)
```

### Types and Structs

```go
// Use PascalCase for exported types
type User struct {
    TelegramID int64  `json:"telegram_id"`
    UserName   string `json:"user_name"`
}

// Use descriptive names for interfaces
type UserRepository interface {
    Get(id int64) (*User, error)
    Add(user *User) error
}
```

## File Organization

### Package Structure

```
package/
├── models.go          # Data structures and their methods
├── repository.go      # Database access layer
├── service.go         # Business logic layer
├── handlers.go        # HTTP handlers
├── validation.go      # Validation functions
└── errors.go          # Package-specific errors
```

### Import Organization

```go
package main

import (
    // Standard library imports first
    "context"
    "fmt"
    "net/http"
    "time"

    // Third-party imports second
    "github.com/labstack/echo/v4"
    "github.com/charmbracelet/log"

    // Local imports last
    "github.com/knackwurstking/pg-vis/pgvis"
)
```

### File Headers

```go
// Package description at the top of each package's main file
// Package pgvis provides core business logic for the pg-vis application.
//
// This package includes user management, trouble reporting, and activity
// tracking functionality with a focus on security and maintainability.
package pgvis
```

## Error Handling

### Use Standardized Error Types

```go
// Use the centralized error types from pgvis/error.go
func ValidateUser(user *User) error {
    multiErr := pgvis.NewMultiError()

    if user.UserName == "" {
        multiErr.Add(pgvis.NewValidationError("user_name", "cannot be empty", user.UserName))
    }

    if multiErr.HasErrors() {
        return multiErr
    }

    return nil
}
```

### Error Wrapping

```go
// Wrap errors to provide context
func GetUser(id int64) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        return nil, pgvis.WrapErrorf(err, "failed to get user with ID %d", id)
    }
    return user, nil
}
```

### Database Errors

```go
// Use DatabaseError for database operations
func (r *UserRepository) Add(user *User) error {
    _, err := r.db.Exec(insertUserQuery, user.TelegramID, user.UserName)
    if err != nil {
        return pgvis.NewDatabaseError("insert", "users",
            "failed to insert user", err)
    }
    return nil
}
```

## Documentation

### Function Documentation

```go
// GetUser retrieves a user by their Telegram ID.
//
// Parameters:
//   - telegramID: The unique Telegram identifier for the user
//
// Returns:
//   - *User: The requested user
//   - error: pgvis.ErrNotFound if user doesn't exist, or database error
func GetUser(telegramID int64) (*User, error) {
    // Implementation
}
```

### Struct Documentation

```go
// User represents a system user with Telegram integration.
// Users can authenticate via API keys and have different permission levels.
type User struct {
    // TelegramID is the unique Telegram user identifier
    TelegramID int64 `json:"telegram_id"`

    // UserName is the display name for the user
    UserName string `json:"user_name"`

    // ApiKey is used for authentication (optional)
    ApiKey string `json:"api_key"`
}
```

### Package Documentation

```go
// Package routes provides HTTP route handlers for the pg-vis web interface.
//
// This package implements the web layer of the application, handling HTTP
// requests, authentication, and response formatting. It integrates with
// the pgvis package for business logic operations.
//
// Key features:
//   - User authentication via cookies and API keys
//   - RESTful API endpoints for data operations
//   - Web interface for trouble report management
//   - Activity feed display
package routes
```

## Database Operations

### Query Constants

```go
const (
    // Group related queries together
    selectUserByIDQuery    = `SELECT * FROM users WHERE telegram_id = ?`
    selectUserByAPIKeyQuery = `SELECT * FROM users WHERE api_key = ?`
    insertUserQuery        = `INSERT INTO users (telegram_id, user_name, api_key) VALUES (?, ?, ?)`
    updateUserQuery        = `UPDATE users SET user_name = ?, api_key = ? WHERE telegram_id = ?`
    deleteUserQuery        = `DELETE FROM users WHERE telegram_id = ?`
)
```

### Parameterized Queries

```go
// NEVER use string concatenation for SQL queries
// BAD:
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", userID)

// GOOD:
query := "SELECT * FROM users WHERE id = ?"
row := db.QueryRow(query, userID)
```

### Transaction Handling

```go
func (r *Repository) UpdateUserWithHistory(user *User) error {
    tx, err := r.db.Begin()
    if err != nil {
        return pgvis.NewDatabaseError("begin_transaction", "",
            "failed to start transaction", err)
    }
    defer tx.Rollback() // Safe to call even after commit

    // Perform operations
    if err := r.updateUser(tx, user); err != nil {
        return err
    }

    if err := r.addHistory(tx, user); err != nil {
        return err
    }

    return tx.Commit()
}
```

## HTTP Handlers

### Handler Structure

```go
func GetUserHandler(db *pgvis.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Parse and validate input
        userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
        if err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, "invalid user ID")
        }

        // Business logic
        user, err := db.Users.Get(userID)
        if err != nil {
            if pgvis.IsNotFound(err) {
                return echo.NewHTTPError(http.StatusNotFound, "user not found")
            }
            return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
        }

        // Return response
        return c.JSON(http.StatusOK, user)
    }
}
```

### Error Response Formatting

```go
func HandleError(c echo.Context, err error) error {
    code := pgvis.GetHTTPStatusCode(err)

    response := map[string]interface{}{
        "error":  err.Error(),
        "code":   code,
        "status": http.StatusText(code),
    }

    if pgvis.IsValidationError(err) {
        // Add validation details
        response["details"] = err
    }

    return c.JSON(code, response)
}
```

## Testing

### Test File Naming

```go
// user_test.go for testing user.go
// repository_test.go for testing repository.go
```

### Test Function Naming

```go
func TestUser_Validate_Success(t *testing.T) {}
func TestUser_Validate_EmptyName_ReturnsError(t *testing.T) {}
func TestUserRepository_Get_UserExists_ReturnsUser(t *testing.T) {}
func TestUserRepository_Get_UserNotFound_ReturnsError(t *testing.T) {}
```

### Test Structure

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name    string
        user    *User
        wantErr bool
        errType string
    }{
        {
            name: "valid user",
            user: &User{TelegramID: 123, UserName: "test"},
            wantErr: false,
        },
        {
            name: "empty username",
            user: &User{TelegramID: 123, UserName: ""},
            wantErr: true,
            errType: "validation",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.user.Validate()
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType == "validation" {
                    assert.True(t, pgvis.IsValidationError(err))
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Security

### Input Validation

```go
// Always validate input at the handler level
func CreateUserHandler(db *pgvis.DB) echo.HandlerFunc {
    return func(c echo.Context) error {
        var req CreateUserRequest
        if err := c.Bind(&req); err != nil {
            return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
        }

        // Validate the request
        if err := req.Validate(); err != nil {
            return HandleError(c, err)
        }

        // Continue with business logic
    }
}
```

### Sensitive Data Handling

```go
// Mask sensitive data in logs
func (u *User) String() string {
    apiKey := "none"
    if u.HasAPIKey() {
        apiKey = pgvis.maskString(u.ApiKey)
    }
    return fmt.Sprintf("User{ID: %d, Name: %s, APIKey: %s}",
        u.TelegramID, u.UserName, apiKey)
}
```

### HTML Escaping

```go
import "html"

// Always escape user-generated content in HTML
func CreateFeedEntry(userName, content string) string {
    escapedName := html.EscapeString(userName)
    escapedContent := html.EscapeString(content)
    return fmt.Sprintf("<p>%s: %s</p>", escapedName, escapedContent)
}
```

## Code Review Checklist

- [ ] Code follows Go formatting standards (`go fmt`, `goimports`)
- [ ] All exported functions and types are documented
- [ ] Error handling uses standardized error types
- [ ] Database queries use parameterized statements
- [ ] Input validation is present and comprehensive
- [ ] Sensitive data is properly masked in logs
- [ ] HTTP status codes are appropriate
- [ ] Tests cover both success and error cases
- [ ] No TODO comments remain in production code
- [ ] Dependencies are minimal and justified

## Tools and Commands

### Development Workflow

```bash
# Format code
go fmt ./...
goimports -w .

# Check for issues
go vet ./...
golint ./...

# Run tests
go test ./...
go test -race ./...

# Build
go build ./...
```

### Pre-commit Hooks

Consider setting up pre-commit hooks to automatically run:

- `go fmt`
- `goimports`
- `go vet`
- `go test`

This ensures code quality and consistency across all commits.
