package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

// Get retrieves a press cycle by ID
func (s *PressCycles) Get(id models.CycleID) (*models.Cycle, error) {
	slog.Debug("Getting press cycle", "id", id)

	query := fmt.Sprintf(`
		SELECT id, press_number, tool_id, tool_position, total_cycles, date, performed_by
		FROM %s
		WHERE id = ?
	`, TableNamePressCycles)

	row := s.DB.QueryRow(query, id)
	cycle, err := ScanSingleRow(row, scanCycle)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("Press cycle with ID %d not found", id))
		}
		return nil, err
	}

	cycle.PartialCycles = s.GetPartialCycles(cycle)
	return cycle, nil
}

// Add creates a new press cycle
func (s *PressCycles) Add(cycle *models.Cycle, user *models.User) (models.CycleID, error) {
	slog.Debug(
		"Adding press cycle",
		"user_name", user.Name,
		"tool_id", cycle.ToolID,
		"cycle.ToolPosition", cycle.ToolPosition,
		"cycle.TotalCycles", cycle.TotalCycles,
	)

	if err := cycle.Validate(); err != nil {
		return 0, err
	}

	if err := user.Validate(); err != nil {
		return 0, err
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
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	cycle.ID = models.CycleID(id)
	return cycle.ID, nil
}

// List retrieves all press cycles
func (s *PressCycles) List() ([]*models.Cycle, error) {
	slog.Debug("Listing press cycles")

	// TODO: ... [WIP]
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		ORDER BY date DESC
	`, TableNamePressCycles)

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	cycles, err := ScanRows(rows, scanCycle)
	if err != nil {
		return nil, err
	}

	cycles = s.injectPartialCycles(cycles)
	return cycles, nil
}

// Update modifies an existing press cycle
func (s *PressCycles) Update(cycle *models.Cycle, user *models.User) error {
	slog.Debug("Updating press cycle", "user_name", user.Name, "cycle.ID", cycle.ID)

	if err := cycle.Validate(); err != nil {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
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

	return s.GetUpdateError(err)
}

// Delete removes a press cycle from the database
func (s *PressCycles) Delete(id models.CycleID) error {
	slog.Debug("Deleting press cycle", "cycle", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNamePressCycles)
	_, err := s.DB.Exec(query, id)

	return s.GetDeleteError(err)
}
