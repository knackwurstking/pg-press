package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Notes struct {
	*Base
}

func NewNotes(db *services.Registry) *Notes {
	return &Notes{
		Base: NewBase(db, logger.NewComponentLogger("Notes")),
	}
}

func (h *Notes) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
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

			utils.NewEchoRoute(http.MethodDelete, "/htmx/notes/delete",
				h.HTMXDeleteNote),
		},
	)
}

// GetNotesPage serves the main notes page
func (h *Notes) GetNotesPage(c echo.Context) error {
	// Get all notes with defensive error handling
	notes, err := h.Registry.Notes.List()
	if err != nil {
		h.Log.Error("Failed to retrieve notes from database: %v", err)
		return HandleError(err, "failed to get notes from database")
	}

	// Handle case where notes might be nil
	if notes == nil {
		h.Log.Debug("No notes found in database, initializing empty slice")
		notes = []*models.Note{}
	}

	h.Log.Debug("Retrieved %d notes from database", len(notes))

	// Get all tools to show relationships
	tools, err := h.Registry.Tools.List()
	if err != nil {
		h.Log.Error("Failed to retrieve tools from database: %v", err)
		return HandleError(err, "failed to get tools from database")
	}

	// Handle case where tools might be nil
	if tools == nil {
		h.Log.Debug("No tools found in database, initializing empty slice")
		tools = []*models.Tool{}
	}

	h.Log.Debug("Retrieved %d tools from database", len(tools))

	page := components.PageNotes(&components.PageNotesProps{
		Notes: notes,
		Tools: tools,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render notes page")
	}

	return nil
}

// HTMXGetEditNoteDialog renders the edit note dialog
func (h *Notes) HTMXGetEditNoteDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
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
	if noteID, _ := ParseQueryInt64(c, "id"); noteID > 0 {
		h.Log.Debug("Opening edit dialog for note %d", noteID)

		note, err := h.Registry.Notes.Get(noteID)
		if err != nil {
			return HandleError(err, "failed to get note from database")
		}
		props.Note = note
	} else {
		h.Log.Debug("Opening create dialog for new note")
	}

	dialog := components.DialogEditNote(*props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render edit note dialog")
	}

	return nil
}

// HTMXPostEditNoteDialog creates a new note
func (h *Notes) HTMXPostEditNoteDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	h.Log.Debug("User %s creating new note", user.Name)

	note, err := h.parseNoteFromForm(c)
	if err != nil {
		return HandleBadRequest(err, "failed to parse note form data")
	}

	// Create the note
	noteID, err := h.Registry.Notes.Add(note)
	if err != nil {
		return HandleError(err, "failed to create note")
	}

	h.Log.Info("User %s created note %d", user.Name, noteID)

	// Create feed entry
	title := "Neue Notiz erstellt"
	content := fmt.Sprintf("Eine neue Notiz wurde erstellt: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for cycle creation: %v", err)
	}

	c.Response().Header().Set("HX-Trigger", "pageLoaded")

	return nil
}

// HTMXPutEditNoteDialog updates an existing note
func (h *Notes) HTMXPutEditNoteDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	noteID, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse note ID")
	}

	h.Log.Debug("User %s updating note %d", user.Name, noteID)

	note, err := h.parseNoteFromForm(c)
	if err != nil {
		return HandleBadRequest(err, "failed to parse note form data")
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	if err := h.Registry.Notes.Update(note); err != nil {
		return HandleError(err, "failed to update note")
	}

	h.Log.Info("User %s updated note %d", user.Name, noteID)

	// Create feed entry
	title := "Notiz aktualisiert"
	content := fmt.Sprintf("Eine Notiz wurde aktualisiert: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for cycle creation: %v", err)
	}

	// Trigger reload of notes sections
	c.Response().Header().Set("HX-Trigger", "notesReload")

	return nil
}

// HTMXDeleteNote deletes a note and unlinks it from all tools
func (h *Notes) HTMXDeleteNote(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	noteID, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse note ID")
	}

	// Delete the note
	if err := h.Registry.Notes.Delete(noteID); err != nil {
		return HandleError(err, "failed to delete note")
	}

	h.Log.Info("User %s deleted note %d", user.Name, noteID)

	// Create feed entry
	feed := models.NewFeed("Notiz gelöscht", "Eine Notiz wurde gelöscht", user.TelegramID)
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for note deletion: %v", err)
	}

	// Trigger reload of notes sections
	c.Response().Header().Set("HX-Trigger", "pageLoaded")

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
