package notes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/features/notes/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(
			db,
			logger.NewComponentLogger("Notes"),
		),
	}
}

// HTMXGetEditNoteDialog renders the edit note dialog
func (h *Handler) HTMXGetEditNoteDialog(c echo.Context) error {
	props := &templates.DialogEditNoteProps{
		Note:         &models.Note{}, // Default empty note for creation
		LinkToTables: []string{},
	}

	// Parse linked tables from query parameter
	if linkToTables := c.QueryParam("link_to_tables"); linkToTables != "" {
		props.LinkToTables = strings.Split(linkToTables, ",")
	}

	// Check if we're editing an existing note
	if noteID, _ := h.ParseInt64Query(c, "id"); noteID > 0 {
		h.LogDebug("Opening edit dialog for note %d", noteID)

		note, err := h.DB.Notes.Get(noteID)
		if err != nil {
			return h.HandleError(c, err, "failed to get note from database")
		}
		props.Note = note
	} else {
		h.LogDebug("Opening create dialog for new note")
	}

	dialog := templates.DialogEditNote(*props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render edit note dialog: "+err.Error())
	}

	return nil
}

// HTMXPostEditNoteDialog creates a new note
func (h *Handler) HTMXPostEditNoteDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("User %s creating new note", user.Name)

	note, err := h.parseNoteFromForm(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note form data: "+err.Error())
	}

	// Create the note
	noteID, err := h.DB.Notes.Add(note)
	if err != nil {
		return h.HandleError(c, err, "failed to create note")
	}

	h.LogInfo("User %s created note %d", user.Name, noteID)

	// Close the dialog and refresh the target
	c.Response().Header().Set("HX-Trigger", "pageLoaded")
	c.Response().WriteHeader(200)
	return nil
}

// HTMXPutEditNoteDialog updates an existing note
func (h *Handler) HTMXPutEditNoteDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	noteID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note ID: "+err.Error())
	}

	h.LogDebug("User %s updating note %d", user.Name, noteID)

	note, err := h.parseNoteFromForm(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse note form data: "+err.Error())
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	if err := h.DB.Notes.Update(note); err != nil {
		return h.HandleError(c, err, "failed to update note")
	}

	h.LogInfo("User %s updated note %d", user.Name, noteID)

	// Close the dialog and refresh the target
	c.Response().Header().Set("HX-Trigger", "pageLoaded")
	c.Response().WriteHeader(200)
	return nil
}

// parseNoteFromForm parses note data from form submission
func (h *Handler) parseNoteFromForm(c echo.Context) (*models.Note, error) {
	note := &models.Note{}

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

	// TODO: Handle linked_tables when the feature is implemented
	// linkedTables := c.Request().Form["linked_tables"]

	return note, nil
}
