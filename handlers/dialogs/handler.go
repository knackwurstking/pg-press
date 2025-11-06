package dialogs

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/handlers/dialogs/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Get, add or edit a cycles table entry
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-cycle", h.GetEditCycle),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-cycle", h.PostEditCycle),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-cycle", h.PutEditCycle),
	})
}

func (h *Handler) GetEditCycle(c echo.Context) error {
	props := &components.DialogEditCycleProps{}

	// Check if we're in tool change mode
	toolChangeMode := utils.ParseQueryBool(c, "tool_change_mode")

	if c.QueryParam("id") != "" {
		cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
		if err != nil {
			return utils.HandleBadRequest(err, "failed to parse cycle ID")
		}
		props.CycleID = models.CycleID(cycleIDQuery)

		// Get cycle data from the database
		cycle, err := h.registry.PressCycles.Get(props.CycleID)
		if err != nil {
			return utils.HandleError(err, "failed to load cycle data")
		}
		props.InputPressNumber = &(cycle.PressNumber)
		props.InputTotalCycles = cycle.TotalCycles
		props.OriginalDate = &cycle.Date

		// Set the cycles (original) tool to props
		if props.Tool, err = h.registry.Tools.Get(cycle.ToolID); err != nil {
			return utils.HandleError(err, "failed to load tool data")
		}

		// If in tool change mode, load all available tools for this press
		if toolChangeMode {
			props.AllowToolChange = true

			// Get all tools
			allTools, err := h.registry.Tools.List()
			if err != nil {
				return utils.HandleError(err, "failed to load available tools")
			}

			// Filter out tools not matching the original tools position
			for _, t := range allTools {
				if t.Position == props.Tool.Position {
					props.AvailableTools = append(props.AvailableTools, t)
				}
			}

			// Sort tools alphabetically by code
			sort.Slice(props.AvailableTools, func(i, j int) bool {
				return props.AvailableTools[i].String() < props.AvailableTools[j].String()
			})
		}
	} else if c.QueryParam("tool_id") != "" {
		toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
		if err != nil {
			return utils.HandleBadRequest(err, "failed to parse tool ID")
		}
		toolID := models.ToolID(toolIDQuery)

		if props.Tool, err = h.registry.Tools.Get(toolID); err != nil {
			return utils.HandleError(err, "failed to load tool data")
		}
	} else {
		return utils.HandleBadRequest(nil, "missing tool or cycle ID")
	}

	cycleEditDialog := components.DialogEditCycle(props)
	if err := cycleEditDialog.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render cycle edit dialog")
	}

	return nil
}

func (h *Handler) PostEditCycle(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "tool_id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return utils.HandleError(err, "failed to load tool data")
	}

	// Parse form data
	form, err := getCycleFormData(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse form data")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return utils.HandleBadRequest(nil, "press_number must be a valid integer")
	}

	pressCycle := models.NewCycle(
		*form.PressNumber,
		tool.ID,
		tool.Position,
		form.TotalCycles,
		user.TelegramID,
	)

	pressCycle.Date = form.Date

	_, err = h.registry.PressCycles.Add(pressCycle, user)
	if err != nil {
		return utils.HandleError(err, "failed to add cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		slog.Info("Starting regeneration", "tool", tool.ID, "user_name", user.Name)
		_, err := h.registry.ToolRegenerations.StartToolRegeneration(tool.ID, "", user)
		if err != nil {
			slog.Error("Failed to start regeneration", "tool", tool.ID, "user_name", user.Name, "error", err)
		}
	}

	{ // Create Feed
		title := fmt.Sprintf("Neuer Zyklus hinzugefügt für %s", tool.String())
		content := fmt.Sprintf("Presse: %d\nWerkzeug: %s\nGesamtzyklen: %d\nDatum: %s",
			*form.PressNumber, tool.String(), form.TotalCycles, form.Date.Format("2006-01-02 15:04:05"))

		if form.Regenerating {
			content += "\nRegenerierung gestartet"
		}

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) PutEditCycle(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleError(err, "failed to get user from context")
	}

	cycleIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse ID from query")
	}
	cycleID := models.CycleID(cycleIDQuery)

	cycle, err := h.registry.PressCycles.Get(cycleID)
	if err != nil {
		return utils.HandleError(err, "failed to get cycle")
	}

	// Get original tool
	originalTool, err := h.registry.Tools.Get(cycle.ToolID)
	if err != nil {
		return utils.HandleError(err, "failed to get original tool")
	}

	form, err := getCycleFormData(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get cycle form data from query")
	}

	if !models.IsValidPressNumber(form.PressNumber) {
		return utils.HandleBadRequest(nil, "press_number must be a valid integer")
	}

	// Determine which tool to use for the cycle
	var tool *models.Tool
	if form.ToolID != nil {
		// Tool change requested - get the new tool
		newTool, err := h.registry.Tools.Get(*form.ToolID)
		if err != nil {
			return utils.HandleError(err, "failed to get new tool")
		}
		tool = newTool
	} else {
		// No tool change - use original tool
		tool = originalTool
	}

	// Update the cycle
	pressCycle := models.NewCycleWithID(
		cycle.ID,
		*form.PressNumber,
		tool.ID, tool.Position, form.TotalCycles,
		user.TelegramID,
		form.Date,
	)

	if err := h.registry.PressCycles.Update(pressCycle, user); err != nil {
		return utils.HandleError(err, "failed to update press cycle")
	}

	// Handle regeneration if requested
	if form.Regenerating {
		if _, err := h.registry.ToolRegenerations.Add(tool.ID, pressCycle.ID, "", user); err != nil {
			slog.Error("Failed to add tool regeneration", "error", err)
		}
	}

	{ // Create Feed
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

		h.createFeed(title, content, user.TelegramID)
	}

	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	return nil
}

func (h *Handler) createFeed(title, content string, userID models.TelegramID) {
	feed := models.NewFeed(title, content, userID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "error", err)
	}
}
