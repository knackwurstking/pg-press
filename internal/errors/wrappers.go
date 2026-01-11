package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Wrap wraps an error with additional context
func Wrap(err error, format string, a ...any) error {
	msg := ""
	if format != "" {
		msg = fmt.Sprintf(format, a...)
	}

	if msg == "" {
		return err
	}

	if err == nil {
		return errors.New(msg)
	}

	// Format the wrapped error with a concise message that starts with lowercase
	return fmt.Errorf("%s: %v", msg, err)
}

func NewRenderError(err error, name string) *echo.HTTPError {
	return echo.NewHTTPError(
		http.StatusInternalServerError,
		fmt.Sprintf("render %s: %s", name, err.Error()),
	)
}
