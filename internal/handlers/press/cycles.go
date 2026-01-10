package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetCycles(c echo.Context) *echo.HTTPError {
	user, merr := urlb.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	pressNumber, merr := urlb.ParseParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	cycles, merr := db.ListCyclesByPressNumber(shared.PressNumber(pressNumber))
	if merr != nil {
		return merr.Echo()
	}

	toolsMap := make(map[shared.EntityID]*shared.Tool)
	tools, merr := db.ListTools()
	if merr != nil {
		return merr.Echo()
	}
	for _, t := range tools {
		toolsMap[t.ID] = t
	}

	t := templates.Cycles(templates.CyclesProps{
		Cycles: cycles,
		Tools:  toolsMap,
		User:   user,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "Cycles")
	}

	return nil
}
