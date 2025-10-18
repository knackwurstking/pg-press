package troublereports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"strconv"
	"time"

	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/troublereports/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/base"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *services.Registry) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(db, logger.NewComponentLogger("Trouble Reports")),
	}
}

func (h *Handler) GetPage(c echo.Context) error {
	h.Log.Debug("Rendering trouble reports page")

	page := templates.Page()
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render trouble reports page: "+err.Error())
	}
	return nil
}

func (h *Handler) GetSharePDF(c echo.Context) error {
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	h.Log.Info("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve trouble report")
	}

	pdfBuffer, err := pdf.GenerateTroubleReportPDF(tr)
	if err != nil {
		return h.HandleError(c, err, "failed to generate PDF")
	}

	h.Log.Info("Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, pdfBuffer.Len())

	return h.shareResponse(c, tr, pdfBuffer)
}

func (h *Handler) GetAttachment(c echo.Context) error {
	attachmentID, err := h.ParseInt64Query(c, "attachment_id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	h.Log.Debug("Fetching attachment %d", attachmentID)

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

	h.Log.Info("Serving attachment %d (size: %d bytes, type: %s)",
		attachmentID, len(attachment.Data), attachment.MimeType)

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}

func (h *Handler) GetModificationsForID(c echo.Context) error {
	h.Log.Info("Handling modifications for trouble report")

	// Parse ID parameter
	id, err := h.ParseInt64Param(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, err.Error())
	}

	// Fetch modifications for this trouble report
	h.Log.Debug("Fetching modifications for trouble report %d", id)

	modifications, err := h.DB.Modifications.ListWithUser(
		models.ModificationTypeTroubleReport,
		id,
		100,
		0,
	)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve modifications")
	}

	h.Log.Debug(
		"Found %d modifications for trouble report %d",
		len(modifications), id)

	// Convert to the format expected by the modifications page
	var m models.Mods[models.TroubleReportModData]
	for _, mod := range modifications {
		var data models.TroubleReportModData
		if err := json.Unmarshal(mod.Modification.Data, &data); err != nil {
			continue
		}

		modEntry := models.NewMod(&mod.User, data)
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
	f := templates.CreateModificationRenderer(id, canRollback)

	// Rendering the page template
	page := base.ModPage(m, f)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c, "failed to render page: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXGetData(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.Log.Debug("User %s fetching trouble reports list", user.Name)

	trs, err := h.DB.TroubleReports.ListWithAttachments()
	if err != nil {
		return h.HandleError(c, err, "failed to load trouble reports")
	}

	h.Log.Debug("Found %d trouble reports for user %s", len(trs), user.Name)

	troubleReportsList := templates.ListReports(user, trs)
	if err := troubleReportsList.Render(c.Request().Context(), c.Response()); err != nil {
		h.HandleError(c, err, "failed to render trouble reports list component")
	}

	return nil
}

func (h *Handler) HTMXDeleteTroubleReport(c echo.Context) error {
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse trouble report ID: "+err.Error())
	}

	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	if !user.IsAdmin() {
		return h.RenderUnauthorized(c, "administrator privileges required")
	}

	h.Log.Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.Name, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReports.RemoveWithAttachments(id, user); err != nil {
		return h.HandleError(c, err, "failed to delete trouble report")
	} else {
		h.Log.Info("Successfully deleted trouble report %d (%s)",
			removedReport.ID, removedReport.Title)

		// Create feed entry
		feedTitle := "Problembericht gelöscht"
		feedContent := fmt.Sprintf("Titel: %s", removedReport.Title)
		feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.Log.Error("Failed to create feed for trouble report deletion: %v", err)
		}
	}

	return h.HTMXGetData(c)
}

