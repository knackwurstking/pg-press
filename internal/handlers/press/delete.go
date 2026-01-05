package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/labstack/echo/v4"
)

func DeletePress(c echo.Context) *echo.HTTPError {
	var pressNumber shared.PressNumber
	if pn, merr := shared.ParseParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		pressNumber = shared.PressNumber(pn)
	}

	if merr := db.DeletePress(pressNumber); merr != nil {
		return merr.WrapEcho("delete press %d", pressNumber)
	}

	urlb.SetHXRedirect(c, urlb.UrlPress(pressNumber).Page)

	return nil
}
