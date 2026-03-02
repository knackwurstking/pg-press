package tools

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

// Delete deletes a tool
func Delete(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	merr = db.DeleteTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}
	log.Debug("Deleted tool with ID: %d", id)

	utils.SetHXRedirect(c, urlb.Tools())

	return nil
}
