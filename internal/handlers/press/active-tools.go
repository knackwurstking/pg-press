package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetActiveTools(c echo.Context) *echo.HTTPError {
	pressNumber, merr := utils.GetParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	pu, merr := db.GetPressUtilizations([]shared.PressNumber{shared.PressNumber(pressNumber)}...)
	if merr != nil {
		return merr.WrapEcho("get press utilizations for press %d", pressNumber)
	}

	toolsForSelection := make(map[shared.Slot][]*shared.Tool)
	toolsForSelection[shared.SlotUpper] = []*shared.Tool{}
	toolsForSelection[shared.SlotLower] = []*shared.Tool{}
	tools, merr := db.ListTools()
	if merr != nil {
		return merr.WrapEcho("list tools for active tools selection")
	}
	for _, tool := range tools {
		switch tool.Position {
		case shared.SlotUpper, shared.SlotLower:
			toolsForSelection[tool.Position] = append(toolsForSelection[tool.Position], tool)
		}
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr.Echo()
	}

	t := templates.ActiveToolsSection(pu[shared.PressNumber(pressNumber)], toolsForSelection, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Active Tools Section")
	}

	return nil
}
