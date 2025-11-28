package umbau

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/umbau/components"
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
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			utils.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau", h.GetUmbauPage),
			utils.NewEchoRoute(http.MethodPost, "/tools/press/:press/umbau", h.PostUmbauPage),
		},
	)
}

func (h *Handler) GetUmbauPage(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	pressNumberParam, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return errors.BadRequest(err, "parse press number")
	}

	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return errors.BadRequest(nil, "invalid press number: %d", pressNumber)
	}

	slog.Info("Rendering the umbau page", "user_name", user.Name)

	tools, err := h.registry.Tools.List()
	if err != nil {
		return errors.Handler(err, "list tools")
	}

	umbauPage := components.PageUmbau(&components.PageUmbauProps{
		PressNumber: pressNumber,
		User:        user,
		Tools:       tools,
	})

	if err := umbauPage.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press umbau page")
	}

	return nil
}

func (h *Handler) PostUmbauPage(c echo.Context) error {
	// Get the user from the request context
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Get the press number from the request context
	pressNumberParam, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return errors.BadRequest(err, "parse press number")
	}

	// Validate the press number
	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return errors.BadRequest(nil, "invalid press number: %d", pressNumber)
	}

	slog.Info("Handle a active tool change", "press", pressNumber, "user_name", user.Name)

	// Get form value for the press cycles
	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return errors.BadRequest(nil, "missing total cycles")
	}

	// Parse this press cycles to int64
	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return errors.BadRequest(err, "invalid total cycles")
	}

	// Get form value for the top tool
	var topToolID models.ToolID
	if id, err := strconv.ParseInt(c.FormValue("top"), 10, 64); err != nil {
		return errors.BadRequest(nil, "missing top tool")
	} else {
		topToolID = models.ToolID(id)
	}

	// Get form value for the bottom tool
	var bottomToolID models.ToolID
	if id, err := strconv.ParseInt(c.FormValue("bottom"), 10, 64); err != nil {
		return errors.BadRequest(nil, "missing bottom tool")
	} else {
		bottomToolID = models.ToolID(id)
	}

	// Get a list with all tools
	tools, err := h.registry.Tools.List()
	if err != nil {
		return errors.Handler(err, "get tools")
	}

	// Get the top tool
	topTool, err := h.findToolByID(tools, topToolID)
	if err != nil {
		return errors.BadRequest(err, "invalid top tool")
	}

	// Get the bottom tool
	bottomTool, err := h.findToolByID(tools, bottomToolID)
	if err != nil {
		return errors.BadRequest(err, "invalid bottom tool")
	}

	// Check if the tools are compatible with each other
	if topTool.Format.String() != bottomTool.Format.String() {
		return errors.BadRequest(nil, "tools are not compatible")
	}

	// Get current tools for press
	currentTools, err := h.registry.Tools.GetByPress(&pressNumber)
	if err != nil {
		return errors.Handler(err, "get current tools for press")
	}

	// Create final cycle entries for current tools with total cycles
	for _, tool := range currentTools {
		cycle := models.NewCycle(pressNumber, tool.ID, tool.Position, totalCycles, user.TelegramID)
		if _, err := h.registry.PressCycles.Add(cycle, user); err != nil {
			return errors.Handler(err, "create final cycle for tool %d", tool.ID)
		}
	}

	// Unassign current tools from press
	for _, tool := range currentTools {
		if err := h.registry.Tools.UpdatePress(tool.ID, nil, user); err != nil {
			return errors.Handler(err, "unassign tool %d", tool.ID)
		}
	}

	// Assign new tools to press
	newTools := []*models.Tool{topTool, bottomTool}
	for _, tool := range newTools {
		if err := h.registry.Tools.UpdatePress(tool.ID, &pressNumber, user); err != nil {
			return errors.Handler(err, "assign tool %d to press", tool.ID)
		}
	}

	// Create feed entry
	title := fmt.Sprintf("Werkzeugwechsel Presse %d", pressNumber)
	content := fmt.Sprintf(
		"Umbau abgeschlossen für Presse %d.\nEingebautes Oberteil: %s\nEingebautes Unterteil: %s\nGesamtzyklen: %d",
		pressNumber, topTool.String(), bottomTool.String(), totalCycles,
	)

	// Create feed entry
	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed", "press", pressNumber, "error", err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) findToolByID(tools []*models.Tool, toolID models.ToolID) (*models.Tool, error) {
	for _, tool := range tools {
		if tool.ID == toolID {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %d", toolID)
}
