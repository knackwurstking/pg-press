// TODO: Split this handlers, on file per dialog
package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/dialogs/templates"
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
		// Edit cycle dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-cycle", h.GetEditCycle),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-cycle", h.PostEditCycle),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-cycle", h.PutEditCycle),

		// Edit tool dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-tool", h.GetEditTool),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-tool", h.PostEditTool),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool", h.PutEditTool),

		// Edit metal sheet dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-metal-sheet", h.GetEditMetalSheet),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-metal-sheet", h.PostEditMetalSheet),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-metal-sheet", h.PutEditMetalSheet),

		// Edit note dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-note", h.GetEditNote),
		ui.NewEchoRoute(http.MethodPost, path+"/edit-note", h.PostEditNote),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-note", h.PutEditNote),

		// Edit tool regeneration dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-tool-regeneration", h.GetEditToolRegeneration),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-tool-regeneration", h.PutEditToolRegeneration),

		// Edit press regeneration dialog
		ui.NewEchoRoute(http.MethodGet, path+"/edit-press-regeneration", h.GetEditPressRegeneration),
		ui.NewEchoRoute(http.MethodPut, path+"/edit-press-regeneration", h.PutEditPressRegeneration),
	})
}

func (h *Handler) GetEditCycle(c echo.Context) error {
	// Check if we're in tool change mode
	toolChangeMode := utils.ParseQueryBool(c, "tool_change_mode")

	var (
		tool             *models.Tool
		cycle            *models.Cycle
		tools            []*models.Tool
		inputPressNumber *models.PressNumber
		inputTotalCycles int64
		originalDate     *time.Time
	)

	if c.QueryParam("id") != "" {
		cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
		if err != nil {
			return errors.NewBadRequestError(err, "parse cycle ID")
		}
		cycleID := models.CycleID(cycleIDQuery)

		// Get cycle data from the database
		var dberr *errors.DBError
		cycle, dberr = h.registry.PressCycles.Get(cycleID)
		if dberr != nil {
			return errors.HandlerError(dberr, "load cycle data")
		}
		inputPressNumber = &(cycle.PressNumber)
		inputTotalCycles = cycle.TotalCycles
		originalDate = &cycle.Date

		// Set the cycles (original) tool to props
		tool, dberr = h.registry.Tools.Get(cycle.ToolID)
		if dberr != nil {
			return errors.HandlerError(dberr, "load tool data")
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			// Get all tools
			allTools, dberr := h.registry.Tools.List()
			if dberr != nil {
				return errors.HandlerError(dberr, "load available tools")
			}

			// Filter out tools not matching the original tools position
			for _, t := range allTools {
				if t.Position == tool.Position {
					tools = append(tools, t)
				}
			}

			// Sort tools alphabetically by code
			sort.Slice(tools, func(i, j int) bool {
				return tools[i].String() < tools[j].String()
			})
		}
	} else if c.QueryParam("tool_id") != "" {
		toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
		if err != nil {
			return errors.NewBadRequestError(err, "parse tool ID")
		}
		toolID := models.ToolID(toolIDQuery)

		var dberr *errors.DBError
		tool, dberr = h.registry.Tools.Get(toolID)
		if dberr != nil {
			return errors.HandlerError(dberr, "load tool data")
		}
	} else {
		return errors.NewBadRequestError(nil, "missing tool or cycle ID")
	}

	var dialog templ.Component
	if cycle != nil {
		dialog = templates.EditCycleDialog(
			tool, cycle, tools, inputPressNumber, inputTotalCycles, originalDate,
		)
	} else {
		dialog = templates.NewCycleDialog(
			tool, inputPressNumber, inputTotalCycles, originalDate,
		)
	}

	err := dialog.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "CycleDialog")
	}

	return nil
}

