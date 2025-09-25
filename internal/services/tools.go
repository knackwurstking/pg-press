package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/interfaces"
	"github.com/knackwurstking/pgpress/internal/logger"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Tools struct {
	db    *sql.DB
	notes *Notes
}

func NewTools(db *sql.DB, notes *Notes) *Tools {
	t := &Tools{
		db:    db,
		notes: notes,
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
			notes BLOB NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`

	if _, err := t.db.Exec(query); err != nil {
		return fmt.Errorf("failed to create tools table: %v", err)
	}

	return nil
}

func (t *Tools) Add(tool *models.Tool, user *models.User) (int64, error) {
	if err := t.validateToolUniqueness(tool, 0); err != nil {
		return 0, err
	}

	formatBytes, notesBytes, err := t.marshalToolData(tool)
	if err != nil {
		return 0, err
	}

	const insertQuery = `
		INSERT INTO tools (position, format, type, code, regenerating, press, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	result, err := t.db.Exec(insertQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, notesBytes)
	if err != nil {
		return 0, fmt.Errorf("insert error: tools: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert error: tools: %v", err)
	}

	return id, nil
}

func (t *Tools) AddWithNotes(tool *models.Tool, user *models.User, notes ...*models.Note) (*models.ToolWithNotes, error) {
	var noteIDs []int64
	for _, note := range notes {
		noteID, err := t.notes.Add(note)
		if err != nil {
			return nil, err
		}
		noteIDs = append(noteIDs, noteID)
	}

	tool.LinkedNotes = noteIDs
	toolID, err := t.Add(tool, user)
	if err != nil {
		return nil, err
	}

	tool.ID = toolID
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Tools) Delete(id int64, user *models.User) error {
	const deleteQuery = `DELETE FROM tools WHERE id = $1`
	_, err := t.db.Exec(deleteQuery, id)
	if err != nil {
		return fmt.Errorf("delete error: tools: %v", err)
	}

	return nil
}

func (t *Tools) Get(id int64) (*models.Tool, error) {
	const query = `SELECT id, position, format, type, code, regenerating, press, notes FROM tools WHERE id = $1`
	row := t.db.QueryRow(query, id)

	tool, err := t.scanTool(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool with ID %d not found", id))
		}
		return nil, fmt.Errorf("select error: tools: %v", err)
	}

	return tool, nil
}

func (t *Tools) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	const query = `
		SELECT id, position, format, type, code, regenerating, press, notes
		FROM tools WHERE regenerating = 0 AND press = ?
	`
	rows, err := t.db.Query(query, pressNumber)
	if err != nil {
		logger.DBTools().Error("Failed to query active tools: %v", err)
		return nil
	}
	defer rows.Close()

	var tools []*models.Tool
	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			return nil
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		logger.DBTools().Error("Error iterating over tool rows: %v", err)
		return nil
	}
	return tools
}

func (t *Tools) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	const query = `
		SELECT id, position, format, type, code, regenerating, press, notes
		FROM tools WHERE press = $1 AND regenerating = 0
	`
	rows, err := t.db.Query(query, pressNumber)
	if err != nil {
		return nil, fmt.Errorf("select error: tools: %v", err)
	}
	defer rows.Close()

	var tools []*models.Tool
	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: tools: %v", err)
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: tools: %v", err)
	}

	return tools, nil
}

func (t *Tools) GetPressUtilization() ([]models.PressUtilization, error) {
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

	return utilization, nil
}

func (t *Tools) GetWithNotes(id int64) (*models.ToolWithNotes, error) {
	tool, err := t.Get(id)
	if err != nil {
		return nil, err
	}

	notes, err := t.notes.GetByIDs(tool.LinkedNotes)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes for tool")
	}

	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Tools) List() ([]*models.Tool, error) {
	const query = `
		SELECT
			id, position, format, type, code, regenerating, press, notes
		FROM
			tools
		ORDER BY format ASC, code ASC
	`

	rows, err := t.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("select error: tools: %v", err)
	}
	defer rows.Close()

	var tools []*models.Tool
	for rows.Next() {
		tool, err := t.scanTool(rows)
		if err != nil {
			return nil, fmt.Errorf("scan error: tools: %v", err)
		}
		tools = append(tools, tool)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select error: tools: %v", err)
	}

	return tools, nil
}

