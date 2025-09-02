package htmxhandler

import (
	"errors"
	"net/http"
	"strconv"

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
			utils.NewEchoRoute(http.MethodGet, "/htmx/tools/list", h.handleList),

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
			utils.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles", h.handleTotalCycles),

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

func (h *Tools) handleList(c echo.Context) error {
	logger.HTMXHandlerTools().Debug("Fetching all tools with notes")

	// Get tools from database
	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Retrieved %d tools", len(tools))

	toolsListAll := toolscomp.List(tools)
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

	formData, err := h.getToolFormData(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool form data: "+err.Error())
	}
	tool := database.NewTool(formData.Position)
	tool.Format = formData.Format
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Mods.Add(user, database.ToolMod{})

	logger.HTMXHandlerTools().Debug("Adding tool: Type=%s, Code=%s, Position=%s",
		tool.Type, tool.Code, tool.Position)

	if t, err := h.DB.ToolsHelper.AddWithNotes(tool, user); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to add tool: "+err.Error())
	} else {
		tool.ID = t.ID
	}

	logger.HTMXHandlerTools().Info("Successfully created tool with ID %d", tool.ID)

	return h.handleEdit(c, &toolscomp.EditDialogProps{
		ID:              tool.ID,
		Close:           true,
		ReloadToolsList: true,
	})
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s is updating a tool", user.UserName)

	formData, err := h.getToolFormData(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool form data: "+err.Error())
	}
	tool := database.NewTool(formData.Position)
	tool.Format = formData.Format
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Mods.Add(user, database.ToolMod{})

	logger.HTMXHandlerTools().Debug("Received tool data: %#v", tool)

	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("Updating tool %d", id)
	tool.ID = id
	if err := h.DB.Tools.Update(tool, user); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to update tool: "+err.Error())
	}

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
	cycles, err := h.DB.PressCyclesHelper.GetPressCyclesForTool(toolID)
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

	totalCycles, err := h.DB.PressCyclesHelper.GetTotalCyclesSinceRegeneration(toolID)

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

func (h *Tools) handleTotalCycles(c echo.Context) error {
	toolID, err := utils.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing tool_id parameter: "+err.Error())
	}

	colorClass, err := utils.ParseStringQuery(c, constants.QueryParamColorClass)
	if err != nil {
		logger.HTMXHandlerTools().Warn("Failed to parse color class query parameter: %v", err)
	}

	totalCycles, err := h.DB.PressCyclesHelper.GetTotalCyclesSinceRegeneration(toolID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get total cycles: "+err.Error())
	}

	return toolscomp.TotalCycles(
		totalCycles,
		utils.ParseBoolQuery(c, constants.QueryParamInput),
		colorClass,
	).Render(c.Request().Context(), c.Response())
}

// handleCycleEditGET "/htmx/tools/cycle/edit?tool_id=%d?cycle_id=%d" cycle_id is optional and only required for editing a cycle
func (h *Tools) handleCycleEditGET(props *toolscomp.CycleEditDialogProps, c echo.Context) error {
	if props == nil {
		props = &toolscomp.CycleEditDialogProps{}
	}

	if props.Tool == nil {
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
	}

	close := utils.ParseBoolQuery(c, constants.QueryParamClose)
	if close || props.Close {
		props.Close = true

		cycleEditDialog := toolscomp.CycleEditDialog(props)
		if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to close cycle edit dialog: "+err.Error())
		}
		return nil
	}

	cycleID, err := utils.ParseInt64Query(c, constants.QueryParamCycleID)
	if err == nil {
		props.CycleID = cycleID
		// Get cycle data from the database
		cycle, err := h.DB.PressCycles.Get(cycleID)
		if err != nil {
			props.Error = "Fehler beim Laden der Zyklusdaten: " + err.Error()
		} else {
			props.InputTotalCycles = cycle.TotalCycles
			pressNumber := cycle.PressNumber
			props.InputPressNumber = &pressNumber
		}
	}

	logger.HTMXHandlerTools().Debug(
		"Handling cycle edit GET request for tool %d and cycle %d",
		props.Tool.ID, cycleID,
	)

	cycleEditDialog := toolscomp.CycleEditDialog(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cycle edit dialog: "+err.Error())
	}

	return nil
}

