package services

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/knackwurstking/pg-press/models"
)

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