func (t *Tools) ListWithNotes() ([]*models.ToolWithNotes, error) {
	tools, err := t.List()
	if err != nil {
		return nil, err
	}

	var result []*models.ToolWithNotes
	for _, tool := range tools {
		notes, err := t.notes.GetByIDs(tool.LinkedNotes)
		if err != nil {
			return nil, fmt.Errorf("failed to load notes for tool %d: %v", tool.ID, err)
		}

		result = append(result, &models.ToolWithNotes{Tool: tool, LoadedNotes: notes})
	}

	return result, nil
}

func (t *Tools) Update(tool *models.Tool, user *models.User) error {
	if err := t.validateToolUniqueness(tool, tool.ID); err != nil {
		return err
	}

	formatBytes, notesBytes, err := t.marshalToolData(tool)
	if err != nil {
		return err
	}

	const updateQuery = `
		UPDATE tools SET position = $1, format = $2, type = $3, code = $4,
		regenerating = $5, press = $6, notes = $7 WHERE id = $8
	`
	_, err = t.db.Exec(updateQuery, tool.Position, formatBytes, tool.Type, tool.Code,
		tool.Regenerating, tool.Press, notesBytes, tool.ID)
	if err != nil {
		return fmt.Errorf("update error: tools: %v", err)
	}

	return nil
}

func (t *Tools) UpdatePress(toolID int64, press *models.PressNumber, user *models.User) error {
	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %v", err)
	}

	if equalPressNumbers(tool.Press, press) {
		return nil
	}

	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %v", toolID, err)
	}

	const query = `UPDATE tools SET press = ? WHERE id = ?`
	_, err = t.db.Exec(query, press, toolID)
	if err != nil {
		return fmt.Errorf("update error: tools: %v", err)
	}

	return nil
}

func (t *Tools) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for regenerating status update: %v", err)
	}

	if tool.Regenerating == regenerating {
		return nil
	}

	const query = `UPDATE tools SET regenerating = ? WHERE id = ?`
	_, err = t.db.Exec(query, regenerating, tool.ID)
	if err != nil {
		return fmt.Errorf("update error: tools: %v", err)
	}

	return nil
}

func (t *Tools) marshalToolData(tool *models.Tool) ([]byte, []byte, error) {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal error: tools: %v", err)
	}

	notesBytes, err := json.Marshal(tool.LinkedNotes)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal error: tools: %v", err)
	}

	return formatBytes, notesBytes, nil
}

func (t *Tools) scanTool(scanner interfaces.Scannable) (*models.Tool, error) {
	tool := &models.Tool{}
	var format, linkedNotes []byte

	if err := scanner.Scan(&tool.ID, &tool.Position, &format, &tool.Type,
		&tool.Code, &tool.Regenerating, &tool.Press, &linkedNotes); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(format, &tool.Format); err != nil {
		return nil, fmt.Errorf("failed to unmarshal format data")
	}

	if err := json.Unmarshal(linkedNotes, &tool.LinkedNotes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal linked notes data")
	}

	return tool, nil
}

func (t *Tools) validateToolUniqueness(tool *models.Tool, excludeID int64) error {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return fmt.Errorf("failed to marshal format data: %v", err)
	}

	var count int
	const query = `
		SELECT COUNT(*) FROM tools
		WHERE id != $1 AND position = $2 AND format = $3 AND code = $4
	`
	err = t.db.QueryRow(query, excludeID, tool.Position, formatBytes, tool.Code).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for existing tool: %v", err)
	}

	if count > 0 {
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
