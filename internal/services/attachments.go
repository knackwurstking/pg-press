package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/pkg/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// Attachments provides database operations for managing attachments with lazy loading.
type Attachments struct {
	db  *sql.DB
	log *logger.Logger
}

// NewAttachment creates a new Service instance and initializes the database table.
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
		panic(fmt.Errorf("failed to create attachments table: %v", err))
	}

	return &Attachments{
		db:  db,
		log: logger.GetComponentLogger("Service: Attachments"),
	}
}

// List retrieves all attachments ordered by ID ascending.
func (a *Attachments) List() ([]*models.Attachment, error) {
	a.log.Debug("Listing all attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select attachments: %v", err)
	}
	defer rows.Close()

	var attachments []*models.Attachment

	for rows.Next() {
		attachment, err := a.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan attachments: %v", err)
		}
		attachments = append(attachments, attachment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select attachments: %v", err)
	}

	return attachments, nil
}

// Get retrieves a specific attachment by ID.
func (a *Attachments) Get(id int64) (*models.Attachment, error) {
	a.log.Debug("Getting attachment, id: %d", id)

	query := `SELECT id, mime_type, data FROM attachments WHERE id = ?`
	row := a.db.QueryRow(query, id)
	attachment, err := a.scan(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("attachment with ID %d not found", id))
		}

		return nil, fmt.Errorf("select attachments: %v", err)
	}

	return attachment, nil
}

// GetByIDs retrieves multiple attachments by their IDs in the order specified.
func (s *Attachments) GetByIDs(ids []int64) ([]*models.Attachment, error) {
	if len(ids) == 0 {
		return []*models.Attachment{}, nil
	}

	s.log.Debug("Getting attachments by IDs: %v", ids)

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
		return nil, fmt.Errorf("select attachments: %v", err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap := make(map[int64]*models.Attachment)

	for rows.Next() {
		attachment, err := s.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scan attachments: %v", err)
		}
		attachmentMap[attachment.GetID()] = attachment
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select attachments: %v", err)
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
func (a *Attachments) Add(attachment *models.Attachment) (int64, error) {
	a.log.Debug("Adding attachment: %s", attachment.String())

	if attachment == nil {
		return 0, utils.NewValidationError("attachment cannot be nil")
	}

	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	query := `INSERT INTO attachments (mime_type, data) VALUES (?, ?)`
	result, err := a.db.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, fmt.Errorf("insert attachments: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert attachments: %v", err)
	}

	return id, nil
}

// Update modifies an existing attachment.
func (a *Attachments) Update(attachment *models.Attachment) error {
	id := attachment.GetID()
	a.log.Debug("Updating attachment, id: %d", id)

	if attachment == nil {
		return utils.NewValidationError("attachment: attachment cannot be nil")
	}

	if err := attachment.Validate(); err != nil {
		return err
	}

	query := `UPDATE attachments SET mime_type = ?, data = ? WHERE id = ?`
	result, err := a.db.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return fmt.Errorf("update attachments: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update attachments: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError(fmt.Sprintf("id: %d", id))
	}

	return nil
}

// Delete deletes an attachment by ID.
func (a *Attachments) Delete(id int64) error {
	a.log.Debug("Removing attachment, id: %d", id)

	query := `DELETE FROM attachments WHERE id = ?`
	result, err := a.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete attachments: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete attachments: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError(fmt.Sprintf("id: %d", id))
	}

	return nil
}

func (s *Attachments) scan(scanner interfaces.Scannable) (*models.Attachment, error) {
	attachment := &models.Attachment{}
	var id int64

	if err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data); err != nil {
		// The `Get` method needs the original sql.ErrNoRows error
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan row: %v", err)
	}

	// Set the ID using string conversion to maintain compatibility
	attachment.ID = fmt.Sprintf("%d", id)

	return attachment, nil
}
