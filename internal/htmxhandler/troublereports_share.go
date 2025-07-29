package htmxhandler

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/labstack/echo/v4"

	"github.com/knackwurstking/pg-vis/internal/database"
	"github.com/knackwurstking/pg-vis/internal/logger"
	"github.com/knackwurstking/pg-vis/internal/utils"
)

// pdfOptions contains common options for PDF generation
type pdfOptions struct {
	PDF        *gofpdf.Fpdf
	Translator func(string) string
	Report     *database.TroubleReportWithAttachments
}

// imageLayoutOptions contains layout options for image rendering
type imageLayoutOptions struct {
	PageWidth   float64
	LeftMargin  float64
	RightMargin float64
	UsableWidth float64
	ImageWidth  float64
}

// imagePositionOptions contains positioning options for image rendering
type imagePositionOptions struct {
	StartIndex  int
	TotalImages int
	ImageWidth  float64
	LeftMargin  float64
	RightX      float64
	CurrentY    *float64
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
			"Failed to retrieve trouble report %d for PDF generation: %v",
			id, err,
		)
		return utils.HandlePgvisError(c, err)
	}

	pdfBuffer, err := h.generateTroubleReportPDF(tr)
	if err != nil {
		logger.TroubleReport().Error(
			"Failed to generate PDF for trouble report %d: %v", tr.ID, err,
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"Fehler beim Erstellen der PDF",
		)
	}

	logger.TroubleReport().Info(
		"Successfully generated PDF for trouble report %d (size: %d bytes)",
		tr.ID, pdfBuffer.Len())

	return h.setupPDFResponse(c, tr, pdfBuffer)
}

func (h *TroubleReports) generateTroubleReportPDF(
	tr *database.TroubleReportWithAttachments,
) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	translator := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	opts := &pdfOptions{
		PDF:        pdf,
		Translator: translator,
		Report:     tr,
	}

	h.addPDFHeader(opts)
	h.addPDFTitleSection(opts)
	h.addPDFContentSection(opts)
	h.addPDFMetadataSection(opts)
	h.addPDFImagesSection(opts)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func (h *TroubleReports) addPDFHeader(opts *pdfOptions) {
	opts.PDF.SetFont("Arial", "B", 20)
	opts.PDF.SetTextColor(0, 51, 102)
	opts.PDF.Cell(0, 15, opts.Translator("Fehlerbericht"))
	opts.PDF.Ln(10)

	opts.PDF.SetFont("Arial", "", 12)
	opts.PDF.SetTextColor(128, 128, 128)
	opts.PDF.Cell(0, 8, fmt.Sprintf("Report-ID: #%d", opts.Report.ID))
	opts.PDF.Ln(15)

	opts.PDF.SetTextColor(0, 0, 0)
}

func (h *TroubleReports) addPDFTitleSection(opts *pdfOptions) {
	opts.PDF.SetFont("Arial", "B", 14)
	opts.PDF.SetFillColor(240, 248, 255)
	opts.PDF.CellFormat(0, 10, "TITEL", "1", 1, "L", true, 0, "")
	opts.PDF.Ln(5)
	opts.PDF.SetFont("Arial", "", 12)
	opts.PDF.MultiCell(0, 8, opts.Translator(opts.Report.Title), "", "", false)
	opts.PDF.Ln(8)
}

func (h *TroubleReports) addPDFContentSection(opts *pdfOptions) {
	opts.PDF.SetFont("Arial", "B", 14)
	opts.PDF.SetFillColor(240, 248, 255)
	opts.PDF.CellFormat(0, 10, "INHALT", "1", 1, "L", true, 0, "")
	opts.PDF.Ln(5)
	opts.PDF.SetFont("Arial", "", 11)
	opts.PDF.MultiCell(0, 6, opts.Translator(opts.Report.Content), "", "", false)
	opts.PDF.Ln(8)
}

