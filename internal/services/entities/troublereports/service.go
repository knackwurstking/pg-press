package troublereports

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
	attachments   AttachmentsService
	modifications ModificationsService
}

// AttachmentsService defines the interface for attachments service methods
type AttachmentsService interface {
	Add(attachment *models.Attachment) (int64, error)
	Get(id int64) (*models.Attachment, error)
	Delete(id int64) error
}

// ModificationsService defines the interface for modifications service methods
type ModificationsService interface {
	Add(userID int64, entityType models.ModificationType, entityID int64, data any) (*models.Modification[any], error)
}

func NewService(db *sql.DB, a AttachmentsService, m ModificationsService) *Service {
	baseService := base.NewBaseService(db, "Trouble Reports")

	troubleReports := &Service{
		BaseService:   baseService,
		attachments:   a,
		modifications: m,
	}

	if err := troubleReports.createTable(); err != nil {
		panic(err)
	}

	return troubleReports
}

func (s *Service) createTable() error {
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

func (s *Service) migrateUseMarkdownColumn() error {
	// Check if use_markdown column exists
	const checkColumnQuery = `PRAGMA table_info(trouble_reports)`
	rows, err := s.DB.Query(checkColumnQuery)
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
		s.Log.Info("Adding use_markdown column to existing trouble_reports table")
		const addColumnQuery = `ALTER TABLE trouble_reports ADD COLUMN use_markdown BOOLEAN DEFAULT 0`
		if _, err := s.DB.Exec(addColumnQuery); err != nil {
			return fmt.Errorf("failed to add use_markdown column: %v", err)
		}
		s.Log.Info("Successfully added use_markdown column")
	}

	return nil
}

// List retrieves all trouble reports ordered by ID descending.
func (s *Service) List() ([]*models.TroubleReport, error) {
	s.Log.Info("Listing trouble reports")

	const listQuery = `SELECT * FROM trouble_reports ORDER BY id DESC`
	rows, err := s.DB.Query(listQuery)
	if err != nil {
		return nil, s.HandleSelectError(err, "trouble_reports")
	}
	defer rows.Close()

	reports, err := ScanTroubleReportsFromRows(rows)
	if err != nil {
		return nil, err
	}

	s.Log.Info("Listed trouble reports: count: %d", len(reports))
	return reports, nil
}

// Get retrieves a specific trouble report by ID.
func (s *Service) Get(id int64) (*models.TroubleReport, error) {
	if err := validation.ValidateID(id, "trouble_report"); err != nil {
		return nil, err
	}

	s.Log.Info("Getting trouble report: %d", id)

	const getQuery = `SELECT * FROM trouble_reports WHERE id = ?`
	row := s.DB.QueryRow(getQuery, id)

	report, err := scanner.ScanSingleRow(row, ScanTroubleReport, "trouble_reports")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("trouble report with ID %d", id))
		}
		return nil, err
	}

	s.Log.Info("Retrieved trouble report: id: %d, title: %s", id, report.Title)
	return report, nil
}

