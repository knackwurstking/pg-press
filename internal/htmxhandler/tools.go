package htmxhandler

import (
	"errors"
	"net/http"
	"strconv"

	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	toolscomp "github.com/knackwurstking/pgpress/internal/templates/components/tools"
	"github.com/knackwurstking/pgpress/internal/utils"
	"github.com/labstack/echo/v4"
)

type Tools struct {
	DB *database.DB
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			utils.NewEchoRoute(http.MethodGet, "/htmx/tools/list-all", h.handleListAll),

			// Get, Post or Edit a tool
			utils.NewEchoRoute(http.MethodGet, "/htmx/tools/edit", func(c echo.Context) error {
				return h.handleEdit(c, nil)
			}),

			utils.NewEchoRoute(http.MethodPost, "/htmx/tools/edit", h.handleEditPOST),
			utils.NewEchoRoute(http.MethodPut, "/htmx/tools/edit", h.handleEditPUT),

			// Delete a tool
			utils.NewEchoRoute(http.MethodDelete, "/htmx/tools/delete", h.handleDelete),

			// Cycles table rows
			utils.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles", h.handleCyclesSection),

			// Get, add or edit a cycles table entry
			utils.NewEchoRoute(http.MethodGet, "/htmx/tools/cycle/edit", func(c echo.Context) error {
				return h.handleCycleEditGET(nil, c)
			}),
			utils.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit", h.handleCycleEditPOST),
			utils.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit", h.handleCycleEditPUT),

			// Delete a cycle table entry
			utils.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete", h.handleCycleDELETE),
		},
	)
}

func (h *Tools) handleListAll(c echo.Context) error {
	logger.HTMXHandlerTools().Debug("Fetching all tools with notes")

	// Get tools from database
	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Retrieved %d tools", len(tools))

	toolsListAll := toolscomp.ListAll(tools)
	if err := toolsListAll.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools list all: "+err.Error())
	}
	return nil
}

