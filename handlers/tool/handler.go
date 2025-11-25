package tool

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/tool/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Main Page
		utils.NewEchoRoute(http.MethodGet, "/tools/tool/:id", h.GetToolPage),

		// Regenerations Table
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

func (h *Handler) GetToolPage(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "get user from context")
	}

	idParam, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

	slog.Debug("Fetching tool with notes", "tool", toolID)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool")
	}

	page := components.PageTool(&components.PageToolProps{
		User:       user,
		ToolString: tool.String(),
		ToolID:     tool.ID,
		Position:   tool.Position,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool page")
	}

	return nil
}

func (h *Handler) HTMXPatchToolBinding(c echo.Context) error {
	var isAdmin bool
	{ // Check for admin
		user, err := utils.GetUserFromContext(c)
		if err != nil {
			return errors.BadRequest(err, "get user from context")
		}
		isAdmin = user.IsAdmin()
	}

	// Get tool from param "/:id"
	toolIDQuery, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	var tool *models.ResolvedTool
	if t, err := h.registry.Tools.Get(toolID); err != nil {
		return errors.Handler(err, "get tool")
	} else {
		tool, err = services.ResolveTool(h.registry, t)
		if err != nil {
			return errors.Handler(err, "resolve tool")
		}
	}

	var targetID models.ToolID
	{ // Get target_id from form value
		targetIDString := c.FormValue("target_id")
		if targetIDString == "" {
			return errors.BadRequest(nil, fmt.Sprintf(
				"failed to parse target_id: %+v", targetIDString))
		}

		targetIDParsed, err := strconv.ParseInt(targetIDString, 10, 64)
		if err != nil {
			return errors.BadRequest(err, "invalid target_id")
		}
		targetID = models.ToolID(targetIDParsed)
	}

	slog.Info("Updating tool binding", "tool", toolID, "target", targetID)

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
		if err = h.registry.Tools.Bind(cassette, target); err != nil {
			return errors.Handler(err, "bind tool")
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
		return errors.Handler(err, "render binding section")
	}

	return nil
}

func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) error {
	var isAdmin bool
	{ // Check for admin
		user, err := utils.GetUserFromContext(c)
		if err != nil {
			return errors.BadRequest(err, "get user from context")
		}
		isAdmin = user.IsAdmin()
	}

	// Get tool from param "/:id"
	toolIDParam, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse tool_id")
	}
	toolID := models.ToolID(toolIDParam)

	if err := h.registry.Tools.UnBind(toolID); err != nil {
		return errors.Handler(err, "unbind tool")
	}

	// Get tools for rendering the template
	var tool *models.ResolvedTool
	if t, err := h.registry.Tools.Get(toolID); err != nil {
		return errors.BadRequest(err, "get tool")
	} else {
		tool, err = services.ResolveTool(h.registry, t)
		if err != nil {
			return errors.Handler(err, "resolve tool")
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
		return errors.Handler(err, "render binding section")
	}

	return nil
}

func (h *Handler) HTMXGetCycles(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "get user from context")
	}

	toolIDParam, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.BadRequest(err, "parse tool_id")
	}
	toolID := models.ToolID(toolIDParam)

	var tool *models.ResolvedTool
	if t, err := h.registry.Tools.Get(toolID); err != nil {
		return errors.Handler(err, "get tool")
	} else {
		tool, err = services.ResolveTool(h.registry, t)
		if err != nil {
			return errors.Handler(err, "resolve tool")
		}
	}

	var filteredCycles []*models.Cycle
	{
		cycles, err := h.registry.PressCycles.GetPressCyclesForTool(toolID)
		if err != nil {
			return errors.Handler(err, "get press cycles")
		}

		filteredCycles = models.FilterCyclesByToolPosition(
			tool.Position, cycles...)
	}

	var resolvedRegenerations []*models.ResolvedToolRegeneration
	{ // Get (resolved) regeneration history for this tool
		regenerations, err := h.registry.ToolRegenerations.GetRegenerationHistory(toolID)
		if err != nil {
			slog.Error("Failed to get tool regenerations", "tool", toolID, "error", err)
		}

		// Resolve regenerations
		for _, r := range regenerations {
			rr, err := services.ResolveRegeneration(h.registry, r)
			if err != nil {
				return err
			}

			resolvedRegenerations = append(resolvedRegenerations, rr)
		}
	}

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
		TotalCycles:     h.getTotalCycles(toolID, filteredCycles...),
		Cycles:          filteredCycles,
		Regenerations:   resolvedRegenerations,
	})

	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		errors.Handler(err, "failed to render tool cycles")
	}

	return nil
}

