// Package pgvis provides error handling utilities and types.
package pgvis

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard errors for common operations.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrRateLimited        = errors.New("rate limited")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// ValidationError represents a field-specific validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string, value any) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// DatabaseError represents database-related errors with contextual information.
type DatabaseError struct {
	Operation string `json:"operation"`
	Table     string `json:"table,omitempty"`
	Message   string `json:"message"`
	Cause     error  `json:"-"`
}

func (e *DatabaseError) Error() string {
	if e.Table != "" {
		return fmt.Sprintf("database error during %s on table '%s': %s: %s", e.Operation, e.Table, e.Message, e.Cause)
	}
	return fmt.Sprintf("database error during %s: %s: %s", e.Operation, e.Message, e.Cause)
}

// NewDatabaseError creates a new database error.
func NewDatabaseError(operation, table, message string, cause error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Table:     table,
		Message:   message,
		Cause:     cause,
	}
}

// IsValidationError checks if an error is a validation error.
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// IsDatabaseError checks if an error is a database error.
func IsDatabaseError(err error) bool {
	var dbErr *DatabaseError
	return errors.As(err, &dbErr)
}

// GetHTTPStatusCode returns the appropriate HTTP status code for an error.
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, ErrNotFound):
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
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
