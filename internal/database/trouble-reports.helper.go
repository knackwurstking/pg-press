package database

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/logger"
)

// TroubleReportWithAttachments represents a trouble report with its attachments loaded.
type TroubleReportWithAttachments struct {
	*TroubleReport
	LoadedAttachments []*Attachment `json:"loaded_attachments"`
}

// TroubleReportsHelper provides high-level operations for trouble reports
// with attachment management.
type TroubleReportsHelper struct {
	troubleReports DataOperations[*TroubleReport]
	attachments    *Attachments
}

// NewTroubleReportsHelper creates a new helper instance.
func NewTroubleReportsHelper(
	troubleReports DataOperations[*TroubleReport],
	attachments *Attachments,
) *TroubleReportsHelper {
	return &TroubleReportsHelper{
		troubleReports: troubleReports,
		attachments:    attachments,
	}
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (s *TroubleReportsHelper) GetWithAttachments(
	id int64,
) (*TroubleReportWithAttachments, error) {
	logger.DBTroubleReportsHelper().Debug(
		"Getting trouble report with attachments, id: %d", id)

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
func (s *TroubleReportsHelper) ListWithAttachments() ([]*TroubleReportWithAttachments, error) {
	logger.DBTroubleReportsHelper().Debug("Listing trouble reports with attachments")

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
			return nil, WrapError(err,
				fmt.Sprintf("failed to load attachments for trouble report %d", tr.ID))
		}

		result = append(result, &TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (s *TroubleReportsHelper) AddWithAttachments(
	user *User,
	troubleReport *TroubleReport,
	attachments []*Attachment,
) error {
	logger.DBTroubleReportsHelper().Info("Adding trouble report with %d attachments", len(attachments))

	if troubleReport == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, err := s.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range attachmentIDs {
				s.attachments.Delete(addedID, user)
			}
			return WrapError(err, "failed to add attachment")
		}
		attachmentIDs = append(attachmentIDs, id)
	}

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Add the trouble report
	if _, err := s.troubleReports.Add(troubleReport, user); err != nil {
		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			s.attachments.Delete(id, user)
		}
		return WrapError(err, "failed to add trouble report")
	}

	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (s *TroubleReportsHelper) UpdateWithAttachments(
	user *User,
	id int64,
	troubleReport *TroubleReport,
	newAttachments []*Attachment,
) error {
	logger.DBTroubleReportsHelper().Info(
		"Updating trouble report %d with %d new attachments", id, len(newAttachments))

	if troubleReport == nil {
		return NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// Add new attachments
	var newAttachmentIDs []int64
	for _, attachment := range newAttachments {
		if attachment == nil {
			continue
		}

		attachmentID, err := s.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range newAttachmentIDs {
				s.attachments.Delete(addedID, user)
			}
			return WrapError(err, "failed to add new attachment")
		}
		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	// Combine existing and new attachment IDs
	allAttachmentIDs := append(troubleReport.LinkedAttachments, newAttachmentIDs...)
	troubleReport.LinkedAttachments = allAttachmentIDs
	troubleReport.ID = id

	// Update the trouble report
	if err := s.troubleReports.Update(troubleReport, user); err != nil {
		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			s.attachments.Delete(attachmentID, user)
		}
		return WrapError(err, "failed to update trouble report")
	}

	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (s *TroubleReportsHelper) RemoveWithAttachments(id int64, user *User) (*TroubleReport, error) {
	logger.DBTroubleReportsHelper().Info("Removing trouble report %d with attachments", id)

	// Get the trouble report to find its attachments
	tr, err := s.troubleReports.Get(id)
	if err != nil {
		return tr, WrapError(err, "failed to get trouble report for removal")
	}

	// Remove the trouble report first
	if err := s.troubleReports.Delete(id, user); err != nil {
		return tr, WrapError(err, "failed to remove trouble report")
	}

	// Remove associated attachments
	for _, attachmentID := range tr.LinkedAttachments {
		if err := s.attachments.Delete(attachmentID, user); err != nil {
			logger.DBTroubleReportsHelper().Warn("Failed to remove attachment %d: %v", attachmentID, err)
		}
	}

	return tr, nil
}

// LoadAttachments loads attachments for a trouble report.
func (s *TroubleReportsHelper) LoadAttachments(tr *TroubleReport) ([]*Attachment, error) {
	logger.DBTroubleReportsHelper().Debug("Loading attachments for trouble report")

	if tr == nil {
		return nil, NewValidationError("report", "trouble report cannot be nil", nil)
	}

	return s.attachments.GetByIDs(tr.LinkedAttachments)
}

// GetAttachment retrieves a specific attachment by ID.
func (s *TroubleReportsHelper) GetAttachment(id int64) (*Attachment, error) {
	logger.DBTroubleReportsHelper().Debug("Getting attachment with ID %d", id)
	return s.attachments.Get(id)
}
