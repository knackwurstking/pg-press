package tools

import (
	"net/http"
	"sync"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/logger"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB  *common.DB
	Log *ui.Logger
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		DB:  db,
		Log: logger.New("handler: tools"),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, h.GetToolsPage),
		ui.NewEchoRoute(http.MethodDelete, path+"/delete", h.Delete),
		ui.NewEchoRoute(http.MethodPatch, path+"/mark-dead", h.MarkAsDead),
		ui.NewEchoRoute(http.MethodGet, path+"/section/press", h.PressSection),
		ui.NewEchoRoute(http.MethodGet, path+"/section/tools", h.ToolsSection),
		ui.NewEchoRoute(http.MethodGet, path+"/admin/overlapping-tools", h.AdminTools),
	})
}

func (h *Handler) GetToolsPage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(templates.PageProps{User: user})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tools Page")
	}

	return nil
}

// Delete deletes a tool or a cassette if "is_cassette" query parameter is set to true.
func (h *Handler) Delete(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	isCassette := shared.ParseQueryBool(c, "is_cassette")

	if isCassette {
		merr = h.DB.Tool.Cassette.Delete(toolID)
		if merr != nil {
			return merr.Echo()
		}
		h.Log.Debug("Deleted cassette with ID: %#v", toolID)
	} else {
		merr = h.DB.Tool.Tool.Delete(toolID)
		if merr != nil {
			return merr.Echo()
		}
		h.Log.Debug("Deleted tool with ID: %#v", toolID)
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

// MarkAsDead marks a tools, or cassette if "is_cassette" query parameter is set to true, as dead.
func (h *Handler) MarkAsDead(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	if shared.ParseQueryBool(c, "is_cassette") {
		cassette, merr := h.DB.Tool.Cassette.GetByID(toolID)
		if merr != nil {
			return merr.Echo()
		}

		if cassette.IsDead {
			return nil
		}
		cassette.IsDead = true

		merr = h.DB.Tool.Cassette.Update(cassette)
		if merr != nil {
			return merr.Echo()
		}
	} else {
		tool, merr := h.DB.Tool.Tool.GetByID(toolID)
		if merr != nil {
			return merr.Echo()
		}

		if tool.IsDead {
			return nil
		}
		tool.IsDead = true

		merr = h.DB.Tool.Tool.Update(tool)
		if merr != nil {
			return merr.Echo()
		}
	}

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

func (h *Handler) PressSection(c echo.Context) *echo.HTTPError {
	return h.renderPressSection(c)
}

func (h *Handler) ToolsSection(c echo.Context) *echo.HTTPError {
	return h.renderToolsSection(c)
}

// TODO: Fix all other stuff first
func (h *Handler) AdminTools(c echo.Context) *echo.HTTPError {
	//overlappingTools, merr := h.registry.PressCycles.GetOverlappingTools()
	//if merr != nil {
	//	return merr.Echo()
	//}

	t := templates.AdminToolsSectionContent()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "AdminToolsSectionContent")
	}

	return nil
}

func (h *Handler) renderToolsSection(c echo.Context) *echo.HTTPError {
	var (
		tools     []*shared.Tool
		cassettes []*shared.Cassette
		errCh     = make(chan *echo.HTTPError, 2)
	)

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		var merr *errors.MasterError
		tools, merr = h.DB.Tool.Tool.List()
		if merr != nil {
			errCh <- merr.Echo()
		}
		errCh <- nil
	})

	wg.Go(func() {
		var merr *errors.MasterError
		cassettes, merr = h.DB.Tool.Cassette.List()
		if merr != nil {
			errCh <- merr.Echo()
		}
		errCh <- nil
	})

	regenerationsMap := make(map[shared.EntityID][]*shared.ToolRegeneration)
	wg.Go(func() {
		regenerations, merr := h.DB.Tool.Regeneration.List()
		if merr != nil {
			errCh <- merr.Echo()
		}

		for _, r := range regenerations {
			if _, ok := regenerationsMap[r.ToolID]; !ok {
				regenerationsMap[r.ToolID] = []*shared.ToolRegeneration{}
			}
			regenerationsMap[r.ToolID] = append(regenerationsMap[r.ToolID], r)
		}

		errCh <- nil
	})

	notesMap := make(map[shared.EntityID][]*shared.Note)
	wg.Go(func() {
		notes, merr := h.DB.Note.Note.List()
		if merr != nil {
			errCh <- merr.Echo()
		}

		for _, n := range notes {
			if _, ok := notesMap[n.ID]; !ok {
				notesMap[n.ID] = []*shared.Note{}
			}
			notesMap[n.ID] = append(notesMap[n.ID], n)
		}

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

	t := templates.SectionTools(templates.SectionToolsProps{
		Tools:         tools,
		Cassettes:     cassettes,
		User:          user,
		Regenerations: regenerationsMap,
		Notes:         notesMap,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionTools")
	}

	return nil
}

func (h *Handler) renderPressSection(c echo.Context) *echo.HTTPError {
	pressUtilizations, merr := helper.GetPressUtilizations(h.DB, shared.AllPressNumbers)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.SectionPress(templates.SectionPressProps{
		PressUtilizations: pressUtilizations,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionPress")
	}

	return nil
}
