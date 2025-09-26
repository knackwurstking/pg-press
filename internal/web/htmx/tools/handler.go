package tools

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/html/tools"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/internal/web/shared/dialogs"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type EditFormData struct {
	Position models.Position     // Position form field name "position"
	Format   models.Format       // Format form field names "width" and "height"
	Type     string              // Type form field name "type"
	Code     string              // Code form field name "code"
	Press    *models.PressNumber // Press form field name "press-selection"
}

type Tools struct {
	*handlers.BaseHandler
}

func NewTools(db *database.DB) *Tools {
	return &Tools{
		BaseHandler: handlers.NewBaseHandler(db, logger.HTMXHandlerTools()),
	}
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/list", h.GetToolsList),

			// Get, Post or Edit a tool
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/edit", h.GetEditDialog),
			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/edit",
				h.AddToolOnEditDialogSubmit),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/edit",
				h.UpdateToolOnEditDialogSubmit),

			// Delete a tool
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete",
				h.DeleteTool),

			// Tool status management
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-edit", h.GetStatusEdit),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/status-display", h.GetStatusDisplay),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/status", h.UpdateToolStatus),

			// Press page sections
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/active-tools", h.GetActiveToolsSection),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/metal-sheets", h.GetMetalSheetsSection),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/press/cycles", h.GetCyclesSection),
		},
	)
}

func (h *Tools) GetToolsList(c echo.Context) error {
	start := time.Now()
	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	dbElapsed := time.Since(start)
	if dbElapsed > 100*time.Millisecond {
		h.LogWarn("Slow tools query took %v for %d tools", dbElapsed, len(tools))
	}

	toolsList := components.ToolsList(tools)
	if err := toolsList.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render tools list all: "+err.Error())
	}

	return nil
}

func (h *Tools) GetEditDialog(c echo.Context) error {
	h.LogDebug("Rendering edit tool dialog")

	props := &dialogs.EditToolProps{}

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

	toolEdit := dialogs.EditTool(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render tool edit dialog: "+err.Error())
	}

	return nil
}

func (h *Tools) AddToolOnEditDialogSubmit(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("User %s creating new tool", user.Name)

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
		h.LogInfo("Created tool ID %d (Type=%s, Code=%s) by user %s",
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
			h.LogError("Failed to create feed for tool creation: %v", err)
		}
	}

	return h.closeDialog(c)
}

func (h *Tools) UpdateToolOnEditDialogSubmit(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	h.LogWarn("User %s updating tool %d", user.Name, toolID)

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
		h.LogInfo("Updated tool %d (Type=%s, Code=%s) by user %s",
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
			h.LogError("Failed to create feed for tool update: %v", err)
		}
	}

	return h.closeDialog(c)
}

func (h *Tools) closeDialog(c echo.Context) error {
	dialog := dialogs.EditTool(&dialogs.EditToolProps{
		CloseDialog: true,
	})

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render tool edit dialog: "+err.Error())
	}

	return nil
}

func (h *Tools) DeleteTool(c echo.Context) error {
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

	h.LogDebug("User %s deleting tool %d", user.Name, toolID)

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
		h.LogError("Failed to create feed for tool deletion: %v", err)
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Tools) GetStatusEdit(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	statusEdit := h.renderStatusComponent(tool, true, user)
	if err := statusEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool status edit: "+err.Error())
	}

	return nil
}

func (h *Tools) GetStatusDisplay(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse tool ID: "+err.Error())
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	statusDisplay := h.renderStatusComponent(tool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool status display: "+err.Error())
	}

	return nil
}

func (h *Tools) UpdateToolStatus(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	toolIDStr := c.FormValue("tool_id")
	if toolIDStr == "" {
		return h.RenderBadRequest(c, "tool_id is required")
	}

	toolID, err := strconv.ParseInt(toolIDStr, 10, 64)
	if err != nil {
		return h.RenderBadRequest(c, "invalid tool_id: "+err.Error())
	}

	statusStr := c.FormValue("status")
	if statusStr == "" {
		return h.RenderBadRequest(c, "status is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool from database")
	}

	h.LogInfo("User %s updating status for tool %d from %s to %s", user.Name, toolID, tool.Status(), statusStr)

	// Handle regeneration start/stop/abort only
	switch statusStr {
	case "regenerating":
		// Start regeneration
		if err := h.DB.Tools.UpdateRegenerating(toolID, true, user); err != nil {
			return h.HandleError(c, err, "failed to start tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung gestartet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.LogError("Failed to create feed for regeneration start: %v", err)
		}

	case "active":
		// Stop regeneration (return to active status)
		if err := h.DB.Tools.UpdateRegenerating(toolID, false, user); err != nil {
			return h.HandleError(c, err, "failed to stop tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung beendet"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.LogError("Failed to create feed for regeneration stop: %v", err)
		}

	case "abort":
		// Abort regeneration (remove regeneration record and set status to false)
		if err := h.DB.ToolRegenerations.AbortToolRegeneration(toolID, user); err != nil {
			return h.HandleError(c, err, "failed to abort tool regeneration")
		}

		// Create feed entry
		title := "Werkzeug Regenerierung abgebrochen"
		content := fmt.Sprintf("Werkzeug: %s", tool.String())
		feed := models.NewFeed(title, content, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.LogError("Failed to create feed for regeneration abort: %v", err)
		}

	default:
		return h.RenderBadRequest(c, "invalid status: must be 'regenerating', 'active', or 'abort'")
	}

	// Get updated tool and render status display
	updatedTool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return h.HandleError(c, err, "failed to get updated tool from database")
	}

	statusDisplay := h.renderStatusComponent(updatedTool, false, user)
	if err := statusDisplay.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render updated tool status: "+err.Error())
	}

	return nil
}

