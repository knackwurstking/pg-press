package troublereports

import (
	"fmt"
	"io"
	"io/fs"
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

type DialogEditTemplateData struct {
	ID                int                 `json:"id"`
	Submitted         bool                `json:"submitted"`
	Title             string              `json:"title"`
	Content           string              `json:"content"`
	LinkedAttachments []*pgvis.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                `json:"invalid_title"`
	InvalidContent    bool                `json:"invalid_content"`
	AttachmentError   string              `json:"attachment_error,omitempty"`
}

type DialogEditHandler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func (h *DialogEditHandler) RegisterRoutes(e *echo.Echo) {
	editDialogPath := h.serverPathPrefix + "/trouble-reports/dialog-edit"

	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})

	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)
}

func (h *DialogEditHandler) handleGetDialogEdit(c echo.Context, pageData *DialogEditTemplateData) *echo.HTTPError {
	if pageData == nil {
		pageData = &DialogEditTemplateData{}
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

			// Load attachments for display
			if loadedAttachments, err := h.db.TroubleReportService.LoadAttachments(tr); err == nil {
				pageData.LinkedAttachments = loadedAttachments
			}
		}
	}

	return utils.HandleTemplate(c, pageData,
		h.templates,
		[]string{
			constants.TroubleReportsDialogEditComponentTemplatePath,
		},
	)
}

func (h *DialogEditHandler) handlePostDialogEdit(c echo.Context) error {
	dialogEditData := &DialogEditTemplateData{
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
			LinkedAttachments: []int64{}, // Will be set by the service
		})
		tr := pgvis.NewTroubleReport(title, content, modified)

		if err := h.db.TroubleReportService.AddWithAttachments(tr, attachments); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *DialogEditHandler) handlePutDialogEdit(c echo.Context) error {
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

	dialogEditData := &DialogEditTemplateData{
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

		// Convert existing attachments to IDs for reordering
		var existingAttachmentIDs []int64
		for _, att := range attachments {
			if att.GetID() > 0 { // Only include existing attachments (with IDs)
				existingAttachmentIDs = append(existingAttachmentIDs, att.GetID())
			}
		}

		// Filter out new attachments (those without valid IDs)
		var newAttachments []*pgvis.Attachment
		for _, att := range attachments {
			if att.GetID() == 0 { // New attachments don't have IDs yet
				newAttachments = append(newAttachments, att)
			}
		}

		tr.LinkedAttachments = existingAttachmentIDs
		tr.Mods = append(tr.Mods, pgvis.NewModified(user, pgvis.TroubleReportMod{
			Title:             tr.Title,
			Content:           tr.Content,
			LinkedAttachments: []int64{}, // Will be set by the service
		}))

		if err := h.db.TroubleReportService.UpdateWithAttachments(id, tr, newAttachments); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *DialogEditHandler) validateDialogEditFormData(ctx echo.Context) (title, content string, attachments []*pgvis.Attachment, httpErr *echo.HTTPError) {
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
func (h *DialogEditHandler) processAttachments(ctx echo.Context) ([]*pgvis.Attachment, error) {
	var attachments []*pgvis.Attachment

	// Get existing attachments if we're editing an existing trouble report
	if idStr := ctx.QueryParam(constants.QueryParamID); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if existingTR, err := h.db.TroubleReports.Get(id); err == nil {
				// Load existing attachments using the service
				if loadedAttachments, err := h.db.TroubleReportService.LoadAttachments(existingTR); err == nil {
					attachments = make([]*pgvis.Attachment, len(loadedAttachments))
					copy(attachments, loadedAttachments)
				}
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

			// Parse the attachment ID
			attachmentID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}

			// Find attachment with this ID in current attachments
			for _, att := range attachments {
				if att.GetID() == attachmentID {
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

// processFileUpload processes a single uploaded file
func (h *DialogEditHandler) processFileUpload(fileHeader *multipart.FileHeader, index int) (*pgvis.Attachment, error) {
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

	// For new attachments, use a temporary ID (will be replaced with database ID)
	// We'll use a combination of filename and timestamp for identification during upload
	sanitizedFilename := h.sanitizeFilename(fileHeader.Filename)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	attachmentID := fmt.Sprintf("temp_%s_%s_%d", sanitizedFilename, timestamp, index)

	// Ensure ID doesn't exceed maximum length
	if len(attachmentID) > pgvis.MaxAttachmentIDLength {
		// Keep the timestamp and index, truncate filename part
		maxFilenameLen := pgvis.MaxAttachmentIDLength - len(timestamp) - len(fmt.Sprintf("temp_%d", index)) - 2
		if maxFilenameLen > 0 {
			if len(sanitizedFilename) > maxFilenameLen {
				sanitizedFilename = sanitizedFilename[:maxFilenameLen]
			}
			attachmentID = fmt.Sprintf("temp_%s_%s_%d", sanitizedFilename, timestamp, index)
		} else {
			// Fallback to just timestamp and index
			attachmentID = fmt.Sprintf("temp_%s_%d", timestamp, index)
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
func (h *DialogEditHandler) sanitizeFilename(filename string) string {
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
func (h *DialogEditHandler) getMimeTypeFromFilename(filename string) string {
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
