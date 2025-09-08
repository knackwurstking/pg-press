package htmx

import (
	"net/http"
	"strconv"
	"time"

	database "github.com/knackwurstking/pgpress/internal/database/core"
	"github.com/knackwurstking/pgpress/internal/database/dberror"
	presscyclemodels "github.com/knackwurstking/pgpress/internal/database/models/presscycle"
	regenerationmodels "github.com/knackwurstking/pgpress/internal/database/models/regeneration"
	toolmodels "github.com/knackwurstking/pgpress/internal/database/models/tool"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	toolscomp "github.com/knackwurstking/pgpress/internal/web/templates/components/tools"
	"github.com/labstack/echo/v4"
)

type Cycles struct {
	DB *database.DB
}

func (h *Cycles) RegisterRoutes(e *echo.Echo) {
	webhelpers.RegisterEchoRoutes(
		e,
		[]*webhelpers.EchoRoute{
			// Cycles table rows
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycles", h.handle),
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/total-cycles", h.handleTotalCycles),

			// Get, add or edit a cycles table entry
			webhelpers.NewEchoRoute(http.MethodGet, "/htmx/tools/cycle/edit", func(c echo.Context) error {
				return h.handleEditGET(nil, c)
			}),
			webhelpers.NewEchoRoute(http.MethodPost, "/htmx/tools/cycle/edit", h.handleEditPOST),
			webhelpers.NewEchoRoute(http.MethodPut, "/htmx/tools/cycle/edit", h.handleEditPUT),

			// Delete a cycle table entry
			webhelpers.NewEchoRoute(http.MethodDelete, "/htmx/tools/cycle/delete", h.handleDELETE),
		},
	)
}

func (h *Cycles) handle(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Get slot parameters
	slotTop, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTop)
	if err != nil {
		slotTop = 0
	}
	slotTopCassette, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTopCassette)
	if err != nil {
		slotTopCassette = 0
	}
	slotBottom, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotBottom)
	if err != nil {
		slotBottom = 0
	}

	// Validate that at least one slot is provided
	if slotTop == 0 && slotTopCassette == 0 && slotBottom == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one slot must be provided")
	}

	logger.HTMXHandlerTools().Debug("Fetching cycles for slots: top=%d, top_cassette=%d, bottom=%d",
		slotTop, slotTopCassette, slotBottom)

	// Get all press cycles (we'll filter by slots in frontend for now)
	allCycles, err := h.DB.PressCycles.List()
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}

	// TODO: Need to handle regenerations here
	var regenerations []*regenerationmodels.ToolRegeneration

	// Filter cycles that match any of the provided slots
	filteredCycles := presscyclemodels.FilterPressCycleSlots(slotTop, slotTopCassette, slotBottom, allCycles...)

	// Get total cycles and lastPartialCycles from filtered cycles
	totalCycles := h.getTotalCycles(filteredCycles...)

	// Render the component
	cyclesSection := toolscomp.CyclesSection(&toolscomp.CyclesSectionProps{
		User:            user,
		SlotTop:         slotTop,
		SlotTopCassette: slotTopCassette,
		SlotBottom:      slotBottom,
		TotalCycles:     totalCycles,
		Cycles:          filteredCycles,
		Regenerations:   regenerations,
	})
	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render tool cycles: "+err.Error())
	}

	return nil
}

func (h *Cycles) handleTotalCycles(c echo.Context) error {
	// Get slot parameters
	slotTop, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTop)
	if err != nil {
		slotTop = 0
	}
	slotTopCassette, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTopCassette)
	if err != nil {
		slotTopCassette = 0
	}
	slotBottom, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotBottom)
	if err != nil {
		slotBottom = 0
	}

	// Validate that at least one slot is provided
	if slotTop == 0 && slotTopCassette == 0 && slotBottom == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one slot must be provided")
	}

	// Get all press cycles and filter by slots
	allCycles, err := h.DB.PressCycles.List()
	if err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get press cycles: "+err.Error())
	}
	// TODO: Need to handle regenerations somehow

	// Filter cycles that match any of the provided slots
	filteredCycles := presscyclemodels.FilterPressCycleSlots(slotTop, slotTopCassette, slotBottom, allCycles...)

	// Get total cycles from filtered cycles
	totalCycles := h.getTotalCycles(filteredCycles...)

	return toolscomp.TotalCycles(
		totalCycles,
		webhelpers.ParseBoolQuery(c, constants.QueryParamInput),
	).Render(c.Request().Context(), c.Response())
}

