package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/models"
)

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

func (s *PressCycles) injectPartialCycles(cycles []*models.Cycle) []*models.Cycle {
	for _, cycle := range cycles {
		cycle.PartialCycles = s.GetPartialCycles(cycle)
	}

	return cycles
}
