package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNamePressCycles = "press_cycles"

type PressCycles struct {
	*Base
}

func NewPressCycles(r *Registry) *PressCycles {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL CHECK(press_number >= 0 AND press_number <= 5),
			tool_id INTEGER NOT NULL,
			tool_position TEXT NOT NULL,
			total_cycles INTEGER NOT NULL DEFAULT 0,
			date DATETIME NOT NULL,
			performed_by INTEGER NOT NULL,
			FOREIGN KEY (tool_id) REFERENCES %[2]s(id),
			FOREIGN KEY (performed_by) REFERENCES %[3]s(telegram_id) ON DELETE SET NULL
		);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_tool_id ON %[1]s(tool_id);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_tool_position ON %[1]s(tool_position);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_press_number ON %[1]s(press_number);
	`, TableNamePressCycles, TableNameTools, TableNameUsers)

	if err := base.CreateTable(query, TableNamePressCycles); err != nil {
		panic(err)
	}

	return &PressCycles{Base: base}
}

// CRUD Operations

// Get retrieves a press cycle by ID
func (s *PressCycles) Get(id models.CycleID) (*models.Cycle, error) {
	slog.Debug("Getting press cycle", "id", id)

	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE id = ?
	`, TableNamePressCycles)

	row := s.DB.QueryRow(query, id)
	cycle, err := ScanSingleRow(row, scanCycle)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("Press cycle with ID %d not found", id))
		}
		return nil, err
	}

	cycle.PartialCycles = s.GetPartialCycles(cycle)
	return cycle, nil
}

// Cycle Calculations

// GetPartialCycles calculates the partial cycles for a given cycle
func (s *PressCycles) GetPartialCycles(cycle *models.Cycle) int64 {
	if err := cycle.Validate(); err != nil {
		slog.Error("Invalid cycle for partial calculation", "error", err)
		return cycle.TotalCycles
	}

	query := s.buildPartialCyclesQuery(cycle.ToolPosition)
	args := s.buildPartialCyclesArgs(cycle)

	var previousTotalCycles int64
	if err := s.DB.QueryRow(query, args...).Scan(&previousTotalCycles); err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Failed to get previous total cycles", "error", err)
		}
		return cycle.TotalCycles
	}

	return cycle.TotalCycles - previousTotalCycles
}

// Query Methods

// GetLastToolCycle retrieves the most recent cycle for a specific tool
func (s *PressCycles) GetLastToolCycle(toolID models.ToolID) (*models.Cycle, error) {
	slog.Debug("Getting last press cycle for tool", "tool", toolID)

	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY date DESC
		LIMIT 1
	`, TableNamePressCycles)

	row := s.DB.QueryRow(query, toolID)
	cycle, err := ScanSingleRow(row, scanCycle)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("No cycles found for tool %d", toolID))
		}
		return nil, err
	}

	return cycle, nil
}

// GetPressCyclesForTool retrieves all cycles for a specific tool
func (s *PressCycles) GetPressCyclesForTool(toolID models.ToolID) ([]*models.Cycle, error) {
	slog.Debug("Getting press cycles for tool", "tool", toolID)

	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY date DESC
	`, TableNamePressCycles)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	cycles, err := ScanRows(rows, scanCycle)
	if err != nil {
		return nil, err
	}

	cycles = s.injectPartialCycles(cycles)

	return cycles, nil
}

