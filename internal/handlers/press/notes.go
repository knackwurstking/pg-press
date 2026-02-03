package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetNotes(c echo.Context) *echo.HTTPError {
	id, merr := utils.GetParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}
	pressID := shared.EntityID(id)

	pressNotes, merr := db.ListNotesForLinked("press", int(pressID))
	if merr != nil {
		return merr.Echo()
	}

	u, merr := db.GetPressUtilization(pressID)
	if merr != nil {
		return merr.Echo()
	}

	toolsMap := make(map[shared.EntityID]*shared.Tool)
	if u.SlotUpper != nil {
		toolsMap[u.SlotUpper.ID] = u.SlotUpper
	}
	if u.SlotUpperCassette != nil {
		toolsMap[u.SlotUpperCassette.ID] = u.SlotUpperCassette
	}
	if u.SlotLower != nil {
		toolsMap[u.SlotLower.ID] = u.SlotLower
	}

	var toolNotes []*shared.Note
	for id := range toolsMap {
		notes, merr := db.ListNotesForLinked("tool", int(id))
		if merr != nil {
			return merr.Echo()
		}
		toolNotes = append(toolNotes, notes...)
	}

	t := templates.Notes(templates.NotesProps{
		PressNotes: pressNotes,
		ToolNotes:  toolNotes,
		Tools:      toolsMap,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes")
	}

	return nil
}
