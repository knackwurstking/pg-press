package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func DeletePress(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetParamInt64(c, "press")
	if merr != nil {
		return merr.Echo()
	}
	pressID := shared.EntityID(id)

	if merr := db.DeletePress(pressID); merr != nil {
		return merr.Echo()
	}

	utils.SetHXRedirect(c, urlb.Tools())

	return nil
}
