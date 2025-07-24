// Package pgvis provides service layer for trouble report operations with attachment lazy loading.
package pgvis

import (
	"fmt"

	"github.com/knackwurstking/pg-vis/pgvis/logger"
)

// TroubleReportWithAttachments represents a trouble report with its attachments loaded.
type TroubleReportWithAttachments struct {
	*TroubleReport
	LoadedAttachments []*Attachment `json:"loaded_attachments"`
}

// TroubleReportService provides high-level operations for trouble reports with attachment management.
type TroubleReportService struct {
	troubleReports *TroubleReports
	attachments    *Attachments
}

// NewTroubleReportService creates a new service instance.
func NewTroubleReportService(troubleReports *TroubleReports, attachments *Attachments) *TroubleReportService {
	return &TroubleReportService{
		troubleReports: troubleReports,
		attachments:    attachments,
	}
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *TroubleReportService) GetWithAttachments(id int64) (*TroubleReportWithAttachments, error) {
	logger.TroubleReport().Debug("Getting trouble report with attachments, id: %d", id)

	// Get the trouble report
	tr, err := s.troubleReports.Get(id)
	if err != nil {
		return nil, err
	}

	// Load attachments
	attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		return nil, WrapError(err, "failed to load attachments for trouble report")
	}

	return &TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (s *TroubleReportService) ListWithAttachments() ([]*TroubleReportWithAttachments, error) {
	logger.TroubleReport().Debug("Listing trouble reports with attachments")

	// Get all trouble reports
	reports, err := s.troubleReports.List()
	if err != nil {
		return nil, err
	}

	var result []*TroubleReportWithAttachments

	for _, tr := range reports {
		// Load attachments for each report
		attachments, err := s.attachments.GetByIDs(tr.LinkedAttachments)
		if err != nil {
			return nil, WrapError(err, fmt.Sprintf("failed to load attachments for trouble report %d", tr.ID))
		}

		result = append(result, &TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (s *TroubleReportService) AddWithAttachments(troubleReport *TroubleReport, attachments []*Attachment) error {
	logger.TroubleReport().Info("Adding trouble report with %d attachments", len(attachments))

	if troubleReport == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, err := s.attachments.Add(attachment)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range attachmentIDs {
				s.attachments.Remove(addedID)
			}
			return WrapError(err, "failed to add attachment")
		}
		attachmentIDs = append(attachmentIDs, id)
	}

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Update the mods to include the attachment IDs
	if len(troubleReport.Mods) > 0 {
		currentMod := troubleReport.Mods[len(troubleReport.Mods)-1]
		currentMod.Data.LinkedAttachments = attachmentIDs
	}

	// Add the trouble report
	if err := s.troubleReports.Add(troubleReport); err != nil {
		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			s.attachments.Remove(id)
		}
		return WrapError(err, "failed to add trouble report")
	}

	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (s *TroubleReportService) UpdateWithAttachments(id int64, troubleReport *TroubleReport, newAttachments []*Attachment) error {
	logger.TroubleReport().Info("Updating trouble report %d with %d new attachments", id, len(newAttachments))

	if troubleReport == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// Get existing trouble report to find old attachments
	oldTR, err := s.troubleReports.Get(id)
	if err != nil {
		return WrapError(err, "failed to get existing trouble report")
	}

	// Add new attachments
	var newAttachmentIDs []int64
	for _, attachment := range newAttachments {
		if attachment == nil {
			continue
		}

		attachmentID, err := s.attachments.Add(attachment)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range newAttachmentIDs {
				s.attachments.Remove(addedID)
			}
			return WrapError(err, "failed to add new attachment")
		}
		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	// Combine existing and new attachment IDs
	allAttachmentIDs := append(troubleReport.LinkedAttachments, newAttachmentIDs...)
	troubleReport.LinkedAttachments = allAttachmentIDs

	// Update the mods to include the attachment IDs
	if len(troubleReport.Mods) > 0 {
		currentMod := troubleReport.Mods[len(troubleReport.Mods)-1]
		currentMod.Data.LinkedAttachments = allAttachmentIDs
	}

	// Update the trouble report
	if err := s.troubleReports.Update(id, troubleReport); err != nil {
		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			s.attachments.Remove(attachmentID)
		}
		return WrapError(err, "failed to update trouble report")
	}

	// Clean up orphaned attachments from the old report
	// (attachments that were removed from the report)
	orphanedIDs := findRemovedAttachments(oldTR.LinkedAttachments, allAttachmentIDs)
	for _, orphanedID := range orphanedIDs {
		if err := s.attachments.Remove(orphanedID); err != nil {
			logger.TroubleReport().Warn("Failed to remove orphaned attachment %d: %v", orphanedID, err)
		}
	}

	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (s *TroubleReportService) RemoveWithAttachments(id int64) error {
	logger.TroubleReport().Info("Removing trouble report %d with attachments", id)

	// Get the trouble report to find its attachments
	tr, err := s.troubleReports.Get(id)
	if err != nil {
		return WrapError(err, "failed to get trouble report for removal")
	}

	// Remove the trouble report first
	if err := s.troubleReports.Remove(id); err != nil {
		return WrapError(err, "failed to remove trouble report")
	}

	// Remove associated attachments
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Remove(attachmentID); err != nil {
			logger.TroubleReport().Warn("Failed to remove attachment %d: %v", attachmentID, err)
		}
	}

	return nil
}

// LoadAttachments loads attachments for a trouble report.
func (s *TroubleReportService) LoadAttachments(tr *TroubleReport) ([]*Attachment, error) {
	if tr == nil {
		return nil, NewValidationError("report", "trouble report cannot be nil", nil)
	}

	return s.attachments.GetByIDs(tr.LinkedAttachments)
}

// GetAttachment retrieves a specific attachment by ID.
func (s *TroubleReportService) GetAttachment(id int64) (*Attachment, error) {
	return s.attachments.Get(id)
}

// CleanupOrphanedAttachments removes attachments not referenced by any trouble report.
func (s *TroubleReportService) CleanupOrphanedAttachments() (int64, error) {
	return s.attachments.CleanupOrphaned()
}

// Helper function to find attachments that were removed
func findRemovedAttachments(oldIDs, newIDs []int64) []int64 {
	newIDMap := make(map[int64]bool)
	for _, id := range newIDs {
		newIDMap[id] = true
	}

	var removed []int64
	for _, id := range oldIDs {
		if !newIDMap[id] {
			removed = append(removed, id)
		}
	}

	return removed
}

// ConvertToLegacyFormat converts attachment IDs to legacy attachment objects for backward compatibility.
func (s *TroubleReportService) ConvertToLegacyFormat(tr *TroubleReport) (*struct {
	*TroubleReport
	LegacyAttachments []*Attachment `json:"linked_attachments"`
}, error) {
	attachments, err := s.LoadAttachments(tr)
	if err != nil {
		return nil, WrapError(err, "failed to load attachments for legacy format")
	}

	return &struct {
		*TroubleReport
		LegacyAttachments []*Attachment `json:"linked_attachments"`
	}{
		TroubleReport:     tr,
		LegacyAttachments: attachments,
	}, nil
}
