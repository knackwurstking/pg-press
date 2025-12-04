package services

import (
	"fmt"
	"sort"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

// GetCycleSummaryData retrieves complete cycle summary data for a press
func (s *PressCycles) GetCycleSummaryData(
	pressNumber models.PressNumber,
) ([]*models.Cycle, map[models.ToolID]*models.Tool, map[models.TelegramID]*models.User, *errors.DBError) {
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
	if cycles == nil {
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

// GetToolSummaries creates consolidated tool summaries with start/end dates
func (s *PressCycles) GetToolSummaries(
	cycles []*models.Cycle, toolsMap map[models.ToolID]*models.Tool,
) ([]*models.ToolSummary, *errors.DBError) {
	toolSummaries := s.createInitialSummaries(cycles, toolsMap)
	s.sortToolSummariesChronologically(toolSummaries)
	consolidatedSummaries := s.consolidateToolSummaries(toolSummaries)
	s.fixToolSummaryStartDates(consolidatedSummaries)
	s.sortToolSummariesByCycles(consolidatedSummaries)

	return consolidatedSummaries, nil
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

		posOrderI := models.GetPositionOrder(summaries[i].Position)
		posOrderJ := models.GetPositionOrder(summaries[j].Position)
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

		return models.GetPositionOrder(summaries[i].Position) < models.GetPositionOrder(summaries[j].Position)
	})
}
