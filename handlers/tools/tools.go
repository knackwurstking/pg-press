package tools

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/tools/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
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
	user, merr := utils.GetUserFromContext(c)
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

	user, merr := utils.GetUserFromContext(c)
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

	user, merr := utils.GetUserFromContext(c)
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

func (h *Handler) HTMXGetSectionTools(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	allTools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

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

func (h *Handler) HTMXGetAdminOverlappingTools(c echo.Context) error {
	overlappingTools, merr := h.registry.PressCycles.GetOverlappingTools()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.AdminToolsSectionContent(overlappingTools)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "AdminToolsSectionContent")
	}

	return nil
}

func (h *Handler) createToolFeed(user *models.User, tool *models.Tool, title string) {
	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))
	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	if _, err := h.registry.Feeds.AddSimple(title, content, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed", "error", err)
	}
}
