package tool

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetCyclesSectionContent(c echo.Context) *echo.HTTPError {
	return renderCyclesSectionContent(c)
}

func renderCyclesSection(c echo.Context, tool shared.ModelTool) *echo.HTTPError {
	// Render out-of-band swap for cycles section to trigger reload
	t := CyclesSection(true, tool.GetID(), !tool.IsCassette())
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

	tool, merr := helper.GetToolByID(DB, toolID)
	if merr != nil {
		return merr.WrapEcho("could not get tool by ID")
	}

	// Get cycles for this specific tool
	toolCycles, merr := helper.ListCyclesForTool(DB, toolID)
	if merr != nil {
		return merr.WrapEcho("could not list cycles for tool")
	}

	// Get active press number for this tool, -1 if none
	activePressNumber := helper.GetPressNumberForTool(DB, toolID)

	// Get bindable cassettes for this tool, if it is a tool and not a cassette
	var cassettesForBinding []*shared.Cassette
	if !tool.IsCassette() {
		cassettesForBinding, merr = helper.ListAvailableCassettesForBinding(DB, toolID)
		if merr != nil {
			return merr.WrapEcho("could not list available cassettes for binding")
		}
	}

	// Get regenerations for this tool
	regenerations, merr := helper.GetRegenerationsForTool(DB, toolID)
	if merr != nil {
		return merr.WrapEcho("could not get regenerations for tool")
	}

	// Get user from context
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.WrapEcho("could not get user from context")
	}

	t := CyclesSectionContent(CyclesSectionContentProps{
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
