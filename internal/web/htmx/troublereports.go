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

	"github.com/knackwurstking/pgpress/internal/constants"
	"github.com/knackwurstking/pgpress/internal/database"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/web/helpers"
	"github.com/knackwurstking/pgpress/internal/web/templates/components"
	"github.com/knackwurstking/pgpress/internal/web/templates/dialogs"
	"github.com/knackwurstking/pgpress/internal/web/templates/troublereportspage"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"

	"github.com/labstack/echo/v4"
)

type TroubleReports struct {
	DB *database.DB
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	helpers.RegisterEchoRoutes(
		e,
		[]*helpers.EchoRoute{
			// Dialog edit routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/edit",
				func(c echo.Context) error {
					return h.handleGetDialogEdit(c, nil)
				}),

			helpers.NewEchoRoute(http.MethodPost, "/htmx/trouble-reports/edit",
				h.handleDialogEditPOST),

			helpers.NewEchoRoute(http.MethodPut, "/htmx/trouble-reports/edit",
				h.handleDialogEditPUT),

			// Data routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/data",
				h.handleDataGET,
			),

			helpers.NewEchoRoute(http.MethodDelete, "/htmx/trouble-reports/data",
				h.handleDataDELETE,
			),

			// Attachments preview routes
			helpers.NewEchoRoute(http.MethodGet, "/htmx/trouble-reports/attachments-preview",
				h.handleGetAttachmentsPreview),
		},
	)
}

func (h *TroubleReports) handleDataGET(c echo.Context) error {
	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("User %s fetching trouble reports list", user.Name)

	trs, err := h.DB.TroubleReports.ListWithAttachments()
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to load trouble reports: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Debug("Found %d trouble reports for user %s", len(trs), user.Name)

	troubleReportsList := troublereportspage.List(user, trs)
	if err := troubleReportsList.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to render trouble reports list component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleDataDELETE(c echo.Context) error {
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		return err
	}

	if !user.IsAdmin() {
		return echo.NewHTTPError(
			http.StatusForbidden,
			"administrator privileges required",
		)
	}

	logger.HTMXHandlerTroubleReports().Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.Name, user.TelegramID, id)

	if removedReport, err := h.DB.TroubleReports.RemoveWithAttachments(id, user); err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to delete trouble report %d: %v", id, err)
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to delete trouble report: "+err.Error())
	} else {
		logger.HTMXHandlerTroubleReports().Info("Successfully deleted trouble report %d (%s)", removedReport.ID, removedReport.Title)
	}

	return h.handleDataGET(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("Fetching attachments preview for trouble report %d", id)

	tr, err := h.DB.TroubleReports.GetWithAttachments(id)
	if err != nil {
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to load trouble report: "+err.Error())
	}

	logger.HTMXHandlerTroubleReports().Debug(
		"Rendering attachments preview with %d attachments", len(tr.LoadedAttachments),
	)

	attachmentsPreview := components.AttachmentsPreview(tr.LoadedAttachments)
	if err := attachmentsPreview.Render(c.Request().Context(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError,
			"failed to render attachments preview component: "+err.Error())
	}
	return nil
}

