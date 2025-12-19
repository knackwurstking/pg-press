package tool

import (
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func HTMXDeleteToolCycle(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	cycleID := shared.EntityID(id)

	merr = DB.Press.Cycles.Delete(cycleID)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, "reload-cycles")

	return nil
}
