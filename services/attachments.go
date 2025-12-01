package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameAttachments = "attachments"

type Attachments struct {
	*Base
}

func NewAttachments(r *Registry) *Attachments {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			mime_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`, TableNameAttachments)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameAttachments))
	}

	return &Attachments{
		Base: base,
	}
}

func (a *Attachments) List() ([]*models.Attachment, error) {
	slog.Info("Listing attachments")

	query := fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s ORDER BY id ASC`,
		TableNameAttachments,
	)

	rows, err := a.DB.Query(query)
	if err != nil {
		return nil, a.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanAttachment)
}

func (a *Attachments) Get(id models.AttachmentID) (*models.Attachment, error) {
	slog.Info("Getting attachment for ID", "id", id)

	query := fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s WHERE id = ?`,
		TableNameAttachments,
	)

	row := a.DB.QueryRow(query, id)
	attachment, err := ScanSingleRow(row, scanAttachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(
				fmt.Sprintf("attachment with ID %d not found", id),
			)
		}
		return nil, a.GetSelectError(err)
	}

	return attachment, nil
}

func (a *Attachments) GetByIDs(ids []models.AttachmentID) ([]*models.Attachment, error) {
	slog.Info("Getting attachments by IDs", "ids", len(ids))

	if len(ids) == 0 {
		return []*models.Attachment{}, nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s WHERE id IN (%s) ORDER BY id ASC`,
		TableNameAttachments,
		strings.Join(placeholders, ","),
	)

	rows, err := a.DB.Query(query, args...)
	if err != nil {
		return nil, a.GetSelectError(err)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachmentMap, err := scanAttachmentsIntoMap(rows)
	if err != nil {
		return nil, err
	}

	// Return attachments in the order of the requested IDs
	attachments := make([]*models.Attachment, 0, len(ids))
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

func (a *Attachments) Add(attachment *models.Attachment) (models.AttachmentID, error) {
	slog.Debug("Adding attachment", "mime_type", attachment.MimeType, "size", len(attachment.Data))

	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (mime_type, data) VALUES (?, ?)`,
		TableNameAttachments,
	)

	result, err := a.DB.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, a.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, a.GetInsertError(err)
	}

	return models.AttachmentID(id), nil
}

func (a *Attachments) Update(attachment *models.Attachment) error {
	slog.Debug("Updating attachment", "attachment", attachment.ID)

	if err := attachment.Validate(); err != nil {
		return err
	}

	id := attachment.GetID()
	query := fmt.Sprintf(
		`UPDATE %s SET mime_type = ?, data = ? WHERE id = ?`,
		TableNameAttachments,
	)

	_, err := a.DB.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return a.GetUpdateError(err)
	}

	return nil
}

func (a *Attachments) Delete(id models.AttachmentID) error {
	slog.Debug("Deleting attachment", "id", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameAttachments)
	_, err := a.DB.Exec(query, id)
	if err != nil {
		return a.GetDeleteError(err)
	}

	return nil
}

func scanAttachment(scanner Scannable) (*models.Attachment, error) {
	var id int64
	attachment := &models.Attachment{}

	err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan attachment: %v", err)
	}

	attachment.ID = fmt.Sprintf("%d", id)
	return attachment, nil
}

func scanAttachmentsIntoMap(rows *sql.Rows) (map[models.AttachmentID]*models.Attachment, error) {
	resultMap := make(map[models.AttachmentID]*models.Attachment)

	for rows.Next() {
		attachment, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}

		id := attachment.GetID()
		resultMap[id] = attachment
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return resultMap, nil
}
