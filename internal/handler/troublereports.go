package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/constants"
	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/logger"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

const (
	adminPrivilegesRequiredMessage   = "administrator privileges required"
	invalidContentFormFieldMessage   = "invalid content form value"
	invalidTitleFormFieldMessage     = "invalid title form value"
	attachmentTooLargeMessage        = "image exceeds maximum size limit (10MB)"
	attachmentNotFoundMessage        = "image not found"
	invalidAttachmentMessage         = "invalid image data"
	nonImageFileMessage              = "only image files are allowed (JPG, PNG, GIF, BMP, SVG, WebP)"
	tooManyAttachmentsMessage        = "too many images (maximum 10 allowed)"
	attachmentProcessingErrorMessage = "failed to process image"
)

// Template data structures
type TroubleReportsDataTemplateData struct {
	TroubleReports []*database.TroubleReportWithAttachments `json:"trouble_reports"`
	User           *database.User                           `json:"user"`
}

type DialogEditTemplateData struct {
	ID                int                    `json:"id"`
	Submitted         bool                   `json:"submitted"`
	Title             string                 `json:"title"`
	Content           string                 `json:"content"`
	LinkedAttachments []*database.Attachment `json:"linked_attachments,omitempty"`
	InvalidTitle      bool                   `json:"invalid_title"`
	InvalidContent    bool                   `json:"invalid_content"`
	AttachmentError   string                 `json:"attachment_error,omitempty"`
}

type AttachmentsPreviewTemplateData struct {
	TroubleReport *database.TroubleReportWithAttachments `json:"trouble_report"`
}

type ModificationsTemplateData struct {
	User              *database.User
	TroubleReport     *database.TroubleReport
	LoadedAttachments []*database.Attachment
	Mods              database.Mods[database.TroubleReportMod]
}

// TroubleReportModification is a type alias for better readability
type TroubleReportModification = *database.Modified[database.TroubleReportMod]

func (mtd *ModificationsTemplateData) FirstModified() TroubleReportModification {
	if len(mtd.TroubleReport.Mods) == 0 {
		return nil
	}
	return mtd.TroubleReport.Mods[0]
}

type ModificationAttachmentsTemplateData struct {
	TroubleReport *database.TroubleReport
	Modification  *database.Modified[database.TroubleReportMod]
	Attachments   []*database.Attachment
}

type TroubleReports struct {
	*Base
}

func (h *TroubleReports) RegisterRoutes(e *echo.Echo) {
	e.GET(h.ServerPathPrefix+"/trouble-reports", h.handleMainPage)

	// Dialog edit routes
	editDialogPath := h.ServerPathPrefix + "/trouble-reports/dialog-edit"
	e.GET(editDialogPath, func(c echo.Context) error {
		return h.handleGetDialogEdit(c, nil)
	})
	e.POST(editDialogPath, h.handlePostDialogEdit)
	e.PUT(editDialogPath, h.handlePutDialogEdit)

	// Data routes
	dataPath := h.ServerPathPrefix + "/trouble-reports/data"
	e.GET(dataPath, h.handleGetData)
	e.DELETE(dataPath, h.handleDeleteData)

	attachmentsPreviewPath := h.ServerPathPrefix + "/trouble-reports/attachments-preview"
	e.GET(attachmentsPreviewPath, h.handleGetAttachmentsPreview)

	sharePdfPath := h.ServerPathPrefix + "/trouble-reports/share-pdf"
	e.GET(sharePdfPath, h.handleGetSharePdf)

	// Modifications routes
	modificationsPath := h.ServerPathPrefix + "/trouble-reports/modifications/:id"
	modificationsAttachmentsPath := h.ServerPathPrefix +
		"/trouble-reports/modifications/attachments-preview/:id"
	e.GET(modificationsPath, func(c echo.Context) error {
		return h.handleGetModifications(c, nil)
	})
	e.GET(modificationsAttachmentsPath, h.handleGetModificationAttachmentsPreview)
	e.POST(modificationsPath, h.handlePostModifications)

	// Attachment routes
	attachmentReorderPath := h.ServerPathPrefix + "/trouble-reports/attachments/reorder"
	e.POST(attachmentReorderPath, h.handlePostAttachmentReorder)

	attachmentPath := h.ServerPathPrefix + "/trouble-reports/attachments"
	e.GET(attachmentPath, h.handleGetAttachment)
	e.DELETE(attachmentPath, h.handleDeleteAttachment)
}

