package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
)

const TableNameTroubleReports = "trouble_reports"

type TroubleReports struct {
	*Base
}

func NewTroubleReports(r *Registry) *TroubleReports {
	base := NewBase(r, logger.NewComponentLogger("Service: TroubleReports"))

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

	if err := base.CreateTable(query, TableNameTroubleReports); err != nil {
		panic(err)
	}

	return &TroubleReports{Base: base}
}

func (s *TroubleReports) List() ([]*models.TroubleReport, error) {
	s.Log.Debug("Listing trouble reports")

	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY id DESC`, TableNameTroubleReports)
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, s.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTroubleReport)
}

func (s *TroubleReports) Get(id int64) (*models.TroubleReport, error) {
	s.Log.Debug("Getting trouble report: %d", id)

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
	s.Log.Debug("Adding trouble report by %s: title: %s, attachments: %d",
		u.String(), tr.Title, len(tr.LinkedAttachments))

	if err := tr.Validate(); err != nil {
		return 0, err
	}

	if err := u.Validate(); err != nil {
		return 0, err
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal linked attachments: %v", err)
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
	tr.ID = id

	if _, err := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		id,
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	); err != nil {
		s.Log.Error("Failed to save initial modification for trouble report %d: %v", id, err)
	}

	return id, nil
}

func (s *TroubleReports) Update(tr *models.TroubleReport, u *models.User) error {
	s.Log.Debug("Updating trouble report by %s: id: %d, title: %s, attachments: %d",
		u.String(), tr.ID, tr.Title, len(tr.LinkedAttachments))

	if err := tr.Validate(); err != nil {
		return err
	}

	if err := u.Validate(); err != nil {
		return err
	}

	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	query := fmt.Sprintf(`UPDATE %s SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`,
		TableNameTroubleReports)

	_, err = s.DB.Exec(query, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown, tr.ID)
	if err != nil {
		return s.GetInsertError(err)
	}

	if _, err := s.Registry.Modifications.Add(
		models.ModificationTypeTroubleReport,
		tr.ID,
		models.NewTroubleReportModData(tr),
		u.TelegramID,
	); err != nil {
		s.Log.Error("Failed to save modification for trouble report %d: %v", tr.ID, err)
	}

	return nil
}

func (s *TroubleReports) Delete(id int64, user *models.User) error {
	s.Log.Debug("Deleting trouble report by %s: id: %d", user.String(), id)

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

func (s *TroubleReports) GetWithAttachments(id int64) (*models.TroubleReportWithAttachments, error) {
	s.Log.Debug("Getting trouble report with attachments: %d", id)

	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	attachments, err := s.LoadAttachments(tr)
	if err != nil {
		return nil, fmt.Errorf("failed to load attachments: %v", err)
	}

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

func (s *TroubleReports) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	s.Log.Debug("Getting all trouble reports with attachments")

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
	s.Log.Debug("Adding trouble report with attachments by %s: title: %s, attachments: %d",
		user.String(), troubleReport.Title, len(attachments))

	if err := user.Validate(); err != nil {
		return err
	}

	var attachmentIDs []int64
	for i, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, err := s.Registry.Attachments.Add(attachment)
		if err != nil {
			s.Log.Error("Failed to add attachment %d, cleaning up %d existing: %v",
				i+1, len(attachmentIDs), err)
			s.cleanupAttachments(attachmentIDs)
			return err
		}

		attachmentIDs = append(attachmentIDs, id)
	}

	troubleReport.LinkedAttachments = attachmentIDs

	id, err := s.Add(troubleReport, user)
	if err != nil {
		s.Log.Error("Failed to add trouble report, cleaning up %d attachments: %v",
			len(attachmentIDs), err)
		s.cleanupAttachments(attachmentIDs)
		return err
	}

	troubleReport.ID = id
	return nil
}

func (s *TroubleReports) UpdateWithAttachments(id int64, tr *models.TroubleReport, user *models.User, newAttachments ...*models.Attachment) error {
	s.Log.Debug("Updating trouble report with attachments by %s: id: %d, title: %s, new_attachments: %d",
		user.String(), id, tr.Title, len(newAttachments))

	if err := tr.Validate(); err != nil {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
	}

	var newAttachmentIDs []int64
	for i, attachment := range newAttachments {
		if attachment == nil {
			s.Log.Debug("Skipping nil attachment: index: %d", i+1)
			continue
		}

		s.Log.Debug("Adding new attachment: index: %d, size: %d bytes",
			i+1, len(attachment.Data))

		attachmentID, err := s.Registry.Attachments.Add(attachment)
		if err != nil {
			s.Log.Error("Failed to add new attachment %d, cleaning up %d existing: %v",
				i+1, len(newAttachmentIDs), err)
			s.cleanupAttachments(newAttachmentIDs)
			return err
		}

		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	originalAttachmentCount := len(tr.LinkedAttachments)
	tr.LinkedAttachments = append(tr.LinkedAttachments, newAttachmentIDs...)
	tr.ID = id

	s.Log.Debug("Combined attachments for update: id: %d, existing: %d, new: %d, total: %d",
		id, originalAttachmentCount, len(newAttachmentIDs), len(tr.LinkedAttachments))

	if err := s.Update(tr, user); err != nil {
		s.Log.Error("Failed to update trouble report %d, cleaning up %d new attachments: %v",
			id, len(newAttachmentIDs), err)
		s.cleanupAttachments(newAttachmentIDs)
		return err
	}

	return nil
}

func (s *TroubleReports) RemoveWithAttachments(id int64, user *models.User) (*models.TroubleReport, error) {
	s.Log.Debug("Removing trouble report with attachments by %s: id: %d", user.String(), id)

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
			s.Log.Error("Failed to delete attachment %d for trouble report %d: %v",
				attachmentID, tr.ID, err)
			failedDeletes++
		}
	}

	if failedDeletes > 0 {
		s.Log.Error("Failed to remove %d/%d attachments for trouble report %d",
			failedDeletes, len(tr.LinkedAttachments), tr.ID)
	}

	return tr, nil
}

func (s *TroubleReports) LoadAttachments(report *models.TroubleReport) ([]*models.Attachment, error) {
	s.Log.Debug("Loading attachments for trouble report: id: %d, attachments: %d",
		report.ID, len(report.LinkedAttachments))

	if err := report.Validate(); err != nil {
		return nil, err
	}

	var attachments []*models.Attachment
	for _, attachmentID := range report.LinkedAttachments {
		attachment, err := s.Registry.Attachments.Get(attachmentID)
		if err != nil {
			s.Log.Error("Failed to load attachment %d: %v", attachmentID, err)
			continue
		}
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

func (s *TroubleReports) cleanupAttachments(attachmentIDs []int64) {
	for _, id := range attachmentIDs {
		if err := s.Registry.Attachments.Delete(id); err != nil {
			s.Log.Error("Failed to cleanup attachment %d: %v", id, err)
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
		return nil, fmt.Errorf("failed to scan trouble report: %v", err)
	}

	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal linked attachments: %v", err)
	}

	return report, nil
}
