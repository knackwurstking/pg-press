package dialogs

import (
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditNote(c echo.Context) *echo.HTTPError {
	linked := c.QueryParam("linked")

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr.Echo()
	}

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

	if note != nil {
		log.Debug("Rendering edit note dialog [note=%v, linked=%v, user_name=%s]", note, linked, c.Get("user-name"))
		t := EditNoteDialog(note.ID, NoteDialogProps{
			NoteFormData: NoteFormData{
				Level:   note.Level,
				Content: note.Content,
				Linked:  note.Linked,
			},
			User: user,
			Open: true,
			OOB:  true,
		})
		if err := t.Render(c.Request().Context(), c.Response()); err != nil {
			return errors.NewRenderError(err, "EditNoteDialog")
		}
		return nil
	}

	log.Debug("Rendering new note dialog [linked=%v, user_name=%s]", linked, c.Get("user-name"))

	t := NewNoteDialog(NoteDialogProps{
		User: user,
		Open: true,
		OOB:  true,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewNoteDialog")
	}
	return nil
}

func PostNote(c echo.Context) *echo.HTTPError {
	id, _ := utils.GetQueryInt64(c, "id")
	if id > 0 {
		return updateNote(c, shared.EntityID(id))
	}

	data, ierrs := parseNoteForm(c)
	if len(ierrs) > 0 {
		return reRenderNewNoteDialog(c, true, data, ierrs...)
	}

	log.Debug("Creating new note [data=%#v, user_name=%s]", data, c.Get("user-name"))

	note := &shared.Note{
		Level:     data.Level,
		Content:   data.Content,
		CreatedAt: shared.NewUnixMilli(time.Now()),
		Linked:    data.Linked,
	}
	if herr := db.AddNote(note); herr != nil {
		ierr := errors.NewInputError("", herr.Error())
		return reRenderNewNoteDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-notes")

	return reRenderNewNoteDialog(c, false, NoteFormData{Linked: data.Linked})
}

func updateNote(c echo.Context, id shared.EntityID) *echo.HTTPError {
	// Parse form data
	data, ierrs := parseNoteForm(c)
	if len(ierrs) > 0 {
		return reRenderEditNoteDialog(c, id, true, data, ierrs...)
	}

	// Get note for id
	note, herr := db.GetNote(id)
	if herr != nil {
		ierr := errors.NewInputError("", "invalid note id")
		return reRenderEditNoteDialog(c, id, true, data, ierr)
	}
	note.Level = data.Level
	note.Content = data.Content
	note.Linked = data.Linked

	log.Debug("Updating note [data=%#v, user_name=%s]", data, c.Get("user-name"))

	// Update the note
	if herr := db.UpdateNote(note); herr != nil {
		ierr := errors.NewInputError("", herr.Error())
		return reRenderEditNoteDialog(c, id, true, data, ierr)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, "reload-notes")

	return reRenderEditNoteDialog(c, id, false, data)
}

// parseNoteForm will return only general input errors (inputID set to "")
func parseNoteForm(c echo.Context) (data NoteFormData, ierrs []*errors.InputError) {
	// Parse level (required)
	level, err := utils.SanitizeInt(c.FormValue("level"))
	data.Level = shared.NoteLevel(level)
	if err != nil {
		ierr := errors.NewInputError("", "level must be an integer")
		ierrs = append(ierrs, ierr)
	}

	// Parse content (required)
	data.Content = utils.SanitizeText(c.FormValue("content"))
	if data.Content == "" {
		ierr := errors.NewInputError("", "content is required")
		ierrs = append(ierrs, ierr)
	}

	return
}

func reRenderNewNoteDialog(c echo.Context, open bool, data NoteFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	u, e := utils.GetUserFromContext(c)
	if e != nil {
		return e.Echo()
	}

	t := NewNoteDialog(NoteDialogProps{
		NoteFormData: data,
		User:         u,
		Open:         open,
		OOB:          true,
		Error:        ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "NewNoteDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "user input error")
	}
	return nil
}

func reRenderEditNoteDialog(c echo.Context, noteID shared.EntityID, open bool, data NoteFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	u, e := utils.GetUserFromContext(c)
	if e != nil {
		return e.Echo()
	}

	t := EditNoteDialog(noteID, NoteDialogProps{
		NoteFormData: data,
		User:         u,
		Open:         open,
		OOB:          true,
		Error:        ierrs,
	})
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "EditNoteDialog")
	}
	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "user input error")
	}
	return nil
}
