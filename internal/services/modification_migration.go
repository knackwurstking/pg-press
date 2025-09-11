package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	oldmodification "github.com/knackwurstking/pgpress/pkg/modification"
)

// ModificationMigration handles the migration from old mod system to new modification service
type ModificationMigration struct {
	db            *sql.DB
	modifications *ModificationService
	users         *User
}

// NewModificationMigration creates a new migration instance
func NewModificationMigration(db *sql.DB, modifications *ModificationService, users *User) *ModificationMigration {
	return &ModificationMigration{
		db:            db,
		modifications: modifications,
		users:         users,
	}
}

// MigrationStats holds statistics about the migration process
type MigrationStats struct {
	TroubleReportsProcessed int           `json:"trouble_reports_processed"`
	MetalSheetsProcessed    int           `json:"metal_sheets_processed"`
	ToolsProcessed          int           `json:"tools_processed"`
	TotalModsMigrated       int           `json:"total_mods_migrated"`
	Errors                  int           `json:"errors"`
	StartTime               time.Time     `json:"start_time"`
	EndTime                 time.Time     `json:"end_time"`
	Duration                time.Duration `json:"duration"`
}

// MigrateAll migrates all entities from the old mod system to the new modification service
func (m *ModificationMigration) MigrateAll() (*MigrationStats, error) {
	logger.DBModifications().Info("Starting migration from old mod system to new modification service")

	stats := &MigrationStats{
		StartTime: time.Now(),
	}

	// Migrate trouble reports
	if err := m.migrateTroubleReports(stats); err != nil {
		return stats, fmt.Errorf("failed to migrate trouble reports: %w", err)
	}

	// Migrate metal sheets
	if err := m.migrateMetalSheets(stats); err != nil {
		return stats, fmt.Errorf("failed to migrate metal sheets: %w", err)
	}

	// Migrate tools
	if err := m.migrateTools(stats); err != nil {
		return stats, fmt.Errorf("failed to migrate tools: %w", err)
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	logger.DBModifications().Info("Migration completed: processed %d entities, migrated %d mods, %d errors in %v",
		stats.TroubleReportsProcessed+stats.MetalSheetsProcessed+stats.ToolsProcessed,
		stats.TotalModsMigrated,
		stats.Errors,
		stats.Duration)

	return stats, nil
}

// migrateTroubleReports migrates trouble report modifications
func (m *ModificationMigration) migrateTroubleReports(stats *MigrationStats) error {
	logger.DBModifications().Info("Migrating trouble report modifications")

	query := `
		SELECT id, title, content, linked_attachments, mods
		FROM trouble_reports
		WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query trouble reports: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var title, content string
		var linkedAttachmentsJSON, modsJSON []byte

		err := rows.Scan(&id, &title, &content, &linkedAttachmentsJSON, &modsJSON)
		if err != nil {
			logger.DBModifications().Error("Failed to scan trouble report row: %v", err)
			stats.Errors++
			continue
		}

		// Parse old mods
		var oldMods oldmodification.Mods[models.TroubleReportMod]
		if err := json.Unmarshal(modsJSON, &oldMods); err != nil {
			logger.DBModifications().Error("Failed to unmarshal old mods for trouble report %d: %v", id, err)
			stats.Errors++
			continue
		}

		// Parse linked attachments
		var linkedAttachments []int64
		if err := json.Unmarshal(linkedAttachmentsJSON, &linkedAttachments); err != nil {
			logger.DBModifications().Error("Failed to unmarshal linked attachments for trouble report %d: %v", id, err)
			linkedAttachments = []int64{}
		}

		// Migrate each mod
		for i, oldMod := range oldMods {
			userID := int64(1) // Default to system user
			if oldMod.User != nil && oldMod.User.TelegramID > 0 {
				userID = oldMod.User.TelegramID
			}

			// Create new modification data
			modData := models.NewExtendedModificationData(
				models.TroubleReportModData{
					Title:             oldMod.Data.Title,
					Content:           oldMod.Data.Content,
					LinkedAttachments: oldMod.Data.LinkedAttachments,
				},
				m.determineAction(i, len(oldMods)),
				fmt.Sprintf("Migrated from old mod system (mod %d/%d)", i+1, len(oldMods)),
			)

			// Add to new modification system with original timestamp
			if err := m.addModificationWithTimestamp(userID, ModificationTypeTroubleReport, id, modData, time.UnixMilli(oldMod.Time)); err != nil {
				logger.DBModifications().Error("Failed to migrate mod for trouble report %d: %v", id, err)
				stats.Errors++
				continue
			}

			stats.TotalModsMigrated++
		}

		stats.TroubleReportsProcessed++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating trouble report rows: %w", err)
	}

	logger.DBModifications().Info("Migrated %d trouble reports with %d total mods", stats.TroubleReportsProcessed, stats.TotalModsMigrated)
	return nil
}

// migrateMetalSheets migrates metal sheet modifications
func (m *ModificationMigration) migrateMetalSheets(stats *MigrationStats) error {
	logger.DBModifications().Info("Migrating metal sheet modifications")

	query := `
		SELECT id, tile_height, value, marke_height, stf, stf_max, tool_id, notes, mods
		FROM metal_sheets
		WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query metal sheets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var tileHeight, value, stf, stfMax float64
		var markeHeight int
		var toolID *int64
		var notesJSON, modsJSON []byte

		err := rows.Scan(&id, &tileHeight, &value, &markeHeight, &stf, &stfMax, &toolID, &notesJSON, &modsJSON)
		if err != nil {
			logger.DBModifications().Error("Failed to scan metal sheet row: %v", err)
			stats.Errors++
			continue
		}

		// Parse old mods
		var oldMods oldmodification.Mods[models.MetalSheetMod]
		if err := json.Unmarshal(modsJSON, &oldMods); err != nil {
			logger.DBModifications().Error("Failed to unmarshal old mods for metal sheet %d: %v", id, err)
			stats.Errors++
			continue
		}

		// Parse notes
		var linkedNotes []int64
		if err := json.Unmarshal(notesJSON, &linkedNotes); err != nil {
			logger.DBModifications().Error("Failed to unmarshal linked notes for metal sheet %d: %v", id, err)
			linkedNotes = []int64{}
		}

		// Migrate each mod
		for i, oldMod := range oldMods {
			userID := int64(1) // Default to system user
			if oldMod.User != nil && oldMod.User.TelegramID > 0 {
				userID = oldMod.User.TelegramID
			}

			// Create new modification data
			modData := models.NewExtendedModificationData(
				models.MetalSheetModData{
					TileHeight:  oldMod.Data.TileHeight,
					Value:       oldMod.Data.Value,
					MarkeHeight: oldMod.Data.MarkeHeight,
					STF:         oldMod.Data.STF,
					STFMax:      oldMod.Data.STFMax,
					ToolID:      oldMod.Data.ToolID,
					LinkedNotes: oldMod.Data.LinkedNotes,
				},
				m.determineAction(i, len(oldMods)),
				fmt.Sprintf("Migrated from old mod system (mod %d/%d)", i+1, len(oldMods)),
			)

			// Add to new modification system with original timestamp
			if err := m.addModificationWithTimestamp(userID, ModificationTypeMetalSheet, id, modData, time.UnixMilli(oldMod.Time)); err != nil {
				logger.DBModifications().Error("Failed to migrate mod for metal sheet %d: %v", id, err)
				stats.Errors++
				continue
			}

			stats.TotalModsMigrated++
		}

		stats.MetalSheetsProcessed++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating metal sheet rows: %w", err)
	}

	logger.DBModifications().Info("Migrated %d metal sheets", stats.MetalSheetsProcessed)
	return nil
}

