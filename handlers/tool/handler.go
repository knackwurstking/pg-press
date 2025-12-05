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
	ui "github.com/knackwurstking/ui/ui-templ"
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

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool")
	}

	page := templates.Page(&templates.PageProps{
		User:       user,
		ToolString: tool.String(),
		ToolID:     tool.ID,
		Position:   tool.Position,
	})

	err := page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolPage")
	}

	return nil
}

func (h *Handler) HTMXPatchToolBinding(c echo.Context) error {
	slog.Info("Initiating tool binding operation")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

	t, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool")
	}

	rTool, err := services.ResolveTool(h.registry, t)
	if err != nil {
		return errors.HandlerError(err, "resolve tool")
	}

	var targetID models.ToolID
	{ // Get target_id from form value
		targetIDString := c.FormValue("target_id")
		if targetIDString == "" {
			return errors.HandlerError(nil, "failed to parse target_id: %#v", targetIDString)
		}

		targetIDParsed, err := strconv.ParseInt(targetIDString, 10, 64)
		if err != nil {
			return errors.NewBadRequestError(err, "invalid target_id")
		}
		targetID = models.ToolID(targetIDParsed)
	}

	{ // Make sure to check for position first (target == top && toolID == cassette)
		var (
			cassette models.ToolID
			target   models.ToolID // top position
		)

		if rTool.Position == models.PositionTopCassette {
			cassette = rTool.ID
			target = targetID
		} else {
			cassette = targetID // If this is not a cassette, the bind method will return an error
			target = rTool.ID
		}

		// Bind tool to target, this will get an error if target already has a binding
		if err := h.registry.Tools.Bind(cassette, target); err != nil {
			return errors.HandlerError(err, "bind tool")
		}
	}

	// Update tools binding, no need to re fetch this tools data from the database
	rTool.Binding = &targetID
	rTool, _ = services.ResolveTool(h.registry, rTool.Tool)

	// Get tools for binding
	toolsForBinding, eerr := h.getToolsForBinding(rTool.Tool)
	if eerr != nil {
		return eerr
	}

	// Render the template
	bs := templates.BindingSection(templates.BindingSectionProps{
		Tool:            rTool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         user.IsAdmin(),
	})

	err = bs.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}

	return nil
}

func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) error {
	slog.Info("Initiating tool unbinding operation")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

	dberr := h.registry.Tools.UnBind(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "unbind tool")
	}

	// Get tools for rendering the template
	t, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.NewBadRequestError(dberr, "get tool")
	}

	rTool, err := services.ResolveTool(h.registry, t)
	if err != nil {
		return errors.HandlerError(err, "resolve tool")
	}

	// Get tools for binding
	toolsForBinding, eerr := h.getToolsForBinding(rTool.Tool)
	if eerr != nil {
		return eerr
	}

	// Render the template
	bs := templates.BindingSection(templates.BindingSectionProps{
		Tool:            rTool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         user.IsAdmin(),
	})

	err = bs.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
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

	err := cyclesSection.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cycles")
	}

	return nil
}

func (h *Handler) HTMXGetToolTotalCycles(c echo.Context) error {
	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool")
	}

	// Get cycles for this specific tool
	toolCycles, dberr := h.registry.PressCycles.GetPressCyclesForTool(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get press cycles")
	}

	// Filter cycles by position
	filteredCycles := models.FilterCyclesByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	tc := templates.TotalCycles(totalCycles, utils.ParseQueryBool(c, "input"))
	err := tc.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "TotalCycles")
	}

	return nil
}

func (h *Handler) HTMXDeleteToolCycle(c echo.Context) error {
	slog.Info("Initiating cycle deletion operation")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse ID query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	// Get cycle data before deletion for the feed
	cycle, dberr := h.registry.PressCycles.Get(cycleID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get cycle for deletion")
	}

	tool, dberr := h.registry.Tools.Get(cycle.ToolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool for deletion")
	}

	// Check if there are any regenerations associated with this cycle
	hasRegenerations, dberr := h.registry.ToolRegenerations.HasRegenerationsForCycle(cycleID)
	if dberr != nil {
		return errors.HandlerError(dberr, "check for regenerations")
	}

	if hasRegenerations {
		return errors.NewBadRequestError(nil,
			"Cannot delete cycle: it has associated regenerations. "+
				"Please remove regenerations first.")
	}

	if err := h.registry.PressCycles.Delete(cycleID); err != nil {
		return errors.HandlerError(err, "delete press cycle")
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
	slog.Info("Retrieving metal sheet entries for tool")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolID, eerr := h.getToolIDFromParam(c)
	if eerr != nil {
		return eerr
	}

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool")
	}

	// Fetch metal sheets assigned to this tool
	metalSheets, dberr := h.registry.MetalSheets.ListByToolID(toolID)
	if dberr != nil {
		// Log error but don't fail - metal sheets are supplementary data
		slog.Error("Failed to fetch metal sheets", "error", dberr, "user_name", user.Name)
		metalSheets = []*models.MetalSheet{}
	}

	metalSheetsSection := templates.MetalSheets(user, metalSheets, tool)

	err := metalSheetsSection.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "metal-sheets")
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
		return errors.HandlerError(err, "get tool")
	}

	// Get notes for this tool
	notes, err := h.registry.Notes.GetByTool(toolID)
	if err != nil {
		return errors.HandlerError(err, "get notes for tool")
	}

	// Create ToolWithNotes for template compatibility
	resolvedTool := models.NewResolvedTool(tool, nil, notes, nil)
	notesSection := templates.Notes(resolvedTool)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.HandlerError(err, "render tool notes section")
	}

	return nil
}

