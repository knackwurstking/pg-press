package database

import (
	"time"
)

// ToolCyclesSummary represents a summary of tool cycles since last regeneration
type ToolCyclesSummary struct {
	ToolID               int64               `json:"tool_id"`
	TotalCycles          int64               `json:"total_cycles"`
	LastRegenerationDate *time.Time          `json:"last_regeneration_date"`
	CurrentUsage         *PressCycle         `json:"current_usage"`
	History              []*PressCycle       `json:"history"`
	RegenerationHistory  []*ToolRegeneration `json:"regeneration_history"`
}

// ToolPressHistory represents the complete press usage history for a tool
type ToolPressHistory struct {
	ToolID              int64                         `json:"tool_id"`
	CurrentPress        *PressNumber                  `json:"current_press"`
	CurrentStatus       ToolStatus                    `json:"current_status"`
	TotalAllTimeCycles  int64                         `json:"total_all_time_cycles"`
	CyclesBetweenRegens []*CyclesBetweenRegenerations `json:"cycles_between_regenerations"`
}

// CyclesBetweenRegenerations represents cycles grouped by regeneration periods
type CyclesBetweenRegenerations struct {
	FromRegeneration *time.Time    `json:"from_regeneration"`
	ToRegeneration   *time.Time    `json:"to_regeneration"`
	TotalCycles      int64         `json:"total_cycles"`
	PressCycles      []*PressCycle `json:"press_cycles"`
}

// ToolCyclesHelper provides helper methods for tool cycles management
type ToolCyclesHelper struct {
	db            *DB
	presses       *Presses
	tools         *Tools
	regenerations *ToolRegenerations
}

// NewToolCyclesHelper creates a new ToolCyclesHelper instance
func NewToolCyclesHelper(db *DB) *ToolCyclesHelper {
	return &ToolCyclesHelper{
		db:            db,
		presses:       db.Presses,
		tools:         db.Tools,
		regenerations: db.ToolRegenerations,
	}
}

// GetToolCyclesSummary gets a comprehensive summary of tool cycles
func (h *ToolCyclesHelper) GetToolCyclesSummary(toolID int64) (*ToolCyclesSummary, error) {
	summary := &ToolCyclesSummary{
		ToolID: toolID,
	}

	// Get last regeneration date
	lastRegen, err := h.regenerations.GetLastRegeneration(toolID)
	if err != nil {
		return nil, err
	}
	if lastRegen != nil {
		summary.LastRegenerationDate = &lastRegen.RegeneratedAt
	}

	// Get current usage
	currentUsage, err := h.presses.GetCurrentToolUsage(toolID)
	if err != nil {
		return nil, err
	}
	summary.CurrentUsage = currentUsage

	// Get history since last regeneration
	history, err := h.presses.GetToolHistorySinceRegeneration(toolID, summary.LastRegenerationDate)
	if err != nil {
		return nil, err
	}
	summary.History = history

	// Calculate total cycles
	totalCycles, err := h.presses.GetTotalCyclesSinceRegeneration(toolID, summary.LastRegenerationDate)
	if err != nil {
		return nil, err
	}
	summary.TotalCycles = totalCycles

	// Get regeneration history
	regenHistory, err := h.regenerations.GetRegenerationHistory(toolID)
	if err != nil {
		return nil, err
	}
	summary.RegenerationHistory = regenHistory

	return summary, nil
}

// GetToolPressHistory gets the complete press usage history for a tool
func (h *ToolCyclesHelper) GetToolPressHistory(toolID int64) (*ToolPressHistory, error) {
	history := &ToolPressHistory{
		ToolID: toolID,
	}

	// Get tool status
	tool, err := h.tools.GetByID(toolID)
	if err != nil {
		return nil, err
	}
	history.CurrentStatus = tool.Status
	history.CurrentPress = tool.Press

	// Get all press cycles
	allCycles, err := h.presses.GetToolHistory(toolID)
	if err != nil {
		return nil, err
	}

	// Calculate total all-time cycles
	for _, cycle := range allCycles {
		history.TotalAllTimeCycles += cycle.TotalCycles
	}

	// Get regeneration history
	regenerations, err := h.regenerations.GetRegenerationHistory(toolID)
	if err != nil {
		return nil, err
	}

	// Group cycles by regeneration periods
	history.CyclesBetweenRegens = h.groupCyclesByRegenerations(allCycles, regenerations)

	return history, nil
}

