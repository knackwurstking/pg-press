package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Database error constructors
func NewDatabaseSelectError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("database select: %v", err)
}

func NewDatabaseInsertError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("database insert: %v", err)
}

func NewDatabaseUpdateError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("database update: %v", err)
}

func NewDatabaseDeleteError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("database delete: %v", err)
}

func GetHTTPStatusCodeFromError(err error) int {
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

func NotFound(err error, format string, a ...any) *echo.HTTPError {
	if err == nil {
		return echo.NewHTTPError(http.StatusNotFound, Wrap(nil, format, a...))
	}

	return echo.NewHTTPError(http.StatusNotFound, Wrap(err, format, a...))
}

func BadRequest(err error, format string, a ...any) *echo.HTTPError {
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, Wrap(nil, format, a...))
	}

	return echo.NewHTTPError(http.StatusBadRequest, Wrap(err, format, a...))
}

func Handler(err error, format string, a ...any) *echo.HTTPError {
	if err == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, Wrap(nil, format, a...))
	}
	statusCode := GetHTTPStatusCodeFromError(err)
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}
	return echo.NewHTTPError(statusCode, Wrap(err, format, a...))
}

func Wrap(err error, format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	if err == nil {
		return errors.New(msg)
	}
	if format == "" {
		return err
	}
	// Format the wrapped error with a concise message that starts with lowercase
	return fmt.Errorf("%s: %v", msg, err)
}
