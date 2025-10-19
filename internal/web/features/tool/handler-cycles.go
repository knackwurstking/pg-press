package tool

import (
	"fmt"
	"sort"

	"github.com/knackwurstking/pgpress/internal/web/features/tool/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/labstack/echo/v4"
)

func (h *Handler) HTMXGetCycles(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	var filteredCycles []*models.Cycle
	{
		cycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
		if err != nil {
			return h.HandleError(c, err, "failed to get press cycles")
		}

		filteredCycles = models.FilterCyclesByToolPosition(
			tool.Position, cycles...)
	}

	var resolvedRegenerations []*models.ResolvedRegeneration
	{ // Get (resolved) regeneration history for this tool
		regenerations, err := h.DB.ToolRegenerations.GetRegenerationHistory(toolID)
		if err != nil {
			h.Log.Error("Failed to get regenerations for tool %d: %v", toolID, err)
		}

		// Resolve regenerations
		for _, r := range regenerations {
			rr, err := h.resolveRegeneration(c, r)
			if err != nil {
				return err
			}

			resolvedRegenerations = append(resolvedRegenerations, rr)
		}
	}

	totalCycles := h.getTotalCycles(
		toolID,
		filteredCycles...,
	)

	// Only get tools for binding if the tool has no binding
	toolsForBinding, err := h.getToolsForBinding(c, tool)
	if err != nil {
		return err
	}

	// Render the template
	cyclesSection := templates.Cycles(&templates.CyclesProps{
		User:            user,
		Tool:            tool,
		ToolsForBinding: toolsForBinding,
		TotalCycles:     totalCycles,
		Cycles:          filteredCycles,
		Regenerations:   resolvedRegenerations,
	})

	if err := cyclesSection.Render(
		c.Request().Context(),
		c.Response(),
	); err != nil {
		h.HandleError(c, err, "failed to render tool cycles")
	}

	return nil
}

func (h *Handler) HTMXGetToolTotalCycles(c echo.Context) error {
	// Get tool and position parameters
	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	// Get cycles for this specific tool
	toolCycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get press cycles")
	}

	// Filter cycles by position
	filteredCycles := models.FilterCyclesByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	return components.TotalCycles(
		totalCycles, h.ParseBoolQuery(c, "input"),
	).Render(c.Request().Context(), c.Response())
}

func (h *Handler) HTMXGetToolCycleEditDialog(c echo.Context) error {
	props := &templates.DialogEditCycleProps{}

	// Check if we're in tool change mode
	toolChangeMode := h.ParseBoolQuery(c, "tool_change_mode")

	if c.QueryParam("id") != "" {
		cycleID, err := h.ParseInt64Query(c, "id")
		if err != nil {
			return h.RenderBadRequest(c, "failed to parse cycle ID: "+err.Error())
		}
		props.CycleID = cycleID

		// Get cycle data from the database
		cycle, err := h.DB.PressCycles.Get(cycleID)
		if err != nil {
			return h.HandleError(c, err, "failed to load cycle data")
		}
		props.InputPressNumber = &(cycle.PressNumber)
		props.InputTotalCycles = cycle.TotalCycles
		props.OriginalDate = &cycle.Date

		// Set the cycles (original) tool to props
		if props.Tool, err = h.DB.Tools.Get(cycle.ToolID); err != nil {
			return h.HandleError(c, err, "failed to load tool data")
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			props.AllowToolChange = true

			// Get all tools
			allTools, err := h.DB.Tools.List()
			if err != nil {
				return h.HandleError(c, err, "failed to load available tools")
			}

			// Filter out tools not matching the original tools position
			for _, t := range allTools {
				if t.Position == props.Tool.Position {
					props.AvailableTools = append(props.AvailableTools, t)
				}
			}

			// Sort tools alphabetically by code
			sort.Slice(props.AvailableTools, func(i, j int) bool {
				return props.AvailableTools[i].String() < props.AvailableTools[j].String()
			})
		}
	} else if c.QueryParam("tool_id") != "" {
		toolID, err := h.ParseInt64Query(c, "tool_id")
		if err != nil {
			return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
		}

		if props.Tool, err = h.DB.Tools.Get(toolID); err != nil {
			return h.HandleError(c, err, "failed to load tool data")
		}
	} else {
		return h.RenderBadRequest(c, "missing tool or cycle ID")
	}

	cycleEditDialog := templates.DialogEditCycle(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render cycle edit dialog")
	}

	return nil
}

