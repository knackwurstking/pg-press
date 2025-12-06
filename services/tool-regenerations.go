package services

import (
	"fmt"
	"net/http"

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

func (s *ToolRegenerations) Get(id models.ToolRegenerationID) (*models.ToolRegeneration, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameToolRegenerations)
	row := s.DB.QueryRow(query, id)

	regeneration, err := ScanToolRegeneration(row)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return regeneration, nil
}

func (s *ToolRegenerations) Add(
	toolID models.ToolID, cycleID models.CycleID, reason string, user *models.User,
) (models.ToolRegenerationID, *errors.MasterError) {
	r := models.NewToolRegeneration(toolID, cycleID, reason, &user.TelegramID)
	if !r.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid tool regeneration: %s", r), http.StatusBadRequest)
	}

	if !user.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
	`, TableNameToolRegenerations)

	result, err := s.DB.Exec(query, r.ToolID, r.CycleID, r.Reason, user.TelegramID)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.ToolRegenerationID(id), nil
}

func (s *ToolRegenerations) Update(r *models.ToolRegeneration, user *models.User) *errors.MasterError {
	if !r.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid tool regeneration: %s", r), http.StatusBadRequest)
	}

	if !user.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`, TableNameToolRegenerations)

	_, err := s.DB.Exec(query, r.CycleID, r.Reason, user.TelegramID, r.ID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolRegenerations) Delete(id models.ToolRegenerationID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameToolRegenerations)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolRegenerations) StartToolRegeneration(
	id models.ToolID, reason string, user *models.User,
) (models.ToolRegenerationID, *errors.MasterError) {
	if !user.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	cycle, merr := s.Registry.PressCycles.GetLastToolCycle(id)
	if merr != nil {
		return 0, merr
	}

	merr = s.Registry.Tools.UpdateRegenerating(id, true, user)
	if merr != nil {
		return 0, merr
	}

	regenerationID, merr := s.Add(id, cycle.ID, reason, user)
	if merr != nil {
		undoErr := s.Registry.Tools.UpdateRegenerating(id, false, user)
		if undoErr != nil {
			return 0, undoErr
		}
		return 0, merr
	}

	return regenerationID, nil
}

func (s *ToolRegenerations) StopToolRegeneration(toolID models.ToolID, user *models.User) *errors.MasterError {
	if !user.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	merr := s.Registry.Tools.UpdateRegenerating(toolID, false, user)
	if merr != nil {
		return merr
	}

	return nil
}

func (s *ToolRegenerations) AbortToolRegeneration(toolID models.ToolID, user *models.User) *errors.MasterError {
	if !user.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	lastRegeneration, merr := s.GetLastRegeneration(toolID)
	if merr != nil {
		return merr
	}

	tool, merr := s.Registry.Tools.Get(toolID)
	if merr != nil {
		return merr
	}
	if !tool.Regenerating {
		return nil
	}

	merr = s.Delete(lastRegeneration.ID)
	if merr != nil {
		return merr
	}

	merr = s.Registry.Tools.UpdateRegenerating(toolID, false, user)
	if merr != nil {
		return merr
	}

	return nil
}

func (s *ToolRegenerations) GetLastRegeneration(toolID models.ToolID) (
	*models.ToolRegeneration, *errors.MasterError,
) {
	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableNameToolRegenerations)

	row := s.DB.QueryRow(query, toolID)
	r, err := ScanToolRegeneration(row)
	if err != nil {
		return r, errors.NewMasterError(err, 0)
	}
	return r, nil
}

func (s *ToolRegenerations) HasRegenerationsForCycle(cycleID models.CycleID) (bool, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE cycle_id = ?`, TableNameToolRegenerations)
	count, merr := s.QueryCount(query, cycleID)
	return count > 0, merr
}

func (s *ToolRegenerations) GetRegenerationHistory(toolID models.ToolID) (
	[]*models.ToolRegeneration, *errors.MasterError,
) {
	query := fmt.Sprintf(`
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM %s
		WHERE tool_id = ?
		ORDER BY id DESC
	`, TableNameToolRegenerations)

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanToolRegeneration)
}
