package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/dialogs/components"
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
		// Edit cycle dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-cycle", h.GetEditCycle),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-cycle", h.PostEditCycle),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-cycle", h.PutEditCycle),

		// Edit tool dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-tool", h.GetEditTool),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-tool", h.PostEditTool),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-tool", h.PutEditTool),

		// Edit metal sheet dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-metal-sheet", h.GetEditMetalSheet),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-metal-sheet", h.PostEditMetalSheet),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-metal-sheet", h.PutEditMetalSheet),

		// Edit note dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-note", h.GetEditNote),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-note", h.PostEditNote),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-note", h.PutEditNote),

		// Edit regeneration dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-regeneration", h.GetEditRegeneration),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-regeneration", h.PutEditRegeneration),
	})
}

func (h *Handler) GetEditCycle(c echo.Context) error {
	props := &components.DialogEditCycleProps{}

	// Check if we're in tool change mode
	toolChangeMode := utils.ParseQueryBool(c, "tool_change_mode")

	if c.QueryParam("id") != "" {
		cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
		if err != nil {
			return errors.BadRequest(err, "failed to parse cycle ID")
		}
		props.CycleID = models.CycleID(cycleIDQuery)

		// Get cycle data from the database
		cycle, err := h.registry.PressCycles.Get(props.CycleID)
		if err != nil {
			return errors.Handler(err, "failed to load cycle data")
		}
		props.InputPressNumber = &(cycle.PressNumber)
		props.InputTotalCycles = cycle.TotalCycles
		props.OriginalDate = &cycle.Date

		// Set the cycles (original) tool to props
		if props.Tool, err = h.registry.Tools.Get(cycle.ToolID); err != nil {
			return errors.Handler(err, "failed to load tool data")
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			props.AllowToolChange = true

			// Get all tools
			allTools, err := h.registry.Tools.List()
			if err != nil {
				return errors.Handler(err, "failed to load available tools")
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
		toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
		if err != nil {
			return errors.BadRequest(err, "failed to parse tool ID")
		}
		toolID := models.ToolID(toolIDQuery)

		if props.Tool, err = h.registry.Tools.Get(toolID); err != nil {
			return errors.Handler(err, "failed to load tool data")
		}
	} else {
		return errors.BadRequest(nil, "missing tool or cycle ID")
	}

	cycleEditDialog := components.DialogEditCycle(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "failed to render cycle edit dialog")
	}

	return nil
}

func (h *Handler) PostEditCycle(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.BadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "failed to load tool data")
	}

	// Parse form data
	form, err := getEditCycleFormData(c)
	if err != nil {
		return errors.BadRequest(err, "failed to parse form data")
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
		return errors.Handler(err, "failed to add cycle")
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

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditCycle(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "failed to get user from context")
	}

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "failed to parse ID from query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	cycle, err := h.registry.PressCycles.Get(cycleID)
	if err != nil {
		return errors.Handler(err, "failed to get cycle")
	}

	// Get original tool
	originalTool, err := h.registry.Tools.Get(cycle.ToolID)
	if err != nil {
		return errors.Handler(err, "failed to get original tool")
	}

	form, err := getEditCycleFormData(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get cycle form data from query")
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
			return errors.Handler(err, "failed to get new tool")
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
		return errors.Handler(err, "failed to update press cycle")
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

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditTool(c echo.Context) error {
	props := &components.DialogEditToolProps{}

	toolIDQuery, _ := utils.ParseQueryInt64(c, "id")
	if toolIDQuery > 0 {
		tool, err := h.registry.Tools.Get(models.ToolID(toolIDQuery))
		if err != nil {
			return errors.Handler(err, "failed to get tool from database")
		}

		props.Tool = tool
		props.InputPosition = string(tool.Position)
		props.InputWidth = tool.Format.Width
		props.InputHeight = tool.Format.Height
		props.InputType = tool.Type
		props.InputCode = tool.Code
		props.InputPressSelection = tool.Press
	}

	dialog := components.DialogEditTool(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "failed to render tool edit dialog")
	}
	return nil
}

