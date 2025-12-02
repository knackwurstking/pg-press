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
	var (
		tool             *models.Tool
		cycle            *models.Cycle
		tools            []*models.Tool
		inputPressNumber *models.PressNumber
		inputTotalCycles int64
		originalDate     *time.Time
	)

	// Check if we're in tool change mode
	toolChangeMode := utils.ParseQueryBool(c, "tool_change_mode")

	if c.QueryParam("id") != "" {
		cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
		if err != nil {
			return errors.BadRequest(err, "parse cycle ID")
		}
		cycleID := models.CycleID(cycleIDQuery)

		// Get cycle data from the database
		cycle, err = h.registry.PressCycles.Get(cycleID)
		if err != nil {
			return errors.Handler(err, "load cycle data")
		}
		inputPressNumber = &(cycle.PressNumber)
		inputTotalCycles = cycle.TotalCycles
		originalDate = &cycle.Date

		// Set the cycles (original) tool to props
		if tool, err = h.registry.Tools.Get(cycle.ToolID); err != nil {
			return errors.Handler(err, "load tool data")
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			// Get all tools
			allTools, err := h.registry.Tools.List()
			if err != nil {
				return errors.Handler(err, "load available tools")
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
			return errors.BadRequest(err, "parse tool ID")
		}
		toolID := models.ToolID(toolIDQuery)

		if tool, err = h.registry.Tools.Get(toolID); err != nil {
			return errors.Handler(err, "load tool data")
		}
	} else {
		return errors.BadRequest(nil, "missing tool or cycle ID")
	}

	var dialog templ.Component
	if cycle != nil {
		dialog = templates.EditCycleDialog(tool, cycle, tools, inputPressNumber, inputTotalCycles, originalDate)
	} else {
		dialog = templates.NewCycleDialog(tool, inputPressNumber, inputTotalCycles, originalDate)
	}

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render cycle edit dialog")
	}

	return nil
}

