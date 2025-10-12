package tools

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/tools/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type EditToolDialogFormData struct {
	Position models.Position     // Position form field name "position"
	Format   models.Format       // Format form field names "width" and "height"
	Type     string              // Type form field name "type"
	Code     string              // Code form field name "code"
	Press    *models.PressNumber // Press form field name "press-selection"
}

type Handler struct {
	*handlers.BaseHandler

	userNameMinLength int
	userNameMaxLength int
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(db,
			logger.NewComponentLogger("Tools")),
		userNameMinLength: 1,
		userNameMaxLength: 100,
	}
}

func (h *Handler) GetToolsPage(c echo.Context) error {
	h.Log.Info("Rendering tools page")

	page := templates.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tools page: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetEditToolDialog(c echo.Context) error {
	h.Log.Debug("Rendering edit tool dialog")

	props := &templates.DialogEditToolProps{}

	toolID, _ := h.ParseInt64Query(c, "id")
	if toolID > 0 {
		var err error
		props.Tool, err = h.DB.Tools.Get(toolID)
		if err != nil {
			return h.HandleError(c, err, "failed to get tool from database")
		}

		props.InputPosition = string(props.Tool.Position)
		props.InputWidth = props.Tool.Format.Width
		props.InputHeight = props.Tool.Format.Height
		props.InputType = props.Tool.Type
		props.InputCode = props.Tool.Code
		props.InputPressSelection = props.Tool.Press
	}

	props.ReloadPage = h.ParseBoolQuery(c, "reload_page")

	toolEdit := templates.DialogEditTool(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render tool edit dialog: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXPostEditToolDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.Log.Debug("User %s creating new tool", user.Name)

	formData, err := h.getEditToolFormData(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to get tool form data: "+err.Error())
	}

	tool := models.NewTool(
		formData.Position, formData.Format, formData.Code, formData.Type,
	)
	tool.Press = formData.Press

	if t, err := h.DB.Tools.AddWithNotes(tool, user); err != nil {
		return h.HandleError(c, err, "failed to add tool")
	} else {
		h.Log.Info("Created tool ID %d (Type=%s, Code=%s) by user %s",
			t.ID, tool.Type, tool.Code, user.Name)

		// Create feed entry
		title := "Neues Werkzeug erstellt"
		content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
			t.String(), t.Type, t.Code, string(t.Position))
		if t.Press != nil {
			content += fmt.Sprintf("\nPresse: %d", *t.Press)
		}

		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for tool creation: %v", err)
		}
	}

	return nil
}

func (h *Handler) HTMXPutEditToolDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	h.Log.Warn("User %s updating tool %d", user.Name, toolID)

	formData, err := h.getEditToolFormData(c)
	if err != nil {
		return h.RenderBadRequest(c, "failed to get tool form data: "+err.Error())
	}

	tool := models.NewTool(
		formData.Position, formData.Format, formData.Code, formData.Type,
	)
	tool.ID = toolID
	tool.Press = formData.Press

	if err := h.DB.Tools.Update(tool, user); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to update tool: "+err.Error())
	} else {
		h.Log.Info("Updated tool %d (Type=%s, Code=%s) by user %s",
			tool.ID, tool.Type, tool.Code, user.Name)

		// Create feed entry
		title := "Werkzeug aktualisiert"
		content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
			tool.String(), tool.Type, tool.Code, string(tool.Position))
		if tool.Press != nil {
			content += fmt.Sprintf("\nPresse: %d", *tool.Press)
		}

		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for tool update: %v", err)
		}
	}

	return nil
}

