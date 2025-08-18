package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/htmxhandler"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/templates/pages"
	"github.com/knackwurstking/pgpress/internal/utils"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	path := "/trouble-reports"
	e.GET(serverPathPrefix+path, h.handleMainPage)

	sharePdfPath := serverPathPrefix + path + "/share-pdf"
	e.GET(sharePdfPath, h.handleGetSharePdf)

	// Attachment routes
	attachmentsPath := serverPathPrefix + "/trouble-reports/attachments"
	e.GET(attachmentsPath, h.handleGetAttachment)

	htmxTroubleReports := htmxhandler.TroubleReports{DB: h.DB}
	htmxTroubleReports.RegisterRoutes(e)
}

func (h *TroubleReports) handleMainPage(c echo.Context) error {
	page := pages.TroubleReportsPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render trouble reports page: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleGetSharePdf(c echo.Context) error {
	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.TroubleReport().Info("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReportsHelper.GetWithAttachments(id)
	if err != nil {
		logger.TroubleReport().Error(
			"Failed to retrieve trouble report %d for PDF generation: %v",
			id, err,
		)
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to retrieve trouble report: "+err.Error())
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		logger.TroubleReport().Error(
			"Failed to generate PDF for trouble report %d: %v", tr.ID, err,
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Fehler beim Erstellen der PDF",
		)
	}

	logger.TroubleReport().Info(
		"Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, pdfBuffer.Len())

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *TroubleReports) shareResponse(
	c echo.Context,
	tr *database.TroubleReportWithAttachments,
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
	attachmentID, err := utils.ParseInt64Query(c, constants.QueryParamAttachmentID)
	if err != nil {
		return err
	}

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
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

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}
