package press

import (
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func ReplaceTool(c echo.Context) *echo.HTTPError {
	var position shared.Slot
	if p, merr := shared.ParseQueryInt(c, "position"); merr != nil {
		return merr.Echo()
	} else {
		position = shared.Slot(p)
	}

	var toolID shared.EntityID
	if id, merr := shared.ParseQueryInt64(c, "tool_id"); merr != nil {
		return merr.Echo()
	} else {
		toolID = shared.EntityID(id)
	}

	// Get form values: "tool_id" - this is the new tool to use for position
	var newToolID shared.EntityID
	if vToolID := c.FormValue("tool_id"); vToolID == "" {
		return echo.NewHTTPError(http.StatusBadRequest,
			"missing form value: tool_id")
	} else {
		i, err := strconv.ParseInt(vToolID, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				"invalid form value: tool_id")
		}
		newToolID = shared.EntityID(i)
	}

	if press, merr := db.GetPress(toolID); merr != nil {
		return merr.Echo()
	} else {
		switch position {
		case shared.SlotUpper:
			press.SlotUp = newToolID
		case shared.SlotLower:
			press.SlotDown = newToolID
		default:
			return echo.NewHTTPError(http.StatusBadRequest,
				"invalid position value: %s", position)
		}
		if merr = db.UpdatePress(press); merr != nil {
			return merr.Echo()
		}
	}

	urlb.SetHXTrigger(c, "reload-active-tools")

	return nil
}