func (h *TroubleReports) handleMainPage(c echo.Context) error {
	return utils.HandleTemplate(c, nil,
		h.Templates,
		constants.TroubleReportsPageTemplates,
	)
}

// Dialog Edit handlers
func (h *TroubleReports) handleGetDialogEdit(
	c echo.Context,
	pageData *DialogEditTemplateData,
) *echo.HTTPError {
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

			tr, err := h.DB.TroubleReports.Get(id)
			if err != nil {
				return utils.HandlePgvisError(c, err)
			}

			pageData.Title = tr.Title
			pageData.Content = tr.Content

			// Load attachments for display
			if loadedAttachments, err := h.DB.TroubleReportService.LoadAttachments(tr); err == nil {
				pageData.LinkedAttachments = loadedAttachments
			}
		}
	}

	return utils.HandleTemplate(c, pageData,
		h.Templates,
		[]string{
			constants.TroubleReportsDialogEditComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handlePostDialogEdit(c echo.Context) error {
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
	dialogEditData.InvalidTitle = title == ""
	dialogEditData.InvalidContent = content == ""

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		dialogEditData.LinkedAttachments = attachments
		modified := database.NewModified[database.TroubleReportMod](user, database.TroubleReportMod{
			Title:             title,
			Content:           content,
			LinkedAttachments: []int64{}, // Will be set by the service
		})
		tr := database.NewTroubleReport(title, content, modified)

		if err := h.DB.TroubleReportService.AddWithAttachments(tr, attachments); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *TroubleReports) handlePutDialogEdit(c echo.Context) error {
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
		Submitted:      true,
		ID:             int(id),
		Title:          title,
		Content:        content,
		InvalidTitle:   title == "",
		InvalidContent: content == "",
	}

	if !dialogEditData.InvalidTitle && !dialogEditData.InvalidContent {
		dialogEditData.LinkedAttachments = attachments
		trOld, err := h.DB.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}

		tr := database.NewTroubleReport(title, content, trOld.Mods...)

		// Convert existing attachments to IDs for reordering
		var existingAttachmentIDs []int64
		for _, att := range attachments {
			if att.GetID() > 0 {
				existingAttachmentIDs = append(existingAttachmentIDs, att.GetID())
			}
		}

		// Filter out new attachments
		var newAttachments []*database.Attachment
		for _, att := range attachments {
			if att.GetID() == 0 {
				newAttachments = append(newAttachments, att)
			}
		}

		tr.LinkedAttachments = existingAttachmentIDs
		tr.Mods = append(tr.Mods, database.NewModified(user, database.TroubleReportMod{
			Title:             tr.Title,
			Content:           tr.Content,
			LinkedAttachments: []int64{},
		}))

		if err := h.DB.TroubleReportService.UpdateWithAttachments(id, tr, newAttachments); err != nil {
			return utils.HandlePgvisError(c, err)
		}
	} else {
		dialogEditData.Submitted = false
	}

	return h.handleGetDialogEdit(c, dialogEditData)
}

