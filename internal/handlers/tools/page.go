package tools

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetToolsPage(c echo.Context) *echo.HTTPError {
	user, merr := urlb.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(templates.PageProps{User: user})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tools Page")
	}

	return nil
}
