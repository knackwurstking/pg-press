package tool

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/tool/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/knackwurstking/ui"
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

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Main Page
		ui.NewEchoRoute(http.MethodGet, path+"/:id", h.GetToolPage),

		// Regenerations Table
		ui.NewEchoRoute(http.MethodDelete,
			path+"/:id/delete-regeneration", h.HTMXDeleteRegeneration),

		// Tool status and regenerations management
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/status-edit", h.HTMXGetStatusEdit),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/status-display", h.HTMXGetStatusDisplay),
		ui.NewEchoRoute(http.MethodPut,
			path+"/:id/status", h.HTMXUpdateToolStatus),

		// Section loading
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/notes", h.HTMXGetToolNotes),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/metal-sheets", h.HTMXGetToolMetalSheets),

		// Cycles table rows
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/cycles", h.HTMXGetCycles),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:id/total-cycles", h.HTMXGetToolTotalCycles),

		// Delete a cycle table entry
		ui.NewEchoRoute(http.MethodDelete,
			path+"/cycle/delete", h.HTMXDeleteToolCycle),

		// Update tools binding data
		ui.NewEchoRoute(http.MethodPatch,
			path+"/:id/bind", h.HTMXPatchToolBinding),
		ui.NewEchoRoute(http.MethodPatch,
			path+"/:id/unbind", h.HTMXPatchToolUnBinding),
	})
}

func (h *Handler) GetToolPage(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool")
	}

	page := templates.Page(&templates.PageProps{
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
	slog.Info("Bind a tool")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

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
			return errors.BadRequest(nil, "failed to parse target_id: %#v", targetIDString)
		}

		targetIDParsed, err := strconv.ParseInt(targetIDString, 10, 64)
		if err != nil {
			return errors.BadRequest(err, "invalid target_id")
		}
		targetID = models.ToolID(targetIDParsed)
	}

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
		if err := h.registry.Tools.Bind(cassette, target); err != nil {
			return errors.Handler(err, "bind tool")
		}
	}

	// Update tools binding, no need to re fetch this tools data from the database
	tool.Binding = &targetID

	// Get tools for binding
	toolsForBinding, eerr := h.getToolsForBinding(tool.Tool)
	if eerr != nil {
		return eerr
	}

	// Render the template
	bs := templates.BindingSection(templates.BindingSectionProps{
		Tool:            tool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         user.IsAdmin(),
	})

	if err := bs.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render binding section")
	}

	return nil
}

func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) error {
	slog.Info("Unbind a tool")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

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
	toolsForBinding, eerr := h.getToolsForBinding(tool.Tool)
	if eerr != nil {
		return eerr
	}

	// Render the template
	bs := templates.BindingSection(templates.BindingSectionProps{
		Tool:            tool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         user.IsAdmin(),
	})

	if err := bs.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render binding section")
	}

	return nil
}

func (h *Handler) HTMXGetCycles(c echo.Context) error {
	// Render the template
	cyclesProps, eerr := h.buildCyclesProps(c)
	if eerr != nil {
		return eerr
	}
	cyclesSection := templates.Cycles(*cyclesProps)

	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		errors.Handler(err, "failed to render tool cycles")
	}

	return nil
}

func (h *Handler) HTMXGetToolTotalCycles(c echo.Context) error {
	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

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

	tc := templates.TotalCycles(totalCycles, utils.ParseQueryBool(c, "input"))
	return tc.Render(c.Request().Context(), c.Response())
}

func (h *Handler) HTMXDeleteToolCycle(c echo.Context) error {
	slog.Info("Deleting a cycle")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
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
	slog.Info("Get metal sheets entries for a tool")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

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

	metalSheetsSection := templates.MetalSheets(user, metalSheets, tool)

	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool metal sheets section")
	}

	return nil
}

func (h *Handler) HTMXGetToolNotes(c echo.Context) error {
	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

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
	notesSection := templates.Notes(resolvedTool)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool notes section")
	}

	return nil
}

