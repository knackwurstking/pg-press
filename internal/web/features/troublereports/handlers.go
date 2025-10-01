package troublereports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/pdf"
	"github.com/knackwurstking/pgpress/internal/services"
	"github.com/knackwurstking/pgpress/internal/web/features/troublereports/templates"
	"github.com/knackwurstking/pgpress/internal/web/shared/base"
	"github.com/knackwurstking/pgpress/internal/web/shared/components"
	"github.com/knackwurstking/pgpress/internal/web/shared/handlers"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/modification"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	*handlers.BaseHandler
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{
		BaseHandler: handlers.NewBaseHandler(db, logger.NewComponentLogger("Trouble Reports")),
	}
}

func (h *Handler) GetPage(c echo.Context) error {
	h.LogDebug("Rendering trouble reports page")

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

func (h *Handler) GetAttachment(c echo.Context) error {
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

func (h *Handler) GetModificationsForID(c echo.Context) error {
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

	h.LogDebug("User %s fetching trouble reports list", user.Name)

	trs, err := h.DB.TroubleReports.ListWithAttachments()
	if err != nil {
		return h.HandleError(c, err, "failed to load trouble reports")
	}

	h.LogDebug("Found %d trouble reports for user %s", len(trs), user.Name)

	troubleReportsList := templates.TroubleReportsList(user, trs)
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

	h.LogInfo("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.Name, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReports.RemoveWithAttachments(id, user); err != nil {
		return h.HandleError(c, err, "failed to delete trouble report")
	} else {
		h.LogInfo("Successfully deleted trouble report %d (%s)",
			removedReport.ID, removedReport.Title)

		// Create feed entry
		feedTitle := "Problembericht gelöscht"
		feedContent := fmt.Sprintf("Titel: %s", removedReport.Title)
		feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
		if err := h.DB.Feeds.Add(feed); err != nil {
			h.LogError("Failed to create feed for trouble report deletion: %v", err)
		}
	}

	return h.HTMXGetData(c)
}

func (h *Handler) HTMXGetAttachmentsPreview(c echo.Context) error {
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID from query")
	}

	h.LogDebug("Fetching attachments preview for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return h.HandleError(c, err, "failed to load trouble report")
	}

	h.LogDebug("Rendering attachments preview with %d attachments",
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

func (h *Handler) HTMXGetEditTroubleReportDialog(c echo.Context) error {
	props := &templates.DialogEditTroubleReportProps{}
	props.ID, _ = h.ParseInt64Query(c, "id")

	if props.ID > 0 {
		h.LogDebug("Open edit dialog for trouble report %d", props.ID)
	} else {
		h.LogDebug("Open edit dialog for new trouble report")
	}

	if props.ID > 0 {
		tr, err := h.DB.TroubleReports.Get(props.ID)
		if err != nil {
			return h.HandleError(c, err, "failed to get trouble report")
		}
		props.Title = tr.Title
		props.Content = tr.Content

		// Load attachments for display
		loadedAttachments, err := h.DB.TroubleReports.LoadAttachments(tr)
		if err == nil {
			props.Attachments = loadedAttachments
		} else {
			h.LogError("Failed to load attachments for trouble report %d: %v",
				props.ID, err)
		}
	}

	dialog := templates.DialogEditTroubleReport(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render trouble report edit dialog: "+err.Error())
	}

	return nil
}

func (h *Handler) HTMXPostEditTroubleReportDialog(c echo.Context) error {
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	h.LogDebug("User %s is creating a new trouble report", user.Name)

	title, content, attachments, err := h.validateDialogEditFormData(c)
	if err != nil {
		return h.RenderBadRequest(c,
			"failed to get trouble report form data: "+err.Error())
	}
	if title == "" || content == "" {
		return h.RenderBadRequest(c, "title and content are required")
	}

	tr := models.NewTroubleReport(title, content)

	h.LogDebug("Creating trouble report: title='%s', attachments=%d",
		title, len(attachments))

	err = h.DB.TroubleReports.AddWithAttachments(tr, user, attachments...)
	if err != nil {
		return h.HandleError(c, err, "failed to add trouble report")
	}

	// Create feed entry
	feedTitle := "Neuer Problembericht erstellt"
	feedContent := fmt.Sprintf("Titel: %s", tr.Title)
	if len(attachments) > 0 {
		feedContent += fmt.Sprintf("\nAnhänge: %d", len(attachments))
	}
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for trouble report creation: %v", err)
	}

	return h.closeDialog(c)
}

func (h *Handler) HTMXPutEditTroubleReportDialog(c echo.Context) error {
	// Get ID from query parameter
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID query")
	}

	h.LogDebug("Updating trouble report %d", id)

	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	// Get Title, Content and Attachments from form data
	title, content, attachments, err := h.validateDialogEditFormData(c)
	if err != nil {
		return h.RenderBadRequest(c,
			"failed to get trouble report form data: "+err.Error())
	}
	if title == "" || content == "" {
		return h.RenderBadRequest(c, "title and content are required")
	}

	// Query previous trouble report
	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return h.HandleError(c, err, "failed to get trouble report")
	}

	// Filter out existing and new attachments
	var existingAttachmentIDs []int64
	var newAttachments []*models.Attachment
	for _, a := range attachments {
		if a.GetID() > 0 {
			existingAttachmentIDs = append(existingAttachmentIDs, a.GetID())
		} else {
			newAttachments = append(newAttachments, a)
		}
	}

	// Update trouble report with existing and new attachments, title content and mods
	h.LogDebug("Updating trouble report %d with %d attachments",
		id, len(attachments))

	// Update the previous trouble report
	tr.Title = title
	tr.Content = content
	tr.LinkedAttachments = existingAttachmentIDs

	err = h.DB.TroubleReports.UpdateWithAttachments(id, tr, user, newAttachments...)
	if err != nil {
		return h.HandleError(c, err, "failed to update trouble report")
	}

	// Create feed entry
	feedTitle := "Problembericht aktualisiert"
	feedContent := fmt.Sprintf("Titel: %s", tr.Title)
	totalAttachments := len(existingAttachmentIDs) + len(newAttachments)
	if totalAttachments > 0 {
		feedContent += fmt.Sprintf("\nAnhänge: %d", totalAttachments)
	}
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for trouble report update: %v", err)
	}

	return h.closeDialog(c)
}

