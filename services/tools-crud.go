package services

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) Add(tool *models.Tool, user *models.User) (models.ToolID, *errors.MasterError) {
	if !tool.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid tool: %s", tool), http.StatusBadRequest)
	}
	if !user.Validate() {
		return 0, errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	if !t.validateToolUniqueness(tool, 0) {
		return 0, errors.NewMasterError(fmt.Errorf("tool not unique: %#v", tool), http.StatusBadRequest)
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
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
		return 0, errors.NewMasterError(err, 0)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewMasterError(err, 0)
	}

	return models.ToolID(id), nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) *errors.MasterError {
	if !tool.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid tool: %s", tool), http.StatusBadRequest)
	}

	if !user.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	if !t.validateToolUniqueness(tool, tool.ID) {
		return errors.NewMasterError(fmt.Errorf("tool not unique: %#v", tool), http.StatusBadRequest)
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return errors.NewMasterError(err, 0)
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
		return errors.NewMasterError(err, 0)
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
		return tool, errors.NewMasterError(err, 0)
	}
	return tool, nil
}

func (t *Tools) Delete(id models.ToolID, user *models.User) *errors.MasterError {
	if !user.Validate() {
		return errors.NewMasterError(fmt.Errorf("invalid user: %s", user), http.StatusBadRequest)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTools)
	_, err := t.DB.Exec(query, id)
	if err != nil {
		return errors.NewMasterError(err, 0)
	}

	return nil
}
