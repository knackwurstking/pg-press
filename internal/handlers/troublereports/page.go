package troublereports

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/troublereports/templates"
	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	t := templates.Page()
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Page")
	}

	return nil
}
