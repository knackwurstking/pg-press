package notes

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

func validateNote(note *models.Note) error {
	if err := validation.ValidateNotNil(note, "note"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(note.Content, "content"); err != nil {
		return err
	}

	if note.Level < 0 {
		return utils.NewValidationError("level must be non-negative")
	}

	return nil
}
