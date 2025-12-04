package services

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pg-press/models"
)

// GetPartialCycles calculates the partial cycles for a given cycle
//
// This function checks for press regenerations that might have occurred after the current cycle,
// which would reset the cycle count, and adjusts the calculation accordingly.
func (s *PressCycles) GetPartialCycles(cycle *models.Cycle) int64 {
	if err := cycle.Validate(); err != nil {
		return cycle.TotalCycles
	}

	query := s.buildPartialCyclesQuery(cycle.ToolPosition)
	args := s.buildPartialCyclesArgs(cycle)

	var previousTotalCycles int64
	if err := s.DB.QueryRow(query, args...).Scan(&previousTotalCycles); err != nil {
		if err != sql.ErrNoRows {
		}
		return cycle.TotalCycles
	}

	// Check if there were any press regenerations for this press that occurred after the previous cycle
	// but before or at the current cycle date. This would reset the cycle count, so we should
	// return the current cycle's total cycles directly.
	regenerationQuery := `
		SELECT COUNT(*) 
		FROM press_regenerations 
		WHERE press_number = :press_number AND started_at > (
			SELECT date 
			FROM press_cycles 
			WHERE total_cycles = :previous_total_cycles AND press_number = :press_number LIMIT 1
		) AND started_at <= :cycle_date
	`
	var regenerationCount int64
	err := s.DB.QueryRow(
		regenerationQuery,
		sql.Named("press_number", cycle.PressNumber),
		sql.Named("previous_total_cycles", previousTotalCycles),
		sql.Named("cycle_date", cycle.Date),
	).Scan(&regenerationCount)
	if err == nil && regenerationCount > 0 {
		return cycle.TotalCycles
	}

	return cycle.TotalCycles - previousTotalCycles
}

func (s *PressCycles) buildPartialCyclesQuery(position models.Position) string {
	baseQuery := `
		SELECT total_cycles
		FROM %s
		WHERE press_number = ? AND %s AND date < ? AND total_cycles < ?
		ORDER BY date DESC
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

	args = append(args, cycle.Date, cycle.TotalCycles)
	return args
}

func (s *PressCycles) injectPartialCycles(cycles []*models.Cycle) {
	for _, cycle := range cycles {
		cycle.PartialCycles = s.GetPartialCycles(cycle)
	}
}
