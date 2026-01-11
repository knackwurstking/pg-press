package tools

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

// MarkAsDead marks a tool as dead
func MarkAsDead(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)
	merr = db.MarkToolAsDead(toolID)
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXRedirect(c, urlb.Tool(toolID))

	return nil
}
