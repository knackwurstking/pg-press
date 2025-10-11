package presscycles

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/entities/users"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
}

// ToolsService defines the interface for tools service methods used by PressCycles
type ToolsService interface {
	ListWithNotes() ([]*models.ToolWithNotes, error)
}

func NewService(db *sql.DB) *Service {
	baseService := base.NewBaseService(db, "Press Cycles")

	query := `
		CREATE TABLE IF NOT EXISTS press_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			tool_position TEXT NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES tools(id),
			FOREIGN KEY (performed_by) REFERENCES users(telegram_id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_id ON press_cycles(tool_id);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_tool_position ON press_cycles(tool_position);
		CREATE INDEX IF NOT EXISTS idx_press_cycles_press_number ON press_cycles(press_number);
	`

	if err := baseService.CreateTable(query, "press_cycles"); err != nil {
		panic(err)
	}

	return &Service{
		BaseService: baseService,
	}
}

// GetPartialCycles calculates the partial cycles for a given cycle
func (s *Service) GetPartialCycles(cycle *models.Cycle) int64 {
	if err := ValidatePressCycle(cycle); err != nil {
		s.Log.Error("Invalid cycle for partial calculation: %v", err)
		return cycle.TotalCycles
	}

	s.Log.Debug("Calculating partial cycles: press: %d, tool: %d, position: %s, total: %d",
		cycle.PressNumber, cycle.ToolID, cycle.ToolPosition, cycle.TotalCycles)

	query := `
		SELECT total_cycles
		FROM press_cycles
		WHERE press_number = ? AND tool_id > 0 AND tool_position = ? AND total_cycles < ?
		ORDER BY total_cycles DESC
		LIMIT 1
	`

	var previousTotalCycles int64
	err := s.DB.QueryRow(query, cycle.PressNumber, cycle.ToolPosition, cycle.TotalCycles).Scan(&previousTotalCycles)
	if err != nil {
		if err != sql.ErrNoRows {
			s.Log.Error("Failed to get previous total cycles: %v", err)
		}
		s.Log.Debug("No previous cycles found, using total cycles: %d", cycle.TotalCycles)
		return cycle.TotalCycles
	}

	partialCycles := cycle.TotalCycles - previousTotalCycles
	s.Log.Debug("Calculated partial cycles: %d (total: %d - previous: %d)",
		partialCycles, cycle.TotalCycles, previousTotalCycles)

	return partialCycles
}

// Get retrieves a specific press cycle by its ID.
func (p *Service) Get(id int64) (*models.Cycle, error) {
	if err := validation.ValidateID(id, "press_cycle"); err != nil {
		return nil, err
	}

	p.Log.Debug("Getting press cycle: %d", id)

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		WHERE id = ?
	`

	row := p.DB.QueryRow(query, id)
	cycle, err := scanner.ScanSingleRow(row, ScanPressCycle, "press_cycles")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("Press cycle with ID %d not found", id))
		}
		return nil, err
	}

	cycle.PartialCycles = p.GetPartialCycles(cycle)
	return cycle, nil
}

// List retrieves all press cycles from the database, ordered by total cycles descending.
func (p *Service) List() ([]*models.Cycle, error) {
	p.Log.Debug("Listing press cycles")

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		ORDER BY total_cycles DESC
	`

	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, p.HandleSelectError(err, "press_cycles")
	}
	defer rows.Close()

	cycles, err := p.scanPressCyclesRows(rows)
	if err != nil {
		return nil, err
	}

	return cycles, nil
}

