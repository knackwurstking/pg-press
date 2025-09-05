package troublereport

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/database/dberror"
	"github.com/knackwurstking/pgpress/internal/database/interfaces"
	"github.com/knackwurstking/pgpress/internal/database/models"
	"github.com/knackwurstking/pgpress/internal/database/services/attachment"
	"github.com/knackwurstking/pgpress/internal/logger"
)

// Helper provides high-level operations for trouble reports
// with attachment management.
type Helper struct {
	troubleReports interfaces.DataOperations[*models.TroubleReport]
	attachments    *attachment.Service
}

// NewTroubleReportsHelper creates a new helper instance.
func NewTroubleReportsHelper(
	troubleReports interfaces.DataOperations[*models.TroubleReport],
	attachments *attachment.Service,
) *Helper {
	return &Helper{
		troubleReports: troubleReports,
		attachments:    attachments,
	}
}

// GetWithAttachments retrieves a trouble report and loads its attachments.
func (h *Helper) GetWithAttachments(
	id int64,
) (*models.TroubleReportWithAttachments, error) {
	logger.DBTroubleReportsHelper().Debug(
		"Getting trouble report with attachments, id: %d", id)

	// Get the trouble report
	tr, err := h.troubleReports.Get(id)
	if err != nil {
		return nil, err
	}

	// Load attachments
	attachments, err := h.attachments.GetByIDs(tr.LinkedAttachments)
	if err != nil {
		return nil, dberror.WrapError(err, "failed to load attachments for trouble report")
	}

	return &models.TroubleReportWithAttachments{
		TroubleReport:     tr,
		LoadedAttachments: attachments,
	}, nil
}

// ListWithAttachments retrieves all trouble reports and loads their attachments.
func (h *Helper) ListWithAttachments() ([]*models.TroubleReportWithAttachments, error) {
	logger.DBTroubleReportsHelper().Debug("Listing trouble reports with attachments")

	// Get all trouble reports
	reports, err := h.troubleReports.List()
	if err != nil {
		return nil, err
	}

	var result []*models.TroubleReportWithAttachments

	for _, tr := range reports {
		// Load attachments for each report
		attachments, err := h.attachments.GetByIDs(tr.LinkedAttachments)
		if err != nil {
			return nil, dberror.WrapError(err,
				fmt.Sprintf("failed to load attachments for trouble report %d", tr.ID))
		}

		result = append(result, &models.TroubleReportWithAttachments{
			TroubleReport:     tr,
			LoadedAttachments: attachments,
		})
	}

	return result, nil
}

// AddWithAttachments creates a new trouble report and its attachments.
func (h *Helper) AddWithAttachments(
	user *models.User,
	troubleReport *models.TroubleReport,
	attachments []*models.Attachment,
) error {
	logger.DBTroubleReportsHelper().Info("Adding trouble report with %d attachments", len(attachments))

	if troubleReport == nil {
		return dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// First, add the attachments and collect their IDs
	var attachmentIDs []int64
	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		id, err := h.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range attachmentIDs {
				h.attachments.Delete(addedID, user)
			}
			return dberror.WrapError(err, "failed to add attachment")
		}
		attachmentIDs = append(attachmentIDs, id)
	}

	// Set the attachment IDs in the trouble report
	troubleReport.LinkedAttachments = attachmentIDs

	// Add the trouble report
	if _, err := h.troubleReports.Add(troubleReport, user); err != nil {
		// Cleanup attachments on failure
		for _, id := range attachmentIDs {
			h.attachments.Delete(id, user)
		}
		return dberror.WrapError(err, "failed to add trouble report")
	}

	return nil
}

// UpdateWithAttachments updates a trouble report and manages its attachments.
func (h *Helper) UpdateWithAttachments(
	user *models.User,
	id int64,
	troubleReport *models.TroubleReport,
	newAttachments []*models.Attachment,
) error {
	logger.DBTroubleReportsHelper().Info(
		"Updating trouble report %d with %d new attachments", id, len(newAttachments))

	if troubleReport == nil {
		return dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	// Add new attachments
	var newAttachmentIDs []int64
	for _, attachment := range newAttachments {
		if attachment == nil {
			continue
		}

		attachmentID, err := h.attachments.Add(attachment, user)
		if err != nil {
			// Cleanup already added attachments on failure
			for _, addedID := range newAttachmentIDs {
				h.attachments.Delete(addedID, user)
			}
			return dberror.WrapError(err, "failed to add new attachment")
		}
		newAttachmentIDs = append(newAttachmentIDs, attachmentID)
	}

	// Combine existing and new attachment IDs
	allAttachmentIDs := append(troubleReport.LinkedAttachments, newAttachmentIDs...)
	troubleReport.LinkedAttachments = allAttachmentIDs
	troubleReport.ID = id

	// Update the trouble report
	if err := h.troubleReports.Update(troubleReport, user); err != nil {
		// Cleanup new attachments on failure
		for _, attachmentID := range newAttachmentIDs {
			h.attachments.Delete(attachmentID, user)
		}
		return dberror.WrapError(err, "failed to update trouble report")
	}

	return nil
}

// RemoveWithAttachments removes a trouble report and its attachments.
func (h *Helper) RemoveWithAttachments(id int64, user *models.User) (*models.TroubleReport, error) {
	logger.DBTroubleReportsHelper().Info("Removing trouble report %d with attachments", id)

	// Get the trouble report to find its attachments
	tr, err := h.troubleReports.Get(id)
	if err != nil {
		return tr, dberror.WrapError(err, "failed to get trouble report for removal")
	}

	// Remove the trouble report first
	if err := h.troubleReports.Delete(id, user); err != nil {
		return tr, dberror.WrapError(err, "failed to remove trouble report")
	}

	// Remove associated attachments
	for _, attachmentID := range tr.LinkedAttachments {
		if err := h.attachments.Delete(attachmentID, user); err != nil {
			logger.DBTroubleReportsHelper().Warn("Failed to remove attachment %d: %v", attachmentID, err)
		}
	}

	return tr, nil
}

// LoadAttachments loads attachments for a trouble report.
func (h *Helper) LoadAttachments(tr *models.TroubleReport) ([]*models.Attachment, error) {
	logger.DBTroubleReportsHelper().Debug("Loading attachments for trouble report")

	if tr == nil {
		return nil, dberror.NewValidationError("report", "trouble report cannot be nil", nil)
	}

	return h.attachments.GetByIDs(tr.LinkedAttachments)
}

// GetAttachment retrieves a specific attachment by ID.
func (h *Helper) GetAttachment(id int64) (*models.Attachment, error) {
	logger.DBTroubleReportsHelper().Debug("Getting attachment with ID %d", id)
	return h.attachments.Get(id)
}
