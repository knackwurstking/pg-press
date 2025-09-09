package htmx

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/components/dialogs"
	toolscomp "github.com/knackwurstking/pgpress/internal/web/templates/components/tools"

	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/list", h.handleList),

			// Get, Post or Edit a tool
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/edit", func(c echo.Context) error {
				return h.handleEdit(c, nil)
			}),

			webhelpers.NewEchoRoute(http.MethodPost, "/htmx/tools/edit", h.handleEditPOST),
			webhelpers.NewEchoRoute(http.MethodPut, "/htmx/tools/edit", h.handleEditPUT),

			// Delete a tool
			webhelpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete", h.handleDelete),
		},
	)
}

func (h *Tools) handleList(c echo.Context) error {
	remoteIP := c.RealIP()
	userAgent := c.Request().UserAgent()
	logger.HTMXHandlerTools().Info("Tools list request from %s (user-agent: %s)", remoteIP, userAgent)

	start := time.Now()
	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to fetch tools from database: %v", err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	dbElapsed := time.Since(start)
	logger.HTMXHandlerTools().Debug("Retrieved %d tools from database in %v", len(tools), dbElapsed)
	if dbElapsed > 100*time.Millisecond {
		logger.HTMXHandlerTools().Warn("Slow tools query took %v for %d tools", dbElapsed, len(tools))
	}

	renderStart := time.Now()
	toolsListAll := toolscomp.List(tools)
	if err := toolsListAll.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTools().Error("Failed to render tools list template: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools list all: "+err.Error())
	}

	renderElapsed := time.Since(renderStart)
	totalElapsed := time.Since(start)
	logger.HTMXHandlerTools().Debug("Tools list rendered in %v (db: %v, render: %v, total: %v)",
		renderElapsed, dbElapsed, renderElapsed, totalElapsed)

	return nil
}

