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
		// Edit cycle dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-cycle", h.GetEditCycle),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-cycle", h.PostEditCycle),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-cycle", h.PutEditCycle),

		// Edit tool dialog
		utils.NewEchoRoute(http.MethodGet, "/htmx/dialogs/edit-tool", h.GetEditTool),
		utils.NewEchoRoute(http.MethodPost, "/htmx/dialogs/edit-tool", h.PostEditTool),
		utils.NewEchoRoute(http.MethodPut, "/htmx/dialogs/edit-tool", h.PutEditTool),
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
	form, err := getEditCycleFormData(c)
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

	form, err := getEditCycleFormData(c)
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

func (h *Handler) GetEditTool(c echo.Context) error {
	props := &components.DialogEditToolProps{}

	toolIDQuery, _ := utils.ParseQueryInt64(c, "id")
	if toolIDQuery > 0 {
		tool, err := h.registry.Tools.Get(models.ToolID(toolIDQuery))
		if err != nil {
			return utils.HandleError(err, "failed to get tool from database")
		}

		props.Tool = tool
		props.InputPosition = string(tool.Position)
		props.InputWidth = tool.Format.Width
		props.InputHeight = tool.Format.Height
		props.InputType = tool.Type
		props.InputCode = tool.Code
		props.InputPressSelection = tool.Press
	}

	dialog := components.DialogEditTool(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return utils.HandleError(err, "failed to render tool edit dialog")
	}
	return nil
}

func (h *Handler) PostEditTool(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	formData, err := getEditToolFormData(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get tool form data")
	}

	tool := models.NewTool(formData.Position, formData.Format, formData.Code, formData.Type)
	tool.SetPress(formData.Press)

	id, err := h.registry.Tools.Add(tool, user)
	if err != nil {
		return utils.HandleError(err, "failed to add tool")
	}

	slog.Info("Created tool", "id", id, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	title := "Neues Werkzeug erstellt"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	h.createFeed(title, content, user.TelegramID)

	utils.SetHXTrigger(c, env.HXGlobalTrigger)
	return nil
}

func (h *Handler) PutEditTool(c echo.Context) error {
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get user from context")
	}

	toolIDQuery, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return utils.HandleBadRequest(err, "failed to parse tool ID")
	}
	toolID := models.ToolID(toolIDQuery)

	formData, err := getEditToolFormData(c)
	if err != nil {
		return utils.HandleBadRequest(err, "failed to get tool form data")
	}

	tool, err := h.registry.Tools.Get(toolID)
	if err != nil {
		return utils.HandleError(err, "failed to get tool")
	}

	tool.Press = formData.Press
	tool.Position = formData.Position
	tool.Format = formData.Format
	tool.Code = formData.Code
	tool.Type = formData.Type

	if err := h.registry.Tools.Update(tool, user); err != nil {
		return utils.HandleError(err, "failed to update tool")
	}

	slog.Info("Updated tool", "id", tool.ID, "type", tool.Type, "code", tool.Code, "user_name", user.Name)

	// Create feed entry
	title := "Werkzeug aktualisiert"

	content := fmt.Sprintf("Werkzeug: %s\nTyp: %s\nCode: %s\nPosition: %s",
		tool.String(), tool.Type, tool.Code, string(tool.Position))

	if tool.Press != nil {
		content += fmt.Sprintf("\nPresse: %d", *tool.Press)
	}

	h.createFeed(title, content, user.TelegramID)

	// Set HX headers
	utils.SetHXTrigger(c, env.HXGlobalTrigger)

	utils.SetHXAfterSettle(c, map[string]interface{}{
		"toolUpdated": map[string]string{
			"pageTitle": fmt.Sprintf("PG Presse | %s %s",
				tool.String(), tool.Position.GermanString()),
			"appBarTitle": fmt.Sprintf("%s %s", tool.String(),
				tool.Position.GermanString()),
		},
	})

	return nil
}

func (h *Handler) createFeed(title, content string, userID models.TelegramID) {
	feed := models.NewFeed(title, content, userID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "error", err)
	}
}