// Add creates a new press cycle entry in the database.
func (p *Service) Add(cycle *models.Cycle, user *models.User) (int64, error) {
	if err := ValidatePressCycle(cycle); err != nil {
		return 0, err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	p.Log.Debug("Adding press cycle by %s (%d): tool: %d, position: %s, press: %d, cycles: %d",
		user.Name, user.TelegramID, cycle.ToolID, cycle.ToolPosition, cycle.PressNumber, cycle.TotalCycles)

	query := `
		INSERT INTO press_cycles (press_number, tool_id, tool_position, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := p.DB.Exec(query,
		cycle.PressNumber,
		cycle.ToolID,
		cycle.ToolPosition,
		cycle.TotalCycles,
		cycle.Date,
		user.TelegramID,
	)
	if err != nil {
		return 0, p.HandleInsertError(err, "press_cycles")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, p.HandleInsertError(err, "press_cycles")
	}

	cycle.ID = id
	return id, nil
}

// Update modifies an existing press cycle entry.
func (p *Service) Update(cycle *models.Cycle, user *models.User) error {
	if err := ValidatePressCycle(cycle); err != nil {
		return err
	}

	if err := validation.ValidateID(cycle.ID, "press_cycle"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	p.Log.Debug("Updating press cycle by %s (%d): id: %d", user.Name, user.TelegramID, cycle.ID)

	query := `
		UPDATE press_cycles
		SET total_cycles = ?, tool_id = ?, tool_position = ?, performed_by = ?, press_number = ?, date = ?
		WHERE id = ?
	`

	result, err := p.DB.Exec(query,
		cycle.TotalCycles,
		cycle.ToolID,
		cycle.ToolPosition,
		user.TelegramID,
		cycle.PressNumber,
		cycle.Date,
		cycle.ID,
	)
	if err != nil {
		return p.HandleUpdateError(err, "press_cycles")
	}

	if err := p.CheckRowsAffected(result, "press_cycle", cycle.ID); err != nil {
		return err
	}

	return nil
}

// Delete removes a press cycle from the database.
func (p *Service) Delete(id int64) error {
	if err := validation.ValidateID(id, "press_cycle"); err != nil {
		return err
	}

	p.Log.Debug("Deleting press cycle: %d", id)

	query := `DELETE FROM press_cycles WHERE id = ?`
	result, err := p.DB.Exec(query, id)
	if err != nil {
		return p.HandleDeleteError(err, "press_cycles")
	}

	if err := p.CheckRowsAffected(result, "press_cycle", id); err != nil {
		return err
	}

	return nil
}

// GetPressCyclesForTool gets all press cycles for a specific tool
func (s *Service) GetPressCyclesForTool(toolID int64) ([]*models.Cycle, error) {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting press cycles for tool: %d", toolID)

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY date DESC
	`

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, s.HandleSelectError(err, "press_cycles")
	}
	defer rows.Close()

	cycles, err := s.scanPressCyclesRows(rows)
	if err != nil {
		return nil, err
	}

	return cycles, nil
}

// GetPressCycles gets all press cycles for a specific press with optional pagination
func (s *Service) GetPressCycles(pressNumber models.PressNumber, limit *int, offset *int) ([]*models.Cycle, error) {
	if err := ValidatePressNumber(pressNumber); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting press cycles for press: %d, limit: %v, offset: %v",
		pressNumber, limit, offset)

	query := `
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY total_cycles DESC
	`

	var queryArgs []any
	queryArgs = append(queryArgs, pressNumber)

	if limit != nil {
		query += " LIMIT ?"
		queryArgs = append(queryArgs, *limit)
	}
	if offset != nil {
		if limit == nil {
			query += " LIMIT -1"
		}
		query += " OFFSET ?"
		queryArgs = append(queryArgs, *offset)
	}

	rows, err := s.DB.Query(query, queryArgs...)
	if err != nil {
		return nil, s.HandleSelectError(err, "press_cycles")
	}
	defer rows.Close()

	cycles, err := s.scanPressCyclesRows(rows)
	if err != nil {
		return nil, err
	}

	return cycles, nil
}

// scanPressCyclesRows scans multiple press cycles from sql.Rows and calculates partial cycles
func (p *Service) scanPressCyclesRows(rows *sql.Rows) ([]*models.Cycle, error) {
	cycles, err := ScanPressCyclesFromRows(rows)
	if err != nil {
		return nil, err
	}

	// Calculate partial cycles for each cycle
	for _, cycle := range cycles {
		cycle.PartialCycles = p.GetPartialCycles(cycle)
	}

	return cycles, nil
}

