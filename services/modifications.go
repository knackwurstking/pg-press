package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/logger"
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
	base := NewBase(r, logger.NewComponentLogger("Service: Modifications"))

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

		CREATE INDEX IF NOT EXISTS idx_%[1]s_entity ON %[1]s(entity_type, entity_id);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_created_at ON %[1]s(created_at);
		CREATE INDEX IF NOT EXISTS idx_%[1]s_user_id ON %[1]s(user_id);
	`, TableNameModifications)

	if err := base.CreateTable(query, TableNameModifications); err != nil {
		panic(err)
	}

	return &Modifications{Base: base}
}

func (s *Modifications) Add(mt models.ModificationType, mtID int64, data any, user models.TelegramID) (models.ModificationID, error) {
	s.Log.Debug("Adding modification: mt: %s, id: %d, user: %d", mt, mtID, user)

	if err := s.validateModificationType(mt, mtID); err != nil {
		return 0, err
	}

	if data == nil {
		return 0, errors.NewValidationError("modification data cannot be nil")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal modification data: %v", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (user_id, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, TableNameModifications)

	result, err := s.DB.Exec(query, user, mt, mtID, jsonData, time.Now())
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	return models.ModificationID(id), nil
}

func (s *Modifications) Get(id models.ModificationID) (*models.Modification[any], error) {
	s.Log.Debug("Getting modification: %v", id)

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE id = ?
	`, TableNameModifications)

	row := s.DB.QueryRow(query, id)
	mod, err := ScanSingleRow(row, scanModification)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

func (s *Modifications) List(mt models.ModificationType, mtID int64, limit, offset int) ([]*models.Modification[any], error) {
	s.Log.Debug("Listing modifications: mt: %s, mtID: %d, limit: %d, offset: %d", mt, mtID, limit, offset)

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
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanModification)
}

func (s *Modifications) ListAll(mt models.ModificationType, mtID int64) ([]*models.Modification[any], error) {
	return s.List(mt, mtID, -1, 0)
}

func (s *Modifications) Count(mt models.ModificationType, mtID int64) (int64, error) {
	s.Log.Debug("Counting modifications: mt: %s, mtID: %d", mt, mtID)

	if err := s.validateModificationType(mt, mtID); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
	`, TableNameModifications)

	var count int64
	err := s.DB.QueryRow(query, mt, mtID).Scan(&count)
	if err != nil {
		return 0, s.GetSelectError(err)
	}

	return count, nil
}

func (s *Modifications) GetLatest(mt models.ModificationType, mtID int64) (*models.Modification[any], error) {
	s.Log.Debug("Getting latest modification: mt: %s, mtID: %d", mt, mtID)

	if err := s.validateModificationType(mt, mtID); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, TableNameModifications)

	row := s.DB.QueryRow(query, mt, mtID)
	return ScanSingleRow(row, scanModification)
}

func (s *Modifications) GetOldest(mt models.ModificationType, mtID int64) (*models.Modification[any], error) {
	s.Log.Debug("Getting oldest modification: mt: %s, mtID: %d", mt, mtID)

	if err := s.validateModificationType(mt, mtID); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
		LIMIT 1
	`, TableNameModifications)

	row := s.DB.QueryRow(query, mt, mtID)
	return ScanSingleRow(row, scanModification)
}

func (s *Modifications) Delete(id models.ModificationID) error {
	s.Log.Debug("Deleting modification: %v", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameModifications)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func (s *Modifications) DeleteAll(mt models.ModificationType, mtID int64) error {
	s.Log.Debug("Deleting all modifications: mt=%s, mtID=%d", mt, mtID)

	if err := s.validateModificationType(mt, mtID); err != nil {
		return err
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE entity_type = ? AND entity_id = ?`, TableNameModifications)
	_, err := s.DB.Exec(query, mt, mtID)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func (s *Modifications) GetByUser(user int64, limit, offset int) ([]*models.Modification[any], error) {
	s.Log.Debug("Getting modifications by user: user: %d, limit: %d, offset: %d", user, limit, offset)

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, TableNameModifications)

	rows, err := s.DB.Query(query, user, limit, offset)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanModification)
}

func (s *Modifications) GetByDateRange(mt models.ModificationType, mtID int64, from, to time.Time) ([]*models.Modification[any], error) {
	s.Log.Debug("Getting modifications by date range: mt: %s, mtID: %d, from: %s, to: %s",
		mt, mtID, from.Format(time.RFC3339), to.Format(time.RFC3339))

	if err := s.validateModificationType(mt, mtID); err != nil {
		return nil, err
	}

	if from.After(to) {
		return nil, errors.NewValidationError("from date must be before to date")
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, data, created_at
		FROM %s
		WHERE entity_type = ? AND entity_id = ? AND created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`, TableNameModifications)

	rows, err := s.DB.Query(query, mt, mtID, from, to)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanModification)
}

func (s *Modifications) validateModificationType(mt models.ModificationType, mtID int64) error {
	if !slices.Contains(ModificationTypes, mt) {
		return errors.NewValidationError(fmt.Sprintf("modification type %s is not supported", mt))
	}

	if mtID <= 0 {
		return errors.NewValidationError("modification ID cannot be zero or negative")
	}

	return nil
}

func scanModification(scanner Scannable) (*models.Modification[any], error) {
	mod := &models.Modification[any]{}
	err := scanner.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan modification: %v", err)
	}
	return mod, nil
}
