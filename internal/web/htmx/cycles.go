package htmx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/components/dialogs"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage/toolpage"

	toolscomp "github.com/knackwurstking/pgpress/internal/web/templates/components/tools"

	"github.com/labstack/echo/v4"
)

type Cycles struct {
	DB *database.DB
}

func (h *Cycles) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// Cycles table rows
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles", h.handleSection),
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles", h.handleTotalCycles),

			// Get, add or edit a cycles table entry
			helpers.NewEchoRoute(
				http.MethodGet, "/htmx/tools/cycle/edit",
				func(c echo.Context) error {
					return h.handleEditGET(nil, c)
				},
			),

			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit", h.handleEditPOST),
			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit", h.handleEditPUT),

			// Delete a cycle table entry
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete", h.handleDELETE),
		},
	)
}

func (h *Cycles) handleSection(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	toolID, err := helpers.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "tool_id parameter is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err), "failed to get tool: "+err.Error())
	}

	toolCycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err), "failed to get press cycles: "+err.Error())
	}

	filteredCycles := models.FilterByToolPosition(
		tool.Position,
		toolCycles...,
	)

	regeneration, err := h.DB.ToolRegenerations.GetLastRegeneration(toolID)
	if err != nil {
		logger.HTMXHandlerTools().Error("Failed to get regenerations for tool %d: %v", toolID, err)
	}

	totalCycles := h.getTotalCycles(
		toolID,
		filteredCycles...,
	)

	cyclesSection := toolpage.CyclesSection(&toolpage.CyclesSectionProps{
		User:             user,
		Tool:             tool,
		TotalCycles:      totalCycles,
		Cycles:           filteredCycles,
		LastRegeneration: regeneration,
	})

	if err := cyclesSection.Render(
		c.Request().Context(),
		c.Response(),
	); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render tool cycles: "+err.Error())
	}

	return nil
}

func (h *Cycles) handleTotalCycles(c echo.Context) error {
	// Get tool and position parameters
	toolID, err := helpers.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "tool_id parameter is required")
	}
	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to get tool")
	}

	// Get cycles for this specific tool
	toolCycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// Filter cycles by position
	filteredCycles := models.FilterByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	return toolscomp.TotalCycles(
		totalCycles,
		helpers.ParseBoolQuery(c, constants.QueryParamInput),
	).Render(c.Request().Context(), c.Response())
}

func (h *Cycles) handleEditGET(props *dialogs.EditCycleProps, c echo.Context) error {
	if props == nil {
		props = &dialogs.EditCycleProps{}
	}

	// Get tool and position from query if not already set
	if props.Tool == nil {
		tool, err := h.getToolFromQuery(c)
		if err != nil {
			return err
		}
		props.Tool = tool
	}

	cycleID, err := helpers.ParseInt64Query(c, "id")
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

	cycleEditDialog := dialogs.EditCycle(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to render cycle edit dialog: "+err.Error())
	}

	return nil
}

func (h *Cycles) handleEditPOST(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	tool, err := h.getToolFromQuery(c)
	if err != nil {
		return err
	}

	// Parse form data
	form, err := h.getCycleFormData(c)
	if err != nil {
		return h.handleEditGET(&dialogs.EditCycleProps{
			Tool:  tool,
			Error: err.Error(),
		}, c)
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return h.handleEditGET(&dialogs.EditCycleProps{
			Tool:             tool,
			Error:            "press_number must be a valid integer",
			InputTotalCycles: form.TotalCycles,
			InputPressNumber: form.PressNumber,
			OriginalDate:     &form.Date,
		}, c)
	}

	pressCycle := models.NewCycle(*form.PressNumber, tool.ID, tool.Position, form.TotalCycles, user.TelegramID)
	pressCycle.Date = form.Date

	cycleID, err := h.DB.PressCycles.Add(pressCycle, user)
	if err != nil {
		return h.handleEditGET(&dialogs.EditCycleProps{
			Tool:             tool,
			Error:            err.Error(),
			InputTotalCycles: form.TotalCycles,
			InputPressNumber: form.PressNumber,
			OriginalDate:     &form.Date,
		}, c)
	}

	// Handle regeneration if requested
	if form.Regenerating {
		if _, err := h.DB.ToolRegenerations.AddToolRegeneration(cycleID, tool.ID, "", user); err != nil {
			logger.HTMXHandlerTools().Error("Failed to start regeneration for tool %d: %v", tool.ID, err)
		}
	}

	return nil
}

