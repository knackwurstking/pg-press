package services

import (
	"fmt"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

// Get retrieves a press cycle by ID
func (s *PressCycles) Get(id models.CycleID) (*models.Cycle, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE id = ?
	`, TableNamePressCycles)

	row := s.DB.QueryRow(query, id)
	cycle, err := ScanCycle(row)
	if err != nil {
		return cycle, errors.NewMasterError(err)
	}

	cycle.PartialCycles = s.GetPartialCycles(cycle)
	return cycle, nil
}

// Add creates a new press cycle
func (s *PressCycles) Add(cycle *models.Cycle, user *models.User) (models.CycleID, *errors.MasterError) {
	if cycle.Validate() {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	if !user.Validate() {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (press_number, tool_id, tool_position, total_cycles, date, performed_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, TableNamePressCycles)

	result, err := s.DB.Exec(query,
		cycle.PressNumber,
		cycle.ToolID,
		cycle.ToolPosition,
		cycle.TotalCycles,
		cycle.Date,
		user.TelegramID,
	)
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	cycle.ID = models.CycleID(id)
	return cycle.ID, nil
}

// List retrieves all press cycles
func (s *PressCycles) List() ([]*models.Cycle, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		ORDER BY date DESC
	`, TableNamePressCycles)

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer rows.Close()

	cycles, dberr := ScanRows(rows, ScanCycle)
	if dberr != nil {
		return nil, dberr
	}

	s.injectPartialCycles(cycles)
	return cycles, nil
}

// Update modifies an existing press cycle
func (s *PressCycles) Update(cycle *models.Cycle, user *models.User) *errors.MasterError {
	if !cycle.Validate() || !user.Validate() {
		return errors.NewMasterError(errors.ErrValidation)
	}

	if cycle.Date.IsZero() {
		cycle.Date = time.Now()
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET total_cycles = ?, tool_id = ?, tool_position = ?, performed_by = ?, press_number = ?, date = ?
		WHERE id = ?
	`, TableNamePressCycles)

	_, err := s.DB.Exec(query,
		cycle.TotalCycles,
		cycle.ToolID,
		cycle.ToolPosition,
		user.TelegramID,
		cycle.PressNumber,
		cycle.Date,
		cycle.ID,
	)
	return errors.NewMasterError(err)
}

// Delete removes a press cycle from the database
func (s *PressCycles) Delete(id models.CycleID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNamePressCycles)
	_, err := s.DB.Exec(query, id)
	return errors.NewMasterError(err)
}
