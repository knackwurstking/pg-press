package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

// GetCyclesSectionContent handles the request to get the cycles section content
func GetCyclesSectionContent(c echo.Context) *echo.HTTPError {
	return renderCyclesSectionContent(c)
}

// renderCyclesSection renders the cycles section for a tool
func renderCyclesSection(c echo.Context, tool *shared.Tool) *echo.HTTPError {
	// Render out-of-band swap for cycles section to trigger reload
	t := templates.CyclesSection(true, tool.ID)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "CyclesSection")
	}
	return nil
}

// renderCyclesSectionContent renders the content for the cycles section
func renderCyclesSectionContent(c echo.Context) *echo.HTTPError {
	// Get tool from URL param "id"
	id, merr := utils.GetParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	toolCycles, herr := db.ListToolCycles(tool.ID)
	if herr != nil {
		return herr.Echo()
	}

	presses, herr := db.ListPress()
	if herr != nil {
		return herr.Echo()
	}

	// Create press map for easy lookup
	pressMap := make(map[shared.EntityID]*shared.Press)
	for _, p := range presses {
		pressMap[p.ID] = p
	}

	// Get active press number for this tool, -1 if none
	activePress, merr := db.GetPressForTool(tool.ID)
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	// Get bindable cassettes for this tool, if it is a tool and not a cassette
	bindableCassettes, eerr := listBindableCassettes(tool)
	if eerr != nil {
		return eerr
	}

	// Get regenerations for this tool
	regenerations, herr := db.ListToolRegenerationsByTool(tool.ID)
	if herr != nil {
		return herr.Echo()
	}

	// Get user from context
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.CyclesSectionContent(&templates.CyclesSectionContentProps{
		Tool:                tool,
		ToolCycles:          toolCycles,
		PressMap:            pressMap,
		ActivePress:         activePress,
		CassettesForBinding: bindableCassettes,
		Regenerations:       regenerations,
		User:                user,
	})

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cycles")
	}

	return nil
}