// GetPressCycles retrieves cycles for a specific press with optional pagination
func (s *PressCycles) GetPressCycles(pressNumber models.PressNumber, limit *int, offset *int) ([]*models.Cycle, error) {
	slog.Debug("Getting press cycles for press",
		"press", pressNumber, "limit", limit, "offset", offset)

	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE press_number = ?
		ORDER BY total_cycles DESC
	`, TableNamePressCycles)

	args := []any{pressNumber}
	query = s.addPaginationToQuery(query, limit, offset, &args)

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	cycles, err := ScanRows(rows, scanCycle)
	if err != nil {
		return nil, err
	}

	cycles = s.injectPartialCycles(cycles)

	return cycles, nil
}

// Summary Data

// GetCycleSummaryData retrieves complete cycle summary data for a press
func (s *PressCycles) GetCycleSummaryData(
	pressNumber models.PressNumber,
) ([]*models.Cycle, map[models.ToolID]*models.Tool, map[models.TelegramID]*models.User, error) {
	slog.Debug("Getting cycle summary data for press", "press", pressNumber)

	cycles, err := s.GetPressCycles(pressNumber, nil, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get cycles for press %d: %w", pressNumber, err)
	}

	tools, err := s.Registry.Tools.List()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get tools: %w", err)
	}

	toolsMap := make(map[models.ToolID]*models.Tool, len(tools))
	for _, tool := range tools {
		toolsMap[tool.ID] = tool
	}

	users, err := s.Registry.Users.List()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get users: %w", err)
	}

	usersMap := make(map[models.TelegramID]*models.User, len(users))
	for _, u := range users {
		usersMap[u.TelegramID] = u
	}

	return cycles, toolsMap, usersMap, nil
}

// GetCycleSummaryStats calculates statistics from cycles data
func (s *PressCycles) GetCycleSummaryStats(cycles []*models.Cycle) (
	totalCycles, totalPartialCycles, activeToolsCount, entriesCount int64,
) {
	slog.Debug("Calculating cycle summary stats")

	if cycles == nil {
		slog.Error("Cannot calculate stats from nil cycles data")
		return 0, 0, 0, 0
	}

	activeTools := make(map[models.ToolID]bool)

	for _, cycle := range cycles {
		if cycle.TotalCycles > totalCycles {
			totalCycles = cycle.TotalCycles
		}
		totalPartialCycles += cycle.PartialCycles
		activeTools[cycle.ToolID] = true
	}

	return totalCycles, totalPartialCycles, int64(len(activeTools)), int64(len(cycles))
}

// Tool Summaries

// GetToolSummaries creates consolidated tool summaries with start/end dates
func (s *PressCycles) GetToolSummaries(
	cycles []*models.Cycle, toolsMap map[models.ToolID]*models.Tool,
) ([]*models.ToolSummary, error) {
	slog.Debug("Creating tool summaries from cycles data")

	toolSummaries := s.createInitialSummaries(cycles, toolsMap)
	s.sortToolSummariesChronologically(toolSummaries)
	consolidatedSummaries := s.consolidateToolSummaries(toolSummaries)
	s.fixToolSummaryStartDates(consolidatedSummaries)
	s.sortToolSummariesByCycles(consolidatedSummaries)

	return consolidatedSummaries, nil
}

// List retrieves all press cycles
func (s *PressCycles) List() ([]*models.Cycle, error) {
	slog.Debug("Listing press cycles")

	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		ORDER BY total_cycles DESC
	`, TableNamePressCycles)

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	cycles, err := ScanRows(rows, scanCycle)
	if err != nil {
		return nil, err
	}

	cycles = s.injectPartialCycles(cycles)
	return cycles, nil
}

