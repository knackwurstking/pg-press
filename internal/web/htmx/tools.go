package htmx

import (
	"errors"
	"net/http"
	"strconv"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	tooltemplates "github.com/knackwurstking/pgpress/internal/web/templates/components/tools"
	"github.com/knackwurstking/pgpress/internal/web/webhelpers"
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

			// Cycles table rows
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles", h.handleCyclesSection),
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles", h.handleTotalCycles),

			// Get, add or edit a cycles table entry
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycle/edit", func(c echo.Context) error {
				return h.handleCycleEditGET(nil, c)
			}),
			webhelpers.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit", h.handleCycleEditPOST),
			webhelpers.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit", h.handleCycleEditPUT),

			// Delete a cycle table entry
			webhelpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete", h.handleCycleDELETE),
		},
	)
}

func (h *Tools) handleList(c echo.Context) error {
	logger.HTMXHandlerTools().Debug("Fetching all tools with notes")

	// Get tools from database
	tools, err := h.DB.ToolsHelper.ListWithNotes()
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get tools from database: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Retrieved %d tools", len(tools))

	toolsListAll := tooltemplates.List(tools)
	if err := toolsListAll.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tools list all: "+err.Error())
	}
	return nil
}

// handleEdit renders a dialog for editing or creating a tool
func (h *Tools) handleEdit(c echo.Context, props *tooltemplates.EditDialogProps) error {
	if props == nil {
		props = &tooltemplates.EditDialogProps{}
		props.ToolID, _ = webhelpers.ParseInt64Query(c, constants.QueryParamID)
		props.Close = webhelpers.ParseBoolQuery(c, constants.QueryParamClose)

		if props.ToolID > 0 {
			logger.HTMXHandlerTools().Debug("Editing tool with ID %d", props.ToolID)
			tool, err := h.DB.ToolsHelper.GetWithNotes(props.ToolID)
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
		} else {
			logger.HTMXHandlerTools().Debug("Creating new tool")
		}
	}

	toolEdit := tooltemplates.EditDialog(props)
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

	logger.HTMXHandlerTools().Info("User %s is creating a new tool", user.UserName)

	formData, err := h.getToolFormData(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool form data: "+err.Error())
	}

	props := &tooltemplates.EditDialogProps{
		InputPosition:       string(formData.Position),
		InputWidth:          formData.Format.Width,
		InputHeight:         formData.Format.Height,
		InputType:           formData.Type,
		InputCode:           formData.Code,
		InputPressSelection: formData.Press,
	}

	tool := models.NewTool(formData.Position)
	tool.Format = formData.Format
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Press = formData.Press

	logger.HTMXHandlerTools().Debug("Adding tool: Type=%s, Code=%s, Position=%s",
		tool.Type, tool.Code, tool.Position)

	if t, err := h.DB.ToolsHelper.AddWithNotes(tool, user); err != nil {
		if err == dberror.ErrAlreadyExists {
			props.Error = "Tool bereits vorhanden"
		} else {
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to add tool: "+err.Error())
		}
	} else {
		props.Close = true
		props.ReloadToolsList = true
		props.ToolID = t.ID // Yeah, there is no need to set the tool ID here
	}

	return h.handleEdit(c, props)
}

func (h *Tools) handleEditPUT(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Info("User %s is updating a tool", user.UserName)

	toolID, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	formData, err := h.getToolFormData(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to get tool form data: "+err.Error())
	}

	props := &tooltemplates.EditDialogProps{
		ToolID:              toolID,
		InputPosition:       string(formData.Position),
		InputWidth:          formData.Format.Width,
		InputHeight:         formData.Format.Height,
		InputType:           formData.Type,
		InputCode:           formData.Code,
		InputPressSelection: formData.Press,
	}

	tool := models.NewTool(formData.Position)
	tool.ID = toolID
	tool.Format = formData.Format
	tool.Type = formData.Type
	tool.Code = formData.Code
	tool.Press = formData.Press

	logger.HTMXHandlerTools().Debug("Received tool data: %#v", tool)

	if err := h.DB.Tools.Update(tool, user); err != nil {
		if err == dberror.ErrAlreadyExists {
			props.Error = "Tool bereits vorhanden"
		} else {
			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to update tool: "+err.Error())
		}
	} else {
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

	logger.HTMXHandlerTools().Info("User %s is deleting tool %d", user.UserName, toolID)

	// Delete the tool from database
	if err := h.DB.Tools.Delete(toolID, user); err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to delete tool: "+err.Error())
	}

	// Set redirect header to tools page
	c.Response().Header().Set("HX-Redirect", env.ServerPathPrefix+"/tools")
	return c.NoContent(http.StatusOK)
}

