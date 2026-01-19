package help

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/help/templates"

	"github.com/labstack/echo/v4"
)

func GetMarkdownPage(c echo.Context) *echo.HTTPError {
	t := templates.MarkdownPage()
	err := t.Render(c.Request().Context(), c.Response().Writer)
	if err != nil {
		return errors.NewRenderError(err, "Markdown Page")
	}

	return nil
}
