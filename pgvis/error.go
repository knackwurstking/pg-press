// Package pgvis error handling utilities and types.
//
// This file defines standard error types, error utilities, and HTTP status code mappings
// for the pgvis application. It provides a consistent error handling strategy across
// all application components.
//
// Error Categories:
//   - Data access errors (not found, duplicates, etc.)
//   - Validation errors with field-specific details
//   - Database operation errors
//   - Network and API errors
//
// Custom Error Types:
//   - ValidationError: Field-specific validation failures
//   - DatabaseError: Database operation failures with context
package pgvis

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard errors for common operations.
// These sentinel errors can be used with errors.Is() for error checking.
var (
	// Data access errors
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	// Authentication and authorization errors
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Network and API errors
	ErrRateLimited        = errors.New("rate limited")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// ValidationError represents a field-specific validation error.
// It provides detailed information about which field failed validation,
// the validation message, and optionally the invalid value.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error with the specified field,
// message, and optional value that caused the validation failure.
func NewValidationError(field, message string, value any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// DatabaseError represents database-related errors with contextual information.
// It includes the operation being performed, the table involved (if applicable),
// and the underlying cause error.
type DatabaseError struct {
	Operation string `json:"operation"`
	Table     string `json:"table,omitempty"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
}

// Error implements the error interface for DatabaseError.
func (e *DatabaseError) Error() string {
	if e.Table != "" {
		return fmt.Sprintf("database error during %s on table '%s': %s", e.Operation, e.Table, e.Message)
	}
	return fmt.Sprintf("database error during %s: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying cause error, enabling error unwrapping
// with errors.Is() and errors.As().
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// NewDatabaseError creates a new database error with contextual information
// about the failed operation, table, and underlying cause.
func NewDatabaseError(operation, table, message string, cause error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Table:     table,
		Message:   message,
		Cause:     cause,
	}
}

// Utility functions for error handling and type checking.
// These functions provide convenient ways to check for specific error types
// and extract information from errors.

// IsNotFound checks if an error represents a "not found" condition
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsDatabaseError checks if an error is a database error
func IsDatabaseError(err error) bool {
	var dbErr *DatabaseError
	return errors.As(err, &dbErr)
}

// GetHTTPStatusCode returns the appropriate HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for specific error types
	switch {
	case IsNotFound(err):
		return http.StatusNotFound
	case IsValidationError(err):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrInvalidToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrTokenExpired):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, ErrServiceUnavailable):
		return http.StatusServiceUnavailable
	case IsDatabaseError(err):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// WrapError wraps an error with additional context message.
// Returns nil if the input error is nil.
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapErrorf wraps an error with formatted additional context message.
// Returns nil if the input error is nil.
func WrapErrorf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}