func (h *Tools) handleCyclesSection(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Get tool ID from query parameter
	toolID, err := webhelpers.ParseInt64Query(c, "tool_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing tool_id parameter: "+err.Error())
	}

	logger.HTMXHandlerTools().Debug("Fetching cycles for tool %d", toolID)
	// Get press cycles for this tool
	cycles, err := h.DB.PressCyclesHelper.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// Get partial cycles for the last entry in cycles
	var lastPartialCycles int64
	if len(cycles) > 0 {
		lastCycle := cycles[len(cycles)-1]
		lastPartialCycles = h.DB.PressCyclesHelper.GetPartialCyclesForPress(lastCycle)
	}

	// Get regenerations for this tool
	regenerations, err := h.DB.ToolRegenerations.GetRegenerationHistory(toolID)
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get tool regenerations: "+err.Error())
	}
	logger.HTMXHandlerTools().Debug("Found %d cycles and %d regenerations for tool %d",
		len(cycles), len(regenerations), toolID)

	// FIXME: Wrong calculation of total cycles
	// Calculate total cycles
	var totalCycles int64
	if len(cycles) > 0 {
		totalCycles = cycles[0].TotalCycles - (cycles[len(cycles)-1].TotalCycles - lastPartialCycles)
	}
	logger.HTMXHandlerTools().Debug("Calculated total cycles for tool %d: %d", toolID, totalCycles)

	// Render the component
	cyclesSection := tooltemplates.CyclesSection(&tooltemplates.CyclesSectionProps{
		User:              user,
		ToolID:            toolID,
		TotalCycles:       totalCycles,
		Cycles:            cycles,
		Regenerations:     regenerations,
		LastPartialCycles: lastPartialCycles,
	})
	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool cycles: "+err.Error())
	}

	return nil
}

func (h *Tools) handleTotalCycles(c echo.Context) error {
	toolID, err := webhelpers.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"invalid or missing tool_id parameter: "+err.Error())
	}

	colorClass, err := webhelpers.ParseStringQuery(c, constants.QueryParamColorClass)
	if err != nil {
		logger.HTMXHandlerTools().Warn("Failed to parse color class query parameter: %v", err)
	}

	cycles, err := h.DB.PressCyclesHelper.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// FIXME: Wrong calculation of total cycles
	// calculate total cycles
	var totalCycles int64
	for _, cycle := range cycles {
		totalCycles += cycle.TotalCycles
	}

	return tooltemplates.TotalCycles(
		totalCycles,
		webhelpers.ParseBoolQuery(c, constants.QueryParamInput),
		colorClass,
	).Render(c.Request().Context(), c.Response())
}

// handleCycleEditGET "/htmx/tools/cycle/edit?tool_id=%d?cycle_id=%d" cycle_id is optional and only required for editing a cycle
func (h *Tools) handleCycleEditGET(props *tooltemplates.CycleEditDialogProps, c echo.Context) error {
	if props == nil {
		props = &tooltemplates.CycleEditDialogProps{}
	}

	if !props.HasActiveSlot() {
		toolTop, toolTopCassette, toolBottom, err := h.getSlotsFromQuery(c)
		if err != nil {
			return err
		}
		props.SlotTop = toolTop
		props.SlotTopCassette = toolTopCassette
		props.SlotBottom = toolBottom
	}

	close := webhelpers.ParseBoolQuery(c, constants.QueryParamClose)
	if close || props.Close {
		props.Close = true

		cycleEditDialog := tooltemplates.CycleEditDialog(props)
		if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to close cycle edit dialog: "+err.Error())
		}
		return nil
	}

	cycleID, err := webhelpers.ParseInt64Query(c, constants.QueryParamCycleID)
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
			props.OriginalDate = &cycle.Date
		}
	}

	cycleEditDialog := tooltemplates.CycleEditDialog(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cycle edit dialog: "+err.Error())
	}

	return nil
}

// handleCycleEditPOST "/htmx/tools/cycle/edit?tool_id=%d"
func (h *Tools) handleCycleEditPOST(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	toolTop, toolTopCassette, toolBottom, err := h.getSlotsFromQuery(c)
	if err != nil {
		return err
	}

	// Parse form data (type: PressCycle)
	formData, err := h.getCycleFormData(c)
	if err != nil {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            err.Error(),
			InputPressNumber: nil, // Don't have form data to repopulate
		}, c)
	}

	if !models.IsValidPressNumber(formData.PressNumber) {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            "press_number must be a valid integer",
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	pressCycle := models.NewPressCycle(
		toolTop.ID, toolTopCassette.ID, toolBottom.ID,
		*formData.PressNumber,
		formData.TotalCycles,
		user.TelegramID,
	)

	if _, err := h.DB.PressCycles.Add(pressCycle, user); err != nil {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
		SlotTop:         toolTop,
		SlotTopCassette: toolTopCassette,
		SlotBottom:      toolBottom,
		Close:           true,
	}, c)
}

