package attachments

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

func validateAttachment(attachment *models.Attachment) error {
	if err := validation.ValidateNotNil(attachment, "attachment"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(attachment.MimeType, "mime_type"); err != nil {
		return err
	}

	if len(attachment.Data) == 0 {
		return utils.NewValidationError("data cannot be empty")
	}

	return nil
}
