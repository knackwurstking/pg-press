package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// cycleSummaryOptions contains options for cycle summary PDF generation
type cycleSummaryOptions struct {
	*imageOptions
	Press    models.PressNumber
	Cycles   []*models.Cycle
	ToolsMap map[int64]*models.Tool
	UsersMap map[int64]*models.User
}

// GenerateCycleSummaryPDF creates a PDF with cycle summary data for a press
func GenerateCycleSummaryPDF(
	press models.PressNumber,
	cycles []*models.Cycle,
	toolsMap map[int64]*models.Tool,
	usersMap map[int64]*models.User,
) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 25)
	pdf.AddPage()
	pdf.SetMargins(20, 20, 20)

	o := &cycleSummaryOptions{
		imageOptions: &imageOptions{
			PDF:        pdf,
			Translator: pdf.UnicodeTranslatorFromDescriptor(""),
		},
		Press:    press,
		Cycles:   cycles,
		ToolsMap: toolsMap,
		UsersMap: usersMap,
	}

	addCycleSummaryHeader(o)
	addCycleSummaryStats(o)
	addCycleSummaryTable(o)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

// addCycleSummaryHeader adds the PDF header section
func addCycleSummaryHeader(o *cycleSummaryOptions) {
	o.PDF.SetFont("Arial", "B", 20)
	o.PDF.SetTextColor(0, 51, 102)
	o.PDF.Cell(0, 15, o.Translator(fmt.Sprintf("Zyklen-Zusammenfassung - Presse %d", o.Press)))
	o.PDF.Ln(10)

	o.PDF.SetFont("Arial", "", 12)
	o.PDF.SetTextColor(128, 128, 128)
	o.PDF.Cell(0, 8, o.Translator(fmt.Sprintf("Erstellt am: %s", time.Now().Format("02.01.2006 15:04"))))
	o.PDF.Ln(15)

	o.PDF.SetTextColor(0, 0, 0)
}

// addCycleSummaryStats adds summary statistics
func addCycleSummaryStats(o *cycleSummaryOptions) {
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10, o.Translator("ZUSAMMENFASSUNG"), "1", 1, "L", true, 0, "")
	o.PDF.Ln(5)

	// Calculate statistics
	totalCycles := int64(0)
	totalPartialCycles := int64(0)
	activeTools := make(map[int64]bool)

	for _, cycle := range o.Cycles {
		if cycle.TotalCycles > totalCycles {
			totalCycles = cycle.TotalCycles
		}
		totalPartialCycles += cycle.PartialCycles
		activeTools[cycle.ToolID] = true
	}

	o.PDF.SetFont("Arial", "", 12)
	o.PDF.Cell(0, 8, o.Translator(fmt.Sprintf("Gesamte Zyklen: %d", totalCycles)))
	o.PDF.Ln(6)
	o.PDF.Cell(0, 8, o.Translator(fmt.Sprintf("Teilzyklen insgesamt: %d", totalPartialCycles)))
	o.PDF.Ln(6)
	o.PDF.Cell(0, 8, o.Translator(fmt.Sprintf("Aktive Werkzeuge: %d", len(activeTools))))
	o.PDF.Ln(6)
	o.PDF.Cell(0, 8, o.Translator(fmt.Sprintf("Anzahl Einträge: %d", len(o.Cycles))))
	o.PDF.Ln(15)
}

// addCycleSummaryTable adds the detailed cycles table
func addCycleSummaryTable(o *cycleSummaryOptions) {
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10, o.Translator("ZYKLEN-DETAILS"), "1", 1, "L", true, 0, "")
	o.PDF.Ln(5)

	// Table headers
	o.PDF.SetFont("Arial", "B", 10)
	o.PDF.SetFillColor(220, 220, 220)

	colWidths := []float64{25, 25, 35, 25, 25}
	headers := []string{"Datum", "Zyklen", "Teil-Zyklen", "Werkzeug", "Position"}

	for i, header := range headers {
		o.PDF.CellFormat(colWidths[i], 8, o.Translator(header), "1", 0, "C", true, 0, "")
	}
	o.PDF.Ln(8)

	// Table data
	o.PDF.SetFont("Arial", "", 9)
	o.PDF.SetFillColor(250, 250, 250)

	for i, cycle := range o.Cycles {
		fill := i%2 == 0 // Alternate row colors

		// Date
		dateStr := cycle.Date.Format("02.01.06")
		o.PDF.CellFormat(colWidths[0], 6, dateStr, "1", 0, "C", fill, 0, "")

		// Total cycles
		o.PDF.CellFormat(colWidths[1], 6, fmt.Sprintf("%d", cycle.TotalCycles), "1", 0, "C", fill, 0, "")

		// Partial cycles
		o.PDF.CellFormat(colWidths[2], 6, fmt.Sprintf("%d", cycle.PartialCycles), "1", 0, "C", fill, 0, "")

		// Tool code
		toolCode := "Unbekannt"
		if tool, exists := o.ToolsMap[cycle.ToolID]; exists && tool != nil {
			toolCode = tool.Code
		}
		if len(toolCode) > 12 {
			toolCode = toolCode[:9] + "..."
		}
		o.PDF.CellFormat(colWidths[3], 6, o.Translator(toolCode), "1", 0, "C", fill, 0, "")

		// Position
		o.PDF.CellFormat(colWidths[4], 6, o.Translator(cycle.ToolPosition.GermanString()), "1", 0, "C", fill, 0, "")

		o.PDF.Ln(6)

		// Add new page if needed
		_, y := o.PDF.GetXY()
		if y > 250 {
			o.PDF.AddPage()
		}
	}
}
