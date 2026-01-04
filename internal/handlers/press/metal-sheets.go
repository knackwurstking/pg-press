package press

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetPressMetalSheets(c echo.Context) *echo.HTTPError {
	var pn shared.PressNumber
	if p, merr := shared.ParseParamInt8(c, "press"); merr != nil {
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
	var lms []*shared.LowerMetalSheet
	var merr *errors.MasterError
	if ums, merr = db.ListUpperMetalSheetsByTool(u.SlotUpper.ID); merr != nil {
		return merr.Echo()
	}

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

	t := MetalSheets(MetalSheetsProps{
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
