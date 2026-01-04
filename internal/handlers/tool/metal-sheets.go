package tool

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func GetToolMetalSheets(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.WrapEcho("could not get tool by ID")
	}

	var t templ.Component

	if !tool.IsLowerTool() && !tool.IsUpperTool() {
		return echo.NewHTTPError(http.StatusBadRequest, "Tool is not supported for metal sheets")
	}

	if tool.IsLowerTool() {
		metalSheets, merr := db.ListLowerMetalSheetsByTool(tool.ID)
		if merr != nil {
			return merr.Echo()
		}

		t = templates.LowerMetalSheets(metalSheets, tool, user)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "LowerMetalSheets")
		}

		return nil
	}

	metalSheets, merr := db.ListUpperMetalSheetsByTool(tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	t = templates.UppperMetalSheets(metalSheets, tool, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "LowerMetalSheets")
	}

	return nil
}
