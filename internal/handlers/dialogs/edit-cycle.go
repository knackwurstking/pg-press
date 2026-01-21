package dialogs

import (
	"net/http"
	"sort"
	"time"

	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/utils"

	"github.com/labstack/echo/v4"
)

func GetEditCycle(c echo.Context) *echo.HTTPError {
	// Check if we're in tool change mode
	toolChangeMode := utils.GetQueryBool(c, "tool_change_mode")

	// Edit Cycle Dialog
	if c.QueryParam("id") != "" {
		cycleIDQuery, merr := utils.GetQueryInt64(c, "id")
		if merr != nil {
			return merr.Echo()
		}

		// Get cycle data from the database
		cycle, merr := db.GetCycle(shared.EntityID(cycleIDQuery))
		if merr != nil {
			return merr.Echo()
		}

		// Set the cycles (original) tool to props
		tool, merr := db.GetTool(cycle.ToolID)
		if merr != nil {
			return merr.Echo()
		}

		// If in tool change mode, load all available tools for this press
		var tools []*shared.Tool
		if toolChangeMode {
			// Get all tools
			tools, merr = db.ListTools()
			if merr != nil {
				return merr.Echo()
			}

			// Filter out tools not matching the original tools position
			for _, t := range tools {
				if t.Position != tool.Position {
					continue
				}
				tools = append(tools, t)
			}

			// Sort tools alphabetically by code
			sort.Slice(tools, func(i, j int) bool {
				return tools[i].String() < tools[j].String()
			})
		}

		t := templates.EditCycleDialog(cycle, tool, tools)
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

	pressNumber, merr := db.GetPressNumberForTool(tool.ID)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.NewCycleDialog(tool, pressNumber)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NewCycleDialog")
	}

	return nil
}

func PostCycle(c echo.Context) *echo.HTTPError {
	cycle, verr := parseCycleForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	log.Debug("Create a new press cycles entry. [cycle=%v]", cycle)

	if merr := db.AddCycle(cycle); merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-cycles")

	return nil
}

func PutCycle(c echo.Context) *echo.HTTPError {
	cycle, verr := parseCycleForm(c)
	if verr != nil {
		return verr.HTTPError().Echo()
	}

	log.Debug("Update existing cycle with ID %d. [cycle=%v]", cycle.ID, cycle)

	if merr := db.UpdateCycle(cycle); merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-cycles")

	return nil
}

func parseCycleForm(c echo.Context) (*shared.Cycle, *errors.ValidationError) {
	// Tool ID
	var toolID shared.EntityID
	originalToolID, err := utils.SanitizeInt64(c.FormValue("original_tool_id"))
	if err != nil {
		return nil, errors.NewValidationError("original_tool_id: %v", err)
	}
	if c.FormValue("tool_id") != "" {
		newToolID, err := utils.SanitizeInt64(c.FormValue("tool_id"))
		if err != nil {
			return nil, errors.NewValidationError("tool_id: %v", err)
		}
		toolID = shared.EntityID(newToolID)
	} else {
		toolID = shared.EntityID(originalToolID)
	}

	// Press Number
	pn, err := utils.SanitizeInt8(c.FormValue("press_number"))
	if err != nil {
		return nil, errors.NewValidationError("press_number: %v", err)
	}
	pressNumber := shared.PressNumber(pn)

	// Press Cycles
	pc, err := utils.SanitizeInt64(c.FormValue("press_cycles"))
	if err != nil {
		return nil, errors.NewValidationError("press_cycles: %v", err)
	}
	pressCycles := pc

	// Stop
	stopTime, err := time.Parse("2006-01-02", c.FormValue("stop"))
	if err != nil {
		return nil, errors.NewValidationError("stop: %v", err)
	}
	stop := shared.NewUnixMilli(stopTime)

	return shared.NewCycle(toolID, pressNumber, pressCycles, stop), nil
}