func (h *Tools) handleEdit(c echo.Context, props *dialogs.EditToolProps) error {
	remoteIP := c.RealIP()

	if props == nil {
		props = &dialogs.EditToolProps{}
		props.ToolID, _ = webhelpers.ParseInt64Query(c, constants.QueryParamID)
		props.Close = webhelpers.ParseBoolQuery(c, constants.QueryParamClose)

		if props.ToolID > 0 {
			logger.HTMXHandlerTools().Info("Opening tool edit dialog for tool ID %d from %s", props.ToolID, remoteIP)
			start := time.Now()
			tool, err := h.DB.Tools.GetWithNotes(props.ToolID)
			if err != nil {
				logger.HTMXHandlerTools().Error("Failed to get tool %d for editing: %v", props.ToolID, err)
				return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
					"failed to get tool from database: "+err.Error())
			}
			elapsed := time.Since(start)
			logger.HTMXHandlerTools().Debug("Loaded tool %d (%s %s) for editing in %v",
				props.ToolID, tool.Type, tool.Code, elapsed)

			props.InputPosition = string(tool.Position)
			props.InputWidth = tool.Format.Width
			props.InputHeight = tool.Format.Height
			props.InputType = tool.Type
			props.InputCode = tool.Code
			props.InputPressSelection = tool.Press
		} else {
			logger.HTMXHandlerTools().Info("Opening new tool creation dialog from %s", remoteIP)
		}
	}

	renderStart := time.Now()
	toolEdit := dialogs.EditTool(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTools().Error("Failed to render tool edit dialog: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}
	renderElapsed := time.Since(renderStart)
	logger.HTMXHandlerTools().Debug("Tool edit dialog rendered in %v", renderElapsed)
	return nil
}

func (h *Tools) handleEditPOST(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get user from context for tool creation: %v", err)
		return err
	}

	remoteIP := c.RealIP()
	logger.HTMXHandlerTools().Info("User %s (ID: %d) creating new tool from %s", user.Name, user.TelegramID, remoteIP)

	start := time.Now()
	formData, err := h.getToolFormData(c)
	if err != nil {
		logger.HTMXHandlerTools().Warn("Invalid form data for tool creation by user %s: %v", user.Name, err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to get tool form data: "+err.Error())
	}

	props := &dialogs.EditToolProps{
		InputPosition:       string(formData.Position),
		InputWidth:          formData.Format.Width,
		InputHeight:         formData.Format.Height,
		InputType:           formData.Type,
		InputCode:           formData.Code,
		InputPressSelection: formData.Press,
	}

	tool := toolmodels.New(formData.Position)
	tool.Format = formData.Format
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Press = formData.Press

	logger.HTMXHandlerTools().Info("Creating tool: Type=%s, Code=%s, Position=%s, Format=%dx%d by user %s",
		tool.Type, tool.Code, tool.Position, tool.Format.Width, tool.Format.Height, user.Name)

	dbStart := time.Now()
	if t, err := h.DB.Tools.AddWithNotes(tool, user); err != nil {
		if err == dberror.ErrAlreadyExists {
			logger.HTMXHandlerTools().Warn("Tool creation failed - already exists: Type=%s, Code=%s, Position=%s by user %s",
				tool.Type, tool.Code, tool.Position, user.Name)
			props.Error = "Tool bereits vorhanden"
		} else {
			logger.HTMXHandlerTools().Error("Failed to add tool (Type=%s, Code=%s) by user %s: %v",
				tool.Type, tool.Code, user.Name, err)
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to add tool: "+err.Error())
		}
	} else {
		dbElapsed := time.Since(dbStart)
		totalElapsed := time.Since(start)
		logger.HTMXHandlerTools().Info("Successfully created tool ID %d (Type=%s, Code=%s) by user %s in %v (db: %v)",
			t.ID, tool.Type, tool.Code, user.Name, totalElapsed, dbElapsed)
		props.Close = true
		props.ToolID = t.ID
	}

	return h.handleEdit(c, props)
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get user from context for tool update: %v", err)
		return err
	}

	remoteIP := c.RealIP()
	toolID, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		logger.HTMXHandlerTools().Error("Invalid tool ID parameter for update from %s: %v", remoteIP, err)
		return err
	}

	logger.HTMXHandlerTools().Info("User %s (ID: %d) updating tool %d from %s", user.Name, user.TelegramID, toolID, remoteIP)

	start := time.Now()
	formData, err := h.getToolFormData(c)
	if err != nil {
		logger.HTMXHandlerTools().Warn("Invalid form data for tool %d update by user %s: %v", toolID, user.Name, err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to get tool form data: "+err.Error())
	}

	props := &dialogs.EditToolProps{
		ToolID:              toolID,
		InputPosition:       string(formData.Position),
		InputWidth:          formData.Format.Width,
		InputHeight:         formData.Format.Height,
		InputType:           formData.Type,
		InputCode:           formData.Code,
		InputPressSelection: formData.Press,
	}

	tool := toolmodels.New(formData.Position)
	tool.ID = toolID
	tool.Format = formData.Format
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Press = formData.Press

	logger.HTMXHandlerTools().Info("Updating tool %d: Type=%s, Code=%s, Position=%s, Format=%dx%d by user %s",
		toolID, tool.Type, tool.Code, tool.Position, tool.Format.Width, tool.Format.Height, user.Name)

	dbStart := time.Now()
	if err := h.DB.Tools.Update(tool, user); err != nil {
		if err == dberror.ErrAlreadyExists {
			logger.HTMXHandlerTools().Warn("Tool update failed - already exists: ID=%d, Type=%s, Code=%s by user %s",
				toolID, tool.Type, tool.Code, user.Name)
			props.Error = "Tool bereits vorhanden"
		} else {
			logger.HTMXHandlerTools().Error("Failed to update tool %d (Type=%s, Code=%s) by user %s: %v",
				toolID, tool.Type, tool.Code, user.Name, err)
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to update tool: "+err.Error())
		}
	} else {
		dbElapsed := time.Since(dbStart)
		totalElapsed := time.Since(start)
		logger.HTMXHandlerTools().Info("Successfully updated tool %d (Type=%s, Code=%s) by user %s in %v (db: %v)",
			toolID, tool.Type, tool.Code, user.Name, totalElapsed, dbElapsed)
		props.Close = true
	}

	return h.handleEdit(c, props)
}

