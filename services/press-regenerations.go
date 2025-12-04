package services

import (
	"fmt"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNamePressRegenerations = "press_regenerations"

type PressRegenerations struct {
	*Base
}

func NewPressRegenerations(r *Registry) *PressRegenerations {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			press_number INTEGER NOT NULL,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			reason TEXT
		);
	`, TableNamePressRegenerations)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNamePressRegenerations))
	}

	return &PressRegenerations{
		Base: base,
	}
}

func (s *PressRegenerations) Get(id models.PressRegenerationID) (*models.PressRegeneration, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNamePressRegenerations)
	row := s.DB.QueryRow(query, id)

	return ScanRow(row, ScanPressRegeneration)
}

func (s *PressRegenerations) Add(r *models.PressRegeneration) (models.PressRegenerationID, *errors.DBError) {
	if err := r.Validate(); err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (press_number, started_at, completed_at, reason)
		VALUES (?, ?, ?, ?)
	`, TableNamePressRegenerations)

	result, err := s.DB.Exec(query, r.PressNumber, r.StartedAt, r.CompletedAt, r.Reason)
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	return models.PressRegenerationID(id), nil
}

func (s *PressRegenerations) Update(r *models.PressRegeneration) *errors.DBError {
	if err := r.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET started_at = ?, completed_at = ?, reason = ?
		WHERE id = ?
	`, TableNamePressRegenerations)

	if _, err := s.DB.Exec(query, r.StartedAt, r.CompletedAt, r.Reason, r.ID); err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (s *PressRegenerations) Delete(id models.PressRegenerationID) *errors.DBError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNamePressRegenerations)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}

func (s *PressRegenerations) StartPressRegeneration(pn models.PressNumber, reason string) (models.PressRegenerationID, *errors.DBError) {
	regenerationID, err := s.Add(models.NewPressRegeneration(pn, time.Now(), reason))
	if err != nil {
		return 0, err
	}

	return regenerationID, nil
}

func (s *PressRegenerations) StopPressRegeneration(id models.PressRegenerationID) *errors.DBError {
	regeneration, dberr := s.Get(id)
	if dberr != nil {
		return dberr
	}

	regeneration.Stop()

	if err := regeneration.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET completed_at = ?
		WHERE id = ?
	`, TableNamePressRegenerations)

	_, err := s.DB.Exec(query, regeneration.CompletedAt, id)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (s *PressRegenerations) GetLastRegeneration(pressNumber models.PressNumber) (*models.PressRegeneration, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT id, press_number, started_at, completed_at, reason
		FROM %s
		WHERE press_number = ?
		ORDER BY id DESC
		LIMIT 1
	`, TableNamePressRegenerations)

	return ScanRow(s.DB.QueryRow(query, pressNumber), ScanPressRegeneration)
}

func (s *PressRegenerations) GetRegenerationHistory(pressNumber models.PressNumber) ([]*models.PressRegeneration, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT id, press_number, started_at, completed_at, reason
		FROM %s
		WHERE press_number = ?
		ORDER BY id DESC
	`, TableNamePressRegenerations)

	rows, err := s.DB.Query(query, pressNumber)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanPressRegeneration)
}
