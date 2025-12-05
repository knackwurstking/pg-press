package services

import (
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameModifications = "modifications"

var ModificationTypes = []models.ModificationType{
	models.ModificationTypeTroubleReport,
	models.ModificationTypeMetalSheet,
	models.ModificationTypeTool,
	models.ModificationTypePressCycle,
	models.ModificationTypeUser,
	models.ModificationTypeNote,
	models.ModificationTypeAttachment,
}

type Modifications struct {
	*Base
}

func NewModifications(r *Registry) *Modifications {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %[1]s (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE SET NULL
		);
	`, TableNameModifications)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameModifications))
	}

	return &Modifications{Base: base}
}

func (s *Modifications) Add(mt models.ModificationType, mtID int64, data any, user models.TelegramID) (models.ModificationID, *errors.MasterError) {
	if dberr := s.validateModificationType(mt, mtID); dberr != nil {
		return 0, dberr
	}

	if data == nil {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (user_id, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, TableNameModifications)

	result, err := s.DB.Exec(query, user, mt, mtID, jsonData, time.Now())
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	return models.ModificationID(id), nil
}

func (s *Modifications) Get(id models.ModificationID) (*models.Modification[any], *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE id = ?
	`, TableNameModifications)

	row := s.DB.QueryRow(query, id)
	m, err := ScanModification(row)
	if err != nil {
		return m, errors.NewMasterError(err)
	}
	return m, nil
}

func (s *Modifications) List(mt models.ModificationType, mtID int64, limit, offset int) ([]*models.Modification[any], *errors.MasterError) {
	if err := s.validateModificationType(mt, mtID); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, TableNameModifications)

	rows, err := s.DB.Query(query, mt, mtID, limit, offset)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) ListAll(mt models.ModificationType, mtID int64) ([]*models.Modification[any], *errors.MasterError) {
	return s.List(mt, mtID, -1, 0)
}

func (s *Modifications) Count(mt models.ModificationType, mtID int64) (int64, *errors.MasterError) {
	if dberr := s.validateModificationType(mt, mtID); dberr != nil {
		return 0, dberr
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
	`, TableNameModifications)

	var count int64
	err := s.DB.QueryRow(query, mt, mtID).Scan(&count)
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	return count, nil
}

func (s *Modifications) GetLatest(mt models.ModificationType, mtID int64) (*models.Modification[any], *errors.MasterError) {
	if dberr := s.validateModificationType(mt, mtID); dberr != nil {
		return nil, dberr
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, TableNameModifications)

	row := s.DB.QueryRow(query, mt, mtID)
	m, err := ScanModification(row)
	if err != nil {
		return m, errors.NewMasterError(err)
	}
	return m, nil
}

func (s *Modifications) GetOldest(mt models.ModificationType, mtID int64) (*models.Modification[any], *errors.MasterError) {
	if dberr := s.validateModificationType(mt, mtID); dberr != nil {
		return nil, dberr
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
		LIMIT 1
	`, TableNameModifications)

	row := s.DB.QueryRow(query, mt, mtID)
	m, err := ScanModification(row)
	if err != nil {
		return m, errors.NewMasterError(err)
	}
	return m, nil
}

func (s *Modifications) Delete(id models.ModificationID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameModifications)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}

func (s *Modifications) DeleteAll(mt models.ModificationType, mtID int64) *errors.MasterError {
	if dberr := s.validateModificationType(mt, mtID); dberr != nil {
		return dberr
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE entity_type = ? AND entity_id = ?`, TableNameModifications)
	_, err := s.DB.Exec(query, mt, mtID)
	if err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}

func (s *Modifications) GetByUser(user int64, limit, offset int) ([]*models.Modification[any], *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, TableNameModifications)

	rows, err := s.DB.Query(query, user, limit, offset)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) GetByDateRange(mt models.ModificationType, mtID int64, from, to time.Time) ([]*models.Modification[any], *errors.MasterError) {
	if err := s.validateModificationType(mt, mtID); err != nil {
		return nil, err
	}

	if from.After(to) {
		return nil, errors.NewMasterError(errors.ErrValidation)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ? AND created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`, TableNameModifications)

	rows, err := s.DB.Query(query, mt, mtID, from, to)
	if err != nil {
		return nil, errors.NewMasterError(err)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) validateModificationType(mt models.ModificationType, mtID int64) *errors.MasterError {
	if !slices.Contains(ModificationTypes, mt) {
		return errors.NewMasterError(errors.ErrValidation)
	}

	if mtID <= 0 {
		return errors.NewMasterError(errors.ErrValidation)
	}

	return nil
}
