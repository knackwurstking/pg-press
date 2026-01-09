package errors

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// AuthorizationError represents an authorization-related error
type AuthorizationError struct {
	Message string
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(format string, v ...any) *AuthorizationError {
	return &AuthorizationError{
		Message: fmt.Sprintf(format, v...),
	}
}

func (a *AuthorizationError) Error() string {
	return a.Message
}

// HTTPError converts AuthorizationError to HTTPError with unauthorized status
func (a *AuthorizationError) HTTPError() *HTTPError {
	return NewHTTPError(a).SetCode(http.StatusUnauthorized)
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Message string
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(format string, v ...any) *NotFoundError {
	return &NotFoundError{
		Message: fmt.Sprintf(format, v...),
	}
}

func (n *NotFoundError) Error() string {
	return n.Message
}

// HTTPError converts NotFoundError to HTTPError with not found status
func (n *NotFoundError) HTTPError() *HTTPError {
	return NewHTTPError(n).SetCode(http.StatusNotFound)
}

// ValidationError represents a validation error
type ValidationError struct {
	Message string
}

// NewValidationError creates a new validation error
func NewValidationError(format string, v ...any) *ValidationError {
	return &ValidationError{
		Message: fmt.Sprintf(format, v...),
	}
}

func (ve *ValidationError) Error() string {
	return ve.Message
}

// HTTPError converts ValidationError to HTTPError with bad request status
func (ve *ValidationError) HTTPError() *HTTPError {
	return NewHTTPError(ve).SetCode(http.StatusBadRequest)
}

// ExistsError represents an error when a resource already exists
type ExistsError struct {
	Name  string
	Value any
}

// NewExistsError creates a new exists error
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

// HTTPError converts ExistsError to HTTPError with conflict status
func (ee *ExistsError) HTTPError() *HTTPError {
	return NewHTTPError(ee).SetCode(http.StatusConflict)
}

// HTTPError is a wrapper that provides HTTP status codes for errors
type HTTPError struct {
	err  error // Err is required
	code int   // Code is optional
}

// NewHTTPError creates a new HTTPError from an error
func NewHTTPError(err error) *HTTPError {
	if err == nil {
		panic("cannot create HTTPError with nil error")
	}

	if e, ok := err.(*HTTPError); ok {
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

	return &HTTPError{
		err:  err,
		code: code,
	}
}

// Error returns the underlying error's message
func (e *HTTPError) Error() string {
	if e == nil {
		panic("error is nil?")
	}

	return e.err.Error()
}

// Err returns the underlying error
func (e *HTTPError) Err() error {
	return e.err
}

// Code returns the associated HTTP status code
func (e *HTTPError) Code() int {
	return e.code
}

// SetCode sets the HTTP status code for this error
func (e *HTTPError) SetCode(code int) *HTTPError {
	e.code = code
	return e
}

// Wrap wraps the error with additional context
func (e *HTTPError) Wrap(format string, a ...any) *HTTPError {
	msg := fmt.Sprintf(format, a...)
	if msg == "" {
		return e
	}
	wrapped := fmt.Errorf("%s: %v", msg, e.err)
	return &HTTPError{err: wrapped, code: e.code}
}

// Echo converts this error to an echo HTTP error
func (e *HTTPError) Echo() *echo.HTTPError {
	return echo.NewHTTPError(e.code, e.err.Error())
}

// WrapEcho wraps the error and converts it to an echo HTTP error
func (e *HTTPError) WrapEcho(format string, a ...any) *echo.HTTPError {
	msg := fmt.Sprintf(format, a...)
	return echo.NewHTTPError(e.code, msg+": "+e.err.Error())
}

// IsValidationError checks if the error is a validation error
func (e *HTTPError) IsValidationError() bool {
	_, ok := e.err.(*ValidationError)
	return ok || e.code == http.StatusBadRequest
}

// IsExistsError checks if the error is an exists error
func (e *HTTPError) IsExistsError() bool {
	_, ok := e.err.(*ExistsError)
	return ok || e.code == http.StatusConflict
}

// IsNotFoundError checks if the error is a not found error
func (e *HTTPError) IsNotFoundError() bool {
	_, ok := e.err.(*NotFoundError)
	return ok || e.code == http.StatusNotFound
}