// ToolSummary represents a summary of a tool's usage during a specific period
type ToolSummary struct {
	ToolID            int64           `json:"tool_id"`
	ToolCode          string          `json:"tool_code"`
	Position          models.Position `json:"position"`
	StartDate         time.Time       `json:"start_date"`
	EndDate           time.Time       `json:"end_date"`
	MaxCycles         int64           `json:"max_cycles"`
	TotalPartial      int64           `json:"total_partial"`
	IsFirstAppearance bool            `json:"is_first_appearance"`
}

// GetCycleSummaryData retrieves complete cycle summary data for a press.
// This is a convenience method that gathers all data needed for cycle summaries:
// - All cycles for the specified press
// - Tools map for looking up tool information by ID
// - Users map for looking up user information by telegram ID
//
// Example usage:
//
//	cycles, toolsMap, usersMap, err := pressCyclesService.GetCycleSummaryData(press, toolsService, usersService)
//	if err != nil {
//	    return err
//	}
//	// Use cycles, toolsMap, and usersMap for PDF generation, reports, etc.
func (s *Service) GetCycleSummaryData(pressNumber models.PressNumber, toolsService ToolsService, usersService *users.Service) ([]*models.Cycle, map[int64]*models.Tool, map[int64]*models.User, error) {
	if err := ValidatePressNumber(pressNumber); err != nil {
		return nil, nil, nil, err
	}

	if err := validation.ValidateNotNil(toolsService, "tools service"); err != nil {
		return nil, nil, nil, err
	}

	if err := validation.ValidateNotNil(usersService, "users service"); err != nil {
		return nil, nil, nil, err
	}

	s.Log.Debug("Getting cycle summary data for press: %d", pressNumber)

	// Get cycles for this press
	cycles, err := s.GetPressCycles(pressNumber, nil, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get cycles for press %d: %v", pressNumber, err)
	}

	// Get tools to create toolsMap
	tools, err := toolsService.ListWithNotes()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get tools: %v", err)
	}

	toolsMap := make(map[int64]*models.Tool)
	for _, toolWithNotes := range tools {
		tool := toolWithNotes.Tool
		toolsMap[tool.ID] = tool
	}

	// Get users for performed_by names
	users, err := usersService.List()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get users: %v", err)
	}

	usersMap := make(map[int64]*models.User)
	for _, u := range users {
		usersMap[u.TelegramID] = u
	}

	return cycles, toolsMap, usersMap, nil
}

// GetCycleSummaryStats calculates statistics from cycles data.
// Returns: totalCycles, totalPartialCycles, activeToolsCount, entriesCount
//
// Example usage:
//
//	totalCycles, totalPartial, activeTools, entries := pressCyclesService.GetCycleSummaryStats(cycles)
//	fmt.Printf("Total cycles: %d, Active tools: %d, Entries: %d", totalCycles, activeTools, entries)
func (s *Service) GetCycleSummaryStats(cycles []*models.Cycle) (int64, int64, int64, int64) {
	if cycles == nil {
		s.Log.Error("Cannot calculate stats from nil cycles data")
		return 0, 0, 0, 0
	}

	totalCycles := int64(0)
	totalPartialCycles := int64(0)
	activeTools := make(map[int64]bool)

	for _, cycle := range cycles {
		if cycle.TotalCycles > totalCycles {
			totalCycles = cycle.TotalCycles
		}
		totalPartialCycles += cycle.PartialCycles
		activeTools[cycle.ToolID] = true
	}

	activeToolsCount := int64(len(activeTools))
	entriesCount := int64(len(cycles))

	return totalCycles, totalPartialCycles, activeToolsCount, entriesCount
}