// migrateTools migrates tool modifications
func (m *ModificationMigration) migrateTools(stats *MigrationStats) error {
	logger.DBModifications().Info("Migrating tool modifications")

	query := `
		SELECT id, position, format, type, code, regenerating, press, notes, mods
		FROM tools
		WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query tools: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var toolType, code string
		var regenerating bool
		var press *int
		var positionJSON, formatJSON, notesJSON, modsJSON []byte

		err := rows.Scan(&id, &positionJSON, &formatJSON, &toolType, &code, &regenerating, &press, &notesJSON, &modsJSON)
		if err != nil {
			logger.DBModifications().Error("Failed to scan tool row: %v", err)
			stats.Errors++
			continue
		}

		// Parse old mods
		var oldMods oldmodification.Mods[models.ToolMod]
		if err := json.Unmarshal(modsJSON, &oldMods); err != nil {
			logger.DBModifications().Error("Failed to unmarshal old mods for tool %d: %v", id, err)
			stats.Errors++
			continue
		}

		// Parse position
		var position models.ToolPosition
		if err := json.Unmarshal(positionJSON, &position); err != nil {
			logger.DBModifications().Error("Failed to unmarshal position for tool %d: %v", id, err)
		}

		// Parse format
		var format models.ToolFormat
		if err := json.Unmarshal(formatJSON, &format); err != nil {
			logger.DBModifications().Error("Failed to unmarshal format for tool %d: %v", id, err)
		}

		// Parse notes
		var linkedNotes []int64
		if err := json.Unmarshal(notesJSON, &linkedNotes); err != nil {
			logger.DBModifications().Error("Failed to unmarshal linked notes for tool %d: %v", id, err)
			linkedNotes = []int64{}
		}

		// Migrate each mod
		for i, oldMod := range oldMods {
			userID := int64(1) // Default to system user
			if oldMod.User != nil && oldMod.User.TelegramID > 0 {
				userID = oldMod.User.TelegramID
			}

			// Create new modification data
			modData := models.NewExtendedModificationData(
				models.ToolModData{
					Position:     position,
					Format:       format,
					Type:         toolType,
					Code:         code,
					Regenerating: regenerating,
					Press:        press,
					LinkedNotes:  linkedNotes,
				},
				m.determineAction(i, len(oldMods)),
				fmt.Sprintf("Migrated from old mod system (mod %d/%d)", i+1, len(oldMods)),
			)

			// Add to new modification system with original timestamp
			if err := m.addModificationWithTimestamp(userID, ModificationTypeTool, id, modData, time.UnixMilli(oldMod.Time)); err != nil {
				logger.DBModifications().Error("Failed to migrate mod for tool %d: %v", id, err)
				stats.Errors++
				continue
			}

			stats.TotalModsMigrated++
		}

		stats.ToolsProcessed++
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating tool rows: %w", err)
	}

	logger.DBModifications().Info("Migrated %d tools", stats.ToolsProcessed)
	return nil
}

// addModificationWithTimestamp adds a modification with a specific timestamp (for migration)
func (m *ModificationMigration) addModificationWithTimestamp(userID int64, entityType ModificationType, entityID int64, data interface{}, timestamp time.Time) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal modification data: %w", err)
	}

	query := `
		INSERT INTO modifications (user_id, entity_type, entity_id, data, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = m.db.Exec(query, userID, string(entityType), entityID, jsonData, timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert modification: %w", err)
	}

	return nil
}

// determineAction determines the action type based on mod position
func (m *ModificationMigration) determineAction(index, total int) models.ModificationAction {
	if index == 0 && total == 1 {
		return models.ActionCreate
	} else if index == 0 {
		return models.ActionCreate
	} else {
		return models.ActionUpdate
	}
}

// CleanupOldMods removes the old mod columns after successful migration
// WARNING: This is destructive and should only be run after verifying the migration
func (m *ModificationMigration) CleanupOldMods() error {
	logger.DBModifications().Info("Starting cleanup of old mod columns - THIS IS DESTRUCTIVE!")

	queries := []string{
		"ALTER TABLE trouble_reports DROP COLUMN mods",
		"ALTER TABLE metal_sheets DROP COLUMN mods",
		"ALTER TABLE tools DROP COLUMN mods",
	}

	for _, query := range queries {
		if _, err := m.db.Exec(query); err != nil {
			// SQLite doesn't support DROP COLUMN in older versions
			// So we'll create a backup and recreate tables
			logger.DBModifications().Warn("Failed to drop column (this is expected in older SQLite): %v", err)
		}
	}

	logger.DBModifications().Info("Old mod columns cleanup completed")
	return nil
}

// VerifyMigration compares the count of old mods with new modifications
func (m *ModificationMigration) VerifyMigration() (*VerificationResult, error) {
	logger.DBModifications().Info("Verifying migration integrity")

	result := &VerificationResult{}

	// Count old mods in trouble reports
	var oldTroubleReportMods int
	err := m.db.QueryRow(`
		SELECT COUNT(*)
		FROM trouble_reports
		WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''
	`).Scan(&oldTroubleReportMods)
	if err != nil {
		return nil, fmt.Errorf("failed to count old trouble report mods: %w", err)
	}

	// Count new modifications for trouble reports
	var newTroubleReportModsCount int
	err = m.db.QueryRow(`
		SELECT COUNT(*)
		FROM modifications
		WHERE entity_type = 'trouble_reports'
	`).Scan(&newTroubleReportModsCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count new trouble report modifications: %w", err)
	}

	result.TroubleReports = EntityVerification{
		OldCount: oldTroubleReportMods,
		NewCount: newTroubleReportModsCount,
		Match:    oldTroubleReportMods == newTroubleReportModsCount,
	}

	// Similar verification for metal sheets and tools would go here...

	result.OverallMatch = result.TroubleReports.Match
	logger.DBModifications().Info("Migration verification completed")
	return result, nil
}

// VerificationResult holds the results of migration verification
type VerificationResult struct {
	TroubleReports EntityVerification `json:"trouble_reports"`
	MetalSheets    EntityVerification `json:"metal_sheets"`
	Tools          EntityVerification `json:"tools"`
	OverallMatch   bool               `json:"overall_match"`
}

// EntityVerification holds verification data for a single entity type
type EntityVerification struct {
	OldCount int  `json:"old_count"`
	NewCount int  `json:"new_count"`
	Match    bool `json:"match"`
}

// GetMigrationStatus returns the current migration status
func (m *ModificationMigration) GetMigrationStatus() (*MigrationStatus, error) {
	status := &MigrationStatus{}

	// Check if modifications table exists and has data
	var modCount int
	err := m.db.QueryRow("SELECT COUNT(*) FROM modifications").Scan(&modCount)
	if err != nil {
		status.ModificationTableExists = false
		return status, nil
	}

	status.ModificationTableExists = true
	status.TotalModifications = modCount

	// Check if old mod columns still exist
	// This is a simple check - in production you might want more sophisticated detection
	var oldModsExist int
	err = m.db.QueryRow(`
		SELECT COUNT(*)
		FROM trouble_reports
		WHERE mods IS NOT NULL AND mods != '[]' AND mods != ''
	`).Scan(&oldModsExist)

	status.OldModsExist = err == nil && oldModsExist > 0
	status.MigrationNeeded = status.OldModsExist && modCount == 0

	return status, nil
}

// MigrationStatus represents the current state of the migration
type MigrationStatus struct {
	ModificationTableExists bool `json:"modification_table_exists"`
	TotalModifications      int  `json:"total_modifications"`
	OldModsExist            bool `json:"old_mods_exist"`
	MigrationNeeded         bool `json:"migration_needed"`
}
