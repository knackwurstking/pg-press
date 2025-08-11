package htmxhandler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/templates/components"
	"github.com/knackwurstking/pgpress/internal/utils"
)

func (h *TroubleReports) processAttachments(ctx echo.Context) ([]*database.Attachment, error) {
	var attachments []*database.Attachment

	// Get existing attachments if editing
	if idStr := ctx.QueryParam(constants.QueryParamID); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if existingTR, err := h.DB.TroubleReports.Get(id); err == nil {
				if loadedAttachments, err := h.DB.TroubleReportsHelper.LoadAttachments(
					existingTR); err == nil {
					attachments = make([]*database.Attachment, len(loadedAttachments))
					copy(attachments, loadedAttachments)
				}
			}
		}
	}

	// Handle attachment reordering
	if existingOrder := ctx.FormValue(constants.AttachmentOrderField); existingOrder != "" {
		orderParts := strings.Split(existingOrder, ",")
		reorderedAttachments := make([]*database.Attachment, 0, len(attachments))

		for _, idStr := range orderParts {
			idStr = strings.TrimSpace(idStr)
			if idStr == "" {
				continue
			}

			attachmentID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				continue
			}

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
		return attachments, nil
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
			return nil, fmt.Errorf("failed to process file %s: %w", fileHeader.Filename, err)
		}

		if attachment != nil {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

func (h *TroubleReports) processFileUpload(
	fileHeader *multipart.FileHeader,
	index int,
) (*database.Attachment, error) {
	if fileHeader.Size > database.MaxAttachmentDataSize {
		return nil, fmt.Errorf("file %s is too large (max 10MB)",
			fileHeader.Filename)
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

	// Create temporary ID
	sanitizedFilename := h.sanitizeFilename(fileHeader.Filename)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	attachmentID := fmt.Sprintf("temp_%s_%s_%d", sanitizedFilename, timestamp, index)

	// Ensure ID doesn't exceed maximum length
	if len(attachmentID) > database.MaxAttachmentIDLength {
		maxFilenameLen := database.MaxAttachmentIDLength - len(timestamp) -
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

	attachment := &database.Attachment{
		ID:       attachmentID,
		MimeType: mimeType,
		Data:     data,
	}

	if err := attachment.Validate(); err != nil {
		return nil, fmt.Errorf("invalid attachment: %w", err)
	}

	return attachment, nil
}

func (h *TroubleReports) sanitizeFilename(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx > 0 {
		filename = filename[:idx]
	}

	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "-", "_")
	filename = strings.ReplaceAll(filename, "(", "_")
	filename = strings.ReplaceAll(filename, ")", "_")
	filename = strings.ReplaceAll(filename, "[", "_")
	filename = strings.ReplaceAll(filename, "]", "_")

	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	filename = strings.Trim(filename, "_")

	if filename == "" {
		filename = "attachment"
	}

	return filename
}

func (h *TroubleReports) getMimeTypeFromFilename(filename string) string {
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

func (h *TroubleReports) handleGetAttachment(c echo.Context) error {
	attachmentID, herr := utils.ParseInt64Query(c, "attachment_id")
	if herr != nil {
		return herr
	}

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		return utils.HandlepgpressError(c, err)
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

func (h *TroubleReports) handlePostAttachmentReorder(c echo.Context) error {
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
	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlepgpressError(c, err)
	}

	// Reorder attachment IDs based on the new order
	reorderedAttachmentIDs := make([]int64, 0, len(tr.LinkedAttachments))
	for _, attachmentIDStr := range orderParts {
		attachmentIDStr = strings.TrimSpace(attachmentIDStr)
		if attachmentIDStr == "" {
			continue
		}

		attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
		if err != nil {
			continue // Skip invalid IDs
		}

		// Check if this ID exists in the current attachments
		if slices.Contains(tr.LinkedAttachments, attachmentID) {
			reorderedAttachmentIDs = append(reorderedAttachmentIDs, attachmentID)
		}
	}

	// Update the trouble report with reordered attachment IDs
	tr.LinkedAttachments = reorderedAttachmentIDs
	tr.Mods = append(tr.Mods, database.NewModified(user, database.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.DB.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlepgpressError(c, err)
	}

	// Load attachments for dialog
	attachments, err := h.DB.TroubleReportsHelper.LoadAttachments(tr)
	if err != nil {
		return utils.HandlepgpressError(c, err)
	}

	return h.handleGetDialogEdit(c, &components.TroubleReportsEditDialogProps{
		Submitted:   true,
		ID:          id,
		Title:       tr.Title,
		Content:     tr.Content,
		Attachments: attachments,
	})
}
