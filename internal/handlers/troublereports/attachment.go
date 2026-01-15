package troublereports

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetAttachment(c echo.Context) *echo.HTTPError {
	// This would normally serve an attachment, but for now return not implemented
	return echo.NewHTTPError(http.StatusNotImplemented, "Attachment serving not implemented")
}
