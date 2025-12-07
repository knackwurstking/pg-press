package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

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
	return &Modifications{
		Base: NewBase(r),
	}
}

func (s *Modifications) Add(mt models.ModificationType, mtID int64, data any, user models.TelegramID) (models.ModificationID, *errors.MasterError) {
	if merr := s.validateModificationType(mt, mtID); merr != nil {
		return 0, merr
	}

	if data == nil {
		return 0, errors.NewMasterError(fmt.Errorf("missing data"), http.StatusBadRequest)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, errors.NewMasterError(fmt.Errorf("marshal data: %#v", err), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (user_id, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, TableNameModifications)

	result, err := s.DB.Exec(query, user, mt, mtID, jsonData, time.Now())
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
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
		return m, errors.NewMasterError(err, 0)
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
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) ListAll(mt models.ModificationType, mtID int64) ([]*models.Modification[any], *errors.MasterError) {
	return s.List(mt, mtID, -1, 0)
}

func (s *Modifications) Count(mt models.ModificationType, mtID int64) (int64, *errors.MasterError) {
	if merr := s.validateModificationType(mt, mtID); merr != nil {
		return 0, merr
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
	`, TableNameModifications)

	var count int64
	err := s.DB.QueryRow(query, mt, mtID).Scan(&count)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return count, nil
}

func (s *Modifications) GetLatest(mt models.ModificationType, mtID int64) (*models.Modification[any], *errors.MasterError) {
	if merr := s.validateModificationType(mt, mtID); merr != nil {
		return nil, merr
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
		return m, errors.NewMasterError(err, 0)
	}
	return m, nil
}

func (s *Modifications) GetOldest(mt models.ModificationType, mtID int64) (*models.Modification[any], *errors.MasterError) {
	if merr := s.validateModificationType(mt, mtID); merr != nil {
		return nil, merr
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
		return m, errors.NewMasterError(err, 0)
	}
	return m, nil
}

func (s *Modifications) Delete(id models.ModificationID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameModifications)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *Modifications) DeleteAll(mt models.ModificationType, mtID int64) *errors.MasterError {
	if merr := s.validateModificationType(mt, mtID); merr != nil {
		return merr
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE entity_type = ? AND entity_id = ?`, TableNameModifications)
	_, err := s.DB.Exec(query, mt, mtID)
	if err != nil {
		return errors.NewMasterError(err, 0)
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
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) GetByDateRange(mt models.ModificationType, mtID int64, from, to time.Time) ([]*models.Modification[any], *errors.MasterError) {
	if err := s.validateModificationType(mt, mtID); err != nil {
		return nil, err
	}

	if from.After(to) {
		return nil, errors.NewMasterError(fmt.Errorf("invalid timing from %#v to %#v", from, to), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ? AND created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`, TableNameModifications)

	rows, err := s.DB.Query(query, mt, mtID, from, to)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanModification)
}

func (s *Modifications) validateModificationType(mt models.ModificationType, mtID int64) *errors.MasterError {
	if !slices.Contains(ModificationTypes, mt) {
		return errors.NewMasterError(fmt.Errorf("invalid modification type: %#v", mt), http.StatusBadRequest)
	}

	if mtID <= 0 {
		return errors.NewMasterError(fmt.Errorf("modidfication type id cannot be lower or equal 0"), http.StatusBadRequest)
	}

	return nil
}
