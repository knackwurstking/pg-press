package errors

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MasterError represents a unified error type that encompasses all error handling patterns
type MasterError struct {
	Err  error // Err is required
	Code int   // Code is optional
}

func NewMasterError(err error, code int) *MasterError {
	if e, ok := err.(*MasterError); ok {
		if code == 0 {
			code = e.Code
		}
		return &MasterError{
			Err:  e.Err,
			Code: code,
		}
	}

	if code == 0 {
		switch err {
		case sql.ErrNoRows:
			code = http.StatusNotFound
		default:
			code = http.StatusInternalServerError
		}
	}

	return &MasterError{
		Err:  err,
		Code: code,
	}
}

func (e *MasterError) Error() string {
	if e == nil {
		return "error is nil?"
	}

	if e.Code > 0 {
		return fmt.Sprintf("%d: %s", e.Code, e.Err.Error())
	}

	return e.Err.Error()
}

func (e *MasterError) Wrap(format string, a ...any) *MasterError {
	msg := fmt.Sprintf(format, a...)
	if msg == "" {
		return e
	}
	wrapped := fmt.Errorf("%s: %w", msg, e.Err)
	return &MasterError{Err: wrapped, Code: e.Code}
}

func (e *MasterError) Echo() *echo.HTTPError {
	return echo.NewHTTPError(e.Code, e.Err.Error())
}

func (e *MasterError) WrapEcho(format string, a ...any) *echo.HTTPError {
	msg := fmt.Sprintf(format, a...)
	return echo.NewHTTPError(e.Code, msg+": "+e.Err.Error())
}
