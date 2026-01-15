package troublereports

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// TODO: Delete trouble reports, take from query parameter "id"
func DeleteTroubleReport(c echo.Context) *echo.HTTPError {
	// This would normally delete a trouble report, but for now return not implemented
	return echo.NewHTTPError(http.StatusNotImplemented, "Delete functionality not implemented")
}
