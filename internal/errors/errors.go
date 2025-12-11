package errors

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ValidationError struct {
	Message string
}

func NewValidationError(format string, v ...any) *ValidationError {
	return &ValidationError{
		Message: fmt.Sprintf(format, v...),
	}
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

func (ve *ValidationError) MasterError() *MasterError {
	return NewMasterError(ve, 0)
}

type ExistsError struct {
	Name  string
	Value any
}

func NewExistsError(name string, v any) *ExistsError {
	return &ExistsError{
		Name:  name,
		Value: v,
	}
}

func (ee *ExistsError) MasterError() *MasterError {
	return NewMasterError(ee, 0)
}

func (ee *ExistsError) Error() string {
	if ee.Value != nil {
		return fmt.Sprintf("%s with value %#v already exists", ee.Name, ee.Value)
	}

	return fmt.Sprintf("%s already exists", ee.Name)
}

// MasterError represents a unified error type that encompasses all error handling patterns
type MasterError struct {
	Err  error // Err is required
	Code int   // Code is optional
}

func NewMasterError(err error, code int) *MasterError {
	if err == nil {
		panic("cannot create MasterError with nil error")
	}

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
			switch err.(type) {
			case *ValidationError:
				code = http.StatusBadRequest
			case *ExistsError:
				code = http.StatusConflict
			default:
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					code = http.StatusConflict
				} else {
					code = http.StatusInternalServerError
				}
			}
		}
	}

	return &MasterError{
		Err:  err,
		Code: code,
	}
}

func (e *MasterError) Error() string {
	if e == nil {
		panic("error is nil?")
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

func (e *MasterError) IsValidationError() bool {
	_, ok := e.Err.(*ValidationError)
	return ok
}

func (e *MasterError) IsExistsError() bool {
	_, ok := e.Err.(*ExistsError)
	return ok
}