func (h *Handler) HTMXPostRollback(c echo.Context) error {
	h.LogInfo("Handling HTMX rollback for trouble report")

	// Parse ID parameter from query
	id, err := h.ParseInt64Query(c, "id")
	if err != nil {
		return h.RenderBadRequest(c, "failed to parse ID query")
	}

	// Get modification timestamp from form data
	modTimeStr := c.FormValue("modification_time")
	if modTimeStr == "" {
		return h.RenderBadRequest(c, "modification_time form value is required")
	}

	modTime, err := strconv.ParseInt(modTimeStr, 10, 64)
	if err != nil {
		return h.RenderBadRequest(c, "invalid modification_time format: "+err.Error())
	}

	// Get user from context
	user, err := h.GetUserFromContext(c)
	if err != nil {
		return h.HandleError(c, err, "failed to get user from context")
	}

	if !user.IsAdmin() {
		return h.RenderUnauthorized(c, "administrator privileges required")
	}

	h.LogInfo("User %s is rolling back trouble report %d to modification %d",
		user.Name, id, modTime)

	// Find the specific modification
	modifications, err := h.DB.Modifications.ListAll(
		services.ModificationTypeTroubleReport, id)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve modifications")
	}

	var targetMod *models.Modification[any]
	for _, mod := range modifications {
		if mod.CreatedAt.UnixMilli() == modTime {
			targetMod = mod
			break
		}
	}

	if targetMod == nil {
		return h.RenderNotFound(c, "modification not found")
	}

	// Unmarshal the modification data
	var modData models.TroubleReportModData
	if err := json.Unmarshal(targetMod.Data, &modData); err != nil {
		return h.RenderInternalError(c,
			"failed to parse modification data: "+err.Error())
	}

	// Get the current trouble report
	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return h.HandleError(c, err, "failed to retrieve trouble report")
	}

	// Apply the rollback
	tr.Title = modData.Title
	tr.Content = modData.Content
	tr.LinkedAttachments = modData.LinkedAttachments

	// Update the trouble report
	if err := h.DB.TroubleReports.Update(tr, user); err != nil {
		return h.HandleError(c, err, "failed to rollback trouble report")
	}

	h.LogInfo("Successfully rolled back trouble report %d", id)

	// Create feed entry
	feedTitle := "Problembericht zurückgesetzt"
	feedContent := fmt.Sprintf("Titel: %s\nZurückgesetzt auf Version vom: %s",
		tr.Title, targetMod.CreatedAt.Format("2006-01-02 15:04:05"))
	feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
	if err := h.DB.Feeds.Add(feed); err != nil {
		h.LogError("Failed to create feed for trouble report rollback: %v", err)
	}

	// Return success message for HTMX
	err = templates.RollbackResponseStatusOK().Render(
		c.Request().Context(), c.Response(),
	)
	if err != nil {
		return h.HandleError(c, err, "failed to render response")
	}

	return nil
}

