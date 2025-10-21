package errors

type ValidationError struct {
	message string
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{message: message}
}

func (v *ValidationError) Error() string {
	return "validation error: " + v.message
}

func IsNotValidationError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*ValidationError); ok {
		return true
	}

	return false
}
