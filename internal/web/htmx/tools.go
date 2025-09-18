package htmx

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
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage"
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

// TODO: Continue here...

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/list", h.handleList),

			// Get, Post or Edit a tool
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/edit", h.handleEditGET),
			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/edit", h.handleEditPOST),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/edit", h.handleEditPUT),

			// Delete a tool
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete", h.handleDelete),
		},
	)
}

func (h *Tools) handleList(c echo.Context) error {
	start := time.Now()
	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	dbElapsed := time.Since(start)
	if dbElapsed > 100*time.Millisecond {
		logger.HTMXHandlerTools().Warn("Slow tools query took %v for %d tools", dbElapsed, len(tools))
	}

	toolsList := toolspage.ListTools(tools)
	if err := toolsList.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools list all: "+err.Error())
	}

	return nil
}

func (h *Tools) handleEditGET(c echo.Context) error {
	logger.HTMXHandlerTools().Debug("Rendering edit tool dialog")

	props := &dialogs.EditToolProps{}

	toolID, _ := helpers.ParseInt64Query(c, "id")
	if toolID > 0 {
		var err error
		props.Tool, err = h.DB.Tools.Get(toolID)
		if err != nil {
			return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
				"failed to get tool from database: "+err.Error())
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
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}

	return nil
}

func (h *Tools) handleEditPOST(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Debug("User %s creating new tool", user.Name)

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

	props.Tool = models.NewTool(formData.Position)
	props.Tool.Format = formData.Format
	props.Tool.Type = formData.Type
	props.Tool.Code = formData.Code
	props.Tool.Press = formData.Press

	if t, err := h.DB.Tools.AddWithNotes(props.Tool, user); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to add tool: "+err.Error())
	} else {
		logger.HTMXHandlerTools().Info("Created tool ID %d (Type=%s, Code=%s) by user %s",
			t.ID, props.Tool.Type, props.Tool.Code, user.Name)
	}

	return nil
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	toolID, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Warn("User %s updating tool %d", user.Name, toolID)

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

	props.Tool = models.NewTool(formData.Position)
	props.Tool.ID = toolID
	props.Tool.Format = formData.Format
	props.Tool.Type = formData.Type
	props.Tool.Code = formData.Code
	props.Tool.Press = formData.Press

	if err := h.DB.Tools.Update(props.Tool, user); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to update tool: "+err.Error())
	} else {
		logger.HTMXHandlerTools().Info("Updated tool %d (Type=%s, Code=%s) by user %s",
			props.Tool.ID, props.Tool.Type, props.Tool.Code, user.Name)
	}

	return nil
}

func (h *Tools) handleDelete(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing id parameter: "+err.Error())
	}

	// Get user from context for audit trail
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Debug("User %s deleting tool %d", user.Name, toolID)

	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to delete tool: "+err.Error())
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Tools) getToolFormData(c echo.Context) (*ToolEditFormData, error) {
	// Parse position with validation
	positionFormValue := c.FormValue("position")
	var position models.Position
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

		pn := models.PressNumber(press)
		data.Press = &pn
		if !models.IsValidPressNumber(data.Press) {
			return nil, errors.New("invalid press number: must be 0, 2, 3, 4, or 5")
		}
	}

	return data, nil
}