func (h *Handler) PostEditCycle(c echo.Context) error {
	slog.Info("Cycle creation request received")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "load tool data")
	}

	// Parse form data
	form, err := getEditCycleFormData(c)
	if err != nil {
		return errors.NewBadRequestError(err, "parse form data")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return errors.NewBadRequestError(nil, "press_number must be a valid integer")
	}

	pressCycle := models.NewCycle(
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
	)

	pressCycle.Date = form.Date

	_, dberr = h.registry.PressCycles.Add(pressCycle, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "add cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		slog.Info("Starting tool regeneration", "tool_id", tool.ID, "user", user.Name)
		_, dberr = h.registry.ToolRegenerations.StartToolRegeneration(tool.ID, "", user)
		if dberr != nil {
			slog.Error(
				"Failed to start tool regeneration",
				"tool_id", tool.ID, "user", user.Name, "error", dberr,
			)
		}
	}

	// Create Feed
	title := fmt.Sprintf("Neuer Zyklus hinzugefügt für %s", tool.String())
	content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
		*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))

	if form.Regenerating {
		content += "\nRegenerierung gestartet"
	}

	_, dberr = h.registry.Feeds.AddSimple(title, content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", dberr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditCycle(c echo.Context) error {
	slog.Info("Updating cycle")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse ID from query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	cycle, dberr := h.registry.PressCycles.Get(cycleID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get cycle")
	}

	// Get original tool
	originalTool, dberr := h.registry.Tools.Get(cycle.ToolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get original tool")
	}

	form, err := getEditCycleFormData(c)
	if err != nil {
		return errors.NewBadRequestError(err, "get cycle form data from query")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return errors.NewBadRequestError(nil, "press_number must be a valid integer")
	}

	// Determine which tool to use for the cycle
	var tool *models.Tool
	if form.ToolID != nil {
		// Tool change requested - get the new tool
		newTool, dberr := h.registry.Tools.Get(*form.ToolID)
		if dberr != nil {
			return errors.HandlerError(dberr, "get new tool")
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

	dberr = h.registry.PressCycles.Update(pressCycle, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "update press cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		_, dberr = h.registry.ToolRegenerations.Add(tool.ID, pressCycle.ID, "", user)
		if dberr != nil {
			slog.Error("Failed to add tool regeneration", "error", dberr)
		}
	}

	// Create Feed
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

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed for cycle update", "error", err)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditTool(c echo.Context) error {
	toolIDQuery, _ := utils.ParseQueryInt64(c, "id")
	var tool *models.Tool
	if toolIDQuery > 0 {
		var dberr *errors.DBError
		tool, dberr = h.registry.Tools.Get(models.ToolID(toolIDQuery))
		if dberr != nil {
			return errors.HandlerError(dberr, "get tool from database")
		}
	}

	var d templ.Component
	if tool != nil {
		d = templates.EditToolDialog(tool)
	} else {
		d = templates.NewToolDialog()
	}

	err := d.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolDialog")
	}

	return nil
}

func (h *Handler) PostEditTool(c echo.Context) error {
	slog.Info("Creating new tool")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	formData, err := getEditToolFormData(c)
	if err != nil {
		return errors.NewBadRequestError(err, "get tool form data")
	}

	tool := models.NewTool(formData.Position, formData.Format, formData.Code, formData.Type)
	tool.SetPress(formData.Press)

	_, dberr := h.registry.Tools.Add(tool, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "add tool")
	}

	// Create feed entry
	title := "Neues Werkzeug erstellt"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	_, dberr = h.registry.Feeds.AddSimple(title, content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for tool creation", "error", dberr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditTool(c echo.Context) error {
	slog.Info("Updating existing tool")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	formData, err := getEditToolFormData(c)
	if err != nil {
		return errors.NewBadRequestError(err, "get tool form data")
	}

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool")
	}

	tool.Press = formData.Press
	tool.Position = formData.Position
	tool.Format = formData.Format
	tool.Code = formData.Code
	tool.Type = formData.Type

	dberr = h.registry.Tools.Update(tool, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "update tool")
	}

	// Create feed entry
	title := "Werkzeug aktualisiert"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	_, dberr = h.registry.Feeds.AddSimple(title, content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for tool update", "error", dberr)
	}

	// Set HX headers
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	utils.SetHXAfterSettle(c, map[string]any{
		"toolUpdated": map[string]string{
			"pageTitle": fmt.Sprintf("PG Presse | %s %s",
				tool.String(), tool.Position.GermanString()),
			"appBarTitle": fmt.Sprintf("%s %s", tool.String(),
				tool.Position.GermanString()),
		},
	})

	return nil
}

func (h *Handler) GetEditMetalSheet(c echo.Context) error {
	// Check if we're editing an existing metal sheet (has ID) or creating new one
	metalSheetIDQuery, _ := utils.ParseQueryInt64(c, "id")
	if metalSheetIDQuery > 0 {
		metalSheetID := models.MetalSheetID(metalSheetIDQuery)

		// Fetch existing metal sheet for editing
		metalSheet, dberr := h.registry.MetalSheets.Get(metalSheetID)
		if dberr != nil {
			return errors.HandlerError(dberr, "fetch metal sheet from database")
		}

		// Fetch the associated tool for the dialog
		tool, dberr := h.registry.Tools.Get(metalSheet.ToolID)
		if dberr != nil {
			return errors.HandlerError(dberr, "get tool from database")
		}

		var d templ.Component
		d = templates.EditMetalSheetDialog(metalSheet, tool)

		err := d.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "MetalSheetDialog")
		}
	}

	// Creating new metal sheet, get tool_id from query
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.HandlerError(err, "get the tool id from query")
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool for the dialog
	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool from database")
	}

	var d templ.Component
	d = templates.NewMetalSheetDialog(tool)

	err = d.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.HandlerError(err, "MetalSheetDialog")
	}

	return nil
}

