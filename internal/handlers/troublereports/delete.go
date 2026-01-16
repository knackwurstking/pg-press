package troublereports

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func DeleteTroubleReport(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	trID := shared.EntityID(id)

	if merr = db.DeleteTroubleReport(trID); merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-trouble-reports")

	return nil
}
