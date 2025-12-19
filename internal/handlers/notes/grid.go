package notes

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/notes/templates"
	"github.com/labstack/echo/v4"
)

func NotesGrid(c echo.Context) error {
	notes, merr := h.registry.Notes.List()
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	ng := templates.NotesGrid(notes, tools)
	err := ng.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NotesGrid")
	}
	return nil
}
