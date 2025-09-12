package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type TroubleReport struct {
	db          *sql.DB
	attachments *Attachment
	feeds       *Feed
}

func NewTroubleReport(db *sql.DB, attachments *Attachment, feeds *Feed) *TroubleReport {
	troubleReport := &TroubleReport{
		db:          db,
		attachments: attachments,
		feeds:       feeds,
	}

	if err := troubleReport.createTable(db); err != nil {
		panic(err)
	}

	return troubleReport
}

// TODO: Remove mods from this table
func (s *TroubleReport) createTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS trouble_reports (
			id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			linked_attachments TEXT NOT NULL,
			mods BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create trouble_reports table: %w", err)
	}

	return nil
}

// List retrieves all trouble reports ordered by ID descending.
func (s *TroubleReport) List() ([]*models.TroubleReport, error) {
	logger.DBTroubleReports().Debug("Starting trouble reports list query")
	start := time.Now()

	query := `SELECT * FROM trouble_reports ORDER BY id DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		logger.DBTroubleReports().Error("Failed to execute trouble reports list query: %v", err)
		return nil, utils.NewDatabaseError("select", "trouble_reports", err)
	}
	defer rows.Close()

	var reports []*models.TroubleReport
	reportCount := 0

	for rows.Next() {
		report, err := s.scanTroubleReport(rows)
		if err != nil {
			return nil, utils.NewDatabaseError("scan", "trouble_reports", err)
		}
		reports = append(reports, report)
		reportCount++
	}

	if err := rows.Err(); err != nil {
		elapsed := time.Since(start)
		logger.DBTroubleReports().Error("Error iterating over %d trouble report rows in %v: %v", reportCount, elapsed, err)
		return nil, utils.NewDatabaseError("select", "trouble_reports", err)
	}

	elapsed := time.Since(start)
	logger.DBTroubleReports().Info("Listed %d trouble reports in %v", len(reports), elapsed)
	if elapsed > 100*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble reports list query took %v for %d reports", elapsed, len(reports))
	}

	return reports, nil
}

// Get retrieves a specific trouble report by ID.
func (s *TroubleReport) Get(id int64) (*models.TroubleReport, error) {
	logger.DBTroubleReports().Debug("Getting trouble report by ID: %d", id)
	start := time.Now()

	query := `SELECT * FROM trouble_reports WHERE id = ?`
	row := s.db.QueryRow(query, id)

	report, err := s.scanTroubleReport(row)
	elapsed := time.Since(start)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("trouble report with ID %d", id))
		}

		return nil, utils.NewDatabaseError("select", "trouble_reports", err)
	}

	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s') in %v", id, report.Title, elapsed)
	return report, nil
}

// Add creates a new trouble report and generates a corresponding activity feed entry.
func (s *TroubleReport) Add(troubleReport *models.TroubleReport, user *models.User) (int64, error) {
	if troubleReport == nil {
		logger.DBTroubleReports().Error("Attempted to add nil trouble report")
		return 0, utils.NewValidationError("report: trouble report cannot be nil")
	}

	userInfo := "unknown user"
	if user != nil {
		userInfo = fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
	}

	logger.DBTroubleReports().Info("Adding trouble report by %s: title='%s', attachments=%d",
		userInfo, troubleReport.Title, len(troubleReport.LinkedAttachments))
	start := time.Now()

	if err := troubleReport.Validate(); err != nil {
		logger.DBTroubleReports().Warn("Trouble report validation failed for %s: %v", userInfo, err)
		return 0, err
	}

	s.updateMods(user, troubleReport)

	marshalStart := time.Now()
	linkedAttachments, err := json.Marshal(troubleReport.LinkedAttachments)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal linked attachments: %w", err)
	}

	mods, err := json.Marshal(troubleReport.Mods)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal mods data: %w", err)
	}
	marshalElapsed := time.Since(marshalStart)

	dbStart := time.Now()
	query := `INSERT INTO trouble_reports
		(title, content, linked_attachments, mods) VALUES (?, ?, ?, ?)`
	result, err := s.db.Exec(
		query,
		troubleReport.Title, troubleReport.Content, string(linkedAttachments), mods,
	)
	if err != nil {
		dbElapsed := time.Since(dbStart)
		logger.DBTroubleReports().Error("Failed to insert trouble report for %s in %v: %v", userInfo, dbElapsed, err)
		return 0, utils.NewDatabaseError("insert", "trouble_reports", err)
	}
	dbElapsed := time.Since(dbStart)

	id, err := result.LastInsertId()
	if err != nil {
		logger.DBTroubleReports().Error("Failed to get last insert ID for trouble report by %s: %v", userInfo, err)
		return 0, utils.NewDatabaseError("insert", "trouble_reports", err)
	}
	troubleReport.ID = id

	feedStart := time.Now()
	feed := models.NewFeed(
		"Neuer Problembericht",
		fmt.Sprintf("Benutzer %s hat einen neuen Problembericht '%s' hinzugefügt.",
			troubleReport.Mods.Current().User.Name, troubleReport.Title),
		troubleReport.Mods.Current().User.TelegramID,
	)
	if err := s.feeds.Add(feed); err != nil {
		feedElapsed := time.Since(feedStart)
		logger.DBTroubleReports().Error("Failed to create feed entry for trouble report %d in %v: %v", id, feedElapsed, err)
		return id, fmt.Errorf("failed to add feed entry: %w", err)
	}
	feedElapsed := time.Since(feedStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Successfully added trouble report %d by %s in %v (marshal: %v, db: %v, feed: %v)",
		id, userInfo, totalElapsed, marshalElapsed, dbElapsed, feedElapsed)

	if totalElapsed > 200*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble report insertion took %v for %s", totalElapsed, userInfo)
	}

	return id, nil
}

// Update modifies an existing trouble report and generates an activity feed entry.
func (s *TroubleReport) Update(troubleReport *models.TroubleReport, user *models.User) error {
	if troubleReport == nil {
		logger.DBTroubleReports().Error("Attempted to update nil trouble report")
		return utils.NewValidationError("report: trouble report cannot be nil")
	}

	id := troubleReport.ID
	userInfo := "unknown user"
	if user != nil {
		userInfo = fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
	}

	logger.DBTroubleReports().Info("Updating trouble report %d by %s: title='%s', attachments=%d",
		id, userInfo, troubleReport.Title, len(troubleReport.LinkedAttachments))
	start := time.Now()

	if err := troubleReport.Validate(); err != nil {
		logger.DBTroubleReports().Warn("Trouble report validation failed for update %d by %s: %v", id, userInfo, err)
		return err
	}

	s.updateMods(user, troubleReport)

	marshalStart := time.Now()
	linkedAttachments, err := json.Marshal(troubleReport.LinkedAttachments)
	if err != nil {
		logger.DBTroubleReports().Error("Failed to marshal linked attachments for update %d by %s: %v", id, userInfo, err)
		return fmt.Errorf("failed to marshal linked attachments: %w", err)
	}

	mods, err := json.Marshal(troubleReport.Mods)
	if err != nil {
		logger.DBTroubleReports().Error("Failed to marshal mods data for update %d by %s: %v", id, userInfo, err)
		return fmt.Errorf("failed to marshal mods data: %w", err)
	}
	marshalElapsed := time.Since(marshalStart)

	dbStart := time.Now()
	query := `UPDATE trouble_reports
		SET title = ?, content = ?, linked_attachments = ?, mods = ? WHERE id = ?`
	_, err = s.db.Exec(
		query,
		troubleReport.Title, troubleReport.Content, string(linkedAttachments), mods, id,
	)
	if err != nil {
		dbElapsed := time.Since(dbStart)
		logger.DBTroubleReports().Error("Failed to update trouble report %d by %s in %v: %v", id, userInfo, dbElapsed, err)
		return utils.NewDatabaseError("update", "trouble_reports", err)
	}
	dbElapsed := time.Since(dbStart)

	feedStart := time.Now()
	feed := models.NewFeed(
		"Problembericht aktualisiert",
		fmt.Sprintf("Benutzer %s hat den Problembericht '%s' aktualisiert.",
			troubleReport.Mods.Current().User.Name, troubleReport.Title),
		troubleReport.Mods.Current().User.TelegramID,
	)
	if err := s.feeds.Add(feed); err != nil {
		feedElapsed := time.Since(feedStart)
		logger.DBTroubleReports().Error("Failed to create feed entry for trouble report update %d in %v: %v", id, feedElapsed, err)
		return err
	}
	feedElapsed := time.Since(feedStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Successfully updated trouble report %d by %s in %v (marshal: %v, db: %v, feed: %v)",
		id, userInfo, totalElapsed, marshalElapsed, dbElapsed, feedElapsed)

	if totalElapsed > 200*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble report update took %v for %d by %s", totalElapsed, id, userInfo)
	}

	return nil
}

// Delete deletes a trouble report by ID and generates an activity feed entry.
func (s *TroubleReport) Delete(id int64, user *models.User) error {
	userInfo := "unknown user"
	if user != nil {
		userInfo = fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
	}

	logger.DBTroubleReports().Info("Removing trouble report %d by %s", id, userInfo)
	start := time.Now()

	// Get the report before deleting for the feed entry and logging
	getStart := time.Now()
	report, err := s.Get(id)
	if err != nil {
		getElapsed := time.Since(getStart)
		logger.DBTroubleReports().Error("Failed to get trouble report %d before deletion by %s in %v: %v", id, userInfo, getElapsed, err)
		return err
	}
	getElapsed := time.Since(getStart)
	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s') for deletion in %v", id, report.Title, getElapsed)

	dbStart := time.Now()
	query := `DELETE FROM trouble_reports WHERE id = ?`
	result, err := s.db.Exec(query, id)
	if err != nil {
		dbElapsed := time.Since(dbStart)
		logger.DBTroubleReports().Error("Failed to delete trouble report %d by %s in %v: %v", id, userInfo, dbElapsed, err)
		return utils.NewDatabaseError("delete", "trouble_reports", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		dbElapsed := time.Since(dbStart)
		logger.DBTroubleReports().Error("Failed to get rows affected for deletion %d in %v: %v", id, dbElapsed, err)
		return utils.NewDatabaseError("delete", "trouble_reports", err)
	}
	dbElapsed := time.Since(dbStart)

	if rowsAffected == 0 {
		totalElapsed := time.Since(start)
		logger.DBTroubleReports().Debug("No trouble report found to delete with ID %d (total time: %v)", id, totalElapsed)
		return utils.NewNotFoundError(fmt.Sprintf("no trouble report found with ID %d", id))
	}

	// Create feed entry for the removed report
	feedStart := time.Now()
	if user != nil {
		feed := models.NewFeed(
			"Problembericht entfernt",
			fmt.Sprintf("Benutzer %s hat den Problembericht '%s' entfernt.",
				user.Name, report.Title),
			user.TelegramID,
		)
		if err := s.feeds.Add(feed); err != nil {
			feedElapsed := time.Since(feedStart)
			logger.DBTroubleReports().Error("Failed to create feed entry for deletion %d in %v: %v", id, feedElapsed, err)
			return err
		}
	}
	feedElapsed := time.Since(feedStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Successfully deleted trouble report %d (title='%s') by %s in %v (get: %v, db: %v, feed: %v)",
		id, report.Title, userInfo, totalElapsed, getElapsed, dbElapsed, feedElapsed)

	return nil
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *TroubleReport) GetWithAttachments(
	id int64,
) (*models.TroubleReportWithAttachments, error) {
	logger.DBTroubleReports().Debug("Getting trouble report with attachments, id: %d", id)
	start := time.Now()

	// Get the trouble report
	trStart := time.Now()
	tr, err := s.Get(id)
	if err != nil {
		trElapsed := time.Since(trStart)
		logger.DBTroubleReports().Error("Failed to get trouble report %d with attachments in %v: %v", id, trElapsed, err)
		return nil, err
	}
	trElapsed := time.Since(trStart)

	// Load attachments
	attachStart := time.Now()
	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		attachElapsed := time.Since(attachStart)
		logger.DBTroubleReports().Error("Failed to load %d attachments for trouble report %d in %v: %v", len(tr.LinkedAttachments), id, attachElapsed, err)
		return nil, err
	}
	attachElapsed := time.Since(attachStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s') with %d attachments in %v (report: %v, attachments: %v)",
		id, tr.Title, len(attachments), totalElapsed, trElapsed, attachElapsed)

	if totalElapsed > 100*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble report with attachments query took %v for %d", totalElapsed, id)
	}

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (s *TroubleReport) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	logger.DBTroubleReports().Debug("Starting trouble reports with attachments list query")
	start := time.Now()

	// Get all trouble reports
	listStart := time.Now()
	reports, err := s.List()
	if err != nil {
		listElapsed := time.Since(listStart)
		logger.DBTroubleReports().Error("Failed to list trouble reports in %v: %v", listElapsed, err)
		return nil, err
	}
	listElapsed := time.Since(listStart)

	var result []*models.TroubleReportWithAttachments
	totalAttachments := 0

	attachStart := time.Now()
	for i, tr := range reports {
		// Load attachments for each report
		attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
		if err != nil {
			logger.DBTroubleReports().Error("Failed to load %d attachments for trouble report %d (report %d/%d): %v",
				len(tr.LinkedAttachments), tr.ID, i+1, len(reports), err)
			return nil, err
		}

		totalAttachments += len(attachments)
		result = append(result, &models.TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}
	attachElapsed := time.Since(attachStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Listed %d trouble reports with %d total attachments in %v (list: %v, attachments: %v)",
		len(result), totalAttachments, totalElapsed, listElapsed, attachElapsed)

	if totalElapsed > 200*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble reports with attachments query took %v for %d reports", totalElapsed, len(result))
	}

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (s *TroubleReport) AddWithAttachments(
	user *models.User,
	troubleReport *models.TroubleReport,
	attachments []*models.Attachment,
) error {
	userInfo := "unknown user"
	if user != nil {
		userInfo = fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
	}

	logger.DBTroubleReports().Info("Adding trouble report with %d attachments by %s: title='%s'",
		len(attachments), userInfo, troubleReport.Title)
	start := time.Now()

	if troubleReport == nil {
		return utils.NewValidationError("report: trouble report cannot be nil")
	}

	// First, add the attachments and collect their IDs
	attachStart := time.Now()
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
	attachElapsed := time.Since(attachStart)

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Add the trouble report
	reportStart := time.Now()
	if _, err := s.Add(troubleReport, user); err != nil {
		reportElapsed := time.Since(reportStart)
		logger.DBTroubleReports().Error("Failed to add trouble report for %s in %v, cleaning up %d attachments: %v",
			userInfo, reportElapsed, len(attachmentIDs), err)
		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			if cleanupErr := s.attachments.Delete(id, user); cleanupErr != nil {
				logger.DBTroubleReports().Error("Failed to cleanup attachment %d: %v", id, cleanupErr)
			}
		}
		return err
	}
	reportElapsed := time.Since(reportStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Successfully added trouble report %d with %d attachments by %s in %v (attachments: %v, report: %v)",
		troubleReport.ID, len(attachmentIDs), userInfo, totalElapsed, attachElapsed, reportElapsed)

	if totalElapsed > 500*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble report with attachments creation took %v for %s", totalElapsed, userInfo)
	}

	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (s *TroubleReport) UpdateWithAttachments(
	user *models.User,
	id int64,
	tr *models.TroubleReport,
	newAttachments []*models.Attachment,
) error {
	userInfo := "unknown user"
	if user != nil {
		userInfo = fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
	}

	logger.DBTroubleReports().Info("Updating trouble report %d with %d new attachments by %s: title='%s'",
		id, len(newAttachments), userInfo, tr.Title)
	start := time.Now()

	if tr == nil {
		logger.DBTroubleReports().Error("Attempted to update with nil trouble report by %s", userInfo)
		return utils.NewValidationError("report: trouble report cannot be nil")
	}

	// Add new attachments
	attachStart := time.Now()
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
	attachElapsed := time.Since(attachStart)

	// Combine existing and new attachment IDs
	originalAttachmentCount := len(tr.LinkedAttachments)
	allAttachmentIDs := append(tr.LinkedAttachments, newAttachmentIDs...)
	tr.LinkedAttachments = allAttachmentIDs
	tr.ID = id

	logger.DBTroubleReports().Debug("Combined attachments for update %d: %d existing + %d new = %d total",
		id, originalAttachmentCount, len(newAttachmentIDs), len(allAttachmentIDs))

	// Update the trouble report
	updateStart := time.Now()
	if err := s.Update(tr, user); err != nil {
		updateElapsed := time.Since(updateStart)
		logger.DBTroubleReports().Error("Failed to update trouble report %d by %s in %v, cleaning up %d new attachments: %v",
			id, userInfo, updateElapsed, len(newAttachmentIDs), err)

		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			if cleanupErr := s.attachments.Delete(attachmentID, user); cleanupErr != nil {
				logger.DBTroubleReports().Error("Failed to cleanup new attachment %d: %v", attachmentID, cleanupErr)
			}
		}

		return err
	}
	updateElapsed := time.Since(updateStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Successfully updated trouble report %d with %d new attachments by %s in %v (attachments: %v, update: %v)",
		id, len(newAttachmentIDs), userInfo, totalElapsed, attachElapsed, updateElapsed)

	if totalElapsed > 500*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble report with attachments update took %v for %d by %s", totalElapsed, id, userInfo)
	}

	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (s *TroubleReport) RemoveWithAttachments(id int64, user *models.User) (*models.TroubleReport, error) {
	userInfo := "unknown user"
	if user != nil {
		userInfo = fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
	}

	logger.DBTroubleReports().Info("Removing trouble report %d with attachments by %s", id, userInfo)
	start := time.Now()

	// Get the trouble report to find its attachments
	getStart := time.Now()
	tr, err := s.Get(id)
	if err != nil {
		getElapsed := time.Since(getStart)
		logger.DBTroubleReports().Error("Failed to get trouble report %d for removal by %s in %v: %v", id, userInfo, getElapsed, err)
		return tr, err
	}
	getElapsed := time.Since(getStart)

	logger.DBTroubleReports().Debug("Retrieved trouble report %d (title='%s') with %d attachments for removal by %s in %v",
		id, tr.Title, len(tr.LinkedAttachments), userInfo, getElapsed)

	// Remove the trouble report first
	deleteStart := time.Now()
	if err := s.Delete(id, user); err != nil {
		deleteElapsed := time.Since(deleteStart)
		logger.DBTroubleReports().Error("Failed to remove trouble report %d by %s in %v: %v", id, userInfo, deleteElapsed, err)
		return tr, err
	}
	deleteElapsed := time.Since(deleteStart)

	// Remove associated attachments
	attachStart := time.Now()
	successfulAttachmentDeletes := 0
	failedAttachmentDeletes := 0
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Delete(attachmentID, user); err != nil {
			logger.DBTroubleReports().Warn("Failed to remove attachment %d for trouble report %d by %s: %v", attachmentID, id, userInfo, err)
			failedAttachmentDeletes++
		} else {
			successfulAttachmentDeletes++
		}
	}
	attachElapsed := time.Since(attachStart)

	totalElapsed := time.Since(start)
	logger.DBTroubleReports().Info("Successfully removed trouble report %d (title='%s') with %d/%d attachments by %s in %v (get: %v, delete: %v, attachments: %v)",
		id, tr.Title, successfulAttachmentDeletes, len(tr.LinkedAttachments), userInfo, totalElapsed, getElapsed, deleteElapsed, attachElapsed)

	if failedAttachmentDeletes > 0 {
		logger.DBTroubleReports().Warn("Failed to remove %d/%d attachments for trouble report %d", failedAttachmentDeletes, len(tr.LinkedAttachments), id)
	}

	if totalElapsed > 200*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow trouble report removal with attachments took %v for %d by %s", totalElapsed, id, userInfo)
	}

	return tr, nil
}

// LoadAttachments loads attachments for a trouble report.
func (s *TroubleReport) LoadAttachments(tr *models.TroubleReport) ([]*models.Attachment, error) {
	if tr == nil {
		logger.DBTroubleReports().Error("Attempted to load attachments for nil trouble report")
		return nil, utils.NewValidationError("report: trouble report cannot be nil")
	}

	logger.DBTroubleReports().Debug("Loading %d attachments for trouble report %d (title='%s')",
		len(tr.LinkedAttachments), tr.ID, tr.Title)
	start := time.Now()

	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	elapsed := time.Since(start)

	if err != nil {
		logger.DBTroubleReports().Error("Failed to load %d attachments for trouble report %d in %v: %v",
			len(tr.LinkedAttachments), tr.ID, elapsed, err)
		return nil, err
	}

	logger.DBTroubleReports().Debug("Successfully loaded %d/%d attachments for trouble report %d in %v",
		len(attachments), len(tr.LinkedAttachments), tr.ID, elapsed)

	if elapsed > 100*time.Millisecond {
		logger.DBTroubleReports().Warn("Slow attachment loading took %v for %d attachments (report %d)",
			elapsed, len(tr.LinkedAttachments), tr.ID)
	}

	return attachments, nil
}

func (s *TroubleReport) scanTroubleReport(scanner interfaces.Scannable) (*models.TroubleReport, error) {
	report := &models.TroubleReport{}
	var linkedAttachments string
	var mods []byte

	if err := scanner.Scan(&report.ID, &report.Title, &report.Content,
		&linkedAttachments, &mods); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		logger.DBTroubleReports().Error("Failed to scan trouble report row: %v", err)
		return nil, fmt.Errorf("failed to scan row %v", err)
	}

	// Try to unmarshal as new format (array of int64 IDs) first
	if err := json.Unmarshal([]byte(linkedAttachments), &report.LinkedAttachments); err != nil {
		logger.DBTroubleReports().Error("Failed to unmarshal linked attachments for report %d: %v", report.ID, err)
		return nil, fmt.Errorf("failed to unmarshal linked attachments %v", err)
	}

	if err := json.Unmarshal(mods, &report.Mods); err != nil {
		logger.DBTroubleReports().Error("Failed to unmarshal mods data for report %d: %v", report.ID, err)
		return nil, fmt.Errorf("failed to unmarshal mods data %v", err)
	}

	return report, nil
}

func (s *TroubleReport) updateMods(user *models.User, report *models.TroubleReport) {
	if user == nil {
		return
	}

	report.Mods.Add(user, models.TroubleReportMod{
		Title:             report.Title,
		Content:           report.Content,
		LinkedAttachments: report.LinkedAttachments,
	})
}
