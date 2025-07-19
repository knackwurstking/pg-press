// Package pgvis provides a PostgreSQL visualization and management system.
package pgvis

import "strings"

const (
	// MinAPIKeyLength defines the minimum length required for API keys
	MinAPIKeyLength = 32
)

// maskString masks sensitive strings by showing only the first and last 4 characters.
// For strings with 8 or fewer characters, all characters are masked.
func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
