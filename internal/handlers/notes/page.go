package notes

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

// GetNotesPage serves the main notes page
func GetPage(c echo.Context) *echo.HTTPError {
	// Get all notes with defensive error handling
	notes, merr := db.Notes.List()
	if merr != nil {
		return merr.Echo()
	}

	// Handle case where notes might be nil
	if notes == nil {
		notes = []*shared.Note{}
	}

	// Get all tools to show relationships
	tools, merr := helper.ListTools(db)
	if merr != nil {
		return merr.Echo()
	}

	t := Page(notes, tools)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes Page")
	}

	return nil
}
