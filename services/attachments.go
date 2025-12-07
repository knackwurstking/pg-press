package services

import (
	"fmt"
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

func (a *Attachments) List() ([]*models.Attachment, *errors.MasterError) {
	query := fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s ORDER BY id ASC`,
		TableNameAttachments,
	)

	rows, err := a.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanAttachment)
}

func (a *Attachments) Get(id models.AttachmentID) (*models.Attachment, *errors.MasterError) {
	query := fmt.Sprintf(
		`SELECT id, mime_type, data FROM %s WHERE id = ?`,
		TableNameAttachments,
	)

	attachment, err := ScanAttachment(a.DB.QueryRow(query, id))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return attachment, nil
}

func (a *Attachments) GetByIDs(ids []models.AttachmentID) ([]*models.Attachment, *errors.MasterError) {
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
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	// Store attachments in a map for efficient lookup
	attachments, merr := ScanRows(rows, ScanAttachment)
	if merr != nil {
		return nil, merr
	}

	// Create a map for O(1) lookup
	attachmentMap := MapAttachments(attachments)

	// Return attachments in the order of the requested IDs
	attachmentsInOrder := make([]*models.Attachment, 0, len(ids))
	for _, id := range ids {
		if attachment, exists := attachmentMap[id]; exists {
			attachmentsInOrder = append(attachmentsInOrder, attachment)
		}
	}

	return attachmentsInOrder, nil
}

func (a *Attachments) Add(attachment *models.Attachment) (models.AttachmentID, *errors.MasterError) {
	verr := attachment.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (mime_type, data) VALUES (?, ?)`,
		TableNameAttachments,
	)

	result, err := a.DB.Exec(query, attachment.MimeType, attachment.Data)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.AttachmentID(id), nil
}

func (a *Attachments) Update(attachment *models.Attachment) *errors.MasterError {
	verr := attachment.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	id := attachment.GetID()
	query := fmt.Sprintf(
		`UPDATE %s SET mime_type = ?, data = ? WHERE id = ?`,
		TableNameAttachments,
	)

	_, err := a.DB.Exec(query, attachment.MimeType, attachment.Data, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (a *Attachments) Delete(id models.AttachmentID) *errors.MasterError {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameAttachments)
	_, err := a.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func MapAttachments(attachments []*models.Attachment) map[models.AttachmentID]*models.Attachment {
	attachmentMap := map[models.AttachmentID]*models.Attachment{}

	for _, a := range attachments {
		attachmentMap[a.GetID()] = a
	}

	return attachmentMap
}
