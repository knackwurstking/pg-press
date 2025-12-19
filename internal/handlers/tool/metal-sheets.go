package tool

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func HTMXGetToolMetalSheets(c echo.Context) *echo.HTTPError {
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

	var t templ.Component

	// Fetch metal sheets for tool
	switch p := tool.GetBase().Position; p {
	case shared.SlotUpper:
		metalSheets, merr := helper.ListUpperMetalSheetsForTool(DB, tool.GetID())
		if merr != nil {
			return merr.Echo()
		}
		t = UppperMetalSheets(metalSheets, tool.(*shared.Tool), user)
	case shared.SlotLower:
		metalSheets, merr := helper.ListLowerMetalSheetsForTool(DB, tool.GetID())
		if merr != nil {
			return merr.Echo()
		}
		t = LowerMetalSheets(metalSheets, tool.(*shared.Tool), user)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Tool is not supported for metal sheets")
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "MetalSheets")
	}
	return nil
}
