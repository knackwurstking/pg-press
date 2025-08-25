package htmxhandler

import (
	"errors"
	"net/http"
	"strconv"

	"time"

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

	e.GET(serverPathPrefix+"/htmx/tools/total-cycles", h.handleTotalCycles)
	e.GET(serverPathPrefix+"/htmx/tools/total-cycles/", h.handleTotalCycles)

	e.DELETE(serverPathPrefix+"/htmx/tools/delete", h.handleDelete)
	e.DELETE(serverPathPrefix+"/htmx/tools/delete/", h.handleDelete)

	e.GET(serverPathPrefix+"/htmx/tools/cycles", h.handleCycles)
	e.GET(serverPathPrefix+"/htmx/tools/cycles/", h.handleCycles)

	// TODO: Add "/htmx/tools/cycles/edit?tool_id=%d"
	// TODO: Add "/htmx/tools/cycles/delete?tool_id=%d"
}

func (h *Tools) handleListAll(c echo.Context) error {
	logger.HTMXHandlerTools().Debug("Fetching all tools with notes")

	// Get tools from database
	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to fetch tools: %v", err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Retrieved %d tools", len(tools))

	toolsListAll := components.ToolsListAll(tools)
	if err := toolsListAll.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTools().Error("Failed to render tools list: %v", err)
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
			logger.HTMXHandlerTools().Debug("Editing tool with ID %d", props.ID)
			// TODO: Get tool from database tools
		} else {
			logger.HTMXHandlerTools().Debug("Creating new tool")
		}
	}

	toolEdit := components.ToolEditDialog(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTools().Error("Failed to render tool edit dialog: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool edit dialog: "+err.Error())
	}
	return nil
}

func (h *Tools) handleEditPOST(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s is creating a new tool", user.UserName)

	tool, err := h.getToolFromForm(c, user)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get tool from form: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool from form: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Adding tool: Type=%s, Code=%s, Position=%s",
		tool.Type, tool.Code, tool.Position)

	if _, err := h.DB.ToolsHelper.AddWithNotes(tool, user); err != nil {
		logger.HTMXHandlerTools().Error("Failed to add tool: %v", err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to add tool: "+err.Error())
	}

	logger.HTMXHandlerTools().Info("Successfully created tool with ID %d", tool.ID)

	return h.handleEdit(c, &components.ToolEditDialogProps{
		ID:    tool.ID,
		Close: true,
	})
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s is updating a tool", user.UserName)

	tool, err := h.getToolFromForm(c, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool from form: "+err.Error())
	}
	logger.HTMXHandlerTools().Debug("Received tool data: %#v", tool)

	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("Updating tool %d", id)
	// TODO: Update tool in database tools

	return h.handleEdit(c, &components.ToolEditDialogProps{
		ID:    id,
		Close: true,
	})
}

func (h *Tools) getToolFromForm(c echo.Context, user *database.User) (*database.Tool, error) {
	logger.HTMXHandlerTools().Debug("Parsing tool form data")

	var position database.Position
	switch positionFormValue := c.FormValue("position"); database.Position(positionFormValue) {
	case database.PositionTop:
		position = database.PositionTop
	case database.PositionBottom:
		position = database.PositionBottom
	default:
		logger.HTMXHandlerTools().Error("Invalid position value: %s", positionFormValue)
		return nil, errors.New("invalid position")
	}

	tool := database.NewTool(position)

	// Parse width and height
	widthStr := c.FormValue("width")
	if widthStr != "" {
		width, err := strconv.Atoi(widthStr)
		if err != nil {
			logger.HTMXHandlerTools().Error("Invalid width value: %s", widthStr)
			return nil, errors.New("invalid width: " + err.Error())
		}
		tool.Format.Width = width
	}

	heightStr := c.FormValue("height")
	if heightStr != "" {
		height, err := strconv.Atoi(heightStr)
		if err != nil {
			logger.HTMXHandlerTools().Error("Invalid height value: %s", heightStr)
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

	tool.Mods.Add(user, database.ToolMod{})

	logger.HTMXHandlerTools().Debug("Successfully parsed tool: Type=%s, Code=%s, Position=%s, Format=%dx%d",
		tool.Type, tool.Code, position, tool.Format.Width, tool.Format.Height)

	return tool, nil
}

func (h *Tools) handleCycles(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Get tool ID from query parameter
	toolID, err := utils.ParseInt64Query(c, "tool_id")
	if err != nil {
		logger.HTMXHandlerTools().Error("Invalid tool_id parameter: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing tool_id parameter: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Fetching cycles for tool %d", toolID)

	// Get press cycles for this tool
	cycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get press cycles for tool %d: %v", toolID, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// Get regenerations for this tool
	regenerations, err := h.DB.ToolRegenerations.GetRegenerationHistory(toolID)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get regenerations for tool %d: %v", toolID, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tool regenerations: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Found %d cycles and %d regenerations for tool %d",
		len(cycles), len(regenerations), toolID)

	// Render the component
	cyclesRows := components.ToolCyclesTableRows(user, cycles, regenerations)
	if err := cyclesRows.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTools().Error("Failed to render tool cycles: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool cycles: "+err.Error())
	}

	return nil
}

func (h *Tools) handleTotalCycles(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := utils.ParseInt64Query(c, "tool_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing tool_id parameter: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Calculating total cycles for tool %d", toolID)

	// Get the last regeneration for this tool
	lastRegeneration, err := h.DB.ToolRegenerations.GetLastRegeneration(toolID)
	if err != nil && err != database.ErrNotFound {
		logger.HTMXHandlerTools().Error(
			"Failed to get last regeneration for tool %d: %v",
			toolID, err,
		)

		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get last regeneration: "+err.Error())
	}

	// Calculate total cycles since last regeneration
	var lastRegenerationDate *time.Time
	if lastRegeneration != nil {
		lastRegenerationDate = &lastRegeneration.RegeneratedAt
	}

	totalCycles, err := h.DB.PressCycles.GetTotalCyclesSinceRegeneration(toolID, lastRegenerationDate)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get total cycles for tool %d: %v", toolID, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get total cycles: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Tool %d has %d total cycles since last regeneration", toolID, totalCycles)

	// Render the component
	totalCyclesComponent := components.ToolTotalCycles(totalCycles)
	if err := totalCyclesComponent.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTools().Error("Failed to render total cycles: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render total cycles: "+err.Error())
	}

	return nil
}

func (h *Tools) handleDelete(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		logger.HTMXHandlerTools().Error("Invalid id parameter: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing id parameter: "+err.Error())
	}

	// Get user from context for audit trail
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s is deleting tool %d", user.UserName, toolID)

	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		logger.HTMXHandlerTools().Error("Failed to delete tool %d: %v", toolID, err)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to delete tool: "+err.Error())
	}

	logger.HTMXHandlerTools().Info("Successfully deleted tool %d", toolID)

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", serverPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}
