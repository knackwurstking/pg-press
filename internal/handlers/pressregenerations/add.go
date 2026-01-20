package pressregenerations

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func Add(c echo.Context) *echo.HTTPError {
	press, merr := getParamPress(c)
	if merr != nil {
		return merr.Echo()
	}

	data, merr := ParseForm(c, press)
	if merr != nil {
		return merr.Echo()
	}

	if _, merr = db.AddPressRegeneration(data); merr != nil {
		return merr.Echo()
	}

	utils.SetHXRedirect(c, urlb.Press(press))

	return nil
}
