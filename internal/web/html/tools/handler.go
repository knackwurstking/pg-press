package tools

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/env"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/handlers"
	"github.com/knackwurstking/pgpress/internal/web/helpers"

	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Tools struct {
	*handlers.BaseHandler
}

func NewTools(db *database.DB) *Tools {
	return &Tools{
		BaseHandler: handlers.NewBaseHandler(db, logger.HandlerTools()),
	}
}

func (h *Tools) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/tools",
				h.HandleTools),

			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press",
				h.HandlePressPage),

			helpers.NewEchoRoute(http.MethodGet, "/tools/press/:press/umbau",
				h.HandleUmbauPage),
			helpers.NewEchoRoute(http.MethodPost, "/tools/press/:press/umbau",
				h.HandleUmbauPagePOST),

			helpers.NewEchoRoute(http.MethodGet, "/tools/tool/:id",
				h.HandleToolPage),
		},
	)
}

func (h *Tools) HandleTools(c echo.Context) error {
	h.LogInfo("Rendering tools page")

	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools")
	}

	h.LogDebug("Retrieved %d tools", len(tools))

	pressUtilization, err := h.DB.Tools.GetPressUtilization()
	if err != nil {
		return h.HandleError(c, err, "failed to get press utilization")
	}

	page := ToolsPage(&ToolsPageProps{
		Tools:            tools,
		PressUtilization: pressUtilization,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tools page: "+err.Error())
	}

	return nil
}

func (h *Tools) HandlePressPage(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get press number from param
	var pn models.PressNumber
	pns, err := h.ParseInt64Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return h.RenderBadRequest(c, fmt.Sprintf("invalid press number: %d", pn))
	}

	// Get cycles for this press
	cycles, err := h.DB.PressCycles.GetPressCycles(pn, nil, nil)
	if err != nil {
		return h.HandleError(c, err, "failed to get press cycles")
	}

	// Get tools
	tools, err := h.DB.Tools.List()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools map")
	}
	// Convert tools to map[int64]*Tool
	toolsMap := make(map[int64]*models.Tool)
	for _, tool := range tools {
		toolsMap[tool.ID] = tool
	}

	// Render page
	h.LogDebug("Rendering page for press %d", pn)
	page := PressPage(PressPageProps{
		Press:    pn,
		Cycles:   cycles,
		User:     user,
		ToolsMap: toolsMap,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press page: "+err.Error())
	}

	return nil
}

func (h *Tools) HandleUmbauPage(c echo.Context) error {
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

	umbaupage := UmbauPage(&UmbauPageProps{
		PressNumber: pn,
		User:        user,
		Tools:       tools,
	})

	if err := umbaupage.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press umbau page: "+err.Error())
	}

	return nil
}

func (h *Tools) HandleUmbauPagePOST(c echo.Context) error {
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
	totalCyclesStr := c.FormValue("press-total-cycles")
	if totalCyclesStr == "" {
		return h.RenderBadRequest(c, "missing total cycles")
	}

	totalCycles, err := strconv.ParseInt(totalCyclesStr, 10, 64)
	if err != nil {
		return h.RenderBadRequest(c, "invalid total cycles: "+err.Error())
	}

	topToolStr := c.FormValue("top")
	if topToolStr == "" {
		return h.RenderBadRequest(c, "missing top tool")
	}

	bottomToolStr := c.FormValue("bottom")
	if bottomToolStr == "" {
		return h.RenderBadRequest(c, "missing bottom tool")
	}

	topCassetteToolStr := c.FormValue("top-cassette") // Optional

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

	var topCassetteTool *models.Tool
	if topCassetteToolStr != "" {
		topCassetteTool, err = h.findToolByString(tools, topCassetteToolStr, models.PositionTopCassette)
		if err != nil {
			return h.RenderBadRequest(c, "invalid top cassette tool: "+err.Error())
		}
	}

	// Get currently assigned tools for this press
	currentTools, err := h.DB.Tools.GetByPress(&pn)
	if err != nil {
		return h.HandleError(c, err, "failed to get current tools for press")
	}

	// Create final cycle entries for current tools (being removed) with the total cycles
	for _, tool := range currentTools {
		cycle := &models.Cycle{
			PressNumber:  pn,
			ToolID:       tool.ID,
			ToolPosition: tool.Position,
			TotalCycles:  totalCycles,
		}

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
	if topCassetteTool != nil {
		toolsToAssign = append(toolsToAssign, topCassetteTool)
	}

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
		"Umbau abgeschlossen für Presse %d. \n"+
			"Eingebautes Oberteil: %s\n"+
			"Eingebautes Unterteil: %s",
		pn, topTool.String(), bottomTool.String(),
	)
	if topCassetteTool != nil {
		content += fmt.Sprintf("\nEingebaute Obere Kassette: %s", topCassetteTool.String())
	}
	content += fmt.Sprintf("\nGesamtzyklen: %d", totalCycles)

	feed := models.NewFeed(title, content, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for press %d: %v", pn, err)
	}

	h.LogInfo("Successfully completed tool change for press %d", pn)

	// Set redirect header for HTMX
	c.Response().Header().Set("HX-Redirect", fmt.Sprintf("%s/tools/press/%d",
		env.ServerPathPrefix, pn))
	return c.NoContent(http.StatusOK)
}

// findToolByString finds a tool by its string representation and position
func (h *Tools) findToolByString(tools []*models.Tool, toolStr string, position models.Position) (*models.Tool, error) {
	for _, tool := range tools {
		if tool.Position == position && tool.String() == toolStr {
			return tool, nil
		}
	}
	return nil, fmt.Errorf("tool not found: %s", toolStr)
}

func (h *Tools) HandleToolPage(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	id, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c,
			"failed to parse id from query parameter:"+err.Error())
	}

	h.LogDebug("Fetching tool %d with notes", id)

	tool, err := h.DB.Tools.GetWithNotes(id)
	if err != nil {
		return h.HandleError(c, err, "failed to get tool")
	}

	h.LogDebug("Successfully fetched tool %d: Type=%s, Code=%s",
		id, tool.Type, tool.Code)

	// Fetch metal sheets assigned to this tool
	metalSheets, err := h.DB.MetalSheets.GetByToolID(id)
	if err != nil {
		// Log error but don't fail - metal sheets are supplementary data
		h.LogError("Failed to fetch metal sheets: %v", err)
		metalSheets = []*models.MetalSheet{}
	}

	h.LogDebug("Rendering tool page for tool %d with %d metal sheets", id, len(metalSheets))

	page := ToolPage(user, tool, metalSheets)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render tool page: "+err.Error())
	}

	return nil
}
