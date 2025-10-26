package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/env"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type ToolCycleEditDialogFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *models.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
	ToolID       *models.ToolID // ToolID form field name "tool_id" (for tool change mode)
}

type ToolRegenerationEditDialogFormData struct {
	Reason string
}

type Tool struct {
	*Base
}

func NewTool(db *services.Registry) *Tool {
	return &Tool{
		Base: NewBase(db, logger.NewComponentLogger("Tool")),
	}
}

func (h *Tool) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Main Page
		utils.NewEchoRoute(http.MethodGet, "/tools/tool/:id", h.GetToolPage),

		// Regnerations Table
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/tool/:id/edit-regeneration", h.HTMXGetEditRegeneration),
		utils.NewEchoRoute(http.MethodPut,
			"/htmx/tools/tool/:id/edit-regeneration", h.HTMXPutEditRegeneration),
		utils.NewEchoRoute(http.MethodDelete,
			"/htmx/tools/tool/:id/delete-regeneration", h.HTMXDeleteRegeneration),

		// Tool status and regenerations management
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/status-edit", h.HTMXGetStatusEdit),
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/status-display", h.HTMXGetStatusDisplay),
		utils.NewEchoRoute(http.MethodPut,
			"/htmx/tools/status", h.HTMXUpdateToolStatus),

		// Section loading
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/notes", h.HTMXGetToolNotes),
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/metal-sheets", h.HTMXGetToolMetalSheets),

		// Cycles table rows
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/cycles", h.HTMXGetCycles),
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/total-cycles", h.HTMXGetToolTotalCycles),

		// Get, add or edit a cycles table entry
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/cycle/edit", h.HTMXGetToolCycleEditDialog),
		utils.NewEchoRoute(http.MethodPost,
			"/htmx/tools/cycle/edit", h.HTMXPostToolCycleEditDialog),
		utils.NewEchoRoute(http.MethodPut,
			"/htmx/tools/cycle/edit", h.HTMXPutToolCycleEditDialog),

		// Delete a cycle table entry
		utils.NewEchoRoute(http.MethodDelete,
			"/htmx/tools/cycle/delete", h.HTMXDeleteToolCycle),

		// Update tools binding data
		utils.NewEchoRoute(http.MethodPatch,
			"/htmx/tools/tool/:id/bind", h.HTMXPatchToolBinding),
		utils.NewEchoRoute(http.MethodPatch,
			"/htmx/tools/tool/:id/unbind", h.HTMXPatchToolUnBinding),
	})
}

func (h *Tool) GetToolPage(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	idQuery, err := ParseParamInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse id from query parameter")
	}
	id := models.ToolID(idQuery)

	h.Log.Debug("Fetching tool %d with notes", id)

	tool, err := h.Registry.Tools.Get(id)
	if err != nil {
		return HandleError(err, "failed to get tool")
	}

	page := components.PageTool(&components.PageToolProps{
		User:       user,
		ToolString: tool.String(),
		ToolID:     tool.ID,
		Position:   tool.Position,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tool page")
	}

	return nil
}

func (h *Tool) HTMXPatchToolBinding(c echo.Context) error {
	var isAdmin bool
	{ // Check for admin
		user, err := GetUserFromContext(c)
		if err != nil {
			return HandleBadRequest(err, "failed to get user from context")
		}
		isAdmin = user.IsAdmin()
	}

	// Get tool from param "/:id"
	toolIDQuery, err := ParseParamInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	var tool *models.ResolvedTool
	if t, err := h.Registry.Tools.Get(toolID); err != nil {
		return HandleError(err, "failed to get tool")
	} else {
		tool, err = services.ResolveTool(h.Registry, t)
		if err != nil {
			return HandleError(err, "failed to resolve tool")
		}
	}

	var targetID models.ToolID
	{ // Get target_id from form value
		targetIDString := c.FormValue("target_id")
		if targetIDString == "" {
			return HandleBadRequest(nil, fmt.Sprintf(
				"failed to parse target_id: %+v", targetIDString))
		}

		targetIDParsed, err := strconv.ParseInt(targetIDString, 10, 64)
		if err != nil {
			return HandleBadRequest(err, "invalid target_id")
		}
		targetID = models.ToolID(targetIDParsed)
	}

	h.Log.Info("Updating tool binding: tool %d -> target %d",
		toolID, targetID)

	{ // Make sure to check for position first (target == top && toolID == cassette)
		var (
			cassette models.ToolID
			target   models.ToolID // top position
		)

		if tool.Position == models.PositionTopCassette {
			cassette = tool.ID
			target = targetID
		} else {
			cassette = targetID // If this is not a cassette, the bind method will return an error
			target = tool.ID
		}

		// Bind tool to target, this will get an error if target already has a binding
		if err = h.Registry.Tools.Bind(cassette, target); err != nil {
			return HandleError(err, "failed to bind tool")
		}
	}

	// Update tools binding, no need to re fetch this tools data from the database
	tool.Binding = &targetID

	// Get tools for binding
	toolsForBinding, err := h.getToolsForBinding(tool.Tool)
	if err != nil {
		return err
	}

	// Render the template
	bs := components.PageTool_BindingSection(components.PageTool_BindingSectionProps{
		Tool:            tool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         isAdmin,
	})

	if err = bs.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render binding section")
	}

	return nil
}

