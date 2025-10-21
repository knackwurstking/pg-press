package errors

type NotFoundError struct {
	message string
}

func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{message: message}
}

func (nf *NotFoundError) Error() string {
	return "not found: " + nf.message
}

func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*NotFoundError); ok {
		return true
	}

	return false
}