// GetToolSummaries creates consolidated tool summaries with start/end dates.
// This method consolidates consecutive entries for the same tool in the same position
// and calculates proper start/end dates for tool usage periods.
//
// The returned summaries are sorted by cycle count (low to high) and then by position.
//
// Example usage:
//
//	cycles, toolsMap, _, err := pressCyclesService.GetCycleSummaryData(press, toolsService, usersService)
//	if err != nil {
//	    return err
//	}
//	summaries, err := pressCyclesService.GetToolSummaries(cycles, toolsMap)
//	if err != nil {
//	    return err
//	}
//	// Use summaries for detailed tool usage reports
func (s *Service) GetToolSummaries(cycles []*models.Cycle, toolsMap map[int64]*models.Tool) ([]*ToolSummary, error) {
	if cycles == nil {
		return nil, fmt.Errorf("cannot create tool summaries from nil cycles data")
	}

	s.Log.Debug("Creating tool summaries from cycles data")

	var toolSummaries []*ToolSummary

	// Create a summary for each individual cycle
	for _, cycle := range cycles {
		// Create tool code string - handle missing tools gracefully
		toolCode := fmt.Sprintf("Tool ID %d", cycle.ToolID)
		if tool, exists := toolsMap[cycle.ToolID]; exists && tool != nil {
			toolCode = fmt.Sprintf("%s %s", tool.Format.String(), tool.Code)
		}

		toolSummaries = append(toolSummaries, &ToolSummary{
			ToolID:            cycle.ToolID,
			ToolCode:          toolCode,
			Position:          cycle.ToolPosition,
			StartDate:         cycle.Date,
			EndDate:           cycle.Date,
			MaxCycles:         cycle.TotalCycles,
			TotalPartial:      cycle.PartialCycles,
			IsFirstAppearance: false, // Will be set during consolidation
		})
	}

	// Sort chronologically for proper consolidation
	s.sortToolSummariesChronologically(toolSummaries)

	// Consolidate consecutive entries for the same tool in the same position
	consolidatedSummaries := s.consolidateToolSummaries(toolSummaries)

	// Fix start dates based on tool changes per position
	s.fixToolSummaryStartDates(consolidatedSummaries)

	// Sort by cycle count and position
	s.sortToolSummariesByCycles(consolidatedSummaries)

	return consolidatedSummaries, nil
}

// sortToolSummariesChronologically sorts tool summaries by date and position
func (s *Service) sortToolSummariesChronologically(summaries []*ToolSummary) {
	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[i].StartDate.After(summaries[j].StartDate) ||
				(summaries[i].StartDate.Equal(summaries[j].StartDate) && s.getPositionOrder(summaries[i].Position) > s.getPositionOrder(summaries[j].Position)) ||
				(summaries[i].StartDate.Equal(summaries[j].StartDate) && summaries[i].Position == summaries[j].Position && summaries[i].EndDate.After(summaries[j].EndDate)) {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}
}

// consolidateToolSummaries consolidates consecutive entries for the same tool in the same position
func (s *Service) consolidateToolSummaries(summaries []*ToolSummary) []*ToolSummary {
	var consolidatedSummaries []*ToolSummary
	lastToolByPosition := make(map[models.Position]int64)
	positionIndexMap := make(map[models.Position]int)

	for _, summary := range summaries {
		lastToolID, hasLastTool := lastToolByPosition[summary.Position]

		// Check if this is the same tool as the last one in this position
		if hasLastTool && lastToolID == summary.ToolID {
			// Consolidate with the existing entry for this position
			existingIndex := positionIndexMap[summary.Position]
			existingSummary := consolidatedSummaries[existingIndex]

			// Extend the date range
			if summary.StartDate.Before(existingSummary.StartDate) {
				existingSummary.StartDate = summary.StartDate
			}
			if summary.EndDate.After(existingSummary.EndDate) {
				existingSummary.EndDate = summary.EndDate
			}

			// Take highest total cycles
			if summary.MaxCycles > existingSummary.MaxCycles {
				existingSummary.MaxCycles = summary.MaxCycles
			}

			// Sum partial cycles
			existingSummary.TotalPartial += summary.TotalPartial
		} else {
			// Create new entry
			newSummary := &ToolSummary{
				ToolID:            summary.ToolID,
				ToolCode:          summary.ToolCode,
				Position:          summary.Position,
				StartDate:         summary.StartDate,
				EndDate:           summary.EndDate,
				MaxCycles:         summary.MaxCycles,
				TotalPartial:      summary.TotalPartial,
				IsFirstAppearance: false, // Will be set in fixToolSummaryStartDates
			}

			consolidatedSummaries = append(consolidatedSummaries, newSummary)

			// Update tracking maps
			lastToolByPosition[summary.Position] = summary.ToolID
			positionIndexMap[summary.Position] = len(consolidatedSummaries) - 1
		}
	}

	return consolidatedSummaries
}

