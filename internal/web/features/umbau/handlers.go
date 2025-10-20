package umbau

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/umbau/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/base"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*base.Handler
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		Handler: base.NewHandler(
			db,
			logger.NewComponentLogger("Umbau"),
		),
	}
}

func (h *Handler) GetUmbauPage(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	tools, err := h.DB.Tools.List()
	if err != nil {
		return h.HandleError(c, err, "failed to list tools")
	}

	umbaupage := templates.Page(&templates.PageProps{
		PressNumber: pn,
		User:        user,
		Tools:       tools,
	})

	if err := umbaupage.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press umbau page: "+err.Error())
	}

	return nil
}

func (h *Handler) PostUmbauPage(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Parse form values
	totalCyclesStr := h.GetSanitizedFormValue(c, "press-total-cycles")
	if totalCyclesStr == "" {
		return h.RenderBadRequest(c, "missing total cycles")
	}

	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return h.RenderBadRequest(c, "invalid total cycles: "+err.Error())
	}

	topToolStr := h.GetSanitizedFormValue(c, "top")
	if topToolStr == "" {
		return h.RenderBadRequest(c, "missing top tool")
	}

	bottomToolStr := h.GetSanitizedFormValue(c, "bottom")
	if bottomToolStr == "" {
		return h.RenderBadRequest(c, "missing bottom tool")
	}

	// Get all tools to find by string representation
	tools, err := h.DB.Tools.List()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools")
	}

	// Find tools by their string representation
	topTool, err := h.findToolByString(tools, topToolStr, models.PositionTop)
	if err != nil {
		return h.RenderBadRequest(c, "invalid top tool: "+err.Error())
	}

	bottomTool, err := h.findToolByString(tools, bottomToolStr, models.PositionBottom)
	if err != nil {
		return h.RenderBadRequest(c, "invalid bottom tool: "+err.Error())
	}

	// Get currently assigned tools for this press
	currentTools, err := h.DB.Tools.GetByPress(&pn)
	if err != nil {
		return h.HandleError(c, err, "failed to get current tools for press")
	}

	// Create final cycle entries for current tools (being removed) with the total cycles
	for _, tool := range currentTools {
		cycle := models.NewCycle(
			pn,
			tool.ID,
			tool.Position,
			totalCycles,
			user.TelegramID,
		)

		_, err := h.DB.PressCycles.Add(cycle, user)
		if err != nil {
			return h.HandleError(c, err, fmt.Sprintf("failed to create final cycle for outgoing tool %d", tool.ID))
		}
	}

	// Unassign current tools from press
	for _, tool := range currentTools {
		if err := h.DB.Tools.UpdatePress(tool.ID, nil, user); err != nil {
			return h.HandleError(c, err, fmt.Sprintf("failed to unassign tool %d", tool.ID))
		}
	}

	// Assign new tools to press (without creating initial cycles)
	toolsToAssign := []*models.Tool{topTool, bottomTool}
	for _, tool := range toolsToAssign {
		// Assign tool to press
		if err := h.DB.Tools.UpdatePress(tool.ID, &pn, user); err != nil {
			return h.HandleError(c, err,
				fmt.Sprintf("failed to assign tool %d to press", tool.ID))
		}
	}

	// Create a feed
	title := fmt.Sprintf("Werkzeugwechsel Presse %d", pn)
	content := fmt.Sprintf(
		"Umbau abgeschlossen für Presse %d.\n"+
			"Eingebautes Oberteil: %s\n"+
			"Eingebautes Unterteil: %s",
		pn, topTool.String(), bottomTool.String(),
	)
	content += fmt.Sprintf("\nGesamtzyklen: %d", totalCycles)

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error("Failed to create feed for press %d: %v", pn, err)
	}

	h.Log.Info("Successfully completed tool change for press %d", pn)

	return c.NoContent(http.StatusOK)
}

func (h *Handler) findToolByString(tools []*models.Tool, toolStr string, position models.Position) (*models.Tool, error) {
	for _, tool := range tools {
		if tool.Position == position && tool.String() == toolStr {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", toolStr)
}
