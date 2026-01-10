package tools

import (
	"sync"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	"github.com/labstack/echo/v4"
)

func ToolsSection(c echo.Context) *echo.HTTPError {
	return renderToolsSection(c)
}

func renderToolsSection(c echo.Context) *echo.HTTPError {
	var (
		tools              []*shared.Tool
		cassettes          []*shared.Tool
		errCh              = make(chan *echo.HTTPError, 3)
		isRegenerating     = make(map[shared.EntityID]bool)
		regenerationsCount = 0
	)

	wg := &sync.WaitGroup{}

	// Load all tools
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

	user, merr := urlb.GetUserFromContext(c)
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