// handleCycleEditPUT "/htmx/tools/cycle/edit?cycle_id=%d"
func (h *Tools) handleCycleEditPUT(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	cycleID, err := webhelpers.ParseInt64Query(c, constants.QueryParamCycleID)
	if err != nil {
		return err
	}

	toolTop, toolTopCassette, toolBottom, err := h.getSlotsFromQuery(c)
	if err != nil {
		return err
	}

	formData, err := h.getCycleFormData(c)
	if err != nil {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputPressNumber: nil, // Don't have form data to repopulate
		}, c)
	}

	if !models.IsValidPressNumber(formData.PressNumber) {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            "press_number must be a valid integer",
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	// Get existing cycle to preserve its original date
	existingCycle, err := h.DB.PressCycles.Get(cycleID)
	if err != nil {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			CycleID:          cycleID,
			Error:            "Failed to get existing cycle: " + err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	// Update only the fields that should change, preserving the original date
	pressCycle := models.NewPressCycleWithID(
		cycleID,
		toolTop.ID, toolTopCassette.ID, toolBottom.ID,
		*formData.PressNumber,
		formData.TotalCycles,
		user.TelegramID,
		existingCycle.Date,
	)

	if err := h.DB.PressCycles.Update(pressCycle, user); err != nil {
		return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
		}, c)
	}

	return h.handleCycleEditGET(&tooltemplates.CycleEditDialogProps{
		SlotTop:         toolTop,
		SlotTopCassette: toolTopCassette,
		SlotBottom:      toolBottom,
		Close:           true,
	}, c)
}

// TODO: Add "DELETE /htmx/tools/cycle/delete?cycle_id=%d"
func (h *Tools) handleCycleDELETE(c echo.Context) error {
	cycleID, err := webhelpers.ParseInt64Query(c, constants.QueryParamCycleID)
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

	var position models.Position
	switch positionFormValue := c.FormValue("position"); models.Position(positionFormValue) {
	case models.PositionTop:
		position = models.PositionTop
	case models.PositionTopCassette:
		position = models.PositionTopCassette
	case models.PositionBottom:
		position = models.PositionBottom
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

	// Parse press selection
	pressStr := c.FormValue("press-selection")
	if pressStr != "" {
		press, err := strconv.Atoi(pressStr)
		if err != nil {
			return nil, errors.New("invalid press number: " + err.Error())
		}

		pn := models.PressNumber(press)
		data.Press = &pn
		if !models.IsValidPressNumber(data.Press) {
			return nil, errors.New("invalid press number")
		}
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

	var pressNumber *models.PressNumber
	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "press_number must be an integer")
		}

		pn := models.PressNumber(press)
		pressNumber = &pn
	}

	return &CycleEditFormData{
		TotalCycles: totalCycles,
		PressNumber: pressNumber,
	}, nil
}

func (h *Tools) getSlotsFromQuery(c echo.Context) (toolTop, toolTopCassette, toolBottom *models.Tool, err error) {
	slotTop, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTop)
	if err != nil {
		return nil, nil, nil, err
	}

	slotTopCassette, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTopCassette)
	if err != nil {
		return nil, nil, nil, err
	}

	slotBottom, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotBottom)
	if err != nil {
		return nil, nil, nil, err
	}

	// Validate slots, at least one must be provided
	if slotTop > 0 || slotTopCassette > 0 || slotBottom > 0 {
		return nil, nil, nil, echo.NewHTTPError(http.StatusBadRequest, "at least one slot must be provided")
	}

	// Fetching tools for slots
	if slotTop > 0 {
		toolTop, err = h.DB.Tools.Get(slotTop)
		if err != nil {
			return nil, nil, nil, echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to get tool for slot %d: "+err.Error(), slotTop)
		}
	}

	if slotTopCassette > 0 {
		toolTopCassette, err = h.DB.Tools.Get(slotTopCassette)
		if err != nil {
			return nil, nil, nil, echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to get tool for slot %d: "+err.Error(), slotTopCassette)
		}
	}

	if slotBottom > 0 {
		toolBottom, err = h.DB.Tools.Get(slotBottom)
		if err != nil {
			return nil, nil, nil, echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to get tool for slot %d: "+err.Error(), slotBottom)
		}
	}

	return toolTop, toolTopCassette, toolBottom, nil
}