func (h *Tools) handleDelete(c echo.Context) error {
	remoteIP := c.RealIP()

	// Get tool ID from query parameter
	toolID, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		logger.HTMXHandlerTools().Error("Invalid tool ID parameter for deletion from %s: %v", remoteIP, err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing id parameter: "+err.Error())
	}

	// Get user from context for audit trail
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get user from context for tool deletion: %v", err)
		return err
	}

	// Get tool info before deletion for better logging
	tool, err := h.DB.Tools.Get(toolID)
	var toolInfo string
	if err != nil {
		logger.HTMXHandlerTools().Warn("Could not get tool info before deletion (ID: %d): %v", toolID, err)
		toolInfo = fmt.Sprintf("ID %d", toolID)
	} else {
		toolInfo = fmt.Sprintf("ID %d (%s %s)", toolID, tool.Type, tool.Code)
	}

	logger.HTMXHandlerTools().Info("User %s (ID: %d) deleting tool %s from %s", user.Name, user.TelegramID, toolInfo, remoteIP)

	start := time.Now()
	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		logger.HTMXHandlerTools().Error("Failed to delete tool %s by user %s: %v", toolInfo, user.Name, err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to delete tool: "+err.Error())
	}

	elapsed := time.Since(start)
	logger.HTMXHandlerTools().Info("Successfully deleted tool %s by user %s in %v", toolInfo, user.Name, elapsed)

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Tools) getToolFormData(c echo.Context) (*ToolEditFormData, error) {
	remoteIP := c.RealIP()
	logger.HTMXHandlerTools().Debug("Parsing tool form data from %s", remoteIP)

	// Parse position with detailed validation logging
	positionFormValue := c.FormValue("position")
	var position toolmodels.Position
	switch toolmodels.Position(positionFormValue) {
	case toolmodels.PositionTop:
		position = toolmodels.PositionTop
	case toolmodels.PositionTopCassette:
		position = toolmodels.PositionTopCassette
	case toolmodels.PositionBottom:
		position = toolmodels.PositionBottom
	default:
		logger.HTMXHandlerTools().Warn("Invalid position value from %s: %s", remoteIP, positionFormValue)
		return nil, errors.New("invalid position: " + positionFormValue)
	}

	data := &ToolEditFormData{
		Position: position,
	}

	// Parse width and height with validation
	widthStr := c.FormValue("width")
	if widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			logger.HTMXHandlerTools().Warn("Invalid width value from %s: %s", remoteIP, widthStr)
			return nil, errors.New("invalid width: " + err.Error())
		}
		if width <= 0 || width > 10000 {
			logger.HTMXHandlerTools().Warn("Width out of range from %s: %d", remoteIP, width)
			return nil, errors.New("width must be between 1 and 10000")
		}
		data.Format.Width = width
	}

	heightStr := c.FormValue("height")
	if heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			logger.HTMXHandlerTools().Warn("Invalid height value from %s: %s", remoteIP, heightStr)
			return nil, errors.New("invalid height: " + err.Error())
		}
		if height <= 0 || height > 10000 {
			logger.HTMXHandlerTools().Warn("Height out of range from %s: %d", remoteIP, height)
			return nil, errors.New("height must be between 1 and 10000")
		}
		data.Format.Height = height
	}

	// Parse type with validation
	data.Type = strings.TrimSpace(c.FormValue("type"))
	if data.Type == "" {
		logger.HTMXHandlerTools().Warn("Empty type field from %s", remoteIP)
		return nil, errors.New("type is required")
	}
	if len(data.Type) > 50 {
		logger.HTMXHandlerTools().Warn("Type too long from %s: %d characters", remoteIP, len(data.Type))
		return nil, errors.New("type must be 50 characters or less")
	}

	// Parse code with validation
	data.Code = strings.TrimSpace(c.FormValue("code"))
	if data.Code == "" {
		logger.HTMXHandlerTools().Warn("Empty code field from %s", remoteIP)
		return nil, errors.New("code is required")
	}
	if len(data.Code) > 50 {
		logger.HTMXHandlerTools().Warn("Code too long from %s: %d characters", remoteIP, len(data.Code))
		return nil, errors.New("code must be 50 characters or less")
	}

	// Parse press selection with validation
	pressStr := c.FormValue("press-selection")
	if pressStr != "" {
		press, err := strconv.Atoi(pressStr)
		if err != nil {
			logger.HTMXHandlerTools().Warn("Invalid press number from %s: %s", remoteIP, pressStr)
			return nil, errors.New("invalid press number: " + err.Error())
		}

		pn := toolmodels.PressNumber(press)
		data.Press = &pn
		if !toolmodels.IsValidPressNumber(data.Press) {
			logger.HTMXHandlerTools().Warn("Press number out of range from %s: %d", remoteIP, press)
			return nil, errors.New("invalid press number: must be 0, 2, 3, 4, or 5")
		}
	}

	logger.HTMXHandlerTools().Info("Successfully parsed tool form data from %s: Type=%s, Code=%s, Position=%s, Format=%dx%d, Press=%v",
		remoteIP, data.Type, data.Code, position, data.Format.Width, data.Format.Height, data.Press)

	return data, nil
}
