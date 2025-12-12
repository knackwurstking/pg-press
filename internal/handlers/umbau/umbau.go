package umbau

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/umbau/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"

	ui "github.com/knackwurstking/ui/ui-templ"

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

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(
		e,
		env.ServerPathPrefix,
		[]*ui.EchoRoute{
			ui.NewEchoRoute(http.MethodGet, path+"/:press", h.GetUmbauPage),
			ui.NewEchoRoute(http.MethodPost, path+"/:press", h.PostUmbauPage),
		},
	)
}

func (h *Handler) GetUmbauPage(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	pressNumberParam, merr := utils.ParseParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	slog.Info("Rendering the umbau page", "user_name", user.Name)

	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.Page(&templates.PageProps{
		PressNumber: pressNumber,
		User:        user,
		Tools:       tools,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Umbau Page")
	}

	return nil
}

func (h *Handler) PostUmbauPage(c echo.Context) error {
	slog.Info("Change the active tools")

	// Get the user from the request context
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get the press number from the request context
	pressNumberParam, merr := utils.ParseParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	// Validate the press number
	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	// Get form value for the press cycles
	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"missing total cycles",
		)
	}

	// Parse this press cycles to int64
	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid total cycles",
		)
	}

	// Get form value for the top tool
	id, err := strconv.ParseInt(c.FormValue("top"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"missing top tool",
		)
	}

	topToolID := models.ToolID(id)

	// Get form value for the bottom tool
	id, err = strconv.ParseInt(c.FormValue("bottom"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"missing bottom tool",
		)
	}

	bottomToolID := models.ToolID(id)

	// Get a list with all tools
	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	// Get the top tool
	topTool, merr := h.findToolByID(tools, topToolID)
	if merr != nil {
		return merr.Echo()
	}

	// Get the bottom tool
	bottomTool, merr := h.findToolByID(tools, bottomToolID)
	if merr != nil {
		return merr.Echo()
	}

	// Check if the tools are compatible with each other
	if topTool.Format.String() != bottomTool.Format.String() {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"tools are not compatible",
		)
	}

	// Get current tools for press
	currentTools, merr := h.registry.Tools.ListByPress(&pressNumber)
	if merr != nil {
		return merr.Echo()
	}

	// Create final cycle entries for current tools with total cycles
	for _, tool := range currentTools {
		cycle := models.NewCycle(pressNumber, tool.ID, tool.Position, totalCycles, user.TelegramID)

		_, merr = h.registry.PressCycles.Add(
			cycle.PressNumber, cycle.ToolID, cycle.ToolPosition, cycle.TotalCycles, cycle.PerformedBy,
		)
		if merr != nil {
			return merr.Echo()
		}
	}

	// Unassign current tools from press
	for _, tool := range currentTools {
		merr := h.registry.Tools.UpdatePress(tool.ID, nil, user)
		if merr != nil {
			return merr.Echo()
		}
	}

	// Assign new tools to press
	newTools := []*models.Tool{topTool, bottomTool}
	for _, tool := range newTools {
		merr := h.registry.Tools.UpdatePress(tool.ID, &pressNumber, user)
		if merr != nil {
			return merr.Echo()
		}
	}

	// Create feed entry
	title := fmt.Sprintf("Werkzeugwechsel Presse %d", pressNumber)
	content := fmt.Sprintf(
		"Umbau abgeschlossen für Presse %d.\nEingebautes Oberteil: %s\nEingebautes Unterteil: %s\nGesamtzyklen: %d",
		pressNumber, topTool.String(), bottomTool.String(), totalCycles,
	)

	// Create feed entry
	merr = h.registry.Feeds.Add(title, content, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed", "press", pressNumber, "error", merr)
	}

	return nil
}

func (h *Handler) findToolByID(tools []*models.Tool, toolID models.ToolID) (*models.Tool, *errors.MasterError) {
	for _, tool := range tools {
		if tool.ID == toolID {
			return tool, errors.NewMasterError(
				fmt.Errorf("tool with ID %d not found in (%d) tools", len(tools), toolID),
				http.StatusBadRequest,
			)
		}
	}
	return nil, errors.NewMasterError(
		fmt.Errorf("tool not found: %d", toolID), http.StatusBadRequest,
	)
}
