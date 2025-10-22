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

	pns, err := ParseParamInt8(c, "press")
	if err != nil {
		return HandleBadRequest(err, "failed to parse press number")
	}

	pn := models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return HandleBadRequest(fmt.Errorf("invalid press number: %d", pn), "invalid press number")
	}

	tools, err := h.Registry.Tools.List()
	if err != nil {
		return HandleError(err, "failed to list tools")
	}

	umbaupage := components.PageUmbau(&components.PageUmbauProps{
		PressNumber: pn,
		User:        user,
		Tools:       tools,
	})

	if err := umbaupage.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render press umbau page")
	}

	return nil
}

func (h *Umbau) PostUmbauPage(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	pns, err := ParseParamInt8(c, "press")
	if err != nil {
		return HandleBadRequest(err, "failed to parse press number")
	}

	pn := models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return HandleBadRequest(fmt.Errorf("invalid press number: %d", pn), "invalid press number")
	}

	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return HandleBadRequest(fmt.Errorf("missing total cycles"), "missing total cycles")
	}

	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return HandleBadRequest(err, "invalid total cycles")
	}

	topToolStr := c.FormValue("top")
	if topToolStr == "" {
		return HandleBadRequest(fmt.Errorf("missing top tool"), "missing top tool")
	}

	bottomToolStr := c.FormValue("bottom")
	if bottomToolStr == "" {
		return HandleBadRequest(fmt.Errorf("missing bottom tool"), "missing bottom tool")
	}

	tools, err := h.Registry.Tools.List()
	if err != nil {
		return HandleError(err, "failed to get tools")
	}

	topTool, err := h.findToolByString(tools, topToolStr, models.PositionTop)
	if err != nil {
		return HandleBadRequest(err, "invalid top tool")
	}

	bottomTool, err := h.findToolByString(tools, bottomToolStr, models.PositionBottom)
	if err != nil {
		return HandleBadRequest(err, "invalid bottom tool")
	}

	currentTools, err := h.Registry.Tools.GetByPress(&pn)
	if err != nil {
		return HandleError(err, "failed to get current tools for press")
	}

	// Create final cycle entries for current tools with total cycles
	for _, tool := range currentTools {
		cycle := models.NewCycle(pn, tool.ID, tool.Position, totalCycles, user.TelegramID)
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
		if err := h.Registry.Tools.UpdatePress(tool.ID, &pn, user); err != nil {
			return HandleError(err, fmt.Sprintf("failed to assign tool %d to press", tool.ID))
		}
	}

	// Create feed entry
	title := fmt.Sprintf("Werkzeugwechsel Presse %d", pn)
	content := fmt.Sprintf(
		"Umbau abgeschlossen für Presse %d.\nEingebautes Oberteil: %s\nEingebautes Unterteil: %s\nGesamtzyklen: %d",
		pn, topTool.String(), bottomTool.String(), totalCycles,
	)

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.Registry.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for press %d: %v", pn, err)
	}

	h.Log.Info("Successfully completed tool change for press %d", pn)

	return c.NoContent(http.StatusOK)
}

func (h *Umbau) findToolByString(tools []*models.Tool, toolStr string, position models.Position) (*models.Tool, error) {
	for _, tool := range tools {
		if tool.Position == position && tool.String() == toolStr {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", toolStr)
}