func (h *Tool) HTMXPatchToolUnBinding(c echo.Context) error {
	var isAdmin bool
	{ // Check for admin
		user, err := GetUserFromContext(c)
		if err != nil {
			return HandleBadRequest(err, "failed to get user from context")
		}
		isAdmin = user.IsAdmin()
	}

	// Get tool from param "/:id"
	toolIDParam, err := ParseParamInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool_id")
	}
	toolID := models.ToolID(toolIDParam)

	if err := h.Registry.Tools.UnBind(toolID); err != nil {
		return HandleError(err, "failed to unbind tool")
	}

	// Get tools for rendering the template
	var tool *models.ResolvedTool
	if t, err := h.Registry.Tools.Get(toolID); err != nil {
		return HandleBadRequest(err, "failed to get tool")
	} else {
		tool, err = services.ResolveTool(h.Registry, t)
		if err != nil {
			return HandleError(err, "failed to resolve tool")
		}
	}

	// Get tools for binding
	toolsForBinding, err := h.getToolsForBinding(tool.Tool)
	if err != nil {
		return err
	}

	// Render the template
	bs := components.PageTool_BindingSection(components.PageTool_BindingSectionProps{
		Tool:            tool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         isAdmin,
	})

	if err = bs.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render binding section")
	}

	return nil
}

func (h *Tool) HTMXGetCycles(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	toolIDParam, err := ParseQueryInt64(c, "tool_id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool_id")
	}
	toolID := models.ToolID(toolIDParam)

	var tool *models.ResolvedTool
	if t, err := h.Registry.Tools.Get(toolID); err != nil {
		return HandleError(err, "failed to get tool")
	} else {
		tool, err = services.ResolveTool(h.Registry, t)
		if err != nil {
			return HandleError(err, "failed to resolve tool")
		}
	}

	var filteredCycles []*models.Cycle
	{
		cycles, err := h.Registry.PressCycles.GetPressCyclesForTool(toolID)
		if err != nil {
			return HandleError(err, "failed to get press cycles")
		}

		filteredCycles = models.FilterCyclesByToolPosition(
			tool.Position, cycles...)
	}

	var resolvedRegenerations []*models.ResolvedRegeneration
	{ // Get (resolved) regeneration history for this tool
		regenerations, err := h.Registry.ToolRegenerations.GetRegenerationHistory(toolID)
		if err != nil {
			h.Log.Error("Failed to get regenerations for tool %d: %v", toolID, err)
		}

		// Resolve regenerations
		for _, r := range regenerations {
			rr, err := services.ResolveRegeneration(h.Registry, r)
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
	toolsForBinding, err := h.getToolsForBinding(tool.Tool)
	if err != nil {
		return err
	}

	// Render the template
	cyclesSection := components.PageTool_Cycles(components.PageTool_CyclesProps{
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
		HandleError(err, "failed to render tool cycles")
	}

	return nil
}

func (h *Tool) HTMXGetToolTotalCycles(c echo.Context) error {
	// Get tool and position parameters
	toolIDQuery, err := ParseQueryInt64(c, "tool_id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool")
	}

	// Get cycles for this specific tool
	toolCycles, err := h.Registry.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return HandleError(err, "failed to get press cycles")
	}

	// Filter cycles by position
	filteredCycles := models.FilterCyclesByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	tc := components.TotalCycles(totalCycles, ParseQueryBool(c, "input"))
	return tc.Render(c.Request().Context(), c.Response())
}