func (h *Handler) HTMXDeleteRegeneration(c echo.Context) error {
	slog.Info("Deleting tool regeneration entry")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	var regenerationID models.ToolRegenerationID
	if id, err := utils.ParseQueryInt64(c, "id"); err != nil {
		return errors.NewBadRequestError(err, "get the regeneration id from url query")
	} else {
		regenerationID = models.ToolRegenerationID(id)
	}

	r, dberr := h.registry.ToolRegenerations.Get(regenerationID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get regeneration before deleting")
	}
	regeneration, err := services.ResolveToolRegeneration(h.registry, r)
	if err != nil {
		return errors.HandlerError(err, "resolve regeneration")
	}

	dberr = h.registry.ToolRegenerations.Delete(regeneration.ID)
	if dberr != nil {
		return errors.HandlerError(dberr, "delete regeneration")
	}

	// Create Feed
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
		return errors.NewBadRequestError(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool from database")
	}

	statusEdit := h.renderStatusComponent(tool, true, user)
	err = statusEdit.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "status")
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
		return errors.NewBadRequestError(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool from database")
	}

	statusDisplay := h.renderStatusComponent(tool, false, user)
	err = statusDisplay.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "status")
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
		return errors.NewBadRequestError(err, "parse id from query parameter")
	}
	toolID := models.ToolID(idParam)

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return errors.NewBadRequestError(nil, "status is required")
	}

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool from database")
	}

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		// Start regeneration
		_, dberr = h.registry.ToolRegenerations.StartToolRegeneration(toolID, "", user)
		if dberr != nil {
			return errors.HandlerError(dberr, "start tool regeneration")
		}

		// Create Feed
		title := "Werkzeug Regenerierung gestartet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		h.createFeed(title, content, user.TelegramID)

	case "active":
		dberr = h.registry.ToolRegenerations.StopToolRegeneration(toolID, user)
		if dberr != nil {
			return errors.HandlerError(dberr, "stop tool regeneration")
		}

		// Create Feed
		title := "Werkzeug Regenerierung beendet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())

		h.createFeed(title, content, user.TelegramID)

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		dberr = h.registry.ToolRegenerations.AbortToolRegeneration(toolID, user)
		if dberr != nil {
			return errors.HandlerError(dberr, "abort tool regeneration")
		}

		// Create Feed
		title := "Werkzeug Regenerierung abgebrochen"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())

		h.createFeed(title, content, user.TelegramID)

	default:
		return errors.NewBadRequestError(nil, "invalid status: must be 'regenerating', 'active', or 'abort'")
	}

	// Get updated tool and render status display
	updatedTool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get updated tool from database")
	}

	// Render the updated status component
	statusDisplay := h.renderStatusComponent(updatedTool, false, user)
	err = statusDisplay.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.HandlerError(err, "render updated tool status")
	}

	// Render out-of-band swap for cycles section to trigger reload
	cyclesProps, _ := h.buildCyclesProps(c)
	oobCyclesReload := templates.OOBCyclesSection(true, toolID, cyclesProps)
	err = oobCyclesReload.Render(c.Request().Context(), c.Response())
	if err != nil {
		slog.Error("Failed to render out-of-band cycles reload", "error", err)
	}

	return nil
}

func (h *Handler) getToolIDFromParam(c echo.Context) (models.ToolID, *echo.HTTPError) {
	toolIDQuery, err := utils.ParseParamInt64(c, "id")

	if err != nil {
		return 0, errors.NewBadRequestError(err, "parse \"id\"")
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
	tools, dberr := h.registry.Tools.List()
	if dberr != nil {
		return nil, errors.HandlerError(dberr, "get tools")
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

	t, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return nil, errors.HandlerError(dberr, "get tool")
	}
	tool, err := services.ResolveTool(h.registry, t)
	if err != nil {
		return nil, errors.HandlerError(err, "resolve tool")
	}

	var filteredCycles []*models.Cycle
	cycles, dberr := h.registry.PressCycles.GetPressCyclesForTool(toolID)
	if dberr != nil {
		return nil, errors.HandlerError(dberr, "get press cycles")
	}

	filteredCycles = models.FilterCyclesByToolPosition(
		tool.Position, cycles...)

	var resolvedRegenerations []*models.ResolvedToolRegeneration
	// Get (resolved) regeneration history for this tool
	regenerations, dberr := h.registry.ToolRegenerations.GetRegenerationHistory(toolID)
	if dberr != nil {
		slog.Error("Failed to get tool regenerations", "tool", toolID, "error", dberr)
	}

	// Resolve regenerations
	for _, r := range regenerations {
		rr, err := services.ResolveToolRegeneration(h.registry, r)
		if err != nil {
			return nil, errors.HandlerError(err, "")
		}

		resolvedRegenerations = append(resolvedRegenerations, rr)
	}

	// Only get tools for binding if the tool has no binding
	toolsForBinding, eerr := h.getToolsForBinding(tool.Tool)
	if eerr != nil {
		return nil, eerr
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
	_, dberr := h.registry.Feeds.AddSimple(title, content, userID)
	if dberr != nil {
		slog.Warn("Failed to create feed", "error", dberr)
	}
}
