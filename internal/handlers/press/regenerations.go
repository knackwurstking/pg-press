package press

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/labstack/echo/v4"
)

func GetRegenerations(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get press regenerations from service
	regenerations, merr := h.registry.PressRegenerations.GetRegenerationHistory(press)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.RegenerationsContent(regenerations, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "RegenerationsContent")
	}

	return nil
}
