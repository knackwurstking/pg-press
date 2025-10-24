package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/errors"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
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
		Base: NewBase(r, logger.NewComponentLogger("Service: Tools")),
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

func (t *Tools) Add(tool *models.Tool, user *models.User) (int64, error) {
	if err := tool.Validate(); err != nil {
		return 0, err
	}

	if err := user.Validate(); err != nil {
		return 0, err
	}

	t.Log.Debug("Adding tool by %s: position: %s, type: %s, code: %s",
		user.String(), tool.Position, tool.Type, tool.Code)

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
		return 0, fmt.Errorf("failed to get last insert ID: %v", err)
	}

	return id, nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) error {
	if err := tool.Validate(); err != nil {
		return err
	}

	if err := user.Validate(); err != nil {
		return err
	}

	t.Log.Debug("Updating tool by %s: id: %d", user.String(), tool.ID)

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

func (t *Tools) Get(id int64) (*models.Tool, error) {
	t.Log.Debug("Getting tool: %d", id)

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
	t.Log.Debug("Listing tools")

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

func (t *Tools) Delete(id int64, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	t.Log.Debug("Deleting tool by %s: id: %d", user.String(), id)

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, TableNameTools)
	_, err := t.DB.Exec(query, id)
	if err != nil {
		return t.GetDeleteError(err)
	}

	return nil
}

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	t.Log.Debug("Getting active tools for press: %d", pressNumber)

	if !models.IsValidPressNumber(&pressNumber) {
		t.Log.Error("Invalid press number: %d (must be 0-5)", pressNumber)
		return nil
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE regenerating = 0 AND is_dead = 0 AND press = ?`,
		ToolQuerySelect, TableNameTools)

	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		t.Log.Error("Failed to query active tools: %v", err)
		return nil
	}
	defer rows.Close()

	tools, err := ScanRows(rows, scanTool)
	if err != nil {
		t.Log.Error("Failed to scan active tools: %v", err)
		return nil
	}

	return tools
}

func (t *Tools) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	t.Log.Debug("Getting tools by press: %v", pressNumber)

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
	t.Log.Debug("Getting press utilization")

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
	t.Log.Debug("Listing active tools")

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
	t.Log.Debug("Listing dead tools")

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

func (t *Tools) UpdatePress(toolID int64, pressNumber *models.PressNumber, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if !models.IsValidPressNumber(pressNumber) {
		return errors.NewValidationError(
			fmt.Sprintf("invalid press number: %d", pressNumber),
		)
	}

	t.Log.Debug("Updating tool press by %s: toolID: %d, pressNumber: %v",
		user.String(), toolID, pressNumber)

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %v", err)
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

func (t *Tools) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	t.Log.Debug("Updating tool regenerating status by %s: toolID: %d, regenerating: %t",
		user.String(), toolID, regenerating)

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get current tool state: %v", err)
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

func (t *Tools) MarkAsDead(toolID int64, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	t.Log.Debug("Marking tool as dead by %s: id: %d", user.String(), toolID)

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 1, press = NULL WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) ReviveTool(toolID int64, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	t.Log.Debug("Reviving dead tool by %s: id: %d", user.String(), toolID)

	query := fmt.Sprintf(`UPDATE %s SET is_dead = 0 WHERE id = ?`, TableNameTools)
	if _, err := t.DB.Exec(query, toolID); err != nil {
		return t.GetUpdateError(err)
	}

	return nil
}

func (t *Tools) Bind(cassetteID, targetID int64) error {
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

func (t *Tools) UnBind(toolID int64) error {
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

func (t *Tools) validateToolUniqueness(tool *models.Tool, excludeID int64) error {
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
func (t *Tools) validateBindingTools(cassetteID, targetID int64) error {
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
		return nil, fmt.Errorf("failed to marshal tool format: %v", err)
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
		return nil, fmt.Errorf("failed to scan tool: %v", err)
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool format: %v", err)
	}

	return tool, nil
}
