package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MasterError represents a unified error type that encompasses all error handling patterns
type MasterError struct {
	Err  error
	Code int
	Type ErrorType
}

func NewMasterError(err error) *MasterError {
	if e, ok := err.(*MasterError); ok {
		return e
	}

	me := &MasterError{
		Code: http.StatusInternalServerError,
		Err:  err,
		Type: ErrorTypeGeneric,
	}

	switch err {
	case sql.ErrNoRows:
		me.Type = ErrorTypeNotFound
	case ErrValidation:
		me.Type = ErrorTypeValidation
	case ErrExists:
		me.Type = ErrorTypeExists
	default:
		slog.Debug(fmt.Sprintf("New unknown error: %#v", err))
	}

	switch me.Type {
	case ErrorTypeNotFound:
		me.Code = http.StatusNotFound
	case ErrorTypeExists:
		me.Code = http.StatusBadRequest
	case ErrorTypeValidation:
		me.Code = http.StatusBadRequest
	}

	return me
}

func (e *MasterError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Type, e.Err)
	}

	return fmt.Sprintf("%d: %s", e.Code, e.Type)
}

func (e *MasterError) Unwrap() error {
	return e.Err
}

func (e *MasterError) Echo() *echo.HTTPError {
	code := e.Code
	if e.Code == 0 {
		code = http.StatusInternalServerError
	}
	return echo.NewHTTPError(code, e.Error())
}

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
