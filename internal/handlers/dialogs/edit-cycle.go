package dialogs

import (
	"fmt"
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditCycle(c echo.Context) *echo.HTTPError {
	// Check if we're in tool change mode
	toolChangeMode := utils.GetQueryBool(c, "tool_change_mode")

	presses, herr := db.ListPress()
	if herr != nil {
		return herr.Echo()
	}

	// Edit Cycle Dialog
	if c.QueryParam("id") != "" {
		cycleIDQuery, herr := utils.GetQueryInt64(c, "id")
		if herr != nil {
			return herr.Echo()
		}

		// Get cycle data from the database
		cycle, herr := db.GetCycle(shared.EntityID(cycleIDQuery))
		if herr != nil {
			return herr.Echo()
		}

		// Set the cycles (original) tool to props
		tool, herr := db.GetTool(cycle.ToolID)
		if herr != nil {
			return herr.Echo()
		}

		// If in tool change mode, load all available tools for this press
		var tools []*shared.Tool
		if toolChangeMode {
			// Get all tools
			allTools, herr := db.ListTools()
			if herr != nil {
				return herr.Echo()
			}

			// Filter out tools not matching the original tools position
			for _, t := range allTools {
				if t.Position != tool.Position {
					continue
				}
				tools = append(tools, t)
			}
		}

		t := EditCycleDialog(cycle.ID, CycleDialogProps{
			CycleFormData: CycleFormData{
				ToolID:      cycle.ToolID,
				Tools:       tools,
				PressID:     cycle.PressID,
				Presses:     presses,
				Stop:        cycle.Stop,
				PressCycles: cycle.PressCycles,
			},
			Open: true,
			OOB:  true,
		})
		err := t.Render(c.Request().Context(), c.Response())
		if err != nil {
			return errors.NewRenderError(err, "EditCycleDialog")
		}

		return nil
	}

	// New Cycle Dialog
	if c.QueryParam("tool_id") == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing tool or cycle ID")
	}

	id, merr := utils.GetQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	tool, merr := db.GetTool(shared.EntityID(id))
	if merr != nil {
		return merr.Echo()
	}

	currentPressID := shared.EntityID(0)
	if p, herr := db.GetPressForTool(tool.ID); herr != nil {
		return herr.Echo()
	} else if p != nil {
		currentPressID = p.ID
	}

	t := NewCycleDialog(CycleDialogProps{
		CycleFormData: CycleFormData{
			ToolID:  tool.ID,
			PressID: currentPressID,
			Presses: presses,
		},
		Open: true,
		OOB:  true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewCycleDialog")
	}

	return nil
}

func PostCycle(c echo.Context) *echo.HTTPError {
	if id, _ := utils.GetQueryInt64(c, "id"); id != 0 {
		return updateCycle(c, shared.EntityID(id))
	}

	data, ierrs := parseCycleForm(c, 0)
	if len(ierrs) > 0 {
		return reRenderNewCycleDialog(c, true, data, ierrs...)
	}

	log.Debug("Create a new press cycles entry. [data=%#v]", data)

	cycle := shared.NewCycle(data.ToolID, data.PressID, data.PressCycles, data.Stop)
	if herr := db.AddCycle(cycle); herr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to create cycle: %v", herr))
		return reRenderNewCycleDialog(c, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-cycles")

	return reRenderNewCycleDialog(c, false, data)
}

func updateCycle(c echo.Context, cycleID shared.EntityID) *echo.HTTPError {
	cycle, herr := db.GetCycle(cycleID)
	if herr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to load cycle with ID %d: %v", cycleID, herr))
		return reRenderEditCycleDialog(c, cycleID, true, CycleFormData{}, ierr)
	}

	data, ierrs := parseCycleForm(c, cycle.ToolID)
	if len(ierrs) > 0 {
		return reRenderEditCycleDialog(c, cycleID, true, data, ierrs...)
	}
	cycle.ToolID = data.ToolID
	cycle.PressID = data.PressID
	cycle.Stop = data.Stop
	cycle.PressCycles = data.PressCycles

	log.Debug("Update existing cycle with ID %d. [data=%#v]", cycle.ID, data)

	if herr := db.UpdateCycle(cycle); herr != nil {
		ierr := errors.NewInputError("", fmt.Sprintf("failed to update cycle: %v", herr))
		return reRenderEditCycleDialog(c, cycleID, true, data, ierr)
	}

	utils.SetHXTrigger(c, "reload-cycles")

	return reRenderEditCycleDialog(c, cycleID, false, data)
}

// FIXME: Why is the tool_id form value always wrong after changing?
func parseCycleForm(c echo.Context, toolID shared.EntityID) (data CycleFormData, ierrs []*errors.InputError) {
	// Tool ID
	if c.FormValue("tool_id") != "" {
		newToolID, err := utils.SanitizeInt64(c.FormValue("tool_id"))
		if err != nil {
			ierr := errors.NewInputError("tool_id", fmt.Sprintf("invalid tool ID: %v", err))
			ierrs = append(ierrs, ierr)
		}
		toolID = shared.EntityID(newToolID)
	}
	if toolID == 0 {
		ierr := errors.NewInputError("tool_id", "tool ID is required")
		ierrs = append(ierrs, ierr)
	}
	data.ToolID = toolID

	// Press Number
	id, err := utils.SanitizeInt8(c.FormValue("press_id"))
	if err != nil {
		ierr := errors.NewInputError("press_id", fmt.Sprintf("invalid press ID: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.PressID = shared.EntityID(id)

	// Press Cycles
	pc, err := utils.SanitizeInt64(c.FormValue("cycles"))
	if err != nil {
		ierr := errors.NewInputError("cycles", fmt.Sprintf("invalid press cycles: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.PressCycles = pc

	// Stop
	stopTime, err := time.Parse("2006-01-02", c.FormValue("stop"))
	if err != nil {
		ierr := errors.NewInputError("stop", fmt.Sprintf("invalid stop time: %v", err))
		ierrs = append(ierrs, ierr)
	}
	data.Stop = shared.NewUnixMilli(stopTime)

	return
}

func reRenderNewCycleDialog(c echo.Context, open bool, formData CycleFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := NewCycleDialog(CycleDialogProps{
		CycleFormData: formData,
		Open:          open,
		OOB:           true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewCycleDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}

func reRenderEditCycleDialog(c echo.Context, cycleID shared.EntityID, open bool, formData CycleFormData, ierrs ...*errors.InputError) *echo.HTTPError {
	t := EditCycleDialog(cycleID, CycleDialogProps{
		CycleFormData: formData,
		Open:          open,
		OOB:           true,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "EditCycleDialog")
	}

	if len(ierrs) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
	}

	return nil
}
