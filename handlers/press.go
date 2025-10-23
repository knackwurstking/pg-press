package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/pdf"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Press struct {
	*Base
}

func NewPress(db *services.Registry) *Press {
	return &Press{
		Base: NewBase(db, logger.NewComponentLogger("Press")),
	}
}

func (h *Press) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Press page
		utils.NewEchoRoute(http.MethodGet,
			"/tools/press/:press", h.GetPressPage),

		// HTMX endpoints for press content
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/press/:press/active-tools", h.HTMXGetPressActiveTools),
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/press/:press/metal-sheets", h.HTMXGetPressMetalSheets),
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/press/:press/cycles", h.HTMXGetPressCycles),
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/press/:press/notes", h.HTMXGetPressNotes),

		// PDF Handlers
		utils.NewEchoRoute(http.MethodGet,
			"/htmx/tools/press/:press/cycle-summary-pdf", h.HTMXGetCycleSummaryPDF),
	})
}

func (h *Press) GetPressPage(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Render page
	page := components.PagePress(components.PagePressProps{
		Press: press,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render press page")
	}

	return nil
}

func (h *Press) HTMXGetPressActiveTools(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get ordered tools for this press with validation
	sortedTools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return HandleError(err, "failed to get tools for press")
	}

	activeToolsSection := components.PagePress_ActiveToolsSection(sortedTools, press)
	if err := activeToolsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render active tools section")
	}

	return nil
}

func (h *Press) HTMXGetPressMetalSheets(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get ordered tools for this press with validation
	_, toolsMap, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return HandleError(err, "failed to get tools for press")
	}

	// Get metal sheets for tools on this press with automatic machine type filtering
	// Press 0 and 5 use SACMI machines, all others use SITI machines
	metalSheets, err := h.Registry.MetalSheets.GetForPress(press, toolsMap)
	if err != nil {
		return HandleError(err, "failed to get metal sheets for press")
	}

	metalSheetsSection := components.PagePress_MetalSheetsSection(
		components.PagePress_MetalSheetSectionProps{
			MetalSheets: metalSheets,
			ToolsMap:    toolsMap,
			Press:       press,
		},
	)
	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render metal sheets section")
	}

	return nil
}

func (h *Press) HTMXGetPressCycles(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get user for permissions
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	// Get cycles for this press
	cycles, err := h.Registry.PressCycles.GetPressCycles(press, nil, nil)
	if err != nil {
		return HandleError(err, "failed to get cycles from database")
	}

	// Get tools for this press to create toolsMap
	tools, err := h.Registry.Tools.List()
	if err != nil {
		return HandleError(err, "failed to get tools from database")
	}

	toolsMap := make(map[int64]*models.Tool)
	for _, t := range tools {
		tool := t
		toolsMap[tool.ID] = tool
	}

	cyclesSection := components.PagePress_CyclesSection(
		components.PagePress_CyclesSectionProps{
			Cycles:   cycles,
			ToolsMap: toolsMap,
			User:     user,
			Press:    press,
		},
	)

	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render cycles section")
	}

	return nil
}

func (h *Press) HTMXGetPressNotes(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get notes directly linked to this press
	notes, err := h.Registry.Notes.GetByPress(press)
	if err != nil {
		return HandleError(err, "failed to get notes for press")
	}

	// Get tools for this press for context
	sortedTools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return HandleError(err, "failed to get tools for press")
	}

	// Get notes for tools
	for _, t := range sortedTools {
		n, err := h.Registry.Notes.GetByTool(t.ID)
		if err != nil {
			return HandleError(err, fmt.Sprintf("failed to get notes for tool %d", t.ID))
		}
		notes = append(notes, n...)
	}

	notesSection := components.PagePress_NotesSection(
		components.PagePress_NotesSectionProps{
			Notes: notes,
			Tools: sortedTools,
			Press: press,
		},
	)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render press notes section")
	}

	return nil
}

func (h *Press) HTMXGetCycleSummaryPDF(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get user for logging purposes
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	h.Log.Info("Generating cycle summary PDF for press %d requested by user %s", press, user.Name)

	// Get cycle summary data using service
	cycles, toolsMap, usersMap, err := h.Registry.PressCycles.GetCycleSummaryData(press)
	if err != nil {
		return HandleError(err, "failed to get cycle summary data")
	}

	// Generate PDF
	pdfBuffer, err := pdf.GenerateCycleSummaryPDF(press, cycles, toolsMap, usersMap)
	if err != nil {
		return HandleError(err, "failed to generate PDF")
	}

	// Set response headers
	filename := fmt.Sprintf("press_%d_cycle_summary_%s.pdf", press, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	return c.Stream(http.StatusOK, "application/pdf", pdfBuffer)
}

func (h *Press) getPressNumberFromParam(c echo.Context) (models.PressNumber, error) {
	pressNum, err := ParseParamInt8(c, "press")
	if err != nil {
		return -1, HandleBadRequest(err, "invalid or missing press parameter")
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, HandleBadRequest(err, "invalid press number")
	}

	return press, nil
}

func (h *Press) getOrderedToolsForPress(press models.PressNumber) ([]*models.Tool, map[int64]*models.Tool, error) {
	// Get tools from database
	tools, err := h.Registry.Tools.List()
	if err != nil {
		return nil, nil, err
	}

	// Filter tools for this press
	var pressTools []*models.Tool
	for _, t := range tools {
		tool := t
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
