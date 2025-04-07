package httperrors

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func NewUnderConstruction() *echo.HTTPError {
	return echo.NewHTTPError(http.StatusInternalServerError, "This route is under construction")
}