func (h *Handler) PostEditCycle(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	slog.Info("Creating a new cycle", "user_name", user.Name)

	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.BadRequest(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "load tool data")
	}

	// Parse form data
	form, err := getEditCycleFormData(c)
	if err != nil {
		return errors.BadRequest(err, "parse form data")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return errors.BadRequest(nil, "press_number must be a valid integer")
	}

	pressCycle := models.NewCycle(
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
	)

	pressCycle.Date = form.Date

	_, err = h.registry.PressCycles.Add(pressCycle, user)
	if err != nil {
		return errors.Handler(err, "add cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		slog.Info("Starting regeneration", "tool", tool.ID, "user_name", user.Name)
		_, err := h.registry.ToolRegenerations.StartToolRegeneration(tool.ID, "", user)
		if err != nil {
			slog.Error("Failed to start regeneration", "tool", tool.ID, "user_name", user.Name, "error", err)
		}
	}

	{ // Create Feed
		title := fmt.Sprintf("Neuer Zyklus hinzugefügt für %s", tool.String())
		content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))

		if form.Regenerating {
			content += "\nRegenerierung gestartet"
		}

		if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
			slog.Error("Failed to create feed for cycle creation", "error", err)
		}
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditCycle(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	slog.Info("Updating cycle", "user_name", user.Name)

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse ID from query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	cycle, err := h.registry.PressCycles.Get(cycleID)
	if err != nil {
		return errors.Handler(err, "get cycle")
	}

	// Get original tool
	originalTool, err := h.registry.Tools.Get(cycle.ToolID)
	if err != nil {
		return errors.Handler(err, "get original tool")
	}

	form, err := getEditCycleFormData(c)
	if err != nil {
		return errors.BadRequest(err, "get cycle form data from query")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return errors.BadRequest(nil, "press_number must be a valid integer")
	}

	// Determine which tool to use for the cycle
	var tool *models.Tool
	if form.ToolID != nil {
		// Tool change requested - get the new tool
		newTool, err := h.registry.Tools.Get(*form.ToolID)
		if err != nil {
			return errors.Handler(err, "get new tool")
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

	if err := h.registry.PressCycles.Update(pressCycle, user); err != nil {
		return errors.Handler(err, "update press cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		if _, err := h.registry.ToolRegenerations.Add(tool.ID, pressCycle.ID, "", user); err != nil {
			slog.Error("Failed to add tool regeneration", "error", err)
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

		if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
			slog.Error("Failed to create feed for cycle update", "error", err)
		}
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditTool(c echo.Context) error {
	var tool *models.Tool

	toolIDQuery, _ := utils.ParseQueryInt64(c, "id")
	if toolIDQuery > 0 {
		var err error
		tool, err = h.registry.Tools.Get(models.ToolID(toolIDQuery))
		if err != nil {
			return errors.Handler(err, "get tool from database")
		}
		slog.Info("Opening edit dialog for tool", "tool_id", tool.ID)
	} else {
		slog.Info("Opening new tool dialog")
	}

	var d templ.Component
	if tool != nil {
		d = templates.EditToolDialog(tool)
	} else {
		d = templates.NewToolDialog()
	}

	if err := d.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render tool dialog")
	}

	return nil
}

func (h *Handler) PostEditTool(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	formData, err := getEditToolFormData(c)
	if err != nil {
		return errors.BadRequest(err, "get tool form data")
	}

	tool := models.NewTool(formData.Position, formData.Format, formData.Code, formData.Type)
	tool.SetPress(formData.Press)

	id, err := h.registry.Tools.Add(tool, user)
	if err != nil {
		return errors.Handler(err, "add tool")
	}

	slog.Info("Created tool", "id", id, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	title := "Neues Werkzeug erstellt"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Error("Failed to create feed for tool creation", "error", err)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)
	return nil
}

func (h *Handler) PutEditTool(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	formData, err := getEditToolFormData(c)
	if err != nil {
		return errors.BadRequest(err, "get tool form data")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool")
	}

	tool.Press = formData.Press
	tool.Position = formData.Position
	tool.Format = formData.Format
	tool.Code = formData.Code
	tool.Type = formData.Type

	if err := h.registry.Tools.Update(tool, user); err != nil {
		return errors.Handler(err, "update tool")
	}

	slog.Info("Updated tool", "id", tool.ID, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	title := "Werkzeug aktualisiert"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Error("Failed to create feed for tool update", "error", err)
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
	var (
		toolID     models.ToolID
		metalSheet *models.MetalSheet
	)

	// Check if we're editing an existing metal sheet (has ID) or creating new one
	if metalSheetIDQuery, _ := utils.ParseQueryInt64(c, "id"); metalSheetIDQuery > 0 {
		metalSheetID := models.MetalSheetID(metalSheetIDQuery)

		// Fetch existing metal sheet for editing
		var err error
		if metalSheet, err = h.registry.MetalSheets.Get(metalSheetID); err != nil {
			return errors.Handler(err, "fetch metal sheet from database")
		}
		toolID = metalSheet.ToolID
		slog.Info("Opening edit dialog for metal sheet", "metal_sheet_id", metalSheetID)
	} else {
		// Creating new metal sheet, get tool_id from query
		toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
		if err != nil {
			return errors.Handler(err, "get the tool id from query")
		}
		toolID = models.ToolID(toolIDQuery)
		slog.Info("Opening new metal sheet dialog", "tool_id", toolID)
	}

	// Fetch the associated tool for the dialog
	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	var d templ.Component
	if metalSheet != nil {
		d = templates.EditMetalSheetDialog(metalSheet, tool)
	} else {
		d = templates.NewMetalSheetDialog(tool)
	}

	if err = d.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render metal sheet dialog")
	}

	return nil
}

func (h *Handler) PostEditMetalSheet(c echo.Context) error {
	// Get current user for feed creation
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	slog.Info("Creating a new metal sheet", "user_name", user.Name)

	// Extract tool ID from query parameters
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.Handler(err, "get tool_id from query")
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool
	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	// Parse form data into metal sheet model
	metalSheet, err := getMetalSheetFormData(c)
	if err != nil {
		return errors.Handler(err, "parse metal sheet form data")
	}

	// Associate metal sheet with the tool
	metalSheet.ToolID = toolID

	// Save new metal sheet to database
	if _, err := h.registry.MetalSheets.Add(metalSheet); err != nil {
		return errors.Handler(err, "create metal sheet in database")
	}

	h.createNewMetalSheetFeed(user, tool, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditMetalSheet(c echo.Context) error {
	// Get current user for feed creation
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	slog.Info("Updating metal sheet", "user_name", user.Name)

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "get id from query")
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet to preserve ID and tool association
	existingSheet, err := h.registry.MetalSheets.Get(metalSheetID)
	if err != nil {
		return errors.Handler(err, "get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.registry.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return errors.Handler(err, "get tool from database")
	}

	// Parse updated form data
	metalSheet, err := getMetalSheetFormData(c)
	if err != nil {
		return errors.Handler(err, "parse metal sheet form data")
	}

	// Preserve the original ID and tool association
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	if err := h.registry.MetalSheets.Update(metalSheet); err != nil {
		return errors.Handler(err, "update metal sheet in database")
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

		slog.Info("Opening edit dialog for note", "note", noteID, "user_name", user.Name)

		var err error
		note, err = h.registry.Notes.Get(noteID)
		if err != nil {
			return errors.Handler(err, "get note from database")
		}
	} else {
		slog.Info("Opening new note dialog", "user_name", user.Name)
	}

	var d templ.Component
	if note != nil {
		d = templates.EditNoteDialog(note, linkToTables, user)
	} else {
		d = templates.NewNoteDialog(linkToTables, user)
	}

	if err := d.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render edit note dialog")
	}

	return nil
}

func (h *Handler) PostEditNote(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	note, err := getNoteFromFormData(c)
	if err != nil {
		return errors.BadRequest(err, "parse note form data")
	}

	// Create the note
	noteID, err := h.registry.Notes.Add(note)
	if err != nil {
		return errors.Handler(err, "create note")
	}

	slog.Info("Created note", "note", noteID, "user_name", user.Name)

	// Create feed entry
	title := "Neue Notiz erstellt"
	content := fmt.Sprintf("Eine neue Notiz wurde erstellt: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Error("Failed to create feed for cycle creation", "error", err)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditNote(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	idq, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse note ID")
	}
	noteID := models.NoteID(idq)

	note, err := getNoteFromFormData(c)
	if err != nil {
		return errors.BadRequest(err, "parse note form data")
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	if err := h.registry.Notes.Update(note); err != nil {
		return errors.Handler(err, "update note")
	}

	slog.Info("Updated note", "user_name", user.Name, "note", noteID)

	// Create feed entry
	title := "Notiz aktualisiert"
	content := fmt.Sprintf("Eine Notiz wurde aktualisiert: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Error("Failed to create feed for cycle creation", "error", err)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditToolRegeneration(c echo.Context) error {
	rid, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "parse regeneration id")
	}
	regenerationID := models.ToolRegenerationID(rid)

	regeneration, err := h.registry.ToolRegenerations.Get(regenerationID)
	if err != nil {
		return errors.Handler(err, "get regeneration")
	}

	resolvedRegeneration, err := services.ResolveToolRegeneration(h.registry, regeneration)
	if err != nil {
		return err
	}

	slog.Info("Opening edit dialog for tool regeneration", "regeneration_id", regenerationID, "tool", resolvedRegeneration.GetTool().String())

	dialog := templates.EditToolRegenerationDialog(resolvedRegeneration)

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render dialog")
	}

	return nil
}

func (h *Handler) PutEditToolRegeneration(c echo.Context) error {
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

		formData := getEditRegenerationFormData(c)
		regeneration.Reason = formData.Reason
	}

	slog.Info("Update tool regeneration", "id", regeneration.ID, "tool", regeneration.GetTool().String())

	if err := h.registry.ToolRegenerations.Update(regeneration.ToolRegeneration, user); err != nil {
		return errors.Handler(err, "update regeneration")
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

		if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
			slog.Error("Failed to create feed for cycle creation", "error", err)
		}
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditPressRegeneration(c echo.Context) error {
	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "press regeneration id query")
	}

	r, err := h.registry.PressRegenerations.Get(models.PressRegenerationID(id))
	if err != nil {
		return errors.Handler(err, "get press regeneration from database")
	}

	slog.Info("Opening edit dialog for press regeneration", "regeneration_id", id, "press", r.PressNumber)

	d := templates.EditPressRegenerationDialog(r)
	if err = d.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press regeneration dialog")
	}

	return nil
}

