package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) UpdatePress(toolID models.ToolID, pressNumber *models.PressNumber, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if !models.IsValidPressNumber(pressNumber) && pressNumber != nil {
		return errors.NewValidationError(
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("get tool for press update: %v", err)
	}

	query := fmt.Sprintf(`UPDATE %s SET press = ? WHERE id = ?`, TableNameTools)
	if _, err = t.DB.Exec(query, pressNumber, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	// Handle binding - update press for bound tool
	if tool.Binding != nil {
		query = fmt.Sprintf(`UPDATE %s SET press = ? WHERE id = ?`, TableNameTools)
		if _, err = t.DB.Exec(query, pressNumber, *tool.Binding); err != nil {
			return t.GetUpdateError(err)
		}
	}

	return nil
}

func (t *Tools) UpdateRegenerating(toolID models.ToolID, regenerating bool, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("get current tool state: %v", err)
	}

	if currentTool.Regenerating == regenerating {
		return nil
	}

	query := fmt.Sprintf(`UPDATE %s SET regenerating = ? WHERE id = ?`, TableNameTools)
	if _, err = t.DB.Exec(query, regenerating, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) MarkAsDead(toolID models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 1, press = NULL WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) ReviveTool(toolID models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 0 WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}