// fixToolSummaryStartDates fixes start dates based on tool changes per position
func (s *Service) fixToolSummaryStartDates(summaries []*ToolSummary) {
	positionEntries := make(map[models.Position][]*ToolSummary)
	for _, summary := range summaries {
		positionEntries[summary.Position] = append(positionEntries[summary.Position], summary)
	}

	// For each position, sort by start date and fix start dates
	for _, entries := range positionEntries {
		// Sort entries by original start date
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].StartDate.After(entries[j].StartDate) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		// Fix start dates: first entry is unknown, others start when previous ended
		for i, entry := range entries {
			if i == 0 {
				// First tool in this position - unknown start date
				entry.IsFirstAppearance = true
			} else {
				// Tool replacement - starts when previous tool ended
				entry.StartDate = entries[i-1].EndDate
				entry.IsFirstAppearance = false
			}
		}
	}
}

// sortToolSummariesByCycles sorts tool summaries by cycle count and then by position
func (s *Service) sortToolSummariesByCycles(summaries []*ToolSummary) {
	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			// Primary sort: by cycle count
			if summaries[i].MaxCycles > summaries[j].MaxCycles {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			} else if summaries[i].MaxCycles == summaries[j].MaxCycles {
				// Secondary sort: by position (top, top cassette, bottom)
				if s.getPositionOrder(summaries[i].Position) > s.getPositionOrder(summaries[j].Position) {
					summaries[i], summaries[j] = summaries[j], summaries[i]
				}
			}
		}
	}
}

