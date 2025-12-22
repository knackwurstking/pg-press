package errors

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthorizationError struct {
	Message string
}

func NewAuthorizationError(format string, v ...any) *AuthorizationError {
	return &AuthorizationError{
		Message: fmt.Sprintf(format, v...),
	}
}

func (a *AuthorizationError) Error() string {
	return a.Message
}

func (a *AuthorizationError) MasterError() *MasterError {
	return NewMasterError(a).SetCode(http.StatusUnauthorized)
}

type NotFoundError struct {
	Message string
}

func NewNotFoundError(format string, v ...any) *NotFoundError {
	return &NotFoundError{
		Message: fmt.Sprintf(format, v...),
	}
}

func (n *NotFoundError) Error() string {
	return n.Message
}

func (n *NotFoundError) MasterError() *MasterError {
	return NewMasterError(n).SetCode(http.StatusNotFound)
}

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
	return NewMasterError(ve).SetCode(http.StatusBadRequest)
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

func (ee *ExistsError) Error() string {
	if ee.Value != nil {
		return fmt.Sprintf("%s with value %#v already exists", ee.Name, ee.Value)
	}

	return fmt.Sprintf("%s already exists", ee.Name)
}

func (ee *ExistsError) MasterError() *MasterError {
	return NewMasterError(ee).SetCode(http.StatusConflict)
}

type MasterError struct {
	err  error // Err is required
	code int   // Code is optional
}

func NewMasterError(err error) *MasterError {
	if err == nil {
		panic("cannot create MasterError with nil error")
	}

	if e, ok := err.(*MasterError); ok {
		return e
	}

	code := http.StatusInternalServerError
	switch err {
	case sql.ErrNoRows:
		code = http.StatusNotFound
	case sql.ErrConnDone:
		code = http.StatusServiceUnavailable
	case sql.ErrTxDone:
		code = http.StatusConflict
	}

	return &MasterError{
		err:  err,
		code: code,
	}
}

func (e *MasterError) Error() string {
	if e == nil {
		panic("error is nil?")
	}

	return e.err.Error()
}

func (e *MasterError) Err() error {
	return e.err
}

func (e *MasterError) Code() int {
	return e.code
}

func (e *MasterError) SetCode(code int) *MasterError {
	e.code = code
	return e
}

func (e *MasterError) Wrap(format string, a ...any) *MasterError {
	msg := fmt.Sprintf(format, a...)
	if msg == "" {
		return e
	}
	wrapped := fmt.Errorf("%s: %v", msg, e.err)
	return &MasterError{err: wrapped, code: e.code}
}

func (e *MasterError) Echo() *echo.HTTPError {
	return echo.NewHTTPError(e.code, e.err.Error())
}

func (e *MasterError) WrapEcho(format string, a ...any) *echo.HTTPError {
	msg := fmt.Sprintf(format, a...)
	return echo.NewHTTPError(e.code, msg+": "+e.err.Error())
}

func (e *MasterError) IsValidationError() bool {
	_, ok := e.err.(*ValidationError)
	return ok
}

func (e *MasterError) IsExistsError() bool {
	_, ok := e.err.(*ExistsError)
	return ok
}
