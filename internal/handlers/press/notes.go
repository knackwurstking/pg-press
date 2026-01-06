package press

import (
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func GetNotes(c echo.Context) *echo.HTTPError {
	var pressNumber shared.PressNumber
	if press, merr := shared.ParseParamInt8(c, "press"); merr != nil {
		return merr.Echo()
	} else {
		pressNumber = shared.PressNumber(press)
	}

	pressNotes, merr := db.ListNotesForLinked("press", int(pressNumber))
	if merr != nil {
		return merr.Echo()
	}

	pu, merr := db.GetPressUtilizations(pressNumber)
	if merr != nil {
		return merr.Echo()
	}
	u := pu[pressNumber]

	toolsMap := make(map[shared.EntityID]*shared.Tool)
	toolsMap[u.SlotUpper.ID] = u.SlotUpper
	toolsMap[u.SlotUpperCassette.ID] = u.SlotUpperCassette
	toolsMap[u.SlotLower.ID] = u.SlotLower

	var toolNotes []*shared.Note
	for id := range toolsMap {
		notes, merr := db.ListNotesForLinked("tool", int(id))
		if merr != nil {
			return merr.Echo()
		}
		toolNotes = append(toolNotes, notes...)
	}

	t := templates.Notes(templates.NotesProps{
		PressNumber: pressNumber,
		PressNotes:  pressNotes,
		ToolNotes:   toolNotes,
		Tools:       toolsMap,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Notes")
	}

	return nil
}
