package htmx

import (
	"errors"
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
	start := time.Now()
	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	dbElapsed := time.Since(start)
	if dbElapsed > 100*time.Millisecond {
		logger.HTMXHandlerTools().Warn("Slow tools query took %v for %d tools", dbElapsed, len(tools))
	}

	toolsListAll := toolscomp.List(tools)
	if err := toolsListAll.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools list all: "+err.Error())
	}

	return nil
}

func (h *Tools) handleEdit(c echo.Context, props *dialogs.EditToolProps) error {
	if props == nil {
		props = &dialogs.EditToolProps{}
		props.ToolID, _ = webhelpers.ParseInt64Query(c, constants.QueryParamID)
		props.Close = webhelpers.ParseBoolQuery(c, constants.QueryParamClose)

		if props.ToolID > 0 {
			tool, err := h.DB.Tools.GetWithNotes(props.ToolID)
			if err != nil {
				return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
					"failed to get tool from database: "+err.Error())
			}

			props.InputPosition = string(tool.Position)
			props.InputWidth = tool.Format.Width
			props.InputHeight = tool.Format.Height
			props.InputType = tool.Type
			props.InputCode = tool.Code
			props.InputPressSelection = tool.Press
		}
	}

	toolEdit := dialogs.EditTool(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}
	return nil
}

func (h *Tools) handleEditPOST(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s creating new tool", user.Name)

	formData, err := h.getToolFormData(c)
	if err != nil {
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

	if t, err := h.DB.Tools.AddWithNotes(tool, user); err != nil {
		if err == dberror.ErrAlreadyExists {
			props.Error = "Tool bereits vorhanden"
		} else {
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to add tool: "+err.Error())
		}
	} else {
		logger.HTMXHandlerTools().Info("Created tool ID %d (Type=%s, Code=%s) by user %s",
			t.ID, tool.Type, tool.Code, user.Name)
		props.Close = true
		props.ToolID = t.ID
	}

	return h.handleEdit(c, props)
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	toolID, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s updating tool %d", user.Name, toolID)

	formData, err := h.getToolFormData(c)
	if err != nil {
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

	if err := h.DB.Tools.Update(tool, user); err != nil {
		if err == dberror.ErrAlreadyExists {
			props.Error = "Tool bereits vorhanden"
		} else {
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to update tool: "+err.Error())
		}
	} else {
		logger.HTMXHandlerTools().Info("Updated tool %d (Type=%s, Code=%s) by user %s",
			toolID, tool.Type, tool.Code, user.Name)
		props.Close = true
	}

	return h.handleEdit(c, props)
}

func (h *Tools) handleDelete(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing id parameter: "+err.Error())
	}

	// Get user from context for audit trail
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s deleting tool %d", user.Name, toolID)

	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to delete tool: "+err.Error())
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Tools) getToolFormData(c echo.Context) (*ToolEditFormData, error) {
	// Parse position with validation
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

		pn := toolmodels.PressNumber(press)
		data.Press = &pn
		if !toolmodels.IsValidPressNumber(data.Press) {
			return nil, errors.New("invalid press number: must be 0, 2, 3, 4, or 5")
		}
	}

	return data, nil
}