// Add creates a new press cycle
func (s *PressCycles) Add(cycle *models.Cycle, user *models.User) (models.CycleID, error) {
	slog.Debug(
		"Adding press cycle",
		"user_name", user.Name,
		"tool_id", cycle.ToolID,
		"cycle.ToolPosition", cycle.ToolPosition,
		"cycle.TotalCycles", cycle.TotalCycles,
	)

	if err := cycle.Validate(); err != nil {
		return 0, err
	}

	if err := user.Validate(); err != nil {
		return 0, err
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (press_number, tool_id, tool_position, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, TableNamePressCycles)

	result, err := s.DB.Exec(query,
		cycle.PressNumber,
		cycle.ToolID,
		cycle.ToolPosition,
		cycle.TotalCycles,
		cycle.Date,
		user.TelegramID,
	)
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	cycle.ID = models.CycleID(id)
	return cycle.ID, nil
}

// Update modifies an existing press cycle
func (s *PressCycles) Update(cycle *models.Cycle, user *models.User) error {
	slog.Debug("Updating press cycle", "user_name", user.Name, "cycle.ID", cycle.ID)

	if err := cycle.Validate(); err != nil {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET total_cycles = ?, tool_id = ?, tool_position = ?, performed_by = ?, press_number = ?, date = ?
		WHERE id = ?
	`, TableNamePressCycles)

	_, err := s.DB.Exec(query,
		cycle.TotalCycles,
		cycle.ToolID,
		cycle.ToolPosition,
		user.TelegramID,
		cycle.PressNumber,
		cycle.Date,
		cycle.ID,
	)

	return s.GetUpdateError(err)
}

// Delete removes a press cycle from the database
func (s *PressCycles) Delete(id models.CycleID) error {
	slog.Debug("Deleting press cycle", "cycle", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNamePressCycles)
	_, err := s.DB.Exec(query, id)

	return s.GetDeleteError(err)
}

func (s *PressCycles) buildPartialCyclesQuery(position models.Position) string {
	baseQuery := `
		SELECT total_cycles
		FROM %s
		WHERE press_number = ? AND tool_id > 0 AND %s AND total_cycles < ?
		ORDER BY total_cycles DESC
		LIMIT 1
	`

	condition := "tool_position = ?"
	if position == models.PositionTopCassette {
		condition = "(tool_position = ? OR tool_position = ?)"
	}

	return fmt.Sprintf(baseQuery, TableNamePressCycles, condition)
}

func (s *PressCycles) buildPartialCyclesArgs(cycle *models.Cycle) []any {
	args := []any{cycle.PressNumber}

	if cycle.ToolPosition == models.PositionTopCassette {
		args = append(args, models.PositionTopCassette, models.PositionTop)
	} else {
		args = append(args, cycle.ToolPosition)
	}

	args = append(args, cycle.TotalCycles)
	return args
}

func (s *PressCycles) addPaginationToQuery(query string, limit *int, offset *int, args *[]any) string {
	if limit != nil {
		query += " LIMIT ?"
		*args = append(*args, *limit)
	}
	if offset != nil {
		if limit == nil {
			query += " LIMIT -1"
		}
		query += " OFFSET ?"
		*args = append(*args, *offset)
	}
	return query
}

func (s *PressCycles) createInitialSummaries(cycles []*models.Cycle, toolsMap map[models.ToolID]*models.Tool) []*models.ToolSummary {
	summaries := make([]*models.ToolSummary, 0, len(cycles))

	for _, cycle := range cycles {
		toolCode := s.formatToolCode(cycle.ToolID, toolsMap)

		summaries = append(summaries, &models.ToolSummary{
			ToolID:            cycle.ToolID,
			ToolCode:          toolCode,
			Position:          cycle.ToolPosition,
			StartDate:         cycle.Date,
			EndDate:           cycle.Date,
			MaxCycles:         cycle.TotalCycles,
			TotalPartial:      cycle.PartialCycles,
			IsFirstAppearance: false,
		})
	}

	return summaries
}

func (s *PressCycles) formatToolCode(toolID models.ToolID, toolsMap map[models.ToolID]*models.Tool) string {
	if tool, exists := toolsMap[toolID]; exists && tool != nil {
		return fmt.Sprintf("%s %s", tool.Format.String(), tool.Code)
	}
	return fmt.Sprintf("Tool ID %d", toolID)
}

func (s *PressCycles) sortToolSummariesChronologically(summaries []*models.ToolSummary) {
	sort.Slice(summaries, func(i, j int) bool {
		if !summaries[i].StartDate.Equal(summaries[j].StartDate) {
			return summaries[i].StartDate.Before(summaries[j].StartDate)
		}

		posOrderI := s.getPositionOrder(summaries[i].Position)
		posOrderJ := s.getPositionOrder(summaries[j].Position)
		if posOrderI != posOrderJ {
			return posOrderI < posOrderJ
		}

		return summaries[i].EndDate.Before(summaries[j].EndDate)
	})
}

func (s *PressCycles) consolidateToolSummaries(summaries []*models.ToolSummary) []*models.ToolSummary {
	if len(summaries) == 0 {
		return summaries
	}

	consolidated := make([]*models.ToolSummary, 0)
	lastToolByPosition := make(map[models.Position]models.ToolID)
	positionIndexMap := make(map[models.Position]int)

	for _, summary := range summaries {
		if lastToolID, exists := lastToolByPosition[summary.Position]; exists && lastToolID == summary.ToolID {
			s.mergeIntoExistingSummary(consolidated[positionIndexMap[summary.Position]], summary)
		} else {
			newSummary := s.createNewSummary(summary)
			consolidated = append(consolidated, newSummary)
			lastToolByPosition[summary.Position] = summary.ToolID
			positionIndexMap[summary.Position] = len(consolidated) - 1
		}
	}

	return consolidated
}

func (s *PressCycles) mergeIntoExistingSummary(existing, new *models.ToolSummary) {
	if new.StartDate.Before(existing.StartDate) {
		existing.StartDate = new.StartDate
	}
	if new.EndDate.After(existing.EndDate) {
		existing.EndDate = new.EndDate
	}
	if new.MaxCycles > existing.MaxCycles {
		existing.MaxCycles = new.MaxCycles
	}
	existing.TotalPartial += new.TotalPartial
}

func (s *PressCycles) createNewSummary(source *models.ToolSummary) *models.ToolSummary {
	return &models.ToolSummary{
		ToolID:            source.ToolID,
		ToolCode:          source.ToolCode,
		Position:          source.Position,
		StartDate:         source.StartDate,
		EndDate:           source.EndDate,
		MaxCycles:         source.MaxCycles,
		TotalPartial:      source.TotalPartial,
		IsFirstAppearance: false,
	}
}

func (s *PressCycles) fixToolSummaryStartDates(summaries []*models.ToolSummary) {
	positionEntries := s.groupSummariesByPosition(summaries)

	for _, entries := range positionEntries {
		s.sortEntriesByStartDate(entries)
		s.adjustStartDates(entries)
	}
}

func (s *PressCycles) groupSummariesByPosition(summaries []*models.ToolSummary) map[models.Position][]*models.ToolSummary {
	positionEntries := make(map[models.Position][]*models.ToolSummary)
	for _, summary := range summaries {
		positionEntries[summary.Position] = append(positionEntries[summary.Position], summary)
	}
	return positionEntries
}

func (s *PressCycles) sortEntriesByStartDate(entries []*models.ToolSummary) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartDate.Before(entries[j].StartDate)
	})
}

func (s *PressCycles) adjustStartDates(entries []*models.ToolSummary) {
	for i, entry := range entries {
		if i == 0 {
			entry.IsFirstAppearance = true
		} else {
			entry.StartDate = entries[i-1].EndDate
			entry.IsFirstAppearance = false
		}
	}
}

func (s *PressCycles) sortToolSummariesByCycles(summaries []*models.ToolSummary) {
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].MaxCycles != summaries[j].MaxCycles {
			return summaries[i].MaxCycles < summaries[j].MaxCycles
		}
		return s.getPositionOrder(summaries[i].Position) < s.getPositionOrder(summaries[j].Position)
	})
}

func (s *PressCycles) getPositionOrder(position models.Position) int {
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

// Overlapping Tools Detection

// GetOverlappingTools detects tools that appear on multiple presses during overlapping time periods
func (s *PressCycles) GetOverlappingTools() ([]*models.OverlappingTool, error) {
	slog.Debug("Detecting overlapping tools across all presses")

	validPresses := []models.PressNumber{0, 2, 3, 4, 5}
	allToolSummaries := s.collectAllPressSummaries(validPresses)
	toolGroups := s.groupSummariesByToolID(allToolSummaries)

	return s.findOverlappingTools(toolGroups), nil
}

func (s *PressCycles) collectAllPressSummaries(presses []models.PressNumber) map[models.PressNumber][]*models.ToolSummary {
	allSummaries := make(map[models.PressNumber][]*models.ToolSummary)

	for _, press := range presses {
		cycles, toolsMap, _, err := s.GetCycleSummaryData(press)
		if err != nil {
			slog.Error("Failed to get cycle summary data for press", "press", press, "error", err)
			continue
		}

		summaries, err := s.GetToolSummaries(cycles, toolsMap)
		if err != nil {
			slog.Error("Failed to get tool summaries for press", "press", press, "error", err)
			continue
		}

		allSummaries[press] = summaries
	}

	return allSummaries
}

func (s *PressCycles) groupSummariesByToolID(allSummaries map[models.PressNumber][]*models.ToolSummary) map[models.ToolID]map[models.PressNumber][]*models.ToolSummary {
	toolGroups := make(map[models.ToolID]map[models.PressNumber][]*models.ToolSummary)

	for press, summaries := range allSummaries {
		for _, summary := range summaries {
			if toolGroups[summary.ToolID] == nil {
				toolGroups[summary.ToolID] = make(map[models.PressNumber][]*models.ToolSummary)
			}
			toolGroups[summary.ToolID][press] = append(toolGroups[summary.ToolID][press], summary)
		}
	}

	return toolGroups
}

func (s *PressCycles) findOverlappingTools(toolGroups map[models.ToolID]map[models.PressNumber][]*models.ToolSummary) []*models.OverlappingTool {
	var overlappingTools []*models.OverlappingTool

	for toolID, pressSummaries := range toolGroups {
		if len(pressSummaries) < 2 {
			continue
		}

		if overlappingTool := s.checkForOverlaps(toolID, pressSummaries); overlappingTool != nil {
			overlappingTools = append(overlappingTools, overlappingTool)
		}
	}

	return overlappingTools
}

func (s *PressCycles) checkForOverlaps(toolID models.ToolID, pressSummaries map[models.PressNumber][]*models.ToolSummary) *models.OverlappingTool {
	presses := s.extractPresses(pressSummaries)
	overlaps, toolCode, startDate, endDate := s.findAllOverlaps(presses, pressSummaries, toolID)

	if len(overlaps) == 0 {
		return nil
	}

	toolCode = s.appendPositionsToToolCode(toolCode, overlaps)

	return &models.OverlappingTool{
		ToolID:    toolID,
		ToolCode:  toolCode,
		Overlaps:  overlaps,
		StartDate: startDate,
		EndDate:   endDate,
	}
}

func (s *PressCycles) extractPresses(pressSummaries map[models.PressNumber][]*models.ToolSummary) []models.PressNumber {
	presses := make([]models.PressNumber, 0, len(pressSummaries))
	for press := range pressSummaries {
		presses = append(presses, press)
	}
	return presses
}

func (s *PressCycles) findAllOverlaps(
	presses []models.PressNumber,
	pressSummaries map[models.PressNumber][]*models.ToolSummary,
	toolID models.ToolID,
) ([]*models.OverlappingToolInstance, string, time.Time, time.Time) {
	var overlaps []*models.OverlappingToolInstance
	var overallStartDate, overallEndDate time.Time
	toolCode := fmt.Sprintf("Tool ID %d", toolID)

	for i, press1 := range presses {
		for _, summary1 := range pressSummaries[press1] {
			toolCode = s.updateToolCode(toolCode, summary1.ToolCode, toolID)
			overallStartDate, overallEndDate = s.updateDateRange(overallStartDate, overallEndDate, summary1)

			for j := i + 1; j < len(presses); j++ {
				press2 := presses[j]
				s.checkPressPairOverlaps(summary1, press1, pressSummaries[press2], press2, &overlaps)
			}
		}
	}

	return overlaps, toolCode, overallStartDate, overallEndDate
}

func (s *PressCycles) updateToolCode(current, candidate string, toolID models.ToolID) string {
	defaultCode := fmt.Sprintf("Tool ID %d", toolID)
	if candidate != "" && candidate != defaultCode {
		return candidate
	}
	return current
}

func (s *PressCycles) updateDateRange(startDate, endDate time.Time, summary *models.ToolSummary) (time.Time, time.Time) {
	if startDate.IsZero() || summary.StartDate.Before(startDate) {
		startDate = summary.StartDate
	}
	if endDate.IsZero() || summary.EndDate.After(endDate) {
		endDate = summary.EndDate
	}
	return startDate, endDate
}

func (s *PressCycles) checkPressPairOverlaps(
	summary1 *models.ToolSummary,
	press1 models.PressNumber,
	summaries2 []*models.ToolSummary,
	press2 models.PressNumber,
	overlaps *[]*models.OverlappingToolInstance,
) {
	for _, summary2 := range summaries2 {
		if s.timePeriodsOverlap(summary1.StartDate, summary1.EndDate, summary2.StartDate, summary2.EndDate) {
			s.addOverlapInstances(overlaps, summary1, press1, summary2, press2)
		}
	}
}

func (s *PressCycles) addOverlapInstances(
	overlaps *[]*models.OverlappingToolInstance,
	summary1 *models.ToolSummary,
	press1 models.PressNumber,
	summary2 *models.ToolSummary,
	press2 models.PressNumber,
) {
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

	if !s.containsInstance(*overlaps, instance1) {
		*overlaps = append(*overlaps, instance1)
	}
	if !s.containsInstance(*overlaps, instance2) {
		*overlaps = append(*overlaps, instance2)
	}
}

func (s *PressCycles) appendPositionsToToolCode(toolCode string, overlaps []*models.OverlappingToolInstance) string {
	positions := s.collectUniquePositions(overlaps)

	if len(positions) > 0 {
		return fmt.Sprintf("%s (%s)", toolCode, strings.Join(positions, ", "))
	}

	return toolCode
}

func (s *PressCycles) collectUniquePositions(overlaps []*models.OverlappingToolInstance) []string {
	positionSet := make(map[string]bool)
	var positions []string

	for _, instance := range overlaps {
		pos := instance.Position.GermanString()
		if !positionSet[pos] {
			positionSet[pos] = true
			positions = append(positions, pos)
		}
	}

	return positions
}

func (s *PressCycles) timePeriodsOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}

func (s *PressCycles) containsInstance(instances []*models.OverlappingToolInstance, target *models.OverlappingToolInstance) bool {
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

func (s *PressCycles) injectPartialCycles(cycles []*models.Cycle) []*models.Cycle {
	for _, cycle := range cycles {
		cycle.PartialCycles = s.GetPartialCycles(cycle)
	}

	return cycles
}

// Scan Functions

func scanCycle(scannable Scannable) (*models.Cycle, error) {
	cycle := &models.Cycle{}
	var performedBy sql.NullInt64

	err := scannable.Scan(
		&cycle.ID,
		&cycle.PressNumber,
		&cycle.ToolID,
		&cycle.ToolPosition,
		&cycle.TotalCycles,
		&cycle.Date,
		&performedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("scan press cycle: %w", err)
	}

	if performedBy.Valid {
		cycle.PerformedBy = models.TelegramID(performedBy.Int64)
	}

	return cycle, nil
}