func (h *Handler) HTMXGetToolTotalCycles(c echo.Context) error {
	// Get tool and position parameters
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.BadRequest(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool")
	}

	// Get cycles for this specific tool
	toolCycles, err := h.registry.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return errors.Handler(err, "get press cycles")
	}

	// Filter cycles by position
	filteredCycles := models.FilterCyclesByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	tc := components.TotalCycles(totalCycles, utils.ParseQueryBool(c, "input"))
	return tc.Render(c.Request().Context(), c.Response())
}

func (h *Handler) HTMXDeleteToolCycle(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "get user from context")
	}

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse ID query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	// Get cycle data before deletion for the feed
	cycle, err := h.registry.PressCycles.Get(cycleID)
	if err != nil {
		return errors.Handler(err, "get cycle for deletion")
	}

	tool, err := h.registry.Tools.Get(cycle.ToolID)
	if err != nil {
		return errors.Handler(err, "get tool for deletion")
	}

	// Check if there are any regenerations associated with this cycle
	hasRegenerations, err := h.registry.ToolRegenerations.HasRegenerationsForCycle(cycleID)
	if err != nil {
		return errors.Handler(err, "check for regenerations")
	}

	if hasRegenerations {
		return errors.BadRequest(nil, "Cannot delete cycle: it has associated regenerations. Please remove regenerations first.")
	}

	if err := h.registry.PressCycles.Delete(cycleID); err != nil {
		return errors.Handler(err, "delete press cycle")
	}

	{ // Create Feed
		title := fmt.Sprintf("Zyklus gelöscht für %s", tool.String())
		content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			cycle.PressNumber, tool.String(), cycle.TotalCycles, cycle.Date.Format("2006-01-02 15:04:05"))

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) HTMXGetToolMetalSheets(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.BadRequest(err, "parse tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	slog.Debug("Fetching metal sheets", "tool", toolID, "user_name", user.Name)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool")
	}

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.registry.MetalSheets.GetByToolID(toolID)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		slog.Error("Failed to fetch metal sheets", "error", err, "user_name", user.Name)
		metalSheets = []*models.MetalSheet{}
	}

	metalSheetsSection := components.PageTool_MetalSheets(user, metalSheets, tool)

	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool metal sheets section")
	}

	return nil
}

func (h *Handler) HTMXGetToolNotes(c echo.Context) error {
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.BadRequest(err, "parse tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	slog.Debug("Fetching notes for tool", "tool", toolID)

	// Get the tool
	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool")
	}

	// Get notes for this tool
	notes, err := h.registry.Notes.GetByTool(toolID)
	if err != nil {
		return errors.Handler(err, "get notes for tool")
	}

	// Create ToolWithNotes for template compatibility
	resolvedTool := models.NewResolvedTool(tool, nil, notes, nil)
	notesSection := components.PageTool_Notes(resolvedTool)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool notes section")
	}

	return nil
}

func (h *Handler) HTMXDeleteRegeneration(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "get user from context")
	}

	var regenerationID models.RegenerationID
	if id, err := utils.ParseQueryInt64(c, "id"); err != nil {
		return errors.BadRequest(err, "get the regeneration id from url query")
	} else {
		regenerationID = models.RegenerationID(id)
	}

	slog.Info("Deleting the regeneration", "regeneration", regenerationID, "user_name", user.Name)

	var regeneration *models.ResolvedRegeneration
	if r, err := h.registry.ToolRegenerations.Get(regenerationID); err != nil {
		return errors.Handler(err, "get regeneration before deleting")
	} else {
		regeneration, err = services.ResolveRegeneration(h.registry, r)
	}

	if err := h.registry.ToolRegenerations.Delete(regeneration.ID); err != nil {
		return errors.Handler(err, "delete regeneration")
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
			user, err = h.registry.Users.Get(*regeneration.PerformedBy)
			content += fmt.Sprintf("\nPerformed by: %s", user.Name)
		}

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) HTMXGetStatusEdit(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	statusEdit := h.renderStatusComponent(tool, true, user)
	if err := statusEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool status edit")
	}

	return nil
}

