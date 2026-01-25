package tools

import (
	"slices"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func PressSection(c echo.Context) *echo.HTTPError {
	return renderPressSection(c)
}

func renderPressSection(c echo.Context) *echo.HTTPError {
	pressUtilizations, merr := db.GetPressUtilizations(shared.AllPressNumbers...)
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	var pressUtilizationsOrder []shared.PressNumber
	for p, _ := range pressUtilizations {
		pressUtilizationsOrder = append(pressUtilizationsOrder, p)
	}
	slices.Sort(pressUtilizationsOrder)

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr.Echo()
	}

	t := templates.SectionPress(templates.SectionPressProps{
		PressUtilizations:      pressUtilizations,
		PressUtilizationsOrder: pressUtilizationsOrder,
		User:                   user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Section Press")
	}

	return nil
}