func (h *Handler) PostEditTool(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	formData, err := getEditToolFormData(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get tool form data")
	}

	tool := models.NewTool(formData.Position, formData.Format, formData.Code, formData.Type)
	tool.SetPress(formData.Press)

	id, err := h.registry.Tools.Add(tool, user)
	if err != nil {
		return errors.Handler(err, "failed to add tool")
	}

	slog.Info("Created tool", "id", id, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	title := "Neues Werkzeug erstellt"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	h.createFeed(title, content, user.TelegramID)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)
	return nil
}

func (h *Handler) PutEditTool(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	formData, err := getEditToolFormData(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get tool form data")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "failed to get tool")
	}

	tool.Press = formData.Press
	tool.Position = formData.Position
	tool.Format = formData.Format
	tool.Code = formData.Code
	tool.Type = formData.Type

	if err := h.registry.Tools.Update(tool, user); err != nil {
		return errors.Handler(err, "failed to update tool")
	}

	slog.Info("Updated tool", "id", tool.ID, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	title := "Werkzeug aktualisiert"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	h.createFeed(title, content, user.TelegramID)

	// Set HX headers
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	utils.SetHXAfterSettle(c, map[string]interface{}{
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
	renderProps := &components.DialogEditMetalSheetProps{}
	var toolID models.ToolID
	var err error

	// Check if we're editing an existing metal sheet (has ID) or creating new one
	if metalSheetIDQuery, _ := utils.ParseQueryInt64(c, "id"); metalSheetIDQuery > 0 {
		metalSheetID := models.MetalSheetID(metalSheetIDQuery)

		// Fetch existing metal sheet for editing
		if renderProps.MetalSheet, err = h.registry.MetalSheets.Get(metalSheetID); err != nil {
			return errors.Handler(err, "failed to fetch metal sheet from database")
		}
		toolID = renderProps.MetalSheet.ToolID
	} else {
		// Creating new metal sheet, get tool_id from query
		toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
		if err != nil {
			return errors.Handler(err, "failed to get the tool id from query")
		}
		toolID = models.ToolID(toolIDQuery)
	}

	// Fetch the associated tool for the dialog
	if renderProps.Tool, err = h.registry.Tools.Get(toolID); err != nil {
		return errors.Handler(err, "failed to get tool from database")
	}

	// Render the edit dialog template
	d := components.DialogEditMetalSheet(renderProps)
	if err := d.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "failed to render edit metal sheet dialog")
	}

	return nil
}

func (h *Handler) PostEditMetalSheet(c echo.Context) error {
	// Get current user for feed creation
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	// Extract tool ID from query parameters
	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return errors.Handler(err, "failed to get tool_id from query")
	}
	toolID := models.ToolID(toolIDQuery)

	// Fetch the associated tool
	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return errors.Handler(err, "failed to get tool from database")
	}

	// Parse form data into metal sheet model
	metalSheet, err := getMetalSheetFormData(c)
	if err != nil {
		return errors.Handler(err, "failed to parse metal sheet form data")
	}

	// Associate metal sheet with the tool
	metalSheet.ToolID = toolID

	// Save new metal sheet to database
	if _, err := h.registry.MetalSheets.Add(metalSheet); err != nil {
		return errors.Handler(err, "failed to create metal sheet in database")
	}

	h.createNewMetalSheetFeed(user, tool, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditMetalSheet(c echo.Context) error {
	// Get current user for feed creation
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	// Extract metal sheet ID from query parameters
	metalSheetIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "failed to get id from query")
	}
	metalSheetID := models.MetalSheetID(metalSheetIDQuery)

	// Fetch the existing metal sheet to preserve ID and tool association
	existingSheet, err := h.registry.MetalSheets.Get(metalSheetID)
	if err != nil {
		return errors.Handler(err, "failed to get existing metal sheet from database")
	}

	// Fetch the associated tool for feed creation
	tool, err := h.registry.Tools.Get(existingSheet.ToolID)
	if err != nil {
		return errors.Handler(err, "failed to get tool from database")
	}

	// Parse updated form data
	metalSheet, err := getMetalSheetFormData(c)
	if err != nil {
		return errors.Handler(err, "failed to parse metal sheet form data")
	}

	// Preserve the original ID and tool association
	metalSheet.ID = existingSheet.ID
	metalSheet.ToolID = existingSheet.ToolID

	// Update the metal sheet in database
	if err := h.registry.MetalSheets.Update(metalSheet); err != nil {
		return errors.Handler(err, "failed to update metal sheet in database")
	}

	h.createUpdateMetalSheetFeed(user, tool, existingSheet, metalSheet)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditNote(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	props := &components.DialogEditNoteProps{
		Note:         &models.Note{}, // Default empty note for creation
		LinkToTables: []string{},
		User:         user,
	}

	// Parse linked tables from query parameter
	if linkToTables := c.QueryParam("link_to_tables"); linkToTables != "" {
		props.LinkToTables = strings.Split(linkToTables, ",")
	}

	// Check if we're editing an existing note
	if idq, _ := utils.ParseQueryInt64(c, "id"); idq > 0 {
		noteID := models.NoteID(idq)

		slog.Debug("Opening edit dialog for note", "note", noteID)

		note, err := h.registry.Notes.Get(noteID)
		if err != nil {
			return errors.Handler(err, "failed to get note from database")
		}
		props.Note = note
	} else {
		slog.Debug("Opening create dialog for new note")
	}

	dialog := components.DialogEditNote(*props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "failed to render edit note dialog")
	}

	return nil
}