func (h *Handler) HTMXGetStatusDisplay(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	statusDisplay := h.renderStatusComponent(tool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool status display")
	}

	return nil
}

func (h *Handler) HTMXUpdateToolStatus(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "get user from context")
	}

	toolIDStr := c.FormValue("tool_id")
	if toolIDStr == "" {
		return errors.BadRequest(nil, "tool_id is required")
	}

	toolIDQuery, err := strconv.ParseInt(toolIDStr, 10, 64)
	if err != nil {
		return errors.BadRequest(nil, "invalid tool_id")
	}
	toolID := models.ToolID(toolIDQuery)

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return errors.BadRequest(nil, "status is required")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	slog.Info(
		"Updating tool status",
		"user_name", user.Name,
		"tool_id", toolID,
		"status-from", tool.Status(),
		"status-to", statusStr,
	)

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		// Start regeneration
		_, err := h.registry.ToolRegenerations.StartToolRegeneration(toolID, "", user)
		if err != nil {
			return errors.Handler(err, "start tool regeneration")
		}

		{ // Create Feed
			title := "Werkzeug Regenerierung gestartet"
			content := fmt.Sprintf("Werkzeug: %s", tool.String())
			h.createFeed(title, content, user.TelegramID)
		}

	case "active":
		if err := h.registry.ToolRegenerations.StopToolRegeneration(toolID, user); err != nil {
			return errors.Handler(err, "stop tool regeneration")
		}

		{ // Create Feed
			title := "Werkzeug Regenerierung beendet"
			content := fmt.Sprintf("Werkzeug: %s", tool.String())

			h.createFeed(title, content, user.TelegramID)
		}

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		if err := h.registry.ToolRegenerations.AbortToolRegeneration(toolID, user); err != nil {
			return errors.Handler(err, "abort tool regeneration")
		}

		{ // Create Feed
			title := "Werkzeug Regenerierung abgebrochen"
			content := fmt.Sprintf("Werkzeug: %s", tool.String())

			h.createFeed(title, content, user.TelegramID)
		}

	default:
		return errors.BadRequest(nil, "invalid status: must be 'regenerating', 'active', or 'abort'")
	}

	// Get updated tool and render status display
	updatedTool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get updated tool from database")
	}

	// Render the updated status component
	statusDisplay := h.renderStatusComponent(updatedTool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render updated tool status")
	}

	// Render out-of-band swap for cycles section to trigger reload
	oobCyclesReload := components.PageTool_CyclesSection(toolID, true)
	if err := oobCyclesReload.Render(c.Request().Context(), c.Response()); err != nil {
		slog.Error("Failed to render out-of-band cycles reload", "error", err)
	}

	return nil
}

func (h *Handler) getTotalCycles(toolID models.ToolID, cycles ...*models.Cycle) int64 {
	slog.Debug("Get total cycles", "tool", toolID, "cycles", len(cycles))

	// Get regeneration for this tool
	var startCycleID models.CycleID
	if r, err := h.registry.ToolRegenerations.GetLastRegeneration(toolID); err == nil {
		startCycleID = r.CycleID
	}

	var totalCycles int64

	for i, cycle := range cycles {
		if cycle.ID == startCycleID {
			slog.Debug("Stop counting...", "tool", toolID, "index", i, "cycle", cycle)
			break
		}

		totalCycles += cycle.PartialCycles
	}

	return totalCycles
}

func (h *Handler) getToolsForBinding(tool *models.Tool) ([]*models.Tool, error) {
	var filteredToolsForBinding []*models.Tool

	if tool.Position != models.PositionTopCassette && tool.Position != models.PositionTop {
		return filteredToolsForBinding, nil
	}

	// Get all tools
	tools, err := h.registry.Tools.List()
	if err != nil {
		return nil, errors.Handler(err, "get tools")
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

func (h *Handler) renderStatusComponent(tool *models.Tool, editable bool, user *models.User) templ.Component {
	return components.ToolPage_ToolStatusEdit(&components.ToolPage_ToolStatusEditProps{
		Tool:              tool,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
}

func (h *Handler) createFeed(title, content string, userID models.TelegramID) {
	feed := models.NewFeed(title, content, userID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "error", err)
	}
}
