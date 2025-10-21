package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetInternelServerError(err error, message string) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusInternalServerError, message+": "+err.Error())
}
