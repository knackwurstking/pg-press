package utils

import (
	"html"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/knackwurstking/pg-press/internal/errors"
)

// SanitizeString removes or escapes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove or escape potentially dangerous HTML characters
	s = html.EscapeString(s)

	// Optionally trim whitespace
	s = strings.TrimSpace(s)

	return s
}

// SanitizeURL removes or escapes potentially dangerous characters from a URL string
func SanitizeURL(u string) string {
	// First, escape HTML entities to prevent XSS
	u = html.EscapeString(u)

	// Validate URL format - basic check for protocol
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		// If it's not a proper URL, return empty to prevent malformed URLs
		return ""
	}

	// Parse and reconstruct URL to ensure it's valid
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}

	// Reconstruct the URL
	sanitized := parsed.String()
	if sanitized == "" {
		return ""
	}

	return sanitized
}

// SanitizeText removes or escapes potentially dangerous characters from text input
func SanitizeText(s string) string {
	// Remove or escape potentially dangerous HTML characters
	s = html.EscapeString(s)

	// Trim whitespace to prevent manipulation
	s = strings.TrimSpace(s)

	// Remove control characters except common whitespace (tab, newline, carriage return)
	var clean strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			clean.WriteRune(r)
		}
	}

	return clean.String()
}

// SanitizeFloat removes or escapes potentially dangerous characters from float input
func SanitizeFloat(s string) (float64, error) {
	// Clean the string from whitespace and common dangerous characters
	s = strings.TrimSpace(s)

	// Basic validation to ensure it's a valid float format
	if s == "" {
		return 0, nil
	}

	// Parse the float value
	f, err := parseFloat(s)
	if err != nil {
		return 0, err
	}

	return f, nil
}

// parseFloat is a helper to safely parse floats with custom validation
func parseFloat(s string) (float64, error) {
	// Remove common dangerous characters that could be used in injection attacks
	s = strings.ReplaceAll(s, "\x00", "") // Remove null bytes
	s = strings.ReplaceAll(s, "'", "")    // Remove single quotes for SQL injection
	s = strings.ReplaceAll(s, "\"", "")   // Remove double quotes for SQL injection
	s = strings.ReplaceAll(s, "--", "")   // Remove SQL comment markers

	// Ensure it's a valid float format
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

// SanitizeInt removes or escapes potentially dangerous characters from integer input
func SanitizeInt(s string) (int64, error) {
	// Clean the string from whitespace and common dangerous characters
	s = strings.TrimSpace(s)

	// Remove any non numeric characters except minus sign for negative numbers
	var clean strings.Builder
	for i, r := range s {
		if unicode.IsDigit(r) || (r == '-' && i == 0) {
			clean.WriteRune(r)
		}
	}

	s = clean.String()
	if s == "" {
		return 0, nil
	}

	// Parse the integer value
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// ValidateAndSanitizeString validates and sanitizes a string input
func ValidateAndSanitizeString(s string, maxLength int) (string, error) {
	// Sanitize the input
	sanitized := SanitizeString(s)

	// Validate length constraint (if specified)
	if maxLength > 0 && len(sanitized) > maxLength {
		return "", &errors.ValidationError{Message: "input exceeds maximum length of " + strconv.Itoa(maxLength)}
	}

	return sanitized, nil
}
