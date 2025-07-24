// TODO:
//   - Update this dialog to allow uploading attachments, currently with a 10mb limit per attachment
//   - Attachments order will be set per attachment ID value
//   - Before submitting, the user should be allowed to reorder attachments manually
package troublereports

import (
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type EditDialogTemplateData struct {
	ID                int                 `json:"id"`
	Submitted         bool                `json:"submitted"`
	Title             string              `json:"title"`
	Content           string              `json:"content"`
	LinkedAttachments []*pgvis.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                `json:"invalid_title"`
	InvalidContent    bool                `json:"invalid_content"`
	AttachmentError   string              `json:"attachment_error,omitempty"`
}

func (h *Handler) handleGetDialogEdit(c echo.Context, pageData *EditDialogTemplateData) *echo.HTTPError {
	if pageData == nil {
		pageData = &EditDialogTemplateData{}
	}

	if c.QueryParam(constants.QueryParamCancel) == constants.TrueValue {
		pageData.Submitted = true
	}

	if !pageData.Submitted && !pageData.InvalidTitle && !pageData.InvalidContent {
		if idStr := c.QueryParam(constants.QueryParamID); idStr != "" {
			id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
			if herr != nil {
				return herr
			}

			pageData.ID = int(id)

			tr, err := h.db.TroubleReports.Get(id)
			if err != nil {
				return utils.HandlePgvisError(c, err)
			}

			pageData.Title = tr.Title
			pageData.Content = tr.Content
			pageData.LinkedAttachments = tr.LinkedAttachments
		}
	}

	return utils.HandleTemplate(c, pageData,
		h.templates,
		[]string{
			constants.TroubleReportsDialogEditComponentTemplatePath,
		},
	)
}

func (h *Handler) handlePostDialogEdit(c echo.Context) error {
	dialogEditData := &EditDialogTemplateData{
		Submitted: true,
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, attachments, herr := h.validateDialogEditFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData.Title = title
	dialogEditData.Content = content
	dialogEditData.LinkedAttachments = attachments
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		modified := pgvis.NewModified[pgvis.TroubleReportMod](user, pgvis.TroubleReportMod{
			Title:             title,
			Content:           content,
			LinkedAttachments: attachments,
		})
		tr := pgvis.NewTroubleReport(title, content, modified)
		tr.LinkedAttachments = attachments

		if err := h.db.TroubleReports.Add(tr); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *Handler) handlePutDialogEdit(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	title, content, attachments, herr := h.validateDialogEditFormData(c)
	if herr != nil {
		return herr
	}

	dialogEditData := &EditDialogTemplateData{
		Submitted:         true,
		ID:                int(id),
		Title:             title,
		Content:           content,
		LinkedAttachments: attachments,
		InvalidTitle:      title == "",
		InvalidContent:    content == "",
	}

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		trOld, err := h.db.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}

		tr := pgvis.NewTroubleReport(title, content, trOld.Mods...)
		tr.LinkedAttachments = attachments
		tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
			Title:             tr.Title,
			Content:           tr.Content,
			LinkedAttachments: tr.LinkedAttachments,
		}))

		if err := h.db.TroubleReports.Update(id, tr); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *Handler) validateDialogEditFormData(ctx echo.Context) (title, content string, attachments []*pgvis.Attachment, httpErr *echo.HTTPError) {
	var err error

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, invalidTitleFormFieldMessage))
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, invalidContentFormFieldMessage))
	}
	content = utils.SanitizeInput(content)

	// Process existing attachments and their order
	attachments, err = h.processAttachments(ctx)
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			pgvis.WrapError(err, "failed to process attachments"))
	}

	return title, content, attachments, nil
}

