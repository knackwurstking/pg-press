package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := utils.GetParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}
	press, merr := db.GetPress(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(templates.PageProps{
		Press: press,
		User:  user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Page")
	}
	return nil
}
