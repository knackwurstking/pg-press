package presscycle

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/models"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Helper provides additional press cycle-related database operations
// that are not part of the generic DataOperations interface.
type Helper struct {
	pressCycles *Service
}

// NewHelper creates a new PressCyclesHelper instance.
func NewHelper(pressCycles *Service) *Helper {
	return &Helper{
		pressCycles: pressCycles,
	}
}

// StartToolUsage records when a tool starts being used on a press
func (h *Helper) StartToolUsage(toolID int64, pressNumber models.PressNumber, user *models.User) (*models.PressCycle, error) {
	logger.DBPressCycles().Info("Starting tool usage: tool_id=%d, press_number=%d", toolID, pressNumber)

	if !models.IsValidPressNumber(&pressNumber) {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	// Create new press cycle entry
	var performedBy int64
	if user != nil {
		performedBy = user.TelegramID
	}

	query := `
		INSERT INTO press_cycles (press_number, tool_id, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, press_number, tool_id, total_cycles, date, performed_by
	`

	row := h.pressCycles.db.QueryRow(query, pressNumber, toolID, 0, time.Now(), performedBy)
	cycle, err := h.pressCycles.scanPressCycle(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}

		return nil, fmt.Errorf("failed to start tool usage: %w", err)
	}

	// Create feed entry
	if h.pressCycles.feeds != nil {
		h.pressCycles.feeds.Add(models.NewFeed(
			models.FeedTypeToolUpdate,
			&models.FeedToolUpdate{
				ID:         toolID,
				Tool:       fmt.Sprintf("Werkzeug #%d wurde an Presse %d angebracht", toolID, pressNumber),
				ModifiedBy: user,
			},
		))
	}

	return cycle, nil
}

// EndToolUsage is deprecated - we no longer track end dates
// Kept for backward compatibility but does nothing
func (h *Helper) EndToolUsage(toolID int64) error {
	logger.DBPressCycles().Info("EndToolUsage called (deprecated): tool_id=%d", toolID)
	// No-op - we don't track to_date anymore
	return nil
}

// GetCurrentToolUsage gets the current active press cycle for a tool
func (h *Helper) GetCurrentToolUsage(toolID int64) (*models.PressCycle, error) {
	logger.DBPressCycles().Debug("Getting current tool usage: tool_id=%d", toolID)

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	row := h.pressCycles.db.QueryRow(query, toolID)
	cycle, err := h.pressCycles.scanPressCycle(row)
	if err == sql.ErrNoRows {
		return nil, dberror.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current tool usage: %w", err)
	}

	return cycle, nil
}