func (h *Handler) HTMXPostToolCycleEditDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	tool, err := h.getToolFromQuery(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to get tool from query: "+err.Error())
	}

	// Parse form data
	form, err := h.getCycleFormData(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse form data: "+err.Error())
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return h.RenderBadRequest(c, "press_number must be a valid integer")
	}

	pressCycle := models.NewCycle(
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
	)

	pressCycle.Date = form.Date

	cycleID, err := h.DB.PressCycles.Add(pressCycle, user)
	if err != nil {
		return h.HandleError(c, err, "failed to add cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		h.Log.Info("Starting regeneration for tool %d", tool.ID)
		_, err := h.DB.ToolRegenerations.AddToolRegeneration(tool.ID, cycleID, "", user)
		if err != nil {
			h.Log.Error("Failed to start regeneration for tool %d: %v",
				tool.ID, err)
		}
	}

	// Create feed entry
	title := fmt.Sprintf("Neuer Zyklus hinzugefügt für %s", tool.String())
	content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
		*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))
	if form.Regenerating {
		content += "\nRegenerierung gestartet"
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for cycle creation: %v", err)
	}

	return nil
}

func (h *Handler) HTMXPutToolCycleEditDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	cycleID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID from query: "+err.Error())
	}

	cycle, err := h.DB.PressCycles.Get(cycleID)
	if err != nil {
		return h.HandleError(c, err, "failed to get cycle")
	}

	// Get original tool
	originalTool, err := h.DB.Tools.Get(cycle.ToolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get original tool")
	}

	form, err := h.getCycleFormData(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to get cycle form data from query: "+err.Error())
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return h.RenderBadRequest(c, "press_number must be a valid integer")
	}

	// Determine which tool to use for the cycle
	var tool *models.Tool
	if form.ToolID != nil {
		// Tool change requested - get the new tool
		newTool, err := h.DB.Tools.Get(*form.ToolID)
		if err != nil {
			return h.HandleError(c, err, "failed to get new tool")
		}
		tool = newTool
	} else {
		// No tool change - use original tool
		tool = originalTool
	}

	// Update the cycle
	pressCycle := models.NewPressCycleWithID(
		cycle.ID,
		*form.PressNumber,
		tool.ID, tool.Position, form.TotalCycles,
		user.TelegramID,
		form.Date,
	)

	if err := h.DB.PressCycles.Update(pressCycle, user); err != nil {
		return h.HandleError(c, err, "failed to update press cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		h.Log.Info("Starting regeneration for tool %d", tool.ID)
		_, err := h.DB.ToolRegenerations.AddToolRegeneration(
			tool.ID, cycleID, "", user)
		if err != nil {
			h.Log.Error("Failed to start regeneration for tool %d: %v",
				tool.ID, err)
		}

		h.Log.Info("Stopping regeneration for tool %d", tool.ID)
		err = h.DB.ToolRegenerations.StopToolRegeneration(tool.ID, user)
		if err != nil {
			h.Log.Error("Failed to stop regeneration for tool %d: %v",
				tool.ID, err)
		}
	}

	// Create feed entry
	var title string
	var content string

	if form.ToolID != nil {
		// Tool change occurred
		title = "Zyklus aktualisiert mit Werkzeugwechsel"
		content = fmt.Sprintf("Presse: %d\nAltes Werkzeug: %s\nNeues Werkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, originalTool.String(), tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))
	} else {
		// Regular cycle update
		title = fmt.Sprintf("Zyklus aktualisiert für %s", tool.String())
		content = fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))
	}

	if form.Regenerating {
		content += "\nRegenerierung abgeschlossen"
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for cycle update: %v", err)
	}

	return nil
}

func (h *Handler) HTMXDeleteToolCycle(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	cycleID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID query: "+err.Error())
	}

	// Get cycle data before deletion for the feed
	cycle, err := h.DB.PressCycles.Get(cycleID)
	if err != nil {
		return h.HandleError(c, err, "failed to get cycle for deletion")
	}

	tool, err := h.DB.Tools.Get(cycle.ToolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool for deletion")
	}

	// Check if there are any regenerations associated with this cycle
	hasRegenerations, err := h.DB.ToolRegenerations.HasRegenerationsForCycle(cycleID)
	if err != nil {
		return h.HandleError(c, err, "failed to check for regenerations")
	}

	if hasRegenerations {
		return h.RenderBadRequest(c, "Cannot delete cycle: it has associated regenerations. Please remove regenerations first.")
	}

	if err := h.DB.PressCycles.Delete(cycleID); err != nil {
		return h.HandleError(c, err, "failed to delete press cycle")
	}

	// Create feed entry
	title := fmt.Sprintf("Zyklus gelöscht für %s", tool.String())
	content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
		cycle.PressNumber, tool.String(), cycle.TotalCycles, cycle.Date.Format("2006-01-02 15:04:05"))

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for cycle deletion: %v", err)
	}

	return nil
}

func (h *Handler) getTotalCycles(toolID int64, cycles ...*models.Cycle) int64 {
	// Get regeneration for this tool
	var startCycleID int64
	if r, err := h.DB.ToolRegenerations.GetLastRegeneration(toolID); err == nil {
		startCycleID = r.CycleID
	}

	var totalCycles int64

	for _, cycle := range cycles {
		if cycle.ID <= startCycleID {
			continue
		}

		totalCycles += cycle.PartialCycles
	}

	return totalCycles
}
