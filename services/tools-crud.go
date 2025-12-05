package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) Add(tool *models.Tool, user *models.User) (models.ToolID, *errors.MasterError) {
	if !tool.Validate() || !user.Validate() {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	if !t.validateToolUniqueness(tool, 0) {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return 0, errors.NewMasterError(errors.ErrValidation)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (position, format, type, code, regenerating, is_dead, press, binding)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		TableNameTools)

	result, err := t.DB.Exec(query,
		tool.Position,
		formatBytes,
		tool.Type,
		tool.Code,
		tool.Regenerating,
		tool.IsDead,
		tool.Press,
		tool.Binding)
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err)
	}

	return models.ToolID(id), nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) *errors.MasterError {
	if !user.Validate() || !tool.Validate() {
		return errors.NewMasterError(errors.ErrValidation)
	}

	if !t.validateToolUniqueness(tool, tool.ID) {
		return errors.NewMasterError(errors.ErrValidation)
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return errors.NewMasterError(errors.ErrValidation)
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET position = ?, format = ?, type = ?, code = ?, regenerating = ?, is_dead = ?, press = ?, binding = ?
		WHERE id = ?`,
		TableNameTools)

	_, err = t.DB.Exec(query,
		tool.Position,
		formatBytes,
		tool.Type,
		tool.Code,
		tool.Regenerating,
		tool.IsDead,
		tool.Press,
		tool.Binding,
		tool.ID)
	if err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}

func (t *Tools) Get(id models.ToolID) (*models.Tool, *errors.MasterError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE id = ?`,
		ToolQuerySelect, TableNameTools)

	row := t.DB.QueryRow(query, id)

	tool, err := ScanTool(row)
	if err != nil {
		return tool, errors.NewMasterError(err)
	}
	return tool, nil
}

func (t *Tools) Delete(id models.ToolID, user *models.User) *errors.MasterError {
	if !user.Validate() {
		return errors.NewMasterError(errors.ErrValidation)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTools)
	_, err := t.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err)
	}

	return nil
}
