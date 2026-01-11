package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetRegenerations(c echo.Context) *echo.HTTPError {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	var pressNumber shared.PressNumber
	if press, merr := utils.GetParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		pressNumber = shared.PressNumber(press)
	}

	// Get press regenerations from service
	regenerations, merr := db.ListPressRegenerationsByPress(pressNumber)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Regenerations(templates.RegenerationsProps{
		PressRegenerations: regenerations,
		User:               user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Regenerations")
	}

	return nil
}