// handleCycleEditPOST "/htmx/tools/cycle/edit?tool_id=%d"
func (h *Tools) handleCycleEditPOST(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
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

	// Parse form data (type: PressCycle)
	formData, err := h.getCycleFormData(c)
	if err != nil {
		return h.handleCycleEditGET(&toolscomp.CycleEditDialogProps{
			Tool:             tool,
			Error:            err.Error(),
			InputPressNumber: nil, // Don't have form data to repopulate
		}, c)
	}

	logger.HTMXHandlerTools().Debug(
		"Handling cycle edit POST request for tool %d: formData=%#v",
		toolID, formData,
	)

	// TODO: I need to make the press argument optional, because i will allow editing tools not active
	if _, err := h.DB.PressCycles.Add(
		database.NewPressCycle(tool.ID, *formData.PressNumber, formData.TotalCycles, user.TelegramID),
		user,
	); err != nil {
		return h.handleCycleEditGET(&toolscomp.CycleEditDialogProps{
			Tool:             tool,
			Error:            err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	return h.handleCycleEditGET(&toolscomp.CycleEditDialogProps{
		Tool:  tool,
		Close: true,
	}, c)
}

// handleCycleEditPUT "/htmx/tools/cycle/edit?cycle_id=%d"
func (h *Tools) handleCycleEditPUT(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	cycleID, err := utils.ParseInt64Query(c, constants.QueryParamCycleID)
	if err != nil {
		return err
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

	logger.HTMXHandlerTools().Info(
		"User %s is handling cycle edit PUT request for cycle %d",
		user.UserName, cycleID,
	)

	formData, err := h.getCycleFormData(c)
	if err != nil {
		return h.handleCycleEditGET(&toolscomp.CycleEditDialogProps{
			Tool:             tool,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputPressNumber: nil, // Don't have form data to repopulate
		}, c)
	}

	// TODO: I need to make the press argument optional, because i will allow editing tools not active
	if err := h.DB.PressCycles.Update(
		database.NewPressCycle(cycleID, *formData.PressNumber, formData.TotalCycles, user.TelegramID),
		user,
	); err != nil {
		return h.handleCycleEditGET(&toolscomp.CycleEditDialogProps{
			Tool:             tool,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	return h.handleCycleEditGET(&toolscomp.CycleEditDialogProps{
		Tool:  tool,
		Close: true,
	}, c)
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

// getToolFormData parses the tool form data from the request context. [POST/PUT /tools/edit]
func (h *Tools) getToolFormData(c echo.Context) (*ToolEditFormData, error) {
	logger.HTMXHandlerTools().Debug("Parsing tool form data")

	var position database.Position
	switch positionFormValue := c.FormValue("position"); database.Position(positionFormValue) {
	case database.PositionTop:
		position = database.PositionTop
	case database.PositionTopCassette:
		position = database.PositionTopCassette
	case database.PositionBottom:
		position = database.PositionBottom
	default:
		return nil, errors.New("invalid position")
	}

	data := &ToolEditFormData{
		Position: position,
	}

	// Parse width and height
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

	// Parse type
	data.Type = c.FormValue("type")
	if data.Type == "" {
		return nil, errors.New("type is required")
	}

	// Parse code
	data.Code = c.FormValue("code")
	if data.Code == "" {
		return nil, errors.New("code is required")
	}

	logger.HTMXHandlerTools().Debug("Successfully parsed tool: Type=%s, Code=%s, Position=%s, Format=%dx%d",
		data.Type, data.Code, position, data.Format.Width, data.Format.Height)

	return data, nil
}

func (h *Tools) getCycleFormData(c echo.Context) (*CycleEditFormData, error) {
	totalCyclesString := c.FormValue("total_cycles")
	if totalCyclesString == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "total_cycles is required")
	}
	totalCycles, err := strconv.ParseInt(totalCyclesString, 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "total_cycles must be an integer")
	}

	pressString := c.FormValue("press_number")
	if pressString == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "press_number is required")
	}
	press, err := strconv.Atoi(pressString)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "press_number must be an integer")
	}
	pressNumber := database.PressNumber(press)

	return &CycleEditFormData{
		TotalCycles: totalCycles,
		PressNumber: &pressNumber,
	}, nil
}
