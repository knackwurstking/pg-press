package notes

import (
	"log/slog"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func DeleteNote(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	slog.Debug("Deleting note", "id", id, "user_name", c.Get("user_name"))

	// Delete the note
	merr = db.DeleteNote(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, "reload-notes")

	return nil
}
