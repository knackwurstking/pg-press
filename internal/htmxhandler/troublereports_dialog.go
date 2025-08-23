package htmxhandler

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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

func (h *TroubleReports) handleGetDialogEdit(
	c echo.Context,
	props *components.TroubleReportsEditDialogProps,
) error {
	if props == nil {
		props = &components.TroubleReportsEditDialogProps{}
	}

	if !props.Close {
		props.Close = utils.ParseBoolQuery(c, constants.QueryParamClose)
	}

	if !props.Close && !props.InvalidTitle && !props.InvalidContent {
		if idStr := c.QueryParam(constants.QueryParamID); idStr != "" {
			id, err := utils.ParseInt64Query(c, constants.QueryParamID)
			if err != nil {
				return err
			}

			props.ID = id

			tr, err := h.DB.TroubleReports.Get(id)
			if err != nil {
				return echo.NewHTTPError(database.GetHTTPStatusCode(err),
					"failed to get trouble report: "+err.Error())
			}

			props.Title = tr.Title
			props.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReportsHelper.LoadAttachments(tr); err == nil {
				props.Attachments = loadedAttachments
			}
		}
	}

	dialog := components.TroubleReportsEditDialog(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render Trouble Reports Edit Dialog: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handlePostDialogEdit(c echo.Context) error {
	props := &components.TroubleReportsEditDialogProps{
		Close: true,
	}

	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	title, content, attachments, err := h.validateDialogEditFormData(c)
	if err != nil {
		return err
	}

	props.Title = title
	props.Content = content
	props.InvalidTitle = title == ""
	props.InvalidContent = content == ""

	if !props.InvalidTitle && !props.InvalidContent {
		props.Attachments = attachments

		err := h.DB.TroubleReportsHelper.AddWithAttachments(
			database.NewTroubleReport(title, content, user),
			attachments,
		)
		if err != nil {
			return echo.NewHTTPError(database.GetHTTPStatusCode(err),
				"failed to add trouble report: "+err.Error())
		}
	} else {
		props.Close = false
	}

	return h.handleGetDialogEdit(c, props)
}

func (h *TroubleReports) handlePutDialogEdit(c echo.Context) error {
	// Get ID from query parameter
	id, err := utils.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	// Get user from context
	user, err := utils.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Get Title, Content and Attachments from form data
	title, content, attachments, err := h.validateDialogEditFormData(c)
	if err != nil {
		return err
	}

	// Initialize dialog template data
	props := &components.TroubleReportsEditDialogProps{
		Close:          true,
		ID:             id,
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	// Abort if invalid title or content
	if props.InvalidTitle || props.InvalidContent {
		props.Close = false
		return h.handleGetDialogEdit(c, props)
	}

	// Set attachments to handlePutDialogEdit
	props.Attachments = attachments

	// Query previous trouble report
	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to get trouble report: "+err.Error())
	}

	// Create new trouble report
	//tr := database.NewTroubleReportWithMods(title, content, trOld.Mods...)

	// Filter out existing and new attachments
	var existingAttachmentIDs []int64
	var newAttachments []*database.Attachment
	for _, a := range props.Attachments {
		if a.GetID() > 0 {
			existingAttachmentIDs = append(existingAttachmentIDs, a.GetID())
		} else {
			newAttachments = append(newAttachments, a)
		}
	}

	// Update trouble report with existing and new attachments, title content and mods
	tr.Update(user, title, content, existingAttachmentIDs...)

	if err := h.DB.TroubleReportsHelper.UpdateWithAttachments(id, tr, newAttachments); err != nil {
		return echo.NewHTTPError(database.GetHTTPStatusCode(err),
			"failed to update trouble report: "+err.Error())
	}

	return h.handleGetDialogEdit(c, props)
}

func (h *TroubleReports) validateDialogEditFormData(ctx echo.Context) (
	title, content string,
	attachments []*database.Attachment,
	err error,
) {
	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, "invalid title form value"))
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, "invalid content form value"))
	}
	content = utils.SanitizeInput(content)

	// Process existing attachments and their order
	attachments, err = h.processAttachments(ctx)
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, "failed to process attachments"))
	}

	return title, content, attachments, nil
}

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

	// Handle new file uploads
	form, err := ctx.MultipartForm()
	if err != nil {
		return attachments, nil
	}

	existingAttachmentsToRemove := strings.SplitSeq(
		form.Value[constants.ExistingAttachmentsRemoval][0],
		",",
	)
	for a := range existingAttachmentsToRemove {
		for i, a2 := range attachments {
			if a2.ID == a {
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