func (h *TroubleReports) validateDialogEditFormData(ctx echo.Context) (
	title, content string,
	attachments []*database.Attachment,
	httpErr *echo.HTTPError,
) {
	var err error

	title, err = url.QueryUnescape(ctx.FormValue(constants.TitleFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, invalidTitleFormFieldMessage))
	}
	title = utils.SanitizeInput(title)

	content, err = url.QueryUnescape(ctx.FormValue(constants.ContentFormField))
	if err != nil {
		return "", "", nil, echo.NewHTTPError(http.StatusBadRequest,
			database.WrapError(err, invalidContentFormFieldMessage))
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

// Data handlers
func (h *TroubleReports) handleGetData(c echo.Context) error {
	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	trs, err := h.DB.TroubleReportService.ListWithAttachments()
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		TroubleReportsDataTemplateData{
			TroubleReports: trs,
			User:           user,
		},
		h.Templates,
		[]string{
			constants.TroubleReportsDataComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handleDeleteData(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	if !user.IsAdmin() {
		return echo.NewHTTPError(
			http.StatusForbidden,
			adminPrivilegesRequiredMessage,
		)
	}

	logger.TroubleReport().Info("Administrator %s (Telegram ID: %d) is deleting trouble report %d",
		user.UserName, user.TelegramID, id)

	if err := h.DB.TroubleReportService.RemoveWithAttachments(id); err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return h.handleGetData(c)
}

func (h *TroubleReports) handleGetAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	return utils.HandleTemplate(
		c,
		AttachmentsPreviewTemplateData{
			TroubleReport: tr,
		},
		h.Templates,
		[]string{
			constants.TroubleReportsAttachmentsPreviewComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handleGetSharePdf(c echo.Context) error {
	id, herr := utils.ParseInt64Query(c, "id")
	if herr != nil {
		return herr
	}

	logger.TroubleReport().Info("Generating PDF for trouble report %d", id)

	tr, err := h.DB.TroubleReportService.GetWithAttachments(id)
	if err != nil {
		logger.TroubleReport().Error(
			"Failed to retrieve trouble report %d for PDF generation: %v", id, err)
		return utils.HandlePgvisError(c, err)
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	translator := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 51, 102)
	pdf.Cell(0, 15, translator("Fehlerbericht"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 8, fmt.Sprintf("Report-ID: #%d", tr.ID))
	pdf.Ln(15)

	pdf.SetTextColor(0, 0, 0)

	// Title section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "TITEL", "1", 1, "L", true, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 8, translator(tr.Title), "", "", false)
	pdf.Ln(8)

	// Content section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "INHALT", "1", 1, "L", true, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, translator(tr.Content), "", "", false)
	pdf.Ln(8)

	// Metadata
	if len(tr.Mods) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(240, 248, 255)
		pdf.CellFormat(0, 10, "METADATEN", "1", 1, "L", true, 0, "")
		pdf.Ln(5)

		var earliestTime, latestTime int64 = tr.Mods[0].Time, tr.Mods[0].Time
		var creator, lastModifier *database.User

		for _, mod := range tr.Mods {
			if mod.Time < earliestTime {
				earliestTime = mod.Time
				creator = mod.User
			}
			if mod.Time > latestTime {
				latestTime = mod.Time
				lastModifier = mod.User
			}
		}

		pdf.SetFont("Arial", "", 11)
		createdAt := time.Unix(0, earliestTime*int64(time.Millisecond))
		createdText := fmt.Sprintf("Erstellt am: %s", createdAt.Format("02.01.2006 15:04:05"))
		if creator != nil {
			createdText += fmt.Sprintf(" von %s", creator.UserName)
		}
		pdf.MultiCell(0, 6, translator(createdText), "", "", false)

		if latestTime != earliestTime {
			lastModifiedAt := time.Unix(0, latestTime*int64(time.Millisecond))
			modifiedText := fmt.Sprintf("Zuletzt geändert: %s",
				lastModifiedAt.Format("02.01.2006 15:04:05"))
			if lastModifier != nil {
				modifiedText += fmt.Sprintf(" von %s", lastModifier.UserName)
			}
			pdf.MultiCell(0, 6, translator(modifiedText), "", "", false)
		}

		pdf.Cell(0, 6, translator(fmt.Sprintf("Anzahl Änderungen: %d", len(tr.Mods))))
		pdf.Ln(13)
	}

	// Attachments
	if len(tr.LoadedAttachments) > 0 {
		// Collect only image attachments
		var images []*database.Attachment

		for _, attachment := range tr.LoadedAttachments {
			if attachment.IsImage() {
				images = append(images, attachment)
			}
		}

		// Only proceed if there are images
		if len(images) > 0 {
			pdf.SetFont("Arial", "B", 14)
			pdf.SetFillColor(240, 248, 255)
			pdf.CellFormat(0, 10,
				translator(fmt.Sprintf("BILDER (%d)", len(images))),
				"1", 1, "L", true, 0, "")

			// Display images 2 per row
			pageWidth, _ := pdf.GetPageSize()
			leftMargin, _, rightMargin, _ := pdf.GetMargins()
			usableWidth := pageWidth - leftMargin - rightMargin
			imageWidth := (usableWidth - 10) / 2 // 10mm spacing between images
			maxImageHeight := 60.0               // Maximum height for consistency

			pdf.Ln(10)

			// Process images in pairs (2 per row)
			var currentY float64
			_, currentY = pdf.GetXY()

			for i := 0; i < len(images); i += 2 {
				// Check if we need a new page
				if currentY+maxImageHeight+25 > 270 { // Leave space for footer
					pdf.AddPage()
					_, currentY = pdf.GetXY()
				}

				// Calculate positions for this row
				captionY := currentY
				imageY := captionY + 6
				rightX := leftMargin + imageWidth + 10

				// Add captions first (both left and right if applicable)
				pdf.SetFont("Arial", "", 9)

				// Left image caption
				pdf.SetXY(leftMargin, captionY)
				pdf.CellFormat(imageWidth, 4,
					translator(fmt.Sprintf("Anhang %d", i+1)),
					"0", 0, "C", false, 0, "")

				// Right image caption (if exists)
				if i+1 < len(images) {
					pdf.SetXY(rightX, captionY)
					pdf.CellFormat(imageWidth, 4,
						translator(fmt.Sprintf("Anhang %d", i+2)),
						"0", 0, "C", false, 0, "")
				}

				// Process left image
				leftImage := images[i]
				tmpFile1, err := os.CreateTemp("", fmt.Sprintf("attachment_%s_*.jpg", leftImage.ID))
				if err == nil {
					_, err = tmpFile1.Write(leftImage.Data)
					tmpFile1.Close()

					if err == nil {
						// Determine image type from mime type
						var imageType string
						switch leftImage.MimeType {
						case "image/jpeg", "image/jpg":
							imageType = "JPG"
						case "image/png":
							imageType = "PNG"
						case "image/gif":
							imageType = "GIF"
						default:
							imageType = "JPG" // Default fallback
						}

						// Add left image
						pdf.ImageOptions(tmpFile1.Name(), leftMargin, imageY, imageWidth, 0, false,
							gofpdf.ImageOptions{ImageType: imageType}, 0, "")
					}
					os.Remove(tmpFile1.Name())
				}

				// Process right image if it exists
				if i+1 < len(images) {
					rightImage := images[i+1]
					tmpFile2, err := os.CreateTemp("", fmt.Sprintf("attachment_%s_*.jpg", rightImage.ID))
					if err == nil {
						_, err = tmpFile2.Write(rightImage.Data)
						tmpFile2.Close()

						if err == nil {
							// Determine image type from mime type
							var imageType string
							switch rightImage.MimeType {
							case "image/jpeg", "image/jpg":
								imageType = "JPG"
							case "image/png":
								imageType = "PNG"
							case "image/gif":
								imageType = "GIF"
							default:
								imageType = "JPG" // Default fallback
							}

							// Add right image
							pdf.ImageOptions(tmpFile2.Name(), rightX, imageY, imageWidth, 0, false,
								gofpdf.ImageOptions{ImageType: imageType}, 0, "")
						}
						os.Remove(tmpFile2.Name())
					}
				}

				// Move to next row
				currentY = imageY + maxImageHeight + 15
				pdf.SetXY(leftMargin, currentY)
			}
		}
	}

	// Create buffer and write PDF
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		logger.TroubleReport().Error("Failed to generate PDF for trouble report %d: %v", tr.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Fehler beim Erstellen der PDF")
	}

	logger.TroubleReport().Info(
		"Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, buf.Len())

	filename := fmt.Sprintf("fehlerbericht_%d_%s.pdf",
		tr.ID, time.Now().Format("2006-01-02"))
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Length",
		fmt.Sprintf("%d", buf.Len()))

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}

// Modifications handlers
func (h *TroubleReports) handleGetModifications(
	c echo.Context,
	tr *database.TroubleReport,
) *echo.HTTPError {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	if tr == nil {
		var err error
		tr, err = h.DB.TroubleReports.Get(id)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	loadedAttachments, err := h.DB.TroubleReportService.LoadAttachments(tr)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	mods := slices.Clone(tr.Mods)
	slices.Reverse(mods)

	data := &ModificationsTemplateData{
		User:              user,
		TroubleReport:     tr,
		LoadedAttachments: loadedAttachments,
		Mods:              mods,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{
			constants.TroubleReportsModificationsComponentTemplatePath,
		},
	)
}

func (h *TroubleReports) handlePostModifications(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Move modification to the top
	newMods := []*database.Modified[database.TroubleReportMod]{}
	var mod *database.Modified[database.TroubleReportMod]
	for _, m := range tr.Mods {
		if m.Time == timeQuery {
			if mod != nil {
				logger.TroubleReport().Warn(
					"Multiple modifications with the same time, mod: %+v, m: %+v", mod, m)
				newMods = append(newMods, m)
			} else {
				mod = m
			}
		} else {
			newMods = append(newMods, m)
		}
	}

	if mod == nil {
		return utils.HandlePgvisError(c, errors.New("modification not found"))
	}

	mod.Time = time.Now().UnixMilli()

	// Update mods with new order
	tr.Mods = append(newMods, mod)

	// Update trouble reports data
	tr.Title = mod.Data.Title
	tr.Content = mod.Data.Content
	tr.LinkedAttachments = mod.Data.LinkedAttachments

	// Update database
	if err = h.DB.TroubleReports.Update(id, tr); err != nil {
		return utils.HandlePgvisError(c, database.WrapError(err, "failed to update trouble report"))
	}

	return h.handleGetModifications(c, tr)
}

func (h *TroubleReports) handleGetModificationAttachmentsPreview(c echo.Context) error {
	id, herr := utils.ParseInt64Param(c, constants.QueryParamID)
	if herr != nil {
		return herr
	}

	timeQuery, herr := utils.ParseInt64Query(c, constants.QueryParamTime)
	if herr != nil {
		return herr
	}

	tr, err := h.DB.TroubleReports.Get(id)
	if err != nil {
		return utils.HandlePgvisError(c, err)
	}

	// Find the specific modification
	var targetMod *database.Modified[database.TroubleReportMod]
	for _, mod := range tr.Mods {
		if mod.Time == timeQuery {
			targetMod = mod
			break
		}
	}

	if targetMod == nil {
		return utils.HandlePgvisError(c, errors.New("modification not found"))
	}

	// Load attachments for this modification
	var attachments []*database.Attachment
	if len(targetMod.Data.LinkedAttachments) > 0 {
		attachments, err = h.DB.Attachments.GetByIDs(targetMod.Data.LinkedAttachments)
		if err != nil {
			return utils.HandlePgvisError(c, err)
		}
	}

	data := &ModificationAttachmentsTemplateData{
		TroubleReport: tr,
		Modification:  targetMod,
		Attachments:   attachments,
	}

	return utils.HandleTemplate(
		c,
		data,
		h.Templates,
		[]string{
			constants.TroubleReportsModificationAttachmentsPreviewComponentTemplatePath,
		},
	)
}

// Attachment handlers
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
