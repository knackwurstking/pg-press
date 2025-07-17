// Package pgvis provides a PostgreSQL visualization and management system.
// It includes functionality for user management, trouble reporting, activity feeds,
// and session management through a web interface and CLI tools.
//
// The package is organized into several key components:
//   - User management and authentication
//   - Trouble report creation and tracking
//   - Activity feed system
//   - Session and cookie management
//   - Database abstraction layer
//
// All database operations use SQLite as the backend storage.
package pgvis

import "strings"

const (
	// MinAPIKeyLength defines the minimum length required for API keys
	MinAPIKeyLength = 32
)

// maskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
// This function is used to safely display sensitive information like API keys and tokens.
func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
