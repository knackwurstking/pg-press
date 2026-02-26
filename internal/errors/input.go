package errors

type InputError struct {
	InputID string
	Message string
}

func NewInputError(inputID, message string) *InputError {
	return &InputError{
		InputID: inputID,
		Message: message,
	}
}

func (e *InputError) Error() string {
	return e.Message
}
