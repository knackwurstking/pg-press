package tool

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/labstack/echo/v4"
)

func GetToolNotes(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)
	notes, merr := db.ListNotesForLinked("tool", toolID)
	if merr != nil {
		return merr.Echo()
	}

	t := Notes(toolID, notes)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes")
	}
	return nil
}
