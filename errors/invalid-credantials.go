package errors

type InvalidCredentialsError struct {
	message string
}

func NewInvalidCredentialsError(message string) *InvalidCredentialsError {
	return &InvalidCredentialsError{message: message}
}

func (ic *InvalidCredentialsError) Error() string {
	return "invalid credentials: " + ic.message
}

func IsInvalidCredentialsError(err error) bool {
	if err == nil {
		return false
	}

	if _, ok := err.(*InvalidCredentialsError); ok {
		return true
	}

	return false
}
