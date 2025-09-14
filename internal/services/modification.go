package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/logger"
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

// ModificationService handles database operations for modifications
type ModificationService struct {
	db *sql.DB
}

// NewModificationService creates a new ModificationService instance
func NewModificationService(db *sql.DB) *ModificationService {
	service := &ModificationService{
		db: db,
	}

	service.createTable()
	return service
}

// createTable creates the modifications table if it doesn't exist
func (s *ModificationService) createTable() {
	//const dropQuery string = `DROP TABLE IF EXISTS modifications;`
	//if _, err := s.db.Exec(dropQuery); err != nil {
	//	panic(fmt.Errorf("failed to drop feeds table: %v", err))
	//}

	query := `
		CREATE TABLE IF NOT EXISTS modifications (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id INTEGER NOT NULL,
			data BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(telegram_id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_modifications_entity ON modifications(entity_type, entity_id);
		CREATE INDEX IF NOT EXISTS idx_modifications_created_at ON modifications(created_at);
		CREATE INDEX IF NOT EXISTS idx_modifications_user_id ON modifications(user_id);
	`

	if _, err := s.db.Exec(query); err != nil {
		panic(fmt.Errorf("failed to create modifications table: %v", err))
	}
}

// Add creates a new modification record
func (s *ModificationService) Add(userID int64, entityType ModificationType, entityID int64, data interface{}) (*models.Modification[interface{}], error) {
	logger.DBModifications().Info("Adding modification: user_id=%d, entity_type=%s, entity_id=%d", userID, entityType, entityID)

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
		return nil, fmt.Errorf("failed to insert modification: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get modification ID: %v", err)
	}

	mod := &models.Modification[interface{}]{
		ID:        id,
		UserID:    userID,
		Data:      jsonData,
		CreatedAt: now,
	}

	logger.DBModifications().Info("Successfully added modification: id=%d", id)
	return mod, nil
}

// Get retrieves a specific modification by ID
func (s *ModificationService) Get(id int64) (*models.Modification[interface{}], error) {
	logger.DBModifications().Debug("Getting modification: id=%d", id)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE id = ?
	`

	row := s.db.QueryRow(query, id)

	mod := &models.Modification[interface{}]{}
	err := row.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, fmt.Errorf("failed to get modification: %v", err)
	}

	return mod, nil
}

// List retrieves all modifications for a specific entity
func (s *ModificationService) List(entityType ModificationType, entityID int64, limit, offset int) ([]*models.Modification[interface{}], error) {
	logger.DBModifications().Debug("Listing modifications: entity_type=%s, entity_id=%d, limit=%d, offset=%d", entityType, entityID, limit, offset)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, string(entityType), entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifications: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
	for rows.Next() {
		mod := &models.Modification[interface{}]{}
		err := rows.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification: %v", err)
		}
		modifications = append(modifications, mod)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating modifications: %v", err)
	}

	logger.DBModifications().Debug("Found %d modifications", len(modifications))
	return modifications, nil
}

// ListAll retrieves all modifications for a specific entity without pagination
func (s *ModificationService) ListAll(entityType ModificationType, entityID int64) ([]*models.Modification[interface{}], error) {
	return s.List(entityType, entityID, -1, 0)
}

// Count returns the total number of modifications for a specific entity
func (s *ModificationService) Count(entityType ModificationType, entityID int64) (int64, error) {
	logger.DBModifications().Debug("Counting modifications: entity_type=%s, entity_id=%d", entityType, entityID)

	query := `
		SELECT COUNT(*)
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
	`

	var count int64
	err := s.db.QueryRow(query, string(entityType), entityID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count modifications: %v", err)
	}

	return count, nil
}

// GetLatest retrieves the most recent modification for a specific entity
func (s *ModificationService) GetLatest(entityType ModificationType, entityID int64) (*models.Modification[interface{}], error) {
	logger.DBModifications().Debug("Getting latest modification: entity_type=%s, entity_id=%d", entityType, entityID)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := s.db.QueryRow(query, string(entityType), entityID)

	mod := &models.Modification[interface{}]{}
	err := row.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, fmt.Errorf("failed to get latest modification: %v", err)
	}

	return mod, nil
}

// GetOldest retrieves the oldest modification for a specific entity
func (s *ModificationService) GetOldest(entityType ModificationType, entityID int64) (*models.Modification[interface{}], error) {
	logger.DBModifications().Debug("Getting oldest modification: entity_type=%s, entity_id=%d", entityType, entityID)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ?
		ORDER BY created_at ASC
		LIMIT 1
	`

	row := s.db.QueryRow(query, string(entityType), entityID)

	mod := &models.Modification[interface{}]{}
	err := row.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError("modification")
		}
		return nil, fmt.Errorf("failed to get oldest modification: %v", err)
	}

	return mod, nil
}

