package notes

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/services/helper"

	"github.com/labstack/echo/v4"
)

func GetNotesGrid(c echo.Context) *echo.HTTPError {
	notes, merr := db.Notes.List()
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := helper.ListTools(db)
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