func (h *Handler) HTMXDeleteTool(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c,
			"invalid or missing id parameter: "+err.Error())
	}

	// Get user from context for audit trail
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.Log.Info("User %s deleting tool %d", user.Name, toolID)

	// Get tool data before deletion for the feed
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool for deletion")
	}

	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		return h.HandleError(c, err, "failed to delete tool")
	}

	// Create feed entry
	title := "Werkzeug gelöscht"
	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))
	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for tool deletion: %v", err)
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) HTMXMarkToolAsDead(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c,
			"invalid or missing id parameter: "+err.Error())
	}

	// Get user from context for audit trail
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.Log.Info("User %s marking tool %d as dead", user.Name, toolID)

	// Get tool data before marking as dead for the feed
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool for marking as dead")
	}

	// Check if tool is already dead
	if tool.IsDead {
		return h.RenderBadRequest(c, "tool is already marked as dead")
	}

	// Mark the tool as dead in database
	if err := h.DB.Tools.MarkAsDead(toolID, user); err != nil {
		return h.HandleError(c, err, "failed to mark tool as dead")
	}

	// Create feed entry
	title := "Werkzeug als Tod markiert"
	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))
	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for tool marking as dead: %v", err)
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) HTMXGetSectionPress(c echo.Context) error {
	h.Log.Debug("Rendering press section")

	pressUtilization, err := h.DB.Tools.GetPressUtilization()
	if err != nil {
		return h.HandleError(c, err, "failed to get press utilization")
	}

	sectionPress := templates.SectionPress(pressUtilization)

	if err := sectionPress.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press section: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetSectionTools(c echo.Context) error {
	h.Log.Debug("Rendering tools section")

	// Get tools from database
	allTools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	// Filter out dead tools
	var tools []*models.ToolWithNotes
	for _, tool := range allTools {
		if !tool.IsDead {
			tools = append(tools, tool)
		}
	}

	sectionTools := templates.SectionTools(tools)
	if err := sectionTools.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tools section: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetAdminOverlappingTools(c echo.Context) error {
	h.Log.Debug("Getting overlapping tools for admin section")

	// Get user from context for audit trail
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.Log.Info("User %s requested overlapping tools analysis", user.Name)

	// Get overlapping tools from service
	overlappingTools, err := h.DB.PressCycles.GetOverlappingTools(h.DB.Tools, h.DB.Users)
	if err != nil {
		return h.HandleError(c, err, "failed to get overlapping tools")
	}

	// Render the admin overlapping tools section
	adminSection := templates.AdminOverlappingTools(overlappingTools)
	if err := adminSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render admin overlapping tools section: "+err.Error())
	}

	return nil
}

func (h *Handler) getEditToolFormData(c echo.Context) (*EditToolDialogFormData, error) {
	// Parse position with validation
	var position models.Position

	positionFormValue := h.GetSanitizedFormValue(c, "position")
	switch models.Position(positionFormValue) {
	case models.PositionTop:
		position = models.PositionTop
	case models.PositionTopCassette:
		position = models.PositionTopCassette
	case models.PositionBottom:
		position = models.PositionBottom
	default:
		return nil, errors.New("invalid position: " + positionFormValue)
	}

	data := &EditToolDialogFormData{
		Position: position,
	}

	// Parse width and height with validation
	widthStr := c.FormValue("width")
	if widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return nil, errors.New("invalid width: " + err.Error())
		}

		data.Format.Width = width
	}

	heightStr := c.FormValue("height")
	if heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, errors.New("invalid height: " + err.Error())
		}

		data.Format.Height = height
	}

	// Parse type with validation (Optional)
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if len(data.Type) > 25 {
		return nil, errors.New("type must be 25 characters or less")
	}

	// Parse code with validation
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		return nil, errors.New("code is required")
	}
	if len(data.Code) > 25 {
		return nil, errors.New("code must be 25 characters or less")
	}

	// Parse press selection with validation
	pressStr := c.FormValue("press-selection")
	if pressStr != "" {
		press, err := strconv.Atoi(pressStr)
		if err != nil {
			return nil, errors.New("invalid press number: " + err.Error())
		}

		pn := models.PressNumber(press)
		data.Press = &pn
		if !models.IsValidPressNumber(data.Press) {
			return nil, errors.New("invalid press number: must be 0, 2, 3, 4, or 5")
		}
	}

	return data, nil
}