func (h *Tool) HTMXGetToolCycleEditDialog(c echo.Context) error {
	props := &components.DialogEditToolCycleProps{}

	// Check if we're in tool change mode
	toolChangeMode := ParseQueryBool(c, "tool_change_mode")

	if c.QueryParam("id") != "" {
		cycleIDQuery, err := ParseQueryInt64(c, "id")
		if err != nil {
			return HandleBadRequest(err, "failed to parse cycle ID")
		}
		props.CycleID = models.CycleID(cycleIDQuery)

		// Get cycle data from the database
		cycle, err := h.Registry.PressCycles.Get(props.CycleID)
		if err != nil {
			return HandleError(err, "failed to load cycle data")
		}
		props.InputPressNumber = &(cycle.PressNumber)
		props.InputTotalCycles = cycle.TotalCycles
		props.OriginalDate = &cycle.Date

		// Set the cycles (original) tool to props
		if props.Tool, err = h.Registry.Tools.Get(cycle.ToolID); err != nil {
			return HandleError(err, "failed to load tool data")
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			props.AllowToolChange = true

			// Get all tools
			allTools, err := h.Registry.Tools.List()
			if err != nil {
				return HandleError(err, "failed to load available tools")
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
		toolIDQuery, err := ParseQueryInt64(c, "tool_id")
		if err != nil {
			return HandleBadRequest(err, "failed to parse tool ID")
		}
		toolID := models.ToolID(toolIDQuery)

		if props.Tool, err = h.Registry.Tools.Get(toolID); err != nil {
			return HandleError(err, "failed to load tool data")
		}
	} else {
		return HandleBadRequest(nil, "missing tool or cycle ID")
	}

	cycleEditDialog := components.DialogEditToolCycle(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render cycle edit dialog")
	}

	return nil
}

func (h *Tool) HTMXPostToolCycleEditDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	tool, err := h.getToolFromQuery(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get tool from query")
	}

	// Parse form data
	form, err := h.getCycleFormData(c)
	if err != nil {
		return HandleBadRequest(err, "failed to parse form data")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return HandleBadRequest(nil, "press_number must be a valid integer")
	}

	pressCycle := models.NewCycle(
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
	)

	pressCycle.Date = form.Date

	_, err = h.Registry.PressCycles.Add(pressCycle, user)
	if err != nil {
		return HandleError(err, "failed to add cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		h.Log.Info("Starting regeneration for tool %d", tool.ID)
		_, err := h.Registry.ToolRegenerations.StartToolRegeneration(tool.ID, "", user)
		if err != nil {
			h.Log.Error("Failed to start regeneration for tool %d: %v",
				tool.ID, err)
		}
	}

	{ // Create Feed
		title := fmt.Sprintf("Neuer Zyklus hinzugefügt für %s", tool.String())
		content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))

		if form.Regenerating {
			content += "\nRegenerierung gestartet"
		}

		h.createFeed(title, content, user.TelegramID)
	}

	SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Tool) HTMXPutToolCycleEditDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	cycleIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse ID from query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	cycle, err := h.Registry.PressCycles.Get(cycleID)
	if err != nil {
		return HandleError(err, "failed to get cycle")
	}

	// Get original tool
	originalTool, err := h.Registry.Tools.Get(cycle.ToolID)
	if err != nil {
		return HandleError(err, "failed to get original tool")
	}

	form, err := h.getCycleFormData(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get cycle form data from query")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return HandleBadRequest(nil, "press_number must be a valid integer")
	}

	// Determine which tool to use for the cycle
	var tool *models.Tool
	if form.ToolID != nil {
		// Tool change requested - get the new tool
		newTool, err := h.Registry.Tools.Get(*form.ToolID)
		if err != nil {
			return HandleError(err, "failed to get new tool")
		}
		tool = newTool
	} else {
		// No tool change - use original tool
		tool = originalTool
	}

	// Update the cycle
	pressCycle := models.NewCycleWithID(
		cycle.ID,
		*form.PressNumber,
		tool.ID, tool.Position, form.TotalCycles,
		user.TelegramID,
		form.Date,
	)

	if err := h.Registry.PressCycles.Update(pressCycle, user); err != nil {
		return HandleError(err, "failed to update press cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		if _, err := h.Registry.ToolRegenerations.Add(tool.ID, pressCycle.ID, "", user); err != nil {
			h.Log.Error("Failed to add tool regeneration: %v", err)
		}
	}

	{ // Create Feed
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

		h.createFeed(title, content, user.TelegramID)
	}

	SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Tool) HTMXDeleteToolCycle(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	cycleIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse ID query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	// Get cycle data before deletion for the feed
	cycle, err := h.Registry.PressCycles.Get(cycleID)
	if err != nil {
		return HandleError(err, "failed to get cycle for deletion")
	}

	tool, err := h.Registry.Tools.Get(cycle.ToolID)
	if err != nil {
		return HandleError(err, "failed to get tool for deletion")
	}

	// Check if there are any regenerations associated with this cycle
	hasRegenerations, err := h.Registry.ToolRegenerations.HasRegenerationsForCycle(cycleID)
	if err != nil {
		return HandleError(err, "failed to check for regenerations")
	}

	if hasRegenerations {
		return HandleBadRequest(nil, "Cannot delete cycle: it has associated regenerations. Please remove regenerations first.")
	}

	if err := h.Registry.PressCycles.Delete(cycleID); err != nil {
		return HandleError(err, "failed to delete press cycle")
	}

	{ // Create Feed
		title := fmt.Sprintf("Zyklus gelöscht für %s", tool.String())
		content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			cycle.PressNumber, tool.String(), cycle.TotalCycles, cycle.Date.Format("2006-01-02 15:04:05"))

		h.createFeed(title, content, user.TelegramID)
	}

	SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Tool) HTMXGetToolMetalSheets(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := ParseQueryInt64(c, "tool_id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	h.Log.Debug("Fetching metal sheets for tool %d", toolID)

	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool")
	}

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.Registry.MetalSheets.GetByToolID(toolID)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		h.Log.Error("Failed to fetch metal sheets: %v", err)
		metalSheets = []*models.MetalSheet{}
	}

	metalSheetsSection := components.PageTool_MetalSheets(user, metalSheets, tool)

	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tool metal sheets section")
	}

	return nil
}

