package troublereports

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetSharePDF(c echo.Context) *echo.HTTPError {
	// This would normally generate a PDF, but for now return not implemented
	return echo.NewHTTPError(http.StatusNotImplemented, "PDF generation not implemented")
}