// GetToolHistory retrieves all press cycles for a specific tool
func (h *Helper) GetToolHistory(toolID int64, limit, offset int) ([]*models.PressCycle, error) {
	logger.DBPressCycles().Debug("Getting tool history: tool_id=%d, limit=%d, offset=%d", toolID, limit, offset)

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.pressCycles.db.Query(query, toolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool history: %w", err)
	}
	defer rows.Close()

	cycles, err := h.pressCycles.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetPressCyclesForTool gets all press cycles for a specific tool
func (h *Helper) GetPressCyclesForTool(toolID int64) ([]*models.PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles for tool: tool_id=%d", toolID)

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
	`

	rows, err := h.pressCycles.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for tool %d: %w", toolID, err)
	}
	defer rows.Close()

	cycles, err := h.pressCycles.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetPressCycles gets all press cycles (current and historical) for a specific press
func (h *Helper) GetPressCycles(pressNumber models.PressNumber, limit, offset int) ([]*models.PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles: press_number=%d, limit=%d, offset=%d", pressNumber, limit, offset)

	if !models.IsValidPressNumber(&pressNumber) {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.pressCycles.db.Query(query, pressNumber, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for press %d: %w", pressNumber, err)
	}
	defer rows.Close()

	cycles, err := h.pressCycles.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetCurrentToolsOnPress gets all tools currently active on a specific press
func (h *Helper) GetCurrentToolsOnPress(pressNumber models.PressNumber) ([]int64, error) {
	logger.DBPressCycles().Debug("Getting current tools on press: press_number=%d", pressNumber)

	if !models.IsValidPressNumber(&pressNumber) {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	query := `
		SELECT tool_id
		FROM (
			SELECT tool_id, press_number,
			       ROW_NUMBER() OVER (PARTITION BY tool_id ORDER BY id DESC) as rn
			FROM press_cycles
		)
		WHERE press_number = ? AND rn = 1
	`

	rows, err := h.pressCycles.db.Query(query, pressNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get current tools on press: %w", err)
	}
	defer rows.Close()

	var toolIDs []int64
	for rows.Next() {
		var toolID int64
		if err := rows.Scan(&toolID); err != nil {
			return nil, fmt.Errorf("failed to scan tool ID: %w", err)
		}
		toolIDs = append(toolIDs, toolID)
	}

	return toolIDs, nil
}

// GetPressUtilization gets current tool count for each press (0-5)
func (h *Helper) GetPressUtilization() (map[models.PressNumber][]int64, error) {
	logger.DBPressCycles().Debug("Getting press utilization for all presses")

	utilization := map[models.PressNumber][]int64{}

	for i := models.PressNumber(0); i <= 5; i++ {
		utilization[i] = []int64{}
	}

	query := `
		SELECT press_number, tool_id
		FROM (
			SELECT press_number, tool_id,
			       ROW_NUMBER() OVER (PARTITION BY tool_id ORDER BY id DESC) as rn
			FROM press_cycles
		)
		WHERE rn = 1
		ORDER BY press_number, tool_id
	`

	rows, err := h.pressCycles.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get press utilization: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber models.PressNumber
		var toolID int64
		if err := rows.Scan(&pressNumber, &toolID); err != nil {
			return nil, fmt.Errorf("failed to scan utilization data: %w", err)
		}
		utilization[pressNumber] = append(utilization[pressNumber], toolID)
	}

	return utilization, nil
}

type PressCycleStats struct {
	TotalCycles    int64
	ActiveTools    int
	TotalToolsUsed int
}

// GetPressCycleStats gets statistics for all presses
func (h *Helper) GetPressCycleStats() (map[models.PressNumber]*PressCycleStats, error) {
	logger.DBPressCycles().Debug("Getting press cycle statistics for all presses")

	stats := make(map[models.PressNumber]*PressCycleStats)

	for i := models.PressNumber(0); i <= 5; i++ {
		stats[i] = &PressCycleStats{}
	}

	query := `
		SELECT
			press_number,
			SUM(total_cycles) as total_cycles,
			COUNT(DISTINCT tool_id) as total_tools_used,
			(SELECT COUNT(DISTINCT tool_id)
			 FROM (
			   SELECT tool_id,
			          ROW_NUMBER() OVER (PARTITION BY tool_id ORDER BY id DESC) as rn
			   FROM press_cycles pc2
			   WHERE pc2.press_number = press_cycles.press_number
			 ) WHERE rn = 1) as active_tools
		FROM press_cycles
		GROUP BY press_number
	`

	rows, err := h.pressCycles.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycle stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber models.PressNumber
		var totalCycles sql.NullInt64
		var totalToolsUsed, activeTools int

		err := rows.Scan(&pressNumber, &totalCycles, &totalToolsUsed, &activeTools)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		stat := stats[pressNumber]

		if totalCycles.Valid {
			stat.TotalCycles = totalCycles.Int64
		}

		stat.TotalToolsUsed = totalToolsUsed
		stat.ActiveTools = activeTools
	}

	return stats, nil
}

// GetCurrentTotalCycles gets the current total cycles for a specific tool.
// It retrieves the `total_cycles` from the most recent press cycle entry for the given tool ID.
func (h *Helper) GetCurrentTotalCycles(toolID int64) (int64, error) {
	logger.DBPressCycles().Debug("Getting current total cycles for tool: tool_id=%d", toolID)

	query := `
		SELECT total_cycles
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	var totalCycles int64
	err := h.pressCycles.db.QueryRow(query, toolID).Scan(&totalCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			// If no rows are found, it means the tool has no cycle entries yet.
			// In this case, the total cycles should be considered 0.
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current total cycles for tool %d: %w", toolID, err)
	}

	return totalCycles, nil
}

// GetTotalCyclesSinceRegeneration calculates the total cycles of a tool since its last regeneration.
func (h *Helper) GetTotalCyclesSinceRegeneration(toolID int64) (int64, error) {
	logger.DBPressCycles().Debug("Getting total cycles since last regeneration for tool: tool_id=%d", toolID)

	// Step 1: Get the cycle_id of the last regeneration for the tool.
	lastRegenCycleIDQuery := `
		SELECT cycle_id
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`
	var lastRegenCycleID int64
	err := h.pressCycles.db.QueryRow(lastRegenCycleIDQuery, toolID).Scan(&lastRegenCycleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No regeneration found, so return all cycles.
			return h.GetCurrentTotalCycles(toolID)
		}
		return 0, fmt.Errorf("failed to get last regeneration cycle for tool %d: %w", toolID, err)
	}

	// Step 2: Get the total cycles at the last regeneration.
	cyclesAtRegenQuery := `
		SELECT total_cycles
		FROM press_cycles
		WHERE id = ?
	`
	var cyclesAtRegen int64
	err = h.pressCycles.db.QueryRow(cyclesAtRegenQuery, lastRegenCycleID).Scan(&cyclesAtRegen)
	if err != nil {
		if err == sql.ErrNoRows {
			// This should not happen if the foreign key is set up correctly.
			return 0, fmt.Errorf("press cycle for regeneration not found for tool %d, cycle %d: %w", toolID, lastRegenCycleID, err)
		}
		return 0, fmt.Errorf("failed to get cycles at regeneration for tool %d: %w", toolID, err)
	}

	// Step 3: Get the current total cycles.
	currentCycles, err := h.GetCurrentTotalCycles(toolID)
	if err != nil {
		return 0, err
	}

	// Step 4: The difference is the cycles since regeneration.
	return currentCycles - cyclesAtRegen, nil
}

func (h *Helper) GetPartialCycles(cycle *models.PressCycle) int64 {
	logger.DBPressCycles().Debug("Getting partial cycles: %#v", cycle)

	// Get the total_cycles from the previous entry on the same press (regardless of tool_id)
	previousQuery := `
		SELECT total_cycles
		FROM press_cycles
		WHERE press_number = ? AND id < ?
		ORDER BY id DESC
		LIMIT 1
	`
	var previousTotalCycles int64
	err := h.pressCycles.db.QueryRow(previousQuery, cycle.PressNumber, cycle.ID).Scan(&previousTotalCycles)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous entry found, so partial cycles equals total cycles
			logger.DBPressCycles().Debug("No previous entry found for press %d, partial cycles = total cycles (%d)", cycle.PressNumber, cycle.TotalCycles)
		} else {
			logger.DBPressCycles().Error("Failed to get previous total cycles for press %d: %v", cycle.PressNumber, err)
		}
		return cycle.TotalCycles
	}

	partialCycles := cycle.TotalCycles - previousTotalCycles
	logger.DBPressCycles().Debug("Partial cycles calculated: %d (current: %d - previous: %d)", partialCycles, cycle.TotalCycles, previousTotalCycles)

	return partialCycles
}
