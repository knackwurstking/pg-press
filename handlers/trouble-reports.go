package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/components"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/pdf"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	registry *services.Registry
}

func NewTroubleReports(r *services.Registry) *TroubleReports {
	return &TroubleReports{
		registry: r,
	}
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		// Pages
		utils.NewEchoRoute(http.MethodGet, "/trouble-reports", h.GetPage),
		utils.NewEchoRoute(http.MethodGet, "/trouble-reports/share-pdf", h.GetSharePDF),
		utils.NewEchoRoute(http.MethodGet, "/trouble-reports/attachment", h.GetAttachment),
		utils.NewEchoRoute(http.MethodGet, "/trouble-reports/modifications/:id", h.GetModificationsForID),

		// HTMX
		utils.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/data", h.HTMXGetData),
		utils.NewEchoRoute(http.MethodDelete, "/htmx/trouble-reports/data", h.HTMXDeleteTroubleReport),
		utils.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/attachments-preview", h.HTMXGetAttachmentsPreview),
		utils.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/rollback", h.HTMXPostRollback),
	})
}

func (h *TroubleReports) GetPage(c echo.Context) error {
	page := components.PageTroubleReports()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render trouble reports page")
	}
	return nil
}

func (h *TroubleReports) GetSharePDF(c echo.Context) error {
	var troubleReportID models.TroubleReportID
	if id, err := ParseQueryInt64(c, "id"); err != nil {
		return HandleBadRequest(err, "missing id query parameter")
	} else {
		troubleReportID = models.TroubleReportID(id)
	}

	tr, err := h.registry.TroubleReports.GetWithAttachments(troubleReportID)
	if err != nil {
		return HandleError(err, "failed to retrieve trouble report")
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return HandleError(err, "failed to generate PDF")
	}

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *TroubleReports) GetAttachment(c echo.Context) error {
	var attachmentID models.AttachmentID
	if id, err := ParseQueryInt64(c, "attachment_id"); err != nil {
		return HandleBadRequest(err, "invalid attachment ID")
	} else {
		attachmentID = models.AttachmentID(id)
	}

	attachment, err := h.registry.Attachments.Get(attachmentID)
	if err != nil {
		return HandleError(err, "failed to get attachment")
	}

	// Set appropriate headers
	c.Response().Header().Set("Content-Type", attachment.MimeType)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(attachment.Data)))

	// Determine filename
	filename := fmt.Sprintf("attachment_%d", attachmentID)
	if ext := attachment.GetFileExtension(); ext != "" {
		filename += ext
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}

func (h *TroubleReports) GetModificationsForID(c echo.Context) error {
	id, err := ParseParamInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "invalid id in request")
	}

	modifications, err := h.registry.Modifications.List(
		models.ModificationTypeTroubleReport,
		id, 100, 0,
	)
	if err != nil {
		return HandleError(err, "failed to retrieve modifications")
	}

	// Convert to the format expected by the modifications page
	var resolvedModifications []*models.ResolvedModification[models.TroubleReportModData]
	for _, m := range modifications {
		resolvedModification, err := services.ResolveModification[models.TroubleReportModData](h.registry, m)
		if err != nil {
			slog.Error("Failed to resolve modification", "id", m.ID, "error", err)
			continue
		}
		resolvedModifications = append(resolvedModifications, resolvedModification)
	}

	// Render the page
	currentUser, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to retrieve user from context")
	}
	isAdmin := currentUser != nil && currentUser.IsAdmin()
	itemRenderFunc := components.TroubleReportCreateModificationRenderer(models.TroubleReportID(id), isAdmin)
	page := components.PageModifications[models.TroubleReportModData](resolvedModifications, itemRenderFunc)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render page")
	}

	return nil
}

func (h *TroubleReports) HTMXGetData(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	trs, err := h.registry.TroubleReports.ListWithAttachments()
	if err != nil {
		return HandleError(err, "failed to load trouble reports")
	}

	troubleReportsList := components.ListReports(user, trs)
	if err := troubleReportsList.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render trouble reports list component")
	}

	return nil
}