// Delete removes a specific modification by ID
func (s *ModificationService) Delete(id int64) error {
	logger.DBModifications().Info("Deleting modification: id=%d", id)

	query := `DELETE FROM modifications WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete modification: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError("modification")
	}

	logger.DBModifications().Info("Successfully deleted modification: id=%d", id)
	return nil
}

// DeleteAll removes all modifications for a specific entity
func (s *ModificationService) DeleteAll(entityType ModificationType, entityID int64) error {
	logger.DBModifications().Info("Deleting all modifications: entity_type=%s, entity_id=%d", entityType, entityID)

	query := `DELETE FROM modifications WHERE entity_type = ? AND entity_id = ?`
	result, err := s.db.Exec(query, string(entityType), entityID)
	if err != nil {
		return fmt.Errorf("failed to delete modifications: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	logger.DBModifications().Info("Successfully deleted %d modifications", rowsAffected)
	return nil
}

// GetByUser retrieves all modifications made by a specific user
func (s *ModificationService) GetByUser(userID int64, limit, offset int) ([]*models.Modification[interface{}], error) {
	logger.DBModifications().Debug("Getting modifications by user: user_id=%d, limit=%d, offset=%d", userID, limit, offset)

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifications by user: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
	for rows.Next() {
		mod := &models.Modification[interface{}]{}
		err := rows.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification: %v", err)
		}
		modifications = append(modifications, mod)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating modifications: %v", err)
	}

	return modifications, nil
}

// GetByDateRange retrieves modifications within a specific date range
func (s *ModificationService) GetByDateRange(entityType ModificationType, entityID int64, from, to time.Time) ([]*models.Modification[interface{}], error) {
	logger.DBModifications().Debug("Getting modifications by date range: entity_type=%s, entity_id=%d, from=%s, to=%s",
		entityType, entityID, from.Format(time.RFC3339), to.Format(time.RFC3339))

	query := `
		SELECT id, user_id, data, created_at
		FROM modifications
		WHERE entity_type = ? AND entity_id = ? AND created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, string(entityType), entityID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifications by date range: %v", err)
	}
	defer rows.Close()

	var modifications []*models.Modification[interface{}]
	for rows.Next() {
		mod := &models.Modification[interface{}]{}
		err := rows.Scan(&mod.ID, &mod.UserID, &mod.Data, &mod.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan modification: %v", err)
		}
		modifications = append(modifications, mod)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating modifications: %v", err)
	}

	return modifications, nil
}

// Helper methods for specific entity types

// AddTroubleReportMod adds a modification for a trouble report
func (s *ModificationService) AddTroubleReportMod(userID, reportID int64, data interface{}) error {
	_, err := s.Add(userID, ModificationTypeTroubleReport, reportID, data)
	return err
}

// AddMetalSheetMod adds a modification for a metal sheet
func (s *ModificationService) AddMetalSheetMod(userID, sheetID int64, data interface{}) error {
	_, err := s.Add(userID, ModificationTypeMetalSheet, sheetID, data)
	return err
}

// AddToolMod adds a modification for a tool
func (s *ModificationService) AddToolMod(userID, toolID int64, data interface{}) error {
	_, err := s.Add(userID, ModificationTypeTool, toolID, data)
	return err
}

// AddPressCycleMod adds a modification for a press cycle
func (s *ModificationService) AddPressCycleMod(userID, cycleID int64, data interface{}) error {
	_, err := s.Add(userID, ModificationTypePressCycle, cycleID, data)
	return err
}

// AddUserMod adds a modification for a user
func (s *ModificationService) AddUserMod(userID, targetUserID int64, data interface{}) error {
	_, err := s.Add(userID, ModificationTypeUser, targetUserID, data)
	return err
}

// GetWithUser retrieves a modification with user information
func (s *ModificationService) GetWithUser(id int64) (*ModificationWithUser, error) {
	logger.DBModifications().Debug("Getting modification with user: id=%d", id)

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
		return nil, fmt.Errorf("failed to get modification with user: %v", err)
	}

	modWithUser.User.TelegramID = modWithUser.Modification.UserID
	return modWithUser, nil
}

// ModificationWithUser represents a modification with user information
type ModificationWithUser struct {
	Modification models.Modification[interface{}] `json:"modification"`
	User         models.User                      `json:"user"`
}

// ListWithUser retrieves modifications with user information for a specific entity
func (s *ModificationService) ListWithUser(entityType ModificationType, entityID int64, limit, offset int) ([]*ModificationWithUser, error) {
	logger.DBModifications().Debug("Listing modifications with user: entity_type=%s, entity_id=%d", entityType, entityID)

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
		return nil, fmt.Errorf("failed to query modifications with user: %v", err)
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
			return nil, fmt.Errorf("failed to scan modification with user: %v", err)
		}
		modWithUser.User.TelegramID = modWithUser.Modification.UserID
		modifications = append(modifications, modWithUser)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating modifications with user: %v", err)
	}

	return modifications, nil
}