func (h *Handler) PostEditNote(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "failed to get user from context")
	}

	slog.Debug("Creating a new note", "user_name", user.Name)

	note, err := getNoteFromFormData(c)
	if err != nil {
		return errors.BadRequest(err, "failed to parse note form data")
	}

	// Create the note
	noteID, err := h.registry.Notes.Add(note)
	if err != nil {
		return errors.Handler(err, "failed to create note")
	}

	slog.Info("Created note", "note", noteID, "user_name", user.Name)

	// Create feed entry
	title := "Neue Notiz erstellt"
	content := fmt.Sprintf("Eine neue Notiz wurde erstellt: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed for cycle creation", "error", err)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditNote(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.Handler(err, "failed to get user from context")
	}

	idq, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "failed to parse note ID")
	}
	noteID := models.NoteID(idq)

	slog.Debug("Updating note", "note", noteID, "user_name", user.Name)

	note, err := getNoteFromFormData(c)
	if err != nil {
		return errors.BadRequest(err, "failed to parse note form data")
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	if err := h.registry.Notes.Update(note); err != nil {
		return errors.Handler(err, "failed to update note")
	}

	slog.Info("Updated note", "user_name", user.Name, "note", noteID)

	// Create feed entry
	title := "Notiz aktualisiert"
	content := fmt.Sprintf("Eine Notiz wurde aktualisiert: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed for cycle creation", "error", err)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) GetEditRegeneration(c echo.Context) error {
	rid, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.BadRequest(err, "failed to parse regeneration id")
	}
	regenerationID := models.RegenerationID(rid)

	regeneration, err := h.registry.ToolRegenerations.Get(regenerationID)
	if err != nil {
		return errors.Handler(err, "get regeneration failed")
	}

	resolvedRegeneration, err := services.ResolveRegeneration(h.registry, regeneration)
	if err != nil {
		return err
	}

	dialog := components.PageTool_DialogEditRegeneration(resolvedRegeneration)

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render dialog failed")
	}

	return nil
}

func (h *Handler) PutEditRegeneration(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return errors.BadRequest(err, "failed to get user from context")
	}

	var regenerationID models.RegenerationID
	if id, err := utils.ParseQueryInt64(c, "id"); err != nil {
		return errors.BadRequest(err, "failed to get the regeneration id from url query")
	} else {
		regenerationID = models.RegenerationID(id)
	}

	var regeneration *models.ResolvedRegeneration
	if r, err := h.registry.ToolRegenerations.Get(regenerationID); err != nil {
		return errors.Handler(err, "failed to get regeneration before deleting")
	} else {
		regeneration, err = services.ResolveRegeneration(h.registry, r)

		formData := getEditRegenerationFormData(c)
		regeneration.Reason = formData.Reason
	}

	err = h.registry.ToolRegenerations.Update(regeneration.Regeneration, user)
	if err != nil {
		return errors.Handler(err, "failed to update regeneration")
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

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) createFeed(title, content string, userID models.TelegramID) {
	feed := models.NewFeed(title, content, userID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "error", err)
	}
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
	feed := models.NewFeed("Blech erstellt", content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
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
	feed := models.NewFeed("Blech aktualisiert", content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create update feed", "error", err)
	}
}
