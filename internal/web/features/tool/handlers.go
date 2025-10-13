package tool

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/tool/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type EditToolCycleDialogFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *models.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
	ToolID       *int64 // ToolID form field name "tool_id" (for tool change mode)
}

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(
			db,
			logger.NewComponentLogger("Tool"),
		),
	}
}

func (h *Handler) GetToolPage(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	id, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c,
			"failed to parse id from query parameter:"+err.Error())
	}

	h.Log.Debug("Fetching tool %d with notes", id)

	tool, err := h.DB.Tools.GetWithNotes(id)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	page := templates.Page(&templates.PageProps{
		User:       user,
		ToolString: tool.String(),
		ToolID:     tool.ID,
		Position:   tool.Position,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool page: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetStatusEdit(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	statusEdit := h.renderStatusComponent(tool, true, user)
	if err := statusEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool status edit: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetStatusDisplay(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	statusDisplay := h.renderStatusComponent(tool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool status display: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXUpdateToolStatus(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolIDStr := c.FormValue("tool_id")
	if toolIDStr == "" {
		return h.RenderBadRequest(c, "tool_id is required")
	}

	toolID, err := strconv.ParseInt(toolIDStr, 10, 64)
	if err != nil {
		return h.RenderBadRequest(c, "invalid tool_id: "+err.Error())
	}

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return h.RenderBadRequest(c, "status is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	h.Log.Info("User %s updating status for tool %d from %s to %s", user.Name, toolID, tool.Status(), statusStr)

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		// Start regeneration
		if err := h.DB.Tools.UpdateRegenerating(toolID, true, user); err != nil {
			return h.HandleError(c, err, "failed to start tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung gestartet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for regeneration start: %v", err)
		}

	case "active":
		// Stop regeneration (return to active status)
		if err := h.DB.Tools.UpdateRegenerating(toolID, false, user); err != nil {
			return h.HandleError(c, err, "failed to stop tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung beendet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for regeneration stop: %v", err)
		}

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		if err := h.DB.ToolRegenerations.AbortToolRegeneration(toolID, user); err != nil {
			return h.HandleError(c, err, "failed to abort tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung abgebrochen"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for regeneration abort: %v", err)
		}

	default:
		return h.RenderBadRequest(c, "invalid status: must be 'regenerating', 'active', or 'abort'")
	}

	// Get updated tool and render status display
	updatedTool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get updated tool from database")
	}

	statusDisplay := h.renderStatusComponent(updatedTool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render updated tool status: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetToolCycles(c echo.Context) error {
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

	toolCycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get press cycles")
	}

	filteredCycles := models.FilterCyclesByToolPosition(
		tool.Position, toolCycles...)

	regeneration, err := h.DB.ToolRegenerations.GetLastRegeneration(toolID)
	if err != nil {
		h.Log.Error("Failed to get regenerations for tool %d: %v", toolID, err)
	}

	totalCycles := h.getTotalCycles(
		toolID,
		filteredCycles...,
	)

	var tools []*models.Tool
	if tool.Position == models.PositionTopCassette {
		tools, err = h.DB.Tools.List()
		if err != nil {
			return h.HandleError(c, err, "failed to get tools")
		}
	}

	cyclesSection := templates.ToolCycles(&templates.ToolCyclesProps{
		User:             user,
		Tool:             tool,
		Tools:            tools,
		TotalCycles:      totalCycles,
		Cycles:           filteredCycles,
		LastRegeneration: regeneration,
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
	toolChangeMode := c.QueryParam("tool_change_mode") == "true"

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
			allTools, err := h.DB.Tools.ListWithNotes()
			if err != nil {
				return h.HandleError(c, err, "failed to load available tools")
			}

			// Filter out tools not matching the original tools position
			for _, t := range allTools {
				if t.Tool.Position == props.Tool.Position {
					props.AvailableTools = append(props.AvailableTools, t.Tool)
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

	return h.HTMXGetToolCycles(c)
}

func (h *Handler) HTMXGetToolNotes(c echo.Context) error {
	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	h.Log.Debug("Fetching notes for tool %d", toolID)

	// Get the tool
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	// Get notes for this tool
	notes, err := h.DB.Notes.GetByTool(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get notes for tool")
	}

	// Create ToolWithNotes for template compatibility
	toolWithNotes := &models.ToolWithNotes{
		Tool:        tool,
		LoadedNotes: notes,
	}

	notesSection := templates.SectionNotes(toolWithNotes)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render tool notes section")
	}

	return nil
}

func (h *Handler) HTMXGetToolMetalSheets(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool_id: "+err.Error())
	}

	h.Log.Debug("Fetching metal sheets for tool %d", toolID)

	tool, err := h.DB.Tools.GetWithNotes(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(toolID)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		h.Log.Error("Failed to fetch metal sheets: %v", err)
		metalSheets = []*models.MetalSheet{}
	}

	metalSheetsSection := templates.SectionMetalSheets(user, metalSheets, tool)

	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.HandleError(c, err, "failed to render tool metal sheets section")
	}

	return nil
}

func (h *Handler) HTMXPatchToolBinding(c echo.Context) error {
	// TODO: Update tools binding, get id from param and 'target_id' from hx-vals (however)

	return errors.New("under construction")
}

// *************** //
// Private Methods //
// *************** //

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

// ************************ //
// TEMPL: Rendering methods //
// ************************ //

func (h *Handler) renderStatusComponent(tool *models.Tool, editable bool, user *models.User) templ.Component {
	return components.ToolStatusEdit(&components.ToolStatusEditProps{
		Tool:              tool,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
}

// ********************* //
// Get (input) form data //
// ********************* //

func (h *Handler) getCycleFormData(c echo.Context) (*EditToolCycleDialogFormData, error) {
	form := &EditToolCycleDialogFormData{}

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
		form.Date, err = time.Parse(constants.DateFormat, dateString)
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
		toolID, err := strconv.ParseInt(toolIDString, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid tool_id: %v", err)
		}
		form.ToolID = &toolID
	}

	return form, nil
}

func (h *Handler) getToolFromQuery(c echo.Context) (*models.Tool, error) {
	toolID, err := h.ParseInt64Query(c, "tool_id")
	if err != nil {
		return nil, err
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return nil, err
	}

	return tool, nil
}
