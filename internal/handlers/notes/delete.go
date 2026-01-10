package notes

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func DeleteNote(c echo.Context) *echo.HTTPError {
	id, merr := urlb.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	log.Debug("Deleting note with ID %d [user_name=%s]", id, c.Get("user_name"))

	// Delete the note
	merr = db.DeleteNote(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	// Trigger reload of notes sections
	urlb.SetHXTrigger(c, "reload-notes")

	return nil
}