func (h *Cycles) handleEditPUT(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	cycleID, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	tool, err := h.getToolFromQuery(c)
	if err != nil {
		return err
	}

	form, err := h.getCycleFormData(c)
	if err != nil {
		return h.handleEditGET(&dialogs.EditCycleProps{
			Tool:    tool,
			CycleID: cycleID,
			Error:   err.Error(),
		}, c)
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return h.handleEditGET(&dialogs.EditCycleProps{
			Tool:             tool,
			CycleID:          cycleID,
			Error:            "press_number must be a valid integer",
			InputTotalCycles: form.TotalCycles,
			InputPressNumber: form.PressNumber,
			OriginalDate:     &form.Date,
		}, c)
	}

	// Update the cycle
	pressCycle := models.NewPressCycleWithID(cycleID, *form.PressNumber, tool.ID, tool.Position, form.TotalCycles, user.TelegramID, form.Date)

	if err := h.DB.PressCycles.Update(pressCycle, user); err != nil {
		return h.handleEditGET(&dialogs.EditCycleProps{
			Tool:             tool,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputTotalCycles: form.TotalCycles,
			InputPressNumber: form.PressNumber,
			OriginalDate:     &form.Date,
		}, c)
	}

	// Handle regeneration if requested
	if form.Regenerating {
		if _, err := h.DB.ToolRegenerations.AddToolRegeneration(cycleID, tool.ID, "", user); err != nil {
			logger.HTMXHandlerTools().Error("Failed to start regeneration for tool %d: %v", tool.ID, err)
		}

		if err := h.DB.ToolRegenerations.StopToolRegeneration(tool.ID, user); err != nil {
			logger.HTMXHandlerTools().Error("Failed to stop regeneration for tool %d: %v", tool.ID, err)
		}
	}

	return nil
}

func (h *Cycles) handleDELETE(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	cycleID, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Debug("Handling cycle deletion request for ID %d", cycleID)

	if err := h.DB.PressCycles.Delete(cycleID, user); err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err), "failed to delete press cycle: "+err.Error())
	}

	return h.handleSection(c)
}

// getTotalCycles calculates total cycles from a list of cycles
func (h *Cycles) getTotalCycles(toolID int64, cycles ...*models.Cycle) int64 {
	// Get regeneration for this tool
	var startCycleID int64
	if r, err := h.DB.ToolRegenerations.GetLastRegeneration(toolID); err == nil {
		startCycleID = r.CycleID
	}

	var totalCycles int64

	for _, cycle := range cycles {
		if cycle.ID <= startCycleID {
			continue
		}

		totalCycles += cycle.PartialCycles
	}

	return totalCycles
}

// getToolFromQuery extracts tool and tool position from query parameters
func (h *Cycles) getToolFromQuery(c echo.Context) (*models.Tool, error) {
	toolID, err := helpers.ParseInt64Query(c, constants.QueryParamToolID)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "tool_id parameter is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return nil, echo.NewHTTPError(dberror.GetHTTPStatusCode(err), "failed to get tool: "+err.Error())
	}

	return tool, nil
}

// getCycleFormData parses form data for cycle operations
func (h *Cycles) getCycleFormData(c echo.Context) (*CycleEditFormData, error) {
	form := &CycleEditFormData{}

	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "press_number must be an integer")
		}
		pn := models.PressNumber(press)
		form.PressNumber = &pn
	}

	if dateString := c.FormValue("original_date"); dateString != "" {
		var err error
		form.Date, err = time.Parse(constants.DateFormat, dateString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid date format: "+err.Error())
		}
	} else {
		form.Date = time.Now()
	}

	if totalCyclesString := c.FormValue("total_cycles"); totalCyclesString == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "total_cycles is required")
	} else {
		var err error
		form.TotalCycles, err = strconv.ParseInt(totalCyclesString, 10, 64)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "total_cycles must be an integer")
		}
	}

	form.Regenerating = c.FormValue("regenerating") != ""

	return form, nil
}
