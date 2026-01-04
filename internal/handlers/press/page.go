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

	t := templates.Page(templates.PageProps{
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
