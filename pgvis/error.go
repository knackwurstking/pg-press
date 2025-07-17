// Package pgvis error handling utilities and types.
//
// This file defines standard error types, error utilities, and HTTP status code mappings
// for the pgvis application. It provides a consistent error handling strategy across
// all application components.
//
// Error Categories:
//   - Data access errors (not found, duplicates, etc.)
//   - Authentication and authorization errors
//   - Validation errors with field-specific details
//   - Database operation errors
//   - File operation errors
//   - Network and API errors
//
// Custom Error Types:
//   - ValidationError: Field-specific validation failures
//   - AuthError: Authentication and authorization failures
//   - DatabaseError: Database operation failures with context
//   - APIError: HTTP API errors with status codes
//   - MultiError: Collection of multiple errors
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
	ErrInvalidData   = errors.New("invalid data")
	ErrDuplicateKey  = errors.New("duplicate key")

	// Authentication and authorization errors
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Validation errors
	ErrMissingField  = errors.New("missing required field")
	ErrInvalidFormat = errors.New("invalid format")
	ErrValueTooLarge = errors.New("value too large")
	ErrValueTooSmall = errors.New("value too small")
	ErrInvalidLength = errors.New("invalid length")

	// Database errors
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrConstraintViolation = errors.New("constraint violation")

	// File operation errors
	ErrFileNotFound    = errors.New("file not found")
	ErrFilePermission  = errors.New("file permission denied")
	ErrFileSize        = errors.New("file size exceeded")
	ErrInvalidFileType = errors.New("invalid file type")

	// Network and API errors
	ErrNetworkTimeout     = errors.New("network timeout")
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

// AuthError represents authentication and authorization errors.
// It distinguishes between authentication failures (invalid credentials)
// and authorization failures (insufficient permissions).
type AuthError struct {
	Type    string `json:"type"` // "authentication" or "authorization"
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// Error implements the error interface for AuthError.
func (e *AuthError) Error() string {
	if e.UserID != "" {
		return fmt.Sprintf("%s error for user '%s': %s", e.Type, e.UserID, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}

// NewAuthenticationError creates a new authentication error for cases where
// user credentials are invalid or authentication fails.
func NewAuthenticationError(message, userID string) *AuthError {
	return &AuthError{
		Type:    "authentication",
		Message: message,
		UserID:  userID,
	}
}

// NewAuthorizationError creates a new authorization error for cases where
// a user is authenticated but lacks sufficient permissions.
func NewAuthorizationError(message, userID string) *AuthError {
	return &AuthError{
		Type:    "authorization",
		Message: message,
		UserID:  userID,
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

// APIError represents API-related errors with HTTP status codes.
// It provides a standardized way to represent HTTP errors with
// appropriate status codes and optional additional details.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}

// StatusCode returns the HTTP status code associated with this error.
func (e *APIError) StatusCode() int {
	return e.Code
}

// NewAPIError creates a new API error with the specified HTTP status code,
// message, and optional details object.
func NewAPIError(code int, message string, details any) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// MultiError represents multiple errors combined into a single error.
// It's useful for operations that can encounter multiple validation
// or processing errors that should all be reported together.
type MultiError struct {
	Errors []error `json:"errors"`
}

// Error implements the error interface for MultiError.
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return "no errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("multiple errors: %d total", len(e.Errors))
}

// Add adds an error to the multi-error collection.
// Nil errors are ignored.
func (e *MultiError) Add(err error) {
	if err != nil {
		e.Errors = append(e.Errors, err)
	}
}

// HasErrors returns true if the MultiError contains any errors.
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// First returns the first error in the collection or nil if there are no errors.
func (e *MultiError) First() error {
	if len(e.Errors) > 0 {
		return e.Errors[0]
	}
	return nil
}

// NewMultiError creates a new empty MultiError collection.
func NewMultiError() *MultiError {
	return &MultiError{
		Errors: make([]error, 0),
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

// IsAuthError checks if an error is an authentication/authorization error
func IsAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// IsDatabaseError checks if an error is a database error
func IsDatabaseError(err error) bool {
	var dbErr *DatabaseError
	return errors.As(err, &dbErr)
}

// IsAPIError checks if an error is an API error
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// GetHTTPStatusCode returns the appropriate HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for API error with explicit status code
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode()
	}

	// Check for specific error types
	switch {
	case IsNotFound(err):
		return http.StatusNotFound
	case IsValidationError(err):
		return http.StatusBadRequest
	case IsAuthError(err):
		var authErr *AuthError
		if errors.As(err, &authErr) {
			if authErr.Type == "authentication" {
				return http.StatusUnauthorized
			}
			return http.StatusForbidden
		}
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
