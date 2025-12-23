package notes

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"

	"github.com/labstack/echo/v4"
)

func GetNotesGrid(c echo.Context) *echo.HTTPError {
	notes, merr := db.ListNotes()
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := db.ListTools()
	if merr != nil {
		return merr.Echo()
	}

	ng := NotesGrid(notes, tools)
	err := ng.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NotesGrid")
	}
	return nil
}