func (h *TroubleReports) addPDFMetadataSection(opts *pdfOptions) {
	if len(opts.Report.Mods) == 0 {
		return
	}

	opts.PDF.SetFont("Arial", "B", 14)
	opts.PDF.SetFillColor(240, 248, 255)
	opts.PDF.CellFormat(0, 10, "METADATEN", "1", 1, "L", true, 0, "")
	opts.PDF.Ln(5)

	earliestTime, latestTime, creator, lastModifier := h.getMetadataInfo(opts.Report)

	opts.PDF.SetFont("Arial", "", 11)
	createdAt := time.Unix(0, earliestTime*int64(time.Millisecond))
	createdText := fmt.Sprintf("Erstellt am: %s",
		createdAt.Format("02.01.2006 15:04:05"))
	if creator != nil {
		createdText += fmt.Sprintf(" von %s", creator.UserName)
	}
	opts.PDF.MultiCell(0, 6, opts.Translator(createdText), "", "", false)

	if latestTime != earliestTime {
		lastModifiedAt := time.Unix(0, latestTime*int64(time.Millisecond))
		modifiedText := fmt.Sprintf("Zuletzt geändert: %s",
			lastModifiedAt.Format("02.01.2006 15:04:05"))
		if lastModifier != nil {
			modifiedText += fmt.Sprintf(" von %s", lastModifier.UserName)
		}
		opts.PDF.MultiCell(0, 6, opts.Translator(modifiedText), "", "", false)
	}

	opts.PDF.Cell(0, 6, opts.Translator(
		fmt.Sprintf("Anzahl Änderungen: %d", len(opts.Report.Mods)),
	))
	opts.PDF.Ln(13)
}

func (h *TroubleReports) getMetadataInfo(
	tr *database.TroubleReportWithAttachments,
) (
	earliestTime, latestTime int64,
	creator *database.User, lastModifier *database.User,
) {
	earliestTime, latestTime = tr.Mods[0].Time, tr.Mods[0].Time

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

	return earliestTime, latestTime, creator, lastModifier
}

func (h *TroubleReports) addPDFImagesSection(opts *pdfOptions) {
	if len(opts.Report.LoadedAttachments) == 0 {
		return
	}

	images := h.getImageAttachments(opts.Report.LoadedAttachments)
	if len(images) == 0 {
		return
	}

	opts.PDF.AddPage()
	opts.PDF.SetFont("Arial", "B", 14)
	opts.PDF.SetFillColor(240, 248, 255)
	opts.PDF.CellFormat(0, 10,
		opts.Translator(fmt.Sprintf("BILDER (%d)", len(images))),
		"1", 1, "L", true, 0, "")

	h.renderImagesInGrid(opts, images)
}

func (h *TroubleReports) getImageAttachments(
	attachments []*database.Attachment,
) []*database.Attachment {
	var images []*database.Attachment
	for _, attachment := range attachments {
		if attachment.IsImage() {
			images = append(images, attachment)
		}
	}
	return images
}

func (h *TroubleReports) renderImagesInGrid(
	opts *pdfOptions,
	images []*database.Attachment,
) {
	pageWidth, _ := opts.PDF.GetPageSize()
	leftMargin, _, rightMargin, _ := opts.PDF.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin
	imageWidth := (usableWidth - 10) / 2 // 10mm spacing between images

	layout := &imageLayoutOptions{
		PageWidth:   pageWidth,
		LeftMargin:  leftMargin,
		RightMargin: rightMargin,
		UsableWidth: usableWidth,
		ImageWidth:  imageWidth,
	}

	opts.PDF.Ln(10)

	var currentY float64
	_, currentY = opts.PDF.GetXY()

	// Process images in pairs (2 per row)
	for i := 0; i < len(images); i += 2 {
		position := &imagePositionOptions{
			StartIndex:  i,
			TotalImages: len(images),
			ImageWidth:  imageWidth,
			LeftMargin:  leftMargin,
			RightX:      leftMargin + imageWidth + 10,
			CurrentY:    &currentY,
		}

		h.processImageRow(opts, layout, position, images)
		currentY += 15 // Add spacing between rows
		opts.PDF.SetXY(leftMargin, currentY)
	}
}

func (h *TroubleReports) processImageRow(
	opts *pdfOptions,
	layout *imageLayoutOptions,
	position *imagePositionOptions,
	images []*database.Attachment,
) {
	leftHeight, rightHeight := h.calculateImageHeights(opts.PDF, images, position.StartIndex, layout.ImageWidth)
	actualRowHeight := max(leftHeight, rightHeight)
	if actualRowHeight == 0 {
		actualRowHeight = 60.0
	}

	captionY := *position.CurrentY
	imageY := captionY + 6

	// Check if we need a new page
	if imageY+actualRowHeight+25 > 270 {
		opts.PDF.AddPage()
		_, *position.CurrentY = opts.PDF.GetXY()
		captionY = *position.CurrentY
		imageY = captionY + 6
	}

	h.addImageCaptions(opts, position, captionY)
	h.addImages(opts.PDF, images, position, imageY)

	// Update currentY to the bottom of the images
	*position.CurrentY = imageY + actualRowHeight
}

