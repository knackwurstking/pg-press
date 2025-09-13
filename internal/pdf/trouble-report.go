package pdf

import (
	"bytes"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"

	"github.com/jung-kurt/gofpdf/v2"
)

// Options contains common options for PDF generation
type troubleReportOptions struct {
	*imageOptions
	Report *models.TroubleReportWithAttachments
}

func GenerateTroubleReportPDF(
	tr *models.TroubleReportWithAttachments,
) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	o := &troubleReportOptions{
		imageOptions: &imageOptions{
			PDF:        pdf,
			Translator: pdf.UnicodeTranslatorFromDescriptor(""),
		},
		Report: tr,
	}

	addTroubleReportHeader(o)
	addTroubleReportTitleSection(o)
	addTroubleReportContentSection(o)
	addTroubleReportImagesSection(o)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func addTroubleReportHeader(o *troubleReportOptions) {
	o.PDF.SetFont("Arial", "B", 20)
	o.PDF.SetTextColor(0, 51, 102)
	o.PDF.Cell(0, 15, o.Translator("Fehlerbericht"))
	o.PDF.Ln(10)

	o.PDF.SetFont("Arial", "", 12)
	o.PDF.SetTextColor(128, 128, 128)
	o.PDF.Cell(0, 8, fmt.Sprintf("Report-ID: #%d", o.Report.ID))
	o.PDF.Ln(15)

	o.PDF.SetTextColor(0, 0, 0)
}

func addTroubleReportTitleSection(o *troubleReportOptions) {
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10, "TITEL", "1", 1, "L", true, 0, "")
	o.PDF.Ln(5)
	o.PDF.SetFont("Arial", "", 12)
	o.PDF.MultiCell(0, 8, o.Translator(o.Report.Title), "", "", false)
	o.PDF.Ln(8)
}

func addTroubleReportContentSection(o *troubleReportOptions) {
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10, "INHALT", "1", 1, "L", true, 0, "")
	o.PDF.Ln(5)
	o.PDF.SetFont("Arial", "", 11)
	o.PDF.MultiCell(0, 6, o.Translator(o.Report.Content), "", "", false)
	o.PDF.Ln(8)
}

func addTroubleReportImagesSection(o *troubleReportOptions) {
	if len(o.Report.LoadedAttachments) == 0 {
		return
	}

	images := getTroubleReportImageAttachments(o.Report.LoadedAttachments)
	if len(images) == 0 {
		return
	}

	o.PDF.AddPage()
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10,
		o.Translator(fmt.Sprintf("BILDER (%d)", len(images))),
		"1", 1, "L", true, 0, "")

	renderTroubleReportImagesInGrid(o, images)
}

func getTroubleReportImageAttachments(attachments []*models.Attachment) []*models.Attachment {
	var images []*models.Attachment
	for _, a := range attachments {
		if a.IsImage() {
			images = append(images, a)
		}
	}
	return images
}

func renderTroubleReportImagesInGrid(o *troubleReportOptions, images []*models.Attachment) {
	pageWidth, _ := o.PDF.GetPageSize()
	leftMargin, _, rightMargin, _ := o.PDF.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin
	imageWidth := (usableWidth - 10) / 2 // 10mm spacing between images

	layout := &imageLayoutOptions{
		PageWidth:   pageWidth,
		LeftMargin:  leftMargin,
		RightMargin: rightMargin,
		UsableWidth: usableWidth,
		ImageWidth:  imageWidth,
	}

	o.PDF.Ln(10)

	var currentY float64
	_, currentY = o.PDF.GetXY()

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

		processImageRow(o.imageOptions, layout, position, images)
		currentY += 15 // Add spacing between rows
		o.PDF.SetXY(leftMargin, currentY)
	}
}
