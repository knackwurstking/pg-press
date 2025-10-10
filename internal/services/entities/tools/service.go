package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/knackwurstking/pgpress/internal/services/base"
	"github.com/knackwurstking/pgpress/internal/services/shared/scanner"
	"github.com/knackwurstking/pgpress/internal/services/shared/validation"
	"github.com/knackwurstking/pgpress/pkg/models"
	"github.com/knackwurstking/pgpress/pkg/utils"
)

type Service struct {
	*base.BaseService
	notes NotesService
}

// NotesService defines the interface for notes service methods used by Tools
type NotesService interface {
	Add(note *models.Note) (int64, error)
	Delete(id int64, user *models.User) error
	GetByTool(toolID int64) ([]*models.Note, error)
}

func NewService(db *sql.DB, notes NotesService) *Service {
	baseService := base.NewBaseService(db, "Tools")

	t := &Service{
		BaseService: baseService,
		notes:       notes,
	}

	if err := t.createTable(); err != nil {
		panic(err)
	}

	return t
}

func (t *Service) createTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS tools (
			id INTEGER NOT NULL,
			position TEXT NOT NULL,
			format BLOB NOT NULL,
			type TEXT NOT NULL,
			code TEXT NOT NULL,
			regenerating INTEGER NOT NULL DEFAULT 0,
			press INTEGER,
			PRIMARY KEY("id" AUTOINCREMENT)
		);
	`
	return t.CreateTable(query, "tools")
}

func (t *Service) Add(tool *models.Tool, user *models.User) (int64, error) {
	if err := ValidateTool(tool); err != nil {
		return 0, err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return 0, err
	}

	t.Log.Info("Adding tool by %s: position: %s, type: %s, code: %s", user.String(), tool.Position, tool.Type, tool.Code)

	if err := t.validateToolUniqueness(tool, 0); err != nil {
		return 0, err
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return 0, err
	}

	const insertQuery = `
		INSERT INTO tools (position, format, type, code, regenerating, press)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := t.DB.Exec(insertQuery, tool.Position, formatBytes, tool.Type, tool.Code, tool.Regenerating, tool.Press)
	if err != nil {
		return 0, t.HandleInsertError(err, "tools")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %v", err)
	}

	t.Log.Info("Added tool: id: %d", id)
	return id, nil
}

