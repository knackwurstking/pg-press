package tools

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

// MarkAsDead marks a tool as dead
func MarkAsDead(c echo.Context) *echo.HTTPError {
	id, merr := urlb.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)
	merr = db.MarkToolAsDead(toolID)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXRedirect(c, urlb.UrlTool(toolID, 0, 0).Page)

	return nil
}
