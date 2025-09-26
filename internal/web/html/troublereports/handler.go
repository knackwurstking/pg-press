package troublereports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/html/modpage"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/internal/web/shared/helpers"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/modification"
)

type TroubleReports struct {
	*handlers.BaseHandler
}

func NewTroubleReports(db *database.DB) *TroubleReports {
	return &TroubleReports{
		BaseHandler: handlers.NewBaseHandler(db, logger.HandlerTroubleReports()),
	}
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports",
				h.HandleTroubleReportsGET),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/share-pdf",
				h.HandleSharePdfGET),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/attachment",
				h.HandleAttachmentGET),

			helpers.NewEchoRoute(http.MethodGet, "/trouble-reports/modifications/:id",
				h.HandleModificationsGET),
		},
	)
}

func (h *TroubleReports) HandleTroubleReportsGET(c echo.Context) error {
	h.LogDebug("Rendering trouble reports page")

	page := TroubleReportsPage()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render trouble reports page: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) HandleSharePdfGET(c echo.Context) error {
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	h.LogInfo("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve trouble report")
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return h.HandleError(c, err, "failed to generate PDF")
	}

	h.LogInfo("Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, pdfBuffer.Len())

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *TroubleReports) HandleAttachmentGET(c echo.Context) error {
	attachmentID, err := h.ParseInt64Query(c, "attachment_id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	h.LogDebug("Fetching attachment %d", attachmentID)

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		return h.HandleError(c, err, "failed to get attachment")
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

	h.LogInfo("Serving attachment %d (size: %d bytes, type: %s)",
		attachmentID, len(attachment.Data), attachment.MimeType)

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}

func (h *TroubleReports) HandleModificationsGET(c echo.Context) error {
	h.LogInfo("Handling modifications for trouble report")

	// Parse ID parameter
	id, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	// Fetch modifications for this trouble report
	h.LogDebug("Fetching modifications for trouble report %d", id)

	modifications, err := h.DB.Modifications.ListWithUser(
		services.ModificationTypeTroubleReport, id, 100, 0)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve modifications")
	}

	h.LogDebug(
		"Found %d modifications for trouble report %d",
		len(modifications), id)

	// Convert to the format expected by the modifications page
	var m modification.Mods[models.TroubleReportModData]
	for _, mod := range modifications {
		var data models.TroubleReportModData
		if err := json.Unmarshal(mod.Modification.Data, &data); err != nil {
			continue
		}

		modEntry := modification.NewMod(&mod.User, data)
		modEntry.Time = mod.Modification.CreatedAt.UnixMilli()
		m = append(m, modEntry)
	}

	// Get user from context to check permissions
	currentUser, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve user from context")
	}
	canRollback := currentUser != nil && currentUser.IsAdmin()

	// Create render function using the new template
	f := CreateModificationRenderer(id, canRollback)

	// Rendering the page template
	page := modpage.BasePage(m, f)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render page: "+err.Error())
	}

	return nil
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
