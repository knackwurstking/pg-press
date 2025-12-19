package tool

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetToolPage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := helper.GetToolByID(DB, shared.EntityID(id))
	if merr != nil {
		return merr.WrapEcho("could not get tool by ID")
	}

	t := Page(&PageProps{
		Tool: tool,
		User: user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tool Page")
	}

	return nil
}
