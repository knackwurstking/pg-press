package press

import (
	"fmt"
	"net/http"
	"time"

	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/press/templates"
	"github.com/knackwurstking/pg-press/internal/pdf"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"

	"github.com/labstack/echo/v4"
)

func GetPage(c echo.Context) *echo.HTTPError {
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	pressNumber, merr := shared.ParseParamInt8(c, "press")
	if merr != nil {
		return merr.Echo()
	}

	t := Page(PageProps{
		PressNumber: shared.PressNumber(pressNumber),
		User:        user,
	})
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Page")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Old Press Page Handlers, Can be removed after migration
// -----------------------------------------------------------------------------

func (h *Handler) GetPressMetalSheets(c echo.Context) *echo.HTTPError {
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
	cycles, merr := h.registry.PressCycles.ListPressCyclesByPress(press, -1, 0)
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
	notes, merr := h.registry.Notes.ListByLinked("press", int64(press))
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
		n, merr := h.registry.Notes.ListByLinked("tool", int64(t.ID))
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
