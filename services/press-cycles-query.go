package services

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

// GetLastToolCycle retrieves the most recent cycle for a specific tool
func (s *PressCycles) GetLastToolCycle(toolID models.ToolID) (*models.Cycle, *errors.MasterError) {
	row := s.DB.QueryRow(SQLGetLastToolCycle, toolID)
	cycle, err := ScanCycle(row)
	if err != nil {
		return cycle, errors.NewMasterError(err, 0)
	}
	cycle.PartialCycles = s.GetPartialCycles(cycle)
	return cycle, nil
}

// GetPressCyclesForTool retrieves all cycles for a specific tool
func (s *PressCycles) ListPressCyclesForTool(toolID models.ToolID) ([]*models.Cycle, *errors.MasterError) {
	rows, err := s.DB.Query(SQLListPressCyclesForTool, toolID)
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
//
// For all press cycles just set limit to -1 and offset to 0
func (s *PressCycles) ListPressCyclesByPress(
	pressNumber models.PressNumber, limit int, offset int,
) ([]*models.Cycle, *errors.MasterError) {

	rows, err := s.DB.Query(SQLListPressCyclesByPress, pressNumber, limit, offset)
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
