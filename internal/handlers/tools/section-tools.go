package tools

import (
	"slices"
	"strings"
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
	wg := &sync.WaitGroup{}
	errCh := make(chan *echo.HTTPError, 4)

	// All Tools
	tools := []*shared.Tool{}
	cassettes := []*shared.Tool{}
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

		// Sort tools and cassettes alphabetically
		slices.SortFunc(tools, func(a, b *shared.Tool) int {
			return strings.Compare(a.German(), b.German())
		})
		slices.SortFunc(cassettes, func(a, b *shared.Tool) int {
			return strings.Compare(a.German(), b.German())
		})

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
	isRegenerating := make(map[shared.EntityID]bool)
	regenerationsCount := make(map[shared.EntityID]int)
	wg.Go(func() {
		regenerations, merr := db.ListToolRegenerations()
		if merr != nil {
			errCh <- merr.Echo()
			return
		}

		rmap := map[shared.EntityID][]*shared.ToolRegeneration{}
		for _, r := range regenerations {
			if r.Stop == 0 {
				isRegenerating[r.ToolID] = true
			}
			if _, ok := rmap[r.ToolID]; !ok {
				rmap[r.ToolID] = []*shared.ToolRegeneration{}
			}
			rmap[r.ToolID] = append(rmap[r.ToolID], r)
		}
		for toolID, regs := range rmap {
			regenerationsCount[toolID] = len(regs)
		}

		errCh <- nil
	})

	// Notes Count
	notesCount := map[shared.EntityID]int{}
	wg.Go(func() {
		notes, merr := db.ListNotes()
		if merr != nil {
			errCh <- merr.Echo()
			return
		}

		for _, n := range notes {
			l := n.GetLinked()
			if l.Name != "tool" {
				continue
			}

			toolID := shared.EntityID(l.ID)
			if _, ok := notesCount[toolID]; ok {
				notesCount[toolID] = 0
			}
			notesCount[toolID]++
		}

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