func (h *Handler) PostEditMetalSheet(c echo.Context) error {
	slog.Info("Metal sheet creation request received")

	// Get current user for feed creation
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Extract tool ID from query parameters
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.HandlerError(err, "get tool_id from query")
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool
	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool from database")
	}

	// Parse form data into metal sheet model
	metalSheet, err := getMetalSheetFormData(c)
	if err != nil {
		return errors.HandlerError(err, "parse metal sheet form data")
	}

	// Associate metal sheet with the tool
	metalSheet.ToolID = toolID

	// Save new metal sheet to database
	_, dberr = h.registry.MetalSheets.Add(metalSheet)
	if dberr != nil {
		return errors.HandlerError(dberr, "create metal sheet in database")
	}

	h.createNewMetalSheetFeed(user, tool, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditMetalSheet(c echo.Context) error {
	slog.Info("Updating metal sheet")

	// Get current user for feed creation
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "get id from query")
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet to preserve ID and tool association
	existingSheet, dberr := h.registry.MetalSheets.Get(metalSheetID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, dberr := h.registry.Tools.Get(existingSheet.ToolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool from database")
	}

	// Parse updated form data
	metalSheet, err := getMetalSheetFormData(c)
	if err != nil {
		return errors.HandlerError(err, "parse metal sheet form data")
	}

	// Preserve the original ID and tool association
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	dberr = h.registry.MetalSheets.Update(metalSheet)
	if dberr != nil {
		return errors.HandlerError(dberr, "update metal sheet in database")
	}

	h.createUpdateMetalSheetFeed(user, tool, existingSheet, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditNote(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	var (
		linkToTables []string
		note         *models.Note
	)

	// Parse linked tables from query parameter
	if ltt := c.QueryParam("link_to_tables"); ltt != "" {
		linkToTables = strings.Split(ltt, ",")
	}

	// Check if we're editing an existing note
	if id, _ := utils.ParseQueryInt64(c, "id"); id > 0 {
		noteID := models.NoteID(id)

		var dberr *errors.DBError
		note, dberr = h.registry.Notes.Get(noteID)
		if dberr != nil {
			return errors.HandlerError(dberr, "get note from database")
		}
	}

	var d templ.Component
	if note != nil {
		d = templates.EditNoteDialog(note, linkToTables, user)
	} else {
		d = templates.NewNoteDialog(linkToTables, user)
	}

	err := d.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.HandlerError(err, "NoteDialog")
	}

	return nil
}

func (h *Handler) PostEditNote(c echo.Context) error {
	slog.Info("Creating new note")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	note, err := getNoteFromFormData(c)
	if err != nil {
		return errors.NewBadRequestError(err, "parse note form data")
	}

	// Create feed entry
	title := "Neue Notiz erstellt"
	content := fmt.Sprintf("Eine neue Notiz wurde erstellt: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	_, dberr := h.registry.Feeds.AddSimple(title, content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", dberr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditNote(c echo.Context) error {
	slog.Info("Updating note")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	idq, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse note ID")
	}
	noteID := models.NoteID(idq)

	note, err := getNoteFromFormData(c)
	if err != nil {
		return errors.NewBadRequestError(err, "parse note form data")
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	dberr := h.registry.Notes.Update(note)
	if dberr != nil {
		return errors.HandlerError(dberr, "update note")
	}

	// Create feed entry
	title := "Notiz aktualisiert"
	content := fmt.Sprintf("Eine Notiz wurde aktualisiert: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	_, dberr = h.registry.Feeds.AddSimple(title, content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", dberr)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditToolRegeneration(c echo.Context) error {
	rid, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse regeneration id")
	}
	regenerationID := models.ToolRegenerationID(rid)

	regeneration, dberr := h.registry.ToolRegenerations.Get(regenerationID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get regeneration")
	}

	resolvedRegeneration, err := services.ResolveToolRegeneration(h.registry, regeneration)
	if err != nil {
		return err
	}

	dialog := templates.EditToolRegenerationDialog(resolvedRegeneration)
	err = dialog.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.HandlerError(err, "ToolRegeneration")
	}

	return nil
}

func (h *Handler) PutEditToolRegeneration(c echo.Context) error {
	slog.Info("Updating tool regeneration entry")

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

	formData := getEditRegenerationFormData(c)
	regeneration.Reason = formData.Reason

	dberr = h.registry.ToolRegenerations.Update(regeneration.ToolRegeneration, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "update regeneration")
	}

	// Create Feed
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

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", err)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditPressRegeneration(c echo.Context) error {
	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "press regeneration id query")
	}

	r, dberr := h.registry.PressRegenerations.Get(models.PressRegenerationID(id))
	if dberr != nil {
		return errors.HandlerError(dberr, "get press regeneration from database")
	}

	d := templates.EditPressRegenerationDialog(r)
	err = d.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "PressRegeneration")
	}

	return nil
}

