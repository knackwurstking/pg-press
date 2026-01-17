package troublereports

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// TODO: Continue here...
func GetAttachmentsPreview(c echo.Context) *echo.HTTPError {
	// This would normally preview attachments, but for now return not implemented
	return echo.NewHTTPError(http.StatusNotImplemented, "Attachments preview not implemented")
}
