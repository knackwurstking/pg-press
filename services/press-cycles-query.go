package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

// GetLastToolCycle retrieves the most recent cycle for a specific tool
func (s *PressCycles) GetLastToolCycle(toolID models.ToolID) (*models.Cycle, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY date DESC
		LIMIT 1
	`, TableNamePressCycles)

	row := s.DB.QueryRow(query, toolID)
	cycle, err := ScanCycle(row)
	if err != nil {
		return cycle, errors.NewMasterError(err, 0)
	}
	cycle.PartialCycles = s.GetPartialCycles(cycle)
	return cycle, nil
}

// GetPressCyclesForTool retrieves all cycles for a specific tool
func (s *PressCycles) GetPressCyclesForTool(toolID models.ToolID) ([]*models.Cycle, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY date DESC
	`, TableNamePressCycles)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	cycles, merr := ScanRows(rows, ScanCycle)
	if merr != nil {
		return nil, merr
	}

	s.injectPartialCycles(cycles)
	return cycles, nil
}

// GetPressCycles retrieves cycles for a specific press with optional pagination
func (s *PressCycles) GetPressCycles(
	pressNumber models.PressNumber, limit *int, offset *int,
) ([]*models.Cycle, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE press_number = ?
		ORDER BY date DESC
	`, TableNamePressCycles)

	args := []any{pressNumber}
	query = s.addPaginationToQuery(query, limit, offset, &args)

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	cycles, merr := ScanRows(rows, ScanCycle)
	if merr != nil {
		return cycles, nil
	}
	s.injectPartialCycles(cycles)
	return cycles, nil
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
