package pdf

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/knackwurstking/pgpress/models"

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

	if o.Report.UseMarkdown {
		renderMarkdownContentToPDF(o)
	} else {
		o.PDF.SetFont("Arial", "", 11)
		o.PDF.MultiCell(0, 6, o.Translator(o.Report.Content), "", "", false)
	}
	o.PDF.Ln(8)
}

// renderMarkdownContentToPDF renders markdown content with basic formatting in PDF
func renderMarkdownContentToPDF(o *troubleReportOptions) {
	content := o.Report.Content
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			o.PDF.Ln(3)
			continue
		}

		// Handle headers
		if strings.HasPrefix(line, "### ") {
			o.PDF.SetFont("Arial", "B", 11)
			o.PDF.MultiCell(0, 6, o.Translator(strings.TrimSpace(line[4:])), "", "", false)
			o.PDF.Ln(2)
			continue
		}
		if strings.HasPrefix(line, "## ") {
			o.PDF.SetFont("Arial", "B", 12)
			o.PDF.MultiCell(0, 7, o.Translator(strings.TrimSpace(line[3:])), "", "", false)
			o.PDF.Ln(2)
			continue
		}
		if strings.HasPrefix(line, "# ") {
			o.PDF.SetFont("Arial", "B", 13)
			o.PDF.MultiCell(0, 8, o.Translator(strings.TrimSpace(line[2:])), "", "", false)
			o.PDF.Ln(3)
			continue
		}

		// Handle lists
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			o.PDF.SetFont("Arial", "", 10)
			o.PDF.Cell(10, 6, "•")
			o.PDF.MultiCell(0, 6, o.Translator(strings.TrimSpace(line[2:])), "", "", false)
			continue
		}

		// Handle numbered lists
		if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
			o.PDF.SetFont("Arial", "", 10)
			parts := regexp.MustCompile(`^(\d+\.\s)(.*)`).FindStringSubmatch(line)
			if len(parts) == 3 {
				o.PDF.Cell(10, 6, parts[1])
				o.PDF.MultiCell(0, 6, o.Translator(parts[2]), "", "", false)
			}
			continue
		}

		// Handle regular paragraphs with basic formatting
		o.PDF.SetFont("Arial", "", 10)
		formattedLine := renderBasicMarkdownFormatting(line)
		o.PDF.MultiCell(0, 6, o.Translator(formattedLine), "", "", false)
		o.PDF.Ln(1)
	}
}

// renderBasicMarkdownFormatting removes markdown syntax for PDF rendering
func renderBasicMarkdownFormatting(text string) string {
	// Remove bold formatting
	text = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`__(.*?)__`).ReplaceAllString(text, "$1")

	// Remove italic formatting
	text = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`_(.*?)_`).ReplaceAllString(text, "$1")

	// Remove strikethrough
	text = regexp.MustCompile(`~~(.*?)~~`).ReplaceAllString(text, "$1")

	// Remove code formatting
	text = regexp.MustCompile("`([^`]*)`").ReplaceAllString(text, "$1")

	// Remove link formatting, keep text
	text = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(text, "$1")

	return text
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
