package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetPressMetalSheets(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetParamInt64(c, "press")
	if merr != nil {
		return merr.Echo()
	}
	pressID := shared.EntityID(id)

	u, merr := db.GetPressUtilization(pressID)
	if merr != nil {
		return merr.WrapEcho("get press utilizations for press %d (ID: %d)", u.PressNumber, u.PressID)
	}

	var ums []*shared.UpperMetalSheet
	if u.SlotUpper != nil {
		var merr *errors.HTTPError
		if ums, merr = db.ListUpperMetalSheetsByTool(u.SlotUpper.ID); merr != nil {
			return merr.Echo()
		}
	}

	var lms []*shared.LowerMetalSheet
	if u.SlotLower != nil {
		var merr *errors.HTTPError
		if lms, merr = db.ListLowerMetalSheetsByTool(u.SlotLower.ID); merr != nil {
			return merr.Echo()
		} else {
			i := 0
			for _, lm := range lms {
				if lm.Identifier == u.PressType {
					lms[i] = lm
					i++
				}
			}
			lms = lms[:i]
		}
	}

	t := templates.MetalSheets(templates.MetalSheetsProps{
		Utilization:      u,
		UpperMetalSheets: ums,
		LowerMetalSheets: lms,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "MetalSheetsSection")
	}
	return nil
}
