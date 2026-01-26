package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetCyclesSectionContent(c echo.Context) *echo.HTTPError {
	return renderCyclesSectionContent(c)
}

func renderCyclesSection(c echo.Context, tool *shared.Tool) *echo.HTTPError {
	// Render out-of-band swap for cycles section to trigger reload
	t := templates.CyclesSection(true, tool.ID)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "CyclesSection")
	}
	return nil
}

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

	// Get cycles for this specific tool
	toolCycles, merr := db.ListToolCycles(tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	// Get active press number for this tool, -1 if none
	activePressNumber, merr := db.GetPressNumberForTool(tool.ID)
	if merr != nil && !merr.IsNotFoundError() {
		return merr.Echo()
	}

	// Get bindable cassettes for this tool, if it is a tool and not a cassette
	bindableCassettes, eerr := listBindableCassettes(tool)
	if eerr != nil {
		return eerr
	}

	// Get regenerations for this tool
	regenerations, merr := db.ListToolRegenerationsByTool(tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	// Get user from context
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.CyclesSectionContent(templates.CyclesSectionContentProps{
		Tool:                tool,
		ToolCycles:          toolCycles,
		ActivePressNumber:   activePressNumber,
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