// handleEdit renders a dialog for editing or creating a tool
func (h *Tools) handleEdit(c echo.Context, props *toolscomp.EditDialogProps) error {
	if props == nil {
		props = &toolscomp.EditDialogProps{}
		props.ID, _ = utils.ParseInt64Query(c, constants.QueryParamID)
		props.Close = utils.ParseBoolQuery(c, constants.QueryParamClose)

		if props.ID > 0 {
			logger.HTMXHandlerTools().Debug("Editing tool with ID %d", props.ID)
			// TODO: Get tool from database tools
		} else {
			logger.HTMXHandlerTools().Debug("Creating new tool")
		}
	}

	toolEdit := toolscomp.EditDialog(props)
	if err := toolEdit.Render(c.Request().Context(), c.Response()); err != nil {
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

	tool, err := h.getToolFormData(c, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool form data: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Adding tool: Type=%s, Code=%s, Position=%s",
		tool.Type, tool.Code, tool.Position)

	if _, err := h.DB.ToolsHelper.AddWithNotes(tool, user); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to add tool: "+err.Error())
	}

	logger.HTMXHandlerTools().Info("Successfully created tool with ID %d", tool.ID)

	return h.handleEdit(c, &toolscomp.EditDialogProps{
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

	tool, err := h.getToolFormData(c, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool form data: "+err.Error())
	}
	logger.HTMXHandlerTools().Debug("Received tool data: %#v", tool)

	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("Updating tool %d", id)
	// TODO: Update tool in database tools

	return h.handleEdit(c, &toolscomp.EditDialogProps{
		ID:    id,
		Close: true,
	})
}

func (h *Tools) handleDelete(c echo.Context) error {
	// Get tool ID from query parameter
	toolID, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
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
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to delete tool: "+err.Error())
	}

	logger.HTMXHandlerTools().Info("Successfully deleted tool %d", toolID)

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", constants.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

// getToolFormData parses the tool form data from the request context. [POST/PUT /tools/edit]
func (h *Tools) getToolFormData(c echo.Context, user *database.User) (*database.Tool, error) {
	logger.HTMXHandlerTools().Debug("Parsing tool form data")

	var position database.Position
	switch positionFormValue := c.FormValue("position"); database.Position(positionFormValue) {
	case database.PositionTop:
		position = database.PositionTop
	case database.PositionBottom:
		position = database.PositionBottom
	default:
		return nil, errors.New("invalid position")
	}

	tool := database.NewTool(position)

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

	tool.Mods.Add(user, database.ToolMod{})

	logger.HTMXHandlerTools().Debug("Successfully parsed tool: Type=%s, Code=%s, Position=%s, Format=%dx%d",
		tool.Type, tool.Code, position, tool.Format.Width, tool.Format.Height)

	return tool, nil
}

func (h *Tools) handleCyclesSection(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Get tool ID from query parameter
	toolID, err := utils.ParseInt64Query(c, "tool_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing tool_id parameter: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Fetching cycles for tool %d", toolID)

	// Get press cycles for this tool
	cycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// Get regenerations for this tool
	regenerations, err := h.DB.ToolRegenerations.GetRegenerationHistory(toolID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tool regenerations: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Found %d cycles and %d regenerations for tool %d",
		len(cycles), len(regenerations), toolID)

	// Calculate total cycles since last regeneration
	var lastRegenerationDate *time.Time
	if len(regenerations) > 0 {
		lastRegenerationDate = &regenerations[len(regenerations)-1].RegeneratedAt
	}

	totalCycles, err := h.DB.PressCycles.GetTotalCyclesSinceRegeneration(
		toolID, lastRegenerationDate,
	)

	// Render the component
	cyclesSection := toolscomp.CyclesSection(&toolscomp.CyclesSectionProps{
		User:          user,
		ToolID:        toolID,
		TotalCycles:   totalCycles,
		Cycles:        cycles,
		Regenerations: regenerations,
	})
	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool cycles: "+err.Error())
	}

	return nil
}

// handleCycleEditGET "/htmx/tools/cycle/edit?tool_id=%d?cycle_id=%d" cycle_id is optional and only required for editing a cycle
func (h *Tools) handleCycleEditGET(props *toolscomp.CycleEditDialogProps, c echo.Context) error {
	if props == nil {
		props = &toolscomp.CycleEditDialogProps{}
	}

	toolID, err := utils.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return err
	}
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}
	props.Tool = tool

	close := utils.ParseBoolQuery(c, constants.QueryParamClose)
	if close {
		props.Close = true

		cycleEditDialog := toolscomp.CycleEditDialog(props)
		if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to close cycle edit dialog: "+err.Error())
		}
		return nil
	}

	// TODO: Get tool data from the database

	cycleID, err := utils.ParseInt64Query(c, constants.QueryParamCycleID)
	if err == nil {
		props.CycleID = cycleID
		// TODO: Get cycle data from the database
	}

	logger.HTMXHandlerTools().Debug(
		"Handling cycle edit GET request for tool %d and cycle %d",
		toolID, cycleID,
	)

	cycleEditDialog := toolscomp.CycleEditDialog(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cycle edit dialog: "+err.Error())
	}

	return nil
}

// TODO: Add "/htmx/tools/cycle/edit?tool_id=%d"
func (h *Tools) handleCycleEditPOST(c echo.Context) error {
	toolID, err := utils.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return err
	}
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	// TODO: Parse form data (type: PressCycle)

	logger.HTMXHandlerTools().Debug(
		"Handling cycle edit POST request for tool %d",
		toolID,
	)

	return errors.New("under construction")
}

// TODO: Add "PUT    /htmx/tools/cycle/edit?cycle_id=%d"
func (h *Tools) handleCycleEditPUT(c echo.Context) error {
	cycleID, err := utils.ParseInt64Query(c, constants.QueryParamCycleID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Debug(
		"Handling cycle edit PUT request for cycle %d",
		cycleID,
	)

	return errors.New("under construction")
}

// TODO: Add "DELETE /htmx/tools/cycle/delete?cycle_id=%d"
func (h *Tools) handleCycleDELETE(c echo.Context) error {
	cycleID, err := utils.ParseInt64Query(c, constants.QueryParamCycleID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Debug(
		"Handling cycle delete request for cycle %d",
		cycleID,
	)

	return errors.New("under construction")
}
