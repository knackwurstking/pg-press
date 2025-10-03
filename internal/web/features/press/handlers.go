package press

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/web/features/press/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(
			db,
			logger.NewComponentLogger("Press"),
		),
	}
}

func (h *Handler) GetPressPage(c echo.Context) error {
	// Get press number from param
	var pn models.PressNumber
	pns, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse id: "+err.Error())
	}
	pn = models.PressNumber(pns)
	if !models.IsValidPressNumber(&pn) {
		return h.RenderBadRequest(c, fmt.Sprintf("invalid press number: %d", pn))
	}

	// Render page
	page := templates.Page(templates.PageProps{
		Press: pn,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press page: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetPressActiveTools(c echo.Context) error {
	pressNum, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	// Filter tools for this press and create toolsMap
	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		if tool.Press != nil && *tool.Press == press {
			toolsMap[tool.ID] = tool
		}
	}

	activeToolsSection := templates.PressActiveToolsSection(toolsMap, press)
	if err := activeToolsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render active tools section: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetPressMetalSheets(c echo.Context) error {
	pressNum, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get tools for this press
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	// Filter tools for this press and create toolsMap
	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		if tool.Press != nil && *tool.Press == press {
			toolsMap[tool.ID] = tool
		}
	}

	// Get metal sheets for tools on this press with automatic machine type filtering
	// Press 0 and 5 use SACMI machines, all others use SITI machines
	metalSheets, err := h.DB.MetalSheets.GetForPress(press, toolsMap)
	if err != nil {
		return h.HandleError(c, err, "failed to get metal sheets for press")
	}

	metalSheetsSection := templates.MetalSheetsSection(metalSheets, toolsMap, press)
	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render metal sheets section: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetPressCycles(c echo.Context) error {
	pressNum, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get user for permissions
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get cycles for this press
	cycles, err := h.DB.PressCycles.GetPressCycles(press, nil, nil)
	if err != nil {
		return h.HandleError(c, err, "failed to get cycles from database")
	}

	// Get tools for this press to create toolsMap
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		toolsMap[tool.ID] = tool
	}

	cyclesSection := templates.PressCyclesSection(cycles, toolsMap, user, press)
	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render cycles section: "+err.Error())
	}

	return nil
}
