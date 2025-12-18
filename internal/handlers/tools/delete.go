package tools

import (
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

// Delete deletes a tool or a cassette if "is_cassette" query parameter is set to true.
func Delete(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	isCassette := shared.ParseQueryBool(c, "is_cassette")

	if isCassette {
		merr = DB.Tool.Cassette.Delete(toolID)
		if merr != nil {
			return merr.Echo()
		}
		Log.Debug("Deleted cassette with ID: %#v", toolID)
	} else {
		merr = DB.Tool.Tool.Delete(toolID)
		if merr != nil {
			return merr.Echo()
		}
		Log.Debug("Deleted tool with ID: %#v", toolID)
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}
