package troublereports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/troublereports/templates"
	"github.com/knackwurstking/pg-press/internal/pdf"
	"github.com/knackwurstking/pg-press/models"
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
	t := templates.Page()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Trouble Reports Page")
	}

	return nil
}

func (h *Handler) GetSharePDF(c echo.Context) error {
	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	troubleReportID := models.TroubleReportID(id)

	tr, merr := h.registry.TroubleReports.GetWithAttachments(troubleReportID)
	if merr != nil {
		return merr.Echo()
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			errors.Wrap(err, "generate PDF"),
		)
	}

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *Handler) GetAttachment(c echo.Context) error {
	id, merr := utils.ParseQueryInt64(c, "attachment_id")
	if merr != nil {
		return merr.Echo()
	}

	attachmentID := models.AttachmentID(id)

	attachment, merr := h.registry.Attachments.Get(attachmentID)
	if merr != nil {
		return merr.Echo()
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

	err := c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return nil
}

func (h *Handler) GetModificationsForID(c echo.Context) error {
	id, merr := utils.ParseParamInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}

	modifications, merr := h.registry.Modifications.List(
		models.ModificationTypeTroubleReport,
		id, 100, 0,
	)
	if merr != nil {
		return merr.Echo()
	}

	// Convert to the format expected by the modifications page
	var resolvedModifications []*models.ResolvedModification[models.TroubleReportModData]
	for _, m := range modifications {
		resolvedModification, merr := services.ResolveModification[models.TroubleReportModData](h.registry, m)
		if merr != nil {
			slog.Error("Failed to resolve modification", "id", m.ID, "error", merr)
			continue
		}
		resolvedModifications = append(resolvedModifications, resolvedModification)
	}

	// Render the page
	currentUser, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}
	isAdmin := currentUser != nil && currentUser.IsAdmin()
	itemRenderFunc := templates.CreateModificationRenderer(models.TroubleReportID(id), isAdmin)

	t := templates.ModificationsPage(resolvedModifications, itemRenderFunc)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Modifications Page")
	}

	return nil
}

func (h *Handler) HTMXGetData(c echo.Context) error {
	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	trs, merr := h.registry.TroubleReports.ListWithAttachments()
	if merr != nil {
		return merr.Echo()
	}

	t := templates.ListReports(user, trs)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "ListReports")
	}

	return nil
}

func (h *Handler) HTMXDeleteTroubleReport(c echo.Context) error {
	slog.Info("Remove a trouble report")

	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	troubleReportID := models.TroubleReportID(id)

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	removedReport, merr := h.registry.TroubleReports.RemoveWithAttachments(troubleReportID, user)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	feedTitle := "Problembericht gelöscht"
	feedContent := fmt.Sprintf("Titel: %s", removedReport.Title)

	merr = h.registry.Feeds.Add(feedTitle, feedContent, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for trouble report deletion", "error", merr)
	}

	err := h.HTMXGetData(c)
	if err != nil {
		return errors.NewMasterError(err, 0).Echo()
	}

	return nil
}

func (h *Handler) HTMXGetAttachmentsPreview(c echo.Context) error {
	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	troubleReportID := models.TroubleReportID(id)

	tr, merr := h.registry.TroubleReports.GetWithAttachments(troubleReportID)
	if merr != nil {
		return merr.Echo()
	}

	t := templates.AttachmentsPreview(
		templates.AttachmentPathTroubleReports,
		tr.LoadedAttachments,
	)
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "AttachmentsPreview")
	}

	return nil
}

func (h *Handler) HTMXPostRollback(c echo.Context) error {
	slog.Info("Rollback a trouble report to an other version")

	user, merr := utils.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	id, merr := utils.ParseQueryInt64(c, "id")
	if merr != nil {
		return merr.Echo()
	}
	troubleReportID := models.TroubleReportID(id)

	modTimeStr := c.FormValue("modification_time")
	if modTimeStr == "" {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"modification_time form value is required",
		)
	}

	modTime, err := strconv.ParseInt(modTimeStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"invalid modification_time format",
		)
	}

	modifications, merr := h.registry.Modifications.ListAll(
		models.ModificationTypeTroubleReport,
		id,
	)
	if merr != nil {
		return merr.Echo()
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
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"parse modification data",
		)
	}

	tr, merr := h.registry.TroubleReports.Get(troubleReportID)
	if merr != nil {
		return merr.Echo()
	}

	merr = h.registry.TroubleReports.Update(modData.CopyTo(tr), user)
	if merr != nil {
		return merr.Echo()
	}

	// Create feed entry
	feedTitle := "Problembericht zurückgesetzt"
	feedContent := fmt.Sprintf("Titel: %s\nZurückgesetzt auf Version vom: %s",
		tr.Title, targetMod.CreatedAt.Format("2006-01-02 15:04:05"))

	merr = h.registry.Feeds.Add(feedTitle, feedContent, user.TelegramID)
	if merr != nil {
		slog.Warn("Failed to create feed for trouble report rollback", "error", merr)
	}

	return nil
}

func (h *Handler) shareResponse(
	c echo.Context, tr *models.TroubleReportWithAttachments, buf *bytes.Buffer,
) *echo.HTTPError {
	if buf == nil || buf.Len() == 0 {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"PDF buffer is empty",
		)
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

	err := c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			err,
		)
	}

	return nil
}