func (h *Handler) HTMXGetAttachmentsPreview(c echo.Context) error {
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID from query")
	}

	h.Log.Debug("Fetching attachments preview for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return h.HandleError(c, err, "failed to load trouble report")
	}

	h.Log.Debug("Rendering attachments preview with %d attachments",
		len(tr.LoadedAttachments))

	attachmentsPreview := components.AttachmentsPreview(
		components.AttachmentPathTroubleReports,
		tr.LoadedAttachments,
	)

	err = attachmentsPreview.Render(c.Request().Context(), c.Response())
	if err != nil {
		return h.RenderInternalError(c,
			"failed to render attachments preview component: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXPostRollback(c echo.Context) error {
	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Parse ID parameter from query
	trID, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID query")
	}

	var modTime int64
	{ // Get modification timestamp from form data
		modTimeStr := c.FormValue("modification_time")
		if modTimeStr == "" {
			return h.RenderBadRequest(c,
				"modification_time form value is required")
		}

		modTime, err = strconv.ParseInt(modTimeStr, 10, 64)
		if err != nil {
			return h.RenderBadRequest(c,
				"invalid modification_time format: "+err.Error())
		}
	}

	h.Log.Info("User %s is rolling back trouble report %d to modification %d",
		user.Name, trID, modTime)

	// Find the specific modification
	modifications, err := h.DB.Modifications.ListAll(
		models.ModificationTypeTroubleReport,
		trID,
	)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve modifications")
	}

	var targetMod *models.Modification[any]
	{ // Try to get the requested mod
		for _, mod := range modifications {
			if mod.CreatedAt.UnixMilli() == modTime {
				targetMod = mod
				break
			}
		}

		if targetMod == nil {
			return h.RenderNotFound(c, "modification not found")
		}
	}

	var tr *models.TroubleReport
	{ // Rollback trouble report to the mod
		var modData models.TroubleReportModData
		if err := json.Unmarshal(targetMod.Data, &modData); err != nil {
			return h.RenderInternalError(c,
				"failed to parse modification data: "+err.Error())
		}

		// Get the current trouble report
		tr, err = h.DB.TroubleReports.Get(trID)
		if err != nil {
			return h.HandleError(c, err, "failed to retrieve trouble report")
		}

		// Update the trouble report
		err := h.DB.TroubleReports.Update(modData.CopyTo(tr), user)
		if err != nil {
			return h.HandleError(c, err, "failed to rollback trouble report")
		}
	}

	// Create feed entry
	feedTitle := "Problembericht zurückgesetzt"

	feedContent := fmt.Sprintf("Titel: %s\nZurückgesetzt auf Version vom: %s",
		tr.Title, targetMod.CreatedAt.Format("2006-01-02 15:04:05"))

	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)

	if err := h.DB.Feeds.Add(feed); err != nil {
		h.Log.Error(
			"Failed to create feed for trouble report rollback: %v",
			err,
		)
	}

	return nil
}

func (h *Handler) shareResponse(
	c echo.Context,
	tr *models.TroubleReportWithAttachments,
	buf *bytes.Buffer,
) error {
	// Check if buffer is empty or nil
	if buf == nil || buf.Len() == 0 {
		return h.RenderInternalError(c, "PDF buffer is empty")
	}

	// Sanitize the title for filename
	sanitizedTitle := h.SanitizeFilename(tr.Title)
	if sanitizedTitle == "" {
		sanitizedTitle = "fehlerbericht"
	}

	filename := fmt.Sprintf("%s_%d_%s.pdf",
		sanitizedTitle, tr.ID, time.Now().Format("2006-01-02"))

	// Set comprehensive PDF headers
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Response().Header().Set("Content-Length",
		fmt.Sprintf("%d", buf.Len()))

	// Add caching headers
	c.Response().Header().Set("Cache-Control", "private, max-age=0, no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")

	// Add security headers
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")
	c.Response().Header().Set("X-Frame-Options", "DENY")
	c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

	// Add content description
	c.Response().Header().Set("Content-Description", "Trouble Report PDF")

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}
