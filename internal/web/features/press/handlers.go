package press

import (
	"fmt"
	"net/http"
	"time"

	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/press/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *services.Registry) *Handler {
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
	tools, err := h.DB.Tools.List()
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

func (h *Handler) HTMXGetPressActiveTools(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
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
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Render page
	page := templates.Page(templates.PageProps{
		Press: press,
	})

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press page: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetPressMetalSheets(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
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
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
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
	tools, err := h.DB.Tools.List()
	if err != nil {
		return h.HandleError(c, err, "failed to get tools from database")
	}

	toolsMap := make(map[int64]*models.Tool)
	for _, t := range tools {
		tool := t
		toolsMap[tool.ID] = tool
	}

	cyclesSection := templates.PressCyclesSection(cycles, toolsMap, user, press)
	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render cycles section: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetPressNotes(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get notes directly linked to this press
	notes, err := h.DB.Notes.GetByPress(press)
	if err != nil {
		return h.HandleError(c, err, "failed to get notes for press")
	}

	// Get tools for this press for context
	sortedTools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return h.HandleError(c, err, "failed to get tools for press")
	}

	// Get notes for tools
	for _, t := range sortedTools {
		n, err := h.DB.Notes.GetByTool(t.ID)
		if err != nil {
			return h.HandleError(c, err, fmt.Sprintf("failed to get notes for tool %d", t.ID))
		}
		notes = append(notes, n...)
	}

	notesSection := templates.PressNotesSection(notes, sortedTools, press)
	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render press notes section: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetCycleSummaryPDF(c echo.Context) error {
	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Get user for logging purposes
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.Log.Info("Generating cycle summary PDF for press %d requested by user %s", press, user.Name)

	// Get cycle summary data using service
	cycles, toolsMap, usersMap, err := h.DB.PressCycles.GetCycleSummaryData(press, h.DB.Tools, h.DB.Users)
	if err != nil {
		return h.HandleError(c, err, "failed to get cycle summary data")
	}

	// Generate PDF
	pdfBuffer, err := pdf.GenerateCycleSummaryPDF(press, cycles, toolsMap, usersMap)
	if err != nil {
		return h.HandleError(c, err, "failed to generate PDF")
	}

	// Set response headers
	filename := fmt.Sprintf("press_%d_cycle_summary_%s.pdf", press, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	return c.Stream(http.StatusOK, "application/pdf", pdfBuffer)
}

// ********************** //
// Request Helper Methods //
// ********************** //

func (h *Handler) getPressNumberFromParam(c echo.Context) (models.PressNumber, error) {
	pressNum, err := h.ParseInt8Param(c, "press")
	if err != nil {
		return -1, h.RenderBadRequest(c, "invalid or missing press parameter: "+err.Error())
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, h.RenderBadRequest(c, "invalid press number")
	}

	return press, nil
}
