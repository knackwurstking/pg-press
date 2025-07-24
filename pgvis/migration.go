// Package pgvis provides database migration utilities.
package pgvis

import (
	"database/sql"
	"encoding/json"

	"github.com/knackwurstking/pg-vis/pgvis/logger"
)

// Migration handles database schema and data migrations.
type Migration struct {
	db *sql.DB
}

// NewMigration creates a new migration instance.
func NewMigration(db *sql.DB) *Migration {
	return &Migration{db: db}
}

// MigrateAttachmentsToSeparateTable migrates existing attachments from trouble_reports
// table to the new attachments table and updates trouble_reports to store attachment IDs.
func (m *Migration) MigrateAttachmentsToSeparateTable() error {
	logger.TroubleReport().Info("Starting migration of attachments to separate table")

	// Check if migration is needed by looking for BLOB data in linked_attachments
	needsMigration, err := m.checkIfMigrationNeeded()
	if err != nil {
		return WrapError(err, "failed to check if migration is needed")
	}

	if !needsMigration {
		logger.TroubleReport().Info("Migration not needed - attachments already migrated")
		return nil
	}

	// Start transaction for migration
	tx, err := m.db.Begin()
	if err != nil {
		return NewDatabaseError("transaction", "migration",
			"failed to start transaction", err)
	}
	defer tx.Rollback()

	// Get all trouble reports with their current attachment data
	rows, err := tx.Query(`
		SELECT id, title, content, linked_attachments, mods
		FROM trouble_reports
		WHERE linked_attachments != '[]' AND linked_attachments != ''
	`)
	if err != nil {
		return NewDatabaseError("select", "trouble_reports",
			"failed to query trouble reports for migration", err)
	}
	defer rows.Close()

	var migratedCount int

	for rows.Next() {
		var reportID int64
		var title, content string
		var linkedAttachmentsJSON, modsJSON []byte

		if err := rows.Scan(&reportID, &title, &content, &linkedAttachmentsJSON, &modsJSON); err != nil {
			logger.TroubleReport().Error("Failed to scan trouble report %d: %v", reportID, err)
			continue
		}

		// Parse existing attachments
		var existingAttachments []*Attachment
		if err := json.Unmarshal(linkedAttachmentsJSON, &existingAttachments); err != nil {
			logger.TroubleReport().Error("Failed to unmarshal attachments for report %d: %v", reportID, err)
			continue
		}

		// Skip if no attachments or if already migrated (contains only IDs)
		if len(existingAttachments) == 0 {
			continue
		}

		// Check if already migrated by looking at first attachment structure
		if len(existingAttachments) > 0 && existingAttachments[0].Data == nil {
			logger.TroubleReport().Debug("Report %d already migrated", reportID)
			continue
		}

		// Migrate attachments to separate table
		var attachmentIDs []int64
		for _, attachment := range existingAttachments {
			if attachment == nil || attachment.Data == nil {
				continue
			}

			// Insert attachment into attachments table
			result, err := tx.Exec(
				"INSERT INTO attachments (mime_type, data) VALUES (?, ?)",
				attachment.MimeType, attachment.Data,
			)
			if err != nil {
				logger.TroubleReport().Error("Failed to insert attachment for report %d: %v", reportID, err)
				continue
			}

			attachmentID, err := result.LastInsertId()
			if err != nil {
				logger.TroubleReport().Error("Failed to get attachment ID for report %d: %v", reportID, err)
				continue
			}

			attachmentIDs = append(attachmentIDs, attachmentID)
		}

		if len(attachmentIDs) == 0 {
			continue
		}

		// Update trouble report with attachment IDs
		attachmentIDsJSON, err := json.Marshal(attachmentIDs)
		if err != nil {
			logger.TroubleReport().Error("Failed to marshal attachment IDs for report %d: %v", reportID, err)
			continue
		}

		// Parse and update mods data
		var mods Mods[TroubleReportMod]
		if err := json.Unmarshal(modsJSON, &mods); err != nil {
			logger.TroubleReport().Error("Failed to unmarshal mods for report %d: %v", reportID, err)
			continue
		}

		// Update attachment references in mods
		for i := range mods {
			mods[i].Data.LinkedAttachments = attachmentIDs
		}

		updatedModsJSON, err := json.Marshal(mods)
		if err != nil {
			logger.TroubleReport().Error("Failed to marshal updated mods for report %d: %v", reportID, err)
			continue
		}

		// Update trouble report
		_, err = tx.Exec(
			"UPDATE trouble_reports SET linked_attachments = ?, mods = ? WHERE id = ?",
			string(attachmentIDsJSON), updatedModsJSON, reportID,
		)
		if err != nil {
			logger.TroubleReport().Error("Failed to update trouble report %d: %v", reportID, err)
			continue
		}

		migratedCount++
		logger.TroubleReport().Debug("Migrated report %d with %d attachments", reportID, len(attachmentIDs))
	}

	if err := rows.Err(); err != nil {
		return NewDatabaseError("select", "trouble_reports",
			"error iterating over migration rows", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return NewDatabaseError("transaction", "migration",
			"failed to commit migration transaction", err)
	}

	logger.TroubleReport().Info("Migration completed successfully. Migrated %d trouble reports", migratedCount)
	return nil
}

// checkIfMigrationNeeded checks if there are any trouble reports with BLOB attachment data
func (m *Migration) checkIfMigrationNeeded() (bool, error) {
	// Check if there are any trouble reports with non-empty linked_attachments that contain BLOB data
	var count int
	err := m.db.QueryRow(`
		SELECT COUNT(*)
		FROM trouble_reports
		WHERE linked_attachments != '[]'
		  AND linked_attachments != ''
		  AND linked_attachments NOT LIKE '[%]'
	`).Scan(&count)

	if err != nil {
		// If there's an error, assume migration is needed to be safe
		logger.TroubleReport().Warn("Could not determine if migration is needed: %v", err)
		return true, nil
	}

	// Also check for any records that might have the old BLOB format
	var blobCount int
	err = m.db.QueryRow(`
		SELECT COUNT(*)
		FROM trouble_reports
		WHERE typeof(linked_attachments) = 'blob'
	`).Scan(&blobCount)

	if err != nil {
		// If there's an error checking for blobs, check for any non-empty attachments
		logger.TroubleReport().Debug("Could not check for BLOB attachments: %v", err)
		return count > 0, nil
	}

	return count > 0 || blobCount > 0, nil
}

// MigrateAttachmentColumnType migrates the linked_attachments column from BLOB to TEXT
func (m *Migration) MigrateAttachmentColumnType() error {
	logger.TroubleReport().Info("Migrating linked_attachments column from BLOB to TEXT")

	// Check current column type
	var columnType string
	err := m.db.QueryRow(`
		SELECT type FROM pragma_table_info('trouble_reports')
		WHERE name = 'linked_attachments'
	`).Scan(&columnType)

	if err != nil {
		return NewDatabaseError("pragma", "trouble_reports",
			"failed to get column type", err)
	}

	if columnType == "TEXT" {
		logger.TroubleReport().Info("Column already migrated to TEXT")
		return nil
	}

	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return NewDatabaseError("transaction", "migration",
			"failed to start transaction", err)
	}
	defer tx.Rollback()

	// Create new table with correct schema
	_, err = tx.Exec(`
		CREATE TABLE trouble_reports_new (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		)
	`)
	if err != nil {
		return NewDatabaseError("create", "trouble_reports_new",
			"failed to create new table", err)
	}

	// Copy data, converting BLOB to TEXT where necessary
	_, err = tx.Exec(`
		INSERT INTO trouble_reports_new (id, title, content, linked_attachments, mods)
		SELECT id, title, content,
			CASE
				WHEN typeof(linked_attachments) = 'blob' THEN linked_attachments
				ELSE linked_attachments
			END,
			mods
		FROM trouble_reports
	`)
	if err != nil {
		return NewDatabaseError("insert", "trouble_reports_new",
			"failed to copy data", err)
	}

	// Drop old table and rename new one
	_, err = tx.Exec("DROP TABLE trouble_reports")
	if err != nil {
		return NewDatabaseError("drop", "trouble_reports",
			"failed to drop old table", err)
	}

	_, err = tx.Exec("ALTER TABLE trouble_reports_new RENAME TO trouble_reports")
	if err != nil {
		return NewDatabaseError("rename", "trouble_reports_new",
			"failed to rename table", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return NewDatabaseError("transaction", "migration",
			"failed to commit column migration", err)
	}

	logger.TroubleReport().Info("Column migration completed successfully")
	return nil
}

// RunAllMigrations runs all necessary migrations in the correct order
func (m *Migration) RunAllMigrations() error {
	logger.TroubleReport().Info("Running all migrations")

	// First, ensure tables exist (this is handled by the constructors)

	// Migrate column type if needed
	if err := m.MigrateAttachmentColumnType(); err != nil {
		return WrapError(err, "failed to migrate column type")
	}

	// Migrate attachment data
	if err := m.MigrateAttachmentsToSeparateTable(); err != nil {
		return WrapError(err, "failed to migrate attachments to separate table")
	}

	logger.TroubleReport().Info("All migrations completed successfully")
	return nil
}
