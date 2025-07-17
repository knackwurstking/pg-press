package pgvis

import "strings"

const (
	MinAPIKeyLength = 32
)

// Helper function to mask sensitive strings (defined in cookie.go, reused here)
func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}
