package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) validateToolUniqueness(tool *models.Tool, excludeID models.ToolID) error {
	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE id != ? AND position = ? AND format = ? AND code = ?`,
		TableNameTools)

	count, err := t.QueryCount(query, excludeID, tool.Position, formatBytes, tool.Code)
	if err != nil {
		return t.GetSelectError(err)
	}

	if count > 0 {
		return errors.NewAlreadyExistsError(
			fmt.Sprintf("tool with position %s, format %s, and code %s already exists",
				tool.Position, tool.Format, tool.Code))
	}

	return nil
}

// validateBindingTools validates that two tools can be bound together.
// It ensures:
// - The cassette tool is a top cassette position tool
// - The target tool is a top position tool
// - Neither tool is already bound to another tool (prevents multiple bindings)
func (t *Tools) validateBindingTools(cassetteID, targetID models.ToolID) error {
	cassetteTool, err := t.Get(cassetteID)
	if err != nil {
		return err
	}

	if cassetteTool.Position != models.PositionTopCassette {
		return errors.NewValidationError(
			fmt.Sprintf("tool %d is not a top cassette", cassetteID))
	}

	if cassetteTool.Binding != nil {
		return errors.NewValidationError(
			fmt.Sprintf("cassette tool %d is already bound to tool %d", cassetteID, *cassetteTool.Binding))
	}

	targetTool, err := t.Get(targetID)
	if err != nil {
		return err
	}

	if targetTool.Position != models.PositionTop {
		return errors.NewValidationError(
			fmt.Sprintf("tool %d is not a top tool", targetID))
	}

	if targetTool.Binding != nil {
		return errors.NewValidationError(
			fmt.Sprintf("target tool %d is already bound to tool %d", targetID, *targetTool.Binding))
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

func scanTool(scannable Scannable) (*models.Tool, error) {
	tool := &models.Tool{}
	var format []byte

	err := scannable.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.IsDead, &tool.Press, &tool.Binding)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("scan tool: %v", err)
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("unmarshal tool format: %v", err)
	}

	return tool, nil
}
