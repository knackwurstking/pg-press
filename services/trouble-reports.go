package services

import (
	"database/sql"
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

func (s *TroubleReports) List() ([]*models.TroubleReport, error) {
	slog.Debug("Listing trouble reports")

	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY id DESC`, TableNameTroubleReports)
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTroubleReport)
}

func (s *TroubleReports) Get(id models.TroubleReportID) (*models.TroubleReport, error) {
	slog.Debug("Getting trouble report", "trouble_report_id", id)

	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, TableNameTroubleReports)
	row := s.DB.QueryRow(query, id)

	report, err := ScanSingleRow(row, scanTroubleReport)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("trouble report with ID %d", id))
		}
		return nil, err
	}

	return report, nil
}

func (s *TroubleReports) Add(tr *models.TroubleReport, u *models.User) (int64, error) {
	slog.Debug("Adding trouble report", "user_name", u.Name, "title", tr.Title, "attachments", len(tr.LinkedAttachments))

	if err := tr.Validate(); err != nil {
		return 0, err
	}

	if err := u.Validate(); err != nil {
		return 0, err
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, fmt.Errorf("marshal linked attachments: %v", err)
	}

	query := fmt.Sprintf(`INSERT INTO %s (title, content, linked_attachments, use_markdown) VALUES (?, ?, ?, ?)`,
		TableNameTroubleReports)

	result, err := s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown)
	if err != nil {
		return 0, s.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.GetInsertError(err)
	}
	tr.ID = models.TroubleReportID(id)

	if _, err := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		id,
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	); err != nil {
		slog.Error("Failed to save initial modification for trouble report",
			"trouble_report_id", id, "error", err)
	}

	return id, nil
}

func (s *TroubleReports) Update(tr *models.TroubleReport, u *models.User) error {
	slog.Debug(
		"Updating trouble report",
		"user", u,
		"trouble_report_id", tr.ID,
		"title", tr.Title,
		"linked_attachments", len(tr.LinkedAttachments),
	)

	if err := tr.Validate(); err != nil {
		return err
	}

	if err := u.Validate(); err != nil {
		return err
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("marshal linked attachments: %v", err)
	}

	query := fmt.Sprintf(`UPDATE %s SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`,
		TableNameTroubleReports)

	_, err = s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown, tr.ID)
	if err != nil {
		return s.GetInsertError(err)
	}

	if _, err := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		int64(tr.ID),
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	); err != nil {
		slog.Error("Failed to save modification for trouble report",
			"trouble_report_id", tr.ID, "error", err)
	}

	return nil
}

func (s *TroubleReports) Delete(id models.TroubleReportID, user *models.User) error {
	slog.Debug("Deleting trouble report", "user_name", user.Name, "trouble_report_id", id)

	if err := user.Validate(); err != nil {
		return err
	}

	if _, err := s.Get(id); err != nil {
		return err
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTroubleReports)
	_, err := s.DB.Exec(query, id)
	if err != nil {
		return s.GetDeleteError(err)
	}

	return nil
}

func (s *TroubleReports) GetWithAttachments(id models.TroubleReportID) (*models.TroubleReportWithAttachments, error) {
	slog.Debug("Getting trouble report with attachments", "trouble_report_id", id)

	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	attachments, err := s.LoadAttachments(tr)
	if err != nil {
		return nil, fmt.Errorf("load attachments: %v", err)
	}

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

func (s *TroubleReports) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	slog.Debug("Getting all trouble reports with attachments")

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

func (s *TroubleReports) AddWithAttachments(troubleReport *models.TroubleReport, user *models.User, attachments ...*models.Attachment) error {
	slog.Debug(
		"Adding trouble report with attachments",
		"user", user,
		"title", troubleReport.Title,
		"attachments", len(attachments),
	)

	if err := user.Validate(); err != nil {
		return err
	}

	var attachmentIDs []models.AttachmentID
	for i, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, err := s.Registry.Attachments.Add(attachment)
		if err != nil {
			slog.Error("Failed to add attachment, cleaning up...",
				"index", i, "attachments", len(attachmentIDs), "error", err)
			s.cleanupAttachments(attachmentIDs)
			return err
		}

		attachmentIDs = append(attachmentIDs, id)
	}

	troubleReport.LinkedAttachments = attachmentIDs

	id, err := s.Add(troubleReport, user)
	if err != nil {
		slog.Error("Failed to add trouble report, cleaning up attachments...",
			"attachments", len(attachmentIDs), "error", err)
		s.cleanupAttachments(attachmentIDs)
		return err
	}

	troubleReport.ID = models.TroubleReportID(id)
	return nil
}

func (s *TroubleReports) UpdateWithAttachments(id models.TroubleReportID, tr *models.TroubleReport, user *models.User, newAttachments ...*models.Attachment) error {
	slog.Debug(
		"Updating trouble report with attachments",
		"user_name", user.Name,
		"trouble_report_id", id,
		"title", tr.Title,
		"new_attachments", len(newAttachments),
	)

	if err := tr.Validate(); err != nil {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
	}

	var newAttachmentIDs []models.AttachmentID
	for i, attachment := range newAttachments {
		if attachment == nil {
			slog.Debug("Skipping nil attachment", "index", i)
			continue
		}

		slog.Debug("Adding new attachment", "index", i, "attachment_data_size", len(attachment.Data))

		attachmentID, err := s.Registry.Attachments.Add(attachment)
		if err != nil {
			slog.Error("Failed to add new attachment, cleaning up...",
				"index", i, "new_attachments", len(newAttachmentIDs), "error", err)
			s.cleanupAttachments(newAttachmentIDs)
			return err
		}

		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	tr.LinkedAttachments = append(tr.LinkedAttachments, newAttachmentIDs...)
	tr.ID = id

	if err := s.Update(tr, user); err != nil {
		slog.Error("Failed to update trouble report, cleaning up...",
			"trouble_report_id", id, "new_attachments", len(newAttachmentIDs), "error", err)
		s.cleanupAttachments(newAttachmentIDs)
		return err
	}

	return nil
}

func (s *TroubleReports) RemoveWithAttachments(id models.TroubleReportID, user *models.User) (*models.TroubleReport, error) {
	slog.Debug("Removing trouble report with attachments", "user_name", user.Name, "trouble_report_id", id)

	if err := user.Validate(); err != nil {
		return nil, err
	}

	tr, err := s.Get(id)
	if err != nil {
		return tr, err
	}

	if err := s.Delete(id, user); err != nil {
		return tr, err
	}

	failedDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.Registry.Attachments.Delete(attachmentID); err != nil {
			slog.Error("Failed to delete attachment for trouble report",
				"attachment_id", attachmentID, "trouble_report_id", tr.ID, "error", err)
			failedDeletes++
		}
	}

	if failedDeletes > 0 {
		slog.Error(
			"Failed to remove attachments for trouble report",
			"trouble_report_id", tr.ID,
			"linked_attachments", len(tr.LinkedAttachments),
			"failed_deletes", failedDeletes,
		)
	}

	return tr, nil
}

func (s *TroubleReports) LoadAttachments(report *models.TroubleReport) ([]*models.Attachment, error) {
	slog.Debug("Loading attachments for trouble report",
		"trouble_report_id", report.ID, "attachments", len(report.LinkedAttachments))

	if err := report.Validate(); err != nil {
		return nil, err
	}

	var attachments []*models.Attachment
	for _, attachmentID := range report.LinkedAttachments {
		attachment, err := s.Registry.Attachments.Get(attachmentID)
		if err != nil {
			slog.Error("Failed to load attachment", "attachment_id", attachmentID, "error", err)
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

func scanTroubleReport(scannable Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string

	err := scannable.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments, &report.UseMarkdown)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan trouble report: %v", err)
	}

	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		return nil, fmt.Errorf("unmarshal linked attachments: %v", err)
	}

	return report, nil
}
