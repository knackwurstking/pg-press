// Package database provides attachment management with lazy loading.
package database

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/dberror"
	"github.com/knackwurstking/pgpress/internal/dbutils"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/internal/models"
)

// Attachments provides database operations for managing attachments with lazy loading.
type Attachments struct {
	db *sql.DB
}

var _ DataOperations[*models.Attachment] = (*Attachments)(nil)

// NewAttachments creates a new Attachments instance and initializes the database table.
func NewAttachments(db *sql.DB) *Attachments {
	query := `
		CREATE TABLE IF NOT EXISTS attachments (
			id INTEGER NOT NULL,
			mime_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(dberror.NewDatabaseError(
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
func (a *Attachments) List() ([]*models.Attachment, error) {
	logger.DBAttachments().Debug("Listing all attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "attachments",
			"failed to query attachments", err)
	}
	defer rows.Close()

	var attachments []*models.Attachment

	for rows.Next() {
		attachment, err := a.scanAttachment(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan attachment")
		}
		attachments = append(attachments, attachment)
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "attachments",
			"error iterating over rows", err)
	}

	return attachments, nil
}

// Get retrieves a specific attachment by ID.
func (a *Attachments) Get(id int64) (*models.Attachment, error) {
	logger.DBAttachments().Debug("Getting attachment, id: %d", id)

	query := `SELECT id, mime_type, data FROM attachments WHERE id = ?`
	row := a.db.QueryRow(query, id)
	attachment, err := a.scanAttachment(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dberror.ErrNotFound
		}
		return nil, dberror.NewDatabaseError("select", "attachments",
			fmt.Sprintf("failed to get attachment with ID %d", id), err)
	}

	return attachment, nil
}

// GetByIDs retrieves multiple attachments by their IDs in the order specified.
func (a *Attachments) GetByIDs(ids []int64) ([]*models.Attachment, error) {
	if len(ids) == 0 {
		return []*models.Attachment{}, nil
	}

	logger.DBAttachments().Debug("Getting attachments by IDs: %v", ids)

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`
			SELECT id, mime_type, data FROM attachments
			WHERE id IN (%s)
			ORDER BY id ASC
		`,
		dbutils.JoinStrings(placeholders, ","),
	)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "attachments",
			"failed to query attachments by IDs", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap := make(map[int64]*models.Attachment)

	for rows.Next() {
		attachment, err := a.scanAttachment(rows)
		if err != nil {
			return nil, dberror.WrapError(err, "failed to scan attachment")
		}
		attachmentMap[attachment.GetID()] = attachment
	}

	if err := rows.Err(); err != nil {
		return nil, dberror.NewDatabaseError("select", "attachments",
			"error iterating over rows", err)
	}

	// Return attachments in the order of the requested IDs
	var attachments []*models.Attachment
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

// Add creates a new attachment and returns its generated ID.
func (a *Attachments) Add(attachment *models.Attachment, _ *models.User) (int64, error) {
	logger.DBAttachments().Debug("Adding attachment: %s", attachment.String())

	if attachment == nil {
		return 0, dberror.NewValidationError("attachment", "attachment cannot be nil", nil)
	}

	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	query := `INSERT INTO attachments (mime_type, data) VALUES (?, ?)`
	result, err := a.db.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "attachments",
			"failed to insert attachment", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, dberror.NewDatabaseError("insert", "attachments",
			"failed to get last insert ID", err)
	}

	return id, nil
}

// Update modifies an existing attachment.
func (a *Attachments) Update(attachment *models.Attachment, _ *models.User) error {
	id := attachment.GetID()
	logger.DBAttachments().Debug("Updating attachment, id: %d", id)

	if attachment == nil {
		return dberror.NewValidationError("attachment", "attachment cannot be nil", nil)
	}

	if err := attachment.Validate(); err != nil {
		return err
	}

	query := `UPDATE attachments SET mime_type = ?, data = ? WHERE id = ?`
	result, err := a.db.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return dberror.NewDatabaseError("update", "attachments",
			fmt.Sprintf("failed to update attachment with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberror.NewDatabaseError("update", "attachments",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return dberror.ErrNotFound
	}

	return nil
}

// Delete deletes an attachment by ID.
func (a *Attachments) Delete(id int64, _ *models.User) error {
	logger.DBAttachments().Debug("Removing attachment, id: %d", id)

	query := `DELETE FROM attachments WHERE id = ?`
	result, err := a.db.Exec(query, id)
	if err != nil {
		return dberror.NewDatabaseError("delete", "attachments",
			fmt.Sprintf("failed to delete attachment with ID %d", id), err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return dberror.NewDatabaseError("delete", "attachments",
			"failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return dberror.ErrNotFound
	}

	return nil
}

// scannable is an interface that abstracts sql.Row and sql.Rows for scanning.
type scannable interface {
	Scan(dest ...any) error
}

func (a *Attachments) scanAttachment(scanner scannable) (*models.Attachment, error) {
	attachment := &models.Attachment{}
	var id int64

	if err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data); err != nil {
		// The `Get` method needs the original sql.ErrNoRows error
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, dberror.NewDatabaseError("scan", "attachments", "failed to scan row", err)
	}

	// Set the ID using string conversion to maintain compatibility
	attachment.ID = fmt.Sprintf("%d", id)

	return attachment, nil
}
