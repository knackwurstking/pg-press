package tools

import (
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

// MarkAsDead marks a tools, or cassette if "is_cassette" query parameter is set to true, as dead.
func MarkAsDead(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	if shared.ParseQueryBool(c, "is_cassette") {
		cassette, merr := DB.Tool.Cassette.GetByID(toolID)
		if merr != nil {
			return merr.WrapEcho("failed to get cassette by ID")
		}

		if cassette.IsDead {
			return nil
		}
		cassette.IsDead = true

		merr = DB.Tool.Cassette.Update(cassette)
		if merr != nil {
			return merr.WrapEcho("failed to update cassette")
		}

		urlb.SetHXRedirect(c, urlb.UrlTool(cassette.ID, 0, 0).Page)
		return nil
	}

	tool, merr := DB.Tool.Tool.GetByID(toolID)
	if merr != nil {
		return merr.WrapEcho("failed to get tool by ID")
	}

	if tool.IsDead {
		return nil
	}
	tool.IsDead = true

	merr = DB.Tool.Tool.Update(tool)
	if merr != nil {
		return merr.WrapEcho("failed to update tool")
	}

	urlb.SetHXRedirect(c, urlb.UrlTool(tool.ID, 0, 0).Page)

	return nil
}
