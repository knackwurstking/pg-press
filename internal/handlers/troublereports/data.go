package troublereports

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetData(c echo.Context) *echo.HTTPError {
	// This would normally return HTMX data, but for now return not implemented
	return echo.NewHTTPError(http.StatusNotImplemented, "HTMX data not implemented")
}

func DeleteTroubleReport(c echo.Context) *echo.HTTPError {
	// This would normally delete a trouble report, but for now return not implemented
	return echo.NewHTTPError(http.StatusNotImplemented, "Delete functionality not implemented")
}
