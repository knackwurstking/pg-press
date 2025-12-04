package errors

// ValidationError represents a validation error
type ValidationError struct {
	message string
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

// Error returns the error message
func (v *ValidationError) Error() string {
	return "validation error: " + v.message
}

// IsNotValidationError checks if an error is a validation error
func IsNotValidationError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*ValidationError); ok {
		return true
	}

	return false
}
