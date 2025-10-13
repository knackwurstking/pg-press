package tools

import (
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

// ValidateTool performs comprehensive tool validation
func ValidateTool(tool *models.Tool) error {
	if err := validation.ValidateNotNil(tool, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(string(tool.Position), "position"); err != nil {
		return err
	}

	if err := validation.ValidateNotEmpty(tool.Code, "code"); err != nil {
		return err
	}

	return nil
}

// ************************** //
// Service validation methods //
// ************************** //

func (t *Service) validateToolUniqueness(tool *models.Tool, excludeID int64) error {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return fmt.Errorf("failed to marshal tool format: %v", err)
	}

	exists, err := t.CheckExistence(`
		SELECT COUNT(*) FROM tools
		WHERE id != ? AND position = ? AND format = ? AND code = ?
	`, excludeID, tool.Position, formatBytes, tool.Code)

	if err != nil {
		return t.HandleSelectError(err, "tools")
	}

	if exists {
		return utils.NewAlreadyExistsError(
			fmt.Sprintf(
				"tool with position %s, format %s, and code %s already exists",
				tool.Position, tool.Format, tool.Code,
			),
		)
	}

	return nil
}

func (s *Service) validateBindingTools(cassetteID, targetID int64) error {
	// Validate cassete tool, has to be a top cassette position tool
	tool, err := s.Get(cassetteID)
	if err != nil {
		return err
	}
	if tool.Position != models.PositionTopCassette {
		return utils.NewValidationError(
			fmt.Sprintf("tool %d is not a top cassette", cassetteID))
	}

	// Validate target tools position, has to be a top position tool
	tool, err = s.Get(targetID)
	if err != nil {
		return err
	}
	if tool.Position != models.PositionTop {
		return utils.NewValidationError(
			fmt.Sprintf("tool %d is not a top tool", targetID))
	}

	return nil
}
