package press

import (
	"fmt"
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/press/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/pdf"
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
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		// Press page
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press", h.GetPressPage),

		// HTMX endpoints for press content
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press/active-tools", h.HTMXGetPressActiveTools),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press/metal-sheets", h.HTMXGetPressMetalSheets),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press/cycles", h.HTMXGetPressCycles),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press/notes", h.HTMXGetPressNotes),
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press/press-regenerations", h.HTMXGetPressRegenerations),

		// PDF Handlers
		ui.NewEchoRoute(http.MethodGet,
			path+"/:press/cycle-summary-pdf", h.HTMXGetCycleSummaryPDF),
	})
}

func (h *Handler) GetPressPage(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	press, err := h.getPressNumberFromParam(c)
	if err != nil {
		return err
	}

	// Render page
	page := templates.Page(press, user)

	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press page")
	}

	return nil
}

func (h *Handler) HTMXGetPressActiveTools(c echo.Context) error {
	press, eerr := h.getPressNumberFromParam(c)
	if eerr != nil {
		return eerr
	}

	// Get ordered tools for this press with validation
	tools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return errors.Handler(err, "get tools for press")
	}

	// Resolve tools, notes not needed, only the binding tool
	resolvedTools := make([]*models.ResolvedTool, 0, len(tools))
	for _, tool := range tools {
		if tool.Position == models.PositionTopCassette {
			continue
		}

		rt, err := services.ResolveTool(h.registry, tool)
		if err != nil {
			return errors.Handler(err, "resolve tool %d", tool.ID)
		}
		resolvedTools = append(resolvedTools, rt)
	}

	activeToolsSection := templates.ActiveToolsSection(resolvedTools, press)
	if err := activeToolsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render active tools section")
	}

	return nil
}

func (h *Handler) HTMXGetPressMetalSheets(c echo.Context) error {
	press, eerr := h.getPressNumberFromParam(c)
	if eerr != nil {
		return eerr
	}

	// Get ordered tools for this press with validation
	_, toolsMap, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return errors.Handler(err, "get tools for press")
	}

	// Get metal sheets for tools on this press with automatic machine type filtering
	// Press 0 and 5 use SACMI machines, all others use SITI machines
	metalSheets, err := h.registry.MetalSheets.GetForPress(press, toolsMap)
	if err != nil {
		return errors.Handler(err, "get metal sheets for press")
	}

	metalSheetsSection := templates.MetalSheetsSection(press, toolsMap, metalSheets)
	if err := metalSheetsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render metal sheets section")
	}

	return nil
}

func (h *Handler) HTMXGetPressCycles(c echo.Context) error {
	press, eerr := h.getPressNumberFromParam(c)
	if eerr != nil {
		return eerr
	}

	// Get user for permissions
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Get cycles for this press
	cycles, err := h.registry.PressCycles.GetPressCycles(press, nil, nil)
	if err != nil {
		return errors.Handler(err, "get cycles from database")
	}

	// Get tools for this press to create toolsMap
	tools, err := h.registry.Tools.List()
	if err != nil {
		return errors.Handler(err, "get tools from database")
	}

	toolsMap := make(map[models.ToolID]*models.Tool)
	for _, t := range tools {
		tool := t
		toolsMap[tool.ID] = tool
	}

	cyclesSection := templates.CyclesSection(cycles, toolsMap, user)

	if err := cyclesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render cycles section")
	}

	return nil
}

func (h *Handler) HTMXGetPressNotes(c echo.Context) error {
	press, eerr := h.getPressNumberFromParam(c)
	if eerr != nil {
		return eerr
	}

	// Get notes directly linked to this press
	notes, err := h.registry.Notes.GetByPress(press)
	if err != nil {
		return errors.Handler(err, "get notes for press")
	}

	// Get tools for this press for context
	sortedTools, _, err := h.getOrderedToolsForPress(press)
	if err != nil {
		return errors.Handler(err, "get tools for press")
	}

	// Get notes for tools
	for _, t := range sortedTools {
		n, err := h.registry.Notes.GetByTool(t.ID)
		if err != nil {
			return errors.Handler(err, "get notes for tool %d", t.ID)
		}
		notes = append(notes, n...)
	}

	notesSection := templates.NotesSection(press, notes, sortedTools)

	if err := notesSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press notes section")
	}

	return nil
}

func (h *Handler) HTMXGetPressRegenerations(c echo.Context) error {
	press, eerr := h.getPressNumberFromParam(c)
	if eerr != nil {
		return eerr
	}

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Get press regenerations from service
	regenerations, err := h.registry.PressRegenerations.GetRegenerationHistory(press)
	if err != nil {
		if !errors.IsNotFoundError(err) {
			return errors.Handler(err, "get press regenerations")
		}
	}

	regenerationsSection := templates.RegenerationsContent(regenerations, user)
	if err := regenerationsSection.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render press regenerations section")
	}

	return nil
}

func (h *Handler) HTMXGetCycleSummaryPDF(c echo.Context) error {
	press, eerr := h.getPressNumberFromParam(c)
	if eerr != nil {
		return eerr
	}

	// Get cycle summary data using service
	cycles, toolsMap, usersMap, err := h.registry.PressCycles.GetCycleSummaryData(press)
	if err != nil {
		return errors.Handler(err, "get cycle summary data")
	}

	// Generate PDF
	pdfBuffer, err := pdf.GenerateCycleSummaryPDF(press, cycles, toolsMap, usersMap)
	if err != nil {
		return errors.Handler(err, "generate PDF")
	}

	// Set response headers
	filename := fmt.Sprintf("press_%d_cycle_summary_%s.pdf", press, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	return c.Stream(http.StatusOK, "application/pdf", pdfBuffer)
}

func (h *Handler) getPressNumberFromParam(c echo.Context) (models.PressNumber, *echo.HTTPError) {
	pressNum, err := utils.ParseParamInt8(c, "press")
	if err != nil {
		return -1, errors.BadRequest(err, "invalid or missing press parameter")
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, errors.BadRequest(err, "invalid press number")
	}

	return press, nil
}

func (h *Handler) getOrderedToolsForPress(press models.PressNumber) ([]*models.Tool, map[models.ToolID]*models.Tool, error) {
	// Get tools from database
	tools, err := h.registry.Tools.List()
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
	toolsMap := make(map[models.ToolID]*models.Tool)
	for _, tool := range sortedTools {
		toolsMap[tool.ID] = tool
	}

	return sortedTools, toolsMap, nil
}
