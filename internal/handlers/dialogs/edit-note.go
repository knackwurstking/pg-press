package dialogs

import (
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetEditNote(c echo.Context) *echo.HTTPError {
	linked := c.QueryParam("linked")

	// Check if we're editing an existing note
	var note *shared.Note
	if id, _ := shared.ParseQueryInt64(c, "id"); id > 0 {
		noteID := shared.EntityID(id)

		var merr *errors.HTTPError
		note, merr = db.GetNote(noteID)
		if merr != nil {
			return merr.Echo()
		}
	}

	user, _ := shared.GetUserFromContext(c)

	if note != nil {
		log.Debug("Rendering edit note dialog [note=%v, linked=%v, user_name=%s]", note, linked, c.Get("user-name"))
		t := templates.EditNoteDialog(note, linked, user)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditNoteDialog")
		}
		return nil
	}

	log.Debug("Rendering new note dialog [linked=%v, user_name=%s]", linked, c.Get("user-name"))
	t := templates.NewNoteDialog(linked, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewNoteDialog")
	}
	return nil
}

func PostNote(c echo.Context) *echo.HTTPError {
	note, merr := parseNoteForm(c)
	if merr != nil {
		return merr.WrapEcho("failed to get note form data")
	}

	log.Debug("Creating new note [note=%v, user_name=%s]", note, c.Get("user-name"))

	merr = db.AddNote(note)
	if merr != nil {
		return merr.WrapEcho("failed to create note")
	}

	urlb.SetHXTrigger(c, "reload-notes")

	return nil
}

func PutNote(c echo.Context) *echo.HTTPError {
	note, merr := parseNoteForm(c)
	if merr != nil {
		return merr.WrapEcho("failed to get note form data")
	}

	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	note.ID = shared.EntityID(id)

	log.Debug("Updating note [note=%v, user_name=%s]", note, c.Get("user-name"))

	// Update the note
	merr = db.UpdateNote(note)
	if merr != nil {
		return merr.WrapEcho("failed to update note")
	}

	// Trigger reload of notes sections
	urlb.SetHXTrigger(c, "reload-notes")

	return nil
}

func parseNoteForm(c echo.Context) (*shared.Note, *errors.HTTPError) {
	// Parse level (required)
	levelStr := c.FormValue("level")
	if levelStr == "" {
		return nil, errors.NewValidationError("level is required").HTTPError()
	}

	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		return nil, errors.NewValidationError("level must be an integer").HTTPError()
	}

	// Validate level is within valid range (0=INFO, 1=ATTENTION, 2=BROKEN)
	level := shared.NoteLevel(levelInt)
	if !level.IsValid() {
		return nil, errors.NewValidationError("level is invalid").HTTPError()
	}

	// Parse content (required)
	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		return nil, errors.NewValidationError("content is required").HTTPError()
	}

	return &shared.Note{
		Level:     level,
		Content:   content,
		CreatedAt: shared.NewUnixMilli(time.Now()),
		Linked:    c.FormValue("linked"),
	}, nil
}