// processAttachments handles file uploads and existing attachment reordering
func (h *Handler) processAttachments(ctx echo.Context) ([]*pgvis.Attachment, error) {
	var attachments []*pgvis.Attachment

	// Get existing attachments if we're editing an existing trouble report
	if idStr := ctx.QueryParam(constants.QueryParamID); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if existingTR, err := h.db.TroubleReports.Get(id); err == nil {
				// Start with existing attachments
				attachments = make([]*pgvis.Attachment, len(existingTR.LinkedAttachments))
				copy(attachments, existingTR.LinkedAttachments)
			}
		}
	}

	// Handle attachment reordering if specified
	if existingOrder := ctx.FormValue(constants.AttachmentOrderField); existingOrder != "" {
		orderParts := strings.Split(existingOrder, ",")
		reorderedAttachments := make([]*pgvis.Attachment, 0, len(attachments))

		// Reorder based on the provided order
		for _, idStr := range orderParts {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			// Find attachment with this ID in current attachments
			for _, att := range attachments {
				if att.ID == idStr {
					reorderedAttachments = append(reorderedAttachments, att)
					break
				}
			}
		}

		attachments = reorderedAttachments
	}

	// Handle new file uploads
	form, err := ctx.MultipartForm()
	if err != nil {
		// No multipart form is okay, just return existing attachments
		return attachments, nil
	}

	files := form.File[constants.AttachmentsFormField]
	for i, fileHeader := range files {
		if len(attachments) >= 10 { // Limit number of attachments
			break
		}

		// Skip empty files
		if fileHeader.Size == 0 {
			continue
		}

		attachment, err := h.processFileUpload(fileHeader, i+len(attachments))
		if err != nil {
			return nil, fmt.Errorf("failed to process file %s: %w", fileHeader.Filename, err)
		}

		if attachment != nil {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

// handlePostAttachmentReorder handles reordering of attachments
func (h *Handler) handlePostAttachmentReorder(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Get the new order from form data
	newOrder := c.FormValue("new_order")
	if newOrder == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing new_order parameter")
	}

	// Parse the new order (comma-separated attachment IDs)
	orderParts := strings.Split(newOrder, ",")

	// Get existing trouble report
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Reorder attachments based on the new order
	reorderedAttachments := make([]*pgvis.Attachment, 0, len(tr.LinkedAttachments))
	for _, attachmentID := range orderParts {
		attachmentID = strings.TrimSpace(attachmentID)
		if attachmentID == "" {
			continue
		}

		// Find the attachment with this ID
		for _, attachment := range tr.LinkedAttachments {
			if attachment.ID == attachmentID {
				reorderedAttachments = append(reorderedAttachments, attachment)
				break
			}
		}
	}

	// Update the trouble report with reordered attachments
	tr.LinkedAttachments = reorderedAttachments
	tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return updated dialog
	pageData := &EditDialogTemplateData{
		Submitted:         true, // Prevent reloading from database
		ID:                int(id),
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}

	return h.handleGetDialogEdit(c, pageData)
}

// handleDeleteAttachment handles deletion of a specific attachment
func (h *Handler) handleDeleteAttachment(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentID := c.QueryParam("attachment_id")
	if attachmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Get existing trouble report
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find and remove the attachment
	newAttachments := make([]*pgvis.Attachment, 0, len(tr.LinkedAttachments))
	found := false

	// Trim whitespace from attachment ID to avoid comparison issues
	attachmentID = strings.TrimSpace(attachmentID)

	for _, attachment := range tr.LinkedAttachments {
		trimmedAttachmentID := strings.TrimSpace(attachment.ID)
		if trimmedAttachmentID != attachmentID {
			newAttachments = append(newAttachments, attachment)
		} else {
			found = true
		}
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// Update the trouble report
	tr.LinkedAttachments = newAttachments
	tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.db.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return only the attachments section HTML
	return h.renderAttachmentsSection(c, int(id), tr.LinkedAttachments)
}

// handleGetAttachment handles downloading/viewing of attachments
func (h *Handler) handleGetAttachment(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentID := c.QueryParam("attachment_id")
	if attachmentID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	// Get trouble report
	tr, err := h.db.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find the attachment
	var attachment *pgvis.Attachment
	for _, att := range tr.LinkedAttachments {
		if att.ID == attachmentID {
			attachment = att
			break
		}
	}

	if attachment == nil {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// Set appropriate headers
	c.Response().Header().Set("Content-Type", attachment.MimeType)
	c.Response().Header().Set("Content-Length", strconv.Itoa(len(attachment.Data)))

	// Try to determine filename from attachment ID
	filename := attachment.ID
	if ext := attachment.GetFileExtension(); ext != "" {
		if !strings.HasSuffix(filename, ext) {
			filename += ext
		}
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	return c.Blob(http.StatusOK, attachment.MimeType, attachment.Data)
}

// processFileUpload processes a single uploaded file
func (h *Handler) processFileUpload(fileHeader *multipart.FileHeader, index int) (*pgvis.Attachment, error) {
	if fileHeader.Size > pgvis.MaxAttachmentDataSize {
		return nil, fmt.Errorf("file %s is too large (max 10MB)", fileHeader.Filename)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate a unique ID for the attachment using timestamp and sanitized filename
	sanitizedFilename := h.sanitizeFilename(fileHeader.Filename)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	attachmentID := fmt.Sprintf("%s_%s_%d", sanitizedFilename, timestamp, index)

	// Ensure ID doesn't exceed maximum length
	if len(attachmentID) > pgvis.MaxAttachmentIDLength {
		// Keep the timestamp and index, truncate filename part
		maxFilenameLen := pgvis.MaxAttachmentIDLength - len(timestamp) - len(fmt.Sprintf("_%d", index)) - 2
		if maxFilenameLen > 0 {
			if len(sanitizedFilename) > maxFilenameLen {
				sanitizedFilename = sanitizedFilename[:maxFilenameLen]
			}
			attachmentID = fmt.Sprintf("%s_%s_%d", sanitizedFilename, timestamp, index)
		} else {
			// Fallback to just timestamp and index
			attachmentID = fmt.Sprintf("attachment_%s_%d", timestamp, index)
		}
	}

	// Detect MIME type with fallback
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" || mimeType == "application/octet-stream" {
		// Use Go's built-in content type detection
		detectedType := http.DetectContentType(data)
		if detectedType != "application/octet-stream" {
			mimeType = detectedType
		} else {
			// Fallback to extension-based detection
			mimeType = h.getMimeTypeFromFilename(fileHeader.Filename)
		}
	}

	attachment := &pgvis.Attachment{
		ID:       attachmentID,
		MimeType: mimeType,
		Data:     data,
	}

	// Validate the attachment
	if err := attachment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid attachment: %w", err)
	}

	return attachment, nil
}

// sanitizeFilename removes or replaces invalid characters from filename
func (h *Handler) sanitizeFilename(filename string) string {
	// Remove file extension for ID generation
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}

	// Replace spaces and special characters with underscores
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "-", "_")
	filename = strings.ReplaceAll(filename, "(", "_")
	filename = strings.ReplaceAll(filename, ")", "_")
	filename = strings.ReplaceAll(filename, "[", "_")
	filename = strings.ReplaceAll(filename, "]", "_")

	// Remove consecutive underscores
	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	// Trim underscores from start and end
	filename = strings.Trim(filename, "_")

	// Ensure we have a valid filename
	if filename == "" {
		filename = "attachment"
	}

	return filename
}

// getMimeTypeFromFilename tries to determine MIME type from file extension
func (h *Handler) getMimeTypeFromFilename(filename string) string {
	ext := strings.ToLower(filename)
	if idx := strings.LastIndex(ext, "."); idx >= 0 {
		ext = ext[idx:]

		// Common MIME types
		switch ext {
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".png":
			return "image/png"
		case ".gif":
			return "image/gif"
		case ".pdf":
			return "application/pdf"
		case ".txt":
			return "text/plain"
		case ".doc":
			return "application/msword"
		case ".docx":
			return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case ".zip":
			return "application/zip"
		case ".svg":
			return "image/svg+xml"
		case ".webp":
			return "image/webp"
		case ".bmp":
			return "image/bmp"
		case ".rtf":
			return "application/rtf"
		case ".odt":
			return "application/vnd.oasis.opendocument.text"
		case ".rar":
			return "application/vnd.rar"
		case ".7z":
			return "application/x-7z-compressed"
		case ".tar":
			return "application/x-tar"
		case ".gz":
			return "application/gzip"
		case ".bz2":
			return "application/x-bzip2"
		}
	}

	// Default fallback
	return "application/octet-stream"
}

// renderAttachmentsSection renders only the attachments section HTML
func (h *Handler) renderAttachmentsSection(c echo.Context, reportID int, attachments []*pgvis.Attachment) error {
	data := struct {
		ID                int                 `json:"id"`
		LinkedAttachments []*pgvis.Attachment `json:"linked_attachments"`
	}{
		ID:                reportID,
		LinkedAttachments: attachments,
	}

	// Return just the attachments section content
	attachmentsSectionHTML := `
		<div class="attachments-label">Anhänge (max. 10MB pro Datei, max. 10 Dateien)</div>

		{{if .LinkedAttachments}}
		<div class="attachments-section">
			<div class="attachments-label">Vorhandene Anhänge:</div>
			<div id="existing-attachments">
				{{range .LinkedAttachments}}
				<div class="attachment-item flex row gap" data-id="{{.ID}}">
					<div class="attachment-info">
						<i class="bi bi-grip-vertical"></i>
						{{if .IsImage}}
						<i class="bi bi-image attachment-icon"></i>
						{{else if .IsDocument}}
						<i class="bi bi-file-text attachment-icon"></i>
						{{else if .IsArchive}}
						<i class="bi bi-file-zip attachment-icon"></i>
						{{else}}
						<i class="bi bi-file-earmark attachment-icon"></i>
						{{end}}
						<span class="ellipsis">{{.ID}}</span>
						<span class="text-muted">({{.GetMimeType}})</span>
					</div>
					<div class="attachment-actions">
						<button type="button" class="secondary small" onclick="viewAttachment({{$.ID}}, '{{.ID}}')">
							<i class="bi bi-eye"></i> Anzeigen
						</button>
						<button type="button" class="destructive small" onclick="deleteAttachment({{$.ID}}, '{{.ID}}')">
							<i class="bi bi-trash"></i> Löschen
						</button>
					</div>
				</div>
				{{end}}
			</div>
		</div>
		{{end}}

		<!-- File Upload Area -->
		<div class="file-input-area"
			 onclick="document.getElementById('attachments').click()"
			 ondrop="handleFileDrop(event)"
			 ondragover="handleDragOver(event)"
			 ondragleave="handleDragLeave(event)">
			<i class="bi bi-cloud-upload" style="font-size: 2em; margin-bottom: 8px;"></i>
			<div>Klicken Sie hier oder ziehen Sie Dateien hierher</div>
			<div class="text-muted">Unterstützte Formate: Bilder, PDFs, Dokumente, Archive</div>
			<input type="file"
				   name="attachments"
				   id="attachments"
				   multiple
				   accept="image/*,.pdf,.doc,.docx,.txt,.rtf,.odt,.zip,.rar,.7z,.tar,.gz,.bz2"
				   onchange="handleFileSelect(event)">
		</div>

		<!-- File Preview Area -->
		<div id="file-preview" class="file-preview">
			<div class="attachments-label">Neue Dateien:</div>
			<div id="new-attachments"></div>
		</div>
	`

	tmpl, err := template.New("attachments").Parse(attachmentsSectionHTML)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse attachments template")
	}

	return tmpl.Execute(c.Response(), data)
}
