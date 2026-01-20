package pressregenerations

import (
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/pressregenerations/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	press, merr := getParamPress(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.PageProps{
		Press: press,
	}
	if err := templates.Page(t).Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Press Regenration Page")
	}

	return nil
}

func PostPage(c echo.Context) *echo.HTTPError {
	press, merr := getParamPress(c)
	if merr != nil {
		return merr.Echo()
	}

	data, merr := parseForm(c, press)
	if merr != nil {
		return merr.Echo()
	}

	if merr = db.AddPressRegeneration(data); merr != nil {
		return merr.Echo()
	}

	utils.SetHXRedirect(c, urlb.Press(press))

	return nil
}

func parseForm(c echo.Context, press shared.PressNumber) (*shared.PressRegeneration, *errors.HTTPError) {
	startStr := c.FormValue("started")
	if startStr == "" {
		return nil, errors.NewValidationError("started date is required").HTTPError()
	}

	stopStr := c.FormValue("completed")
	if stopStr == "" {
		return nil, errors.NewValidationError("completed date is required").HTTPError()
	}

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid started date format: %s", err.Error()).HTTPError()
	}

	stop, err := time.Parse("2006-01-02", stopStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid completed date format: %s", err.Error()).HTTPError()
	}

	data := &shared.PressRegeneration{
		PressNumber: press,
		Start:       shared.UnixMilli(start.UnixMilli()),
		Stop:        shared.UnixMilli(stop.UnixMilli()),
	}

	if err := data.Validate(); err != nil {
		return nil, errors.NewValidationError(err.Error()).HTTPError()
	}

	return data, nil
}
