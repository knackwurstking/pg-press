package pdf

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/knackwurstking/pg-press/internal/shared"

	"github.com/jung-kurt/gofpdf/v2"
)

// Options contains common options for PDF generation
type troubleReportOptions struct {
	*imageOptions
	Report *shared.TroubleReport
}

func GenerateTroubleReportPDF(tr *shared.TroubleReport) (*bytes.Buffer, error) {
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
	err := addTroubleReportImagesSection(o)
	if err != nil {
		return nil, fmt.Errorf("failed to add images to PDF: %w", err)
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF output: %w", err)
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

	for line := range strings.SplitSeq(content, "\n") {
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
			o.PDF.Cell(10, 6, o.Translator("•"))
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

		// Handle blockquotes
		if strings.HasPrefix(line, "> ") {
			o.PDF.SetFont("Arial", "I", 10)
			o.PDF.SetTextColor(100, 100, 100)
			o.PDF.Cell(5, 6, o.Translator("│"))
			o.PDF.MultiCell(0, 6, o.Translator(strings.TrimSpace(line[2:])), "", "", false)
			o.PDF.SetTextColor(0, 0, 0)
			continue
		}

		// Handle regular paragraphs with basic formatting
		o.PDF.SetFont("Arial", "", 10)
		formattedLine := renderBasicMarkdownFormatting(line)
		o.PDF.MultiCell(0, 6, o.Translator(formattedLine), "", "", false)
		o.PDF.Ln(1)
	}
}

// renderBasicMarkdownFormatting processes markdown formatting for PDF rendering
// Note: gofpdf has limited styling support, so we render styled elements where possible
func renderBasicMarkdownFormatting(text string) string {
	// Handle bold: **text** or __text__ (must have content between)
	text = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`__(.+?)__`).ReplaceAllString(text, "$1")

	// Handle italic: *text* - use word boundary to avoid matching single asterisks
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(text, "$1")

	// Remove strikethrough
	text = regexp.MustCompile(`~~(.+?)~~`).ReplaceAllString(text, "$1")

	// Remove inline code formatting - keep the code content
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "$1")

	// Remove link formatting, keep text
	text = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`).ReplaceAllString(text, "$1")

	return text
}

func addTroubleReportImagesSection(o *troubleReportOptions) error {
	if len(o.Report.LinkedAttachments) == 0 {
		return nil
	}

	images, err := getTroubleReportImages(o.Report.LinkedAttachments)
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}

	o.PDF.AddPage()
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10,
		o.Translator(fmt.Sprintf("BILDER (%d)", len(images))),
		"1", 1, "L", true, 0, "")

	renderTroubleReportImagesInGrid(o, images)

	return nil
}

func getTroubleReportImages(attachments []string) ([]*shared.Image, error) {
	var images []*shared.Image
	for _, a := range attachments {
		i := shared.NewImage(a, nil)
		if err := i.ReadFile(); err != nil {
			return images, err
		}
		images = append(images, i)
	}
	return images, nil
}

func renderTroubleReportImagesInGrid(o *troubleReportOptions, images []*shared.Image) {
	pageWidth, _ := o.PDF.GetPageSize()
	leftMargin, _, rightMargin, _ := o.PDF.GetMargins()
	usableWidth := pageWidth - leftMargin - rightMargin
	imageWidth := (usableWidth - 10) / 2 // 10mm spacing between images

	layout := &imageLayoutProps{
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
