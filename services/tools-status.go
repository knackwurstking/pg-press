package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) UpdatePress(toolID models.ToolID, pressNumber *models.PressNumber, user *models.User) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	if !models.IsValidPressNumber(pressNumber) && pressNumber != nil {
		return errors.NewDBError(
			fmt.Errorf("invalid press number: %d", pressNumber),
			errors.DBTypeValidation,
		)
	}

	tool, dberr := t.Get(toolID)
	if dberr != nil {
		return dberr
	}

	query := fmt.Sprintf(`UPDATE %s SET press = ? WHERE id = ?`, TableNameTools)
	_, err = t.DB.Exec(query, pressNumber, toolID)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	// Handle binding - update press for bound tool
	if tool.Binding == nil {
		return nil
	}

	query = fmt.Sprintf(`UPDATE %s SET press = ? WHERE id = ?`, TableNameTools)
	_, err = t.DB.Exec(query, pressNumber, *tool.Binding)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (t *Tools) UpdateRegenerating(toolID models.ToolID, regenerating bool, user *models.User) *errors.DBError {
	err := user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, dberr := t.Get(toolID)
	if dberr != nil {
		return dberr
	}

	if currentTool.Regenerating == regenerating {
		return nil
	}

	query := fmt.Sprintf(`UPDATE %s SET regenerating = ? WHERE id = ?`, TableNameTools)
	_, err = t.DB.Exec(query, regenerating, toolID)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (t *Tools) MarkAsDead(toolID models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 1, press = NULL WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (t *Tools) ReviveTool(toolID models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 0 WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}
