package tools

import (
	"sync"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func ToolsSection(c echo.Context) *echo.HTTPError {
	return renderToolsSection(c)
}

func renderToolsSection(c echo.Context) *echo.HTTPError {
	var (
		tools              []*shared.Tool
		cassettes          []*shared.Tool
		errCh              = make(chan *echo.HTTPError, 4)
		isRegenerating     = make(map[shared.EntityID]bool)
		regenerationsCount = 0
	)

	wg := &sync.WaitGroup{}

	// All Tools
	wg.Go(func() {
		allTools, merr := db.ListTools()
		if merr != nil {
			errCh <- merr.Echo()
			return
		}
		// Filter tools into cassettes and non-cassettes
		for _, t := range allTools {
			if t.IsCassette() {
				cassettes = append(cassettes, t)
			} else {
				tools = append(tools, t)
			}
		}
		errCh <- nil
	})

	// Active Tools
	activeTools := make(map[shared.EntityID]shared.PressNumber)
	wg.Go(func() {
		for _, press := range shared.AllPressNumbers {
			p, herr := db.GetPress(press)
			if herr != nil {
				errCh <- herr.Echo()
				return
			}
			if p.SlotUp > 0 {
				activeTools[p.SlotUp] = press
			}
			if p.SlotDown > 0 {
				activeTools[p.SlotDown] = press
			}
		}

		errCh <- nil
	})

	// Tool Regenerations
	wg.Go(func() {
		regenerations, merr := db.ListToolRegenerations()
		if merr != nil {
			errCh <- merr.Echo()
			return
		}
		for _, r := range regenerations {
			if r.Stop == 0 {
				isRegenerating[r.ToolID] = true
			}
		}
		regenerationsCount = len(regenerations)
		errCh <- nil
	})

	// Notes Count
	notesCount := 0
	wg.Go(func() {
		notes, merr := db.ListNotes()
		if merr != nil {
			errCh <- merr.Echo()
			return
		}
		notesCount = len(notes)
		errCh <- nil
	})

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		if e != nil {
			return e
		}
	}

	t := templates.SectionTools(templates.SectionToolsProps{
		Tools:              tools,
		Cassettes:          cassettes,
		ActiveTools:        activeTools,
		User:               user,
		IsRegenerating:     isRegenerating,
		RegenerationsCount: regenerationsCount,
		NotesCount:         notesCount,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionTools")
	}

	return nil
}
