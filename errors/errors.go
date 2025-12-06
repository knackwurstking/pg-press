package errors

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MasterError represents a unified error type that encompasses all error handling patterns
type MasterError struct {
	Err  error
	Code int
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
	if e.Code > 0 {
		return fmt.Sprintf("%d: %s", e.Code, e.Err.Error())
	}

	return e.Err.Error()
}

func (e *MasterError) Echo() *echo.HTTPError {
	return echo.NewHTTPError(e.Code, e.Err.Error())
}
