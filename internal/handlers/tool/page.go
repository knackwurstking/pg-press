package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetToolPage(c echo.Context) *echo.HTTPError {
	user, merr := urlb.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := urlb.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.WrapEcho("could not get tool by ID")
	}

	t := templates.Page(&templates.PageProps{
		Tool: tool,
		User: user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tool Page")
	}

	return nil
}
