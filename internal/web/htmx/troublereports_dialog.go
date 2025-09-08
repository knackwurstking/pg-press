package htmx

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

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	attachmentmodels "github.com/knackwurstking/pgpress/internal/database/models/attachment"
	trmodels "github.com/knackwurstking/pgpress/internal/database/models/troublereport"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/constants"
	webhelpers "github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/components/dialogs"

	"github.com/labstack/echo/v4"
)

func (h *TroubleReports) handleGetDialogEdit(
	c echo.Context,
	props *dialogs.EditTroubleReportProps,
) error {
	if props == nil {
		props = &dialogs.EditTroubleReportProps{}
	}

	if !props.Close {
		props.Close = webhelpers.ParseBoolQuery(c, constants.QueryParamClose)
	}

	if !props.Close && !props.InvalidTitle && !props.InvalidContent {
		if idStr := c.QueryParam(constants.QueryParamID); idStr != "" {
			id, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
			if err != nil {
				return err
			}

			props.ID = id

			logger.HTMXHandlerTroubleReports().Debug("Loading trouble report %d for editing", id)
			tr, err := h.DB.TroubleReports.Get(id)
			if err != nil {
				logger.HTMXHandlerTroubleReports().Error("Failed to get trouble report %d: %v", id, err)
				return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
					"failed to get trouble report: "+err.Error())
			}

			props.Title = tr.Title
			props.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReports.LoadAttachments(tr); err == nil {
				props.Attachments = loadedAttachments
				logger.HTMXHandlerTroubleReports().Debug("Loaded %d attachments for trouble report %d", len(loadedAttachments), id)
			} else {
				logger.HTMXHandlerTroubleReports().Error("Failed to load attachments for trouble report %d: %v", id, err)
			}
		}
	}

	dialog := dialogs.EditTroubleReport(props)
	if err := dialog.Render(c.Request().Context(), c.Response()); err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to render edit dialog: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render Trouble Reports Edit Dialog: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handlePostDialogEdit(c echo.Context) error {
	props := &dialogs.EditTroubleReportProps{
		Close: true,
	}

	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Form validation failed: %v", err)
		return err
	}

	logger.HTMXHandlerTroubleReports().Info("User %s is creating a new trouble report", user.Name)

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
		tr := trmodels.New(title, content)

		logger.HTMXHandlerTroubleReports().Debug(
			"Creating trouble report: title='%s', attachments=%d",
			title, len(attachments),
		)

		err := h.DB.TroubleReports.AddWithAttachments(user, tr, attachments)
		if err != nil {
			logger.HTMXHandlerTroubleReports().Error(
				"Failed to add trouble report: %v",
				err,
			)

			return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
				"failed to add trouble report: "+err.Error())
		}

		logger.HTMXHandlerTroubleReports().Info(
			"Successfully created trouble report %d",
			tr.ID,
		)
	} else {
		props.Close = false
	}

	return h.handleGetDialogEdit(c, props)
}

func (h *TroubleReports) handlePutDialogEdit(c echo.Context) error {
	// Get ID from query parameter
	id, err := webhelpers.ParseInt64Query(c, constants.QueryParamID)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Info("Updating trouble report %d", id)

	// Get user from context
	user, err := webhelpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	// Get Title, Content and Attachments from form data
	title, content, attachments, err := h.validateDialogEditFormData(c)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Form validation failed: %v", err)
		return err
	}

	// Initialize dialog template data
	props := &dialogs.EditTroubleReportProps{
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
		logger.HTMXHandlerTroubleReports().Error("Failed to get trouble report %d: %v", id, err)
		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to get trouble report: "+err.Error())
	}

	// Filter out existing and new attachments
	var existingAttachmentIDs []int64
	var newAttachments []*attachmentmodels.Attachment
	for _, a := range props.Attachments {
		if a.GetID() > 0 {
			existingAttachmentIDs = append(existingAttachmentIDs, a.GetID())
		} else {
			newAttachments = append(newAttachments, a)
		}
	}

	// Update trouble report with existing and new attachments, title content and mods
	logger.HTMXHandlerTroubleReports().Debug(
		"Updating trouble report %d: title='%s', existing attachments=%d, new attachments=%d",
		id, title, len(existingAttachmentIDs), len(newAttachments),
	)

	tr.Title = title
	tr.Content = content
	tr.LinkedAttachments = existingAttachmentIDs

	err = h.DB.TroubleReports.UpdateWithAttachments(user, id, tr, newAttachments)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error(
			"Failed to update trouble report %d: %v",
			id, err,
		)

		return echo.NewHTTPError(dberror.GetHTTPStatusCode(err),
			"failed to update trouble report: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Info(
		"Successfully updated trouble report %d",
		id,
	)

	return h.handleGetDialogEdit(c, props)
}

