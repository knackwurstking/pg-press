package tools

import (
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
)

// ValidateTool performs comprehensive tool validation
func ValidateTool(tool *models.Tool) error {
	if err := validation.ValidateNotNil(tool, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(string(tool.Position), "position"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(tool.Type, "type"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(tool.Code, "code"); err != nil {
		return err
	}

	// Format validation would be handled by the model's own validation
	// since models.Format might not be a pointer type

	return nil
}
