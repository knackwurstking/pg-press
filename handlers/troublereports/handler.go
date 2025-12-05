package troublereports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/env"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/troublereports/templates"
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
		// Pages
		ui.NewEchoRoute(http.MethodGet, path, h.GetPage),
		ui.NewEchoRoute(http.MethodGet, path+"/share-pdf", h.GetSharePDF),
		ui.NewEchoRoute(http.MethodGet, path+"/attachment", h.GetAttachment),
		ui.NewEchoRoute(http.MethodGet, path+"/modifications/:id", h.GetModificationsForID),

		// HTMX
		ui.NewEchoRoute(http.MethodGet, path+"/data", h.HTMXGetData),
		ui.NewEchoRoute(http.MethodDelete, path+"/data", h.HTMXDeleteTroubleReport),
		ui.NewEchoRoute(http.MethodGet, path+"/attachments-preview", h.HTMXGetAttachmentsPreview),
		ui.NewEchoRoute(http.MethodPost, path+"/rollback", h.HTMXPostRollback),
	})
}

func (h *Handler) GetPage(c echo.Context) error {
	page := templates.Page()
	err := page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "TroubleReportsPage")
	}
	return nil
}

func (h *Handler) GetSharePDF(c echo.Context) error {
	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "missing id query parameter")
	}
	troubleReportID := models.TroubleReportID(id)

	tr, dberr := h.registry.TroubleReports.GetWithAttachments(troubleReportID)
	if dberr != nil {
		return errors.HandlerError(dberr, "retrieve trouble report")
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return errors.HandlerError(err, "generate PDF")
	}

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *Handler) GetAttachment(c echo.Context) error {
	var attachmentID models.AttachmentID
	if id, err := utils.ParseQueryInt64(c, "attachment_id"); err != nil {
		return errors.NewBadRequestError(err, "invalid attachment ID")
	} else {
		attachmentID = models.AttachmentID(id)
	}

	attachment, dberr := h.registry.Attachments.Get(attachmentID)
	if dberr != nil {
		return errors.HandlerError(dberr, "get attachment")
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

func (h *Handler) GetModificationsForID(c echo.Context) error {
	id, err := utils.ParseParamInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "invalid id in request")
	}

	modifications, dberr := h.registry.Modifications.List(
		models.ModificationTypeTroubleReport,
		id, 100, 0,
	)
	if dberr != nil {
		return errors.HandlerError(dberr, "retrieve modifications")
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
	currentUser, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}
	isAdmin := currentUser != nil && currentUser.IsAdmin()
	itemRenderFunc := templates.CreateModificationRenderer(models.TroubleReportID(id), isAdmin)

	page := templates.ModificationsPage(resolvedModifications, itemRenderFunc)
	err = page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ModificationsPage")
	}

	return nil
}

func (h *Handler) HTMXGetData(c echo.Context) error {
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	trs, dberr := h.registry.TroubleReports.ListWithAttachments()
	if dberr != nil {
		return errors.HandlerError(dberr, "load trouble reports")
	}

	troubleReportsList := templates.ListReports(user, trs)
	err := troubleReportsList.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ListReports")
	}

	return nil
}

func (h *Handler) HTMXDeleteTroubleReport(c echo.Context) error {
	slog.Info("Remove a trouble report")

	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse trouble report ID")
	}
	troubleReportID := models.TroubleReportID(id)

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	removedReport, dberr := h.registry.TroubleReports.RemoveWithAttachments(troubleReportID, user)
	if dberr != nil {
		return errors.HandlerError(dberr, "delete trouble report")
	}

	// Create feed entry
	feedTitle := "Problembericht gelöscht"
	feedContent := fmt.Sprintf("Titel: %s", removedReport.Title)
	if _, err := h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID); err != nil {
		slog.Warn("Failed to create feed for trouble report deletion", "error", err)
	}

	return h.HTMXGetData(c)
}

func (h *Handler) HTMXGetAttachmentsPreview(c echo.Context) error {
	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse ID from query")
	}
	troubleReportID := models.TroubleReportID(id)

	tr, dberr := h.registry.TroubleReports.GetWithAttachments(troubleReportID)
	if dberr != nil {
		return errors.HandlerError(dberr, "load trouble report")
	}

	attachmentsPreview := templates.AttachmentsPreview(
		templates.AttachmentPathTroubleReports,
		tr.LoadedAttachments,
	)

	err = attachmentsPreview.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "AttachmentsPreview")
	}

	return nil
}

func (h *Handler) HTMXPostRollback(c echo.Context) error {
	slog.Info("Rollback a trouble report to an other version")

	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	id, err := utils.ParseQueryInt64(c, "id")
	if err != nil {
		return errors.NewBadRequestError(err, "parse ID query")
	}
	troubleReportID := models.TroubleReportID(id)

	modTimeStr := c.FormValue("modification_time")
	if modTimeStr == "" {
		return errors.NewBadRequestError(nil, "modification_time form value is required")
	}

	modTime, err := strconv.ParseInt(modTimeStr, 10, 64)
	if err != nil {
		return errors.NewBadRequestError(err, "invalid modification_time format")
	}

	modifications, dberr := h.registry.Modifications.ListAll(
		models.ModificationTypeTroubleReport,
		id,
	)
	if dberr != nil {
		return errors.HandlerError(dberr, "retrieve modifications")
	}

	var targetMod *models.Modification[any]
	for _, mod := range modifications {
		if mod.CreatedAt.UnixMilli() == modTime {
			targetMod = mod
			break
		}
	}

	if targetMod == nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("modification %d not found", modTime))
	}

	var modData models.TroubleReportModData
	err = json.Unmarshal(targetMod.Data, &modData)
	if err != nil {
		return errors.HandlerError(err, "parse modification data")
	}

	tr, dberr := h.registry.TroubleReports.Get(troubleReportID)
	if dberr != nil {
		return errors.HandlerError(dberr, "retrieve trouble report")
	}

	dberr = h.registry.TroubleReports.Update(modData.CopyTo(tr), user)
	if dberr != nil {
		return errors.HandlerError(dberr, "rollback trouble report")
	}

	// Create feed entry
	feedTitle := "Problembericht zurückgesetzt"
	feedContent := fmt.Sprintf("Titel: %s\nZurückgesetzt auf Version vom: %s",
		tr.Title, targetMod.CreatedAt.Format("2006-01-02 15:04:05"))

	_, dberr = h.registry.Feeds.AddSimple(feedTitle, feedContent, user.TelegramID)
	if dberr != nil {
		slog.Warn("Failed to create feed for trouble report rollback", "error", dberr)
	}

	return nil
}

func (h *Handler) shareResponse(
	c echo.Context, tr *models.TroubleReportWithAttachments, buf *bytes.Buffer,
) error {
	if buf == nil || buf.Len() == 0 {
		return errors.HandlerError(nil, "PDF buffer is empty")
	}

	sanitizedTitle := utils.SanitizeFilename(tr.Title)
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
