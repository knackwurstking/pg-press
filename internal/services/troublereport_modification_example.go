package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// TroubleReportWithModificationService demonstrates how to integrate the new modification service
// This is an example of how you could refactor the existing TroubleReport service
type TroubleReportWithModificationService struct {
	db            *sql.DB
	modifications *ModificationService
	attachments   *Attachment
	feeds         *Feed
}

// NewTroubleReportWithModificationService creates a new service instance with modification support
func NewTroubleReportWithModificationService(db *sql.DB, modifications *ModificationService, attachments *Attachment, feeds *Feed) *TroubleReportWithModificationService {
	return &TroubleReportWithModificationService{
		db:            db,
		modifications: modifications,
		attachments:   attachments,
		feeds:         feeds,
	}
}

// Add creates a new trouble report and records the initial modification
func (s *TroubleReportWithModificationService) Add(report *models.TroubleReport, user *models.User) error {
	logger.DBModifications().Info("Adding trouble report with modification tracking")

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w: %w", err)
	}
	defer tx.Rollback()

	// Insert the trouble report
	query := `
		INSERT INTO trouble_reports (title, content, linked_attachments, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	attachmentsJSON, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %w: %w", err)
	}

	result, err := tx.Exec(query, report.Title, report.Content, attachmentsJSON)
	if err != nil {
		return fmt.Errorf("failed to insert trouble report: %w: %w", err)
	}

	reportID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get report ID: %w: %w", err)
	}
	report.ID = reportID

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w: %w", err)
	}

	// Record the initial modification
	modData := models.NewExtendedModificationData(
		models.TroubleReportModData{
			Title:             report.Title,
			Content:           report.Content,
			LinkedAttachments: report.LinkedAttachments,
		},
		models.ActionCreate,
		"Initial trouble report creation",
	)

	if err = s.modifications.AddTroubleReportMod(user.TelegramID, reportID, modData); err != nil {
		logger.DBModifications().Error("Failed to record initial modification for trouble report %d: %v", reportID, err)
		// Don't fail the operation, just log the error
	}

	logger.DBModifications().Info("Successfully added trouble report: id=%d", reportID)
	return nil
}

// Update modifies an existing trouble report and records the modification
func (s *TroubleReportWithModificationService) Update(report *models.TroubleReport, user *models.User) error {
	logger.DBModifications().Info("Updating trouble report with modification tracking: id=%d", report.ID)

	// Get current report for comparison
	current, err := s.Get(report.ID)
	if err != nil {
		return fmt.Errorf("failed to get current report: %w: %w", err)
	}

	// Check if there are actual changes
	hasChanges := current.Title != report.Title ||
		current.Content != report.Content ||
		!equalInt64Slices(current.LinkedAttachments, report.LinkedAttachments)

	if !hasChanges {
		logger.DBModifications().Debug("No changes detected for trouble report: id=%d", report.ID)
		return nil
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w: %w", err)
	}
	defer tx.Rollback()

	// Update the trouble report
	query := `
		UPDATE trouble_reports
		SET title = ?, content = ?, linked_attachments = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	attachmentsJSON, err := json.Marshal(report.LinkedAttachments)
	if err != nil {
		return fmt.Errorf("failed to marshal linked attachments: %w: %w", err)
	}

	result, err := tx.Exec(query, report.Title, report.Content, attachmentsJSON, report.ID)
	if err != nil {
		return fmt.Errorf("failed to update trouble report: %w: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trouble report not found: id=%d", report.ID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w: %w", err)
	}

	// Record the modification with details about what changed
	description := s.buildChangeDescription(current, report)
	modData := models.NewExtendedModificationData(
		models.TroubleReportModData{
			Title:             report.Title,
			Content:           report.Content,
			LinkedAttachments: report.LinkedAttachments,
		},
		models.ActionUpdate,
		description,
	)

	if err = s.modifications.AddTroubleReportMod(user.TelegramID, report.ID, modData); err != nil {
		logger.DBModifications().Error("Failed to record modification for trouble report %d: %v", report.ID, err)
		// Don't fail the operation, just log the error
	}

	logger.DBModifications().Info("Successfully updated trouble report: id=%d", report.ID)
	return nil
}

// Delete removes a trouble report and records the deletion
func (s *TroubleReportWithModificationService) Delete(reportID int64, user *models.User) error {
	logger.DBModifications().Info("Deleting trouble report with modification tracking: id=%d", reportID)

	// Get the report before deletion for the modification record
	report, err := s.Get(reportID)
	if err != nil {
		return fmt.Errorf("failed to get report before deletion: %w: %w", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w: %w", err)
	}
	defer tx.Rollback()

	// Delete the trouble report
	query := `DELETE FROM trouble_reports WHERE id = ?`
	result, err := tx.Exec(query, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete trouble report: %w: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trouble report not found: id=%d", reportID)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w: %w", err)
	}

	// Record the deletion modification
	modData := models.NewExtendedModificationData(
		models.TroubleReportModData{
			Title:             report.Title,
			Content:           report.Content,
			LinkedAttachments: report.LinkedAttachments,
		},
		models.ActionDelete,
		fmt.Sprintf("Trouble report '%s' was deleted", report.Title),
	)

	if err = s.modifications.AddTroubleReportMod(user.TelegramID, reportID, modData); err != nil {
		logger.DBModifications().Error("Failed to record deletion modification for trouble report %d: %v", reportID, err)
		// Don't fail the operation, just log the error
	}

	logger.DBModifications().Info("Successfully deleted trouble report: id=%d", reportID)
	return nil
}