func (h *Cycles) handleEditGET(props *toolscomp.CycleEditDialogProps, c echo.Context) error {
	if props == nil {
		props = &toolscomp.CycleEditDialogProps{}
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

		cycleEditDialog := toolscomp.CycleEditDialog(props)
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

	cycleEditDialog := toolscomp.CycleEditDialog(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render cycle edit dialog: "+err.Error())
	}

	return nil
}

func (h *Cycles) handleEditPOST(c echo.Context) error {
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
		return h.handleEditGET(&toolscomp.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            err.Error(),
			InputPressNumber: nil, // Don't have form data to repopulate
		}, c)
	}

	if !presscyclemodels.IsValidPressNumber(formData.PressNumber) {
		return h.handleEditGET(&toolscomp.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            "press_number must be a valid integer",
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
			OriginalDate:     &formData.Date,
		}, c)
	}

	var slotTopID, slotTopCassetteID, slotBottomID int64
	if toolTop != nil {
		slotTopID = toolTop.ID
	}
	if toolTopCassette != nil {
		slotTopCassetteID = toolTopCassette.ID
	}
	if toolBottom != nil {
		slotBottomID = toolBottom.ID
	}

	pressCycle := presscyclemodels.NewPressCycle(
		slotTopID, slotTopCassetteID, slotBottomID,
		*formData.PressNumber,
		formData.TotalCycles,
		user.TelegramID,
	)
	pressCycle.Date = formData.Date

	if _, err := h.DB.PressCycles.Add(pressCycle, user); err != nil {
		return h.handleEditGET(&toolscomp.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
			OriginalDate:     &formData.Date,
		}, c)
	}

	return h.handleEditGET(&toolscomp.CycleEditDialogProps{
		SlotTop:         toolTop,
		SlotTopCassette: toolTopCassette,
		SlotBottom:      toolBottom,
		Close:           true,
	}, c)
}

func (h *Cycles) handleEditPUT(c echo.Context) error {
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
		return h.handleEditGET(&toolscomp.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputPressNumber: nil, // Don't have form data to repopulate
		}, c)
	}

	if !presscyclemodels.IsValidPressNumber(formData.PressNumber) {
		return h.handleEditGET(&toolscomp.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			Error:            "press_number must be a valid integer",
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
			OriginalDate:     &formData.Date,
		}, c)
	}

	var slotTopID, slotTopCassetteID, slotBottomID int64
	if toolTop != nil {
		slotTopID = toolTop.ID
	}
	if toolTopCassette != nil {
		slotTopCassetteID = toolTopCassette.ID
	}
	if toolBottom != nil {
		slotBottomID = toolBottom.ID
	}

	// Update only the fields that should change, preserving the original date
	pressCycle := presscyclemodels.NewPressCycleWithID(
		cycleID,
		slotTopID, slotTopCassetteID, slotBottomID,
		*formData.PressNumber,
		formData.TotalCycles,
		user.TelegramID,
		formData.Date,
	)

	if err := h.DB.PressCycles.Update(pressCycle, user); err != nil {
		return h.handleEditGET(&toolscomp.CycleEditDialogProps{
			SlotTop:          toolTop,
			SlotTopCassette:  toolTopCassette,
			SlotBottom:       toolBottom,
			CycleID:          cycleID,
			Error:            err.Error(),
			InputTotalCycles: formData.TotalCycles,
			InputPressNumber: formData.PressNumber,
			OriginalDate:     &formData.Date,
		}, c)
	}

	return h.handleEditGET(&toolscomp.CycleEditDialogProps{
		SlotTop:         toolTop,
		SlotTopCassette: toolTopCassette,
		SlotBottom:      toolBottom,
		Close:           true,
	}, c)
}

func (h *Cycles) handleDELETE(c echo.Context) error {
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	cycleID, err := webhelpers.ParseInt64Query(c, constants.QueryParamCycleID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTools().Debug("Handling cycle deletion request for ID %d", cycleID)

	if err := h.DB.PressCycles.Delete(cycleID, user); err != nil {
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to delete press cycle: "+err.Error())
	}

	return h.handle(c)
}

// NOTE: The database will always sort IDs DESC
func (h *Cycles) getTotalCycles(cycles ...*presscyclemodels.PressCycle) int64 {
	var totalCycles int64

	for _, cycle := range cycles {
		totalCycles += cycle.PartialCycles
	}

	return totalCycles
}

func (h *Cycles) getSlotsFromQuery(c echo.Context) (toolTop, toolTopCassette, toolBottom *toolmodels.Tool, err error) {
	slotTop, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTop)
	if err != nil {
		slotTop = 0
	}

	slotTopCassette, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotTopCassette)
	if err != nil {
		slotTopCassette = 0
	}

	slotBottom, err := webhelpers.ParseInt64Query(c, constants.QueryParamSlotBottom)
	if err != nil {
		slotBottom = 0
	}

	// Validate slots, at least one must be provided
	if slotTop == 0 && slotTopCassette == 0 && slotBottom == 0 {
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

func (h *Cycles) getCycleFormData(c echo.Context) (*CycleEditFormData, error) {
	totalCyclesString := c.FormValue("total_cycles")
	if totalCyclesString == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "total_cycles is required")
	}
	totalCycles, err := strconv.ParseInt(totalCyclesString, 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "total_cycles must be an integer")
	}

	var pressNumber *presscyclemodels.PressNumber
	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "press_number must be an integer")
		}

		pn := presscyclemodels.PressNumber(press)
		pressNumber = &pn
	}

	var date time.Time
	if dateString := c.FormValue(constants.QueryParamOriginalDate); dateString != "" {
		// Create time (date) object from dateString
		date, err = time.Parse(toolscomp.DateFormat, dateString)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid date format: "+err.Error())
		}
	} else {
		date = time.Now()
	}

	return &CycleEditFormData{
		TotalCycles: totalCycles,
		PressNumber: pressNumber,
		Date:        date,
	}, nil
}
