package tools

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/knackwurstking/pg-press/handlers/tools/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
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

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/tools", h.GetToolsPage),

		utils.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete", h.HTMXDeleteTool),
		utils.NewEchoRoute(http.MethodPatch, "/htmx/tools/mark-dead", h.HTMXMarkToolAsDead),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/section/press", h.HTMXGetSectionPress),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/section/tools", h.HTMXGetSectionTools),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/admin/overlapping-tools", h.HTMXGetAdminOverlappingTools),
	})
}

func (h *Handler) GetToolsPage(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	page := components.PageTools(user)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render tools page")
	}
	return nil
}

func (h *Handler) HTMXDeleteTool(c echo.Context) error {
	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "invalid or missing id parameter")
	}
	toolID := models.ToolID(toolIDQuery)

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return utils.HandleError(err, "failed to get tool for deletion")
	}

	if err := h.registry.Tools.Delete(toolID, user); err != nil {
		return utils.HandleError(err, "failed to delete tool")
	}

	slog.Info("Tool deleted", "id", toolID, "user_name", user.Name)

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug gelöscht")

	utils.SetHXRedirect(c, "/tools")
	return nil
}

func (h *Handler) HTMXMarkToolAsDead(c echo.Context) error {
	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "invalid or missing id parameter")
	}
	toolID := models.ToolID(toolIDQuery)

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return utils.HandleError(err, "failed to get tool for marking as dead")
	}

	if tool.IsDead {
		return utils.HandleBadRequest(nil, "tool is already marked as dead")
	}

	if err := h.registry.Tools.MarkAsDead(toolID, user); err != nil {
		return utils.HandleError(err, "failed to mark tool as dead")
	}

	slog.Info("Tool marked as dead", "id", toolID, "user_name", user.Name)

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug als Tot markiert")

	utils.SetHXRedirect(c, "/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) HTMXGetSectionPress(c echo.Context) error {
	pressUtilization, err := h.registry.Tools.GetPressUtilization()
	if err != nil {
		return utils.HandleError(err, "failed to get press utilization")
	}

	section := components.SectionPress(pressUtilization)
	if err := section.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render press section")
	}
	return nil
}

func (h *Handler) HTMXGetSectionTools(c echo.Context) error {
	allTools, err := h.registry.Tools.List()
	if err != nil {
		return utils.HandleError(err, "failed to get tools from database")
	}

	var tools []*models.ResolvedTool
	for _, t := range allTools {
		if t.IsDead {
			continue
		}

		var bindingTool *models.Tool
		if t.IsBound() {
			bindingTool, err = h.registry.Tools.Get(*t.Binding)
			if err != nil {
				return utils.HandleError(err, "failed to get binding tool")
			}
		}

		notes, err := h.registry.Notes.GetByTool(t.ID)
		if err != nil {
			return utils.HandleError(err, "failed to get notes for tool")
		}

		tools = append(tools, models.NewResolvedTool(t, bindingTool, notes))
	}

	section := components.SectionTools(tools)
	if err := section.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render tools section")
	}
	return nil
}

func (h *Handler) HTMXGetAdminOverlappingTools(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleError(err, "failed to get user from context")
	}

	slog.Info("User requested overlapping tools analysis", "user_name", user.Name)

	overlappingTools, err := h.registry.PressCycles.GetOverlappingTools()
	if err != nil {
		return utils.HandleError(err, "failed to get overlapping tools")
	}

	section := components.AdminToolsSectionContent(overlappingTools)
	if err := section.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render admin overlapping tools section")
	}
	return nil
}

func (h *Handler) createToolFeed(user *models.User, tool *models.Tool, title string) {
	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))
	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "error", err)
	}
}
