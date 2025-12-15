package tools

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/helper"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB     *common.DB
	Logger *ui.Logger
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		DB:     db,
		Logger: env.NewLogger("handler: tools"),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, h.GetToolsPage),
		ui.NewEchoRoute(http.MethodDelete, path+"/delete", h.HTMXDeleteTool),
		ui.NewEchoRoute(http.MethodPatch, path+"/mark-dead", h.HTMXMarkToolAsDead),
		ui.NewEchoRoute(http.MethodGet, path+"/section/press", h.HTMXGetSectionPress),
		ui.NewEchoRoute(http.MethodGet, path+"/section/tools", h.HTMXGetSectionTools),
		ui.NewEchoRoute(http.MethodGet, path+"/admin/overlapping-tools", h.HTMXGetAdminOverlappingTools),
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

func (h *Handler) HTMXDeleteTool(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

	merr = h.DB.Tool.Tool.Delete(toolID)
	if merr != nil {
		return merr.Echo()
	}

	h.Logger.Debug("Deleted tool with ID: %#v", toolID)

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

func (h *Handler) HTMXMarkToolAsDead(c echo.Context) *echo.HTTPError {
	id, merr := shared.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := shared.EntityID(id)

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

	urlb.SetHXTrigger(c, "tools-tab")

	return nil
}

func (h *Handler) HTMXGetSectionPress(c echo.Context) *echo.HTTPError {
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

func (h *Handler) HTMXGetSectionTools(c echo.Context) *echo.HTTPError {
	return h.renderSectionTools(c)
}

// TODO: Fix all other stuff first
func (h *Handler) HTMXGetAdminOverlappingTools(c echo.Context) *echo.HTTPError {
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

func (h *Handler) renderSectionTools(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	tools, merr := h.DB.Tool.Tool.List()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.SectionTools(templates.SectionToolsProps{
		Tools: tools,
		User:  user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionTools")
	}

	return nil
}
