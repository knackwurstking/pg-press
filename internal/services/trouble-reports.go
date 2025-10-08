package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type TroubleReports struct {
	*BaseService
	attachments   *Attachments
	modifications *Modifications
}

func NewTroubleReports(db *sql.DB, a *Attachments, m *Modifications) *TroubleReports {
	base := NewBaseService(db, "Trouble Reports")

	troubleReports := &TroubleReports{
		BaseService:   base,
		attachments:   a,
		modifications: m,
	}

	if err := troubleReports.createTable(); err != nil {
		panic(err)
	}

	return troubleReports
}

func (s *TroubleReports) createTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			use_markdown BOOLEAN DEFAULT 0,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if err := s.CreateTable(query, "trouble_reports"); err != nil {
		return err
	}

	// Run migration to add use_markdown column if it doesn't exist
	return s.migrateUseMarkdownColumn()
}

func (s *TroubleReports) migrateUseMarkdownColumn() error {
	// Check if use_markdown column exists
	const checkColumnQuery = `PRAGMA table_info(trouble_reports)`
	rows, err := s.db.Query(checkColumnQuery)
	if err != nil {
		return fmt.Errorf("failed to check table schema: %v", err)
	}
	defer rows.Close()

	hasUseMarkdownColumn := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue interface{}

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return fmt.Errorf("failed to scan column info: %v", err)
		}

		if name == "use_markdown" {
			hasUseMarkdownColumn = true
			break
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error checking columns: %v", err)
	}

	// Add use_markdown column if it doesn't exist
	if !hasUseMarkdownColumn {
		s.LogOperation("Adding use_markdown column to existing trouble_reports table")
		const addColumnQuery = `ALTER TABLE trouble_reports ADD COLUMN use_markdown BOOLEAN DEFAULT 0`
		if _, err := s.db.Exec(addColumnQuery); err != nil {
			return fmt.Errorf("failed to add use_markdown column: %v", err)
		}
		s.LogOperation("Successfully added use_markdown column")
	}

	return nil
}

// List retrieves all trouble reports ordered by ID descending.
func (s *TroubleReports) List() ([]*models.TroubleReport, error) {
	s.LogOperation("Listing trouble reports")

	const listQuery = `SELECT * FROM trouble_reports ORDER BY id DESC`
	rows, err := s.db.Query(listQuery)
	if err != nil {
		return nil, s.HandleSelectError(err, "trouble_reports")
	}
	defer rows.Close()

	reports, err := ScanTroubleReportsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Listed trouble reports", fmt.Sprintf("count: %d", len(reports)))
	return reports, nil
}

// Get retrieves a specific trouble report by ID.
func (s *TroubleReports) Get(id int64) (*models.TroubleReport, error) {
	if err := ValidateID(id, "trouble_report"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting trouble report", id)

	const getQuery = `SELECT * FROM trouble_reports WHERE id = ?`
	row := s.db.QueryRow(getQuery, id)

	report, err := ScanSingleRow(row, ScanTroubleReport, "trouble_reports")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("trouble report with ID %d", id))
		}
		return nil, err
	}

	s.LogOperation("Retrieved trouble report", fmt.Sprintf("id: %d, title: %s", id, report.Title))
	return report, nil
}