// Add creates a new trouble report and generates a corresponding activity feed entry.
func (s *Service) Add(tr *models.TroubleReport, u *models.User) (int64, error) {
	if err := ValidateTroubleReport(tr); err != nil {
		return 0, err
	}

	if err := validation.ValidateNotNil(u, "user"); err != nil {
		return 0, err
	}

	// Call the model's validate method for additional checks
	if err := tr.Validate(); err != nil {
		return 0, err
	}

	s.Log.Info("Adding trouble report by %s: title: %s, attachments: %d", u.String(), tr.Title, len(tr.LinkedAttachments))

	// Marshal linked attachments
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	// Insert trouble report
	const insertQuery = `INSERT INTO trouble_reports (title, content, linked_attachments, use_markdown) VALUES (?, ?, ?, ?)`
	result, err := s.DB.Exec(insertQuery, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown)
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
	if _, err := s.modifications.Add(
		u.TelegramID, models.ModificationTypeTroubleReport, id, modData,
	); err != nil {
		s.Log.Error("Failed to save initial modification for trouble report %d: %v", id, err)
		// Don't fail the entire operation for modification tracking
	}

	s.Log.Info("Added trouble report: id: %d", id)
	return id, nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
// Update updates an existing trouble report in the database.
func (s *Service) Update(tr *models.TroubleReport, u *models.User) error {
	if err := ValidateTroubleReport(tr); err != nil {
		return err
	}

	if err := validation.ValidateID(tr.ID, "trouble_report"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(u, "user"); err != nil {
		return err
	}

	// Call the model's validate method for additional checks
	if err := tr.Validate(); err != nil {
		return err
	}

	s.Log.Info("Updating trouble report by %s: id: %d, title: %s, attachments: %d", u.String(), tr.ID, tr.Title, len(tr.LinkedAttachments))

	// Marshal linked attachments
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	const updateQuery = `UPDATE trouble_reports SET title = ?, content = ?, linked_attachments = ?, use_markdown = ? WHERE id = ?`
	result, err := s.DB.Exec(updateQuery, tr.Title, tr.Content, linkedAttachments, tr.UseMarkdown, tr.ID)
	if err != nil {
		return s.HandleUpdateError(err, "trouble_reports")
	}

	if err := s.CheckRowsAffected(result, "trouble_report", tr.ID); err != nil {
		return err
	}

	// Save modification
	// Save modification for tracking changes
	modData := models.TroubleReportModData{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
		UseMarkdown:       tr.UseMarkdown,
	}
	if _, err := s.modifications.Add(
		u.TelegramID,
		models.ModificationTypeTroubleReport,

		tr.ID,
		modData,
	); err != nil {
		s.Log.Error("Failed to save modification for trouble report %d: %v", tr.ID, err)
		// Don't fail the entire operation for modification tracking
	}

	s.Log.Info("Updated trouble report: id: %d", tr.ID)
	return nil
}

// Delete deletes a trouble report by ID and generates an activity feed entry.
// Delete removes a trouble report from the database.
func (s *Service) Delete(id int64, user *models.User) error {
	if err := validation.ValidateID(id, "trouble_report"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	// Get the report before deleting for logging and cleanup
	report, err := s.Get(id)
	if err != nil {
		return err
	}

	s.Log.Info("Deleting trouble report by %s: id: %d, title: %s", user.String(), id, report.Title)

	const deleteQuery = `DELETE FROM trouble_reports WHERE id = ?`
	result, err := s.DB.Exec(deleteQuery, id)
	if err != nil {
		return s.HandleDeleteError(err, "trouble_reports")
	}

	if err := s.CheckRowsAffected(result, "trouble_report", id); err != nil {
		return err
	}

	// Delete modifications
	// Note: Modifications service may not have DeleteAll method, skipping for now
	// if err := s.modifications.DeleteAll(modifications.ModificationTypeTroubleReport, report.ID); err != nil {
	//	s.Log.Error("Failed to delete modifications for trouble report %d: %v", report.ID, err)
	// }

	s.Log.Info("Deleted trouble report: id: %d", id)
	return nil
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *Service) GetWithAttachments(id int64) (*models.TroubleReportWithAttachments, error) {
	if err := validation.ValidateID(id, "trouble_report"); err != nil {
		return nil, err
	}

	s.Log.Info("Getting trouble report with attachments: %d", id)

	// Get the trouble report
	tr, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	// Load attachments
	attachments, err := s.LoadAttachments(tr)
	if err != nil {
		return nil, fmt.Errorf("failed to load attachments: %v", err)
	}

	s.Log.Info("Retrieved trouble report with attachments: id: %d, attachments: %d", id, len(attachments))

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (s *Service) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	s.Log.Info("Getting all trouble reports with attachments")

	// Get all trouble reports
	reports, err := s.List()
	if err != nil {
		return nil, err
	}

	var result []*models.TroubleReportWithAttachments
	totalAttachments := 0

	for _, tr := range reports {
		// Load attachments for each report
		attachments, err := s.LoadAttachments(tr)
		if err != nil {
			return nil, fmt.Errorf("failed to load attachments for report %d: %v", tr.ID, err)
		}

		totalAttachments += len(attachments)
		result = append(result, &models.TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	s.Log.Info("Listed trouble reports with attachments: reports: %d, total_attachments: %d", len(result), totalAttachments)

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
// AddWithAttachments creates a trouble report with attachments in a single transaction
func (s *Service) AddWithAttachments(troubleReport *models.TroubleReport, user *models.User, attachments ...*models.Attachment) error {
	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	s.Log.Info("Adding trouble report with attachments by %s: title: %s, attachments: %d", user.String(), troubleReport.Title, len(attachments))

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for i, attachment := range attachments {
		if attachment == nil {
			s.Log.Info("Skipping nil attachment: index: %d", i+1)
			continue
		}

		s.Log.Info("Adding attachment: index: %d, size: %d bytes", i+1, len(attachment.Data))

		id, err := s.attachments.Add(attachment)
		if err != nil {
			// Cleanup already added attachments on failure
			s.Log.Error("Failed to add attachment %d, cleaning up %d existing: %v", i+1, len(attachmentIDs), err)
			// Cleanup attachments on failure
			for _, attachmentID := range attachmentIDs {
				if cleanupErr := s.attachments.Delete(attachmentID); cleanupErr != nil {
					s.Log.Error("Failed to cleanup attachment %d: %v", attachmentID, cleanupErr)
				}
			}
			return err
		}

		attachmentIDs = append(attachmentIDs, id)
	}

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Add the trouble report
	id, err := s.Add(troubleReport, user)
	if err != nil {
		s.Log.Error("Failed to add trouble report, cleaning up %d attachments: %v", len(attachmentIDs), err)

		// Cleanup attachments on failure
		for _, attachmentID := range attachmentIDs {
			if cleanupErr := s.attachments.Delete(attachmentID); cleanupErr != nil {
				s.Log.Error("Failed to cleanup attachment %d: %v", attachmentID, cleanupErr)
			}
		}
		return err
	}

	troubleReport.ID = id

	s.Log.Info("Added trouble report with attachments: id: %d, attachments: %d", troubleReport.ID, len(attachmentIDs))
	return nil
}

// UpdateWithAttachments updates a trouble report with new attachments
func (s *Service) UpdateWithAttachments(id int64, tr *models.TroubleReport, user *models.User, newAttachments ...*models.Attachment) error {
	if err := validation.ValidateID(id, "trouble_report"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	s.Log.Info("Updating trouble report with attachments by %s: id: %d, title: %s, new_attachments: %d", user.String(), id, tr.Title, len(newAttachments))

	// Add new attachments
	var newAttachmentIDs []int64
	for i, attachment := range newAttachments {
		if attachment == nil {
			s.Log.Info("Skipping nil attachment: index: %d", i+1)
			continue
		}

		s.Log.Info("Adding new attachment: index: %d, size: %d bytes", i+1, len(attachment.Data))

		attachmentID, err := s.attachments.Add(attachment)
		if err != nil {
			// Cleanup already added attachments on failure
			s.Log.Error("Failed to add new attachment %d, cleaning up %d existing: %v", i+1, len(newAttachmentIDs), err)
			for _, addedID := range newAttachmentIDs {
				if cleanupErr := s.attachments.Delete(addedID); cleanupErr != nil {
					s.Log.Error("Failed to cleanup new attachment %d: %v", addedID, cleanupErr)
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

	s.Log.Info("Combined attachments for update: id: %d, existing: %d, new: %d, total: %d",
		id, originalAttachmentCount, len(newAttachmentIDs), len(allAttachmentIDs))

	// Update the trouble report
	if err := s.Update(tr, user); err != nil {
		s.Log.Error("Failed to update trouble report %d, cleaning up %d new attachments: %v", id, len(newAttachmentIDs), err)

		// Cleanup new attachments on failure
		for _, attachmentID := range tr.LinkedAttachments {
			if err := s.attachments.Delete(attachmentID); err != nil {
				s.Log.Error("Failed to delete attachment %d: %v", attachmentID, err)
			}
		}
		return err
	}

	s.Log.Info("Updated trouble report with attachments: id: %d, new_attachments: %d", tr.ID, len(newAttachmentIDs))
	return nil
}

// RemoveWithAttachments removes a trouble report and its associated attachments
func (s *Service) RemoveWithAttachments(id int64, user *models.User) (*models.TroubleReport, error) {
	if err := validation.ValidateID(id, "trouble_report"); err != nil {
		return nil, err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	s.Log.Info("Removing trouble report with attachments by %s: id: %d", user.String(), id)

	// Get the trouble report to find its attachments
	tr, err := s.Get(id)
	if err != nil {
		return tr, err
	}

	s.Log.Info("Retrieved trouble report for removal: id: %d, title: %s, attachments: %d", id, tr.Title, len(tr.LinkedAttachments))

	// Remove the trouble report first
	if err := s.Delete(id, user); err != nil {
		return tr, err
	}

	// Remove associated attachments
	successfulAttachmentDeletes := 0
	failedAttachmentDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Delete(attachmentID); err != nil {
			s.Log.Error("Failed to delete attachment %d for trouble report %d: %v", attachmentID, tr.ID, err)
			failedAttachmentDeletes++
		} else {
			successfulAttachmentDeletes++
		}
	}

	if failedAttachmentDeletes > 0 {
		s.Log.Error("Failed to remove %d/%d attachments for trouble report %d",
			failedAttachmentDeletes, len(tr.LinkedAttachments), tr.ID)
	}

	s.Log.Info("Removed trouble report with attachments: id: %d, attachments_removed: %d, attachments_failed: %d",
		tr.ID, successfulAttachmentDeletes, failedAttachmentDeletes)

	return tr, nil
}

// LoadAttachments loads attachments for a given trouble report
func (s *Service) LoadAttachments(report *models.TroubleReport) ([]*models.Attachment, error) {
	if err := validation.ValidateID(report.ID, "trouble_report"); err != nil {
		return nil, err
	}

	s.Log.Info("Loading attachments for trouble report: id: %d, attachments: %d", report.ID, len(report.LinkedAttachments))

	// Load attachments individually
	var attachments []*models.Attachment
	for _, attachmentID := range report.LinkedAttachments {
		attachment, err := s.attachments.Get(attachmentID)
		if err != nil {
			s.Log.Error("Failed to load attachment %d: %v", attachmentID, err)
			continue
		}
		attachments = append(attachments, attachment)
	}

	s.Log.Info("Loaded attachments for trouble report: id: %d, loaded: %d", report.ID, len(attachments))
	return attachments, nil
}
