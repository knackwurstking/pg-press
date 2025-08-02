// Package database provides database migration utilities.
package database

import (
	"database/sql"
	"encoding/json"

	"github.com/knackwurstking/pgpress/internal/logger"
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

	// Check if there are any trouble reports at all
	var totalCount int
	err = m.db.QueryRow("SELECT COUNT(*) FROM trouble_reports").Scan(&totalCount)
	if err != nil {
		return NewDatabaseError("count", "trouble_reports",
			"failed to count trouble reports", err)
	}

	if totalCount == 0 {
		logger.TroubleReport().Info("No trouble reports found - skipping migration")
		return nil
	}

	logger.TroubleReport().Info("Found %d trouble reports, proceeding with migration", totalCount)

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

		// Check if already migrated by trying to unmarshal as array of IDs first
		var attachmentIDs []int64
		if err := json.Unmarshal(linkedAttachmentsJSON, &attachmentIDs); err == nil {
			// Already migrated - data is array of IDs
			logger.TroubleReport().Debug("Report %d already migrated (contains IDs)", reportID)
			continue
		}

		// Try to parse as old format (full attachment objects)
		var existingAttachments []*Attachment
		if err := json.Unmarshal(linkedAttachmentsJSON, &existingAttachments); err != nil {
			// Check if it's just an empty array or invalid JSON
			var rawArray []interface{}
			if err2 := json.Unmarshal(linkedAttachmentsJSON, &rawArray); err2 == nil {
				if len(rawArray) == 0 {
					logger.TroubleReport().Debug("Report %d has empty attachments array", reportID)
					continue
				}
				logger.TroubleReport().Error(
					"Report %d has unexpected attachment format: %s",
					reportID, string(linkedAttachmentsJSON))
			} else {
				logger.TroubleReport().Error(
					"Failed to unmarshal attachments for report %d: %v, data: %s",
					reportID, err, string(linkedAttachmentsJSON))
			}
			continue
		}

		// Skip if no attachments
		if len(existingAttachments) == 0 {
			continue
		}

		// Validate that we have old-format attachments with data
		hasOldFormatData := false
		for _, att := range existingAttachments {
			if att != nil && att.Data != nil && len(att.Data) > 0 {
				hasOldFormatData = true
				break
			}
		}

		if !hasOldFormatData {
			logger.TroubleReport().Debug("Report %d has no attachment data to migrate", reportID)
			continue
		}

		// Migrate attachments to separate table
		var newAttachmentIDs []int64
		for _, attachment := range existingAttachments {
			if attachment == nil || attachment.Data == nil || len(attachment.Data) == 0 {
				continue
			}

			// Validate attachment before migration
			if attachment.MimeType == "" {
				logger.TroubleReport().Warn(
					"Attachment for report %d has empty MIME type, setting default", reportID)
				attachment.MimeType = "application/octet-stream"
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

			newAttachmentIDs = append(newAttachmentIDs, attachmentID)
		}

		if len(newAttachmentIDs) == 0 {
			logger.TroubleReport().Debug("No valid attachments to migrate for report %d", reportID)
			continue
		}

		// Update trouble report with attachment IDs
		attachmentIDsJSON, err := json.Marshal(newAttachmentIDs)
		if err != nil {
			logger.TroubleReport().Error(
				"Failed to marshal attachment IDs for report %d: %v", reportID, err)
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
			mods[i].Data.LinkedAttachments = newAttachmentIDs
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
		logger.TroubleReport().Debug(
			"Migrated report %d with %d attachments", reportID, len(newAttachmentIDs))
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

	logger.TroubleReport().Info(
		"Migration completed successfully. Migrated %d trouble reports", migratedCount)
	return nil
}

// checkIfMigrationNeeded checks if there are any trouble reports with BLOB attachment data
func (m *Migration) checkIfMigrationNeeded() (bool, error) {
	// Get a sample of trouble reports with non-empty linked_attachments
	rows, err := m.db.Query(`
		SELECT linked_attachments
		FROM trouble_reports
		WHERE linked_attachments != '[]'
		  AND linked_attachments != ''
		  AND linked_attachments IS NOT NULL
		LIMIT 10
	`)
	if err != nil {
		logger.TroubleReport().Warn("Could not query for migration check: %v", err)
		return true, nil // Assume migration needed to be safe
	}
	defer rows.Close()

	needsMigration := false
	for rows.Next() {
		var linkedAttachmentsJSON []byte
		if err := rows.Scan(&linkedAttachmentsJSON); err != nil {
			continue
		}

		// Try to unmarshal as array of int64 (new format)
		var attachmentIDs []int64
		if err := json.Unmarshal(linkedAttachmentsJSON, &attachmentIDs); err == nil {
			// This is already in new format, continue checking others
			continue
		}

		// Try to unmarshal as array of attachment objects (old format)
		var attachments []*Attachment
		if err := json.Unmarshal(linkedAttachmentsJSON, &attachments); err == nil {
			// Check if any attachment has actual data (indicating old format)
			for _, att := range attachments {
				if att != nil && att.Data != nil && len(att.Data) > 0 {
					needsMigration = true
					break
				}
			}
			if needsMigration {
				break
			}
		} else {
			// Unknown format, assume migration needed
			logger.TroubleReport().Debug("Unknown attachment format found, assuming migration needed")
			needsMigration = true
			break
		}
	}

	if err := rows.Err(); err != nil {
		logger.TroubleReport().Warn("Error during migration check iteration: %v", err)
		return true, nil // Assume migration needed to be safe
	}

	return needsMigration, nil
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

// DiagnoseAttachmentData provides diagnostic information about attachment data in the database
func (m *Migration) DiagnoseAttachmentData() error {
	logger.TroubleReport().Info("Starting attachment data diagnosis")

	// Get sample of trouble reports with attachment data
	rows, err := m.db.Query(`
		SELECT id, title, linked_attachments
		FROM trouble_reports
		WHERE linked_attachments != '[]'
		  AND linked_attachments != ''
		  AND linked_attachments IS NOT NULL
		LIMIT 5
	`)
	if err != nil {
		return NewDatabaseError("select", "trouble_reports",
			"failed to query for diagnosis", err)
	}
	defer rows.Close()

	var reportCount int
	for rows.Next() {
		var reportID int64
		var title string
		var linkedAttachmentsJSON []byte

		if err := rows.Scan(&reportID, &title, &linkedAttachmentsJSON); err != nil {
			logger.TroubleReport().Error("Failed to scan report for diagnosis: %v", err)
			continue
		}

		reportCount++
		logger.TroubleReport().Info("=== Report %d: %s ===", reportID, title)
		logger.TroubleReport().Info("Raw attachment data: %s", string(linkedAttachmentsJSON))

		// Try different parsing approaches
		var attachmentIDs []int64
		if err := json.Unmarshal(linkedAttachmentsJSON, &attachmentIDs); err == nil {
			logger.TroubleReport().Info("✅ Successfully parsed as int64 array: %v", attachmentIDs)
		} else {
			logger.TroubleReport().Info("❌ Failed to parse as int64 array: %v", err)

			var attachments []*Attachment
			if err := json.Unmarshal(linkedAttachmentsJSON, &attachments); err == nil {
				logger.TroubleReport().Info(
					"✅ Successfully parsed as Attachment array: %d items", len(attachments))
				for i, att := range attachments {
					if att == nil {
						logger.TroubleReport().Info("  [%d] nil attachment", i)
						continue
					}
					dataSize := 0
					if att.Data != nil {
						dataSize = len(att.Data)
					}
					logger.TroubleReport().Info(
						"  [%d] ID: %s, MimeType: %s, DataSize: %d",
						i, att.ID, att.MimeType, dataSize)
				}
			} else {
				logger.TroubleReport().Info("❌ Failed to parse as Attachment array: %v", err)

				var rawData interface{}
				if err := json.Unmarshal(linkedAttachmentsJSON, &rawData); err == nil {
					logger.TroubleReport().Info("✅ Successfully parsed as generic interface: %T", rawData)
				} else {
					logger.TroubleReport().Info("❌ Failed to parse as any JSON: %v", err)
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		return NewDatabaseError("select", "trouble_reports",
			"error during diagnosis iteration", err)
	}

	if reportCount == 0 {
		logger.TroubleReport().Info("No trouble reports with attachment data found")
	} else {
		logger.TroubleReport().Info("Diagnosed %d trouble reports", reportCount)
	}

	// Check attachments table
	var attachmentCount int
	err = m.db.QueryRow("SELECT COUNT(*) FROM attachments").Scan(&attachmentCount)
	if err != nil {
		logger.TroubleReport().Warn("Could not count attachments table: %v", err)
	} else {
		logger.TroubleReport().Info("Attachments table contains %d records", attachmentCount)
	}

	logger.TroubleReport().Info("Attachment data diagnosis completed")
	return nil
}
