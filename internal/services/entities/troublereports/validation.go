package troublereports

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ValidateTroubleReport performs comprehensive trouble report validation
func ValidateTroubleReport(report *models.TroubleReport) error {
	if err := validation.ValidateNotNil(report, "trouble_report"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(report.Title, "title"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(report.Content, "content"); err != nil {
		return err
	}

	// LinkedAttachments can be empty, but if present should contain valid IDs
	for i, attachmentID := range report.LinkedAttachments {
		if err := validation.ValidateID(attachmentID, fmt.Sprintf("attachment_%d", i)); err != nil {
			return err
		}
	}

	return nil
}
