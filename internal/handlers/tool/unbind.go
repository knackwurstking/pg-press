package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func ToolUnBinding(c echo.Context) *echo.HTTPError {
	id, merr := urlb.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)
	merr = db.UnbindTool(toolID)
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(toolID)
	if merr != nil {
		return merr.Echo()
	}

	return renderBindingSection(c, tool)
}