func (h *TroubleReports) handleGetDialogEdit(
	c echo.Context,
	props *dialogs.EditTroubleReportProps,
) error {
	if props == nil {
		props = &dialogs.EditTroubleReportProps{}
	}

	if !props.Close {
		props.Close = helpers.ParseBoolQuery(c, "close")
	}

	if !props.Close && !props.InvalidTitle && !props.InvalidContent {
		if idStr := c.QueryParam("id"); idStr != "" {
			id, err := helpers.ParseInt64Query(c, "id")
			if err != nil {
				return err
			}

			props.ID = id

			logger.HTMXHandlerTroubleReports().Debug("Loading trouble report %d for editing", id)
			tr, err := h.DB.TroubleReports.Get(id)
			if err != nil {
				logger.HTMXHandlerTroubleReports().Error("Failed to get trouble report %d: %v", id, err)
				return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
					"failed to get trouble report: "+err.Error())
			}

			props.Title = tr.Title
			props.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReports.LoadAttachments(tr); err == nil {
				props.Attachments = loadedAttachments
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

func (h *TroubleReports) handleDialogEditPOST(c echo.Context) error {
	props := &dialogs.EditTroubleReportProps{
		Close: true,
	}

	user, err := helpers.GetUserFromContext(c)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Form validation failed: %v", err)
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("User %s is creating a new trouble report", user.Name)

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
		tr := models.NewTroubleReport(title, content)

		logger.HTMXHandlerTroubleReports().Debug(
			"Creating trouble report: title='%s', attachments=%d",
			title, len(attachments),
		)

		err := h.DB.TroubleReports.AddWithAttachments(tr, user, attachments...)
		if err != nil {
			logger.HTMXHandlerTroubleReports().Error(
				"Failed to add trouble report: %v",
				err,
			)

			return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
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

func (h *TroubleReports) handleDialogEditPUT(c echo.Context) error {
	// Get ID from query parameter
	id, err := helpers.ParseInt64Query(c, "id")
	if err != nil {
		return err
	}

	logger.HTMXHandlerTroubleReports().Debug("Updating trouble report %d", id)

	// Get user from context
	user, err := helpers.GetUserFromContext(c)
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
		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
			"failed to get trouble report: "+err.Error())
	}

	// Filter out existing and new attachments
	var existingAttachmentIDs []int64
	var newAttachments []*models.Attachment
	for _, a := range props.Attachments {
		if a.GetID() > 0 {
			existingAttachmentIDs = append(existingAttachmentIDs, a.GetID())
		} else {
			newAttachments = append(newAttachments, a)
		}
	}

	// Update trouble report with existing and new attachments, title content and mods
	logger.HTMXHandlerTroubleReports().Debug("Updating trouble report %d with %d attachments",
		id, len(props.Attachments))

	tr.Title = title
	tr.Content = content
	tr.LinkedAttachments = existingAttachmentIDs

	err = h.DB.TroubleReports.UpdateWithAttachments(id, tr, user, newAttachments...)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error(
			"Failed to update trouble report %d: %v",
			id, err,
		)

		return echo.NewHTTPError(utils.GetHTTPStatusCode(err),
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
	attachments []*models.Attachment,
	err error,
) {

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Invalid title form value: %v", err)
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			fmt.Errorf("invalid title form value: %v", err))
	}
	title = helpers.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Invalid content form value: %v", err)
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			fmt.Errorf("invalid content form value: %v", err))
	}
	content = helpers.SanitizeInput(content)

	// Process existing attachments and their order
	attachments, err = h.processAttachments(ctx)
	if err != nil {
		logger.HTMXHandlerTroubleReports().Error("Failed to process attachments: %v", err)
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			fmt.Errorf("failed to process attachments: %v", err))
	}

	return title, content, attachments, nil
}

func (h *TroubleReports) processAttachments(ctx echo.Context) ([]*models.Attachment, error) {

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

	existingAttachmentsToRemove := strings.Split(
		form.Value[constants.ExistingAttachmentsRemovalFormField][0],
		",",
	)

	for _, a := range existingAttachmentsToRemove {
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
			logger.HTMXHandlerTroubleReports().Error("Failed to process file %s: %v", fileHeader.Filename, err)
			return nil, fmt.Errorf("failed to process file %s: %v", fileHeader.Filename, err)
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
) (*models.Attachment, error) {

	if fileHeader.Size > models.MaxDataSize {
		logger.HTMXHandlerTroubleReports().Error("File %s is too large: %d bytes (max: %d)", fileHeader.Filename, fileHeader.Size, models.MaxDataSize)
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
	sanitizedFilename := h.sanitizeFilename(fileHeader.Filename)
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	attachmentID := fmt.Sprintf("temp_%s_%s_%d", sanitizedFilename, timestamp, index)

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
		logger.HTMXHandlerTroubleReports().Error("File %s is not an image: %s", fileHeader.Filename, mimeType)
		return nil, fmt.Errorf("only image files are allowed (JPG, PNG, GIF, BMP, SVG, WebP)")
	}

	attachment := &models.Attachment{
		ID:       attachmentID,
		MimeType: mimeType,
		Data:     data,
	}

	if err := attachment.Validate(); err != nil {
		logger.HTMXHandlerTroubleReports().Error("Invalid attachment: %v", err)
		return nil, fmt.Errorf("invalid attachment: %v", err)
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
