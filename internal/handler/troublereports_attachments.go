package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

func (h *TroubleReports) processAttachments(ctx echo.Context) ([]*database.Attachment, error) {
	var attachments []*database.Attachment

	// Get existing attachments if editing
	if idStr := ctx.QueryParam(constants.QueryParamID); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if existingTR, err := h.DB.TroubleReports.Get(id); err == nil {
				if loadedAttachments, err := h.DB.TroubleReportService.LoadAttachments(
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
		return nil, fmt.Errorf(nonImageFileMessage)
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
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentIDStr := c.QueryParam("attachment_id")
	if attachmentIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment_id parameter")
	}

	// Get trouble report to verify attachment belongs to it
	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Check if attachment ID is in the trouble report's linked attachments
	var found bool
	for _, linkedID := range tr.LinkedAttachments {
		if linkedID == attachmentID {
			found = true
			break
		}
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found in this trouble report")
	}

	// Get the attachment from the attachments table
	attachment, err := h.DB.Attachments.Get(attachmentID)
	if err != nil {
		return utils.HandlePgvisError(c, err)
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
		return utils.HandlePgvisError(c, err)
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
		for _, existingID := range tr.LinkedAttachments {
			if existingID == attachmentID {
				reorderedAttachmentIDs = append(reorderedAttachmentIDs, attachmentID)
				break
			}
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
		return utils.HandlePgvisError(c, err)
	}

	// Load attachments for dialog
	attachments, err := h.DB.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Return updated dialog
	pageData := &DialogEditTemplateData{
		Submitted:         true, // Prevent reloading from database
		ID:                int(id),
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: attachments,
	}

	return h.handleGetDialogEdit(c, pageData)
}

func (h *TroubleReports) handleDeleteAttachment(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	attachmentIDStr := c.QueryParam("attachment_id")
	if attachmentIDStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing attachment_id parameter")
	}

	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid attachment_id parameter")
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	// Get existing trouble report
	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find and remove the attachment ID
	newAttachmentIDs := make([]int64, 0, len(tr.LinkedAttachments))
	found := false

	for _, linkedID := range tr.LinkedAttachments {
		if linkedID != attachmentID {
			newAttachmentIDs = append(newAttachmentIDs, linkedID)
		} else {
			found = true
		}
	}

	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "attachment not found")
	}

	// Update the trouble report
	tr.LinkedAttachments = newAttachmentIDs
	tr.Mods = append(tr.Mods, database.NewModified(user, database.TroubleReportMod{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}))

	if err := h.DB.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Load attachments and return only the attachments section HTML
	attachments, err := h.DB.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.renderAttachmentsSection(c, int(id), attachments)
}

// renderAttachmentsSection renders only the attachments section HTML
func (h *TroubleReports) renderAttachmentsSection(
	c echo.Context,
	reportID int,
	attachments []*database.Attachment,
) error {
	data := struct {
		ID                int                    `json:"id"`
		LinkedAttachments []*database.Attachment `json:"linked_attachments"`
	}{
		ID:                reportID,
		LinkedAttachments: attachments,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{constants.AttachmentsSectionComponentTemplatePath},
	)
}