func (h *Tool) HTMXGetToolNotes(c echo.Context) error {
	toolIDQuery, err := ParseQueryInt64(c, "tool_id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	h.Log.Debug("Fetching notes for tool %d", toolID)

	// Get the tool
	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool")
	}

	// Get notes for this tool
	notes, err := h.Registry.Notes.GetByTool(toolID)
	if err != nil {
		return HandleError(err, "failed to get notes for tool")
	}

	// Create ToolWithNotes for template compatibility
	resolvedTool := models.NewResolvedTool(tool, nil, notes)
	notesSection := components.PageTool_Notes(resolvedTool)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tool notes section")
	}

	return nil
}

func (h *Tool) HTMXGetEditRegeneration(c echo.Context) error {
	rid, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse regeneration id")
	}
	regenerationID := models.RegenerationID(rid)

	regeneration, err := h.Registry.ToolRegenerations.Get(regenerationID)
	if err != nil {
		return HandleError(err, "get regeneration failed")
	}

	resolvedRegeneration, err := services.ResolveRegeneration(h.Registry, regeneration)
	if err != nil {
		return err
	}

	dialog := components.PageTool_DialogEditRegeneration(resolvedRegeneration)

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "render dialog failed")
	}

	return nil
}

func (h *Tool) HTMXPutEditRegeneration(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	var regenerationID models.RegenerationID
	if id, err := ParseQueryInt64(c, "id"); err != nil {
		return HandleBadRequest(err, "failed to get the regeneration id from url query")
	} else {
		regenerationID = models.RegenerationID(id)
	}

	var regeneration *models.ResolvedRegeneration
	if r, err := h.Registry.ToolRegenerations.Get(regenerationID); err != nil {
		return HandleError(err, "failed to get regeneration before deleting")
	} else {
		regeneration, err = services.ResolveRegeneration(h.Registry, r)

		formData := h.parseRegenerationEditFormData(c)
		regeneration.Reason = formData.Reason
	}

	err = h.Registry.ToolRegenerations.Update(regeneration.Regeneration, user)
	if err != nil {
		return HandleError(err, "failed to update regeneration")
	}

	{ // Create Feed
		title := "Werkzeug Regenerierung aktualisiert"
		content := fmt.Sprintf(
			"Tool: %s\nGebundener Zyklus: %s (Teil Zyklen: %d)",
			regeneration.GetTool().String(),
			regeneration.GetCycle().Date.Format(env.DateFormat),
			regeneration.GetCycle().PartialCycles,
		)

		if regeneration.Reason != "" {
			content += fmt.Sprintf("\nReason: %s", regeneration.Reason)
		}

		h.createFeed(title, content, user.TelegramID)
	}

	SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Tool) HTMXDeleteRegeneration(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	var regenerationID models.RegenerationID
	if id, err := ParseQueryInt64(c, "id"); err != nil {
		return HandleBadRequest(err, "failed to get the regeneration id from url query")
	} else {
		regenerationID = models.RegenerationID(id)
	}

	h.Log.Info("Deleting the regeneration with the ID %d (User: %s)", regenerationID, user.Name)

	var regeneration *models.ResolvedRegeneration
	if r, err := h.Registry.ToolRegenerations.Get(regenerationID); err != nil {
		return HandleError(err, "failed to get regeneration before deleting")
	} else {
		regeneration, err = services.ResolveRegeneration(h.Registry, r)
	}

	if err := h.Registry.ToolRegenerations.Delete(regeneration.ID); err != nil {
		return HandleError(err, "failed to delete regeneration")
	}

	{ // Create Feed
		title := "Werkzeug Regenerierung entfernt"
		content := fmt.Sprintf(
			"Tool: %s\nGebundener Zyklus: %s (Teil Zyklen: %d)",
			regeneration.GetTool().String(),
			regeneration.GetCycle().Date.Format(env.DateFormat),
			regeneration.GetCycle().PartialCycles,
		)

		if regeneration.Reason != "" {
			content += fmt.Sprintf("\nReason: %s", regeneration.Reason)
		}

		if regeneration.PerformedBy != nil {
			user, err = h.Registry.Users.Get(*regeneration.PerformedBy)
			content += fmt.Sprintf("\nPerformed by: %s", user.Name)
		}

		h.createFeed(title, content, user.TelegramID)
	}

	SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Tool) HTMXGetStatusEdit(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool from database")
	}

	statusEdit := h.renderStatusComponent(tool, true, user)
	if err := statusEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tool status edit")
	}

	return nil
}

