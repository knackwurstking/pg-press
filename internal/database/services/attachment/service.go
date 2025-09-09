// Package database provides attachment management with lazy loading.
package attachment

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	"github.com/knackwurstking/pgpress/internal/database/models/attachment"
	"github.com/knackwurstking/pgpress/internal/database/models/user"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Service provides database operations for managing attachments with lazy loading.
type Service struct {
	db *sql.DB
}

// New creates a new Service instance and initializes the database table.
func New(db *sql.DB) *Service {
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

	return &Service{
		db: db,
	}
}

// List retrieves all attachments ordered by ID ascending.
func (a *Service) List() ([]*attachment.Attachment, error) {
	logger.DBAttachments().Debug("Listing all attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "attachments",
			"failed to query attachments", err)
	}
	defer rows.Close()

	var attachments []*attachment.Attachment

	for rows.Next() {
		attachment, err := a.scan(rows)
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
func (a *Service) Get(id int64) (*attachment.Attachment, error) {
	logger.DBAttachments().Debug("Getting attachment, id: %d", id)

	query := `SELECT id, mime_type, data FROM attachments WHERE id = ?`
	row := a.db.QueryRow(query, id)
	attachment, err := a.scan(row)
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
func (s *Service) GetByIDs(ids []int64) ([]*attachment.Attachment, error) {
	if len(ids) == 0 {
		return []*attachment.Attachment{}, nil
	}

	logger.DBAttachments().Debug("Getting attachments by IDs: %v", ids)

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
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
		strings.Join(placeholders, ","),
	)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, dberror.NewDatabaseError("select", "attachments",
			"failed to query attachments by IDs", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap := make(map[int64]*attachment.Attachment)

	for rows.Next() {
		attachment, err := s.scan(rows)
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
	var attachments []*attachment.Attachment
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

// Add creates a new attachment and returns its generated ID.
func (a *Service) Add(attachment *attachment.Attachment, _ *user.User) (int64, error) {
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
func (a *Service) Update(attachment *attachment.Attachment, _ *user.User) error {
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
func (a *Service) Delete(id int64, _ *user.User) error {
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

func (s *Service) scan(scanner interfaces.Scannable) (*attachment.Attachment, error) {
	attachment := &attachment.Attachment{}
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