func (h *TroubleReports) HTMXDeleteTroubleReport(c echo.Context) error {
	var troubleReportID models.TroubleReportID
	if id, err := ParseQueryInt64(c, "id"); err != nil {
		return HandleBadRequest(err, "failed to parse trouble report ID")
	} else {
		troubleReportID = models.TroubleReportID(id)
	}

	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleBadRequest(err, "failed to get user from context")
	}

	removedReport, err := h.registry.TroubleReports.RemoveWithAttachments(troubleReportID, user)
	if err != nil {
		return HandleError(err, "failed to delete trouble report")
	}

	// Create feed entry
	feedTitle := "Problembericht gelöscht"
	feedContent := fmt.Sprintf("Titel: %s", removedReport.Title)
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed for trouble report deletion", "error", err)
	}

	return h.HTMXGetData(c)
}

func (h *TroubleReports) HTMXGetAttachmentsPreview(c echo.Context) error {
	var troubleReportID models.TroubleReportID
	if id, err := ParseQueryInt64(c, "id"); err != nil {
		return HandleBadRequest(err, "failed to parse ID from query")
	} else {
		troubleReportID = models.TroubleReportID(id)
	}

	tr, err := h.registry.TroubleReports.GetWithAttachments(troubleReportID)
	if err != nil {
		return HandleError(err, "failed to load trouble report")
	}

	attachmentsPreview := components.AttachmentsPreview(
		components.AttachmentPathTroubleReports,
		tr.LoadedAttachments,
	)

	if err := attachmentsPreview.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render attachments preview component")
	}

	return nil
}

func (h *TroubleReports) HTMXPostRollback(c echo.Context) error {
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	id, err := ParseQueryInt64(c, "id")
	if err != nil {
		return HandleBadRequest(err, "failed to parse ID query")
	}
	troubleReportID := models.TroubleReportID(id)

	modTimeStr := c.FormValue("modification_time")
	if modTimeStr == "" {
		return HandleBadRequest(nil, "modification_time form value is required")
	}

	modTime, err := strconv.ParseInt(modTimeStr, 10, 64)
	if err != nil {
		return HandleBadRequest(err, "invalid modification_time format")
	}

	modifications, err := h.registry.Modifications.ListAll(
		models.ModificationTypeTroubleReport,
		id,
	)
	if err != nil {
		return HandleError(err, "failed to retrieve modifications")
	}

	var targetMod *models.Modification[any]
	for _, mod := range modifications {
		if mod.CreatedAt.UnixMilli() == modTime {
			targetMod = mod
			break
		}
	}

	if targetMod == nil {
		return HandleNotFound(nil, fmt.Sprintf("modification %d not found", modTime))
	}

	var modData models.TroubleReportModData
	if err := json.Unmarshal(targetMod.Data, &modData); err != nil {
		return HandleError(err, "failed to parse modification data")
	}

	tr, err := h.registry.TroubleReports.Get(troubleReportID)
	if err != nil {
		return HandleError(err, "failed to retrieve trouble report")
	}

	if err := h.registry.TroubleReports.Update(modData.CopyTo(tr), user); err != nil {
		return HandleError(err, "failed to rollback trouble report")
	}

	// Create feed entry
	feedTitle := "Problembericht zurückgesetzt"
	feedContent := fmt.Sprintf("Titel: %s\nZurückgesetzt auf Version vom: %s",
		tr.Title, targetMod.CreatedAt.Format("2006-01-02 15:04:05"))

	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.registry.Feeds.Add(feed); err != nil {
		slog.Error("Failed to create feed for trouble report rollback", "error", err)
	}

	return nil
}

func (h *TroubleReports) shareResponse(
	c echo.Context,
	tr *models.TroubleReportWithAttachments,
	buf *bytes.Buffer,
) error {
	if buf == nil || buf.Len() == 0 {
		return HandleError(nil, "PDF buffer is empty")
	}

	sanitizedTitle := SanitizeFilename(tr.Title)
	if sanitizedTitle == "" {
		sanitizedTitle = "fehlerbericht"
	}

	filename := fmt.Sprintf("%s_%d_%s.pdf",
		sanitizedTitle, tr.ID, time.Now().Format("2006-01-02"))

	// Set headers
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	c.Response().Header().Set("Cache-Control", "private, max-age=0, no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("X-Frame-Options", "DENY")
	c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
	c.Response().Header().Set("Content-Description", "Trouble Report PDF")

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}
