package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type TroubleReports struct {
	db            *sql.DB
	attachments   *Attachment
	modifications *Modifications
}

func NewTroubleReports(db *sql.DB, a *Attachment, m *Modifications) *TroubleReports {
	troubleReport := &TroubleReports{
		db:            db,
		attachments:   a,
		modifications: m,
	}

	if err := troubleReport.createTable(db); err != nil {
		panic(err)
	}

	return troubleReport
}

func (s *TroubleReports) createTable(db *sql.DB) error {
	const createQuery string = `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	if _, err := db.Exec(createQuery); err != nil {
		return fmt.Errorf("failed to create trouble_reports table: %v", err)
	}

	return nil
}

// List retrieves all trouble reports ordered by ID descending.
func (s *TroubleReports) List() ([]*models.TroubleReport, error) {
	logger.DBTroubleReports().Debug("Starting trouble reports list query")

	const listQuery string = `SELECT * FROM trouble_reports ORDER BY id DESC`
	rows, err := s.db.Query(listQuery)
	if err != nil {
		logger.DBTroubleReports().Error("Failed to execute trouble reports list query: %v", err)
		return nil, fmt.Errorf("select error: trouble_reports: %v", err)
	}
	defer rows.Close()

	var reports []*models.TroubleReport
	reportCount := 0

	for rows.Next() {
		report, err := s.scanTroubleReport(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: trouble_reports: %v", err)
		}
		reports = append(reports, report)
		reportCount++
	}

	if err := rows.Err(); err != nil {
		logger.DBTroubleReports().Error("Error iterating over %d trouble report rows: %v", reportCount, err)
		return nil, fmt.Errorf("select error: trouble_reports: %v", err)
	}

	logger.DBTroubleReports().Info("Listed %d trouble reports", len(reports))
	return reports, nil
}

// Get retrieves a specific trouble report by ID.
func (s *TroubleReports) Get(id int64) (*models.TroubleReport, error) {
	logger.DBTroubleReports().Debug("Getting trouble report by ID: %d", id)

	const getQuery = `SELECT * FROM trouble_reports WHERE id = ?`
	row := s.db.QueryRow(getQuery, id)

	report, err := s.scanTroubleReport(row)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("trouble report with ID %d", id))
		}

		return nil, fmt.Errorf("select error: trouble_reports: %v", err)
	}

	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s')", id, report.Title)
	return report, nil
}

// Add creates a new trouble report and generates a corresponding activity feed entry.
func (s *TroubleReports) Add(tr *models.TroubleReport, u *models.User) (int64, error) {
	// Validate
	if u == nil {
		return 0, utils.NewValidationError("user: user cannot be nil")
	}

	if err := tr.Validate(); err != nil {
		return 0, err
	}

	userInfo := createUserInfo(u)

	logger.DBTroubleReports().Info("Adding trouble report by %s: title='%s', attachments=%d",
		userInfo, tr.Title, len(tr.LinkedAttachments))

	// Marshal linked attachments
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	// Insert trouble report
	const addQuery = `INSERT INTO trouble_reports
		(title, content, linked_attachments) VALUES (?, ?, ?)`

	result, err := s.db.Exec(
		addQuery, tr.Title, tr.Content, string(linkedAttachments),
	)
	if err != nil {
		return 0, fmt.Errorf("insert error: trouble_reports: %v", err)
	}

	// Get ID of inserted trouble report
	id, err := result.LastInsertId()
	if err != nil {
		logger.DBTroubleReports().Error("Failed to get last insert ID for trouble report by %s: %v", userInfo, err)
		return 0, fmt.Errorf("insert error: trouble_reports: %v", err)
	}
	tr.ID = id

	// Save initial modification
	modData := models.TroubleReportModData{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}
	if err := s.modifications.AddTroubleReportMod(u.TelegramID, id, modData); err != nil {
		logger.DBTroubleReports().Error("Failed to save initial modification for trouble report %d: %v", id, err)
		// Don't fail the entire operation for modification tracking
	}

	logger.DBTroubleReports().Info("Successfully added trouble report %d by %s", id, userInfo)

	return id, nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (s *TroubleReports) Update(tr *models.TroubleReport, u *models.User) error {
	// Validate
	if u == nil {
		return utils.NewValidationError("user: user cannot be nil")
	}

	if err := tr.Validate(); err != nil {
		return err
	}

	userInfo := createUserInfo(u)

	logger.DBTroubleReports().Info("Updating trouble report %d by %s: title='%s', attachments=%d",
		tr.ID, userInfo, tr.Title, len(tr.LinkedAttachments))

	// Unmarshal linked attachments
	linkedAttachments, err := json.Marshal(tr.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %v", err)
	}

	const updateQuery string = `UPDATE trouble_reports
		SET title = ?, content = ?, linked_attachments = ? WHERE id = ?`
	_, err = s.db.Exec(
		updateQuery,
		tr.Title, tr.Content, string(linkedAttachments), // SET
		tr.ID, // WHERE
	)
	if err != nil {
		return fmt.Errorf("update error: trouble_reports: %v", err)
	}

	// Save modification
	modData := models.TroubleReportModData{
		Title:             tr.Title,
		Content:           tr.Content,
		LinkedAttachments: tr.LinkedAttachments,
	}
	if err := s.modifications.AddTroubleReportMod(u.TelegramID, tr.ID, modData); err != nil {
		logger.DBTroubleReports().Error(
			"Failed to save modification for trouble report %d: %v",
			tr.ID, err,
		)
		// Don't fail the entire operation for modification tracking
	}

	logger.DBTroubleReports().Info("Successfully updated trouble report %d by %s",
		tr.ID, userInfo)

	return nil
}

// Delete deletes a trouble report by ID and generates an activity feed entry.
func (s *TroubleReports) Delete(id int64, user *models.User) error {
	if user == nil {
		return utils.NewValidationError("user: user cannot be nil")
	}

	userInfo := createUserInfo(user)

	logger.DBTroubleReports().Info("Removing trouble report %d by %s", id, userInfo)

	// Get the report before deleting for the feed entry and logging
	report, err := s.Get(id)
	if err != nil {
		return err
	}
	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s') for deletion", id, report.Title)

	const deleteQuery = `DELETE FROM trouble_reports WHERE id = ?`
	result, err := s.db.Exec(deleteQuery, id)
	if err != nil {
		return fmt.Errorf("delete error: trouble_reports: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete error: trouble_reports: %v", err)
	}

	if rowsAffected == 0 {
		return utils.NewNotFoundError(fmt.Sprintf("no trouble report found with ID %d", id))
	}

	// Delete modification
	if err := s.modifications.DeleteAll(ModificationTypeTroubleReport, report.ID); err != nil {
		logger.DBTroubleReports().Error("Failed to delete all modifications for entity type \"%s\" and entity ID %d",
			ModificationTypeTroubleReport, report.ID)
	}

	logger.DBTroubleReports().Info(
		"Successfully deleted trouble report %d (title='%s') by %s",
		id, report.Title, userInfo,
	)

	return nil
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *TroubleReports) GetWithAttachments(
	id int64,
) (*models.TroubleReportWithAttachments, error) {
	logger.DBTroubleReports().Debug("Getting trouble report with attachments, id: %d", id)

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

	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s') with %d attachments",
		id, tr.Title, len(attachments))

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (s *TroubleReports) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	logger.DBTroubleReports().Debug("Starting trouble reports with attachments list query")

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

	logger.DBTroubleReports().Info(
		"Listed %d trouble reports with %d total attachments",
		len(result), totalAttachments,
	)

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (s *TroubleReports) AddWithAttachments(
	troubleReport *models.TroubleReport, user *models.User, attachments ...*models.Attachment,
) error {
	if user == nil {
		return utils.NewValidationError("user: user cannot be nil")
	}

	userInfo := createUserInfo(user)

	logger.DBTroubleReports().Info("Adding trouble report with %d attachments by %s: title='%s'",
		len(attachments), userInfo, troubleReport.Title)

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for i, attachment := range attachments {
		if attachment == nil {
			logger.DBTroubleReports().Debug("Skipping nil attachment %d for %s", i, userInfo)
			continue
		}

		logger.DBTroubleReports().Debug("Adding attachment %d/%d (size: %d bytes) for %s",
			i+1, len(attachments), len(attachment.Data), userInfo)

		id, err := s.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			logger.DBTroubleReports().Error("Failed to add attachment %d for %s, cleaning up %d existing: %v",
				i+1, userInfo, len(attachmentIDs), err)

			for _, addedID := range attachmentIDs {
				if cleanupErr := s.attachments.Delete(addedID, user); cleanupErr != nil {
					logger.DBTroubleReports().Error("Failed to cleanup attachment %d: %v", addedID, cleanupErr)
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
		logger.DBTroubleReports().Error(
			"Failed to add trouble report for %s, cleaning up %d attachments: %v",
			userInfo, len(attachmentIDs), err,
		)

		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			if cleanupErr := s.attachments.Delete(id, user); cleanupErr != nil {
				logger.DBTroubleReports().Error("Failed to cleanup attachment %d: %v", id, cleanupErr)
			}
		}

		return err
	}

	logger.DBTroubleReports().Info("Successfully added trouble report %d with %d attachments by %s",
		troubleReport.ID, len(attachmentIDs), userInfo)

	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (s *TroubleReports) UpdateWithAttachments(
	id int64,
	tr *models.TroubleReport,
	user *models.User,
	newAttachments ...*models.Attachment,
) error {
	if user == nil {
		return utils.NewValidationError("user: user cannot be nil")
	}

	userInfo := createUserInfo(user)

	logger.DBTroubleReports().Info(
		"Updating trouble report %d with %d new attachments by %s: title='%s'",
		id, len(newAttachments), userInfo, tr.Title,
	)

	// Add new attachments
	var newAttachmentIDs []int64
	for i, attachment := range newAttachments {
		if attachment == nil {
			logger.DBTroubleReports().Debug("Skipping nil attachment %d for update %d by %s", i, id, userInfo)
			continue
		}

		logger.DBTroubleReports().Debug("Adding new attachment %d/%d (size: %d bytes) for update %d by %s",
			i+1, len(newAttachments), len(attachment.Data), id, userInfo)

		attachmentID, err := s.attachments.Add(attachment, user)

		if err != nil {
			// Cleanup already added attachments on failure
			logger.DBTroubleReports().Error("Failed to add new attachment %d for update %d by %s, cleaning up %d existing: %v",
				i+1, id, userInfo, len(newAttachmentIDs), err)

			for _, addedID := range newAttachmentIDs {
				if cleanupErr := s.attachments.Delete(addedID, user); cleanupErr != nil {
					logger.DBTroubleReports().Error("Failed to cleanup new attachment %d: %v", addedID, cleanupErr)
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

	logger.DBTroubleReports().Debug("Combined attachments for update %d: %d existing + %d new = %d total",
		id, originalAttachmentCount, len(newAttachmentIDs), len(allAttachmentIDs))

	// Update the trouble report
	if err := s.Update(tr, user); err != nil {
		logger.DBTroubleReports().Error("Failed to update trouble report %d by %s, cleaning up %d new attachments: %v",
			id, userInfo, len(newAttachmentIDs), err)

		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			if cleanupErr := s.attachments.Delete(attachmentID, user); cleanupErr != nil {
				logger.DBTroubleReports().Error("Failed to cleanup new attachment %d: %v", attachmentID, cleanupErr)
			}
		}

		return err
	}

	logger.DBTroubleReports().Info(
		"Successfully updated trouble report %d with %d new attachments by %s",
		id, len(newAttachmentIDs), userInfo,
	)

	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (s *TroubleReports) RemoveWithAttachments(id int64, user *models.User) (*models.TroubleReport, error) {
	if user == nil {
		return nil, utils.NewValidationError("user: user cannot be nil")
	}

	userInfo := createUserInfo(user)

	logger.DBTroubleReports().Info("Removing trouble report %d with attachments by %s", id, userInfo)

	// Get the trouble report to find its attachments
	tr, err := s.Get(id)
	if err != nil {
		return tr, err
	}

	logger.DBTroubleReports().Debug(
		"Retrieved trouble report %d (title='%s') with %d attachments for removal by %s",
		id, tr.Title, len(tr.LinkedAttachments), userInfo,
	)

	// Remove the trouble report first
	if err := s.Delete(id, user); err != nil {
		return tr, err
	}

	// Remove associated attachments
	successfulAttachmentDeletes := 0
	failedAttachmentDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Delete(attachmentID, user); err != nil {
			logger.DBTroubleReports().Warn(
				"Failed to remove attachment %d for trouble report %d by %s: %v",
				attachmentID, id, userInfo, err,
			)

			failedAttachmentDeletes++
		} else {
			successfulAttachmentDeletes++
		}
	}

	logger.DBTroubleReports().Info(
		"Successfully removed trouble report %d (title='%s') with %d/%d attachments by %s",
		id, tr.Title, successfulAttachmentDeletes, len(tr.LinkedAttachments), userInfo,
	)

	if failedAttachmentDeletes > 0 {
		logger.DBTroubleReports().Warn(
			"Failed to remove %d/%d attachments for trouble report %d",
			failedAttachmentDeletes, len(tr.LinkedAttachments), id,
		)
	}

	return tr, nil
}

// LoadAttachments loads attachments for a trouble report.
func (s *TroubleReports) LoadAttachments(tr *models.TroubleReport) ([]*models.Attachment, error) {
	logger.DBTroubleReports().Debug(
		"Loading %d attachments for trouble report %d (title='%s')",
		len(tr.LinkedAttachments), tr.ID, tr.Title,
	)

	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		return nil, err
	}

	logger.DBTroubleReports().Debug("Successfully loaded %d/%d attachments for trouble report %d",
		len(attachments), len(tr.LinkedAttachments), tr.ID)

	return attachments, nil
}

func (s *TroubleReports) scanTroubleReport(scanner interfaces.Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string

	if err := scanner.Scan(&report.ID, &report.Title, &report.Content, &linkedAttachments); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}

		return nil, fmt.Errorf("failed to scan row: %v", err)
	}

	// Try to unmarshal as new format (array of int64 IDs) first
	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		logger.DBTroubleReports().Error("Failed to unmarshal linked attachments for report %d: %v", report.ID, err)
		return nil, fmt.Errorf("failed to unmarshal linked attachments: %v", err)
	}

	return report, nil
}
