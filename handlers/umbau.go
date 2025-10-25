package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Umbau struct {
	*Base
}

func NewUmbau(db *services.Registry) *Umbau {
	return &Umbau{
		Base: NewBase(db, logger.NewComponentLogger("Umbau")),
	}
}

func (h *Umbau) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(
		e,
		[]*utils.EchoRoute{
			utils.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau", h.GetUmbauPage),
			utils.NewEchoRoute(http.MethodPost, "/tools/press/:press/umbau", h.PostUmbauPage),
		},
	)
}

func (h *Umbau) GetUmbauPage(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	pressNumberParam, err := ParseParamInt8(c, "press")
	if err != nil {
		return HandleBadRequest(err, "failed to parse press number")
	}

	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return HandleBadRequest(nil, fmt.Sprintf("invalid press number: %d", pressNumber))
	}

	h.Log.Info("Rendering the umbau page for user %s", user.String())

	tools, err := h.Registry.Tools.List()
	if err != nil {
		return HandleError(err, "failed to list tools")
	}

	umbaupage := components.PageUmbau(&components.PageUmbauProps{
		PressNumber: pressNumber,
		User:        user,
		Tools:       tools,
	})

	if err := umbaupage.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render press umbau page")
	}

	return nil
}

func (h *Umbau) PostUmbauPage(c echo.Context) error {
	// Get the user from the request context
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	// Get the press number from the request context
	pressNumberParam, err := ParseParamInt8(c, "press")
	if err != nil {
		return HandleBadRequest(err, "failed to parse press number")
	}

	// Validate the press number
	pressNumber := models.PressNumber(pressNumberParam)
	if !models.IsValidPressNumber(&pressNumber) {
		return HandleBadRequest(nil, fmt.Sprintf("invalid press number: %d", pressNumber))
	}

	h.Log.Info("Handle a active tool change for the press %d by user %s", pressNumber, user.String())

	// Get form value for the press cycles
	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return HandleBadRequest(nil, "missing total cycles")
	}

	// Parse this press cycles to int64
	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return HandleBadRequest(err, "invalid total cycles")
	}

	// Get form value for the top tool
	topToolID, err := strconv.ParseInt(c.FormValue("top"), 10, 64)
	if err != nil {
		return HandleBadRequest(nil, "missing top tool")
	}

	// Get form value for the bottom tool
	bottomToolID, err := strconv.ParseInt(c.FormValue("bottom"), 10, 64)
	if err != nil {
		return HandleBadRequest(nil, "missing bottom tool")
	}

	// Get a list with all tools
	tools, err := h.Registry.Tools.List()
	if err != nil {
		return HandleError(err, "failed to get tools")
	}

	// Get the top tool
	topTool, err := h.findToolByID(tools, topToolID)
	if err != nil {
		return HandleBadRequest(err, "invalid top tool")
	}

	// Get the bottom tool
	bottomTool, err := h.findToolByID(tools, bottomToolID)
	if err != nil {
		return HandleBadRequest(err, "invalid bottom tool")
	}

	// Check if the tools are compatible with each other
	if topTool.Format.String() != bottomTool.Format.String() {
		return HandleBadRequest(nil, "tools are not compatible")
	}

	// Get current tools for press
	currentTools, err := h.Registry.Tools.GetByPress(&pressNumber)
	if err != nil {
		return HandleError(err, "failed to get current tools for press")
	}

	// Create final cycle entries for current tools with total cycles
	for _, tool := range currentTools {
		cycle := models.NewCycle(pressNumber, tool.ID, tool.Position, totalCycles, user.TelegramID)
		if _, err := h.Registry.PressCycles.Add(cycle, user); err != nil {
			return HandleError(err, fmt.Sprintf("failed to create final cycle for tool %d", tool.ID))
		}
	}

	// Unassign current tools from press
	for _, tool := range currentTools {
		if err := h.Registry.Tools.UpdatePress(tool.ID, nil, user); err != nil {
			return HandleError(err, fmt.Sprintf("failed to unassign tool %d", tool.ID))
		}
	}

	// Assign new tools to press
	newTools := []*models.Tool{topTool, bottomTool}
	for _, tool := range newTools {
		if err := h.Registry.Tools.UpdatePress(tool.ID, &pressNumber, user); err != nil {
			return HandleError(err, fmt.Sprintf("failed to assign tool %d to press", tool.ID))
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
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for press %d: %v", pressNumber, err)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Umbau) findToolByID(tools []*models.Tool, toolID int64) (*models.Tool, error) {
	for _, tool := range tools {
		if tool.ID == toolID {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %d", toolID)
}
