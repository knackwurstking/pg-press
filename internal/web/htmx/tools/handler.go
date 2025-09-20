package tools

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/dialogs"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

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
			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/edit", h.AddTool),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/edit", h.UpdateTool),

			// Delete a tool
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete",
				h.DeleteTool),
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

	toolsList := ListTools(tools)
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

func (h *Tools) AddTool(c echo.Context) error {
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
	}

	return nil
}

func (h *Tools) UpdateTool(c echo.Context) error {
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

	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		return h.HandleError(c, err, "failed to delete tool")
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
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
