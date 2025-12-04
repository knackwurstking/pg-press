package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameToolRegenerations = "tool_regenerations"

type ToolRegenerations struct {
	*Base
}

func NewToolRegenerations(r *Registry) *ToolRegenerations {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL,
			reason TEXT,
			performed_by INTEGER NOT NULL
		);
	`, TableNameToolRegenerations, TableNameTools)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameToolRegenerations))
	}

	return &ToolRegenerations{
		Base: base,
	}
}

func (s *ToolRegenerations) Get(id models.ToolRegenerationID) (*models.ToolRegeneration, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameToolRegenerations)
	row := s.DB.QueryRow(query, id)

	regeneration, err := ScanRow(row, ScanToolRegeneration)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}

	return regeneration, nil
}

func (s *ToolRegenerations) Add(toolID models.ToolID, cycleID models.CycleID, reason string, user *models.User) (
	models.ToolRegenerationID, *errors.DBError,
) {
	err := user.Validate()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	r := models.NewToolRegeneration(toolID, cycleID, reason, &user.TelegramID)
	if err := r.Validate(); err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
	`, TableNameToolRegenerations)

	result, err := s.DB.Exec(query, r.ToolID, r.CycleID, r.Reason, user.TelegramID)
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	return models.ToolRegenerationID(id), nil
}

func (s *ToolRegenerations) Update(r *models.ToolRegeneration, user *models.User) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	err = r.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`, TableNameToolRegenerations)

	_, err = s.DB.Exec(query, r.CycleID, r.Reason, user.TelegramID, r.ID)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (s *ToolRegenerations) Delete(id models.ToolRegenerationID) *errors.DBError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameToolRegenerations)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}

func (s *ToolRegenerations) StartToolRegeneration(id models.ToolID, reason string, user *models.User) (
	models.ToolRegenerationID, *errors.DBError,
) {
	err := user.Validate()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	cycle, dberr := s.Registry.PressCycles.GetLastToolCycle(id)
	if err != nil {
		return 0, dberr
	}

	dberr = s.Registry.Tools.UpdateRegenerating(id, true, user)
	if dberr != nil {
		return 0, dberr
	}

	regenerationID, dberr := s.Add(id, cycle.ID, reason, user)
	if dberr != nil {
		undoErr := s.Registry.Tools.UpdateRegenerating(id, false, user)
		if undoErr != nil {
			return 0, undoErr
		}
		return 0, dberr
	}

	return regenerationID, nil
}

func (s *ToolRegenerations) StopToolRegeneration(toolID models.ToolID, user *models.User) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	dberr := s.Registry.Tools.UpdateRegenerating(toolID, false, user)
	if dberr != nil {
		return dberr
	}

	return nil
}

func (s *ToolRegenerations) AbortToolRegeneration(toolID models.ToolID, user *models.User) *errors.DBError {
	if err := user.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	lastRegeneration, dberr := s.GetLastRegeneration(toolID)
	if dberr != nil {
		return dberr
	}

	tool, dberr := s.Registry.Tools.Get(toolID)
	if dberr != nil {
		return dberr
	}
	if !tool.Regenerating {
		return nil
	}

	dberr = s.Delete(lastRegeneration.ID)
	if dberr != nil {
		return dberr
	}

	dberr = s.Registry.Tools.UpdateRegenerating(toolID, false, user)
	if dberr != nil {
		return dberr
	}

	return nil
}

func (s *ToolRegenerations) GetLastRegeneration(toolID models.ToolID) (*models.ToolRegeneration, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableNameToolRegenerations)

	row := s.DB.QueryRow(query, toolID)
	return ScanRow(row, ScanToolRegeneration)
}

func (s *ToolRegenerations) HasRegenerationsForCycle(cycleID models.CycleID) (bool, *errors.DBError) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE cycle_id = ?`, TableNameToolRegenerations)
	count, dberr := s.QueryCount(query, cycleID)
	return count > 0, dberr
}

func (s *ToolRegenerations) GetRegenerationHistory(toolID models.ToolID) ([]*models.ToolRegeneration, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
	`, TableNameToolRegenerations)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanToolRegeneration)
}