func (h *Handler) validateDialogEditFormData(ctx echo.Context) (
	title, content string,
	attachments []*models.Attachment,
	err error,
) {
	title = h.GetSanitizedFormValue(ctx, constants.TitleFormField)
	if title == "" {
		return "", "", nil, fmt.Errorf("missing title")
	}

	content = h.GetSanitizedFormValue(ctx, constants.ContentFormField)
	if content == "" {
		return "", "", nil, fmt.Errorf("missing content")
	}

	// Process existing attachments and their order
	attachments, err = h.processAttachments(ctx)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to process attachments: %v", err)
	}

	return title, content, attachments, nil
}

func (h *Handler) processAttachments(ctx echo.Context) ([]*models.Attachment, error) {
	var attachments []*models.Attachment

	// Get existing attachments if editing
	if idStr := ctx.QueryParam("id"); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if existingTR, err := h.DB.TroubleReports.Get(id); err == nil {
				if loadedAttachments, err := h.DB.TroubleReports.LoadAttachments(
					existingTR); err == nil {
					attachments = make([]*models.Attachment, len(loadedAttachments))
					copy(attachments, loadedAttachments)
				}
			}
		}
	}

	// Handle new file uploads
	form, err := ctx.MultipartForm()
	if err != nil {
		return attachments, nil
	}

	attachmentsToRemoveSeq := strings.SplitSeq(
		form.Value[constants.ExistingAttachmentsRemovalFormField][0],
		",")
	for atr := range attachmentsToRemoveSeq {
		for i, a := range attachments {
			if a.ID == atr {
				attachments = slices.Delete(attachments, i, 1)
				break
			}
		}
	}

	files := form.File[constants.AttachmentsFormField]
	for i, fileHeader := range files {
		if len(attachments) >= 10 {
			break
		}

		if fileHeader.Size == 0 {
			continue
		}

		attachment, err := h.processFileUpload(fileHeader, i+len(attachments))
		if err != nil {
			return nil, fmt.Errorf("failed to process file %s: %v", fileHeader.Filename, err)
		}

		if attachment != nil {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

func (h *Handler) processFileUpload(
	fileHeader *multipart.FileHeader, index int,
) (*models.Attachment, error) {
	if fileHeader.Size > models.MaxDataSize {
		return nil, fmt.Errorf("file %s is too large (max 10MB)",
			fileHeader.Filename)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Create temporary ID
	sanitizedFilename := h.SanitizeFilename(fileHeader.Filename)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	attachmentID := fmt.Sprintf("temp_%s_%s_%d",
		sanitizedFilename, timestamp, index)

	// Ensure ID doesn't exceed maximum length
	if len(attachmentID) > models.MaxIDLength {
		maxFilenameLen := models.MaxIDLength - len(timestamp) -
			len(fmt.Sprintf("temp_%d", index)) - 2

		if maxFilenameLen > 0 && len(sanitizedFilename) > maxFilenameLen {
			sanitizedFilename = sanitizedFilename[:maxFilenameLen]
			attachmentID = fmt.Sprintf("temp_%s_%s_%d",
				sanitizedFilename, timestamp, index)
		} else {
			attachmentID = fmt.Sprintf("temp_%s_%d", timestamp, index)
		}
	}

	// Detect MIME type
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		detectedType := http.DetectContentType(data)
		if detectedType != "application/octet-stream" {
			mimeType = detectedType
		} else {
			mimeType = h.getMimeTypeFromFilename(fileHeader.Filename)
		}
	}

	// Validate that the file is an image
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, fmt.Errorf("only image files are allowed (JPG, PNG, GIF, BMP, SVG, WebP)")
	}

	attachment := &models.Attachment{
		ID:       attachmentID,
		MimeType: mimeType,
		Data:     data,
	}

	if err := attachment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid attachment: %v", err)
	}

	return attachment, nil
}

func (h *Handler) getMimeTypeFromFilename(filename string) string {
	ext := strings.ToLower(filename)
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]

		switch ext {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".svg":
			return "image/svg+xml"
		case ".webp":
			return "image/webp"
		case ".bmp":
			return "image/bmp"
		}
	}

	return "application/octet-stream"
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

func (h *Handler) closeDialog(c echo.Context) error {
	dialog := templates.DialogEditTroubleReport(
		&templates.DialogEditTroubleReportProps{})

	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return h.RenderInternalError(c,
			"failed to render trouble report edit dialog: "+err.Error())
	}

	return nil
}
