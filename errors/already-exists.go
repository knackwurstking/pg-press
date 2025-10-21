package errors

type AlreadyExistsError struct {
	message string
}

func NewAlreadyExistsError(message string) *AlreadyExistsError {
	return &AlreadyExistsError{message: message}
}

func (ae *AlreadyExistsError) Error() string {
	return "already exists: " + ae.message
}

func IsAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*AlreadyExistsError); ok {
		return true
	}

	return false
}