func (h *Handler) PutEditPressRegeneration(c echo.Context) error {
	slog.Info("Updating press regeneration entry")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "")
	}

	r, dberr := h.registry.PressRegenerations.Get(models.PressRegenerationID(id))
	if dberr != nil {
		return errors.HandlerError(dberr, "press regenerations")
	}

	r.Reason = c.FormValue("reason")
	dberr = h.registry.PressRegenerations.Update(r)
	if dberr != nil {
		return errors.HandlerError(dberr, "press regenerations")
	}

	feedTitle := fmt.Sprintf("\"Regenerierung\" für Presse %d aktualisiert", r.PressNumber)
	feedContent := fmt.Sprintf("Presse: %d\n", r.PressNumber)
	feedContent += fmt.Sprintf("Von: %s, Bis: %s\n", r.StartedAt.Format(env.DateTimeFormat), r.CompletedAt.Format(env.DateTimeFormat))
	feedContent += fmt.Sprintf("Bemerkung: %s", r.Reason)

	_, dberr = h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID)
	if dberr != nil {
		slog.Warn("Add feed", "title", feedTitle, "error", dberr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) createNewMetalSheetFeed(user *models.User, tool *models.Tool, metalSheet *models.MetalSheet) {
	// Build base feed content with tool and metal sheet info
	content := fmt.Sprintf("Werkzeug: %s\nStärke: %.1f mm\nBlech: %.1f mm\nTyp: %s",
		tool.String(), metalSheet.TileHeight, metalSheet.Value, metalSheet.Identifier.String())

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		content += fmt.Sprintf("\nMarke: %d mm\nStf.: %.1f\nStf. Max: %.1f",
			metalSheet.MarkeHeight, metalSheet.STF, metalSheet.STFMax)
	}

	// Create and save the feed entry
	_, dberr := h.registry.Feeds.AddSimple("Blech erstellt", content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed", "error", dberr)
	}
}

func (h *Handler) createUpdateMetalSheetFeed(user *models.User, tool *models.Tool, oldSheet, newSheet *models.MetalSheet) {
	content := fmt.Sprintf("Werkzeug: %s", tool.String())

	// Check for changes in TileHeight
	if oldSheet.TileHeight != newSheet.TileHeight {
		content += fmt.Sprintf("\nStärke: %.1f mm → %.1f mm", oldSheet.TileHeight, newSheet.TileHeight)
	} else {
		content += fmt.Sprintf("\nStärke: %.1f mm", newSheet.TileHeight)
	}

	// Check for changes in Value
	if oldSheet.Value != newSheet.Value {
		content += fmt.Sprintf("\nBlech: %.1f mm → %.1f mm", oldSheet.Value, newSheet.Value)
	} else {
		content += fmt.Sprintf("\nBlech: %.1f mm", newSheet.Value)
	}

	// Check for changes in machine type
	if oldSheet.Identifier != newSheet.Identifier {
		content += fmt.Sprintf("\nTyp: %s → %s", oldSheet.Identifier.String(), newSheet.Identifier.String())
	} else {
		content += fmt.Sprintf("\nTyp: %s", newSheet.Identifier.String())
	}

	// Add additional fields for bottom position tools
	if tool.Position == models.PositionBottom {
		// Check for changes in MarkeHeight
		if oldSheet.MarkeHeight != newSheet.MarkeHeight {
			content += fmt.Sprintf("\nMarke: %d mm → %d mm", oldSheet.MarkeHeight, newSheet.MarkeHeight)
		} else {
			content += fmt.Sprintf("\nMarke: %d mm", newSheet.MarkeHeight)
		}

		// Check for changes in STF
		if oldSheet.STF != newSheet.STF {
			content += fmt.Sprintf("\nStf.: %.1f → %.1f", oldSheet.STF, newSheet.STF)
		} else {
			content += fmt.Sprintf("\nStf.: %.1f", newSheet.STF)
		}

		// Check for changes in STFMax
		if oldSheet.STFMax != newSheet.STFMax {
			content += fmt.Sprintf("\nStf. Max: %.1f → %.1f", oldSheet.STFMax, newSheet.STFMax)
		} else {
			content += fmt.Sprintf("\nStf. Max: %.1f", newSheet.STFMax)
		}
	}

	// Create and save the feed entry
	_, dberr := h.registry.Feeds.AddSimple("Blech aktualisiert", content, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create update feed", "error", dberr)
	}
}