func (t *Service) AddWithNotes(tool *models.Tool, user *models.User, notes ...*models.Note) (*models.ToolWithNotes, error) {
	if err := ValidateTool(tool); err != nil {
		return nil, err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return nil, err
	}

	t.Log.Info("Adding tool with notes by %s: notes_count: %d", user.String(), len(notes))

	var createdNotes []*models.Note
	for _, note := range notes {
		noteID, err := t.notes.Add(note)
		if err != nil {
			// Cleanup previously created notes on failure
			for _, cn := range createdNotes {
				if deleteErr := t.notes.Delete(cn.ID, user); deleteErr != nil {
					t.Log.Error("Failed to cleanup note %d: %v", cn.ID, deleteErr)
				}
			}
			return nil, fmt.Errorf("failed to create note: %v", err)
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

func (t *Service) Delete(id int64, user *models.User) error {
	if err := validation.ValidateID(id, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Info("Deleting tool by %s: id: %d", user.String(), id)

	const deleteQuery = `DELETE FROM tools WHERE id = ?`
	result, err := t.DB.Exec(deleteQuery, id)
	if err != nil {
		return t.HandleDeleteError(err, "tools")
	}

	return t.CheckRowsAffected(result, "tool", id)
}

func (t *Service) Get(id int64) (*models.Tool, error) {
	if err := validation.ValidateID(id, "tool"); err != nil {
		return nil, err
	}

	t.Log.Info("Getting tool: %d", id)

	const query = `SELECT id, position, format, type, code, regenerating, press FROM tools WHERE id = ?`
	row := t.DB.QueryRow(query, id)

	tool, err := scanner.ScanSingleRow(row, ScanTool, "tools")
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NewNotFoundError(fmt.Sprintf("tool with ID %d not found", id))
		}
		return nil, err
	}

	return tool, nil
}

func (t *Service) GetActiveToolsForPress(pressNumber models.PressNumber) []*models.Tool {
	t.Log.Info("Getting active tools for press: %d", pressNumber)

	const query = `
		SELECT id, position, format, type, code, regenerating, press
		FROM tools WHERE regenerating = 0 AND press = ?
	`
	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		t.Log.Error("Failed to query active tools: %v", err)
		return nil
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		t.Log.Error("Failed to scan active tools: %v", err)
		return nil
	}

	t.Log.Info("Found active tools for press: %d, count: %d", pressNumber, len(tools))
	return tools
}

func (t *Service) GetByPress(pressNumber *models.PressNumber) ([]*models.Tool, error) {
	if pressNumber != nil && !models.IsValidPressNumber(pressNumber) {
		return nil, fmt.Errorf("invalid press number: %d (must be 0-5)", *pressNumber)
	}

	t.Log.Info("Getting tools by press: %v", pressNumber)

	const query = `
		SELECT id, position, format, type, code, regenerating, press
		FROM tools WHERE press = ? AND regenerating = 0
	`
	rows, err := t.DB.Query(query, pressNumber)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	t.Log.Info("Found tools by press: %v, count: %d", pressNumber, len(tools))
	return tools, nil
}

func (t *Service) GetPressUtilization() ([]models.PressUtilization, error) {
	t.Log.Info("Getting press utilization")

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

	t.Log.Info("Calculated press utilization: presses_checked: %d", len(validPresses))
	return utilization, nil
}

func (t *Service) GetWithNotes(id int64) (*models.ToolWithNotes, error) {
	if err := validation.ValidateID(id, "tool"); err != nil {
		return nil, err
	}

	t.Log.Info("Getting tool with notes: %d", id)

	tool, err := t.Get(id)
	if err != nil {
		return nil, err
	}

	notes, err := t.notes.GetByTool(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes for tool")
	}

	t.Log.Info("Found tool with notes: id: %d, notes_count: %d", id, len(notes))
	return &models.ToolWithNotes{Tool: tool, LoadedNotes: notes}, nil
}

func (t *Service) List() ([]*models.Tool, error) {
	t.Log.Info("Listing tools")

	const query = `
		SELECT
			id, position, format, type, code, regenerating, press
		FROM
			tools
		ORDER BY format ASC, code ASC
	`

	rows, err := t.DB.Query(query)
	if err != nil {
		return nil, t.HandleSelectError(err, "tools")
	}
	defer rows.Close()

	tools, err := ScanToolsFromRows(rows)
	if err != nil {
		return nil, err
	}

	t.Log.Info("Listed tools: count: %d", len(tools))
	return tools, nil
}

func (t *Service) ListWithNotes() ([]*models.ToolWithNotes, error) {
	t.Log.Info("Listing tools with notes")

	tools, err := t.List()
	if err != nil {
		return nil, err
	}

	var result []*models.ToolWithNotes
	for _, tool := range tools {
		notes, err := t.notes.GetByTool(tool.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load notes for tool %d", tool.ID)
		}
		result = append(result, &models.ToolWithNotes{Tool: tool, LoadedNotes: notes})
	}

	t.Log.Info("Listed tools with notes: count: %d", len(result))
	return result, nil
}

func (t *Service) Update(tool *models.Tool, user *models.User) error {
	if err := ValidateTool(tool); err != nil {
		return err
	}

	if err := validation.ValidateID(tool.ID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Info("Updating tool by %s: id: %d", user.String(), tool.ID)

	if err := t.validateToolUniqueness(tool, tool.ID); err != nil {
		return err
	}

	formatBytes, err := t.marshalFormat(tool.Format)
	if err != nil {
		return err
	}

	const updateQuery = `
		UPDATE tools
		SET position = ?, format = ?, type = ?, code = ?, regenerating = ?, press = ?
		WHERE id = ?
	`

	result, err := t.DB.Exec(updateQuery,
		tool.Position,
		formatBytes,
		tool.Type,
		tool.Code,
		tool.Regenerating,
		tool.Press,
		tool.ID,
	)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", tool.ID); err != nil {
		return err
	}

	t.Log.Info("Updated tool: id: %d", tool.ID)
	return nil
}

func (t *Service) UpdatePress(toolID int64, press *models.PressNumber, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	t.Log.Info("Updating tool press by %s: id: %d, press: %v", user.String(), toolID, press)

	tool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get tool for press update: %v", err)
	}

	if equalPressNumbers(tool.Press, press) {
		t.Log.Info("Tool press unchanged: id: %d", toolID)
		return nil
	}

	if err := tool.SetPress(press); err != nil {
		return fmt.Errorf("failed to set press for tool %d: %v", toolID, err)
	}

	const query = `UPDATE tools SET press = ? WHERE id = ?`
	result, err := t.DB.Exec(query, press, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	t.Log.Info("Updated tool press: id: %d, press: %v", toolID, press)
	return nil
}

func (t *Service) UpdateRegenerating(toolID int64, regenerating bool, user *models.User) error {
	if err := validation.ValidateID(toolID, "tool"); err != nil {
		return err
	}

	if err := validation.ValidateNotNil(user, "user"); err != nil {
		return err
	}

	// Get the current tool to check if the regeneration status is actually changing
	currentTool, err := t.Get(toolID)
	if err != nil {
		return fmt.Errorf("failed to get current tool state: %v", err)
	}

	if currentTool.Regenerating == regenerating {
		t.Log.Info("Tool regeneration status unchanged by %s: id: %d, regenerating: %v", user.String(), toolID, regenerating)
		return nil
	}

	t.Log.Info("Updating tool regenerating status by %s: id: %d, regenerating: %v", user.String(), toolID, regenerating)

	const query = `UPDATE tools SET regenerating = ? WHERE id = ?`
	result, err := t.DB.Exec(query, regenerating, toolID)
	if err != nil {
		return t.HandleUpdateError(err, "tools")
	}

	if err := t.CheckRowsAffected(result, "tool", toolID); err != nil {
		return err
	}

	t.Log.Info("Updated tool regenerating status: id: %d, regenerating: %v", toolID, regenerating)
	return nil
}

func (t *Service) marshalFormat(format models.Format) ([]byte, error) {
	formatBytes, err := json.Marshal(format)
	if err != nil {
		return nil, fmt.Errorf("marshal error: tools: %v", err)
	}

	return formatBytes, nil
}

func (t *Service) validateToolUniqueness(tool *models.Tool, excludeID int64) error {
	formatBytes, err := json.Marshal(tool.Format)
	if err != nil {
		return fmt.Errorf("failed to marshal tool format: %v", err)
	}

	exists, err := t.CheckExistence(`
		SELECT COUNT(*) FROM tools
		WHERE id != ? AND position = ? AND format = ? AND code = ?
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
