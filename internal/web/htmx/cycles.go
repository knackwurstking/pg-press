package htmx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/components"
	"github.com/knackwurstking/pgpress/internal/web/templates/dialogs"
	"github.com/knackwurstking/pgpress/internal/web/templates/toolspage/toolpage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

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
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles",
				h.handleSection),

			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles",
				h.handleTotalCycles),

			// Get, add or edit a cycles table entry
			helpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycle/edit",
				h.handleEditGET),

			helpers.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit",
				h.handleEditPOST),

			helpers.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit",
				h.handleEditPUT),

			// Delete a cycle table entry
			helpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete",
				h.handleDELETE),
		},
	)
}

func (h *Cycles) handleSection(c echo.Context) error {

	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	toolID, err := helpers.ParseInt64Query(c, "tool_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"tool_id parameter is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	toolCycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
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
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool cycles: "+err.Error())
	}

	return nil
}

func (h *Cycles) handleTotalCycles(c echo.Context) error {
	// Get tool and position parameters
	toolID, err := helpers.ParseInt64Query(c, "tool_id")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"tool_id parameter is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest,
			"failed to get tool")
	}

	// Get cycles for this specific tool
	toolCycles, err := h.DB.PressCycles.GetPressCyclesForTool(toolID)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// Filter cycles by position
	filteredCycles := models.FilterByToolPosition(tool.Position, toolCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(toolID, filteredCycles...)

	return components.TotalCycles(
		totalCycles,
		helpers.ParseBoolQuery(c, "input"),
	).Render(c.Request().Context(), c.Response())
}

func (h *Cycles) handleEditGET(c echo.Context) error {
	props := &dialogs.EditCycleProps{}

	var err error
	props.Tools, err = h.DB.Tools.List()
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to load tool data: "+err.Error())
	}

	if c.QueryParam("id") != "" {
		cycleID, err := helpers.ParseInt64Query(c, "id")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				"failed to parse cycle ID: "+err.Error())
		}
		props.CycleID = cycleID

		// Get cycle data from the database
		cycle, err := h.DB.PressCycles.Get(cycleID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to load cycle data: "+err.Error())
		}
		props.InputPressNumber = &(cycle.PressNumber)
		props.InputTotalCycles = cycle.TotalCycles
		props.OriginalDate = &cycle.Date

		if props.Tool, err = h.DB.Tools.Get(cycle.ToolID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to load tool data: "+err.Error())
		}
	} else if c.QueryParam("tool_id") != "" {
		toolID, err := helpers.ParseInt64Query(c, "tool_id")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				"failed to parse tool ID: "+err.Error())
		}

		if props.Tool, err = h.DB.Tools.Get(toolID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError,
				"failed to load tool data: "+err.Error())
		}
	}

	cycleEditDialog := dialogs.EditCycle(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cycle edit dialog: "+err.Error())
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
		return err
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return echo.NewHTTPError(http.StatusBadRequest,
			"press_number must be a valid integer")
	}

	pressCycle := models.NewCycle(
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
	)

	pressCycle.Date = form.Date

	cycleID, err := h.DB.PressCycles.Add(pressCycle, user)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err), err)
	}

	// Handle regeneration if requested
	if form.Regenerating {
		_, err := h.DB.ToolRegenerations.AddToolRegeneration(cycleID, tool.ID, "", user)
		if err != nil {
			logger.HTMXHandlerTools().Error(
				"Failed to start regeneration for tool %d: %v",
				tool.ID, err)
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
		return err
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return echo.NewHTTPError(http.StatusBadRequest,
			"press_number must be a valid integer")
	}

	// Update the cycle
	pressCycle := models.NewPressCycleWithID(
		cycleID,
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
		form.Date,
	)

	if err := h.DB.PressCycles.Update(pressCycle, user); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to update press cycle: "+err.Error())
	}

	// Handle regeneration if requested
	if form.Regenerating {
		_, err := h.DB.ToolRegenerations.AddToolRegeneration(cycleID, tool.ID, "", user)
		if err != nil {
			logger.HTMXHandlerTools().Error(
				"Failed to start regeneration for tool %d: %v",
				tool.ID, err)
		}

		err = h.DB.ToolRegenerations.StopToolRegeneration(tool.ID, user)
		if err != nil {
			logger.HTMXHandlerTools().Error(
				"Failed to stop regeneration for tool %d: %v",
				tool.ID, err)
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

	if err := h.DB.PressCycles.Delete(cycleID, user); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to delete press cycle: "+err.Error())
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
	toolID, err := helpers.ParseInt64Query(c, "tool_id")
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest,
			"tool_id parameter is required")
	}

	tool, err := h.DB.Tools.Get(toolID)
	if err != nil {
		return nil, echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get tool: "+err.Error())
	}

	return tool, nil
}

// getCycleFormData parses form data for cycle operations
func (h *Cycles) getCycleFormData(c echo.Context) (*CycleEditFormData, error) {
	form := &CycleEditFormData{}

	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest,
				"press_number must be an integer")
		}

		pn := models.PressNumber(press)
		form.PressNumber = &pn
	}

	if dateString := c.FormValue("original_date"); dateString != "" {
		var err error
		form.Date, err = time.Parse(constants.DateFormat, dateString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest,
				"invalid date format: "+err.Error())
		}
	} else {
		form.Date = time.Now()
	}

	if totalCyclesString := c.FormValue("total_cycles"); totalCyclesString == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest,
			"total_cycles is required")
	} else {
		var err error
		form.TotalCycles, err = strconv.ParseInt(totalCyclesString, 10, 64)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest,
				"total_cycles must be an integer")
		}
	}

	form.Regenerating = c.FormValue("regenerating") != ""

	return form, nil
}
