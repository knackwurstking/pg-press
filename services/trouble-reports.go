package services

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

type TroubleReports struct {
	*Base
}

func NewTroubleReports(r *Registry) *TroubleReports {
	return &TroubleReports{
		Base: NewBase(r),
	}
}

func (s *TroubleReports) List() ([]*models.TroubleReport, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY id DESC`, TableNameTroubleReports)
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, errors.NewMasterError(err, 0)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTroubleReport)
}

func (s *TroubleReports) Get(id models.TroubleReportID) (*models.TroubleReport, *errors.MasterError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameTroubleReports)
	tr, err := ScanTroubleReport(s.DB.QueryRow(query, id))
	if err != nil {
		return tr, errors.NewMasterError(err, 0)
	}
	return tr, nil
}

func (s *TroubleReports) Add(tr *models.TroubleReport, u *models.User) (int64, *errors.MasterError) {
	verr := tr.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	verr = u.Validate()
	if verr != nil {
		return 0, verr.MasterError()
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	query := fmt.Sprintf(`INSERT INTO %s (title, content, linked_attachments, use_markdown) VALUES (?, ?, ?, ?)`,
		TableNameTroubleReports)

	result, err := s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}
	tr.ID = models.TroubleReportID(id)

	_, merr := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		id,
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	)
	return id, merr
}

func (s *TroubleReports) Update(tr *models.TroubleReport, u *models.User) *errors.MasterError {
	verr := tr.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	verr = u.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	query := fmt.Sprintf(`UPDATE %s SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`,
		TableNameTroubleReports)

	_, err = s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown, tr.ID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	_, merr := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		int64(tr.ID),
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	)
	if merr != nil {
		return merr
	}

	return nil
}

func (s *TroubleReports) Delete(id models.TroubleReportID, user *models.User) *errors.MasterError {
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	_, merr := s.Get(id)
	if merr != nil {
		return merr
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTroubleReports)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (s *TroubleReports) GetWithAttachments(id models.TroubleReportID) (*models.TroubleReportWithAttachments, *errors.MasterError) {
	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	attachments, merr := s.LoadAttachments(tr)
	if merr != nil {
		return nil, merr
	}

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

func (s *TroubleReports) ListWithAttachments() ([]*models.TroubleReportWithAttachments, *errors.MasterError) {
	reports, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []*models.TroubleReportWithAttachments
	for _, tr := range reports {
		attachments, err := s.LoadAttachments(tr)
		if err != nil {
			return nil, err
		}

		result = append(result, &models.TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	return result, nil
}

func (s *TroubleReports) AddWithAttachments(
	troubleReport *models.TroubleReport,
	user *models.User,
	attachments ...*models.Attachment,
) *errors.MasterError {

	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	var attachmentIDs []models.AttachmentID
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, merr := s.Registry.Attachments.Add(attachment.MimeType, attachment.Data)
		if merr != nil {
			s.cleanupAttachments(attachmentIDs)
			return merr
		}

		attachmentIDs = append(attachmentIDs, id)
	}

	troubleReport.LinkedAttachments = attachmentIDs

	id, merr := s.Add(troubleReport, user)
	if merr != nil {
		s.cleanupAttachments(attachmentIDs)
		return merr
	}

	troubleReport.ID = models.TroubleReportID(id)
	return nil
}

func (s *TroubleReports) UpdateWithAttachments(
	id models.TroubleReportID,
	tr *models.TroubleReport,
	user *models.User,
	newAttachments ...*models.Attachment,
) *errors.MasterError {

	verr := tr.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	verr = user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	var newAttachmentIDs []models.AttachmentID
	for _, attachment := range newAttachments {
		if attachment == nil {
			continue
		}

		attachmentID, merr := s.Registry.Attachments.Add(attachment.MimeType, attachment.Data)
		if merr != nil {
			s.cleanupAttachments(newAttachmentIDs)
			return merr
		}

		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	tr.LinkedAttachments = append(tr.LinkedAttachments, newAttachmentIDs...)
	tr.ID = id

	merr := s.Update(tr, user)
	if merr != nil {
		s.cleanupAttachments(newAttachmentIDs)
		return merr
	}

	return nil
}

func (s *TroubleReports) RemoveWithAttachments(
	id models.TroubleReportID,
	user *models.User,
) (*models.TroubleReport, *errors.MasterError) {

	verr := user.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	tr, merr := s.Get(id)
	if merr != nil {
		return tr, merr
	}

	merr = s.Delete(id, user)
	if merr != nil {
		return tr, merr
	}

	failedMessages := []string{}
	failedDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		merr := s.Registry.Attachments.Delete(attachmentID)
		if merr != nil {
			failedDeletes++
			failedMessages = append(
				failedMessages,
				fmt.Sprintf(
					"Failed to delete attachment for trouble report %d: (attachment id: %d) %#v",
					tr.ID, attachmentID, merr,
				),
			)
		}
	}

	if failedDeletes > 0 {
		msg := ""
		for i, e := range failedMessages {
			msg += e
			if len(failedMessages) < i+1 {
				msg += "\n\t"
			}
		}
		return tr, errors.NewMasterError(fmt.Errorf("Some attachments could not be removed:\n\t%s", msg), 0)
	}

	return tr, nil
}

func (s *TroubleReports) LoadAttachments(tr *models.TroubleReport) (
	[]*models.Attachment, *errors.MasterError,
) {

	verr := tr.Validate()
	if verr != nil {
		return nil, verr.MasterError()
	}

	var attachments []*models.Attachment
	for _, attachmentID := range tr.LinkedAttachments {
		attachment, err := s.Registry.Attachments.Get(attachmentID)
		if err != nil {
			continue
		}
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (s *TroubleReports) cleanupAttachments(attachmentIDs []models.AttachmentID) {
	for _, id := range attachmentIDs {
		if err := s.Registry.Attachments.Delete(id); err != nil {
			slog.Error("Failed to cleanup attachment", "attachment_id", id, "error", err)
		}
	}
}
