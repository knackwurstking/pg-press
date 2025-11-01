package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type ToolsDialogEditForm struct {
	Position models.Position
	Format   models.Format
	Type     string
	Code     string
	Press    *models.PressNumber
}

type Tools struct {
	registry *services.Registry
}

func NewTools(r *services.Registry) *Tools {
	return &Tools{
		registry: r,
	}
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/tools", h.GetToolsPage),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/edit", h.HTMXGetEditToolDialog),
		utils.NewEchoRoute(http.MethodPost, "/htmx/tools/edit", h.HTMXPostEditToolDialog),
		utils.NewEchoRoute(http.MethodPut, "/htmx/tools/edit", h.HTMXPutEditToolDialog),
		utils.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete", h.HTMXDeleteTool),
		utils.NewEchoRoute(http.MethodPatch, "/htmx/tools/mark-dead", h.HTMXMarkToolAsDead),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/section/press", h.HTMXGetSectionPress),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/section/tools", h.HTMXGetSectionTools),
		utils.NewEchoRoute(http.MethodGet, "/htmx/tools/admin/overlapping-tools", h.HTMXGetAdminOverlappingTools),
	})
}

func (h *Tools) GetToolsPage(c echo.Context) error {
	page := components.PageTools()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tools page")
	}
	return nil
}

func (h *Tools) HTMXGetEditToolDialog(c echo.Context) error {
	props := &components.DialogEditToolProps{}

	toolIDQuery, _ := ParseQueryInt64(c, "id")
	if toolIDQuery > 0 {
		tool, err := h.registry.Tools.Get(models.ToolID(toolIDQuery))
		if err != nil {
			return HandleError(err, "failed to get tool from database")
		}

		props.Tool = tool
		props.InputPosition = string(tool.Position)
		props.InputWidth = tool.Format.Width
		props.InputHeight = tool.Format.Height
		props.InputType = tool.Type
		props.InputCode = tool.Code
		props.InputPressSelection = tool.Press
	}

	dialog := components.DialogEditTool(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tool edit dialog")
	}
	return nil
}

func (h *Tools) HTMXPostEditToolDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	formData, err := h.getEditToolFormData(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get tool form data")
	}

	tool := models.NewTool(formData.Position, formData.Format, formData.Code, formData.Type)
	tool.SetPress(formData.Press)

	id, err := h.registry.Tools.Add(tool, user)
	if err != nil {
		return HandleError(err, "failed to add tool")
	}

	slog.Info("Created tool", "id", id, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	h.createToolFeed(user, tool, "Neues Werkzeug erstellt")

	SetHXTrigger(c, env.HXGlobalTrigger)
	return nil
}

func (h *Tools) HTMXPutEditToolDialog(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	formData, err := h.getEditToolFormData(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get tool form data")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool")
	}

	tool.Press = formData.Press
	tool.Position = formData.Position
	tool.Format = formData.Format
	tool.Code = formData.Code
	tool.Type = formData.Type

	if err := h.registry.Tools.Update(tool, user); err != nil {
		return HandleError(err, "failed to update tool")
	}

	slog.Info("Updated tool", "id", tool.ID, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug aktualisiert")

	SetHXTrigger(c, env.HXGlobalTrigger)

	SetHXAfterSettle(c, map[string]interface{}{
		"toolUpdated": map[string]string{
			"pageTitle": fmt.Sprintf("PG Presse | %s %s",
				tool.String(), tool.Position.GermanString()),
			"appBarTitle": fmt.Sprintf("%s %s", tool.String(),
				tool.Position.GermanString()),
		},
	})

	return nil
}

func (h *Tools) HTMXDeleteTool(c echo.Context) error {
	toolIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "invalid or missing id parameter")
	}
	toolID := models.ToolID(toolIDQuery)

	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool for deletion")
	}

	if err := h.registry.Tools.Delete(toolID, user); err != nil {
		return HandleError(err, "failed to delete tool")
	}

	slog.Info("Tool deleted", "id", toolID, "user_name", user.Name)

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug gelöscht")

	SetHXRedirect(c, "/tools")
	return nil
}

