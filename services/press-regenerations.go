package services

import (
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

type PressRegenerations struct {
	*Base
}

func NewPressRegenerations(r *Registry) *PressRegenerations {
	return &PressRegenerations{
		Base: NewBase(r),
	}
}

func (s *PressRegenerations) Get(id models.PressRegenerationID) (*models.PressRegeneration, *errors.MasterError) {
	query := `SELECT * FROM press_regenerations WHERE id = ?`
	row := s.DB.QueryRow(query, id)
	r, err := ScanPressRegeneration(row)
	if err != nil {
		return r, errors.NewMasterError(err, 0)
	}
	return r, nil
}

func (s *PressRegenerations) Add(r *models.PressRegeneration) (models.PressRegenerationID, *errors.MasterError) {
	verr := r.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	query := `
		INSERT INTO press_regenerations (press_number, started_at, completed_at, reason)
		VALUES (?, ?, ?, ?)
	`

	result, err := s.DB.Exec(query, r.PressNumber, r.StartedAt, r.CompletedAt, r.Reason)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.PressRegenerationID(id), nil
}

func (s *PressRegenerations) Update(r *models.PressRegeneration) *errors.MasterError {
	verr := r.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	query := `
		UPDATE press_regenerations
		SET started_at = ?, completed_at = ?, reason = ?
		WHERE id = ?
	`

	if _, err := s.DB.Exec(query, r.StartedAt, r.CompletedAt, r.Reason, r.ID); err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressRegenerations) Delete(id models.PressRegenerationID) *errors.MasterError {
	query := `DELETE FROM press_regenerations WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressRegenerations) StartPressRegeneration(
	pn models.PressNumber, reason string,
) (models.PressRegenerationID, *errors.MasterError) {
	regenerationID, err := s.Add(models.NewPressRegeneration(pn, time.Now(), reason))
	if err != nil {
		return 0, err
	}

	return regenerationID, nil
}

func (s *PressRegenerations) StopPressRegeneration(id models.PressRegenerationID) *errors.MasterError {
	regeneration, merr := s.Get(id)
	if merr != nil {
		return merr
	}

	regeneration.Stop()

	verr := regeneration.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	query := `
		UPDATE press_regenerations
		SET completed_at = ?
		WHERE id = ?
	`

	_, err := s.DB.Exec(query, regeneration.CompletedAt, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *PressRegenerations) GetLastRegeneration(
	pressNumber models.PressNumber,
) (*models.PressRegeneration, *errors.MasterError) {
	query := `
		SELECT id, press_number, started_at, completed_at, reason
		FROM press_regenerations
		WHERE press_number = ?
		ORDER BY id DESC
		LIMIT 1
	`

	r, err := ScanPressRegeneration(s.DB.QueryRow(query, pressNumber))
	if err != nil {
		return r, errors.NewMasterError(err, 0)
	}
	return r, nil
}

func (s *PressRegenerations) GetRegenerationHistory(
	pressNumber models.PressNumber,
) ([]*models.PressRegeneration, *errors.MasterError) {
	query := `
		SELECT id, press_number, started_at, completed_at, reason
		FROM press_regenerations
		WHERE press_number = ?
		ORDER BY id DESC
	`

	rows, err := s.DB.Query(query, pressNumber)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanPressRegeneration)
}