// groupCyclesByRegenerations groups press cycles by regeneration periods
func (h *ToolCyclesHelper) groupCyclesByRegenerations(cycles []*PressCycle, regenerations []*ToolRegeneration) []*CyclesBetweenRegenerations {
	if len(cycles) == 0 {
		return nil
	}

	var groups []*CyclesBetweenRegenerations

	// Sort regenerations by date (they should already be sorted DESC from query)
	// We need them in ASC order for grouping
	for i := len(regenerations)/2 - 1; i >= 0; i-- {
		opp := len(regenerations) - 1 - i
		regenerations[i], regenerations[opp] = regenerations[opp], regenerations[i]
	}

	// Create groups between regenerations
	var lastRegenTime *time.Time
	for i, regen := range regenerations {
		group := &CyclesBetweenRegenerations{
			FromRegeneration: lastRegenTime,
			ToRegeneration:   &regen.RegeneratedAt,
		}

		// Find cycles in this period
		for _, cycle := range cycles {
			if (lastRegenTime == nil || cycle.FromDate.After(*lastRegenTime)) &&
				cycle.FromDate.Before(regen.RegeneratedAt) {
				group.PressCycles = append(group.PressCycles, cycle)
				group.TotalCycles += cycle.TotalCycles
			}
		}

		if len(group.PressCycles) > 0 {
			groups = append(groups, group)
		}
		lastRegenTime = &regenerations[i].RegeneratedAt
	}

	// Add final group for cycles after last regeneration
	finalGroup := &CyclesBetweenRegenerations{
		FromRegeneration: lastRegenTime,
		ToRegeneration:   nil, // Still ongoing
	}

	for _, cycle := range cycles {
		if lastRegenTime == nil || cycle.FromDate.After(*lastRegenTime) {
			finalGroup.PressCycles = append(finalGroup.PressCycles, cycle)
			finalGroup.TotalCycles += cycle.TotalCycles
		}
	}

	if len(finalGroup.PressCycles) > 0 {
		groups = append(groups, finalGroup)
	}

	// If no regenerations, all cycles are in one group
	if len(regenerations) == 0 && len(cycles) > 0 {
		totalCycles := int64(0)
		for _, cycle := range cycles {
			totalCycles += cycle.TotalCycles
		}
		groups = []*CyclesBetweenRegenerations{
			{
				FromRegeneration: nil, // Since beginning
				ToRegeneration:   nil, // Still ongoing
				TotalCycles:      totalCycles,
				PressCycles:      cycles,
			},
		}
	}

	return groups
}

// StartToolOnPress starts a tool on a specific press
func (h *ToolCyclesHelper) StartToolOnPress(toolID int64, pressNumber int) error {
	// Update tool status to active and set press number
	err := h.tools.UpdateStatus(toolID, ToolStatusActive)
	if err != nil {
		return err
	}

	err = h.tools.UpdatePress(toolID, &pressNumber)
	if err != nil {
		return err
	}

	// Start press cycle tracking
	_, err = h.presses.StartToolUsage(toolID, pressNumber)
	if err != nil {
		return err
	}

	return nil
}

// RemoveToolFromPress removes a tool from its current press
func (h *ToolCyclesHelper) RemoveToolFromPress(toolID int64) error {
	// End press cycle
	err := h.presses.EndToolUsage(toolID)
	if err != nil {
		return err
	}

	// Update tool status to available and clear press number
	err = h.tools.UpdateStatus(toolID, ToolStatusAvailable)
	if err != nil {
		return err
	}

	err = h.tools.UpdatePress(toolID, nil)
	if err != nil {
		return err
	}

	return nil
}

// RegenerateTool marks a tool as regenerated and resets its cycles
func (h *ToolCyclesHelper) RegenerateTool(toolID int64, reason string, performedBy string) error {
	// Update tool status to regenerating
	err := h.tools.UpdateStatus(toolID, ToolStatusRegenerating)
	if err != nil {
		return err
	}

	// Save regeneration event to tracking table
	_, err = h.regenerations.Create(toolID, reason, performedBy, "")
	if err != nil {
		return err
	}

	// Mark regeneration in press cycles (ends current usage)
	err = h.presses.MarkToolRegeneration(toolID)
	if err != nil {
		return err
	}

	return nil
}

// CompleteToolRegeneration marks a tool regeneration as complete
func (h *ToolCyclesHelper) CompleteToolRegeneration(toolID int64) error {
	// Update tool status back to available
	err := h.tools.UpdateStatus(toolID, ToolStatusAvailable)
	if err != nil {
		return err
	}

	// Clear press assignment
	err = h.tools.UpdatePress(toolID, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateToolCycles updates the cycle counts for a tool currently on a press
func (h *ToolCyclesHelper) UpdateToolCycles(toolID int64, totalCycles, partialCycles int64) error {
	return h.presses.UpdateCycles(toolID, totalCycles, partialCycles)
}

// GetPressUtilization gets current tool utilization for all presses
func (h *ToolCyclesHelper) GetPressUtilization() (map[int][]int64, error) {
	utilization := make(map[int][]int64)

	// Iterate through all press numbers (0-5)
	for pressNumber := 0; pressNumber <= 5; pressNumber++ {
		toolIDs, err := h.presses.GetCurrentToolsOnPress(pressNumber)
		if err != nil {
			return nil, err
		}
		utilization[pressNumber] = toolIDs
	}

	return utilization, nil
}