func (h *Handler) HTMXDeleteRegeneration(c echo.Context) error {
	slog.Info("Delete a tool regeneration entry")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	var regenerationID models.ToolRegenerationID
	if id, err := utils.ParseQueryInt64(c, "id"); err != nil {
		return errors.BadRequest(err, "get the regeneration id from url query")
	} else {
		regenerationID = models.ToolRegenerationID(id)
	}

	var regeneration *models.ResolvedToolRegeneration
	if r, err := h.registry.ToolRegenerations.Get(regenerationID); err != nil {
		return errors.Handler(err, "get regeneration before deleting")
	} else {
		regeneration, err = services.ResolveToolRegeneration(h.registry, r)
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
			user, err := h.registry.Users.Get(*regeneration.PerformedBy)
			if err != nil {
				slog.Warn("User not found", "error", err, "performed_by", regeneration.PerformedBy)
			}
			content += fmt.Sprintf("\nPerformed by: %s", user.Name)
		}

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) HTMXGetStatusEdit(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	idParam, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

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
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	idParam, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

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
	slog.Info("Change the tool status")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	idParam, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return errors.BadRequest(nil, "status is required")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

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
	cyclesProps, _ := h.buildCyclesProps(c)
	oobCyclesReload := templates.OOBCyclesSection(true, toolID, cyclesProps)
	if err := oobCyclesReload.Render(c.Request().Context(), c.Response()); err != nil {
		slog.Error("Failed to render out-of-band cycles reload", "error", err)
	}

	return nil
}

func (h *Handler) getToolIDFromParam(c echo.Context) (models.ToolID, *echo.HTTPError) {
	toolIDQuery, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return 0, errors.BadRequest(err, "parse \"id\"")
	}
	return models.ToolID(toolIDQuery), nil
}

func (h *Handler) getTotalCycles(toolID models.ToolID, cycles ...*models.Cycle) int64 {
	// Get regeneration for this tool
	var startCycleID models.CycleID

	if r, err := h.registry.ToolRegenerations.GetLastRegeneration(toolID); err == nil {
		startCycleID = r.CycleID
	}

	var totalCycles int64

	for _, cycle := range cycles {
		if cycle.ID == startCycleID {
			break
		}

		totalCycles += cycle.PartialCycles
	}

	return totalCycles
}

func (h *Handler) getToolsForBinding(tool *models.Tool) ([]*models.Tool, *echo.HTTPError) {
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

func (h *Handler) buildCyclesProps(c echo.Context) (*templates.CyclesProps, *echo.HTTPError) {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return nil, eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return nil, eerr
	}

	var tool *models.ResolvedTool
	if t, err := h.registry.Tools.Get(toolID); err != nil {
		return nil, errors.Handler(err, "get tool")
	} else {
		tool, err = services.ResolveTool(h.registry, t)
		if err != nil {
			return nil, errors.Handler(err, "resolve tool")
		}
	}

	var filteredCycles []*models.Cycle
	{
		cycles, err := h.registry.PressCycles.GetPressCyclesForTool(toolID)
		if err != nil {
			return nil, errors.Handler(err, "get press cycles")
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
			rr, err := services.ResolveToolRegeneration(h.registry, r)
			if err != nil {
				return nil, errors.Handler(err, "")
			}

			resolvedRegenerations = append(resolvedRegenerations, rr)
		}
	}

	// Only get tools for binding if the tool has no binding
	toolsForBinding, err := h.getToolsForBinding(tool.Tool)
	if err != nil {
		return nil, err
	}

	return &templates.CyclesProps{
		User:            user,
		Tool:            tool,
		ToolsForBinding: toolsForBinding,
		TotalCycles:     h.getTotalCycles(toolID, filteredCycles...),
		Cycles:          filteredCycles,
		Regenerations:   resolvedRegenerations,
	}, nil
}

func (h *Handler) renderStatusComponent(tool *models.Tool, editable bool, user *models.User) templ.Component {
	return templates.ToolStatusEdit(&templates.ToolStatusEditProps{
		Tool:              tool,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
}

func (h *Handler) createFeed(title, content string, userID models.TelegramID) {
	if _, err := h.registry.Feeds.AddSimple(title, content, userID); err != nil {
		slog.Warn("Failed to create feed", "error", err)
	}
}
