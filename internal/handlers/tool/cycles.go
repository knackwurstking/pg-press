package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetCyclesSectionContent(c echo.Context) *echo.HTTPError {
	return renderCyclesSectionContent(c)
}

func renderCyclesSection(c echo.Context, tool *shared.Tool) *echo.HTTPError {
	// Render out-of-band swap for cycles section to trigger reload
	t := templates.CyclesSection(true, tool.ID, !tool.IsCassette())
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "CyclesSection")
	}
	return nil
}

func renderCyclesSectionContent(c echo.Context) *echo.HTTPError {
	// Get tool from URL param "id"
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.WrapEcho("could not parse tool ID from URL parameter")
	}
	toolID := shared.EntityID(id)

	tool, merr := db.GetTool(toolID)
	if merr != nil {
		return merr.WrapEcho("could not get tool by ID")
	}

	// Get cycles for this specific tool
	toolCycles, merr := db.ListToolCycles(toolID)
	if merr != nil {
		return merr.WrapEcho("could not list cycles for tool")
	}

	// Get active press number for this tool, -1 if none
	activePressNumber, merr := db.GetPressNumberForTool(toolID)
	if merr != nil && !merr.IsNotFoundError() {
		return merr.WrapEcho("could not get active press number for tool")
	}

	// Get bindable cassettes for this tool, if it is a tool and not a cassette
	var cassettesForBinding []*shared.Tool
	if !tool.IsCassette() {
		cassettesForBinding, merr = db.ListBindableCassettes(toolID)
		if merr != nil {
			return merr.WrapEcho("could not list available cassettes for binding")
		}
	}

	// Get regenerations for this tool
	regenerations, merr := db.ListToolRegenerationsByTool(toolID)
	if merr != nil {
		return merr.WrapEcho("could not get regenerations for tool")
	}

	// Get user from context
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.WrapEcho("could not get user from context")
	}

	t := templates.CyclesSectionContent(templates.CyclesSectionContentProps{
		Tool:                tool,
		ToolCycles:          toolCycles,
		ActivePressNumber:   activePressNumber,
		CassettesForBinding: cassettesForBinding,
		Regenerations:       regenerations,
		User:                user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cycles")
	}

	return nil
}
