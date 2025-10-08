package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Attachments provides database operations for managing attachments with lazy loading.
type Attachments struct {
	*BaseService
}

// NewAttachments creates a new Service instance and initializes the database table.
func NewAttachments(db *sql.DB) *Attachments {
	base := NewBaseService(db, "Attachments")

	query := `
		CREATE TABLE IF NOT EXISTS attachments (
			id INTEGER NOT NULL,
			mime_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if err := base.CreateTable(query, "attachments"); err != nil {
		panic(err)
	}

	return &Attachments{
		BaseService: base,
	}
}

// List retrieves all attachments ordered by ID ascending.
func (a *Attachments) List() ([]*models.Attachment, error) {
	a.LogOperation("Listing attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, a.HandleSelectError(err, "attachments")
	}
	defer rows.Close()

	attachments, err := ScanAttachmentsFromRows(rows)
	if err != nil {
		return nil, err
	}

	a.LogOperation("Listed attachments", fmt.Sprintf("count: %d", len(attachments)))
	return attachments, nil
}

// Get retrieves a specific attachment by ID.
func (a *Attachments) Get(id int64) (*models.Attachment, error) {
	if err := ValidateID(id, "attachment"); err != nil {
		return nil, err
	}

	a.LogOperation("Getting attachment", id)

	query := `SELECT id, mime_type, data FROM attachments WHERE id = ?`
	row := a.db.QueryRow(query, id)

	attachment, err := ScanSingleRow(row, ScanAttachment, "attachments")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("attachment with ID %d not found", id))
		}
		return nil, err
	}

	return attachment, nil
}

// GetByIDs retrieves multiple attachments by their IDs in the order specified.
func (a *Attachments) GetByIDs(ids []int64) ([]*models.Attachment, error) {
	if len(ids) == 0 {
		return []*models.Attachment{}, nil
	}

	a.LogOperation("Getting attachments by IDs", fmt.Sprintf("count: %d", len(ids)))

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT id, mime_type, data FROM attachments
		WHERE id IN (%s)
		ORDER BY id ASC`,
		strings.Join(placeholders, ","),
	)

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, a.HandleSelectError(err, "attachments")
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap, err := ScanAttachmentsIntoMap(rows)
	if err != nil {
		return nil, err
	}

	// Return attachments in the order of the requested IDs
	var attachments []*models.Attachment
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachments = append(attachments, attachment)
		}
	}

	a.LogOperation("Found attachments by IDs", fmt.Sprintf("found: %d", len(attachments)))
	return attachments, nil
}

// Add creates a new attachment and returns its generated ID.
func (a *Attachments) Add(attachment *models.Attachment) (int64, error) {
	if err := ValidateAttachment(attachment); err != nil {
		return 0, err
	}

	// Call the model's validate method for additional checks
	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	a.LogOperation("Adding attachment", fmt.Sprintf("mime_type: %s, size: %d bytes", attachment.MimeType, len(attachment.Data)))

	query := `INSERT INTO attachments (mime_type, data) VALUES (?, ?)`
	result, err := a.db.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, a.HandleInsertError(err, "attachments")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, a.HandleInsertError(err, "attachments")
	}

	a.LogOperation("Added attachment", fmt.Sprintf("id: %d", id))
	return id, nil
}

// Update modifies an existing attachment.
func (a *Attachments) Update(attachment *models.Attachment) error {
	if err := ValidateAttachment(attachment); err != nil {
		return err
	}

	// Call the model's validate method for additional checks
	if err := attachment.Validate(); err != nil {
		return err
	}

	// Convert string ID to int64
	var id int64
	if _, err := fmt.Sscanf(attachment.ID, "%d", &id); err != nil {
		return utils.NewValidationError("invalid attachment ID format")
	}

	if err := ValidateID(id, "attachment"); err != nil {
		return err
	}

	a.LogOperation("Updating attachment", fmt.Sprintf("id: %d", id))

	query := `UPDATE attachments SET mime_type = ?, data = ? WHERE id = ?`
	result, err := a.db.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return a.HandleUpdateError(err, "attachments")
	}

	if err := a.CheckRowsAffected(result, "attachment", id); err != nil {
		return err
	}

	a.LogOperation("Updated attachment", fmt.Sprintf("id: %d", id))
	return nil
}

// Delete deletes an attachment by ID.
func (a *Attachments) Delete(id int64) error {
	if err := ValidateID(id, "attachment"); err != nil {
		return err
	}

	a.LogOperation("Deleting attachment", id)

	query := `DELETE FROM attachments WHERE id = ?`
	result, err := a.db.Exec(query, id)
	if err != nil {
		return a.HandleDeleteError(err, "attachments")
	}

	if err := a.CheckRowsAffected(result, "attachment", id); err != nil {
		return err
	}

	a.LogOperation("Deleted attachment", id)
	return nil
}