// TODO: Do somehtings like the `get*FormData` method in "tools.go"
func (h *TroubleReports) validateDialogEditFormData(ctx echo.Context) (
	title, content string,
	attachments []*attachmentmodels.Attachment,
	err error,
) {
	logger.HTMXHandlerTroubleReports().Debug("Validating dialog edit form data")

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Invalid title form value: %v", err)
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			dberror.WrapError(err, "invalid title form value"))
	}
	title = webhelpers.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Invalid content form value: %v", err)
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			dberror.WrapError(err, "invalid content form value"))
	}
	content = webhelpers.SanitizeInput(content)

	// Process existing attachments and their order
	attachments, err = h.processAttachments(ctx)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to process attachments: %v", err)
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			dberror.WrapError(err, "failed to process attachments"))
	}

	logger.HTMXHandlerTroubleReports().Debug("Form validation successful: title='%s', attachments=%d", title, len(attachments))
	return title, content, attachments, nil
}

func (h *TroubleReports) processAttachments(ctx echo.Context) ([]*attachmentmodels.Attachment, error) {
	logger.HTMXHandlerTroubleReports().Debug("Processing attachments")
	var attachments []*attachmentmodels.Attachment

	// Get existing attachments if editing
	if idStr := ctx.QueryParam(constants.QueryParamID); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err == nil {
			if existingTR, err := h.DB.TroubleReports.Get(id); err == nil {
				if loadedAttachments, err := h.DB.TroubleReports.LoadAttachments(
					existingTR); err == nil {
					attachments = make([]*attachmentmodels.Attachment, len(loadedAttachments))
					copy(attachments, loadedAttachments)
					logger.HTMXHandlerTroubleReports().Debug("Loaded %d existing attachments for trouble report %d", len(loadedAttachments), id)
				}
			}
		}
	}

	// Handle new file uploads
	form, err := ctx.MultipartForm()
	if err != nil {
		return attachments, nil
	}

	existingAttachmentsToRemove := strings.Split(
		form.Value[constants.ExistingAttachmentsRemovalFormField][0],
		",",
	)
	logger.HTMXHandlerTroubleReports().Debug("Removing %d existing attachments", len(existingAttachmentsToRemove))
	for _, a := range existingAttachmentsToRemove {
		for i, a2 := range attachments {
			if a2.ID == a {
				attachments = slices.Delete(attachments, i, 1)
				break
			}
		}
	}

	files := form.File[constants.AttachmentsFormField]
	logger.HTMXHandlerTroubleReports().Debug("Processing %d new file uploads", len(files))
	for i, fileHeader := range files {
		if len(attachments) >= 10 {
			break
		}

		if fileHeader.Size == 0 {
			continue
		}

		attachment, err := h.processFileUpload(fileHeader, i+len(attachments))
		if err != nil {
			logger.HTMXHandlerTroubleReports().Error("Failed to process file %s: %v", fileHeader.Filename, err)
			return nil, fmt.Errorf("failed to process file %s: %w", fileHeader.Filename, err)
		}

		if attachment != nil {
			attachments = append(attachments, attachment)
		}
	}

	logger.HTMXHandlerTroubleReports().Debug("Successfully processed %d total attachments", len(attachments))
	return attachments, nil
}

func (h *TroubleReports) processFileUpload(
	fileHeader *multipart.FileHeader,
	index int,
) (*attachmentmodels.Attachment, error) {
	logger.HTMXHandlerTroubleReports().Debug("Processing file upload: %s (size: %d bytes)", fileHeader.Filename, fileHeader.Size)

	if fileHeader.Size > attachmentmodels.MaxDataSize {
		logger.HTMXHandlerTroubleReports().Error("File %s is too large: %d bytes (max: %d)", fileHeader.Filename, fileHeader.Size, attachmentmodels.MaxDataSize)
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
	if len(attachmentID) > attachmentmodels.MaxIDLength {
		maxFilenameLen := attachmentmodels.MaxIDLength - len(timestamp) -
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
		logger.HTMXHandlerTroubleReports().Error("File %s is not an image: %s", fileHeader.Filename, mimeType)
		return nil, fmt.Errorf("only image files are allowed (JPG, PNG, GIF, BMP, SVG, WebP)")
	}

	attachment := &attachmentmodels.Attachment{
		ID:       attachmentID,
		MimeType: mimeType,
		Data:     data,
	}

	if err := attachment.Validate(); err != nil {
		logger.HTMXHandlerTroubleReports().Error("Invalid attachment: %v", err)
		return nil, fmt.Errorf("invalid attachment: %w", err)
	}

	logger.HTMXHandlerTroubleReports().Debug("Successfully processed file upload: %s", sanitizedFilename)
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
