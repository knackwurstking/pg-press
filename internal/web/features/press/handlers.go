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

// getOrderedToolsForPress gets tools for a press, validates unique positions, and returns ordered tools + toolsMap
func (h *Handler) getOrderedToolsForPress(press models.PressNumber) ([]*models.Tool, map[int64]*models.Tool, error) {
	// Get tools from database
	tools, err := h.DB.Tools.ListWithNotes()
	if err != nil {
		return nil, nil, err
	}

	// Filter tools for this press
	var pressTools []*models.Tool
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		if tool.Press != nil && *tool.Press == press {
			pressTools = append(pressTools, tool)
		}
	}

	// Validate that each position is unique (only one tool per position)
	if err := models.ValidateUniquePositions(pressTools); err != nil {
		return nil, nil, err
	}

	// Sort tools by position: top, top cassette, bottom
	sortedTools := models.SortToolsByPosition(pressTools)

	// Create toolsMap for lookup purposes
	toolsMap := make(map[int64]*models.Tool)
	for _, tool := range sortedTools {
		toolsMap[tool.ID] = tool
	}

	return sortedTools, toolsMap, nil
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

	// Get ordered tools for this press with validation
	sortedTools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return h.HandleError(c, err, "failed to get tools for press")
	}

	activeToolsSection := templates.PressActiveToolsSection(sortedTools, press)
	if err := activeToolsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render active tools section: "+err.Error())
	}

	return nil
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

func (h *Handler) HTMXGetPressMetalSheets(c echo.Context) error {
	pressNum, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get ordered tools for this press with validation
	_, toolsMap, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return h.HandleError(c, err, "failed to get tools for press")
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

func (h *Handler) HTMXGetPressNotes(c echo.Context) error {
	pressNum, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return h.RenderBadRequest(c, "invalid press number")
	}

	// Get ordered tools for this press with validation
	sortedTools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return h.HandleError(c, err, "failed to get tools for press")
	}

	// Get all notes for tools on this press
	var allNotes []*models.Note
	for _, tool := range sortedTools {
		if len(tool.LinkedNotes) > 0 {
			notes, err := h.DB.Notes.GetByIDs(tool.LinkedNotes)
			if err != nil {
				h.LogError("Failed to get notes for tool %d: %v", tool.ID, err)
				continue
			}
			allNotes = append(allNotes, notes...)
		}
	}

	notesSection := templates.PressNotesSection(allNotes, sortedTools, press)
	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press notes section: "+err.Error())
	}

	return nil
}
