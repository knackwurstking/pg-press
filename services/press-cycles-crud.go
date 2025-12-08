package services

import (
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

// Get retrieves a press cycle by ID
func (s *PressCycles) Get(id models.CycleID) (*models.Cycle, *errors.MasterError) {
	row := s.DB.QueryRow(SQLGetPressCycle, id)
	cycle, err := ScanCycle(row)
	if err != nil {
		return cycle, errors.NewMasterError(err, 0)
	}

	cycle.PartialCycles = s.GetPartialCycles(cycle)
	return cycle, nil
}

// Add creates a new press cycle
func (s *PressCycles) Add(
	press models.PressNumber,
	toolID models.ToolID,
	toolPosition models.Position,
	totalCycles int64,
	performedBy models.TelegramID,
) (models.CycleID, *errors.MasterError) {

	cycle := models.NewCycle(
		press, toolID, toolPosition, totalCycles, performedBy,
	)

	verr := cycle.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	result, err := s.DB.Exec(
		SQLAddPressCycle,
		cycle.PressNumber,
		cycle.ToolID,
		cycle.ToolPosition,
		cycle.TotalCycles,
		cycle.Date,
		cycle.PerformedBy,
	)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	cycle.ID = models.CycleID(id)
	return cycle.ID, nil
}

// List retrieves all press cycles
func (s *PressCycles) List() ([]*models.Cycle, *errors.MasterError) {
	rows, err := s.DB.Query(SQLListPressCycles)
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

// Update modifies an existing press cycle
func (s *PressCycles) Update(
	id models.CycleID,
	press models.PressNumber,
	toolID models.ToolID,
	toolPosition models.Position,
	totalCycles int64,
	date time.Time,
	performedBy models.TelegramID,
) *errors.MasterError {

	cycle := models.NewCycle(
		press, toolID, toolPosition, totalCycles, performedBy,
	)
	cycle.ID = id
	cycle.Date = date

	verr := cycle.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	_, err := s.DB.Exec(
		SQLUpdatePressCycle,
		cycle.TotalCycles,
		cycle.ToolID,
		cycle.ToolPosition,
		performedBy,
		cycle.PressNumber,
		cycle.Date,
		cycle.ID,
	)
	return errors.NewMasterError(err, 0)
}

// Delete removes a press cycle from the database
func (s *PressCycles) Delete(id models.CycleID) *errors.MasterError {
	_, err := s.DB.Exec(SQLDeletePressCycle, id)
	return errors.NewMasterError(err, 0)
}
