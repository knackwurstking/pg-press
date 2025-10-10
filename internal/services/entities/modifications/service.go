package modifications

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Service handles database operations for modifications
type Service struct {
	*base.BaseService
}

// NewService creates a new ModificationService instance
func NewService(db *sql.DB) *Service {
	base := base.NewBaseService(db, "Modifications")

	query := `
		CREATE TABLE IF NOT EXISTS modifications (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE SET NULL
		);

		CREATE INDEX IF NOT EXISTS idx_modifications_entity ON modifications(entity_type, entity_id);
		CREATE INDEX IF NOT EXISTS idx_modifications_created_at ON modifications(created_at);
		CREATE INDEX IF NOT EXISTS idx_modifications_user_id ON modifications(user_id);
	`

	if err := base.CreateTable(query, "modifications"); err != nil {
		panic(err)
	}

	return &Service{
		BaseService: base,
	}
}

// Add creates a new modification record
func (s *Service) Add(
	userID int64, entityType models.ModificationType, entityID int64, data any,
) (*models.Modification[any], error) {
	if err := validation.ValidateID(userID, "user"); err != nil {
		return nil, err
	}

	if err := validateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	if data == nil {
		return nil, utils.NewValidationError("modification data cannot be nil")
	}

	s.Log.Debug("Adding modification: user_id: %d, entity_type: %s, entity_id: %d", userID, entityType, entityID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modification data: %v", err)
	}

	query := `
		INSERT INTO modifications (user_id, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := s.DB.Exec(query,
		userID, string(entityType), entityID, jsonData, now)
	if err != nil {
		return nil, s.HandleInsertError(err, "modifications")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, s.HandleInsertError(err, "modifications")
	}

	mod := &models.Modification[any]{
		ID:        id,
		UserID:    userID,
		Data:      jsonData,
		CreatedAt: now,
	}

	s.Log.Debug("Added modification: id: %d", id)
	return mod, nil
}

// Get retrieves a specific modification by ID
func (s *Service) Get(id int64) (*models.Modification[any], error) {
	if err := validation.ValidateID(id, "modification"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting modification: %v", id)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE id = ?
	`

	row := s.DB.QueryRow(query, id)

	mod, err := scanner.ScanSingleRow(row, modificationScanner, "modifications")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

// List retrieves all modifications for a specific entity
func (s *Service) List(
	entityType models.ModificationType, entityID int64, limit, offset int,
) ([]*models.Modification[any], error) {
	if err := validateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	s.Log.Debug("Listing modifications: entity_type: %s, entity_id: %d, limit: %d, offset: %d", entityType, entityID, limit, offset)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.DB.Query(query, string(entityType), entityID, limit, offset)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	modifications, err := scanModificationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Debug("Listed modifications: count: %d", len(modifications))
	return modifications, nil
}

// ListAll retrieves all modifications for a specific entity without pagination
func (s *Service) ListAll(
	entityType models.ModificationType, entityID int64,
) ([]*models.Modification[any], error) {
	return s.List(entityType, entityID, -1, 0)
}

// Count returns the total number of modifications for a specific entity
func (s *Service) Count(entityType models.ModificationType, entityID int64) (int64, error) {
	if err := validateModificationType(string(entityType)); err != nil {
		return 0, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return 0, err
	}

	s.Log.Debug("Counting modifications: entity_type: %s, entity_id: %d",
		entityType, entityID)

	query := `
		SELECT COUNT(*)
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
	`

	var count int64
	err := s.DB.QueryRow(query, string(entityType), entityID).Scan(&count)
	if err != nil {
		return 0, s.HandleSelectError(err, "modifications")
	}

	s.Log.Debug("Counted modifications: count: %d", count)
	return count, nil
}

// GetLatest retrieves the most recent modification for a specific entity
func (s *Service) GetLatest(
	entityType models.ModificationType, entityID int64,
) (*models.Modification[any], error) {
	if err := validateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting latest modification: entity_type: %s, entity_id: %d", entityType, entityID)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := s.DB.QueryRow(query, string(entityType), entityID)

	mod, err := scanner.ScanSingleRow(row, modificationScanner, "modifications")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

// GetOldest retrieves the oldest modification for a specific entity
func (s *Service) GetOldest(
	entityType models.ModificationType, entityID int64,
) (*models.Modification[any], error) {
	if err := validateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting oldest modification: entity_type: %s, entity_id: %d", entityType, entityID)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
		LIMIT 1
	`

	row := s.DB.QueryRow(query, string(entityType), entityID)

	mod, err := scanner.ScanSingleRow(
		row, modificationScanner, "modifications")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

// Delete removes a specific modification by ID
func (s *Service) Delete(id int64) error {
	if err := validation.ValidateID(id, "modification"); err != nil {
		return err
	}

	s.Log.Debug("Deleting modification: %v", id)

	query := `DELETE FROM modifications WHERE id = ?`
	result, err := s.DB.Exec(query, id)
	if err != nil {
		return s.HandleDeleteError(err, "modifications")
	}

	if err := s.CheckRowsAffected(result, "modification", id); err != nil {
		return err
	}

	s.Log.Debug("Deleted modification: %v", id)
	return nil
}

// DeleteAll removes all modifications for a specific entity
func (s *Service) DeleteAll(entityType models.ModificationType, entityID int64) error {
	if err := validateModificationType(string(entityType)); err != nil {
		return err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return err
	}

	s.Log.Info("Deleting all modifications: entity_type=%s, entity_id=%d", entityType, entityID)

	query := `DELETE FROM modifications WHERE entity_type = ? AND entity_id = ?`
	result, err := s.DB.Exec(query, string(entityType), entityID)
	if err != nil {
		return s.HandleDeleteError(err, "modifications")
	}

	rowsAffected, err := s.GetRowsAffected(result, "delete all modifications")
	if err != nil {
		return err
	}

	s.Log.Info("Successfully deleted %d modifications", rowsAffected)
	return nil
}

// GetByUser retrieves all modifications made by a specific user
func (s *Service) GetByUser(userID int64, limit, offset int) ([]*models.Modification[any], error) {
	if err := validation.ValidateID(userID, "user"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting modifications by user: user_id: %d, limit: %d, offset: %d", userID, limit, offset)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.DB.Query(query, userID, limit, offset)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	modifications, err := scanModificationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Debug("Found modifications by user: count: %d", len(modifications))
	return modifications, nil
}

// GetByDateRange retrieves modifications within a specific date range
func (s *Service) GetByDateRange(
	entityType models.ModificationType, entityID int64, from, to time.Time,
) ([]*models.Modification[any], error) {
	if err := validateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	if from.After(to) {
		return nil, utils.NewValidationError("from date must be before to date")
	}

	s.Log.Debug("Getting modifications by date range: entity_type: %s, entity_id: %d, from: %s, to: %s",
		entityType, entityID, from.Format(time.RFC3339), to.Format(time.RFC3339))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ? AND created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`

	rows, err := s.DB.Query(query, string(entityType), entityID, from, to)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	modifications, err := scanModificationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Debug("Found modifications by date range: count: %d", len(modifications))
	return modifications, nil
}

// Helper methods for specific entity types

// AddTroubleReportMod adds a modification for a trouble report
func (s *Service) AddTroubleReportMod(userID, reportID int64, data any) error {
	_, err := s.Add(userID, models.ModificationTypeTroubleReport, reportID, data)
	return err
}

// AddMetalSheetMod adds a modification for a metal sheet
func (s *Service) AddMetalSheetMod(userID, sheetID int64, data any) error {
	_, err := s.Add(userID, models.ModificationTypeMetalSheet, sheetID, data)
	return err
}

// AddToolMod adds a modification for a tool
func (s *Service) AddToolMod(userID, toolID int64, data any) error {
	_, err := s.Add(userID, models.ModificationTypeTool, toolID, data)
	return err
}

// AddPressCycleMod adds a modification for a press cycle
func (s *Service) AddPressCycleMod(userID, cycleID int64, data any) error {
	_, err := s.Add(userID, models.ModificationTypePressCycle, cycleID, data)
	return err
}

// AddUserMod adds a modification for a user
func (s *Service) AddUserMod(userID, targetUserID int64, data any) error {
	_, err := s.Add(userID, models.ModificationTypeUser, targetUserID, data)
	return err
}

// GetWithUser retrieves a modification with user information
func (s *Service) GetWithUser(id int64) (*models.ModificationWithUser, error) {
	if err := validation.ValidateID(id, "modification"); err != nil {
		return nil, err
	}

	s.Log.Debug("Getting modification with user: %v", id)

	query := `
		SELECT m.id, m.user_id, m.data, m.created_at, u.user_name
		FROM modifications m
		JOIN users u ON m.user_id = u.telegram_id
		WHERE m.id = ?
	`

	row := s.DB.QueryRow(query, id)

	modWithUser := &models.ModificationWithUser{}
	err := row.Scan(
		&modWithUser.Modification.ID,
		&modWithUser.Modification.UserID,
		&modWithUser.Modification.Data,
		&modWithUser.Modification.CreatedAt,
		&modWithUser.User.Name,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, s.HandleSelectError(err, "modifications")
	}

	modWithUser.User.TelegramID = modWithUser.Modification.UserID
	return modWithUser, nil
}

// ListWithUser retrieves modifications with user information for a specific entity
func (s *Service) ListWithUser(
	entityType models.ModificationType, entityID int64, limit, offset int,
) ([]*models.ModificationWithUser, error) {
	if err := validateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := validation.ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	s.Log.Debug("Listing modifications with user: entity_type: %s, entity_id: %d", entityType, entityID)

	query := `
		SELECT m.id, m.user_id, m.data, m.created_at, u.user_name
		FROM modifications m
		JOIN users u ON m.user_id = u.telegram_id
		WHERE m.entity_type = ? AND m.entity_id = ?
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.DB.Query(query, string(entityType), entityID, limit, offset)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	var modifications []*models.ModificationWithUser
	for rows.Next() {
		modWithUser := &models.ModificationWithUser{}
		err := rows.Scan(
			&modWithUser.Modification.ID,
			&modWithUser.Modification.UserID,
			&modWithUser.Modification.Data,
			&modWithUser.Modification.CreatedAt,
			&modWithUser.User.Name,
		)
		if err != nil {
			return nil, s.HandleScanError(err, "modifications")
		}
		modWithUser.User.TelegramID = modWithUser.Modification.UserID
		modifications = append(modifications, modWithUser)
	}

	if err = rows.Err(); err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}

	s.Log.Debug("Listed modifications with user: count: %d", len(modifications))
	return modifications, nil
}
