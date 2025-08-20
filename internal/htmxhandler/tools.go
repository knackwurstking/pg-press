package htmxhandler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	e.GET(serverPathPrefix+"/htmx/tools/list-all", h.handleListAll)

	e.GET(serverPathPrefix+"/htmx/tools/edit", func(c echo.Context) error {
		return h.handleEdit(nil, c)
	})
	e.POST(serverPathPrefix+"/htmx/tools/edit", h.handleEditPOST)
	e.PUT(serverPathPrefix+"/htmx/tools/edit", h.handleEditPUT)
}

func (h *Tools) handleListAll(c echo.Context) error {
	// TODO: Implement list all handler

	return errors.New("under construction")
}

// handleEdit renders a dialog for editing or creating a tool
func (h *Tools) handleEdit(props *components.ToolEditDialogProps, c echo.Context) error {
	if props == nil {
		props = &components.ToolEditDialogProps{}
	}

	props.ID, _ = utils.ParseInt64Query(c, constants.QueryParamID)
	props.Close = utils.ParseBoolQuery(c, constants.QueryParamClose)

	toolEdit := components.ToolEditDialog(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}
	return nil
}

func (h *Tools) handleEditPOST(c echo.Context) error {
	tool, err := h.getToolFromForm(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool from form: "+err.Error())
	}
	logger.Tools().Debug("Received tool data: %#v", tool)

	return errors.New("under construction")
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	tool, err := h.getToolFromForm(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool from form: "+err.Error())
	}
	logger.Tools().Debug("Received tool data: %#v", tool)

	return errors.New("under construction")
}

func (h *Tools) getToolFromForm(c echo.Context) (*database.Tool, error) {
	tool := database.NewTool()

	switch position := c.FormValue("position"); position {
	case database.PositionTop:
		tool.Position = database.PositionTop
	case database.PositionBottom:
		tool.Position = database.PositionBottom
	default:
		return nil, errors.New("invalid position")
	}

	// Parse width and height
	widthStr := c.FormValue("width")
	if widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			return nil, errors.New("invalid width: " + err.Error())
		}
		tool.Format.Width = width
	}

	heightStr := c.FormValue("height")
	if heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			return nil, errors.New("invalid height: " + err.Error())
		}
		tool.Format.Height = height
	}

	// Parse type
	tool.Type = c.FormValue("type")
	if tool.Type == "" {
		return nil, errors.New("type is required")
	}

	// Parse code
	tool.Code = c.FormValue("code")
	if tool.Code == "" {
		return nil, errors.New("code is required")
	}

	return tool, nil
}
