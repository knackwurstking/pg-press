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
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	t := templates.Page(user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ToolsPage")
	}
	return nil
}

func (h *Handler) HTMXDeleteTool(c echo.Context) error {
	slog.Info("Deleting tool")

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "invalid or missing id parameter")
	}
	toolID := models.ToolID(toolIDQuery)

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool for deletion")
	}

	dberr = h.registry.Tools.Delete(toolID, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "delete tool")
	}

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug gelöscht")

	utils.SetHXRedirect(c, "/tools")
	return nil
}

func (h *Handler) HTMXMarkToolAsDead(c echo.Context) error {
	slog.Info("Marking tool as dead")

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "invalid or missing id parameter")
	}
	toolID := models.ToolID(toolIDQuery)

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	tool, dberr := h.registry.Tools.Get(toolID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get tool for marking as dead")
	}

	if tool.IsDead {
		return errors.NewBadRequestError(nil, "tool is already marked as dead")
	}

	dberr = h.registry.Tools.MarkAsDead(toolID, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "mark tool as dead")
	}

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug als Tot markiert")

	utils.SetHXRedirect(c, "/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) HTMXGetSectionPress(c echo.Context) error {
	pressUtilization, dberr := h.registry.Tools.PressUtilization()
	if dberr != nil {
		return errors.HandlerError(dberr, "get press utilization")
	}

	t := templates.SectionPress(pressUtilization)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "SectionPress")
	}
	return nil
}

func (h *Handler) HTMXGetSectionTools(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	allTools, dberr := h.registry.Tools.List()
	if dberr != nil {
		return errors.HandlerError(dberr, "get tools from database")
	}

	var tools []*models.ResolvedTool
	for _, t := range allTools {
		if t.IsDead {
			continue
		}

		rt, err := services.ResolveTool(h.registry, t)
		if err != nil {
			return errors.HandlerError(err, "resolving tool")
		}

		tools = append(tools, rt)
	}

	t := templates.SectionTools(tools, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "render tools section")
	}
	return nil
}

func (h *Handler) HTMXGetAdminOverlappingTools(c echo.Context) error {
	overlappingTools, dberr := h.registry.PressCycles.GetOverlappingTools()
	if dberr != nil {
		return errors.HandlerError(dberr, "get overlapping tools")
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
