package pressregenerations

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func Delete(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id") // PressRegenerationID
	if merr != nil {
		return merr.Echo()
	}

	merr = db.DeletePressRegeneration(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-press-regenerations")

	return nil
}
