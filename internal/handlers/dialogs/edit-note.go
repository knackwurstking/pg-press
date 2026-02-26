package dialogs

import (
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditNote(c echo.Context) *echo.HTTPError {
	linked := c.QueryParam("linked")

	// Check if we're editing an existing note
	var note *shared.Note
	if id, _ := utils.GetQueryInt64(c, "id"); id > 0 {
		noteID := shared.EntityID(id)

		var merr *errors.HTTPError
		note, merr = db.GetNote(noteID)
		if merr != nil {
			return merr.Echo()
		}
	}

	user, _ := utils.GetUserFromContext(c)

	if note != nil {
		log.Debug("Rendering edit note dialog [note=%v, linked=%v, user_name=%s]", note, linked, c.Get("user-name"))
		t := EditNoteDialog(note, linked, user)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditNoteDialog")
		}
		return nil
	}

	log.Debug("Rendering new note dialog [linked=%v, user_name=%s]", linked, c.Get("user-name"))
	t := NewNoteDialog(linked, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewNoteDialog")
	}
	return nil
}

func PostNote(c echo.Context) *echo.HTTPError {
	note, verr := parseNoteForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	log.Debug("Creating new note [note=%v, user_name=%s]", note, c.Get("user-name"))

	if merr := db.AddNote(note); merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-notes")

	return nil
}

func PutNote(c echo.Context) *echo.HTTPError {
	note, verr := parseNoteForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	id, merr := utils.GetQueryInt64(c, "id")
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
	utils.SetHXTrigger(c, "reload-notes")

	return nil
}

func parseNoteForm(c echo.Context) (*shared.Note, *errors.ValidationError) {
	// Parse level (required)
	vLevelString := utils.SanitizeText(c.FormValue("level"))
	if vLevelString == "" {
		return nil, errors.NewValidationError("level is required")
	}
	vLevelInt, err := strconv.Atoi(vLevelString)
	if err != nil {
		return nil, errors.NewValidationError("level must be an integer")
	}
	level := shared.NoteLevel(vLevelInt)
	if !level.IsValid() {
		return nil, errors.NewValidationError("level is invalid")
	}

	// Parse content (required)
	content := utils.SanitizeText(c.FormValue("content"))
	if content == "" {
		return nil, errors.NewValidationError("content is required")
	}

	return &shared.Note{
		Level:     level,
		Content:   content,
		CreatedAt: shared.NewUnixMilli(time.Now()),
		Linked:    utils.SanitizeText(c.FormValue("linked")),
	}, nil
}
