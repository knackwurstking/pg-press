package errors

import (
	"errors"
	"fmt"
	"net/http"
)

func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if IsNotFoundError(err) {
		return http.StatusNotFound
	}

	if IsAlreadyExistsError(err) {
		return http.StatusConflict
	}

	if IsInvalidCredentialsError(err) {
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}

func Wrap(err error, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	if err == nil {
		return errors.New(msg)
	}
	// Format the wrapped error with a concise message that starts with lowercase
	return fmt.Errorf("%s: %v", msg, err)
}
