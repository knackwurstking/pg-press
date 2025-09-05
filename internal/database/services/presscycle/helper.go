package presscycle

import (
	"database/sql"
	"fmt"

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

// GetPressCyclesForTool gets all press cycles for a specific tool
func (h *Helper) GetPressCyclesForTool(toolID int64) ([]*models.PressCycle, error) {
	logger.DBPressCycles().Debug("Getting press cycles for tool: tool_id=%d", toolID)

	query := `
		SELECT id, press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by
		FROM press_cycles
		WHERE slot_top = ? OR slot_top_cassette = ? OR slot_bottom = ?
		ORDER BY id DESC
	`

	rows, err := h.pressCycles.db.Query(query, toolID, toolID, toolID)
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
		SELECT id, press_number, slot_top, slot_top_cassette, slot_bottom, total_cycles, date, performed_by
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

func (h *Helper) GetPartialCyclesForPress(cycle *models.PressCycle) int64 {
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
