package pressregenerations

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func getParamPress(c echo.Context) (shared.PressNumber, *errors.HTTPError) {
	pressNum, merr := utils.GetParamInt8(c, "press")
	if merr != nil {
		return -1, merr
	}

	press := shared.PressNumber(pressNum)
	if !press.IsValid() {
		return -1, merr
	}

	return press, nil
}
