package tools

import (
	"sync"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/labstack/echo/v4"
)

func ToolsSection(c echo.Context) *echo.HTTPError {
	return renderToolsSection(c)
}

func renderToolsSection(c echo.Context) *echo.HTTPError {
	var (
		tools     []*shared.Tool
		cassettes []*shared.Cassette
		errCh     = make(chan *echo.HTTPError, 2)
	)

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		var merr *errors.MasterError
		tools, merr = DB.Tool.Tool.List()
		if merr != nil {
			errCh <- merr.Echo()
		}
		errCh <- nil
	})

	wg.Go(func() {
		var merr *errors.MasterError
		cassettes, merr = DB.Tool.Cassette.List()
		if merr != nil {
			errCh <- merr.Echo()
		}
		errCh <- nil
	})

	isRegenerating := make(map[shared.EntityID]bool)
	regenerationsCount := 0
	wg.Go(func() {
		regenerations, merr := DB.Tool.Regeneration.List()
		if merr != nil {
			errCh <- merr.Echo()
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
		notes, merr := DB.Note.Note.List()
		if merr != nil {
			errCh <- merr.Echo()
		}
		notesCount = len(notes)
		errCh <- nil
	})

	user, merr := shared.GetUserFromContext(c)
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

	t := SectionTools(SectionToolsProps{
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
