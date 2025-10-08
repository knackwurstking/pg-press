package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// ModificationType represents the type of entity being modified
type ModificationType string

const (
	ModificationTypeTroubleReport ModificationType = "trouble_reports"
	ModificationTypeMetalSheet    ModificationType = "metal_sheets"
	ModificationTypeTool          ModificationType = "tools"
	ModificationTypePressCycle    ModificationType = "press_cycles"
	ModificationTypeUser          ModificationType = "users"
	ModificationTypeNote          ModificationType = "notes"
	ModificationTypeAttachment    ModificationType = "attachments"
)

// ModificationWithUser represents a modification with user information
type ModificationWithUser struct {
	Modification models.Modification[any] `json:"modification"`
	User         models.User              `json:"user"`
}

// Modifications handles database operations for modifications
type Modifications struct {
	*BaseService
}

// NewModifications creates a new ModificationService instance
func NewModifications(db *sql.DB) *Modifications {
	base := NewBaseService(db, "Modifications")

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

	return &Modifications{
		BaseService: base,
	}
}

// Add creates a new modification record
func (s *Modifications) Add(
	userID int64, entityType ModificationType, entityID int64, data any,
) (*models.Modification[any], error) {
	if err := ValidateID(userID, "user"); err != nil {
		return nil, err
	}

	if err := ValidateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	if data == nil {
		return nil, utils.NewValidationError("modification data cannot be nil")
	}

	s.LogOperation("Adding modification", fmt.Sprintf("user_id: %d, entity_type: %s, entity_id: %d", userID, entityType, entityID))

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modification data: %v", err)
	}

	query := `
		INSERT INTO modifications (user_id, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := s.db.Exec(query, userID, string(entityType), entityID, jsonData, now)
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

	s.LogOperation("Added modification", fmt.Sprintf("id: %d", id))
	return mod, nil
}

// Get retrieves a specific modification by ID
func (s *Modifications) Get(id int64) (*models.Modification[any], error) {
	if err := ValidateID(id, "modification"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting modification", id)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE id = ?
	`

	row := s.db.QueryRow(query, id)

	mod, err := ScanSingleRow(row, ScanModification, "modifications")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

// List retrieves all modifications for a specific entity
func (s *Modifications) List(
	entityType ModificationType, entityID int64, limit, offset int,
) ([]*models.Modification[any], error) {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	if limit > 0 { // Only validate pagination if limit is specified
		if err := ValidatePagination(limit, offset); err != nil {
			return nil, err
		}
	}

	s.LogOperation("Listing modifications", fmt.Sprintf("entity_type: %s, entity_id: %d, limit: %d, offset: %d", entityType, entityID, limit, offset))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, string(entityType), entityID, limit, offset)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	modifications, err := ScanModificationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Listed modifications", fmt.Sprintf("count: %d", len(modifications)))
	return modifications, nil
}

// ListAll retrieves all modifications for a specific entity without pagination
func (s *Modifications) ListAll(
	entityType ModificationType, entityID int64,
) ([]*models.Modification[any], error) {
	return s.List(entityType, entityID, -1, 0)
}

// Count returns the total number of modifications for a specific entity
func (s *Modifications) Count(entityType ModificationType, entityID int64) (int64, error) {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return 0, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return 0, err
	}

	s.LogOperation("Counting modifications", fmt.Sprintf("entity_type: %s, entity_id: %d", entityType, entityID))

	query := `
		SELECT COUNT(*)
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
	`

	var count int64
	err := s.db.QueryRow(query, string(entityType), entityID).Scan(&count)
	if err != nil {
		return 0, s.HandleSelectError(err, "modifications")
	}

	s.LogOperation("Counted modifications", fmt.Sprintf("count: %d", count))
	return count, nil
}

// GetLatest retrieves the most recent modification for a specific entity
func (s *Modifications) GetLatest(
	entityType ModificationType, entityID int64,
) (*models.Modification[any], error) {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting latest modification", fmt.Sprintf("entity_type: %s, entity_id: %d", entityType, entityID))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := s.db.QueryRow(query, string(entityType), entityID)

	mod, err := ScanSingleRow(row, ScanModification, "modifications")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

// GetOldest retrieves the oldest modification for a specific entity
func (s *Modifications) GetOldest(
	entityType ModificationType, entityID int64,
) (*models.Modification[any], error) {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting oldest modification", fmt.Sprintf("entity_type: %s, entity_id: %d", entityType, entityID))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
		LIMIT 1
	`

	row := s.db.QueryRow(query, string(entityType), entityID)

	mod, err := ScanSingleRow(row, ScanModification, "modifications")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, err
	}

	return mod, nil
}

// Delete removes a specific modification by ID
func (s *Modifications) Delete(id int64) error {
	if err := ValidateID(id, "modification"); err != nil {
		return err
	}

	s.LogOperation("Deleting modification", id)

	query := `DELETE FROM modifications WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return s.HandleDeleteError(err, "modifications")
	}

	if err := s.CheckRowsAffected(result, "modification", id); err != nil {
		return err
	}

	s.LogOperation("Deleted modification", id)
	return nil
}