// Add creates a new trouble report and generates a corresponding activity feed entry.
func (s *TroubleReports) Add(tr *models.TroubleReport, u *models.User) (int64, error) {
	if err := ValidateTroubleReport(tr); err != nil {
		return 0, err
	}

	if err := ValidateNotNil(u, "user"); err != nil {
		return 0, err
	}

	// Call the model's validate method for additional checks
	if err := tr.Validate(); err != nil {
		return 0, err
	}

	userInfo := createUserInfo(u)
	s.LogOperationWithUser("Adding trouble report", userInfo,
		fmt.Sprintf("title: %s, attachments: %d", tr.Title, len(tr.LinkedAttachments)))

	// Marshal linked attachments
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	// Insert trouble report
	const addQuery = `INSERT INTO trouble_reports (title, content, linked_attachments, use_markdown) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(addQuery, tr.Title, tr.Content, string(linkedAttachments), tr.UseMarkdown)
	if err != nil {
		return 0, s.HandleInsertError(err, "trouble_reports")
	}

	// Get ID of inserted trouble report
	id, err := result.LastInsertId()
	if err != nil {
		return 0, s.HandleInsertError(err, "trouble_reports")
	}
	tr.ID = id

	// Save initial modification
	modData := models.TroubleReportModData{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
		UseMarkdown:       tr.UseMarkdown,
	}
	if err := s.modifications.AddTroubleReportMod(u.TelegramID, id, modData); err != nil {
		s.log.Error("Failed to save initial modification for trouble report %d: %v", id, err)
		// Don't fail the entire operation for modification tracking
	}

	s.LogOperation("Added trouble report", fmt.Sprintf("id: %d", id))
	return id, nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (s *TroubleReports) Update(tr *models.TroubleReport, u *models.User) error {
	if err := ValidateTroubleReport(tr); err != nil {
		return err
	}

	if err := ValidateID(tr.ID, "trouble_report"); err != nil {
		return err
	}

	if err := ValidateNotNil(u, "user"); err != nil {
		return err
	}

	// Call the model's validate method for additional checks
	if err := tr.Validate(); err != nil {
		return err
	}

	userInfo := createUserInfo(u)
	s.LogOperationWithUser("Updating trouble report", userInfo,
		fmt.Sprintf("id: %d, title: %s, attachments: %d", tr.ID, tr.Title, len(tr.LinkedAttachments)))

	// Marshal linked attachments
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	const updateQuery = `UPDATE trouble_reports SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`
	result, err := s.db.Exec(updateQuery, tr.Title, tr.Content, string(linkedAttachments), tr.UseMarkdown, tr.ID)
	if err != nil {
		return s.HandleUpdateError(err, "trouble_reports")
	}

	if err := s.CheckRowsAffected(result, "trouble_report", tr.ID); err != nil {
		return err
	}

	// Save modification
	modData := models.TroubleReportModData{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
		UseMarkdown:       tr.UseMarkdown,
	}
	if err := s.modifications.AddTroubleReportMod(u.TelegramID, tr.ID, modData); err != nil {
		s.log.Error("Failed to save modification for trouble report %d: %v", tr.ID, err)
		// Don't fail the entire operation for modification tracking
	}

	s.LogOperation("Updated trouble report", fmt.Sprintf("id: %d", tr.ID))
	return nil
}

// Delete deletes a trouble report by ID and generates an activity feed entry.
func (s *TroubleReports) Delete(id int64, user *models.User) error {
	if err := ValidateID(id, "trouble_report"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	userInfo := createUserInfo(user)

	// Get the report before deleting for logging and cleanup
	report, err := s.Get(id)
	if err != nil {
		return err
	}

	s.LogOperationWithUser("Deleting trouble report", userInfo,
		fmt.Sprintf("id: %d, title: %s", id, report.Title))

	const deleteQuery = `DELETE FROM trouble_reports WHERE id = ?`
	result, err := s.db.Exec(deleteQuery, id)
	if err != nil {
		return s.HandleDeleteError(err, "trouble_reports")
	}

	if err := s.CheckRowsAffected(result, "trouble_report", id); err != nil {
		return err
	}

	// Delete modifications
	if err := s.modifications.DeleteAll(ModificationTypeTroubleReport, report.ID); err != nil {
		s.log.Error("Failed to delete modifications for trouble report %d: %v", report.ID, err)
	}

	s.LogOperation("Deleted trouble report", fmt.Sprintf("id: %d", id))
	return nil
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *TroubleReports) GetWithAttachments(id int64) (*models.TroubleReportWithAttachments, error) {
	if err := ValidateID(id, "trouble_report"); err != nil {
		return nil, err
	}

	s.LogOperation("Getting trouble report with attachments", id)

	// Get the trouble report
	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Load attachments
	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Retrieved trouble report with attachments",
		fmt.Sprintf("id: %d, attachments: %d", id, len(attachments)))

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (s *TroubleReports) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	s.LogOperation("Listing trouble reports with attachments")

	// Get all trouble reports
	reports, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []*models.TroubleReportWithAttachments
	totalAttachments := 0

	for _, tr := range reports {
		// Load attachments for each report
		attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
		if err != nil {
			return nil, err
		}

		totalAttachments += len(attachments)
		result = append(result, &models.TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	s.LogOperation("Listed trouble reports with attachments",
		fmt.Sprintf("reports: %d, total_attachments: %d", len(result), totalAttachments))

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (s *TroubleReports) AddWithAttachments(troubleReport *models.TroubleReport, user *models.User, attachments ...*models.Attachment) error {
	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	userInfo := createUserInfo(user)
	s.LogOperationWithUser("Adding trouble report with attachments", userInfo,
		fmt.Sprintf("title: %s, attachments: %d", troubleReport.Title, len(attachments)))

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for i, attachment := range attachments {
		if attachment == nil {
			s.LogOperation("Skipping nil attachment", fmt.Sprintf("index: %d", i))
			continue
		}

		s.LogOperation("Adding attachment", fmt.Sprintf("index: %d, size: %d bytes", i+1, len(attachment.Data)))

		id, err := s.attachments.Add(attachment)
		if err != nil {
			// Cleanup already added attachments on failure
			s.log.Error("Failed to add attachment %d, cleaning up %d existing: %v", i+1, len(attachmentIDs), err)
			for _, addedID := range attachmentIDs {
				if cleanupErr := s.attachments.Delete(addedID); cleanupErr != nil {
					s.log.Error("Failed to cleanup attachment %d: %v", addedID, cleanupErr)
				}
			}
			return err
		}

		attachmentIDs = append(attachmentIDs, id)
	}

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Add the trouble report
	if _, err := s.Add(troubleReport, user); err != nil {
		s.log.Error("Failed to add trouble report, cleaning up %d attachments: %v", len(attachmentIDs), err)

		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			if cleanupErr := s.attachments.Delete(id); cleanupErr != nil {
				s.log.Error("Failed to cleanup attachment %d: %v", id, cleanupErr)
			}
		}
		return err
	}

	s.LogOperation("Added trouble report with attachments",
		fmt.Sprintf("attachments_added: %d", len(attachmentIDs)))
	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (s *TroubleReports) UpdateWithAttachments(id int64, tr *models.TroubleReport, user *models.User, newAttachments ...*models.Attachment) error {
	if err := ValidateID(id, "trouble_report"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	userInfo := createUserInfo(user)
	s.LogOperationWithUser("Updating trouble report with attachments", userInfo,
		fmt.Sprintf("id: %d, title: %s, new_attachments: %d", id, tr.Title, len(newAttachments)))

	// Add new attachments
	var newAttachmentIDs []int64
	for i, attachment := range newAttachments {
		if attachment == nil {
			s.LogOperation("Skipping nil attachment", fmt.Sprintf("index: %d", i))
			continue
		}

		s.LogOperation("Adding new attachment", fmt.Sprintf("index: %d, size: %d bytes", i+1, len(attachment.Data)))

		attachmentID, err := s.attachments.Add(attachment)
		if err != nil {
			// Cleanup already added attachments on failure
			s.log.Error("Failed to add new attachment %d, cleaning up %d existing: %v", i+1, len(newAttachmentIDs), err)
			for _, addedID := range newAttachmentIDs {
				if cleanupErr := s.attachments.Delete(addedID); cleanupErr != nil {
					s.log.Error("Failed to cleanup new attachment %d: %v", addedID, cleanupErr)
				}
			}
			return err
		}

		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	// Combine existing and new attachment IDs
	originalAttachmentCount := len(tr.LinkedAttachments)
	allAttachmentIDs := append(tr.LinkedAttachments, newAttachmentIDs...)
	tr.LinkedAttachments = allAttachmentIDs
	tr.ID = id

	s.LogOperation("Combined attachments for update",
		fmt.Sprintf("id: %d, existing: %d, new: %d, total: %d",
			id, originalAttachmentCount, len(newAttachmentIDs), len(allAttachmentIDs)))

	// Update the trouble report
	if err := s.Update(tr, user); err != nil {
		s.log.Error("Failed to update trouble report %d, cleaning up %d new attachments: %v", id, len(newAttachmentIDs), err)

		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			if cleanupErr := s.attachments.Delete(attachmentID); cleanupErr != nil {
				s.log.Error("Failed to cleanup new attachment %d: %v", attachmentID, cleanupErr)
			}
		}
		return err
	}

	s.LogOperation("Updated trouble report with attachments",
		fmt.Sprintf("id: %d, new_attachments_added: %d", id, len(newAttachmentIDs)))
	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (s *TroubleReports) RemoveWithAttachments(id int64, user *models.User) (*models.TroubleReport, error) {
	if err := ValidateID(id, "trouble_report"); err != nil {
		return nil, err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	userInfo := createUserInfo(user)
	s.LogOperationWithUser("Removing trouble report with attachments", userInfo, fmt.Sprintf("id: %d", id))

	// Get the trouble report to find its attachments
	tr, err := s.Get(id)
	if err != nil {
		return tr, err
	}

	s.LogOperation("Retrieved trouble report for removal",
		fmt.Sprintf("id: %d, title: %s, attachments: %d", id, tr.Title, len(tr.LinkedAttachments)))

	// Remove the trouble report first
	if err := s.Delete(id, user); err != nil {
		return tr, err
	}

	// Remove associated attachments
	successfulAttachmentDeletes := 0
	failedAttachmentDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Delete(attachmentID); err != nil {
			s.log.Warn("Failed to remove attachment %d for trouble report %d: %v", attachmentID, id, err)
			failedAttachmentDeletes++
		} else {
			successfulAttachmentDeletes++
		}
	}

	if failedAttachmentDeletes > 0 {
		s.log.Warn("Failed to remove %d/%d attachments for trouble report %d",
			failedAttachmentDeletes, len(tr.LinkedAttachments), id)
	}

	s.LogOperation("Removed trouble report with attachments",
		fmt.Sprintf("id: %d, attachments_removed: %d, attachments_failed: %d",
			id, successfulAttachmentDeletes, failedAttachmentDeletes))

	return tr, nil
}

// LoadAttachments loads attachments for a trouble report.
func (s *TroubleReports) LoadAttachments(tr *models.TroubleReport) ([]*models.Attachment, error) {
	if err := ValidateTroubleReport(tr); err != nil {
		return nil, err
	}

	s.LogOperation("Loading attachments for trouble report",
		fmt.Sprintf("id: %d, attachment_count: %d", tr.ID, len(tr.LinkedAttachments)))

	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		return nil, err
	}

	s.LogOperation("Loaded attachments for trouble report",
		fmt.Sprintf("id: %d, loaded: %d/%d", tr.ID, len(attachments), len(tr.LinkedAttachments)))

	return attachments, nil
}
