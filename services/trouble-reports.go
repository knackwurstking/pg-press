package services

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameTroubleReports = "trouble_reports"

type TroubleReports struct {
	*Base
}

func NewTroubleReports(r *Registry) *TroubleReports {
	base := NewBase(r)

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			use_markdown BOOLEAN DEFAULT 0,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`, TableNameTroubleReports)

	if _, err := base.DB.Exec(query); err != nil {
		panic(errors.Wrap(err, "create %s table", TableNameTroubleReports))
	}

	return &TroubleReports{Base: base}
}

func (s *TroubleReports) List() ([]*models.TroubleReport, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY id DESC`, TableNameTroubleReports)
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeSelect)
	}
	defer rows.Close()

	return ScanRows(rows, ScanTroubleReport)
}

func (s *TroubleReports) Get(id models.TroubleReportID) (*models.TroubleReport, *errors.DBError) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameTroubleReports)
	return ScanRow(s.DB.QueryRow(query, id), ScanTroubleReport)
}

func (s *TroubleReports) Add(tr *models.TroubleReport, u *models.User) (int64, *errors.DBError) {
	err := tr.Validate()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	err = u.Validate()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`INSERT INTO %s (title, content, linked_attachments, use_markdown) VALUES (?, ?, ?, ?)`,
		TableNameTroubleReports)

	result, err := s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown)
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}
	tr.ID = models.TroubleReportID(id)

	_, dberr := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		id,
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	)
	return id, dberr
}

func (s *TroubleReports) Update(tr *models.TroubleReport, u *models.User) *errors.DBError {
	err := tr.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	err = u.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`UPDATE %s SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`,
		TableNameTroubleReports)

	_, err = s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown, tr.ID)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeInsert)
	}

	_, dberr := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		int64(tr.ID),
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	)
	if dberr != nil {
		return dberr
	}

	return nil
}

func (s *TroubleReports) Delete(id models.TroubleReportID, user *models.User) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	_, dberr := s.Get(id)
	if dberr != nil {
		return dberr
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTroubleReports)
	_, err = s.DB.Exec(query, id)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}

func (s *TroubleReports) GetWithAttachments(id models.TroubleReportID) (*models.TroubleReportWithAttachments, *errors.DBError) {
	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	attachments, dberr := s.LoadAttachments(tr)
	if dberr != nil {
		return nil, dberr
	}

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

func (s *TroubleReports) ListWithAttachments() ([]*models.TroubleReportWithAttachments, *errors.DBError) {
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
) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	var attachmentIDs []models.AttachmentID
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, dberr := s.Registry.Attachments.Add(attachment)
		if dberr != nil {
			s.cleanupAttachments(attachmentIDs)
			return dberr
		}

		attachmentIDs = append(attachmentIDs, id)
	}

	troubleReport.LinkedAttachments = attachmentIDs

	id, dberr := s.Add(troubleReport, user)
	if err != nil {
		s.cleanupAttachments(attachmentIDs)
		return dberr
	}

	troubleReport.ID = models.TroubleReportID(id)
	return nil
}

func (s *TroubleReports) UpdateWithAttachments(
	id models.TroubleReportID,
	tr *models.TroubleReport,
	user *models.User,
	newAttachments ...*models.Attachment,
) *errors.DBError {
	err := tr.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	err = user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	var newAttachmentIDs []models.AttachmentID
	for _, attachment := range newAttachments {
		if attachment == nil {
			continue
		}

		attachmentID, dberr := s.Registry.Attachments.Add(attachment)
		if dberr != nil {
			s.cleanupAttachments(newAttachmentIDs)
			return dberr
		}

		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	tr.LinkedAttachments = append(tr.LinkedAttachments, newAttachmentIDs...)
	tr.ID = id

	dberr := s.Update(tr, user)
	if dberr != nil {
		s.cleanupAttachments(newAttachmentIDs)
		return dberr
	}

	return nil
}

func (s *TroubleReports) RemoveWithAttachments(
	id models.TroubleReportID,
	user *models.User,
) (*models.TroubleReport, *errors.DBError) {
	err := user.Validate()
	if err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeValidation)
	}

	tr, dberr := s.Get(id)
	if dberr != nil {
		return tr, dberr
	}

	dberr = s.Delete(id, user)
	if dberr != nil {
		return tr, dberr
	}

	failedMessages := []string{}
	failedDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		dberr := s.Registry.Attachments.Delete(attachmentID)
		if dberr != nil {
			failedDeletes++
			failedMessages = append(
				failedMessages,
				fmt.Sprintf(
					"Failed to delete attachment for trouble report %d: (attachment id: %d) %#v",
					tr.ID, attachmentID, dberr,
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
		return tr, errors.NewDBError(
			fmt.Errorf("Some attachments could not be removed:\n\t%s", msg),
			errors.DBTypeDelete,
		)
	}

	return tr, nil
}

func (s *TroubleReports) LoadAttachments(report *models.TroubleReport) (
	[]*models.Attachment, *errors.DBError,
) {
	if err := report.Validate(); err != nil {
		return nil, errors.NewDBError(err, errors.DBTypeValidation)
	}

	var attachments []*models.Attachment
	for _, attachmentID := range report.LinkedAttachments {
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
