package tools

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/services"
	ui "github.com/knackwurstking/ui/ui-templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
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

func (h *Handler) GetToolsPage(c echo.Context) error {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Tools Page")
	}

	return nil
}

func (h *Handler) HTMXDeleteTool(c echo.Context) error {
	slog.Info("Deleting tool")

	toolIDQuery, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(toolIDQuery)

	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	merr = h.registry.Tools.Delete(toolID, user)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug gelöscht")

	utils.SetHXRedirect(c, "/tools")
	return nil
}

func (h *Handler) HTMXMarkToolAsDead(c echo.Context) error {
	slog.Info("Marking tool as dead")

	toolIDQuery, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(toolIDQuery)

	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	if tool.IsDead {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"tool is already marked as dead",
		)
	}

	merr = h.registry.Tools.MarkAsDead(toolID, user)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug als Tot markiert")

	utils.SetHXRedirect(c, "/tools")
	return nil
}

func (h *Handler) HTMXGetSectionPress(c echo.Context) error {
	pressUtilization, merr := h.registry.Tools.PressUtilization()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.SectionPress(pressUtilization)
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

	allTools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	// TODO: Continue here...
	var tools []*models.ResolvedTool
	for _, t := range allTools {
		if t.IsDead {
			continue
		}

		rt, merr := services.ResolveTool(h.registry, t)
		if merr != nil {
			return merr.Echo()
		}

		tools = append(tools, rt)
	}

	t := templates.SectionTools(tools, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionTools")
	}

	return nil
}
