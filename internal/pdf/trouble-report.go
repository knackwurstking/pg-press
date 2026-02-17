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
			renderFormattedLine(o, strings.TrimSpace(line[4:]))
			o.PDF.Ln(2)
			continue
		}
		if strings.HasPrefix(line, "## ") {
			o.PDF.SetFont("Arial", "B", 12)
			renderFormattedLine(o, strings.TrimSpace(line[3:]))
			o.PDF.Ln(2)
			continue
		}
		if strings.HasPrefix(line, "# ") {
			o.PDF.SetFont("Arial", "B", 13)
			renderFormattedLine(o, strings.TrimSpace(line[2:]))
			o.PDF.Ln(3)
			continue
		}

		// Handle lists
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			o.PDF.SetFont("Arial", "", 10)
			o.PDF.Cell(10, 6, o.Translator("•"))
			renderFormattedLine(o, strings.TrimSpace(line[2:]))
			o.PDF.Ln(6)
			continue
		}

		// Handle numbered lists
		if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
			o.PDF.SetFont("Arial", "", 10)
			parts := regexp.MustCompile(`^(\d+\.\s)(.*)`).FindStringSubmatch(line)
			if len(parts) == 3 {
				o.PDF.Cell(10, 6, parts[1])
				renderFormattedLine(o, parts[2])
			}
			o.PDF.Ln(6)
			continue
		}

		// Handle blockquotes
		if strings.HasPrefix(line, "> ") {
			o.PDF.SetFont("Arial", "I", 10)
			o.PDF.SetTextColor(100, 100, 100)
			o.PDF.Cell(5, 6, o.Translator("│"))
			renderFormattedLine(o, strings.TrimSpace(line[2:]))
			o.PDF.SetTextColor(0, 0, 0)
			o.PDF.Ln(6)
			continue
		}

		// Handle regular paragraphs with inline formatting
		o.PDF.SetFont("Arial", "", 10)
		renderFormattedLine(o, line)
		o.PDF.Ln(1)
	}
}

// renderFormattedLine processes inline markdown formatting (bold, italic, strikethrough)
func renderFormattedLine(o *troubleReportOptions, line string) {
	segments := parseMarkdownSegments(line)

	for _, seg := range segments {
		style := ""
		if seg.Bold {
			style += "B"
		}
		if seg.Italic {
			style += "I"
		}
		if seg.Strikethrough {
			style += "S"
		}
		if style == "" {
			style = ""
		}

		fontSize := 10.0
		if seg.Header == 3 {
			fontSize = 11
		} else if seg.Header == 2 {
			fontSize = 12
		} else if seg.Header == 1 {
			fontSize = 13
		}

		o.PDF.SetFont("Arial", style, fontSize)
		o.PDF.Write(fixedLineHeight(fontSize), o.Translator(seg.Text))
	}
}

// markdownSegment represents a text segment with formatting
type markdownSegment struct {
	Text          string
	Bold          bool
	Italic        bool
	Strikethrough bool
	Header        int
}

// parseMarkdownSegments parses a line into formatted segments
func parseMarkdownSegments(line string) []markdownSegment {
	var segments []markdownSegment

	// Regex patterns for inline formatting
	boldPattern := regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicPattern := regexp.MustCompile(`\*(.+?)\*`)

	// Keep track of what has been replaced
	processed := line

	// Find all matches with their positions
	type match struct {
		start   int
		end     int
		text    string
		matched string
		format  string
	}

	var matches []match

	// Find bold matches
	for _, m := range boldPattern.FindAllStringSubmatchIndex(processed, -1) {
		matches = append(matches, match{
			start:   m[0],
			end:     m[1],
			text:    processed[m[2]:m[3]],
			matched: processed[m[0]:m[1]],
			format:  "bold",
		})
	}

	// Find italic matches
	for _, m := range italicPattern.FindAllStringSubmatchIndex(processed, -1) {
		matches = append(matches, match{
			start:   m[0],
			end:     m[1],
			text:    processed[m[2]:m[3]],
			matched: processed[m[0]:m[1]],
			format:  "italic",
		})
	}

	// Sort matches by position
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].start < matches[i].start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Build segments
	lastPos := 0
	for _, m := range matches {
		// Add text before this match
		if m.start > lastPos {
			segments = append(segments, markdownSegment{Text: processed[lastPos:m.start]})
		}

		// Add the formatted text
		seg := markdownSegment{Text: m.text}
		switch m.format {
		case "bold":
			seg.Bold = true
		case "italic":
			seg.Italic = true
		segments = append(segments, seg)

		lastPos = m.end
	}

	// Add remaining text
	if lastPos < len(processed) {
		segments = append(segments, markdownSegment{Text: processed[lastPos:]})
	}

	if len(segments) == 0 {
		segments = append(segments, markdownSegment{Text: line})
	}

	return segments
}

func fixedLineHeight(fontSize float64) float64 {
	return fontSize * 0.4
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
