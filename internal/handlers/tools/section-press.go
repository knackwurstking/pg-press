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
	pressUtilizations, herr := db.GetPressUtilizations()
	if herr != nil && !herr.IsNotFoundError() {
		return herr.Echo()
	}

	var pressUtilizationsOrder []shared.EntityID
	for e := range pressUtilizations {
		pressUtilizationsOrder = append(pressUtilizationsOrder, e)
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
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Section Press")
	}

	return nil
}
