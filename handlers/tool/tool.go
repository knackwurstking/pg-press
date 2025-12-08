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
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(&templates.PageProps{
		User:       user,
		ToolString: tool.String(),
		ToolID:     tool.ID,
		Position:   tool.Position,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tool Page")
	}

	return nil
}

func (h *Handler) HTMXPatchToolBinding(c echo.Context) error {
	slog.Info("Initiating tool binding operation")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	t, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	rTool, merr := services.ResolveTool(h.registry, t)
	if merr != nil {
		return merr.Echo()
	}

	// Get target_id from form value
	targetIDString := c.FormValue("target_id")
	if targetIDString == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("failed to parse target_id: %#v", targetIDString),
		)
	}

	targetIDParsed, err := strconv.ParseInt(targetIDString, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid target_id",
		)
	}
	targetID := models.ToolID(targetIDParsed)

	// Make sure to check for position first (target == top && toolID == cassette)
	var (
		cassette, target models.ToolID
	)
	if rTool.Position == models.PositionTopCassette {
		cassette = rTool.ID
		target = targetID
	} else {
		cassette = targetID // If this is not a cassette, the bind method will return an error
		target = rTool.ID
	}

	// Bind tool to target, this will get an error if target already has a binding
	merr = h.registry.Tools.Bind(cassette, target)
	if merr != nil {
		return merr.Echo()
	}

	// Update tools binding, no need to re fetch this tools data from the database
	rTool.Binding = &targetID
	rTool, _ = services.ResolveTool(h.registry, rTool.Tool)

	// Get tools for binding
	toolsForBinding, merr := h.getToolsForBinding(rTool.Tool)
	if merr != nil {
		return merr.Echo()
	}

	// Render the template
	te := templates.BindingSection(templates.BindingSectionProps{
		Tool:            rTool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         user.IsAdmin(),
	})

	err = te.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}

	return nil
}

func (h *Handler) HTMXPatchToolUnBinding(c echo.Context) error {
	slog.Info("Initiating tool unbinding operation")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	merr = h.registry.Tools.UnBind(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get tools for rendering the template
	t, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	rTool, merr := services.ResolveTool(h.registry, t)
	if merr != nil {
		return merr.Echo()
	}

	// Get tools for binding
	toolsForBinding, merr := h.getToolsForBinding(rTool.Tool)
	if merr != nil {
		return merr.Echo()
	}

	// Render the template
	te := templates.BindingSection(templates.BindingSectionProps{
		Tool:            rTool,
		ToolsForBinding: toolsForBinding,
		IsAdmin:         user.IsAdmin(),
	})
	err := te.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "BindingSection")
	}

	return nil
}

func (h *Handler) HTMXGetCycles(c echo.Context) error {
	// Render the template
	cyclesProps, merr := h.buildCyclesProps(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Cycles(*cyclesProps)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Cycles")
	}

	return nil
}

func (h *Handler) HTMXGetToolTotalCycles(c echo.Context) error {
	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get cycles for this specific tool
	toolCycles, merr := h.registry.PressCycles.GetPressCyclesForTool(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Filter cycles by position
	filteredCycles := models.FilterCyclesByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	t := templates.TotalCycles(totalCycles, utils.ParseQueryBool(c, "input"))
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "TotalCycles")
	}

	return nil
}

func (h *Handler) HTMXDeleteToolCycle(c echo.Context) error {
	slog.Info("Initiating cycle deletion operation")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cycleIDQuery, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	cycleID := models.CycleID(cycleIDQuery)

	// Get cycle data before deletion for the feed
	cycle, merr := h.registry.PressCycles.Get(cycleID)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(cycle.ToolID)
	if merr != nil {
		return merr.Echo()
	}

	// Check if there are any regenerations associated with this cycle
	hasRegenerations, merr := h.registry.ToolRegenerations.HasRegenerationsForCycle(cycleID)
	if merr != nil {
		return merr.Echo()
	}

	if hasRegenerations {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"Cannot delete cycle: it has associated regenerations. "+
				"Please remove regenerations first.",
		)
	}

	merr = h.registry.PressCycles.Delete(cycleID)
	if merr != nil {
		return merr.Echo()
	}

	// Create Feed
	title := fmt.Sprintf("Zyklus gelöscht für %s", tool.String())
	content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
		cycle.PressNumber, tool.String(), cycle.TotalCycles, cycle.Date.Format("2006-01-02 15:04:05"))

	h.createFeed(title, content, user.TelegramID)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) HTMXGetToolMetalSheets(c echo.Context) error {
	slog.Info("Retrieving metal sheet entries for tool")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Fetch metal sheets assigned to this tool
	metalSheets, merr := h.registry.MetalSheets.ListByToolID(toolID)
	if merr != nil {
		// Log error but don't fail - metal sheets are supplementary data
		slog.Error("Failed to fetch metal sheets", "error", merr, "user_name", user.Name)
		metalSheets = []*models.MetalSheet{}
	}

	t := templates.MetalSheets(user, metalSheets, tool)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "MetalSheets")
	}

	return nil
}

func (h *Handler) HTMXGetToolNotes(c echo.Context) error {
	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get the tool
	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get notes for this tool
	notes, merr := h.registry.Notes.GetByTool(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Create ToolWithNotes for template compatibility
	resolvedTool := models.NewResolvedTool(tool, nil, notes, nil)
	t := templates.Notes(resolvedTool)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes")
	}

	return nil
}

