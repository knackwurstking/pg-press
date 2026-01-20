package pressregenerations

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/pressregenerations/templates"
	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	press, merr := getParamPress(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.PageProps{
		Press: press,
	}
	if err := templates.Page(t).Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Press Regenration Page")
	}

	return nil
}
