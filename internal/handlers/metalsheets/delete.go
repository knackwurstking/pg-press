package metalsheets

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func DeleteMetalSheet(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	// Get the metal sheet before deletion to determine if it's upper or lower
	_, merr = db.GetUpperMetalSheet(shared.EntityID(id))
	if merr != nil {
		// If not found as upper, try lower
		_, merr := db.GetLowerMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
		
		// Delete lower metal sheet
		merr = db.DeleteLowerMetalSheet(shared.EntityID(id))
		if merr != nil {
			return merr.Echo()
		}
		
		// Trigger reload of metal sheets sections
		utils.SetHXTrigger(c, "reload-metal-sheets")
		return nil
	}
	
	// Delete upper metal sheet
	merr = db.DeleteUpperMetalSheet(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}
	
	// Trigger reload of metal sheets sections
	utils.SetHXTrigger(c, "reload-metal-sheets")
	return nil
}