func (h *Tool) HTMXGetStatusDisplay(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	toolIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool from database")
	}

	statusDisplay := h.renderStatusComponent(tool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tool status display")
	}

	return nil
}

func (h *Tool) HTMXUpdateToolStatus(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	toolIDStr := c.FormValue("tool_id")
	if toolIDStr == "" {
		return HandleBadRequest(nil, "tool_id is required")
	}

	toolIDQuery, err := strconv.ParseInt(toolIDStr, 10, 64)
	if err != nil {
		return HandleBadRequest(nil, "invalid tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return HandleBadRequest(nil, "status is required")
	}

	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool from database")
	}

	h.Log.Info("User %s updating status for tool %d from %s to %s", user.Name, toolID, tool.Status(), statusStr)

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		// Start regeneration
		_, err := h.Registry.ToolRegenerations.StartToolRegeneration(toolID, "", user)
		if err != nil {
			return HandleError(err, "failed to start tool regeneration")
		}

		{ // Create Feed
			title := "Werkzeug Regenerierung gestartet"
			content := fmt.Sprintf("Werkzeug: %s", tool.String())
			h.createFeed(title, content, user.TelegramID)
		}

	case "active":
		if err := h.Registry.ToolRegenerations.StopToolRegeneration(toolID, user); err != nil {
			return HandleError(err, "failed to stop tool regeneration")
		}

		{ // Create Feed
			title := "Werkzeug Regenerierung beendet"
			content := fmt.Sprintf("Werkzeug: %s", tool.String())

			h.createFeed(title, content, user.TelegramID)
		}

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		if err := h.Registry.ToolRegenerations.AbortToolRegeneration(toolID, user); err != nil {
			return HandleError(err, "failed to abort tool regeneration")
		}

		{ // Create Feed
			title := "Werkzeug Regenerierung abgebrochen"
			content := fmt.Sprintf("Werkzeug: %s", tool.String())

			h.createFeed(title, content, user.TelegramID)
		}

	default:
		return HandleBadRequest(nil, "invalid status: must be 'regenerating', 'active', or 'abort'")
	}

	// Get updated tool and render status display
	updatedTool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get updated tool from database")
	}

	// Render the updated status component
	statusDisplay := h.renderStatusComponent(updatedTool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render updated tool status")
	}

	// Render out-of-band swap for cycles section to trigger reload
	oobCyclesReload := components.PageTool_CyclesSection(toolID, true)
	if err := oobCyclesReload.Render(c.Request().Context(), c.Response()); err != nil {
		h.Log.Error("Failed to render out-of-band cycles reload: %v", err)
	}

	return nil
}

