package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetActiveTools(c echo.Context) *echo.HTTPError {
	pressNumber, merr := shared.ParseParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := db.GetPressUtilizations([]shared.PressNumber{shared.PressNumber(pressNumber)}...)
	if merr != nil {
		return merr.WrapEcho("get press utilizations for press %d", pressNumber)
	}

	t := templates.ActiveToolsSection(tools[shared.PressNumber(pressNumber)])
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ActiveToolsSection")
	}

	return nil
}
