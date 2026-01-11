package press

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetPressMetalSheets(c echo.Context) *echo.HTTPError {
	var pn shared.PressNumber
	if p, merr := utils.GetParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		pn = shared.PressNumber(p)
	}

	var u *shared.PressUtilization
	if pu, merr := db.GetPressUtilizations([]shared.PressNumber{pn}...); merr != nil {
		return merr.WrapEcho("get press utilizations for press %d", pn)
	} else {
		var ok bool
		if u, ok = pu[pn]; !ok {
			return echo.NewHTTPError(http.StatusNotFound, "no active tools for press %d", pn)
		}
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
				if lm.Identifier == u.Type {
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
