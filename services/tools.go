package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/models"
)

const TableNameTools = "tools"

const (
	ToolQueryInsert = `position, format, type, code, regenerating, is_dead, press, binding`
	ToolQuerySelect = `id, position, format, type, code, regenerating, is_dead, press, binding`
	ToolQueryUpdate = `position = ?, format = ?, type = ?, code = ?, regenerating = ?, is_dead = ?, press = ?, binding = ?`
)

type Tools struct {
	*Base
}

func NewTools(r *Registry) *Tools {
	t := &Tools{
		Base: NewBase(r),
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			regenerating INTEGER NOT NULL DEFAULT 0,
			is_dead INTEGER NOT NULL DEFAULT 0,
			press INTEGER,
			binding INTEGER,
			PRIMARY KEY("id" AUTOINCREMENT)
		)`, TableNameTools)

	if err := t.CreateTable(query, TableNameTools); err != nil {
		panic(err)
	}

	return t
}

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
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		TableNameTools, ToolQueryInsert)

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
		SET %s
		WHERE id = ?`,
		TableNameTools, ToolQueryUpdate)

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

func (t *Tools) List() ([]*models.Tool, error) {
	slog.Debug("Listing tools")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
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

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	slog.Debug("Getting active tools for press", "press", pressNumber)

	if !models.IsValidPressNumber(&pressNumber) {
		slog.Error("Invalid press number (must be 0-5)", "press", pressNumber)
		return nil
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE regenerating = 0 AND is_dead = 0 AND press = ?`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		slog.Error("Failed to query active tools", "error", err)
		return nil
	}
	defer rows.Close()

	tools, err := ScanRows(rows, scanTool)
	if err != nil {
		slog.Error("Failed to scan active tools", "error", err)
		return nil
	}

	return tools
}

func (t *Tools) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	slog.Debug("Getting tools by press", "press", pressNumber)

	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, errors.NewValidationError(
			fmt.Sprintf("invalid press number: %d (must be 0-5)", *pressNumber),
		)
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE press = ? AND regenerating = 0 AND is_dead = 0`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) GetPressUtilization() ([]models.PressUtilization, error) {
	slog.Debug("Getting press utilization")

	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}
	utilization := make([]models.PressUtilization, 0, len(validPresses))

	for _, pressNum := range validPresses {
		tools := t.GetActiveToolsForPress(pressNum)
		count := len(tools)

		utilization = append(utilization, models.PressUtilization{
			PressNumber: pressNum,
			Tools:       tools,
			Count:       count,
			Available:   count == 0,
		})
	}

	return utilization, nil
}

func (t *Tools) ListToolsNotDead() ([]*models.Tool, error) {
	slog.Debug("Listing active tools")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 0
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) ListDeadTools() ([]*models.Tool, error) {
	slog.Debug("Listing dead tools")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE is_dead = 1
		ORDER BY format ASC, code ASC`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.GetSelectError(err)
	}
	defer rows.Close()

	return ScanRows(rows, scanTool)
}

func (t *Tools) UpdatePress(toolID models.ToolID, pressNumber *models.PressNumber, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if !models.IsValidPressNumber(pressNumber) && pressNumber != nil {
		return errors.NewValidationError(
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	slog.Debug("Updating tool press", "user_name", user.Name, "tool_id", toolID, "press", pressNumber)

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("get tool for press update: %v", err)
	}

	query := fmt.Sprintf(`UPDATE %s SET press = ? WHERE id = ?`, TableNameTools)
	if _, err = t.DB.Exec(query, pressNumber, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	// Handle binding - update press for bound tool
	if tool.Binding != nil {
		query = fmt.Sprintf(`UPDATE %s SET press = ? WHERE id = ?`, TableNameTools)
		if _, err = t.DB.Exec(query, pressNumber, *tool.Binding); err != nil {
			return t.GetUpdateError(err)
		}
	}

	return nil
}

func (t *Tools) UpdateRegenerating(toolID models.ToolID, regenerating bool, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	slog.Debug("Updating tool regenerating status", "user_name", user.Name, "tool_id", toolID, "regenerating", regenerating)

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("get current tool state: %v", err)
	}

	if currentTool.Regenerating == regenerating {
		return nil
	}

	query := fmt.Sprintf(`UPDATE %s SET regenerating = ? WHERE id = ?`, TableNameTools)
	if _, err = t.DB.Exec(query, regenerating, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) MarkAsDead(toolID models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	slog.Debug("Marking tool as dead", "user_name", user.Name, "tool_id", toolID)

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 1, press = NULL WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) ReviveTool(toolID models.ToolID, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	slog.Debug("Reviving dead tool", "user_name", user.Name, "tool_id", toolID)

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 0 WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) Bind(cassetteID, targetID models.ToolID) error {
	if err := t.validateBindingTools(cassetteID, targetID); err != nil {
		return err
	}

	// Get press from the target tool
	targetTool, err := t.Get(targetID)
	if err != nil {
		return err
	}

	// Execute binding operations
	queries := []string{
		// Set bindings
		fmt.Sprintf(`UPDATE %s SET binding = :target WHERE id = :cassette`, TableNameTools),
		fmt.Sprintf(`UPDATE %s SET binding = :cassette WHERE id = :target`, TableNameTools),
		// Clear press from other cassettes at the same press
		fmt.Sprintf(`UPDATE %s SET press = NULL WHERE press = :press AND position = "cassette top"`, TableNameTools),
		// Assign press to cassette
		fmt.Sprintf(`UPDATE %s SET press = :press WHERE id = :cassette`, TableNameTools),
	}

	for _, query := range queries {
		if _, err := t.DB.Exec(query,
			sql.Named("target", targetID),
			sql.Named("cassette", cassetteID),
			sql.Named("press", targetTool.Press),
		); err != nil {
			return t.GetUpdateError(err)
		}
	}

	return nil
}

func (t *Tools) UnBind(toolID models.ToolID) error {
	tool, err := t.Get(toolID)
	if err != nil {
		return err
	}

	if tool.Binding == nil {
		return nil
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET binding = NULL
		WHERE id = :toolID OR id = :binding`,
		TableNameTools)

	if _, err := t.DB.Exec(query,
		sql.Named("toolID", toolID),
		sql.Named("binding", *tool.Binding),
	); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

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
