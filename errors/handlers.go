package errors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetHTTPStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if e, ok := err.(*echo.HTTPError); ok {
		return e.Code
	}

	if e, ok := err.(*DBError); ok {
		if e.Typ == DBTypeNotFound {
			return http.StatusNotFound
		}
	}

	return http.StatusInternalServerError
}

func HandlerError(err error, format string, a ...any) *echo.HTTPError {
	if err == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, Wrap(nil, format, a...))
	}

	if e, ok := err.(*echo.HTTPError); ok {
		return e
	}

	statusCode := GetHTTPStatusCodeFromError(err)
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	return echo.NewHTTPError(statusCode, Wrap(err, format, a...))
}

func NewRenderError(err error, name string) *echo.HTTPError {
	return HandlerError(err, "%s", name)
}
