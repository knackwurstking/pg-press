package services

import (
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

type ToolRegenerations struct {
	*Base
}

func NewToolRegenerations(r *Registry) *ToolRegenerations {
	return &ToolRegenerations{
		Base: NewBase(r),
	}
}

func (s *ToolRegenerations) Get(id models.ToolRegenerationID) (*models.ToolRegeneration, *errors.MasterError) {
	query := `SELECT * FROM tool_regenerations WHERE id = ?`
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

	verr := r.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	verr = user.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	query := `
		INSERT INTO tool_regenerations (tool_id, cycle_id, reason, performed_by)
		VALUES (?, ?, ?, ?)
	`

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
	verr := r.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	verr = user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	query := `
		UPDATE tool_regenerations
		SET cycle_id = ?, reason = ?, performed_by = ?
		WHERE id = ?
	`

	_, err := s.DB.Exec(query, r.CycleID, r.Reason, user.TelegramID, r.ID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolRegenerations) Delete(id models.ToolRegenerationID) *errors.MasterError {
	query := `DELETE FROM tool_regenerations WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *ToolRegenerations) StartToolRegeneration(
	id models.ToolID, reason string, user *models.User,
) (models.ToolRegenerationID, *errors.MasterError) {

	verr := user.Validate()
	if verr != nil {
		return 0, verr.MasterError()
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
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	merr := s.Registry.Tools.UpdateRegenerating(toolID, false, user)
	if merr != nil {
		return merr
	}

	return nil
}

func (s *ToolRegenerations) AbortToolRegeneration(toolID models.ToolID, user *models.User) *errors.MasterError {
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
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
	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
		LIMIT 1
	`

	row := s.DB.QueryRow(query, toolID)
	r, err := ScanToolRegeneration(row)
	if err != nil {
		return r, errors.NewMasterError(err, 0)
	}
	return r, nil
}

func (s *ToolRegenerations) HasRegenerationsForCycle(cycleID models.CycleID) (bool, *errors.MasterError) {
	query := `SELECT COUNT(*) FROM tool_regenerations WHERE cycle_id = ?`
	count, merr := s.QueryCount(query, cycleID)
	return count > 0, merr
}

func (s *ToolRegenerations) GetRegenerationHistory(toolID models.ToolID) (
	[]*models.ToolRegeneration, *errors.MasterError,
) {
	query := `
		SELECT id, tool_id, cycle_id, reason, performed_by
		FROM tool_regenerations
		WHERE tool_id = ?
		ORDER BY id DESC
	`

	rows, err := s.DB.Query(query, toolID)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanToolRegeneration)
}
