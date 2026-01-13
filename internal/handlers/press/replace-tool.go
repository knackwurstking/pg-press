package press

import (
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func ReplaceTool(c echo.Context) *echo.HTTPError {
	// Get press number from param
	var pressNumber shared.PressNumber
	if p, merr := utils.GetParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		pressNumber = shared.PressNumber(p)
	}

	// Get position from query
	var position shared.Slot
	if p, merr := utils.GetQueryInt(c, "position"); merr != nil {
		return merr.Echo()
	} else {
		position = shared.Slot(p)
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

	if press, merr := db.GetPress(pressNumber); merr != nil {
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

	utils.SetHXTrigger(c, "reload-active-tools")

	return nil
}
