package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/dialogs/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/utils"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) GetEditCycle(c echo.Context) error {
	// Check if we're in tool change mode
	toolChangeMode := utils.ParseQueryBool(c, "tool_change_mode")

	var (
		tool             *models.Tool
		cycle            *models.Cycle
		tools            []*models.Tool
		inputPressNumber *models.PressNumber
		inputTotalCycles int64
		originalDate     *time.Time
	)

	if c.QueryParam("id") != "" {
		cycleIDQuery, merr := utils.ParseQueryInt64(c, "id")
		if merr != nil {
			return merr.Echo()
		}
		cycleID := models.CycleID(cycleIDQuery)

		// Get cycle data from the database
		cycle, merr = h.registry.PressCycles.Get(cycleID)
		if merr != nil {
			return merr.Echo()
		}
		inputPressNumber = &(cycle.PressNumber)
		inputTotalCycles = cycle.TotalCycles
		originalDate = &cycle.Date

		// Set the cycles (original) tool to props
		tool, merr = h.registry.Tools.Get(cycle.ToolID)
		if merr != nil {
			return merr.Echo()
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			// Get all tools
			allTools, merr := h.registry.Tools.List()
			if merr != nil {
				return merr.Echo()
			}

			// Filter out tools not matching the original tools position
			for _, t := range allTools {
				if t.Position == tool.Position {
					tools = append(tools, t)
				}
			}

			// Sort tools alphabetically by code
			sort.Slice(tools, func(i, j int) bool {
				return tools[i].String() < tools[j].String()
			})
		}
	} else if c.QueryParam("tool_id") != "" {
		toolIDQuery, merr := utils.ParseQueryInt64(c, "tool_id")
		if merr != nil {
			return merr.Echo()
		}
		toolID := models.ToolID(toolIDQuery)

		tool, merr = h.registry.Tools.Get(toolID)
		if merr != nil {
			return merr.Echo()
		}
	} else {
		return echo.NewHTTPError(http.StatusBadRequest, "missing tool or cycle ID")
	}

	var t templ.Component
	var tName string
	if cycle != nil {
		t = templates.EditCycleDialog(
			tool, cycle, tools, inputPressNumber, inputTotalCycles, originalDate,
		)
		tName = "EditCycleDialog"
	} else {
		t = templates.NewCycleDialog(
			tool, inputPressNumber, inputTotalCycles, originalDate,
		)
		tName = "NewCycleDialog"
	}

	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, tName)
	}

	return nil
}

