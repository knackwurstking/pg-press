package press

import (
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"

	"github.com/labstack/echo/v4"
)

func GetNotes(c echo.Context) *echo.HTTPError {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get notes directly linked to this press
	notes, merr := h.registry.Notes.ListByLinked("press", int64(press))
	if merr != nil {
		return merr.Echo()
	}

	// Get tools for this press for context
	sortedTools, _, merr := h.getOrderedToolsForPress(press) // Get active tools
	if merr != nil {
		return merr.WrapEcho("get tools for press %d", press)
	}

	// Get notes for tools
	for _, t := range sortedTools {
		n, merr := h.registry.Notes.ListByLinked("tool", int64(t.ID))
		if merr != nil {
			return merr.WrapEcho("get notes for tool %d", t.ID)
		}
		notes = append(notes, n...)
	}

	t := templates.NotesSection(press, notes, sortedTools)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NotesSection")
	}

	return nil
}
