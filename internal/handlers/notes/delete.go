package notes

import (
	"log/slog"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/models"
	"github.com/labstack/echo/v4"
)

func DeleteNote(c echo.Context) *echo.HTTPError {
	slog.Info("Deleting note")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	idq, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	noteID := models.NoteID(idq)

	// Delete the note
	merr = h.registry.Notes.Delete(noteID)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	merr = h.registry.Feeds.Add("Notiz gelöscht", "Eine Notiz wurde gelöscht", user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for note deletion", "error", merr)
	}

	// Trigger reload of notes sections
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}