func (h *Tools) HTMXMarkToolAsDead(c echo.Context) error {
	toolIDQuery, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "invalid or missing id parameter")
	}
	toolID := models.ToolID(toolIDQuery)

	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return HandleError(err, "failed to get tool for marking as dead")
	}

	if tool.IsDead {
		return HandleBadRequest(nil, "tool is already marked as dead")
	}

	if err := h.registry.Tools.MarkAsDead(toolID, user); err != nil {
		return HandleError(err, "failed to mark tool as dead")
	}

	slog.Info("Tool marked as dead", "id", toolID, "user_name", user.Name)

	// Create feed entry
	h.createToolFeed(user, tool, "Werkzeug als Tot markiert")

	SetHXRedirect(c, "/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Tools) HTMXGetSectionPress(c echo.Context) error {
	pressUtilization, err := h.registry.Tools.GetPressUtilization()
	if err != nil {
		return HandleError(err, "failed to get press utilization")
	}

	section := components.PageTools_SectionPress(pressUtilization)
	if err := section.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render press section")
	}
	return nil
}

func (h *Tools) HTMXGetSectionTools(c echo.Context) error {
	allTools, err := h.registry.Tools.List()
	if err != nil {
		return HandleError(err, "failed to get tools from database")
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
				return HandleError(err, "failed to get binding tool")
			}
		}

		notes, err := h.registry.Notes.GetByTool(t.ID)
		if err != nil {
			return HandleError(err, "failed to get notes for tool")
		}

		tools = append(tools, models.NewResolvedTool(t, bindingTool, notes))
	}

	section := components.PageTools_SectionTools(tools)
	if err := section.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render tools section")
	}
	return nil
}

func (h *Tools) HTMXGetAdminOverlappingTools(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	slog.Info("User requested overlapping tools analysis", "user_name", user.Name)

	overlappingTools, err := h.registry.PressCycles.GetOverlappingTools()
	if err != nil {
		return HandleError(err, "failed to get overlapping tools")
	}

	section := components.PageTools_AdminOverlappingToolsSectionContent(overlappingTools)
	if err := section.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render admin overlapping tools section")
	}
	return nil
}

func (h *Tools) createToolFeed(user *models.User, tool *models.Tool, title string) {
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

func (h *Tools) getEditToolFormData(c echo.Context) (*ToolsDialogEditForm, error) {
	positionStr := c.FormValue("position")
	position := models.Position(positionStr)

	switch position {
	case models.PositionTop, models.PositionTopCassette, models.PositionBottom:
		// Valid position
	default:
		return nil, errors.NewValidationError(fmt.Sprintf("invalid position: %s", positionStr))
	}

	data := &ToolsDialogEditForm{Position: position}

	// Parse width
	if widthStr := c.FormValue("width"); widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid width: %v", err))
		}
		data.Format.Width = width
	}

	// Parse height
	if heightStr := c.FormValue("height"); heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid height: %v", err))
		}
		data.Format.Height = height
	}

	// Parse type
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if len(data.Type) > 25 {
		return nil, errors.NewValidationError("type must be 25 characters or less")
	}

	// Parse code
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		return nil, errors.NewValidationError("code is required")
	}
	if len(data.Code) > 25 {
		return nil, errors.NewValidationError("code must be 25 characters or less")
	}

	// Parse press
	if pressStr := c.FormValue("press-selection"); pressStr != "" {
		press, err := strconv.Atoi(pressStr)
		if err != nil {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid press number: %v", err))
		}

		pressNumber := models.PressNumber(press)
		if !models.IsValidPressNumber(&pressNumber) {
			return nil, errors.NewValidationError("invalid press number: must be 0, 2, 3, 4, or 5")
		}
		data.Press = &pressNumber
	}

	return data, nil
}
