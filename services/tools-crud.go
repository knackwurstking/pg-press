package services

import (
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) Add(tool *models.Tool, user *models.User) (models.ToolID, *errors.DBError) {
	err := tool.Validate()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	err = user.Validate()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
	}

	dberr := t.validateToolUniqueness(tool, 0)
	if dberr != nil {
		return 0, errors.NewDBError(dberr, errors.DBTypeValidation)
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeValidation)
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
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.NewDBError(err, errors.DBTypeInsert)
	}

	return models.ToolID(id), nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) *errors.DBError {
	err := tool.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	err = user.Validate()
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	dberr := t.validateToolUniqueness(tool, tool.ID)
	if dberr != nil {
		return dberr
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
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
		return errors.NewDBError(err, errors.DBTypeUpdate)
	}

	return nil
}

func (t *Tools) Get(id models.ToolID) (*models.Tool, *errors.DBError) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE id = ?`,
		ToolQuerySelect, TableNameTools)

	row := t.DB.QueryRow(query, id)

	return ScanRow(row, ScanTool)
}

func (t *Tools) Delete(id models.ToolID, user *models.User) *errors.DBError {
	if err := user.Validate(); err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTools)
	_, err := t.DB.Exec(query, id)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeDelete)
	}

	return nil
}
