package tools

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func PressSection(c echo.Context) *echo.HTTPError {
	return renderPressSection(c)
}

func renderPressSection(c echo.Context) *echo.HTTPError {
	pressUtilizations, merr := db.GetPressUtilizations(shared.AllPressNumbers...)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.SectionPress(templates.SectionPressProps{
		PressUtilizations: pressUtilizations,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionPress")
	}

	return nil
}
