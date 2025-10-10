package modifications

import (
	"fmt"
	"slices"

	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

func validateModificationType(modType string) error {
	if err := validation.ValidateNotEmpty(modType, "modification_type"); err != nil {
		return err
	}

	validTypes := []string{
		"trouble_reports",
		"metal_sheets",
		"tools",
		"press_cycles",
		"users",
		"notes",
		"attachments",
	}

	if slices.Contains(validTypes, modType) {
		return nil
	}

	return utils.NewValidationError(fmt.Sprintf("invalid modification type: %s", modType))
}
