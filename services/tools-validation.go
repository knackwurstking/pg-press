package services

import (
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) validateToolUniqueness(tool *models.Tool, excludeID models.ToolID) *errors.DBError {
	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return errors.NewDBError(err, errors.DBTypeValidation)
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE id != ? AND position = ? AND format = ? AND code = ?`,
		TableNameTools)

	count, dberr := t.QueryCount(query, excludeID, tool.Position, formatBytes, tool.Code)
	if dberr != nil {
		return dberr
	}

	if count > 0 {
		err = fmt.Errorf(
			"tool with position %s, format %s, and code %s already exists",
			tool.Position, tool.Format, tool.Code,
		)
		return errors.NewDBError(err, errors.DBTypeExists)
	}

	return nil
}

// validateBindingTools validates that two tools can be bound together.
// It ensures:
// - The cassette tool is a top cassette position tool
// - The target tool is a top position tool
// - Neither tool is already bound to another tool (prevents multiple bindings)
func (t *Tools) validateBindingTools(cassetteID, targetID models.ToolID) *errors.DBError {
	cassetteTool, dberr := t.Get(cassetteID)
	if dberr != nil {
		return dberr
	}

	if cassetteTool.Position != models.PositionTopCassette {
		return errors.NewDBError(
			fmt.Errorf("tool %d is not a top cassette", cassetteID),
			errors.DBTypeValidation,
		)
	}

	if cassetteTool.Binding != nil {
		return errors.NewDBError(
			fmt.Errorf("cassette tool %d is already bound to tool %d", cassetteID, *cassetteTool.Binding),
			errors.DBTypeValidation,
		)
	}

	targetTool, dberr := t.Get(targetID)
	if dberr != nil {
		return dberr
	}

	if targetTool.Position != models.PositionTop {
		return errors.NewDBError(
			fmt.Errorf("tool %d is not a top tool", targetID),
			errors.DBTypeValidation,
		)
	}

	if targetTool.Binding != nil {
		return errors.NewDBError(
			fmt.Errorf("target tool %d is already bound to tool %d", targetID, *targetTool.Binding),
			errors.DBTypeValidation,
		)
	}

	return nil
}

func (t *Tools) marshalFormat(format models.Format) ([]byte, error) {
	formatBytes, err := json.Marshal(format)
	if err != nil {
		return nil, fmt.Errorf("marshal tool format: %v", err)
	}
	return formatBytes, nil
}