// getPositionOrder returns the sort order for a position
func (s *Service) getPositionOrder(position models.Position) int {
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

// GetOverlappingTools detects tools that appear on multiple presses during overlapping time periods
func (s *Service) GetOverlappingTools(
	toolsService ToolsService, usersService *users.Service,
) ([]*models.OverlappingTool, error) {
	if err := validation.ValidateNotNil(toolsService, "tools service"); err != nil {
		return nil, err
	}
	if err := validation.ValidateNotNil(usersService, "users service"); err != nil {
		return nil, err
	}

	s.Log.Debug("Detecting overlapping tools across all presses")

	// Valid press numbers
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}

	// Get all tool summaries for all presses
	allToolSummaries := make(map[models.PressNumber][]*ToolSummary)

	for _, press := range validPresses {
		cycles, toolsMap, _, err := s.GetCycleSummaryData(press, toolsService, usersService)
		if err != nil {
			s.Log.Error("Failed to get cycle summary data for press: %d, error: %v", press, err)
			continue // Skip this press and continue with others
		}

		summaries, err := s.GetToolSummaries(cycles, toolsMap)
		if err != nil {
			s.Log.Error("Failed to get tool summaries for press: %d, error: %v", press, err)
			continue // Skip this press and continue with others
		}

		allToolSummaries[press] = summaries
	}

	// Group summaries by tool ID
	toolGroups := make(map[int64]map[models.PressNumber][]*ToolSummary)
	for press, summaries := range allToolSummaries {
		for _, summary := range summaries {
			if toolGroups[summary.ToolID] == nil {
				toolGroups[summary.ToolID] = make(map[models.PressNumber][]*ToolSummary)
			}
			toolGroups[summary.ToolID][press] = append(toolGroups[summary.ToolID][press], summary)
		}
	}

	var overlappingTools []*models.OverlappingTool

	// Check each tool for overlaps across presses
	for toolID, pressSummaries := range toolGroups {
		// Only check tools that appear on multiple presses
		if len(pressSummaries) < 2 {
			continue
		}

		// Get all press combinations
		presses := make([]models.PressNumber, 0, len(pressSummaries))
		for press := range pressSummaries {
			presses = append(presses, press)
		}

		// Check for overlaps between all press pairs
		var overlaps []*models.OverlappingToolInstance
		overallStartDate := time.Time{}
		overallEndDate := time.Time{}
		toolCode := fmt.Sprintf("Tool ID %d", toolID)

		for i, press1 := range presses {
			for _, summary1 := range pressSummaries[press1] {
				// Update tool code if we have it
				if summary1.ToolCode != "" && summary1.ToolCode != fmt.Sprintf("Tool ID %d", toolID) {
					toolCode = summary1.ToolCode
				}

				// Track overall date range
				if overallStartDate.IsZero() || summary1.StartDate.Before(overallStartDate) {
					overallStartDate = summary1.StartDate
				}
				if overallEndDate.IsZero() || summary1.EndDate.After(overallEndDate) {
					overallEndDate = summary1.EndDate
				}

				for j := i + 1; j < len(presses); j++ {
					press2 := presses[j]
					for _, summary2 := range pressSummaries[press2] {
						// Check if time periods overlap
						if s.timePeriodsOverlap(summary1.StartDate, summary1.EndDate, summary2.StartDate, summary2.EndDate) {
							// Add both instances if not already added
							instance1 := &models.OverlappingToolInstance{
								PressNumber: press1,
								Position:    summary1.Position,
								StartDate:   summary1.StartDate,
								EndDate:     summary1.EndDate,
							}
							instance2 := &models.OverlappingToolInstance{
								PressNumber: press2,
								Position:    summary2.Position,
								StartDate:   summary2.StartDate,
								EndDate:     summary2.EndDate,
							}

							// Check if we already have these instances
							if !s.containsInstance(overlaps, instance1) {
								overlaps = append(overlaps, instance1)
							}
							if !s.containsInstance(overlaps, instance2) {
								overlaps = append(overlaps, instance2)
							}
						}
					}
				}
			}
		}

		// If we found overlaps, create the overlapping tool entry
		if len(overlaps) > 0 {
			// Collect unique positions and append to toolCode
			positions := make(map[string]bool)
			var positionList []string
			for _, instance := range overlaps {
				pos := instance.Position.GermanString()
				if !positions[pos] {
					positions[pos] = true
					positionList = append(positionList, pos)
				}
			}

			// Append positions to toolCode
			if len(positionList) > 0 {
				positionsStr := ""
				for i, pos := range positionList {
					if i > 0 {
						positionsStr += ", "
					}
					positionsStr += pos
				}
				toolCode = fmt.Sprintf("%s (%s)", toolCode, positionsStr)
			}

			overlappingTool := &models.OverlappingTool{
				ToolID:    toolID,
				ToolCode:  toolCode,
				Overlaps:  overlaps,
				StartDate: overallStartDate,
				EndDate:   overallEndDate,
			}
			overlappingTools = append(overlappingTools, overlappingTool)
		}
	}

	return overlappingTools, nil
}

// timePeriodsOverlap checks if two time periods overlap
func (s *Service) timePeriodsOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

// containsInstance checks if an instance is already in the slice
func (s *Service) containsInstance(
	instances []*models.OverlappingToolInstance, target *models.OverlappingToolInstance,
) bool {
	for _, instance := range instances {
		if instance.PressNumber == target.PressNumber &&
			instance.Position == target.Position &&
			instance.StartDate.Equal(target.StartDate) &&
			instance.EndDate.Equal(target.EndDate) {
			return true
		}
	}
	return false
}
