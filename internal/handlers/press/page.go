package press

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	user, merr := urlb.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	pressNumber, merr := urlb.ParseParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(templates.PageProps{
		PressNumber: shared.PressNumber(pressNumber),
		User:        user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Page")
	}
	return nil
}
