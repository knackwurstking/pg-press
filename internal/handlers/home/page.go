package home

import (
	"github.com/knackwurstking/pg-press/internal/errors"

	"github.com/labstack/echo/v4"
)

func GetHomePage(c echo.Context) *echo.HTTPError {
	t := HomePage()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Home Page")
	}
	return nil
}
