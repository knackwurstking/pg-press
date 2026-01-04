package tools

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/labstack/echo/v4"
)

// TODO: Fix all other stuff first
func AdminOverlappingTools(c echo.Context) *echo.HTTPError {
	//overlappingTools, merr := h.registry.PressCycles.GetOverlappingTools()
	//if merr != nil {
	//	return merr.Echo()
	//}

	t := templates.AdminToolsSectionContent()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "AdminToolsSectionContent")
	}

	return nil
}
