package troublereports

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/handlers/troublereports/templates"
	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	t := templates.Page()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to render page")
	}

	return nil
}