// DeleteAll removes all modifications for a specific entity
func (s *Modifications) DeleteAll(entityType ModificationType, entityID int64) error {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return err
	}

	s.log.Info("Deleting all modifications: entity_type=%s, entity_id=%d", entityType, entityID)

	query := `DELETE FROM modifications WHERE entity_type = ? AND entity_id = ?`
	result, err := s.db.Exec(query, string(entityType), entityID)
	if err != nil {
		return s.HandleDeleteError(err, "modifications")
	}

	rowsAffected, err := s.GetRowsAffected(result, "delete all modifications")
	if err != nil {
		return err
	}

	s.log.Info("Successfully deleted %d modifications", rowsAffected)
	return nil
}

// GetByUser retrieves all modifications made by a specific user
func (s *Modifications) GetByUser(
	userID int64, limit, offset int,
) ([]*models.Modification[any], error) {
	if err := ValidateID(userID, "user"); err != nil {
		return nil, err
	}

	if err := ValidatePagination(limit, offset); err != nil {
		return nil, err
	}

	s.LogOperation("Getting modifications by user", fmt.Sprintf("user_id: %d, limit: %d, offset: %d", userID, limit, offset))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	modifications, err := ScanModificationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Found modifications by user", fmt.Sprintf("count: %d", len(modifications)))
	return modifications, nil
}

// GetByDateRange retrieves modifications within a specific date range
func (s *Modifications) GetByDateRange(
	entityType ModificationType, entityID int64, from, to time.Time,
) ([]*models.Modification[any], error) {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	if from.After(to) {
		return nil, utils.NewValidationError("from date must be before to date")
	}

	s.LogOperation("Getting modifications by date range",
		fmt.Sprintf("entity_type: %s, entity_id: %d, from: %s, to: %s",
			entityType, entityID, from.Format(time.RFC3339), to.Format(time.RFC3339)))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ? AND created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, string(entityType), entityID, from, to)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	modifications, err := ScanModificationsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Found modifications by date range", fmt.Sprintf("count: %d", len(modifications)))
	return modifications, nil
}

// Helper methods for specific entity types

// AddTroubleReportMod adds a modification for a trouble report
func (s *Modifications) AddTroubleReportMod(userID, reportID int64, data any) error {
	_, err := s.Add(userID, ModificationTypeTroubleReport, reportID, data)
	return err
}

// AddMetalSheetMod adds a modification for a metal sheet
func (s *Modifications) AddMetalSheetMod(userID, sheetID int64, data any) error {
	_, err := s.Add(userID, ModificationTypeMetalSheet, sheetID, data)
	return err
}

// AddToolMod adds a modification for a tool
func (s *Modifications) AddToolMod(userID, toolID int64, data any) error {
	_, err := s.Add(userID, ModificationTypeTool, toolID, data)
	return err
}

// AddPressCycleMod adds a modification for a press cycle
func (s *Modifications) AddPressCycleMod(userID, cycleID int64, data any) error {
	_, err := s.Add(userID, ModificationTypePressCycle, cycleID, data)
	return err
}

// AddUserMod adds a modification for a user
func (s *Modifications) AddUserMod(userID, targetUserID int64, data any) error {
	_, err := s.Add(userID, ModificationTypeUser, targetUserID, data)
	return err
}

// GetWithUser retrieves a modification with user information
func (s *Modifications) GetWithUser(id int64) (*ModificationWithUser, error) {
	if err := ValidateID(id, "modification"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting modification with user", id)

	query := `
		SELECT m.id, m.user_id, m.data, m.created_at, u.user_name
		FROM modifications m
		JOIN users u ON m.user_id = u.telegram_id
		WHERE m.id = ?
	`

	row := s.db.QueryRow(query, id)

	modWithUser := &ModificationWithUser{}
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
func (s *Modifications) ListWithUser(
	entityType ModificationType, entityID int64, limit, offset int,
) ([]*ModificationWithUser, error) {
	if err := ValidateModificationType(string(entityType)); err != nil {
		return nil, err
	}

	if err := ValidateID(entityID, "entity"); err != nil {
		return nil, err
	}

	if err := ValidatePagination(limit, offset); err != nil {
		return nil, err
	}

	s.LogOperation("Listing modifications with user", fmt.Sprintf("entity_type: %s, entity_id: %d", entityType, entityID))

	query := `
		SELECT m.id, m.user_id, m.data, m.created_at, u.user_name
		FROM modifications m
		JOIN users u ON m.user_id = u.telegram_id
		WHERE m.entity_type = ? AND m.entity_id = ?
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, string(entityType), entityID, limit, offset)
	if err != nil {
		return nil, s.HandleSelectError(err, "modifications")
	}
	defer rows.Close()

	var modifications []*ModificationWithUser
	for rows.Next() {
		modWithUser := &ModificationWithUser{}
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

	s.LogOperation("Listed modifications with user", fmt.Sprintf("count: %d", len(modifications)))
	return modifications, nil
}
