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
	e.GET(serverPathPrefix+"/htmx/tools/list-all/", h.handleListAll)

	e.GET(serverPathPrefix+"/htmx/tools/edit", func(c echo.Context) error {
		return h.handleEdit(c, nil)
	})
	e.GET(serverPathPrefix+"/htmx/tools/edit/", func(c echo.Context) error {
		return h.handleEdit(c, nil)
	})
	e.POST(serverPathPrefix+"/htmx/tools/edit", h.handleEditPOST)
	e.POST(serverPathPrefix+"/htmx/tools/edit/", h.handleEditPOST)
	e.PUT(serverPathPrefix+"/htmx/tools/edit", h.handleEditPUT)
	e.PUT(serverPathPrefix+"/htmx/tools/edit/", h.handleEditPUT)
}

func (h *Tools) handleListAll(c echo.Context) error {
	// Get tools from database
	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	toolsListAll := components.ToolsListAll(tools)
	if err := toolsListAll.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools list all: "+err.Error())
	}
	return nil
}

// handleEdit renders a dialog for editing or creating a tool
func (h *Tools) handleEdit(c echo.Context, props *components.ToolEditDialogProps) error {
	if props == nil {
		props = &components.ToolEditDialogProps{}
		props.ID, _ = utils.ParseInt64Query(c, constants.QueryParamID)
		props.Close = utils.ParseBoolQuery(c, constants.QueryParamClose)

		if props.ID > 0 {
			// TODO: Get tool from database tools
		}
	}

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

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	if _, err := h.DB.ToolsHelper.AddWithNotes(tool, user); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to add tool: "+err.Error())
	}

	return h.handleEdit(c, &components.ToolEditDialogProps{
		ID:    tool.ID,
		Close: true,
	})
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	tool, err := h.getToolFromForm(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool from form: "+err.Error())
	}
	logger.Tools().Debug("Received tool data: %#v", tool)

	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	// TODO: Update tool in database tools

	return h.handleEdit(c, &components.ToolEditDialogProps{
		ID:    id,
		Close: true,
	})
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
