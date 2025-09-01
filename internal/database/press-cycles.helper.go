package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/logger"
)

// PressCyclesHelper provides additional press cycle-related database operations
// that are not part of the generic DataOperations interface.
type PressCyclesHelper struct {
	pressCycles *PressCycles
}

// NewPressCyclesHelper creates a new PressCyclesHelper instance.
func NewPressCyclesHelper(pressCycles *PressCycles) *PressCyclesHelper {
	return &PressCyclesHelper{
		pressCycles: pressCycles,
	}
}

// StartToolUsage records when a tool starts being used on a press
func (pch *PressCyclesHelper) StartToolUsage(toolID int64, pressNumber PressNumber, user *User) (*PressCycle, error) {
	logger.DBPressCycles().Info("Starting tool usage: tool_id=%d, press_number=%d", toolID, pressNumber)

	if !pressNumber.IsValid() {
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

	row := pch.pressCycles.db.QueryRow(query, pressNumber, toolID, 0, time.Now(), performedBy)
	cycle, err := pch.pressCycles.scanPressCycle(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to start tool usage: %w", err)
	}

	// Create feed entry
	if pch.pressCycles.feeds != nil {
		pch.pressCycles.feeds.Add(NewFeed(
			FeedTypeToolUpdate,
			&FeedToolUpdate{
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
func (pch *PressCyclesHelper) EndToolUsage(toolID int64) error {
	logger.DBPressCycles().Info("EndToolUsage called (deprecated): tool_id=%d", toolID)
	// No-op - we don't track to_date anymore
	return nil
}

// GetCurrentToolUsage gets the current active press cycle for a tool
func (pch *PressCyclesHelper) GetCurrentToolUsage(toolID int64) (*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting current tool usage: tool_id=%d", toolID)

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	row := pch.pressCycles.db.QueryRow(query, toolID)
	cycle, err := pch.pressCycles.scanPressCycle(row)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current tool usage: %w", err)
	}

	return cycle, nil
}

// GetToolHistory retrieves all press cycles for a specific tool
func (pch *PressCyclesHelper) GetToolHistory(toolID int64, limit, offset int) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting tool history: tool_id=%d, limit=%d, offset=%d", toolID, limit, offset)

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := pch.pressCycles.db.Query(query, toolID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool history: %w", err)
	}
	defer rows.Close()

	cycles, err := pch.pressCycles.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetPressCyclesForTool gets all press cycles for a specific tool
func (pch *PressCyclesHelper) GetPressCyclesForTool(toolID int64) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles for tool: tool_id=%d", toolID)

	query := `
		SELECT id, press_number, tool_id, total_cycles, 0 as partial_cycles, date, performed_by
		FROM press_cycles
		WHERE tool_id = ?
		ORDER BY id DESC
	`

	rows, err := pch.pressCycles.db.Query(query, toolID)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for tool %d: %w", toolID, err)
	}
	defer rows.Close()

	cycles, err := pch.pressCycles.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetPressCycles gets all press cycles (current and historical) for a specific press
func (pch *PressCyclesHelper) GetPressCycles(pressNumber PressNumber, limit, offset int) ([]*PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles: press_number=%d, limit=%d, offset=%d", pressNumber, limit, offset)

	if !pressNumber.IsValid() {
		return nil, fmt.Errorf("invalid press number %d: must be between 0 and 5", pressNumber)
	}

	query := `
		SELECT id, press_number, tool_id, total_cycles, date, performed_by
		FROM press_cycles
		WHERE press_number = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := pch.pressCycles.db.Query(query, pressNumber, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycles for press %d: %w", pressNumber, err)
	}
	defer rows.Close()

	cycles, err := pch.pressCycles.scanPressCyclesRows(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan press cycles: %w", err)
	}

	return cycles, nil
}

// GetCurrentToolsOnPress gets all tools currently active on a specific press
func (pch *PressCyclesHelper) GetCurrentToolsOnPress(pressNumber PressNumber) ([]int64, error) {
	logger.DBPressCycles().Debug("Getting current tools on press: press_number=%d", pressNumber)

	if !pressNumber.IsValid() {
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

	rows, err := pch.pressCycles.db.Query(query, pressNumber)
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
func (pch *PressCyclesHelper) GetPressUtilization() (map[PressNumber][]int64, error) {
	logger.DBPressCycles().Debug("Getting press utilization for all presses")

	utilization := map[PressNumber][]int64{}

	for i := PressNumber(0); i <= 5; i++ {
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

	rows, err := pch.pressCycles.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get press utilization: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber PressNumber
		var toolID int64
		if err := rows.Scan(&pressNumber, &toolID); err != nil {
			return nil, fmt.Errorf("failed to scan utilization data: %w", err)
		}
		utilization[pressNumber] = append(utilization[pressNumber], toolID)
	}

	return utilization, nil
}

// GetPressCycleStats gets statistics for all presses
func (pch *PressCyclesHelper) GetPressCycleStats() (map[PressNumber]struct {
	TotalCycles    int64
	ActiveTools    int
	TotalToolsUsed int
}, error) {
	logger.DBPressCycles().Debug("Getting press cycle statistics for all presses")

	stats := make(map[PressNumber]struct {
		TotalCycles    int64
		ActiveTools    int
		TotalToolsUsed int
	})

	for i := PressNumber(0); i <= 5; i++ {
		stats[i] = struct {
			TotalCycles    int64
			ActiveTools    int
			TotalToolsUsed int
		}{}
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

	rows, err := pch.pressCycles.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get press cycle stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pressNumber PressNumber
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
		stats[pressNumber] = stat
	}

	return stats, nil
}
