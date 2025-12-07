package services

import (
	"fmt"
	"strings"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

type Attachments struct {
	*Base
}

func NewAttachments(r *Registry) *Attachments {
	return &Attachments{
		Base: NewBase(r),
	}
}

func (a *Attachments) Get(id models.AttachmentID) (*models.Attachment, *errors.MasterError) {
	attachment, err := ScanAttachment(a.DB.QueryRow(SQLGetAttachmentByID, id))
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}

	return attachment, nil
}

func (a *Attachments) List() ([]*models.Attachment, *errors.MasterError) {
	rows, err := a.DB.Query(SQLListAttachments)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanAttachment)
}

func (a *Attachments) ListByIDs(ids []models.AttachmentID) ([]*models.Attachment, *errors.MasterError) {
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

	rows, err := a.DB.Query(
		fmt.Sprintf(SQLListAttachmentsByIDs, strings.Join(placeholders, ", ")),
		args...,
	)
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

func (a *Attachments) Add(mimeType string, data []byte) (models.AttachmentID, *errors.MasterError) {
	result, err := a.DB.Exec(SQLAddAttachment, mimeType, data)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.AttachmentID(id), nil
}

func (a *Attachments) Update(id int64, mimeType string, data []byte) *errors.MasterError {
	_, err := a.DB.Exec(SQLUpdateAttachment, mimeType, data, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (a *Attachments) Delete(id models.AttachmentID) *errors.MasterError {
	_, err := a.DB.Exec(SQLDeleteAttachment, id)
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
