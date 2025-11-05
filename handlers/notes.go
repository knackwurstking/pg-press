package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Notes struct {
	registry *services.Registry
}

func NewNotes(r *services.Registry) *Notes {
	return &Notes{
		registry: r,
	}
}

func (h *Notes) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Notes page
		utils.NewEchoRoute(http.MethodGet, "/notes",
			h.GetNotesPage),

		// HTMX routes for notes dialog editing
		utils.NewEchoRoute(http.MethodGet, "/htmx/notes/edit",
			h.HTMXGetEditNoteDialog),
		utils.NewEchoRoute(http.MethodPost, "/htmx/notes/edit",
			h.HTMXPostEditNoteDialog),
		utils.NewEchoRoute(http.MethodPut, "/htmx/notes/edit",
			h.HTMXPutEditNoteDialog),

		// HTMX routes for notes deletion
		utils.NewEchoRoute(http.MethodDelete, "/htmx/notes/delete",
			h.HTMXDeleteNote),

		// Render Notes Grid
		utils.NewEchoRoute(http.MethodGet, "/htmx/notes/grid",
			h.HTMXGetNotesGrid),
	})
}

// GetNotesPage serves the main notes page
func (h *Notes) GetNotesPage(c echo.Context) error {
	// Get all notes with defensive error handling
	notes, err := h.registry.Notes.List()
	if err != nil {
		slog.Error("Failed to retrieve notes from database", "error", err)
		return utils.HandleError(err, "failed to get notes from database")
	}

	// Handle case where notes might be nil
	if notes == nil {
		slog.Debug("No notes found in database, initializing empty slice")
		notes = []*models.Note{}
	}

	slog.Debug("Retrieved notes from database", "notes", len(notes))

	// Get all tools to show relationships
	tools, err := h.registry.Tools.List()
	if err != nil {
		slog.Error("Failed to retrieve tools from database", "error", err)
		return utils.HandleError(err, "failed to get tools from database")
	}

	// Handle case where tools might be nil
	if tools == nil {
		slog.Debug("No tools found in database, initializing empty slice")
		tools = []*models.Tool{}
	}

	slog.Debug("Retrieved tools from database", "tools", len(tools))

	page := components.PageNotes(&components.PageNotesProps{
		Notes: notes,
		Tools: tools,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render notes page")
	}

	return nil
}

// HTMXGetEditNoteDialog renders the edit note dialog
func (h *Notes) HTMXGetEditNoteDialog(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
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
			return utils.HandleError(err, "failed to get note from database")
		}
		props.Note = note
	} else {
		slog.Debug("Opening create dialog for new note")
	}

	dialog := components.DialogEditNote(*props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render edit note dialog")
	}

	return nil
}

// HTMXPostEditNoteDialog creates a new note
func (h *Notes) HTMXPostEditNoteDialog(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleError(err, "failed to get user from context")
	}

	slog.Debug("Creating a new note", "user_name", user.Name)

	note, err := h.parseNoteFromForm(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse note form data")
	}

	// Create the note
	noteID, err := h.registry.Notes.Add(note)
	if err != nil {
		return utils.HandleError(err, "failed to create note")
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

// HTMXPutEditNoteDialog updates an existing note
func (h *Notes) HTMXPutEditNoteDialog(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleError(err, "failed to get user from context")
	}

	idq, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse note ID")
	}
	noteID := models.NoteID(idq)

	slog.Debug("Updating note", "note", noteID, "user_name", user.Name)

	note, err := h.parseNoteFromForm(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse note form data")
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	if err := h.registry.Notes.Update(note); err != nil {
		return utils.HandleError(err, "failed to update note")
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

// HTMXDeleteNote deletes a note and unlinks it from all tools
func (h *Notes) HTMXDeleteNote(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleError(err, "failed to get user from context")
	}

	idq, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse note ID")
	}
	noteID := models.NoteID(idq)

	// Delete the note
	if err := h.registry.Notes.Delete(noteID); err != nil {
		return utils.HandleError(err, "failed to delete note")
	}

	slog.Info("Deleted note", "note", noteID, "user_name", user.Name)

	// Create feed entry
	feed := models.NewFeed("Notiz gelöscht", "Eine Notiz wurde gelöscht", user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed for note deletion", "error", err)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Notes) HTMXGetNotesGrid(c echo.Context) error {
	notes, err := h.registry.Notes.List()
	if err != nil {
		return utils.HandleError(err, "failed to list notes")
	}

	tools, err := h.registry.Tools.List()
	if err != nil {
		return utils.HandleError(err, "failed to list tools")
	}

	ng := components.PageNotes_NotesGrid(notes, tools)
	if err := ng.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render notes grid")
	}
	return nil
}

// parseNoteFromForm parses note data from form submission
func (h *Notes) parseNoteFromForm(c echo.Context) (note *models.Note, err error) {
	note = &models.Note{}

	// Parse level
	levelStr := c.FormValue("level")
	if levelStr == "" {
		return nil, fmt.Errorf("level is required")
	}

	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		return nil, fmt.Errorf("invalid level format: %v", err)
	}

	// Validate level is within valid range (0=INFO, 1=ATTENTION, 2=BROKEN)
	if levelInt < 0 || levelInt > 2 {
		return nil, fmt.Errorf("invalid level value: %d (must be 0, 1, or 2)", levelInt)
	}

	note.Level = models.Level(levelInt)

	// Parse content
	note.Content = strings.TrimSpace(c.FormValue("content"))
	if note.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Handle linked field - get first linked_tables value or empty string
	linkedTables := c.Request().Form["linked_tables"]
	if len(linkedTables) > 0 {
		note.Linked = linkedTables[0]
	}

	return note, nil
}
