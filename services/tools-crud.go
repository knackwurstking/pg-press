package services

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

func (t *Tools) Add(tool *models.Tool, user *models.User) (models.ToolID, error) {
	if err := tool.Validate(); err != nil {
		return 0, err
	}

	if err := user.Validate(); err != nil {
		return 0, err
	}

	slog.Debug("Adding tool", "user_name", user.Name, "position", tool.Position, "type", tool.Type, "code", tool.Code)

	if err := t.validateToolUniqueness(tool, 0); err != nil {
		return 0, err
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return 0, err
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
		return 0, t.GetInsertError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get last insert ID: %v", err)
	}

	return models.ToolID(id), nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) error {
	if err := tool.Validate(); err != nil {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
	}

	slog.Debug("Updating tool", "user_name", user.Name, "tool_id", tool.ID)

	if err := t.validateToolUniqueness(tool, tool.ID); err != nil {
		return err
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return err
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
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) Get(id models.ToolID) (*models.Tool, error) {
	slog.Debug("Getting tool", "tool", id)

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE id = ?`,
		ToolQuerySelect, TableNameTools)

	row := t.DB.QueryRow(query, id)

	tool, err := ScanSingleRow(row, scanTool)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError(fmt.Sprintf("tool with ID %d not found", id))
		}
		return nil, t.GetSelectError(err)
	}

	return tool, nil
}

func (t *Tools) Delete(id models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	slog.Debug("Deleting tool", "user_name", user.Name, "tool_id", id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTools)
	_, err := t.DB.Exec(query, id)
	if err != nil {
		return t.GetDeleteError(err)
	}

	return nil
}