func (h *Tool) getTotalCycles(toolID models.ToolID, cycles ...*models.Cycle) int64 {
	// Get regeneration for this tool
	var startCycleID models.CycleID
	if r, err := h.Registry.ToolRegenerations.GetLastRegeneration(toolID); err == nil {
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

func (h *Tool) getToolsForBinding(tool *models.Tool) ([]*models.Tool, error) {
	var filteredToolsForBinding []*models.Tool

	if tool.Position != models.PositionTopCassette && tool.Position != models.PositionTop {
		return filteredToolsForBinding, nil
	}

	// Get all tools
	tools, err := h.Registry.Tools.List()
	if err != nil {
		return nil, HandleError(err, "failed to get tools")
	}
	// Filter tools for binding
	for _, t := range tools {
		if t.Format != tool.Format || t.IsBound() {
			continue
		}

		if tool.Position == models.PositionTop {
			if t.Position == models.PositionTopCassette {
				filteredToolsForBinding = append(filteredToolsForBinding, t)
			}

			continue
		}

		// Else: position top cassette
		if t.Position == models.PositionTop {
			filteredToolsForBinding = append(filteredToolsForBinding, t)
		}
	}

	return filteredToolsForBinding, nil
}

func (h *Tool) getCycleFormData(c echo.Context) (*ToolCycleEditDialogFormData, error) {
	form := &ToolCycleEditDialogFormData{}

	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, err
		}

		pn := models.PressNumber(press)
		form.PressNumber = &pn
	}

	if dateString := c.FormValue("original_date"); dateString != "" {
		var err error
		form.Date, err = time.Parse(env.DateFormat, dateString)
		if err != nil {
			return nil, err
		}
	} else {
		form.Date = time.Now()
	}

	if totalCyclesString := c.FormValue("total_cycles"); totalCyclesString == "" {
		return nil, fmt.Errorf("form value total_cycles is required")
	} else {
		var err error
		form.TotalCycles, err = strconv.ParseInt(totalCyclesString, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	form.Regenerating = c.FormValue("regenerating") != ""

	// Parse tool_id if present (for tool change mode)
	if toolIDString := c.FormValue("tool_id"); toolIDString != "" {
		toolIDParsed, err := strconv.ParseInt(toolIDString, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid tool_id: %v", err)
		}
		toolID := models.ToolID(toolIDParsed)
		form.ToolID = &toolID
	}

	return form, nil
}

func (h *Tool) getToolFromQuery(c echo.Context) (*models.Tool, error) {
	toolIDQuery, err := ParseQueryInt64(c, "tool_id")
	if err != nil {
		return nil, err
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.Registry.Tools.Get(toolID)
	if err != nil {
		return nil, err
	}

	return tool, nil
}

func (h *Tool) parseRegenerationEditFormData(c echo.Context) *ToolRegenerationEditDialogFormData {
	return &ToolRegenerationEditDialogFormData{
		Reason: c.FormValue("reason"),
	}
}

func (h *Tool) renderStatusComponent(tool *models.Tool, editable bool, user *models.User) templ.Component {
	return components.ToolPage_ToolStatusEdit(&components.ToolPage_ToolStatusEditProps{
		Tool:              tool,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
}

func (h *Tool) createFeed(title, content string, userID models.TelegramID) {
	feed := models.NewFeed(title, content, userID)
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for cycle creation: %v", err)
	}
}
