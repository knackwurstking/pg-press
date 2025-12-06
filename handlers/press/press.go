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
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Render page
	t := templates.Page(press, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Press Page")
	}

	return nil
}

func (h *Handler) HTMXGetPressActiveTools(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get ordered tools for this press with validation
	tools, _, merr := h.getOrderedToolsForPress(press)
	if merr != nil {
		return merr.WrapEcho("get tools for press %d", press)
	}

	// Resolve tools, notes not needed, only the binding tool
	resolvedTools := make([]*models.ResolvedTool, 0, len(tools))
	for _, tool := range tools {
		rt, merr := services.ResolveTool(h.registry, tool)
		if merr != nil {
			return merr.WrapEcho("resolve tool %d", tool.ID)
		}
		resolvedTools = append(resolvedTools, rt)
	}

	t := templates.ActiveToolsSection(resolvedTools, press)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ActiveToolsSection")
	}

	return nil
}

func (h *Handler) HTMXGetPressMetalSheets(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get ordered tools for this press with validation
	_, toolsMap, merr := h.getOrderedToolsForPress(press)
	if merr != nil {
		return merr.WrapEcho("get tools for press %d", press)
	}

	// Get metal sheets for tools on this press with automatic machine type filtering
	// Press 0 and 5 use SACMI machines, all others use SITI machines
	metalSheets, merr := h.registry.MetalSheets.ListByPress(press, toolsMap)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.MetalSheetsSection(press, toolsMap, metalSheets)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "MetalSheetsSection")
	}

	return nil
}

func (h *Handler) HTMXGetPressCycles(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get user for permissions
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get cycles for this press
	cycles, merr := h.registry.PressCycles.GetPressCycles(press, nil, nil)
	if merr != nil {
		return merr.Echo()
	}

	// Get tools for this press to create toolsMap
	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return merr.Echo()
	}

	toolsMap := make(map[models.ToolID]*models.Tool)
	for _, t := range tools {
		tool := t
		toolsMap[tool.ID] = tool
	}

	t := templates.CyclesSection(cycles, toolsMap, user)
	if err := t.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.NewRenderError(err, "CyclesSection")
	}

	return nil
}

func (h *Handler) HTMXGetPressNotes(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get notes directly linked to this press
	notes, merr := h.registry.Notes.GetByPress(press)
	if merr != nil {
		return merr.Echo()
	}

	// Get tools for this press for context
	sortedTools, _, merr := h.getOrderedToolsForPress(press) // Get active tools
	if merr != nil {
		return merr.WrapEcho("get tools for press %d", press)
	}

	// Get notes for tools
	for _, t := range sortedTools {
		n, merr := h.registry.Notes.GetByTool(t.ID)
		if merr != nil {
			return merr.WrapEcho("get notes for tool %d", t.ID)
		}
		notes = append(notes, n...)
	}

	t := templates.NotesSection(press, notes, sortedTools)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "NotesSection")
	}

	return nil
}

func (h *Handler) HTMXGetPressRegenerations(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get press regenerations from service
	regenerations, merr := h.registry.PressRegenerations.GetRegenerationHistory(press)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.RegenerationsContent(regenerations, user)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "RegenerationsContent")
	}

	return nil
}

func (h *Handler) HTMXGetCycleSummaryPDF(c echo.Context) error {
	press, merr := h.getPressNumberFromParam(c)
	if merr != nil {
		return merr.Echo()
	}

	// Get cycle summary data using service
	cycles, toolsMap, usersMap, merr := h.registry.PressCycles.GetCycleSummaryData(press)
	if merr != nil {
		return merr.Echo()
	}

	// Generate PDF
	pdfBuffer, err := pdf.GenerateCycleSummaryPDF(press, cycles, toolsMap, usersMap)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "generate PDF"))
	}

	// Set response headers
	filename := fmt.Sprintf("press_%d_cycle_summary_%s.pdf", press, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	err = c.Stream(http.StatusOK, "application/pdf", pdfBuffer)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "stream"))
	}
	return nil
}

func (h *Handler) getPressNumberFromParam(c echo.Context) (models.PressNumber, *errors.MasterError) {
	pressNum, merr := utils.ParseParamInt8(c, "press")
	if merr != nil {
		return -1, merr
	}

	press := models.PressNumber(pressNum)
	if !models.IsValidPressNumber(&press) {
		return -1, errors.NewMasterError(fmt.Errorf("invalid press number"), http.StatusBadRequest)
	}

	return press, nil
}

func (h *Handler) getOrderedToolsForPress(press models.PressNumber) ([]*models.Tool, map[models.ToolID]*models.Tool, *errors.MasterError) {
	// Get tools from database
	tools, merr := h.registry.Tools.List()
	if merr != nil {
		return nil, nil, merr
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
	if !models.ValidateUniquePositions(pressTools) {
		return nil, nil, errors.NewMasterError(
			fmt.Errorf("tool duplicates in active tools list"),
			http.StatusBadRequest,
		)
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
