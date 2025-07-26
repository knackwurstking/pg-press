// Package database provides attachment management with lazy loading.
package database

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pg-vis/internal/logger"
)

const (
	createAttachmentsTableQuery = `
		CREATE TABLE IF NOT EXISTS attachments (
			id INTEGER NOT NULL,
			mime_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	selectAllAttachmentsQuery      = `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	selectAttachmentByIDQuery      = `SELECT id, mime_type, data FROM attachments WHERE id = ?`
	selectAttachmentsByIDsQuery    = `SELECT id, mime_type, data FROM attachments WHERE id IN (%s) ORDER BY id ASC`
	insertAttachmentQuery          = `INSERT INTO attachments (mime_type, data) VALUES (?, ?)`
	updateAttachmentQuery          = `UPDATE attachments SET mime_type = ?, data = ? WHERE id = ?`
	deleteAttachmentQuery          = `DELETE FROM attachments WHERE id = ?`
	deleteOrphanedAttachmentsQuery = `DELETE FROM attachments WHERE id NOT IN (
		SELECT DISTINCT json_extract(value, '$')
		FROM trouble_reports, json_each(linked_attachments)
		WHERE json_valid(linked_attachments)
	)`
)

// Attachments provides database operations for managing attachments with lazy loading.
type Attachments struct {
	db *sql.DB
}

// NewAttachments creates a new Attachments instance and initializes the database table.
func NewAttachments(db *sql.DB) *Attachments {
	if _, err := db.Exec(createAttachmentsTableQuery); err != nil {
		panic(NewDatabaseError(
			"create_table",
			"attachments",
			"failed to create attachments table",
			err,
		))
	}

	return &Attachments{
		db: db,
	}
}

// List retrieves all attachments ordered by ID ascending.
func (a *Attachments) List() ([]*Attachment, error) {
	logger.TroubleReport().Debug("Listing all attachments")

	rows, err := a.db.Query(selectAllAttachmentsQuery)
	if err != nil {
		return nil, NewDatabaseError("select", "attachments",
			"failed to query attachments", err)
	}
	defer rows.Close()

	var attachments []*Attachment

	for rows.Next() {
		attachment, err := a.scanAttachment(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan attachment")
		}
		attachments = append(attachments, attachment)
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "attachments",
			"error iterating over rows", err)
	}

	return attachments, nil
}

// Get retrieves a specific attachment by ID.
func (a *Attachments) Get(id int64) (*Attachment, error) {
	logger.TroubleReport().Debug("Getting attachment, id: %d", id)

	row := a.db.QueryRow(selectAttachmentByIDQuery, id)

	attachment, err := a.scanAttachmentRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, NewDatabaseError("select", "attachments",
			fmt.Sprintf("failed to get attachment with ID %d", id), err)
	}

	return attachment, nil
}

// GetByIDs retrieves multiple attachments by their IDs in the order specified.
func (a *Attachments) GetByIDs(ids []int64) ([]*Attachment, error) {
	if len(ids) == 0 {
		return []*Attachment{}, nil
	}

	logger.TroubleReport().Debug("Getting attachments by IDs: %v", ids)

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(selectAttachmentsByIDsQuery,
		joinStrings(placeholders, ","))

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, NewDatabaseError("select", "attachments",
			"failed to query attachments by IDs", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap := make(map[int64]*Attachment)

	for rows.Next() {
		attachment, err := a.scanAttachment(rows)
		if err != nil {
			return nil, WrapError(err, "failed to scan attachment")
		}
		attachmentMap[attachment.GetID()] = attachment
	}

	if err := rows.Err(); err != nil {
		return nil, NewDatabaseError("select", "attachments",
			"error iterating over rows", err)
	}

	// Return attachments in the order of the requested IDs
	var attachments []*Attachment
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

// Add creates a new attachment and returns its generated ID.
func (a *Attachments) Add(attachment *Attachment) (int64, error) {
	logger.TroubleReport().Debug("Adding attachment: %s", attachment.String())

	if attachment == nil {
		return 0, NewValidationError("attachment", "attachment cannot be nil", nil)
	}

	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	result, err := a.db.Exec(insertAttachmentQuery, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, NewDatabaseError("insert", "attachments",
			"failed to insert attachment", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, NewDatabaseError("insert", "attachments",
			"failed to get last insert ID", err)
	}

	return id, nil
}

// Update modifies an existing attachment.
func (a *Attachments) Update(id int64, attachment *Attachment) error {
	logger.TroubleReport().Debug("Updating attachment, id: %d", id)

	if attachment == nil {
		return NewValidationError("attachment", "attachment cannot be nil", nil)
	}

	if err := attachment.Validate(); err != nil {
		return err
	}

	result, err := a.db.Exec(updateAttachmentQuery, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return NewDatabaseError("update", "attachments",
			fmt.Sprintf("failed to update attachment with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("update", "attachments",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Remove deletes an attachment by ID.
func (a *Attachments) Remove(id int64) error {
	logger.TroubleReport().Debug("Removing attachment, id: %d", id)

	result, err := a.db.Exec(deleteAttachmentQuery, id)
	if err != nil {
		return NewDatabaseError("delete", "attachments",
			fmt.Sprintf("failed to delete attachment with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewDatabaseError("delete", "attachments",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// CleanupOrphaned removes attachments that are not referenced by any trouble report.
func (a *Attachments) CleanupOrphaned() (int64, error) {
	logger.TroubleReport().Info("Cleaning up orphaned attachments")

	result, err := a.db.Exec(deleteOrphanedAttachmentsQuery)
	if err != nil {
		return 0, NewDatabaseError("delete", "attachments",
			"failed to cleanup orphaned attachments", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, NewDatabaseError("delete", "attachments",
			"failed to get rows affected", err)
	}

	logger.TroubleReport().Info("Cleaned up %d orphaned attachments", rowsAffected)
	return rowsAffected, nil
}

func (a *Attachments) scanAttachment(rows *sql.Rows) (*Attachment, error) {
	attachment := &Attachment{}
	var id int64

	if err := rows.Scan(&id, &attachment.MimeType, &attachment.Data); err != nil {
		return nil, NewDatabaseError("scan", "attachments",
			"failed to scan row", err)
	}

	// Set the ID using string conversion to maintain compatibility
	attachment.ID = fmt.Sprintf("%d", id)

	return attachment, nil
}

func (a *Attachments) scanAttachmentRow(row *sql.Row) (*Attachment, error) {
	attachment := &Attachment{}
	var id int64

	if err := row.Scan(&id, &attachment.MimeType, &attachment.Data); err != nil {
		return nil, err
	}

	// Set the ID using string conversion to maintain compatibility
	attachment.ID = fmt.Sprintf("%d", id)

	return attachment, nil
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	var result string
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