func (h *Handler) PostEditCycle(c echo.Context) error {
	slog.Info("Cycle creation request received")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	toolIDQuery, merr := utils.ParseQueryInt64(c, "tool_id")
	if merr != nil {
		return merr.Echo()
	}
	toolID := models.ToolID(toolIDQuery)

	tool, merr := h.registry.Tools.Get(toolID)
	if merr != nil {
		return merr.Echo()
	}

	// Parse form data
	form, merr := GetEditCycleFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	pc := models.NewCycle(*form.PressNumber, tool.ID, tool.Position,
		form.TotalCycles, user.TelegramID)

	pc.Date = form.Date

	_, merr = h.registry.PressCycles.Add(
		pc.PressNumber, pc.ToolID, pc.ToolPosition, pc.TotalCycles, pc.PerformedBy,
	)
	if merr != nil {
		return merr.Echo()
	}

	// Handle regeneration if requested
	if form.Regenerating {
		slog.Info("Starting tool regeneration", "tool_id", tool.ID, "user", user.Name)

		_, merr = h.registry.ToolRegenerations.StartToolRegeneration(tool.ID, "", user)
		if merr != nil {
			slog.Error(
				"Failed to start tool regeneration",
				"tool_id", tool.ID, "user", user.Name, "error", merr,
			)
		}
	}

	// Create Feed
	title := fmt.Sprintf("Neuer Zyklus hinzugefügt für %s", tool.String())
	content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
		*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))

	if form.Regenerating {
		content += "\nRegenerierung gestartet"
	}

	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for cycle creation", "error", merr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditCycle(c echo.Context) error {
	slog.Info("Updating cycle")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewMasterError(err, http.StatusBadRequest).Echo()
	}
	cycleID := models.CycleID(cycleIDQuery)

	cycle, merr := h.registry.PressCycles.Get(cycleID)
	if merr != nil {
		return merr.Echo()
	}

	// Get original tool
	originalTool, merr := h.registry.Tools.Get(cycle.ToolID)
	if merr != nil {
		return merr.Echo()
	}

	form, merr := GetEditCycleFormData(c)
	if merr != nil {
		return merr.Echo()
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return errors.NewMasterError(fmt.Errorf("press_number must be a valid integer"), http.StatusBadRequest).Echo()
	}

	// Determine which tool to use for the cycle
	var tool *models.Tool
	if form.ToolID != nil {
		// Tool change requested - get the new tool
		newTool, merr := h.registry.Tools.Get(*form.ToolID)
		if merr != nil {
			return merr.Echo()
		}
		tool = newTool
	} else {
		// No tool change - use original tool
		tool = originalTool
	}

	// Update the cycle
	pc := models.NewCycleWithID(
		cycle.ID,
		*form.PressNumber,
		tool.ID, tool.Position, form.TotalCycles,
		user.TelegramID,
		form.Date,
	)

	merr = h.registry.PressCycles.Update(
		pc.ID,
		pc.PressNumber,
		pc.ToolID,
		pc.ToolPosition,
		pc.TotalCycles,
		pc.Date,
		pc.PerformedBy,
	)
	if merr != nil {
		return merr.Echo()
	}

	// Handle regeneration if requested
	if form.Regenerating {
		_, merr = h.registry.ToolRegenerations.Add(tool.ID, pc.ID, "", user)
		if merr != nil {
			slog.Error("Failed to add tool regeneration", "error", merr)
		}
	}

	// Create Feed
	var title string
	var content string
	if form.ToolID != nil {
		// Tool change occurred
		title = "Zyklus aktualisiert mit Werkzeugwechsel"
		content = fmt.Sprintf("Presse: %d\nAltes Werkzeug: %s\nNeues Werkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, originalTool.String(), tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))
	} else {
		// Regular cycle update
		title = fmt.Sprintf("Zyklus aktualisiert für %s", tool.String())
		content = fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))
	}

	if form.Regenerating {
		content += "\nRegenerierung abgeschlossen"
	}

	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for cycle update", "error", merr)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

type EditCycleFormData struct {
	TotalCycles  int64 // TotalCycles form field name "total_cycles"
	PressNumber  *models.PressNumber
	Date         time.Time // OriginalDate form field name "original_date"
	Regenerating bool
	ToolID       *models.ToolID // ToolID form field name "tool_id" (for tool change mode)
}

func GetEditCycleFormData(c echo.Context) (*EditCycleFormData, *errors.MasterError) {
	form := &EditCycleFormData{}

	// Parse press number
	if pressString := c.FormValue("press_number"); pressString != "" {
		press, err := strconv.Atoi(pressString)
		if err != nil {
			return nil, errors.NewMasterError(err, http.StatusBadRequest)
		}
		pn := models.PressNumber(press)
		form.PressNumber = &pn

		if !models.IsValidPressNumber(form.PressNumber) {
			return nil, errors.NewMasterError(
				fmt.Errorf("press_number must be a valid integer"),
				http.StatusBadRequest,
			)
		}
	}

	// Parse date
	if dateString := c.FormValue("original_date"); dateString != "" {
		var err error
		form.Date, err = time.Parse(env.DateFormat, dateString)
		if err != nil {
			return nil, errors.NewMasterError(err, http.StatusBadRequest)
		}
	} else {
		form.Date = time.Now()
	}

	// Parse total cycles (required)
	totalCyclesString := c.FormValue("total_cycles")
	if totalCyclesString == "" {
		return nil, errors.NewMasterError(
			fmt.Errorf("form value total_cycles is required"),
			http.StatusBadRequest,
		)
	}

	var err error
	form.TotalCycles, err = strconv.ParseInt(totalCyclesString, 10, 64)
	if err != nil {
		return nil, errors.NewMasterError(err, http.StatusBadRequest)
	}

	// Parse regenerating flag
	form.Regenerating = c.FormValue("regenerating") != ""

	// Parse tool_id if present (for tool change mode)
	if toolIDString := c.FormValue("tool_id"); toolIDString != "" {
		toolIDParsed, err := strconv.ParseInt(toolIDString, 10, 64)
		if err != nil {
			return nil, errors.NewMasterError(err, http.StatusBadRequest)
		}
		toolID := models.ToolID(toolIDParsed)
		form.ToolID = &toolID
	}

	return form, nil
}
