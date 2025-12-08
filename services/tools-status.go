package services

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) UpdatePress(toolID models.ToolID, pressNumber *models.PressNumber, user *models.User) *errors.MasterError {
	if !models.IsValidPressNumber(pressNumber) {
		return errors.NewMasterError(fmt.Errorf("invalid press number: %d", pressNumber), http.StatusBadRequest)
	}

	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	tool, merr := t.Get(toolID)
	if merr != nil {
		return merr
	}

	query := `UPDATE tools SET press = ? WHERE id = ?`
	_, err := t.DB.Exec(query, pressNumber, toolID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	// Handle binding - update press for bound tool
	if tool.Binding == nil {
		return nil
	}

	query = `UPDATE tools SET press = ? WHERE id = ?`
	_, err = t.DB.Exec(query, pressNumber, *tool.Binding)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (t *Tools) UpdateRegenerating(toolID models.ToolID, regenerating bool, user *models.User) *errors.MasterError {
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, merr := t.Get(toolID)
	if merr != nil {
		return merr
	}

	if currentTool.Regenerating == regenerating {
		return nil
	}

	query := `UPDATE tools SET regenerating = ? WHERE id = ?`
	_, err := t.DB.Exec(query, regenerating, toolID)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (t *Tools) MarkAsDead(toolID models.ToolID, user *models.User) *errors.MasterError {
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	query := `UPDATE tools SET is_dead = 1, press = NULL WHERE id = ?`
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}

func (t *Tools) ReviveTool(toolID models.ToolID, user *models.User) *errors.MasterError {
	verr := user.Validate()
	if verr != nil {
		return verr.MasterError()
	}

	query := `UPDATE tools SET is_dead = 0 WHERE id = ?`
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
