package tools

import (
	"database/sql"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

func (t *Service) Add(tool *models.Tool, user *models.User) (int64, error) {
	if err := ValidateTool(tool); err != nil {
		return 0, err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	t.Log.Debug("Adding tool by %s: position: %s, type: %s, code: %s", user.String(), tool.Position, tool.Type, tool.Code)

	if err := t.validateToolUniqueness(tool, 0); err != nil {
		return 0, err
	}

	formatBytes, err := marshalFormat(tool.Format)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf(`
		INSERT INTO tools (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, ToolQueryInsert)

	result, err := t.DB.Exec(query, tool.Position, formatBytes, tool.Type, tool.Code, tool.Regenerating, tool.IsDead, tool.Press, tool.Binding)
	if err != nil {
		return 0, t.HandleInsertError(err, "tools")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %v", err)
	}

	return id, nil
}

func (t *Service) Get(id int64) (*models.Tool, error) {
	if err := validation.ValidateID(id, "tool"); err != nil {
		return nil, err
	}

	t.Log.Debug("Getting tool: %d", id)

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			tools
		WHERE
			id = ?
	`, ToolQuerySelect)

	row := t.DB.QueryRow(query, id)

	tool, err := scanner.ScanSingleRow(
		row, ScanTool, "tools")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool with ID %d not found", id))
		}
		return nil, err
	}

	return tool, nil
}

func (t *Service) List() ([]*models.Tool, error) {
	t.Log.Debug("Listing tools")

	query := fmt.Sprintf(`
		SELECT
			%s
		FROM
			tools
		ORDER BY format ASC, code ASC
	`, ToolQuerySelect)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	return tools, nil
}

func (t *Service) Delete(id int64, user *models.User) error {
	if err := validation.ValidateID(id, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Debug("Deleting tool by %s: id: %d", user.String(), id)

	query := `DELETE FROM tools WHERE id = ?`
	result, err := t.DB.Exec(query, id)
	if err != nil {
		return t.HandleDeleteError(err, "tools")
	}

	return t.CheckRowsAffected(result, "tool", id)
}