func (h *Handler) PutEditPressRegeneration(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "")
	}

	r, err := h.registry.PressRegenerations.Get(models.PressRegenerationID(id))
	if err != nil {
		return errors.Handler(err, "press regenerations")
	}

	slog.Info("Update press regeneration", "id", id, "press", r.PressNumber, "reason", r.Reason)

	r.Reason = c.FormValue("reason")
	if err = h.registry.PressRegenerations.Update(r); err != nil {
		return errors.Handler(err, "press regenerations")
	}

	feedTitle := "Pressen Regenerierung Aktualisiert"
	feedContent := fmt.Sprintf("Presse: %d", r.PressNumber)
	feedContent += fmt.Sprintf("Von: %s, Bis: %s", r.StartedAt.Format(env.DateTimeFormat), r.CompletedAt.Format(env.DateTimeFormat))
	feedContent += fmt.Sprintf("Bemerkung: %s", r.Reason)
	if _, err = h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID); err != nil {
		slog.Warn("Add feed", "error", err, "title", feedTitle)
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
	if _, err := h.registry.Feeds.AddSimple("Blech erstellt", content, user.TelegramID); err != nil {
		slog.Error("Failed to create feed", "error", err)
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
	if _, err := h.registry.Feeds.AddSimple("Blech aktualisiert", content, user.TelegramID); err != nil {
		slog.Error("Failed to create update feed", "error", err)
	}
}
