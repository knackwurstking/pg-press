package html

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/troublereportspage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type TroubleReports struct {
	DB *database.DB
}

// TODO: Add "/trouble-reports/modifications/:id" the view
func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports", h.handleTroubleReports),
			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/share-pdf", h.handleGetSharePdf),
			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/attachment", h.handleGetAttachment),
		},
	)
}

func (h *TroubleReports) handleTroubleReports(c echo.Context) error {
	logger.HandlerTroubleReports().Debug("Rendering trouble reports page")

	page := troublereportspage.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HandlerTroubleReports().Error("Failed to render trouble reports page: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble reports page: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleGetSharePdf(c echo.Context) error {
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		logger.HandlerTroubleReports().Error("Invalid trouble report ID parameter: %v", err)
		return err
	}

	logger.HandlerTroubleReports().Info("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		logger.HandlerTroubleReports().Error(
			"Failed to retrieve trouble report %d for PDF generation: %v",
			id, err,
		)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to retrieve trouble report: "+err.Error())
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		logger.HandlerTroubleReports().Error(
			"Failed to generate PDF for trouble report %d: %v", tr.ID, err,
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Fehler beim Erstellen der PDF",
		)
	}

	logger.HandlerTroubleReports().Info(
		"Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, pdfBuffer.Len())

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *TroubleReports) shareResponse(
	c echo.Context,
	tr *models.TroubleReportWithAttachments,
	buf *bytes.Buffer,
) error {
	filename := fmt.Sprintf("fehlerbericht_%d_%s.pdf",
		tr.ID, time.Now().Format("2006-01-02"))

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Length",
		fmt.Sprintf("%d", buf.Len()))

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}

func (h *TroubleReports) handleGetAttachment(c echo.Context) error {
	attachmentID, err := helpers.ParseInt64Query(c, "attachment_id")
	if err != nil {
		logger.HandlerTroubleReports().Error("Invalid attachment ID parameter: %v", err)
		return err
	}

	logger.HandlerTroubleReports().Debug("Fetching attachment %d", attachmentID)

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		logger.HandlerTroubleReports().Error("Failed to get attachment %d: %v", attachmentID, err)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get attachment: "+err.Error())
	}

	// Set appropriate headers
	c.Response().Header().Set("Content-Type", attachment.MimeType)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(attachment.Data)))

	// Try to determine filename from attachment ID
	filename := fmt.Sprintf("attachment_%d", attachmentID)
	if ext := attachment.GetFileExtension(); ext != "" {
		filename += ext
	}
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filename))

	logger.HandlerTroubleReports().Info("Serving attachment %d (size: %d bytes, type: %s)",
		attachmentID, len(attachment.Data), attachment.MimeType)

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}
