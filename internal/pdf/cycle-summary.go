package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
	"github.com/knackwurstking/pgpress/internal/constants"
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

	// Create individual cycle summaries without combining by ToolID
	type toolSummary struct {
		toolID            int64
		toolCode          string
		position          models.Position
		startDate         time.Time
		endDate           time.Time
		maxCycles         int64
		totalPartial      int64
		isFirstAppearance bool
	}

	var toolSummaries []*toolSummary

	// Create a summary for each individual cycle
	for _, cycle := range o.Cycles {
		// Create tool code string - handle missing tools gracefully
		toolCode := fmt.Sprintf("Tool ID %d", cycle.ToolID)
		if tool, exists := o.ToolsMap[cycle.ToolID]; exists && tool != nil {
			toolCode = fmt.Sprintf("%s %s", tool.Format.String(), tool.Code)
		}

		toolSummaries = append(toolSummaries, &toolSummary{
			toolID:            cycle.ToolID,
			toolCode:          toolCode,
			position:          cycle.ToolPosition,
			startDate:         cycle.Date,
			endDate:           cycle.Date,
			maxCycles:         cycle.TotalCycles,
			totalPartial:      cycle.PartialCycles,
			isFirstAppearance: false, // Will be set during consolidation
		})
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

	// First sort chronologically for proper consolidation
	for i := 0; i < len(toolSummaries)-1; i++ {
		for j := i + 1; j < len(toolSummaries); j++ {
			if toolSummaries[i].startDate.After(toolSummaries[j].startDate) ||
				(toolSummaries[i].startDate.Equal(toolSummaries[j].startDate) && getPositionOrder(toolSummaries[i].position) > getPositionOrder(toolSummaries[j].position)) ||
				(toolSummaries[i].startDate.Equal(toolSummaries[j].startDate) && toolSummaries[i].position == toolSummaries[j].position && toolSummaries[i].endDate.After(toolSummaries[j].endDate)) {
				toolSummaries[i], toolSummaries[j] = toolSummaries[j], toolSummaries[i]
			}
		}
	}

	// Consolidate consecutive entries for the same tool in the same position
	var consolidatedSummaries []*toolSummary
	lastToolByPosition := make(map[models.Position]int64)
	positionIndexMap := make(map[models.Position]int)

	for _, summary := range toolSummaries {
		lastToolID, hasLastTool := lastToolByPosition[summary.position]

		// Check if this is the same tool as the last one in this position
		if hasLastTool && lastToolID == summary.toolID {
			// Consolidate with the existing entry for this position
			existingIndex := positionIndexMap[summary.position]
			existingSummary := consolidatedSummaries[existingIndex]

			// Extend the date range
			if summary.startDate.Before(existingSummary.startDate) {
				existingSummary.startDate = summary.startDate
			}
			if summary.endDate.After(existingSummary.endDate) {
				existingSummary.endDate = summary.endDate
			}

			// Take highest total cycles
			if summary.maxCycles > existingSummary.maxCycles {
				existingSummary.maxCycles = summary.maxCycles
			}

			// Sum partial cycles
			existingSummary.totalPartial += summary.totalPartial
		} else {
			// Create new entry (start date will be fixed in second pass)
			newSummary := &toolSummary{
				toolID:            summary.toolID,
				toolCode:          summary.toolCode,
				position:          summary.position,
				startDate:         summary.startDate,
				endDate:           summary.endDate,
				maxCycles:         summary.maxCycles,
				totalPartial:      summary.totalPartial,
				isFirstAppearance: false, // Will be set in second pass
			}

			consolidatedSummaries = append(consolidatedSummaries, newSummary)

			// Update tracking maps
			lastToolByPosition[summary.position] = summary.toolID
			positionIndexMap[summary.position] = len(consolidatedSummaries) - 1
		}
	}

	// Second pass: Fix start dates based on tool changes per position
	positionEntries := make(map[models.Position][]*toolSummary)
	for _, summary := range consolidatedSummaries {
		positionEntries[summary.position] = append(positionEntries[summary.position], summary)
	}

	// For each position, sort by start date and fix start dates
	for _, entries := range positionEntries {
		// Sort entries by original start date
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].startDate.After(entries[j].startDate) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Fix start dates: first entry is unknown, others start when previous ended
		for i, entry := range entries {
			if i == 0 {
				// First tool in this position - unknown start date
				entry.isFirstAppearance = true
			} else {
				// Tool replacement - starts when previous tool ended
				entry.startDate = entries[i-1].endDate
				entry.isFirstAppearance = false
			}
		}
	}

	toolSummaries = consolidatedSummaries

	// Sort by total cycle count from low to high, then by position
	for i := 0; i < len(toolSummaries)-1; i++ {
		for j := i + 1; j < len(toolSummaries); j++ {
			// Primary sort: by cycle count
			if toolSummaries[i].maxCycles > toolSummaries[j].maxCycles {
				toolSummaries[i], toolSummaries[j] = toolSummaries[j], toolSummaries[i]
			} else if toolSummaries[i].maxCycles == toolSummaries[j].maxCycles {
				// Secondary sort: by position (top, top cassette, bottom)
				if getPositionOrder(toolSummaries[i].position) > getPositionOrder(toolSummaries[j].position) {
					toolSummaries[i], toolSummaries[j] = toolSummaries[j], toolSummaries[i]
				}
			}
		}
	}

	// Table data
	o.PDF.SetFont("Arial", "", 9)

	// Group tools by total cycle count for highlighting
	var currentCycleCount int64 = -1
	var cycleGroupIndex int

	for _, summary := range toolSummaries {
		// Check if this is a new cycle count group
		if summary.maxCycles != currentCycleCount {
			currentCycleCount = summary.maxCycles
			cycleGroupIndex++
		}

		// Use different background colors for different cycle count groups
		fill := cycleGroupIndex%2 == 0
		if fill {
			o.PDF.SetFillColor(240, 248, 255) // Light blue for even groups
		} else {
			o.PDF.SetFillColor(255, 248, 240) // Light orange for odd groups
		}

		// Tool code with format
		o.PDF.CellFormat(colWidths[0], 6, o.Translator(summary.toolCode), "1", 0, "C", fill, 0, "")

		// Position
		o.PDF.CellFormat(colWidths[1], 6, o.Translator(summary.position.GermanString()), "1", 0, "C", fill, 0, "")

		// Start date - show empty for first appearance of tool
		var startDateStr string
		if summary.isFirstAppearance {
			startDateStr = ""
		} else {
			startDateStr = summary.startDate.Format(constants.DateFormat)
		}

		o.PDF.CellFormat(colWidths[2], 6, startDateStr, "1", 0, "C", fill, 0, "")

		// End date
		endDateStr := summary.endDate.Format(constants.DateFormat)
		o.PDF.CellFormat(colWidths[3], 6, endDateStr, "1", 0, "C", fill, 0, "")

		// Max cycles
		o.PDF.CellFormat(colWidths[4], 6, fmt.Sprintf("%d", summary.maxCycles), "1", 0, "C", fill, 0, "")

		// Total partial cycles
		o.PDF.CellFormat(colWidths[5], 6, fmt.Sprintf("%d", summary.totalPartial), "1", 0, "C", fill, 0, "")

		o.PDF.Ln(6)

		// Add new page if needed
		_, y := o.PDF.GetXY()
		if y > 250 {
			o.PDF.AddPage()
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