func (h *Handler) HTMXDeleteRegeneration(c echo.Context) error {
	slog.Info("Deleting tool regeneration entry")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	regenerationID := models.ToolRegenerationID(id)

	r, merr := h.registry.ToolRegenerations.Get(regenerationID)
	if merr != nil {
		return merr.Echo()
	}
	regeneration, merr := services.ResolveToolRegeneration(h.registry, r)
	if merr != nil {
		return merr.Echo()
	}

	merr = h.registry.ToolRegenerations.Delete(regeneration.ID)
	if merr != nil {
		return merr.Echo()
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
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	idParam, merr := utils.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(idParam)

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	t := h.renderStatusComponent(tool, true, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolStatusEdit")
	}

	return nil
}

func (h *Handler) HTMXGetStatusDisplay(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	idParam, merr := utils.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(idParam)

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	t := h.renderStatusComponent(tool, false, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolStatusEdit")
	}

	return nil
}

func (h *Handler) HTMXUpdateToolStatus(c echo.Context) error {
	slog.Info("Change the tool status")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	idParam, merr := utils.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(idParam)

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		// Start regeneration
		_, merr = h.registry.ToolRegenerations.StartToolRegeneration(toolID, "", user)
		if merr != nil {
			return merr.Echo()
		}

		// Create Feed
		title := "Werkzeug Regenerierung gestartet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		h.createFeed(title, content, user.TelegramID)

	case "active":
		merr = h.registry.ToolRegenerations.StopToolRegeneration(toolID, user)
		if merr != nil {
			return merr.Echo()
		}

		// Create Feed
		title := "Werkzeug Regenerierung beendet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())

		h.createFeed(title, content, user.TelegramID)

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		merr = h.registry.ToolRegenerations.AbortToolRegeneration(toolID, user)
		if merr != nil {
			return merr.Echo()
		}

		// Create Feed
		title := "Werkzeug Regenerierung abgebrochen"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())

		h.createFeed(title, content, user.TelegramID)

	default:
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid status: must be 'regenerating', 'active', or 'abort'",
		)
	}

	// Get updated tool and render status display
	updatedTool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Render the updated status component
	t := h.renderStatusComponent(updatedTool, false, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolStatusEdit")
	}

	// Render out-of-band swap for cycles section to trigger reload
	cyclesProps, _ := h.buildCyclesProps(c)
	t = templates.OOBCyclesSection(true, toolID, cyclesProps)
	err = t.Render(c.Request().Context(), c.Response())
	if err != nil {
		slog.Error("Failed to render out-of-band cycles reload", "error", err)
	}

	return nil
}

func (h *Handler) getToolIDFromParam(c echo.Context) (models.ToolID, *errors.MasterError) {
	toolIDQuery, merr := utils.ParseParamInt64(c, "id")
	if merr != nil {
		return 0, merr
	}

	return models.ToolID(toolIDQuery), nil
}

func (h *Handler) getTotalCycles(toolID models.ToolID, cycles ...*models.Cycle) int64 {
	// Get regeneration for this tool
	var startCycleID models.CycleID

	if r, merr := h.registry.ToolRegenerations.GetLastRegeneration(toolID); merr == nil {
		startCycleID = r.CycleID
	} else {
		if merr.Code != http.StatusNotFound {
			slog.Warn("Failed to get the last regeneration data", "tool_id", toolID)
		}
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

func (h *Handler) getToolsForBinding(tool *models.Tool) ([]*models.Tool, *errors.MasterError) {
	var filteredToolsForBinding []*models.Tool

	if tool.Position != models.PositionTopCassette && tool.Position != models.PositionTop {
		return filteredToolsForBinding, nil
	}

	// Get all tools
	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return nil, merr
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

func (h *Handler) buildCyclesProps(c echo.Context) (*templates.CyclesProps, *errors.MasterError) {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return nil, merr
	}

	toolID, merr := h.getToolIDFromParam(c)
	if merr != nil {
		return nil, merr
	}

	t, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return nil, merr
	}
	tool, merr := services.ResolveTool(h.registry, t)
	if merr != nil {
		return nil, merr
	}

	var filteredCycles []*models.Cycle
	cycles, merr := h.registry.PressCycles.GetPressCyclesForTool(toolID)
	if merr != nil {
		return nil, merr
	}

	filteredCycles = models.FilterCyclesByToolPosition(
		tool.Position, cycles...)

	var resolvedRegenerations []*models.ResolvedToolRegeneration
	// Get (resolved) regeneration history for this tool
	regenerations, merr := h.registry.ToolRegenerations.GetRegenerationHistory(toolID)
	if merr != nil {
		slog.Error("Failed to get tool regenerations", "tool", toolID, "error", merr)
	}

	// Resolve regenerations
	for _, r := range regenerations {
		rr, merr := services.ResolveToolRegeneration(h.registry, r)
		if merr != nil {
			return nil, merr
		}

		resolvedRegenerations = append(resolvedRegenerations, rr)
	}

	// Only get tools for binding if the tool has no binding
	toolsForBinding, merr := h.getToolsForBinding(tool.Tool)
	if merr != nil {
		return nil, merr
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
	merr := h.registry.Feeds.Add(title, content, userID)
	if merr != nil {
		slog.Warn("Failed to create feed", "error", merr)
	}
}
