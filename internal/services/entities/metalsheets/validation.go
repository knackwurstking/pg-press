package metalsheets

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

func validateMetalSheet(sheet *models.MetalSheet) error {
	if err := validation.ValidateNotNil(sheet, "metal_sheet"); err != nil {
		return err
	}

	if sheet.TileHeight <= 0 {
		return utils.NewValidationError("tile_height must be positive")
	}

	if sheet.Value <= 0 {
		return utils.NewValidationError("value must be positive")
	}

	if sheet.MarkeHeight <= 0 {
		return utils.NewValidationError("marke_height must be positive")
	}

	if sheet.STF <= 0 {
		return utils.NewValidationError("stf must be positive")
	}

	if sheet.STFMax <= 0 {
		return utils.NewValidationError("stf_max must be positive")
	}

	if err := validation.ValidateID(sheet.ToolID, "tool"); err != nil {
		return err
	}

	// Validate machine type identifier
	if !sheet.Identifier.IsValid() {
		return utils.NewValidationError("invalid machine type identifier")
	}

	return nil
}
