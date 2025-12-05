package services

import (
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) validateToolUniqueness(tool *models.Tool, excludeID models.ToolID) bool {
	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return false
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE id != ? AND position = ? AND format = ? AND code = ?`,
		TableNameTools)

	count, err := t.QueryCount(query, excludeID, tool.Position, formatBytes, tool.Code)
	if err != nil {
		return false
	}

	if count > 0 {
		err = fmt.Errorf(
			"tool with position %s, format %s, and code %s already exists",
			tool.Position, tool.Format, tool.Code,
		)
		return false
	}

	return true
}

// validateBindingTools validates that two tools can be bound together.
// It ensures:
// - The cassette tool is a top cassette position tool
// - The target tool is a top position tool
// - Neither tool is already bound to another tool (prevents multiple bindings)
func (t *Tools) validateBindingTools(cassetteID, targetID models.ToolID) bool {
	cassetteTool, err := t.Get(cassetteID)
	if err != nil {
		return false
	}

	if cassetteTool.Position != models.PositionTopCassette {
		return false
	}

	if cassetteTool.Binding != nil {
		return false
	}

	targetTool, err := t.Get(targetID)
	if err != nil {
		return false
	}

	if targetTool.Position != models.PositionTop {
		return false
	}

	if targetTool.Binding != nil {
		return false
	}

	return true
}

func (t *Tools) marshalFormat(format models.Format) ([]byte, error) {
	formatBytes, err := json.Marshal(format)
	if err != nil {
		return nil, fmt.Errorf("marshal tool format: %v", err)
	}
	return formatBytes, nil
}
