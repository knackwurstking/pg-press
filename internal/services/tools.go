package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Tools struct {
	*BaseService
	notes *Notes
}

func NewTools(db *sql.DB, notes *Notes) *Tools {
	base := NewBaseService(db, "Tools")

	t := &Tools{
		BaseService: base,
		notes:       notes,
	}

	if err := t.createTable(); err != nil {
		panic(err)
	}

	return t
}

func (t *Tools) createTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			regenerating BOOLEAN NOT NULL DEFAULT 0,
			press INTEGER,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	return t.CreateTable(query, "tools")
}

func (t *Tools) Add(tool *models.Tool, user *models.User) (int64, error) {
	if err := ValidateTool(tool); err != nil {
		return 0, err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	if err := t.validateToolUniqueness(tool, 0); err != nil {
		return 0, err
	}

	formatBytes, err := t.marshalToolData(tool)
	if err != nil {
		return 0, err
	}

	t.LogOperationWithUser("Adding tool", createUserInfo(user), fmt.Sprintf("position: %s, type: %s, code: %s", tool.Position, tool.Type, tool.Code))

	const insertQuery = `
		INSERT INTO tools (position, format, type, code, regenerating, press)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	result, err := t.db.Exec(insertQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press)
	if err != nil {
		return 0, t.HandleInsertError(err, "tools")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, t.HandleInsertError(err, "tools")
	}

	t.LogOperation("Added tool", fmt.Sprintf("id: %d", id))
	return id, nil
}

func (t *Tools) AddWithNotes(tool *models.Tool, user *models.User, notes ...*models.Note) (*models.ToolWithNotes, error) {
	if err := ValidateTool(tool); err != nil {
		return nil, err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	t.LogOperationWithUser("Adding tool with notes", createUserInfo(user), fmt.Sprintf("notes_count: %d", len(notes)))

	var createdNotes []*models.Note
	for _, note := range notes {
		if err := ValidateNote(note); err != nil {
			return nil, err
		}
		// Link the note to this tool using the generic linked field
		note.Linked = fmt.Sprintf("tool_%d", tool.ID)
		noteID, err := t.notes.Add(note)
		if err != nil {
			return nil, err
		}
		note.ID = noteID
		createdNotes = append(createdNotes, note)
	}

	toolID, err := t.Add(tool, user)
	if err != nil {
		return nil, err
	}

	tool.ID = toolID
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: createdNotes}, nil
}

func (t *Tools) Delete(id int64, user *models.User) error {
	if err := ValidateID(id, "tool"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.LogOperationWithUser("Deleting tool", createUserInfo(user), fmt.Sprintf("id: %d", id))

	const deleteQuery = `DELETE FROM tools WHERE id = $1`
	result, err := t.db.Exec(deleteQuery, id)
	if err != nil {
		return t.HandleDeleteError(err, "tools")
	}

	return t.CheckRowsAffected(result, "tool", id)
}

func (t *Tools) Get(id int64) (*models.Tool, error) {
	if err := ValidateID(id, "tool"); err != nil {
		return nil, err
	}

	t.LogOperation("Getting tool", id)

	const query = `SELECT id, position, format, type, code, regenerating, press FROM tools WHERE id = $1`
	row := t.db.QueryRow(query, id)

	tool, err := ScanSingleRow(row, ScanTool, "tools")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool with ID %d not found", id))
		}
		return nil, err
	}

	return tool, nil
}

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	t.LogOperation("Getting active tools for press", pressNumber)

	const query = `
		SELECT id, position, format, type, code, regenerating, press
		FROM tools WHERE regenerating = 0 AND press = ?
	`
	rows, err := t.db.Query(query, pressNumber)
	if err != nil {
		t.log.Error("Failed to query active tools: %v", err)
		return nil
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		t.log.Error("Failed to scan active tools: %v", err)
		return nil
	}

	t.LogOperation("Found active tools for press", fmt.Sprintf("press: %d, count: %d", pressNumber, len(tools)))
	return tools
}

func (t *Tools) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	t.LogOperation("Getting tools by press", fmt.Sprintf("press: %v", pressNumber))

	const query = `
		SELECT id, position, format, type, code, regenerating, press
		FROM tools WHERE press = $1 AND regenerating = 0
	`
	rows, err := t.db.Query(query, pressNumber)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	t.LogOperation("Found tools by press", fmt.Sprintf("press: %v, count: %d", pressNumber, len(tools)))
	return tools, nil
}

func (t *Tools) GetPressUtilization() ([]models.PressUtilization, error) {
	t.LogOperation("Getting press utilization")

	var utilization []models.PressUtilization

	// Valid press numbers: 0, 2, 3, 4, 5
	validPresses := []models.PressNumber{0, 2, 3, 4, 5}

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

	t.LogOperation("Calculated press utilization", fmt.Sprintf("presses_checked: %d", len(validPresses)))
	return utilization, nil
}

func (t *Tools) GetWithNotes(id int64) (*models.ToolWithNotes, error) {
	if err := ValidateID(id, "tool"); err != nil {
		return nil, err
	}

	t.LogOperation("Getting tool with notes", id)

	tool, err := t.Get(id)
	if err != nil {
		return nil, err
	}

	notes, err := t.notes.GetByTool(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes for tool")
	}

	t.LogOperation("Found tool with notes", fmt.Sprintf("id: %d, notes_count: %d", id, len(notes)))
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Tools) List() ([]*models.Tool, error) {
	t.LogOperation("Listing tools")

	const query = `
		SELECT
			id, position, format, type, code, regenerating, press
		FROM
			tools
		ORDER BY format ASC, code ASC
	`

	rows, err := t.db.Query(query)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	t.LogOperation("Listed tools", fmt.Sprintf("count: %d", len(tools)))
	return tools, nil
}

func (t *Tools) ListWithNotes() ([]*models.ToolWithNotes, error) {
	t.LogOperation("Listing tools with notes")

	tools, err := t.List()
	if err != nil {
		return nil, err
	}

	var result []*models.ToolWithNotes
	for _, tool := range tools {
		notes, err := t.notes.GetByTool(tool.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load notes for tool %d: %v", tool.ID, err)
		}

		result = append(result, &models.ToolWithNotes{Tool: tool, LoadedNotes: notes})
	}

	t.LogOperation("Listed tools with notes", fmt.Sprintf("count: %d", len(result)))
	return result, nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) error {
	if err := ValidateTool(tool); err != nil {
		return err
	}

	if err := ValidateID(tool.ID, "tool"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	if err := t.validateToolUniqueness(tool, tool.ID); err != nil {
		return err
	}

	formatBytes, err := t.marshalToolData(tool)
	if err != nil {
		return err
	}

	t.LogOperationWithUser("Updating tool", createUserInfo(user), fmt.Sprintf("id: %d", tool.ID))

	const updateQuery = `
		UPDATE tools SET position = $1, format = $2, type = $3, code = $4,
		regenerating = $5, press = $6 WHERE id = $7
	`

	result, err := t.db.Exec(updateQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, tool.ID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", tool.ID); err != nil {
		return err
	}

	t.LogOperation("Updated tool", fmt.Sprintf("id: %d", tool.ID))
	return nil
}

func (t *Tools) UpdatePress(toolID int64, press *models.PressNumber, user *models.User) error {
	if err := ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %v", err)
	}

	if equalPressNumbers(tool.Press, press) {
		t.LogOperation("Tool press unchanged", fmt.Sprintf("id: %d", toolID))
		return nil
	}

	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %v", toolID, err)
	}

	t.LogOperationWithUser("Updating tool press", createUserInfo(user), fmt.Sprintf("id: %d, press: %v", toolID, press))

	const query = `UPDATE tools SET press = ? WHERE id = ?`
	result, err := t.db.Exec(query, press, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	t.LogOperation("Updated tool press", fmt.Sprintf("id: %d", toolID))
	return nil
}

func (t *Tools) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	if err := ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := ValidateNotNil(user, "user"); err != nil {
		return err
	}

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %v", err)
	}

	if tool.Regenerating == regenerating {
		t.LogOperation("Tool regenerating status unchanged", fmt.Sprintf("id: %d", toolID))
		return nil
	}

	t.LogOperationWithUser("Updating tool regenerating status", createUserInfo(user), fmt.Sprintf("id: %d, regenerating: %v", toolID, regenerating))

	const query = `UPDATE tools SET regenerating = ? WHERE id = ?`
	result, err := t.db.Exec(query, regenerating, tool.ID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	t.LogOperation("Updated tool regenerating status", fmt.Sprintf("id: %d", toolID))
	return nil
}

func (t *Tools) marshalToolData(tool *models.Tool) ([]byte, error) {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return nil, fmt.Errorf("marshal error: tools: %v", err)
	}

	return formatBytes, nil
}

func (t *Tools) validateToolUniqueness(tool *models.Tool, excludeID int64) error {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return fmt.Errorf("failed to marshal format data: %v", err)
	}

	exists, err := t.CheckExistence(`
		SELECT COUNT(*) FROM tools
		WHERE id != $1 AND position = $2 AND format = $3 AND code = $4
	`, excludeID, tool.Position, formatBytes, tool.Code)

	if err != nil {
		return t.HandleSelectError(err, "tools")
	}

	if exists {
		return utils.NewAlreadyExistsError(
			fmt.Sprintf(
				"tool with position %s, format %s, and code %s already exists",
				tool.Position, tool.Format, tool.Code,
			),
		)
	}

	return nil
}

func equalPressNumbers(a, b *models.PressNumber) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
