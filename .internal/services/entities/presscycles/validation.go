package presscycles

import (
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// ValidatePressCycle performs comprehensive press cycle validation
func ValidatePressCycle(cycle *models.Cycle) error {
	if err := validation.ValidateNotNil(cycle, "cycle"); err != nil {
		return err
	}

	if err := ValidatePressNumber(cycle.PressNumber); err != nil {
		return err
	}

	if err := validation.ValidateID(cycle.ToolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(string(cycle.ToolPosition), "tool_position"); err != nil {
		return err
	}

	if err := ValidatePositive(cycle.TotalCycles, "total_cycles"); err != nil {
		return err
	}

	if cycle.Date.IsZero() {
		return utils.NewValidationError("date cannot be zero")
	}

	return nil
}

// ValidatePressNumber validates that a press number is within valid range (0-5)
func ValidatePressNumber(pressNumber models.PressNumber) error {
	if pressNumber < 0 || pressNumber > 5 {
		return utils.NewValidationError(fmt.Sprintf("press_number must be between 0 and 5, got: %d", pressNumber))
	}
	return nil
}

// ValidatePositive checks if a value is positive
func ValidatePositive(value int64, fieldName string) error {
	if value <= 0 {
		return utils.NewValidationError(fmt.Sprintf("%s must be positive", fieldName))
	}
	return nil
}