// GetModificationHistory retrieves the modification history for a trouble report
func (s *TroubleReportWithModificationService) GetModificationHistory(reportID int64, limit, offset int) ([]*ModificationWithUser, error) {
	logger.DBModifications().Debug("Getting modification history for trouble report: id=%d", reportID)

	modifications, err := s.modifications.ListWithUser(ModificationTypeTroubleReport, reportID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get modification history: %w: %w", err)
	}

	return modifications, nil
}

// GetModificationCount returns the total number of modifications for a trouble report
func (s *TroubleReportWithModificationService) GetModificationCount(reportID int64) (int64, error) {
	count, err := s.modifications.Count(ModificationTypeTroubleReport, reportID)
	if err != nil {
		return 0, fmt.Errorf("failed to get modification count: %w: %w", err)
	}
	return count, nil
}

// GetLatestModification returns the most recent modification for a trouble report
func (s *TroubleReportWithModificationService) GetLatestModification(reportID int64) (*models.Modification[any], error) {
	mod, err := s.modifications.GetLatest(ModificationTypeTroubleReport, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest modification: %w: %w", err)
	}
	return mod, nil
}

// RestoreFromModification restores a trouble report to a previous state
func (s *TroubleReportWithModificationService) RestoreFromModification(reportID, modificationID int64, user *models.User) error {
	logger.DBModifications().Info("Restoring trouble report from modification: report_id=%d, mod_id=%d", reportID, modificationID)

	// Get the modification to restore from
	mod, err := s.modifications.Get(modificationID)
	if err != nil {
		return fmt.Errorf("failed to get modification: %w: %w", err)
	}

	// Unmarshal the modification data
	var extendedData models.ExtendedModificationData[models.TroubleReportModData]
	var dataBytes []byte = mod.Data
	if err = json.Unmarshal(dataBytes, &extendedData); err != nil {
		return fmt.Errorf("failed to unmarshal modification data: %w: %w", err)
	}

	// Create a trouble report with the restored data
	restoredReport := &models.TroubleReport{
		ID:                reportID,
		Title:             extendedData.Data.Title,
		Content:           extendedData.Data.Content,
		LinkedAttachments: extendedData.Data.LinkedAttachments,
	}

	// Update the report
	if err = s.Update(restoredReport, user); err != nil {
		return fmt.Errorf("failed to restore report: %w: %w", err)
	}

	// Record the restoration as a modification
	restoreModData := models.NewExtendedModificationData(
		models.TroubleReportModData{
			Title:             restoredReport.Title,
			Content:           restoredReport.Content,
			LinkedAttachments: restoredReport.LinkedAttachments,
		},
		models.ActionUpdate,
		fmt.Sprintf("Restored from modification %d", modificationID),
	)

	if err = s.modifications.AddTroubleReportMod(user.TelegramID, reportID, restoreModData); err != nil {
		logger.DBModifications().Error("Failed to record restoration modification: %v: %w", err)
	}

	logger.DBModifications().Info("Successfully restored trouble report from modification")
	return nil
}

// Helper methods

// Get retrieves a trouble report by ID (placeholder - implement based on your existing logic)
func (s *TroubleReportWithModificationService) Get(reportID int64) (*models.TroubleReport, error) {
	// This would contain your existing Get logic
	// Returning a placeholder for compilation
	return &models.TroubleReport{}, nil
}

// buildChangeDescription creates a human-readable description of changes
func (s *TroubleReportWithModificationService) buildChangeDescription(old, new *models.TroubleReport) string {
	var changes []string

	if old.Title != new.Title {
		changes = append(changes, fmt.Sprintf("Title changed from '%s' to '%s'", old.Title, new.Title))
	}

	if old.Content != new.Content {
		changes = append(changes, "Content was modified")
	}

	if !equalInt64Slices(old.LinkedAttachments, new.LinkedAttachments) {
		changes = append(changes, "Linked attachments were modified")
	}

	if len(changes) == 0 {
		return "No significant changes detected"
	}

	if len(changes) == 1 {
		return changes[0]
	}

	return fmt.Sprintf("%d changes: %s", len(changes), changes[0])
}

// equalInt64Slices compares two int64 slices for equality
func equalInt64Slices(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
