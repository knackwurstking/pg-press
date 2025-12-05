package errors

import "errors"

// ErrorType represents the type of error for classification
type ErrorType string

const (
	ErrorTypeGeneric ErrorType = "generic"

	// Database error types
	ErrorTypeCount      ErrorType = "count"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeExists     ErrorType = "exists"
)

var (
	ErrValidation error = errors.New(string(ErrorTypeValidation))
	ErrExists     error = errors.New(string(ErrorTypeExists))
)