func (h *Tools) renderStatusComponent(tool *models.Tool, editable bool, user *models.User) templ.Component {
	return components.ToolStatusEdit(&components.ToolStatusEditProps{
		Tool:              tool,
		Editable:          editable,
		UserHasPermission: user.IsAdmin(),
	})
}

func (h *Tools) getEditToolFormData(c echo.Context) (*EditFormData, error) {
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

	data := &EditFormData{
		Position: position,
	}

	// Parse width and height with validation
	widthStr := c.FormValue("width")
	if widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return nil, errors.New("invalid width: " + err.Error())
		}
		if width <= 0 || width > 10000 {
			return nil, errors.New("width must be between 1 and 10000")
		}
		data.Format.Width = width
	}

	heightStr := c.FormValue("height")
	if heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, errors.New("invalid height: " + err.Error())
		}
		if height <= 0 || height > 10000 {
			return nil, errors.New("height must be between 1 and 10000")
		}
		data.Format.Height = height
	}

	// Parse type with validation
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if data.Type == "" {
		return nil, errors.New("type is required")
	}
	if len(data.Type) > 50 {
		return nil, errors.New("type must be 50 characters or less")
	}

	// Parse code with validation
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		return nil, errors.New("code is required")
	}
	if len(data.Code) > 50 {
		return nil, errors.New("code must be 50 characters or less")
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

// GetActiveToolsSection handles HTMX requests for the active tools section
func (h *Tools) GetActiveToolsSection(c echo.Context) error {
	pressNum, err := h.ParseInt64Query(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	// Filter tools for this press and create toolsMap
	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		if tool.Press != nil && *tool.Press == press {
			toolsMap[tool.ID] = tool
		}
	}

	activeToolsSection := h.renderActiveToolsSection(toolsMap, press)
	if err := activeToolsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render active tools section: "+err.Error())
	}

	return nil
}

// GetMetalSheetsSection handles HTMX requests for the metal sheets section
func (h *Tools) GetMetalSheetsSection(c echo.Context) error {
	pressNum, err := h.ParseInt64Query(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get tools for this press
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	// Filter tools for this press and create toolsMap
	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		if tool.Press != nil && *tool.Press == press {
			toolsMap[tool.ID] = tool
		}
	}

	// Get metal sheets for tools on this press
	var metalSheets []*models.MetalSheet
	for toolID := range toolsMap {
		sheets, err := h.DB.MetalSheets.GetByToolID(toolID)
		if err != nil {
			h.LogError("Failed to get metal sheets for tool %d: %v", toolID, err)
			continue
		}
		metalSheets = append(metalSheets, sheets...)
	}

	metalSheetsSection := h.renderMetalSheetsSection(metalSheets, toolsMap)
	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render metal sheets section: "+err.Error())
	}

	return nil
}

// GetCyclesSection handles HTMX requests for the cycles section
func (h *Tools) GetCyclesSection(c echo.Context) error {
	pressNum, err := h.ParseInt64Query(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get user for permissions
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get cycles for this press
	cycles, err := h.DB.PressCycles.GetPressCycles(press, nil, nil)
	if err != nil {
		return h.HandleError(c, err, "failed to get cycles from database")
	}

	// Get tools for this press to create toolsMap
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		toolsMap[tool.ID] = tool
	}

	cyclesSection := h.renderCyclesSection(cycles, toolsMap, user, press)
	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render cycles section: "+err.Error())
	}

	return nil
}

// renderActiveToolsSection renders the active tools section content
func (h *Tools) renderActiveToolsSection(toolsMap map[int64]*models.Tool, press models.PressNumber) templ.Component {
	return tools.ActiveToolsSection(toolsMap, press)
}

// renderMetalSheetsSection renders the metal sheets section content
func (h *Tools) renderMetalSheetsSection(metalSheets []*models.MetalSheet, toolsMap map[int64]*models.Tool) templ.Component {
	return tools.MetalSheetsSection(metalSheets, toolsMap)
}

// renderCyclesSection renders the cycles section content
func (h *Tools) renderCyclesSection(cycles []*models.Cycle, toolsMap map[int64]*models.Tool, user *models.User, press models.PressNumber) templ.Component {
	return tools.CyclesSection(cycles, toolsMap, user, press)
}
