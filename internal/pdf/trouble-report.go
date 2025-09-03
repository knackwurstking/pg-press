package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/knackwurstking/pgpress/internal/database/models"
	"github.com/knackwurstking/pgpress/internal/database/services/troublereport"
)

// Options contains common options for PDF generation
type troubleReportOptions struct {
	*imageOptions
	Report *troublereport.TroubleReportWithAttachments
}

func GenerateTroubleReportPDF(
	tr *troublereport.TroubleReportWithAttachments,
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
	addTroubleReportMetadataSection(o)
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

func addTroubleReportMetadataSection(o *troubleReportOptions) {
	if len(o.Report.Mods) == 0 {
		return
	}

	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10, "METADATEN", "1", 1, "L", true, 0, "")
	o.PDF.Ln(5)

	earliestTime, latestTime, creator, lastModifier := getTroubleReportMetadataInfo(o)

	o.PDF.SetFont("Arial", "", 11)
	createdAt := time.Unix(0, earliestTime*int64(time.Millisecond))
	createdText := fmt.Sprintf("Erstellt am: %s",
		createdAt.Format("02.01.2006 15:04:05"))
	if creator != nil {
		createdText += fmt.Sprintf(" von %s", creator.UserName)
	}
	o.PDF.MultiCell(0, 6, o.Translator(createdText), "", "", false)

	if latestTime != earliestTime {
		lastModifiedAt := time.Unix(0, latestTime*int64(time.Millisecond))
		modifiedText := fmt.Sprintf("Zuletzt geändert: %s",
			lastModifiedAt.Format("02.01.2006 15:04:05"))
		if lastModifier != nil {
			modifiedText += fmt.Sprintf(" von %s", lastModifier.UserName)
		}
		o.PDF.MultiCell(0, 6, o.Translator(modifiedText), "", "", false)
	}

	o.PDF.Cell(0, 6, o.Translator(
		fmt.Sprintf("Anzahl Änderungen: %d", len(o.Report.Mods)),
	))
	o.PDF.Ln(13)
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

func getTroubleReportMetadataInfo(o *troubleReportOptions) (
	earliestTime, latestTime int64,
	creator *models.User, lastModifier *models.User,
) {
	earliestTime, latestTime = o.Report.Mods[0].Time, o.Report.Mods[0].Time

	for _, mod := range o.Report.Mods {
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

func getTroubleReportImageAttachments(attachments []*models.Attachment) []*models.Attachment {
	var images []*models.Attachment
	for _, attachment := range attachments {
		if attachment.IsImage() {
			images = append(images, attachment)
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
