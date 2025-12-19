package tools

import (
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

// MarkAsDead marks a tools, or cassette if "is_cassette" query parameter is set to true, as dead.
func MarkAsDead(c echo.Context) *echo.HTTPError {
	isCassette := shared.ParseQueryBool(c, "is_cassette")

	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	var tool shared.ModelTool
	if isCassette {
		tool, merr = DB.Tool.Cassette.GetByID(toolID)
		if merr != nil {
			return merr.WrapEcho("failed to get cassette by ID")
		}
	} else {
		tool, merr = DB.Tool.Tool.GetByID(toolID)
		if merr != nil {
			return merr.WrapEcho("failed to get tool by ID")
		}
	}

	if tool.GetBase().IsDead {
		return nil
	}
	tool.GetBase().IsDead = true

	if isCassette {
		merr = DB.Tool.Cassette.Update(tool.(*shared.Cassette))
		if merr != nil {
			return merr.WrapEcho("failed to update cassette")
		}
	} else {
		merr = DB.Tool.Tool.Update(tool.(*shared.Tool))
		if merr != nil {
			return merr.WrapEcho("failed to update tool")
		}
	}

	urlb.SetHXRedirect(c, urlb.UrlTool(tool.GetID(), 0, 0).Page)
	return nil
}
