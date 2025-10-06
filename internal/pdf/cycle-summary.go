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

// addCycleSummaryTable adds the summarized cycles table grouped by tool
func addCycleSummaryTable(o *cycleSummaryOptions) {
	o.PDF.SetFont("Arial", "B", 14)
	o.PDF.SetFillColor(240, 248, 255)
	o.PDF.CellFormat(0, 10, o.Translator("WERKZEUG-ÜBERSICHT"), "1", 1, "L", true, 0, "")
	o.PDF.Ln(5)

	// Group cycles by date range
	type toolInfo struct {
		toolID       int64
		toolCode     string
		position     models.Position
		maxCycles    int64
		totalPartial int64
	}

	type dateRangeSummary struct {
		startDate    time.Time
		endDate      time.Time
		tools        []*toolInfo
		maxCycles    int64
		totalPartial int64
	}

	toolSummaries := make(map[int64]*toolInfo)
	dateRangeSummaries := make(map[string]*dateRangeSummary)

	// First pass: group cycles by tool
	for _, cycle := range o.Cycles {
		if summary, exists := toolSummaries[cycle.ToolID]; exists {
			// Update existing summary
			if cycle.TotalCycles > summary.maxCycles {
				summary.maxCycles = cycle.TotalCycles
			}
			summary.totalPartial += cycle.PartialCycles
		} else {
			// Create new summary
			toolCode := "Unbekannt"
			if tool, exists := o.ToolsMap[cycle.ToolID]; exists && tool != nil {
				toolCode = fmt.Sprintf("%s %s", tool.Format.String(), tool.Code)
			}

			toolSummaries[cycle.ToolID] = &toolInfo{
				toolID:       cycle.ToolID,
				toolCode:     toolCode,
				position:     cycle.ToolPosition,
				maxCycles:    cycle.TotalCycles,
				totalPartial: cycle.PartialCycles,
			}
		}
	}

	// Second pass: find date ranges for each tool
	toolDateRanges := make(map[int64]struct{ start, end time.Time })
	for _, cycle := range o.Cycles {
		if existing, exists := toolDateRanges[cycle.ToolID]; exists {
			start := existing.start
			end := existing.end
			if cycle.Date.Before(start) {
				start = cycle.Date
			}
			if cycle.Date.After(end) {
				end = cycle.Date
			}
			toolDateRanges[cycle.ToolID] = struct{ start, end time.Time }{start, end}
		} else {
			toolDateRanges[cycle.ToolID] = struct{ start, end time.Time }{cycle.Date, cycle.Date}
		}
	}

	// Third pass: group tools by identical date ranges
	for toolID, tool := range toolSummaries {
		dateRange := toolDateRanges[toolID]
		dateKey := fmt.Sprintf("%s-%s", dateRange.start.Format("2006-01-02"), dateRange.end.Format("2006-01-02"))

		if dateSummary, exists := dateRangeSummaries[dateKey]; exists {
			// Add tool to existing date range
			dateSummary.tools = append(dateSummary.tools, tool)
			if tool.maxCycles > dateSummary.maxCycles {
				dateSummary.maxCycles = tool.maxCycles
			}
			dateSummary.totalPartial += tool.totalPartial
		} else {
			// Create new date range summary
			dateRangeSummaries[dateKey] = &dateRangeSummary{
				startDate:    dateRange.start,
				endDate:      dateRange.end,
				tools:        []*toolInfo{tool},
				maxCycles:    tool.maxCycles,
				totalPartial: tool.totalPartial,
			}
		}
	}

	// Sort tools within each date range group by position
	for _, dateSummary := range dateRangeSummaries {
		// Sort tools by position order
		for i := 0; i < len(dateSummary.tools)-1; i++ {
			for j := i + 1; j < len(dateSummary.tools); j++ {
				iOrder := getPositionOrder(dateSummary.tools[i].position)
				jOrder := getPositionOrder(dateSummary.tools[j].position)
				if iOrder > jOrder {
					dateSummary.tools[i], dateSummary.tools[j] = dateSummary.tools[j], dateSummary.tools[i]
				}
			}
		}
	}

	// Table headers
	o.PDF.SetFont("Arial", "B", 10)
	o.PDF.SetFillColor(220, 220, 220)

	colWidths := []float64{40, 22, 28, 28, 22, 30}
	headers := []string{"Werkzeug", "Position", "Start Datum", "End Datum", "Zyklen", "Teil-Zyklen"}

	for i, header := range headers {
		o.PDF.CellFormat(colWidths[i], 8, o.Translator(header), "1", 0, "C", true, 0, "")
	}
	o.PDF.Ln(8)

	// Sort date range summaries by max cycles from low to high
	var summaries []*dateRangeSummary
	for _, summary := range dateRangeSummaries {
		summaries = append(summaries, summary)
	}

	// Sort by total press cycles from low to high
	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[i].maxCycles > summaries[j].maxCycles {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	// Table data
	o.PDF.SetFont("Arial", "", 9)
	o.PDF.SetFillColor(250, 250, 250)

	for i, summary := range summaries {
		fill := i%2 == 0 // Alternate row colors

		// Create separate rows for each tool in the group
		for j, tool := range summary.tools {
			// Determine border style based on position in group
			var border string
			if len(summary.tools) == 1 {
				// Single tool - all borders
				border = "1"
			} else if j == 0 {
				// First tool in group - top, left, right borders
				border = "TLR"
			} else if j == len(summary.tools)-1 {
				// Last tool in group - bottom, left, right borders
				border = "BLR"
			} else {
				// Middle tools in group - left, right borders only
				border = "LR"
			}

			// Tool code with format
			o.PDF.CellFormat(colWidths[0], 6, o.Translator(tool.toolCode), border, 0, "C", fill, 0, "")

			// Position
			o.PDF.CellFormat(colWidths[1], 6, o.Translator(tool.position.GermanString()), border, 0, "C", fill, 0, "")

			// Show date and cycle info on the middle row of each group for centering
			middleIndex := len(summary.tools) / 2
			if j == middleIndex {
				// Start date
				startDateStr := summary.startDate.Format("02.01.06")
				o.PDF.CellFormat(colWidths[2], 6, startDateStr, border, 0, "C", fill, 0, "")

				// End date
				endDateStr := summary.endDate.Format("02.01.06")
				o.PDF.CellFormat(colWidths[3], 6, endDateStr, border, 0, "C", fill, 0, "")

				// Max cycles
				o.PDF.CellFormat(colWidths[4], 6, fmt.Sprintf("%d", summary.maxCycles), border, 0, "C", fill, 0, "")

				// Total partial cycles
				o.PDF.CellFormat(colWidths[5], 6, fmt.Sprintf("%d", summary.totalPartial), border, 0, "C", fill, 0, "")
			} else {
				// Empty cells for subsequent tools in the same date group
				o.PDF.CellFormat(colWidths[2], 6, "", border, 0, "C", fill, 0, "")
				o.PDF.CellFormat(colWidths[3], 6, "", border, 0, "C", fill, 0, "")
				o.PDF.CellFormat(colWidths[4], 6, "", border, 0, "C", fill, 0, "")
				o.PDF.CellFormat(colWidths[5], 6, "", border, 0, "C", fill, 0, "")
			}

			o.PDF.Ln(6)

			// Add new page if needed
			_, y := o.PDF.GetXY()
			if y > 250 {
				o.PDF.AddPage()
			}
		}
	}
}

// getPositionOrder returns the sort order for a position
func getPositionOrder(position models.Position) int {
	switch position {
	case models.PositionTop:
		return 1
	case models.PositionTopCassette:
		return 2
	case models.PositionBottom:
		return 3
	default:
		return 999
	}
}
