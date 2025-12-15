package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetEditNote(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	var (
		linkToTables []string
		note         *shared.Note
	)

	// Parse linked tables from query parameter
	if ltt := c.QueryParam("link_to_tables"); ltt != "" {
		linkToTables = strings.Split(ltt, ",")
	}

	// Check if we're editing an existing note
	if id, _ := shared.ParseQueryInt64(c, "id"); id > 0 {
		noteID := shared.EntityID(id)

		note, merr = h.db.Note.Note.Get(noteID)
		if merr != nil {
			return merr.Echo()
		}
	}

	var (
		t     templ.Component
		tName string
	)
	if note != nil {
		t = templates.EditNoteDialog(note, linkToTables, user)
		tName = "EditNoteDialog"
	} else {
		t = templates.NewNoteDialog(linkToTables, user)
		tName = "NewNoteDialog"
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func (h *Handler) PostEditNote(c echo.Context) *echo.HTTPError {
	slog.Info("Creating new note")

	note, merr := GetNoteFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	urlb.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditNote(c echo.Context) *echo.HTTPError {
	slog.Info("Updating note")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	idq, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	noteID := models.NoteID(idq)

	note, merr := GetNoteFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	// Set the ID for update
	note.ID = noteID

	// Update the note
	merr = h.registry.Notes.Update(note)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	title := "Notiz aktualisiert"
	content := fmt.Sprintf("Eine Notiz wurde aktualisiert: %s", note.Content)

	// Add linked info if any
	if note.Linked != "" {
		content += fmt.Sprintf("\nVerknüpft mit: %s", note.Linked)
	}

	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", merr)
	}

	// Trigger reload of notes sections
	urlb.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func GetNoteFormData(c echo.Context) (*shared.Note, *errors.MasterError) {
	note := &shared.Note{}

	// Parse level (required)
	levelStr := c.FormValue("level")
	if levelStr == "" {
		return nil, errors.NewMasterError(
			fmt.Errorf("level is required"),
			http.StatusBadRequest,
		)
	}

	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		return nil, errors.NewMasterError(err, http.StatusBadRequest)
	}

	// Validate level is within valid range (0=INFO, 1=ATTENTION, 2=BROKEN)
	if levelInt < 0 || levelInt > 2 {
		return nil, errors.NewMasterError(
			fmt.Errorf("invalid level value: %d (must be 0, 1, or 2)", levelInt),
			http.StatusBadRequest,
		)
	}

	note.Level = shared.NoteLevel(levelInt)

	// Parse content (required)
	note.Content = strings.TrimSpace(c.FormValue("content"))
	if note.Content == "" {
		return nil, errors.NewMasterError(
			fmt.Errorf("content is required"),
			http.StatusBadRequest,
		)
	}

	// Handle linked field - get first linked_tables value or empty string
	linkedTables := c.Request().Form["linked_tables"]
	if len(linkedTables) > 0 {
		note.Linked = linkedTables[0]
	}

	return note, nil
}
