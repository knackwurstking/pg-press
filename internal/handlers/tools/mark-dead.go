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
			return merr.Echo()
		}

		if cassette.IsDead {
			return nil
		}
		cassette.IsDead = true

		merr = DB.Tool.Cassette.Update(cassette)
		if merr != nil {
			return merr.Echo()
		}
	} else {
		tool, merr := DB.Tool.Tool.GetByID(toolID)
		if merr != nil {
			return merr.Echo()
		}

		if tool.IsDead {
			return nil
		}
		tool.IsDead = true

		merr = DB.Tool.Tool.Update(tool)
		if merr != nil {
			return merr.Echo()
		}
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}