func (h *TroubleReports) calculateImageHeights(
	pdf *gofpdf.Fpdf,
	images []*database.Attachment,
	startIndex int,
	imageWidth float64,
) (leftHeight, rightHeight float64) {
	// Calculate left image height
	if startIndex < len(images) {
		leftHeight = h.calculateSingleImageHeight(pdf, images[startIndex], imageWidth)
	}

	// Calculate right image height if it exists
	if startIndex+1 < len(images) {
		rightHeight = h.calculateSingleImageHeight(pdf, images[startIndex+1], imageWidth)
	}

	return leftHeight, rightHeight
}

func (h *TroubleReports) calculateSingleImageHeight(
	pdf *gofpdf.Fpdf,
	image *database.Attachment,
	imageWidth float64,
) (height float64) {
	tmpFile, err := h.createTempImageFile(image)
	if err != nil {
		return 60.0
	}
	defer os.Remove(tmpFile)

	imageType := h.getImageType(image.MimeType)
	info := pdf.RegisterImage(tmpFile, imageType)
	if info != nil {
		return (imageWidth * info.Height()) / info.Width()
	}

	return 60.0
}

func (h *TroubleReports) addImageCaptions(
	opts *pdfOptions,
	position *imagePositionOptions,
	captionY float64,
) {
	opts.PDF.SetFont("Arial", "", 9)

	// Left image caption
	opts.PDF.SetXY(position.LeftMargin, captionY)
	opts.PDF.CellFormat(position.ImageWidth, 4,
		opts.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+1)),
		"0", 0, "C", false, 0, "")

	// Right image caption (if exists)
	if position.StartIndex+1 < position.TotalImages {
		opts.PDF.SetXY(position.RightX, captionY)
		opts.PDF.CellFormat(position.ImageWidth, 4,
			opts.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+2)),
			"0", 0, "C", false, 0, "")
	}
}

func (h *TroubleReports) addImages(
	pdf *gofpdf.Fpdf,
	images []*database.Attachment,
	position *imagePositionOptions,
	imageY float64,
) {
	// Add left image
	if position.StartIndex < len(images) {
		h.addSingleImage(pdf, images[position.StartIndex], position.LeftMargin, imageY, position.ImageWidth)
	}

	// Add right image (if it exists)
	if position.StartIndex+1 < len(images) {
		h.addSingleImage(pdf, images[position.StartIndex+1], position.RightX, imageY, position.ImageWidth)
	}
}

func (h *TroubleReports) addSingleImage(
	pdf *gofpdf.Fpdf,
	image *database.Attachment,
	x, y, width float64,
) {
	tmpFile, err := h.createTempImageFile(image)
	if err != nil {
		return
	}
	defer os.Remove(tmpFile)

	imageType := h.getImageType(image.MimeType)
	pdf.Image(tmpFile, x, y, width, 0, false, imageType, 0, "")
}

func (h *TroubleReports) createTempImageFile(
	image *database.Attachment,
) (string, error) {
	tmpFile, err := os.CreateTemp("",
		fmt.Sprintf("attachment_%s_*.jpg", image.ID))
	if err != nil {
		return "", err
	}

	_, err = tmpFile.Write(image.Data)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func (h *TroubleReports) getImageType(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return "JPG"
	case "image/png":
		return "PNG"
	case "image/gif":
		return "GIF"
	default:
		return "JPG"
	}
}

func (h *TroubleReports) setupPDFResponse(
	c echo.Context,
	tr *database.TroubleReportWithAttachments,
	buf *bytes.Buffer,
) error {
	filename := fmt.Sprintf("fehlerbericht_%d_%s.pdf",
		tr.ID, time.Now().Format("2006-01-02"))

	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=%s", filename))
	c.Response().Header().Set("Content-Length",
		fmt.Sprintf("%d", buf.Len()))

	return c.Blob(http.StatusOK, "application/pdf", buf.Bytes())
}
