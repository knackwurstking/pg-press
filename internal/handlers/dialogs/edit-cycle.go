package dialogs

import (
	"net/http"
	"sort"
	"strconv"
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
	cycle, merr := parseCycleForm(c)
	if merr != nil {
		return merr.Echo()
	}

	log.Debug("Create a new press cycles entry. [cycle=%v]", cycle)

	merr = db.AddCycle(cycle)
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-cycles")

	return nil
}

func PutCycle(c echo.Context) *echo.HTTPError {
	cycle, merr := parseCycleForm(c)
	if merr != nil {
		return merr.Echo()
	}

	log.Debug("Update existing cycle with ID %d. [cycle=%v]", cycle.ID, cycle)

	merr = db.UpdateCycle(cycle)
	if merr != nil {
		return merr.Echo()
	}

	utils.SetHXTrigger(c, "reload-cycles")

	return nil
}

func parseCycleForm(c echo.Context) (*shared.Cycle, *errors.HTTPError) {
	cycle := &shared.Cycle{}

	// Tool
	originalToolID, err := strconv.ParseInt(c.FormValue("original_tool_id"), 10, 64)
	if err != nil {
		return cycle, errors.NewHTTPError(err).Wrap("original_tool_id")
	}
	if c.FormValue("tool_id") != "" {
		newToolID, err := strconv.ParseInt(c.FormValue("tool_id"), 10, 64)
		if err != nil {
			return cycle, errors.NewHTTPError(err).Wrap("tool_id")
		}
		cycle.ToolID = shared.EntityID(newToolID)
	} else {
		cycle.ToolID = shared.EntityID(originalToolID)
	}

	// Press Number
	pressNumber, err := strconv.ParseInt(c.FormValue("press_number"), 10, 8)
	if err != nil {
		return cycle, errors.NewHTTPError(err).Wrap("press_number")
	}
	cycle.PressNumber = shared.PressNumber(pressNumber)

	// Total Cycles
	totalCycles, err := strconv.ParseInt(c.FormValue("press_cycles"), 10, 64)
	if err != nil {
		return cycle, errors.NewHTTPError(err).Wrap("press_cycles")
	}
	cycle.PressCycles = totalCycles

	// Stop
	stop, err := time.Parse("2006-01-02", c.FormValue("stop"))
	if err != nil {
		return cycle, errors.NewHTTPError(err).Wrap("stop")
	}
	cycle.Stop = shared.NewUnixMilli(stop)

	return cycle, nil
}
