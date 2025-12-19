package dialogs

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func GetEditNote(c echo.Context) *echo.HTTPError {
	var linkToTables []string
	var note *shared.Note
	var merr *errors.MasterError

	// Parse linked tables from query parameter
	if ltt := c.QueryParam("link_to_tables"); ltt != "" {
		linkToTables = strings.Split(ltt, ",")
	}

	// Check if we're editing an existing note
	if id, _ := shared.ParseQueryInt64(c, "id"); id > 0 {
		noteID := shared.EntityID(id)

		note, merr = db.Notes.GetByID(noteID)
		if merr != nil {
			return merr.Echo()
		}
	}

	user, _ := shared.GetUserFromContext(c)

	if note != nil {
		log.Debug("Rendering edit note dialog [note=%s] [linkToTables=%v] [user=\"%s\"]", note.String(), linkToTables, c.Get("user-name"))
		t := EditNoteDialog(note, linkToTables, user)
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditNoteDialog")
		}
		return nil
	}

	log.Debug("Rendering new note dialog [linkToTables=%v] [user=\"%s\"]", linkToTables, c.Get("user-name"))
	t := NewNoteDialog(linkToTables, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewNoteDialog")
	}
	return nil
}

func PostEditNote(c echo.Context) *echo.HTTPError {
	note, merr := GetNoteFormData(c)
	if merr != nil {
		return merr.WrapEcho("failed to get note form data")
	}

	log.Debug("Creating new note [note=%s] [user=\"%s\"]", note.String(), c.Get("user-name"))

	merr = db.Notes.Create(note)
	if merr != nil {
		return merr.WrapEcho("failed to create note")
	}

	urlb.SetHXTrigger(c, "reload-notes")

	return nil
}

func PutEditNote(c echo.Context) *echo.HTTPError {
	note, merr := GetNoteFormData(c)
	if merr != nil {
		return merr.WrapEcho("failed to get note form data")
	}

	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	note.ID = shared.EntityID(id)

	log.Debug("Updating note [note=%s] [user=\"%s\"]", note.String(), c.Get("user-name"))

	// Update the note
	merr = db.Notes.Update(note)
	if merr != nil {
		return merr.WrapEcho("failed to update note")
	}

	// Trigger reload of notes sections
	urlb.SetHXTrigger(c, "reload-notes")

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
