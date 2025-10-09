package attachments

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
}

func NewService(db *sql.DB) *Service {
	base := base.NewBaseService(db, "Attachments")

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

	return &Service{
		BaseService: base,
	}
}

func (a *Service) List() ([]*models.Attachment, error) {
	a.Log.Debug("Listing attachments")

	query := `SELECT id, mime_type, data FROM attachments ORDER BY id ASC`
	rows, err := a.DB.Query(query)
	if err != nil {
		return nil, a.HandleSelectError(err, "attachments")
	}
	defer rows.Close()

	attachments, err := scanAttachmentsFromRows(rows)
	if err != nil {
		return nil, err
	}

	a.Log.Debug("Listed attachments: count: %d", len(attachments))
	return attachments, nil
}

func (a *Service) Get(id int64) (*models.Attachment, error) {
	if err := validation.ValidateID(id, "attachment"); err != nil {
		return nil, err
	}

	a.Log.Debug("Getting attachment: %d", id)

	query := `SELECT id, mime_type, data FROM attachments WHERE id = ?`
	row := a.DB.QueryRow(query, id)

	attachment, err := scanner.ScanSingleRow(row, scanAttachment, "attachments")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("attachment with ID %d not found", id))
		}
		return nil, err
	}

	return attachment, nil
}

func (a *Service) GetByIDs(ids []int64) ([]*models.Attachment, error) {
	if len(ids) == 0 {
		return []*models.Attachment{}, nil
	}

	a.Log.Debug("Getting attachments by IDs: count: %d", len(ids))

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

	rows, err := a.DB.Query(query, args...)
	if err != nil {
		return nil, a.HandleSelectError(err, "attachments")
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap, err := scanAttachmentsIntoMap(rows)
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

	a.Log.Debug("Found attachments by IDs: found: %d", len(attachments))
	return attachments, nil
}

func (a *Service) Add(attachment *models.Attachment) (int64, error) {
	if err := validateAttachment(attachment); err != nil {
		return 0, err
	}

	// Call the model's validate method for additional checks
	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	a.Log.Debug("Adding attachment: mime_type: %s, size: %d bytes",
		attachment.MimeType, len(attachment.Data))

	query := `INSERT INTO attachments (mime_type, data) VALUES (?, ?)`
	result, err := a.DB.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, a.HandleInsertError(err, "attachments")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, a.HandleInsertError(err, "attachments")
	}

	a.Log.Debug("Added attachment: id: %d", id)
	return id, nil
}

func (a *Service) Update(attachment *models.Attachment) error {
	if err := validateAttachment(attachment); err != nil {
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

	if err := validation.ValidateID(id, "attachment"); err != nil {
		return err
	}

	a.Log.Debug("Updating attachment: id: %d", id)

	query := `UPDATE attachments SET mime_type = ?, data = ? WHERE id = ?`
	result, err := a.DB.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return a.HandleUpdateError(err, "attachments")
	}

	if err := a.CheckRowsAffected(result, "attachment", id); err != nil {
		return err
	}

	a.Log.Debug("Updated attachment: id: %d", id)
	return nil
}

func (a *Service) Delete(id int64) error {
	if err := validation.ValidateID(id, "attachment"); err != nil {
		return err
	}

	a.Log.Debug("Deleting attachment: %d", id)

	query := `DELETE FROM attachments WHERE id = ?`
	result, err := a.DB.Exec(query, id)
	if err != nil {
		return a.HandleDeleteError(err, "attachments")
	}

	if err := a.CheckRowsAffected(result, "attachment", id); err != nil {
		return err
	}

	a.Log.Debug("Deleted attachment: %d", id)
	return nil
}
