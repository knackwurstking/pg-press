package handler

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

	h.addPDFHeader(pdf, translator, tr)
	h.addPDFTitleSection(pdf, translator, tr)
	h.addPDFContentSection(pdf, translator, tr)
	h.addPDFMetadataSection(pdf, translator, tr)
	h.addPDFImagesSection(pdf, translator, tr)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func (h *TroubleReports) addPDFHeader(
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	tr *database.TroubleReportWithAttachments,
) {
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 51, 102)
	pdf.Cell(0, 15, translator("Fehlerbericht"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 8, fmt.Sprintf("Report-ID: #%d", tr.ID))
	pdf.Ln(15)

	pdf.SetTextColor(0, 0, 0)
}

func (h *TroubleReports) addPDFTitleSection(
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	tr *database.TroubleReportWithAttachments,
) {
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "TITEL", "1", 1, "L", true, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 8, translator(tr.Title), "", "", false)
	pdf.Ln(8)
}

func (h *TroubleReports) addPDFContentSection(
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	tr *database.TroubleReportWithAttachments,
) {
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "INHALT", "1", 1, "L", true, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, translator(tr.Content), "", "", false)
	pdf.Ln(8)
}

func (h *TroubleReports) addPDFMetadataSection(
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	tr *database.TroubleReportWithAttachments,
) {
	if len(tr.Mods) == 0 {
		return
	}

	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10, "METADATEN", "1", 1, "L", true, 0, "")
	pdf.Ln(5)

	earliestTime, latestTime, creator, lastModifier := h.getMetadataInfo(tr)

	pdf.SetFont("Arial", "", 11)
	createdAt := time.Unix(0, earliestTime*int64(time.Millisecond))
	createdText := fmt.Sprintf("Erstellt am: %s",
		createdAt.Format("02.01.2006 15:04:05"))
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

	pdf.Cell(0, 6, translator(
		fmt.Sprintf("Anzahl Änderungen: %d", len(tr.Mods)),
	))
	pdf.Ln(13)
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

func (h *TroubleReports) addPDFImagesSection(
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	tr *database.TroubleReportWithAttachments,
) {
	if len(tr.LoadedAttachments) == 0 {
		return
	}

	images := h.getImageAttachments(tr.LoadedAttachments)
	if len(images) == 0 {
		return
	}

	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(240, 248, 255)
	pdf.CellFormat(0, 10,
		translator(fmt.Sprintf("BILDER (%d)", len(images))),
		"1", 1, "L", true, 0, "")

	h.renderImagesInGrid(pdf, translator, images)
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
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	images []*database.Attachment,
) {
	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin
	imageWidth := (usableWidth - 10) / 2 // 10mm spacing between images

	pdf.Ln(10)

	var currentY float64
	_, currentY = pdf.GetXY()

	// Process images in pairs (2 per row)
	for i := 0; i < len(images); i += 2 {
		h.processImageRow(
			pdf,
			translator,
			images,
			i,
			imageWidth,
			leftMargin,
			&currentY,
		)
		currentY += 15 // Add spacing between rows
		pdf.SetXY(leftMargin, currentY)
	}
}

func (h *TroubleReports) processImageRow(
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	images []*database.Attachment,
	startIndex int,
	imageWidth, leftMargin float64,
	currentY *float64,
) {
	leftHeight, rightHeight := h.calculateImageHeights(pdf, images, startIndex, imageWidth)
	actualRowHeight := max(leftHeight, rightHeight)
	if actualRowHeight == 0 {
		actualRowHeight = 60.0
	}

	captionY := *currentY
	imageY := captionY + 6
	rightX := leftMargin + imageWidth + 10

	// Check if we need a new page
	if imageY+actualRowHeight+25 > 270 {
		pdf.AddPage()
		_, *currentY = pdf.GetXY()
		captionY = *currentY
		imageY = captionY + 6
	}

	h.addImageCaptions(
		pdf,
		translator,
		startIndex,
		len(images),
		imageWidth,
		leftMargin,
		rightX,
		captionY,
	)
	h.addImages(pdf, images, startIndex, imageWidth, leftMargin, rightX, imageY)

	// Update currentY to the bottom of the images
	*currentY = imageY + actualRowHeight
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
	pdf *gofpdf.Fpdf,
	translator func(string) string,
	startIndex, totalImages int,
	imageWidth, leftMargin, rightX, captionY float64,
) {
	pdf.SetFont("Arial", "", 9)

	// Left image caption
	pdf.SetXY(leftMargin, captionY)
	pdf.CellFormat(imageWidth, 4,
		translator(fmt.Sprintf("Anhang %d", startIndex+1)),
		"0", 0, "C", false, 0, "")

	// Right image caption (if exists)
	if startIndex+1 < totalImages {
		pdf.SetXY(rightX, captionY)
		pdf.CellFormat(imageWidth, 4,
			translator(fmt.Sprintf("Anhang %d", startIndex+2)),
			"0", 0, "C", false, 0, "")
	}
}

func (h *TroubleReports) addImages(
	pdf *gofpdf.Fpdf,
	images []*database.Attachment,
	startIndex int,
	imageWidth, leftMargin, rightX, imageY float64,
) {
	// Add left image
	if startIndex < len(images) {
		h.addSingleImage(pdf, images[startIndex], leftMargin, imageY, imageWidth)
	}

	// Add right image (if it exists)
	if startIndex+1 < len(images) {
		h.addSingleImage(pdf, images[startIndex+1], rightX, imageY, imageWidth)
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
