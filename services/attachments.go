package services

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
)

const TableNameAttachments = "attachments"

type Attachments struct {
	*Base
}

func NewAttachments(r *Registry) *Attachments {
	base := NewBase(r, logger.NewComponentLogger("Service: Attachments"))

	query := `
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			mime_type TEXT NOT NULL,
			data BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if err := base.CreateTable(query, TableNameAttachments); err != nil {
		panic(err)
	}

	return &Attachments{
		Base: base,
	}
}

func (a *Attachments) List() ([]*models.Attachment, error) {
	a.Log.Debug("Listing attachments")

	rows, err := a.DB.Query(fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s ORDER BY id ASC`, TableNameAttachments,
	))
	if err != nil {
		return nil, a.GetSelectError(err)
	}
	defer rows.Close()

	attachments, err := ScanRows(rows, scanAttachment)
	if err != nil {
		return nil, err
	}

	return attachments, nil
}

func (a *Attachments) Get(id int64) (*models.Attachment, error) {
	a.Log.Debug("Getting attachment for ID: %d", id)

	row := a.DB.QueryRow(fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s WHERE id = ?`, TableNameAttachments),
		id,
	)
	attachment, err := ScanSingleRow(row, scanAttachment)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf(
				"attachment with ID %d not found", id))
		}
		return nil, a.GetSelectError(err)
	}

	return attachment, nil
}

func (a *Attachments) GetByIDs(ids []int64) ([]*models.Attachment, error) {
	a.Log.Debug("Getting attachments by IDs: count: %d", len(ids))

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
		`
			SELECT
				*
			FROM
				%s
			WHERE
				id
			IN
				(%s)
			ORDER BY
				id ASC
		`,
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
	var attachments []*models.Attachment
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachments = append(attachments, attachment)
		}
	}

	return attachments, nil
}

func (a *Attachments) Add(attachment *models.Attachment) (int64, error) {
	a.Log.Debug("Adding attachment: mime_type: %s, size: %d bytes",
		attachment.MimeType, len(attachment.Data))

	if err := attachment.Validate(); err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`INSERT INTO %s (mime_type, data) VALUES (?, ?)`,
		TableNameAttachments)
	result, err := a.DB.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, a.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, a.GetInsertError(err)
	}

	return id, nil
}

func (a *Attachments) Update(attachment *models.Attachment) error {
	a.Log.Debug("Updating attachment: id: %s", attachment.ID)

	if err := attachment.Validate(); err != nil {
		return err
	}

	// Convert string ID to int64
	id := attachment.GetID()
	query := fmt.Sprintf(`UPDATE %s SET mime_type = ?, data = ? WHERE id = ?`,
		TableNameAttachments)
	_, err := a.DB.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return a.GetUpdateError(err)
	}

	return nil
}

func (a *Attachments) Delete(id int64) error {
	a.Log.Debug("Deleting attachment: %d", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameAttachments)
	_, err := a.DB.Exec(query, id)
	if err != nil {
		return a.GetDeleteError(err)
	}

	return nil
}

func scanAttachment(scanner Scannable) (*models.Attachment, error) {
	attachment := &models.Attachment{}
	var id int64

	err := scanner.Scan(&id, &attachment.MimeType, &attachment.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan attachment: %v", err)
	}

	// Set the ID using string conversion to maintain compatibility
	attachment.ID = fmt.Sprintf("%d", id)
	return attachment, nil
}

func scanAttachmentsIntoMap(rows *sql.Rows) (map[int64]*models.Attachment, error) {
	resultMap := make(map[int64]*models.Attachment)

	for rows.Next() {
		attachment, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}

		// Convert string ID back to int64 for map key
		id := attachment.GetID()
		resultMap[id] = attachment
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return resultMap, nil
}
