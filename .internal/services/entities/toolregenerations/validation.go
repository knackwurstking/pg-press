package toolregenerations

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ValidateToolRegeneration performs comprehensive tool regeneration validation
func ValidateToolRegeneration(regen *models.Regeneration) error {
	if err := validation.ValidateNotNil(regen, "regeneration"); err != nil {
		return err
	}

	if err := validation.ValidateID(regen.ToolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateID(regen.CycleID, "cycle"); err != nil {
		return err
	}

	// Reason is optional, but if provided should not be empty
	if regen.Reason != "" {
		if err := validation.ValidateNotEmpty(regen.Reason, "reason"); err != nil {
			return err
		}
	}

	return nil
}
